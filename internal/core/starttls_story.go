package core

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	startTLSConnectTimeout = 4 * time.Second
	startTLSIOTimeout      = 4 * time.Second
	startTLSMaxLineBytes   = 4096
	startTLSMaxLines       = 64
	startTLSLineEvidence   = 1024
)

type StartTLSStage struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Summary  string   `json:"summary"`
	Evidence []string `json:"evidence,omitempty"`
}

type StartTLSStory struct {
	Protocol           string          `json:"protocol"`
	Target             string          `json:"target"`
	ServerName         string          `json:"server_name"`
	Port               string          `json:"port"`
	StartedAt          time.Time       `json:"started_at"`
	Status             string          `json:"status"`
	Conclusion         string          `json:"conclusion"`
	FirstProblem       string          `json:"first_problem,omitempty"`
	Greeting           string          `json:"greeting,omitempty"`
	Capabilities       []string        `json:"capabilities,omitempty"`
	StartTLSAdvertised bool            `json:"starttls_advertised"`
	UpgradeCommand     string          `json:"upgrade_command,omitempty"`
	UpgradeResponse    string          `json:"upgrade_response,omitempty"`
	TLS                *TLSEvidence    `json:"tls,omitempty"`
	Stages             []StartTLSStage `json:"stages"`
	ScopeNotes         []string        `json:"scope_notes,omitempty"`
}

var startTLSDial = func(ctx context.Context, target string) (net.Conn, error) {
	dialer := net.Dialer{Timeout: startTLSConnectTimeout}
	return dialer.DialContext(ctx, "tcp", target)
}

func BuildStartTLSStory(ctx context.Context, protocol, target string) (StartTLSStory, error) {
	normalized, err := normalizeStartTLSProtocol(protocol)
	if err != nil {
		return StartTLSStory{}, err
	}
	host, port, err := validateStartTLSTarget(target)
	if err != nil {
		return StartTLSStory{}, err
	}

	story := StartTLSStory{
		Protocol:   normalized,
		Target:     strings.TrimSpace(target),
		ServerName: host,
		Port:       port,
		StartedAt:  time.Now().UTC(),
		ScopeNotes: []string{
			"Only greeting, capability discovery, the fixed STARTTLS/STLS upgrade command, TLS negotiation, and certificate metadata are inspected.",
			"No credentials, authentication commands, mail submission, mailbox access, message commands, or message contents are sent or read.",
			"Protocol reads are bounded by fixed line length, line count, and I/O deadlines.",
		},
	}

	raw, err := startTLSDial(ctx, story.Target)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"tcp", "TCP connection", "fail",
			"TCP connection to the mail endpoint failed",
			boundedEvidence(err.Error(), 512),
		))
		finalizeStartTLSStory(&story)
		return story, nil
	}
	defer raw.Close()
	_ = raw.SetDeadline(time.Now().Add(startTLSIOTimeout))
	story.Stages = append(story.Stages, startTLSStage(
		"tcp", "TCP connection", "pass",
		"TCP connection accepted",
		"target="+story.Target,
	))

	reader := bufio.NewReaderSize(raw, startTLSMaxLineBytes)
	ready := false
	switch normalized {
	case "smtp":
		ready = prepareSMTPStartTLS(raw, reader, &story)
	case "imap":
		ready = prepareIMAPStartTLS(raw, reader, &story)
	case "pop3":
		ready = preparePOP3StartTLS(raw, reader, &story)
	}
	if !ready {
		finalizeStartTLSStory(&story)
		return story, nil
	}

	_ = raw.SetDeadline(time.Now().Add(startTLSIOTimeout))
	handshakeCtx, cancel := context.WithTimeout(ctx, tlsProbeTimeout)
	defer cancel()
	tlsConn, evidence := probeTLSOnConnection(handshakeCtx, raw, host)
	if tlsConn != nil {
		defer tlsConn.Close()
	}
	story.TLS = evidence
	if evidence == nil {
		story.Stages = append(story.Stages, startTLSStage(
			"tls", "TLS negotiation", "unknown",
			"TLS negotiation returned no evidence",
		))
		finalizeStartTLSStory(&story)
		return story, nil
	}

	tlsSummary := strings.TrimSpace(evidence.Protocol + " " + evidence.CipherSuite)
	if evidence.HandshakeStatus != "pass" {
		detail := evidence.HandshakeError
		if detail == "" {
			detail = "no additional TLS error was returned"
		}
		story.Stages = append(story.Stages, startTLSStage(
			"tls", "TLS negotiation", normalizeStartTLSStatus(evidence.HandshakeStatus),
			"Server accepted the protocol upgrade, but TLS did not complete",
			detail,
		))
		finalizeStartTLSStory(&story)
		return story, nil
	}
	story.Stages = append(story.Stages, startTLSStage(
		"tls", "TLS negotiation", "pass",
		"TLS negotiation completed",
		tlsSummary,
	))

	certCheck := tlsCertificateCheck(evidence)
	story.Stages = append(story.Stages, startTLSStage(
		"certificate", "Certificate validity", normalizeStartTLSStatus(certCheck.Status),
		startTLSCheckSummary(certCheck, "Served certificate is within its validity window", "Served certificate validity check failed"),
		certCheck.Evidence,
	))
	hostCheck := tlsHostnameCheck(evidence)
	story.Stages = append(story.Stages, startTLSStage(
		"hostname", "Certificate hostname", normalizeStartTLSStatus(hostCheck.Status),
		startTLSCheckSummary(hostCheck, "Served certificate matches the requested host", "Served certificate does not match the requested host"),
		hostCheck.Evidence,
	))
	trustCheck := tlsTrustCheck(evidence)
	story.Stages = append(story.Stages, startTLSStage(
		"trust", "Certificate trust", normalizeStartTLSStatus(trustCheck.Status),
		startTLSCheckSummary(trustCheck, "Certificate chain is trusted by this host", "Certificate chain is not trusted by this host"),
		trustCheck.Evidence,
	))

	finalizeStartTLSStory(&story)
	return story, nil
}

