package core

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const firewallOutputLimit = 64 * 1024

type boundedCommandResult struct {
	Output    string
	Truncated bool
}

type cappedBuffer struct {
	mu        sync.Mutex
	buf       bytes.Buffer
	max       int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	original := len(p)
	remaining := b.max - b.buf.Len()
	if remaining <= 0 {
		if original > 0 {
			b.truncated = true
		}
		return original, nil
	}
	if len(p) > remaining {
		p = p[:remaining]
		b.truncated = true
	}
	_, _ = b.buf.Write(p)
	return original, nil
}

func (b *cappedBuffer) result() boundedCommandResult {
	b.mu.Lock()
	defer b.mu.Unlock()
	return boundedCommandResult{Output: b.buf.String(), Truncated: b.truncated}
}

func runBoundedCommand(ctx context.Context, maxBytes int, name string, args ...string) (boundedCommandResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out := &cappedBuffer{max: maxBytes}
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	return out.result(), err
}

var nftRulesetLookup = func(ctx context.Context) (boundedCommandResult, error) {
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	command := resolveCommand("nft", "/usr/sbin/nft", "/sbin/nft")
	return runBoundedCommand(lookupCtx, firewallOutputLimit, command, "-nn", "list", "ruleset")
}

func resolveCommand(name string, candidates ...string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return path
		}
	}
	return name
}

func firewallCheck(ctx context.Context, port string) Check {
	result, err := nftRulesetLookup(ctx)
	ruleset := strings.TrimSpace(result.Output)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return Check{Name: "firewall", Status: "unknown", Evidence: "nft command is unavailable"}
		}
		if ruleset != "" {
			return Check{Name: "firewall", Status: "unknown", Evidence: "nftables active-policy read unavailable: " + boundedEvidence(ruleset, 512)}
		}
		return Check{Name: "firewall", Status: "unknown", Evidence: "nftables active-policy read unavailable: " + boundedEvidence(err.Error(), 384)}
	}

	if ruleset == "" {
		return Check{Name: "firewall", Status: "pass", Evidence: "nftables ruleset is empty"}
	}

	evidence := summarizeNFTRuleset(ruleset, port)
	if result.Truncated {
		evidence += "; ruleset scan truncated at 64 KiB"
	}
	return Check{Name: "firewall", Status: "unknown", Evidence: boundedEvidence(evidence, 1024)}
}

func summarizeNFTRuleset(ruleset, port string) string {
	lines := strings.Split(ruleset, "\n")
	candidates := make([]string, 0, 4)
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		basePolicy := (strings.Contains(lower, "hook input") || strings.Contains(lower, "hook output")) && strings.Contains(lower, "policy ")
		portRule := lineMentionsTCPDPort(lower, port) && containsFirewallVerdict(lower)
		if !basePolicy && !portRule {
			continue
		}
		candidates = append(candidates, boundedEvidence(line, 240))
		if len(candidates) == 4 {
			break
		}
	}

	if len(candidates) == 0 {
		return "nftables ruleset present; no direct TCP/" + port + " rule or input/output base-policy candidate found in bounded scan"
	}
	return "nftables candidate evidence (not a proven verdict): " + strings.Join(candidates, " | ")
}

func containsFirewallVerdict(line string) bool {
	fields := strings.FieldsFunc(line, func(r rune) bool {
		switch r {
		case ' ', '\t', ';', ',', '{', '}':
			return true
		default:
			return false
		}
	})
	for _, field := range fields {
		switch field {
		case "accept", "drop", "reject":
			return true
		}
	}
	return false
}

func lineMentionsTCPDPort(line, port string) bool {
	needle := "tcp dport"
	for {
		idx := strings.Index(line, needle)
		if idx < 0 {
			return false
		}
		rest := strings.TrimSpace(line[idx+len(needle):])
		if rest == "" {
			return false
		}
		if rest[0] == '{' {
			end := strings.IndexByte(rest, '}')
			if end < 0 {
				return false
			}
			if tokenListContains(rest[1:end], port) {
				return true
			}
			line = rest[end+1:]
			continue
		}
		field := strings.Fields(rest)
		if len(field) == 0 {
			return false
		}
		value := strings.Trim(field[0], ",;{}")
		if value == port {
			return true
		}
		line = rest[len(field[0]):]
	}
}

func tokenListContains(value, target string) bool {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == ' ' || r == '\t' || r == ','
	})
	for _, field := range fields {
		if strings.TrimSpace(field) == target {
			return true
		}
	}
	return false
}
