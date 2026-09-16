package core

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	workbenchFileBytesLimit = int64(16 * 1024 * 1024 * 1024)
	workbenchDNSRecordLimit = 128
	workbenchHTTPHopLimit   = 10
	workbenchPEMBytesLimit  = 256 * 1024
)

type FileInspection struct {
	Path              string    `json:"path"`
	Size              int64     `json:"size"`
	Mode              string    `json:"mode"`
	Permissions       string    `json:"permissions"`
	ModifiedAt        time.Time `json:"modified_at"`
	UID               uint32    `json:"uid,omitempty"`
	GID               uint32    `json:"gid,omitempty"`
	Owner             string    `json:"owner,omitempty"`
	Group             string    `json:"group,omitempty"`
	SHA256            string    `json:"sha256"`
	SHA512            string    `json:"sha512"`
	ExpectedAlgorithm string    `json:"expected_algorithm,omitempty"`
	ExpectedChecksum  string    `json:"expected_checksum,omitempty"`
	ChecksumStatus    string    `json:"checksum_status,omitempty"`
}

type FileComparison struct {
	Left       *FileInspection `json:"left"`
	Right      *FileInspection `json:"right"`
	SameSize   bool            `json:"same_size"`
	SameSHA256 bool            `json:"same_sha256"`
	Conclusion string          `json:"conclusion"`
}

type DNSRecord struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type DNSLookupEvidence struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type DNSInspection struct {
	Name     string              `json:"name"`
	Resolver string              `json:"resolver"`
	Records  []DNSRecord         `json:"records"`
	Lookups  []DNSLookupEvidence `json:"lookups"`
}