func normalizeStartTLSProtocol(protocol string) (string, error) {
	switch value := strings.ToLower(strings.TrimSpace(protocol)); value {
	case "smtp", "imap", "pop3":
		return value, nil
	default:
		return "", errors.New("protocol must be one of: smtp, imap, pop3")
	}
}

func validateStartTLSTarget(target string) (string, string, error) {
	target = strings.TrimSpace(target)
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		return "", "", errors.New("target must be in host:port form")
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return "", "", errors.New("target host is required")
	}
	value, err := strconv.ParseUint(port, 10, 16)
	if err != nil || value == 0 {
		return "", "", errors.New("target port must be numeric and between 1 and 65535")
	}
	return host, port, nil
}

func prepareSMTPStartTLS(conn net.Conn, reader *bufio.Reader, story *StartTLSStory) bool {
	code, lines, err := readSMTPResponse(reader)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"greeting", "SMTP greeting", "fail",
			"SMTP greeting could not be read",
			err.Error(),
		))
		return false
	}
	story.Greeting = joinProtocolEvidence(lines)
	if code != 220 {
		story.Stages = append(story.Stages, startTLSStage(
			"greeting", "SMTP greeting", "fail",
			fmt.Sprintf("SMTP greeting returned %d instead of 220", code),
			story.Greeting,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"greeting", "SMTP greeting", "pass",
		"SMTP service returned a 220 greeting",
		story.Greeting,
	))

	if err := writeStartTLSCommand(conn, "EHLO hostsleuth.invalid\r\n"); err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "SMTP STARTTLS capability", "fail",
			"Could not send the bounded EHLO capability request",
			err.Error(),
		))
		return false
	}
	code, lines, err = readSMTPResponse(reader)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "SMTP STARTTLS capability", "fail",
			"SMTP EHLO response could not be read",
			err.Error(),
		))
		return false
	}
	story.Capabilities = smtpCapabilityLines(lines)
	story.StartTLSAdvertised = capabilityPresent(story.Capabilities, "STARTTLS")
	if code != 250 {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "SMTP STARTTLS capability", "fail",
			fmt.Sprintf("SMTP EHLO returned %d instead of 250", code),
			story.Capabilities...,
		))
		return false
	}
	if !story.StartTLSAdvertised {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "SMTP STARTTLS capability", "fail",
			"SMTP server did not advertise STARTTLS",
			story.Capabilities...,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"capability", "SMTP STARTTLS capability", "pass",
		"SMTP server advertised STARTTLS before authentication",
		story.Capabilities...,
	))

	story.UpgradeCommand = "STARTTLS"
	if err := writeStartTLSCommand(conn, "STARTTLS\r\n"); err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "SMTP STARTTLS command", "fail",
			"Could not send STARTTLS",
			err.Error(),
		))
		return false
	}
	code, lines, err = readSMTPResponse(reader)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "SMTP STARTTLS command", "fail",
			"SMTP STARTTLS response could not be read",
			err.Error(),
		))
		return false
	}
	story.UpgradeResponse = joinProtocolEvidence(lines)
	if code != 220 {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "SMTP STARTTLS command", "fail",
			fmt.Sprintf("SMTP server rejected STARTTLS with reply %d", code),
			story.UpgradeResponse,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"upgrade", "SMTP STARTTLS command", "pass",
		"SMTP server accepted STARTTLS and requested TLS negotiation",
		story.UpgradeResponse,
	))
	return true
}

