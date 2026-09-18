package core

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStartTLSProtocolsReachTLSHandshakeWithoutAuthentication(t *testing.T) {
	for _, protocol := range []string{"smtp", "imap", "pop3"} {
		t.Run(protocol, func(t *testing.T) {
			target, cleanup := startFakeStartTLSServer(t, protocol, true, false)
			defer cleanup()

			story, err := BuildStartTLSStory(t.Context(), protocol, target)
			if err != nil {
				t.Fatal(err)
			}
			if !story.StartTLSAdvertised {
				t.Fatalf("expected advertised upgrade capability: %#v", story)
			}
			if story.TLS == nil || story.TLS.HandshakeStatus != "pass" {
				t.Fatalf("expected successful TLS handshake: %#v", story.TLS)
			}
			if story.TLS.HostnameStatus != "pass" {
				t.Fatalf("expected localhost hostname match: %#v", story.TLS)
			}
			if story.TLS.TrustStatus != "fail" {
				t.Fatalf("self-signed fixture should remain untrusted: %#v", story.TLS)
			}
			if story.FirstProblem != "trust" {
				t.Fatalf("expected trust as first problem, got %q (%#v)", story.FirstProblem, story.Stages)
			}
			if stageStatus(story, "tcp") != "pass" ||
				stageStatus(story, "greeting") != "pass" ||
				stageStatus(story, "capability") != "pass" ||
				stageStatus(story, "upgrade") != "pass" ||
				stageStatus(story, "tls") != "pass" {
				t.Fatalf("protocol stages did not reach TLS cleanly: %#v", story.Stages)
			}
		})
	}
}

func TestSMTPMissingSTARTTLSStopsBeforeUpgrade(t *testing.T) {
	target, cleanup := startFakeStartTLSServer(t, "smtp", false, false)
	defer cleanup()

	story, err := BuildStartTLSStory(t.Context(), "smtp", target)
	if err != nil {
		t.Fatal(err)
	}
	if story.Status != "fail" || story.FirstProblem != "capability" {
		t.Fatalf("unexpected story outcome: %#v", story)
	}
	if story.TLS != nil {
		t.Fatalf("TLS must not be attempted without advertised STARTTLS: %#v", story.TLS)
	}
}

func TestIMAPRejectedSTARTTLSIsUpgradeFailure(t *testing.T) {
	target, cleanup := startFakeStartTLSServer(t, "imap", true, true)
	defer cleanup()

	story, err := BuildStartTLSStory(t.Context(), "imap", target)
	if err != nil {
		t.Fatal(err)
	}
	if story.Status != "fail" || story.FirstProblem != "upgrade" {
		t.Fatalf("unexpected story outcome: %#v", story)
	}
	if story.TLS != nil {
		t.Fatalf("TLS must not be attempted after rejected STARTTLS: %#v", story.TLS)
	}
}

func TestStartTLSInputValidation(t *testing.T) {
	for _, tc := range []struct {
		protocol string
		target   string
	}{
		{"smtp", "missing-port"},
		{"imap", ":143"},
		{"pop3", "localhost:0"},
		{"smtp", "localhost:not-a-port"},
		{"ftp", "localhost:21"},
	} {
		if _, err := BuildStartTLSStory(t.Context(), tc.protocol, tc.target); err == nil {
			t.Fatalf("expected validation failure for protocol=%q target=%q", tc.protocol, tc.target)
		}
	}
}

func TestReadProtocolLineRejectsOversizedInput(t *testing.T) {
	reader := bufio.NewReaderSize(strings.NewReader(strings.Repeat("A", startTLSMaxLineBytes+1)+"\r\n"), startTLSMaxLineBytes)
	if _, err := readProtocolLine(reader); err == nil || !strings.Contains(err.Error(), "bounded line length") {
		t.Fatalf("expected bounded line failure, got %v", err)
	}
}

func TestReadProtocolLineRejectsUnterminatedInput(t *testing.T) {
	reader := bufio.NewReaderSize(strings.NewReader("220 partial greeting"), startTLSMaxLineBytes)
	if _, err := readProtocolLine(reader); err == nil || !strings.Contains(err.Error(), "newline terminator") {
		t.Fatalf("expected unterminated line failure, got %v", err)
	}
}

func TestSMTPResponseRejectsInvalidSeparator(t *testing.T) {
	reader := bufio.NewReaderSize(strings.NewReader("250Xinvalid\r\n"), startTLSMaxLineBytes)
	if _, _, err := readSMTPResponse(reader); err == nil || !strings.Contains(err.Error(), "invalid reply separator") {
		t.Fatalf("expected invalid SMTP separator failure, got %v", err)
	}
}

func TestPOP3FailureWordingUsesSTLS(t *testing.T) {
	if got := startTLSFailureConclusion("pop3", "capability"); !strings.Contains(got, "STLS") || strings.Contains(got, "STARTTLS") {
		t.Fatalf("unexpected POP3 wording: %q", got)
	}
}