type HTTPHop struct {
	URL         string `json:"url"`
	StatusCode  int    `json:"status_code"`
	Status      string `json:"status"`
	Location    string `json:"location,omitempty"`
	Server      string `json:"server,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type HTTPInspection struct {
	URL         string    `json:"url"`
	Hops        []HTTPHop `json:"hops"`
	FinalURL    string    `json:"final_url,omitempty"`
	FinalStatus int       `json:"final_status,omitempty"`
	Redirected  bool      `json:"redirected"`
	Truncated   bool      `json:"truncated,omitempty"`
	Conclusion  string    `json:"conclusion"`
}

type CertificateFileInspection struct {
	Path               string               `json:"path"`
	Certificate        *CertificateEvidence `json:"certificate"`
	PublicKeyAlgorithm string               `json:"public_key_algorithm,omitempty"`
	SignatureAlgorithm string               `json:"signature_algorithm,omitempty"`
	IsCA               bool                 `json:"is_ca"`
}

type CertificateFileComparison struct {
	Local      *CertificateFileInspection `json:"local"`
	Target     string                     `json:"target"`
	Served     *TLSEvidence               `json:"served,omitempty"`
	Status     string                     `json:"status"`
	Conclusion string                     `json:"conclusion"`
}

func InspectFile(ctx context.Context, path, expected string) (*FileInspection, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("file path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("path must be a regular file")
	}
	if info.Size() > workbenchFileBytesLimit {
		return nil, fmt.Errorf("file is larger than the Workbench limit of %d bytes", workbenchFileBytesLimit)
	}

	sha256Value, sha512Value, err := hashFile(ctx, absolute)
	if err != nil {
		return nil, err
	}
	result := &FileInspection{
		Path:        absolute,
		Size:        info.Size(),
		Mode:        info.Mode().String(),
		Permissions: fmt.Sprintf("%#o", info.Mode().Perm()),
		ModifiedAt:  info.ModTime().UTC(),
		SHA256:      sha256Value,
		SHA512:      sha512Value,
	}
	populateOwnership(result, info)

	if strings.TrimSpace(expected) != "" {
		algorithm, normalized, err := normalizeExpectedChecksum(expected)
		if err != nil {
			return nil, err
		}
		result.ExpectedAlgorithm = algorithm
		result.ExpectedChecksum = normalized
		actual := result.SHA256
		if algorithm == "sha512" {
			actual = result.SHA512
		}
		if strings.EqualFold(actual, normalized) {
			result.ChecksumStatus = "match"
		} else {
			result.ChecksumStatus = "mismatch"
		}
	}
	return result, nil
}

func CompareFiles(ctx context.Context, leftPath, rightPath string) (*FileComparison, error) {
	left, err := InspectFile(ctx, leftPath, "")
	if err != nil {
		return nil, fmt.Errorf("left file: %w", err)
	}
	right, err := InspectFile(ctx, rightPath, "")
	if err != nil {
		return nil, fmt.Errorf("right file: %w", err)
	}
	comparison := &FileComparison{
		Left:       left,
		Right:      right,
		SameSize:   left.Size == right.Size,
		SameSHA256: strings.EqualFold(left.SHA256, right.SHA256),
	}
	if comparison.SameSHA256 {
		comparison.Conclusion = "files are byte-for-byte identical by SHA-256 fingerprint"
	} else {
		comparison.Conclusion = "files have different SHA-256 fingerprints"
	}
	return comparison, nil
}

func hashFile(ctx context.Context, path string) (string, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	h256 := sha256.New()
	h512 := sha512.New()
	writer := io.MultiWriter(h256, h512)
	buffer := make([]byte, 1024*1024)
	for {
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		default:
		}
		n, readErr := f.Read(buffer)
		if n > 0 {
			if _, err := writer.Write(buffer[:n]); err != nil {
				return "", "", err
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return "", "", readErr
		}
	}
	return strings.ToUpper(hex.EncodeToString(h256.Sum(nil))), strings.ToUpper(hex.EncodeToString(h512.Sum(nil))), nil
}

func normalizeExpectedChecksum(value string) (string, string, error) {
	value = strings.TrimSpace(value)
	algorithm := ""
	lower := strings.ToLower(value)
	for _, prefix := range []string{"sha256:", "sha256=", "sha512:", "sha512="} {
		if strings.HasPrefix(lower, prefix) {
			algorithm = strings.TrimSuffix(prefix, ":")
			algorithm = strings.TrimSuffix(algorithm, "=")
			value = strings.TrimSpace(value[len(prefix):])
			break
		}
	}
	value = strings.ReplaceAll(value, ":", "")
	value = strings.ReplaceAll(value, " ", "")
	if algorithm == "" {
		switch len(value) {
		case 64:
			algorithm = "sha256"
		case 128:
			algorithm = "sha512"
		default:
			return "", "", errors.New("expected checksum must be 64 hex characters for SHA-256 or 128 for SHA-512")
		}
	}
	expectedLength := 64
	if algorithm == "sha512" {
		expectedLength = 128
	}
	if len(value) != expectedLength {
		return "", "", fmt.Errorf("%s checksum must be %d hex characters", strings.ToUpper(algorithm), expectedLength)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return "", "", errors.New("expected checksum must contain only hexadecimal characters")
	}
	return algorithm, strings.ToUpper(value), nil
}

func populateOwnership(result *FileInspection, info os.FileInfo) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return
	}
	result.UID = stat.Uid
	result.GID = stat.Gid
	if u, err := user.LookupId(strconv.FormatUint(uint64(stat.Uid), 10)); err == nil {
		result.Owner = u.Username
	}
	if g, err := user.LookupGroupId(strconv.FormatUint(uint64(stat.Gid), 10)); err == nil {
		result.Group = g.Name
	}
}

func InspectDNS(ctx context.Context, name string) (*DNSInspection, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("DNS name or IP is required")
	}
	lookupCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	result := &DNSInspection{Name: name, Resolver: "system resolver"}
	resolver := net.DefaultResolver
	addStatus := func(kind string, err error) {
		status := DNSLookupEvidence{Type: kind, Status: "pass"}
		if err != nil {
			status.Status = "unknown"
			status.Error = boundedEvidence(err.Error(), 384)
		}
		result.Lookups = append(result.Lookups, status)
	}
	addRecord := func(kind, value string) {
		value = strings.TrimSpace(value)
		if value == "" || len(result.Records) >= workbenchDNSRecordLimit {
			return
		}
		result.Records = append(result.Records, DNSRecord{Type: kind, Value: boundedEvidence(value, 1024)})
	}

	if ip := net.ParseIP(name); ip != nil {
		names, err := resolver.LookupAddr(lookupCtx, name)
		addStatus("PTR", err)
		for _, value := range names {
			addRecord("PTR", value)
		}
	} else {
		ips, err := resolver.LookupIP(lookupCtx, "ip", name)
		addStatus("A/AAAA", err)
		for _, ip := range ips {
			kind := "AAAA"
			if ip.To4() != nil {
				kind = "A"
			}
			addRecord(kind, ip.String())
		}

		cname, err := resolver.LookupCNAME(lookupCtx, name)
		addStatus("CNAME", err)
		if err == nil && !sameDNSName(cname, name) {
			addRecord("CNAME", cname)
		}

		mx, err := resolver.LookupMX(lookupCtx, name)
		addStatus("MX", err)
		for _, value := range mx {
			addRecord("MX", fmt.Sprintf("%d %s", value.Pref, value.Host))
		}

		ns, err := resolver.LookupNS(lookupCtx, name)
		addStatus("NS", err)
		for _, value := range ns {
			addRecord("NS", value.Host)
		}

		txt, err := resolver.LookupTXT(lookupCtx, name)
		addStatus("TXT", err)
		for _, value := range txt {
			addRecord("TXT", value)
		}
	}

	sort.Slice(result.Records, func(i, j int) bool {
		if result.Records[i].Type == result.Records[j].Type {
			return result.Records[i].Value < result.Records[j].Value
		}
		return result.Records[i].Type < result.Records[j].Type
	})
	result.Records = uniqueDNSRecords(result.Records)
	return result, nil
}

func sameDNSName(a, b string) bool {
	normalize := func(value string) string {
		return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), ".")
	}
	return normalize(a) == normalize(b)
}

func uniqueDNSRecords(records []DNSRecord) []DNSRecord {
	if len(records) == 0 {
		return records
	}
	out := records[:0]
	for i, record := range records {
		if i == 0 || record != records[i-1] {
			out = append(out, record)
		}
	}
	return out
}

func InspectHTTP(ctx context.Context, rawURL string) (*HTTPInspection, error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("a complete http:// or https:// URL is required")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("only http:// and https:// URLs are supported")
	}
	if parsed.User != nil {
		return nil, errors.New("URLs containing credentials are not accepted")
	}

	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	result := &HTTPInspection{URL: parsed.String()}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           (&net.Dialer{Timeout: 4 * time.Second}).DialContext,
		TLSHandshakeTimeout:   4 * time.Second,
		ResponseHeaderTimeout: 6 * time.Second,
		DisableCompression:    true,
	}
	defer transport.CloseIdleConnections()
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.Response != nil {
				appendHTTPHop(result, req.Response)
			}
			if len(via) >= workbenchHTTPHopLimit {
				result.Truncated = true
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	req, err := http.NewRequestWithContext(probeCtx, http.MethodHead, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "HostSleuth-Workbench/1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	appendHTTPHop(result, resp)
	result.FinalURL = resp.Request.URL.String()
	result.FinalStatus = resp.StatusCode
	result.Redirected = len(result.Hops) > 1
	if result.Truncated {
		result.Conclusion = fmt.Sprintf("redirect inspection stopped after %d hops", len(result.Hops))
	} else if result.Redirected {
		result.Conclusion = fmt.Sprintf("HEAD request followed %d redirect(s) and finished with HTTP %d", len(result.Hops)-1, resp.StatusCode)
	} else {
		result.Conclusion = fmt.Sprintf("HEAD request returned HTTP %d without a redirect", resp.StatusCode)
	}
	return result, nil
}

func appendHTTPHop(result *HTTPInspection, resp *http.Response) {
	if result == nil || resp == nil || resp.Request == nil {
		return
	}
	hop := HTTPHop{
		URL:         resp.Request.URL.String(),
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Location:    boundedEvidence(resp.Header.Get("Location"), 1024),
		Server:      boundedEvidence(resp.Header.Get("Server"), 256),
		ContentType: boundedEvidence(resp.Header.Get("Content-Type"), 256),
	}
	if len(result.Hops) > 0 {
		last := result.Hops[len(result.Hops)-1]
		if last.URL == hop.URL && last.StatusCode == hop.StatusCode && last.Location == hop.Location {
			return
		}
	}
	if len(result.Hops) < workbenchHTTPHopLimit+1 {
		result.Hops = append(result.Hops, hop)
	}
}

func InspectCertificateFile(path string) (*CertificateFileInspection, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, errors.New("certificate path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	cert, err := readCertificatePEMOnly(absolute)
	if err != nil {
		return nil, err
	}
	return &CertificateFileInspection{
		Path:               absolute,
		Certificate:        certificateEvidence(cert, time.Now().UTC()),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		IsCA:               cert.IsCA,
	}, nil
}

func CompareCertificateFileToServed(ctx context.Context, path, target string) (*CertificateFileComparison, error) {
	local, err := InspectCertificateFile(path)
	if err != nil {
		return nil, err
	}
	target = strings.TrimSpace(target)
	host, _, err := net.SplitHostPort(target)
	if err != nil || strings.TrimSpace(host) == "" {
		return nil, errors.New("target must be host:port")
	}
	served := probeTLS(ctx, target, host)
	comparison := &CertificateFileComparison{Local: local, Target: target, Served: served, Status: "unknown"}
	if served == nil || served.HandshakeStatus != "pass" || served.Certificate == nil {
		comparison.Conclusion = "served certificate could not be captured for comparison"
		return comparison, nil
	}
	if sameCertificateFingerprint(local.Certificate, served.Certificate) {
		comparison.Status = "match"
		comparison.Conclusion = "certificate file exactly matches the certificate served by the endpoint"
	} else {
		comparison.Status = "mismatch"
		comparison.Conclusion = "certificate file fingerprint differs from the certificate served by the endpoint"
	}
	return comparison, nil
}

// readCertificatePEMOnly stops as soon as the first CERTIFICATE block is read.
// If a private-key block appears before a certificate, it refuses the file
// without reading the private-key payload. This keeps Workbench certificate
// inspection on the same public-certificate boundary as M6.
func readCertificatePEMOnly(path string) (*x509.Certificate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(io.LimitReader(f, int64(workbenchPEMBytesLimit)+1))
	scanner.Buffer(make([]byte, 4096), 64*1024)
	var pemText strings.Builder
	capturing := false
	readBytes := 0
	for scanner.Scan() {
		line := scanner.Text()
		readBytes += len(line) + 1
		if readBytes > workbenchPEMBytesLimit {
			return nil, errors.New("certificate PEM exceeds Workbench size limit")
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-----BEGIN ") && strings.Contains(trimmed, "PRIVATE KEY-----") {
			return nil, errors.New("private-key PEM blocks are not accepted by certificate inspection")
		}
		if trimmed == "-----BEGIN CERTIFICATE-----" {
			capturing = true
			pemText.Reset()
		}
		if capturing {
			pemText.WriteString(line)
			pemText.WriteByte('\n')
		}
		if capturing && trimmed == "-----END CERTIFICATE-----" {
			block, _ := pem.Decode([]byte(pemText.String()))
			if block == nil || block.Type != "CERTIFICATE" {
				return nil, errors.New("invalid certificate PEM block")
			}
			return x509.ParseCertificate(block.Bytes)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, errors.New("no certificate PEM block found")
}