func prepareIMAPStartTLS(conn net.Conn, reader *bufio.Reader, story *StartTLSStory) bool {
	greeting, err := readProtocolLine(reader)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"greeting", "IMAP greeting", "fail",
			"IMAP greeting could not be read",
			err.Error(),
		))
		return false
	}
	story.Greeting = greeting
	upper := strings.ToUpper(greeting)
	if strings.HasPrefix(upper, "* PREAUTH") {
		story.Stages = append(story.Stages, startTLSStage(
			"greeting", "IMAP greeting", "fail",
			"IMAP server greeted in PREAUTH state; STARTTLS is not valid after authentication",
			greeting,
		))
		return false
	}
	if !strings.HasPrefix(upper, "* OK") {
		story.Stages = append(story.Stages, startTLSStage(
			"greeting", "IMAP greeting", "fail",
			"IMAP server did not return an OK greeting",
			greeting,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"greeting", "IMAP greeting", "pass",
		"IMAP service returned an unauthenticated OK greeting",
		greeting,
	))

	if err := writeStartTLSCommand(conn, "a001 CAPABILITY\r\n"); err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "IMAP STARTTLS capability", "fail",
			"Could not send the bounded CAPABILITY request",
			err.Error(),
		))
		return false
	}
	lines, status, err := readIMAPTaggedResponse(reader, "a001")
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "IMAP STARTTLS capability", "fail",
			"IMAP CAPABILITY response could not be read",
			err.Error(),
		))
		return false
	}
	story.Capabilities = imapCapabilityLines(lines)
	story.StartTLSAdvertised = capabilityPresent(story.Capabilities, "STARTTLS")
	if status != "OK" {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "IMAP STARTTLS capability", "fail",
			"IMAP CAPABILITY command did not complete with OK",
			lines...,
		))
		return false
	}
	if !story.StartTLSAdvertised {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "IMAP STARTTLS capability", "fail",
			"IMAP server did not advertise STARTTLS",
			story.Capabilities...,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"capability", "IMAP STARTTLS capability", "pass",
		"IMAP server advertised STARTTLS before authentication",
		story.Capabilities...,
	))

	story.UpgradeCommand = "STARTTLS"
	if err := writeStartTLSCommand(conn, "a002 STARTTLS\r\n"); err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "IMAP STARTTLS command", "fail",
			"Could not send STARTTLS",
			err.Error(),
		))
		return false
	}
	lines, status, err = readIMAPTaggedResponse(reader, "a002")
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "IMAP STARTTLS command", "fail",
			"IMAP STARTTLS response could not be read",
			err.Error(),
		))
		return false
	}
	story.UpgradeResponse = joinProtocolEvidence(lines)
	if status != "OK" {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "IMAP STARTTLS command", "fail",
			"IMAP server rejected STARTTLS",
			lines...,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"upgrade", "IMAP STARTTLS command", "pass",
		"IMAP server accepted STARTTLS and requested TLS negotiation",
		story.UpgradeResponse,
	))
	return true
}