func TestFinalizeStartTLSStoryUsesFirstProvenFailure(t *testing.T) {
	story := StartTLSStory{
		Protocol: "smtp",
		Stages: []StartTLSStage{
			{ID: "tcp", Title: "TCP", Status: "pass"},
			{ID: "greeting", Title: "Greeting", Status: "unknown"},
			{ID: "capability", Title: "Capability", Status: "fail"},
			{ID: "trust", Title: "Trust", Status: "fail"},
		},
	}
	finalizeStartTLSStory(&story)
	if story.Status != "fail" || story.FirstProblem != "capability" {
		t.Fatalf("unexpected precedence: %#v", story)
	}
}

func stageStatus(story StartTLSStory, id string) string {
	for _, stage := range story.Stages {
		if stage.ID == id {
			return stage.Status
		}
	}
	return ""
}

func startFakeStartTLSServer(t *testing.T, protocol string, advertise, reject bool) (string, func()) {
	t.Helper()
	certPEM, keyPEM, _ := makeTestCertificate(t, "localhost", time.Now().Add(-time.Hour), time.Now().Add(24*time.Hour))
	pair, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	errCh := make(chan error, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)

		send := func(value string) error {
			if _, err := writer.WriteString(value); err != nil {
				return err
			}
			return writer.Flush()
		}
		read := func() (string, error) {
			value, err := reader.ReadString('\n')
			return strings.TrimSpace(value), err
		}

		switch protocol {
		case "smtp":
			if err := send("220 localhost ESMTP fixture\r\n"); err != nil {
				errCh <- err
				return
			}
			line, err := read()
			if err != nil || !strings.HasPrefix(line, "EHLO ") {
				errCh <- fmt.Errorf("unexpected SMTP EHLO: %q err=%v", line, err)
				return
			}
			if advertise {
				if err := send("250-localhost\r\n250-STARTTLS\r\n250 SIZE 1000\r\n"); err != nil {
					errCh <- err
					return
				}
			} else {
				if err := send("250-localhost\r\n250 SIZE 1000\r\n"); err != nil {
					errCh <- err
				}
				return
			}
			line, err = read()
			if err != nil || line != "STARTTLS" {
				errCh <- fmt.Errorf("unexpected SMTP STARTTLS: %q err=%v", line, err)
				return
			}
			if reject {
				errCh <- send("454 TLS temporarily unavailable\r\n")
				return
			}
			if err := send("220 Ready to start TLS\r\n"); err != nil {
				errCh <- err
				return
			}

		case "imap":
			if err := send("* OK IMAP fixture ready\r\n"); err != nil {
				errCh <- err
				return
			}
			line, err := read()
			if err != nil || line != "a001 CAPABILITY" {
				errCh <- fmt.Errorf("unexpected IMAP CAPABILITY: %q err=%v", line, err)
				return
			}
			capability := "* CAPABILITY IMAP4rev2 LOGINDISABLED"
			if advertise {
				capability += " STARTTLS"
			}
			if err := send(capability + "\r\na001 OK CAPABILITY completed\r\n"); err != nil {
				errCh <- err
				return
			}
			if !advertise {
				return
			}
			line, err = read()
			if err != nil || line != "a002 STARTTLS" {
				errCh <- fmt.Errorf("unexpected IMAP STARTTLS: %q err=%v", line, err)
				return
			}
			if reject {
				errCh <- send("a002 NO TLS unavailable\r\n")
				return
			}
			if err := send("a002 OK Begin TLS negotiation\r\n"); err != nil {
				errCh <- err
				return
			}

		case "pop3":
			if err := send("+OK POP3 fixture ready\r\n"); err != nil {
				errCh <- err
				return
			}
			line, err := read()
			if err != nil || line != "CAPA" {
				errCh <- fmt.Errorf("unexpected POP3 CAPA: %q err=%v", line, err)
				return
			}
			capabilities := "+OK Capability list follows\r\nTOP\r\nUIDL\r\n"
			if advertise {
				capabilities += "STLS\r\n"
			}
			capabilities += ".\r\n"
			if err := send(capabilities); err != nil {
				errCh <- err
				return
			}
			if !advertise {
				return
			}
			line, err = read()
			if err != nil || line != "STLS" {
				errCh <- fmt.Errorf("unexpected POP3 STLS: %q err=%v", line, err)
				return
			}
			if reject {
				errCh <- send("-ERR TLS unavailable\r\n")
				return
			}
			if err := send("+OK Begin TLS negotiation\r\n"); err != nil {
				errCh <- err
				return
			}
		default:
			errCh <- fmt.Errorf("unsupported fixture protocol %q", protocol)
			return
		}

		tlsServer := tls.Server(conn, &tls.Config{Certificates: []tls.Certificate{pair}})
		if err := tlsServer.Handshake(); err != nil {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	target := net.JoinHostPort("localhost", port)
	cleanup := func() {
		_ = listener.Close()
		select {
		case err := <-errCh:
			if err != nil {
				t.Errorf("fixture server: %v", err)
			}
		case <-done:
			select {
			case err := <-errCh:
				if err != nil {
					t.Errorf("fixture server: %v", err)
				}
			default:
			}
		case <-time.After(2 * time.Second):
			t.Error("fixture server did not exit")
		}
	}
	return target, cleanup
}