func preparePOP3StartTLS(conn net.Conn, reader *bufio.Reader, story *StartTLSStory) bool {
	greeting, err := readProtocolLine(reader)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"greeting", "POP3 greeting", "fail",
			"POP3 greeting could not be read",
			err.Error(),
		))
		return false
	}
	story.Greeting = greeting
	if !strings.HasPrefix(strings.ToUpper(greeting), "+OK") {
		story.Stages = append(story.Stages, startTLSStage(
			"greeting", "POP3 greeting", "fail",
			"POP3 server did not return a +OK greeting",
			greeting,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"greeting", "POP3 greeting", "pass",
		"POP3 service returned a +OK greeting",
		greeting,
	))

	if err := writeStartTLSCommand(conn, "CAPA\r\n"); err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "POP3 STLS capability", "fail",
			"Could not send the bounded CAPA request",
			err.Error(),
		))
		return false
	}
	first, err := readProtocolLine(reader)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "POP3 STLS capability", "fail",
			"POP3 CAPA response could not be read",
			err.Error(),
		))
		return false
	}
	if !strings.HasPrefix(strings.ToUpper(first), "+OK") {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "POP3 STLS capability", "fail",
			"POP3 CAPA command was rejected",
			first,
		))
		return false
	}
	lines, err := readPOP3Multiline(reader)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "POP3 STLS capability", "fail",
			"POP3 CAPA capability list could not be read",
			err.Error(),
		))
		return false
	}
	story.Capabilities = lines
	story.StartTLSAdvertised = capabilityPresent(story.Capabilities, "STLS")
	if !story.StartTLSAdvertised {
		story.Stages = append(story.Stages, startTLSStage(
			"capability", "POP3 STLS capability", "fail",
			"POP3 server did not advertise STLS",
			story.Capabilities...,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"capability", "POP3 STLS capability", "pass",
		"POP3 server advertised STLS in AUTHORIZATION state",
		story.Capabilities...,
	))

	story.UpgradeCommand = "STLS"
	if err := writeStartTLSCommand(conn, "STLS\r\n"); err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "POP3 STLS command", "fail",
			"Could not send STLS",
			err.Error(),
		))
		return false
	}
	response, err := readProtocolLine(reader)
	if err != nil {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "POP3 STLS command", "fail",
			"POP3 STLS response could not be read",
			err.Error(),
		))
		return false
	}
	story.UpgradeResponse = response
	if !strings.HasPrefix(strings.ToUpper(response), "+OK") {
		story.Stages = append(story.Stages, startTLSStage(
			"upgrade", "POP3 STLS command", "fail",
			"POP3 server rejected STLS",
			response,
		))
		return false
	}
	story.Stages = append(story.Stages, startTLSStage(
		"upgrade", "POP3 STLS command", "pass",
		"POP3 server accepted STLS and requested TLS negotiation",
		response,
	))
	return true
}

func readSMTPResponse(reader *bufio.Reader) (int, []string, error) {
	var lines []string
	var code int
	for len(lines) < startTLSMaxLines {
		line, err := readProtocolLine(reader)
		if err != nil {
			return 0, lines, err
		}
		if len(line) < 3 {
			return 0, lines, errors.New("SMTP response line is shorter than a reply code")
		}
		current, err := strconv.Atoi(line[:3])
		if err != nil {
			return 0, lines, errors.New("SMTP response did not begin with a numeric reply code")
		}
		if code == 0 {
			code = current
		} else if current != code {
			return 0, lines, errors.New("SMTP multiline response changed reply code")
		}
		lines = append(lines, line)
		if len(line) == 3 {
			return code, lines, nil
		}
		switch line[3] {
		case '-':
			continue
		case ' ':
			return code, lines, nil
		default:
			return 0, lines, errors.New("SMTP response used an invalid reply separator")
		}
	}
	return 0, lines, errors.New("SMTP response exceeded the bounded line limit")
}

func readIMAPTaggedResponse(reader *bufio.Reader, tag string) ([]string, string, error) {
	var lines []string
	tagUpper := strings.ToUpper(tag)
	for len(lines) < startTLSMaxLines {
		line, err := readProtocolLine(reader)
		if err != nil {
			return lines, "", err
		}
		lines = append(lines, line)
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.ToUpper(fields[0]) == tagUpper {
			return lines, strings.ToUpper(fields[1]), nil
		}
	}
	return lines, "", errors.New("IMAP response exceeded the bounded line limit")
}

func readPOP3Multiline(reader *bufio.Reader) ([]string, error) {
	var lines []string
	for len(lines) < startTLSMaxLines {
		line, err := readProtocolLine(reader)
		if err != nil {
			return lines, err
		}
		if line == "." {
			return lines, nil
		}
		if strings.HasPrefix(line, "..") {
			line = line[1:]
		}
		lines = append(lines, line)
	}
	return lines, errors.New("POP3 response exceeded the bounded line limit")
}

func readProtocolLine(reader *bufio.Reader) (string, error) {
	raw, err := reader.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) {
		return "", errors.New("protocol line exceeded the bounded line length")
	}
	if err != nil {
		if errors.Is(err, io.EOF) {
			if len(raw) == 0 {
				return "", io.EOF
			}
			return "", errors.New("protocol line ended before a newline terminator")
		}
		return "", err
	}
	return safeProtocolLine(string(raw)), nil
}

func writeStartTLSCommand(conn net.Conn, command string) error {
	if err := conn.SetWriteDeadline(time.Now().Add(startTLSIOTimeout)); err != nil {
		return err
	}
	_, err := io.WriteString(conn, command)
	if err != nil {
		return err
	}
	return conn.SetReadDeadline(time.Now().Add(startTLSIOTimeout))
}

func safeProtocolLine(value string) string {
	value = strings.TrimRight(value, "\r\n")
	value = strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case r < 0x20 || r == 0x7f:
			return -1
		default:
			return r
		}
	}, value)
	return boundedEvidence(value, startTLSLineEvidence)
}

func joinProtocolEvidence(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return boundedEvidence(strings.Join(lines, " | "), 4096)
}

func smtpCapabilityLines(lines []string) []string {
	var out []string
	for _, line := range lines {
		if len(line) <= 4 {
			continue
		}
		value := strings.TrimSpace(line[4:])
		if value != "" {
			out = append(out, safeProtocolLine(value))
		}
	}
	return uniqueBoundedCapabilities(out)
}

func imapCapabilityLines(lines []string) []string {
	var out []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != "*" || !strings.EqualFold(fields[1], "CAPABILITY") {
			continue
		}
		for _, capability := range fields[2:] {
			out = append(out, safeProtocolLine(capability))
		}
	}
	return uniqueBoundedCapabilities(out)
}

func uniqueBoundedCapabilities(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToUpper(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
		if len(out) >= startTLSMaxLines {
			break
		}
	}
	return out
}

func capabilityPresent(capabilities []string, want string) bool {
	for _, capability := range capabilities {
		fields := strings.Fields(capability)
		if len(fields) > 0 && strings.EqualFold(fields[0], want) {
			return true
		}
	}
	return false
}

func startTLSStage(id, title, status, summary string, evidence ...string) StartTLSStage {
	clean := make([]string, 0, len(evidence))
	for _, item := range evidence {
		if value := safeProtocolLine(item); value != "" {
			clean = append(clean, value)
		}
	}
	return StartTLSStage{
		ID:       id,
		Title:    title,
		Status:   normalizeStartTLSStatus(status),
		Summary:  summary,
		Evidence: clean,
	}
}

func normalizeStartTLSStatus(status string) string {
	switch status {
	case "pass", "fail", "unknown":
		return status
	default:
		return "unknown"
	}
}

func startTLSCheckSummary(check Check, passText, failText string) string {
	switch check.Status {
	case "pass":
		return passText
	case "fail":
		return failText
	default:
		return "Evidence was incomplete"
	}
}

func finalizeStartTLSStory(story *StartTLSStory) {
	if story == nil {
		return
	}
	for _, stage := range story.Stages {
		if stage.Status == "fail" {
			story.Status = "fail"
			story.FirstProblem = stage.ID
			story.Conclusion = startTLSFailureConclusion(story.Protocol, stage.ID)
			return
		}
	}
	for _, stage := range story.Stages {
		if stage.Status == "unknown" {
			story.Status = "unknown"
			story.FirstProblem = stage.ID
			story.Conclusion = "STARTTLS evidence is incomplete at " + stage.Title
			return
		}
	}
	story.Status = "pass"
	story.Conclusion = strings.ToUpper(story.Protocol) + " " + startTLSUpgradeName(story.Protocol) + " negotiated successfully and certificate checks passed"
}

func startTLSUpgradeName(protocol string) string {
	if strings.EqualFold(protocol, "pop3") {
		return "STLS"
	}
	return "STARTTLS"
}

func startTLSFailureConclusion(protocol, stage string) string {
	name := strings.ToUpper(protocol)
	upgrade := startTLSUpgradeName(protocol)
	switch stage {
	case "tcp":
		return "TCP connection to the " + name + " endpoint failed"
	case "greeting":
		return name + " service greeting did not establish a valid " + upgrade + "-capable pre-authentication session"
	case "capability":
		return name + " service did not advertise the required " + upgrade + " upgrade capability"
	case "upgrade":
		return name + " service advertised the upgrade capability but rejected the " + upgrade + " negotiation request"
	case "tls":
		return name + " service accepted the protocol upgrade but the TLS handshake failed"
	case "certificate":
		return upgrade + " negotiated, but the served certificate is outside its validity window"
	case "hostname":
		return upgrade + " negotiated, but the served certificate does not match the requested host"
	case "trust":
		return upgrade + " negotiated, but the served certificate chain is not trusted by this host"
	default:
		return upgrade + " negotiation failed"
	}
}
