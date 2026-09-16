package core

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

const rebootStorySchemaVersion = 4

func collectBootIdentity() (string, time.Time) {
	bootID := ""
	if b, err := os.ReadFile("/proc/sys/kernel/random/boot_id"); err == nil {
		bootID = strings.TrimSpace(string(b))
		if len(bootID) > 128 {
			bootID = bootID[:128]
		}
	}

	startedAt := time.Time{}
	if f, err := os.Open("/proc/stat"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if parsed, ok := parseProcStatBootStartedAt(scanner.Text()); ok {
				startedAt = parsed
				break
			}
		}
	}
	return bootID, startedAt
}

func parseProcStatBootStartedAt(line string) (time.Time, bool) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) != 2 || fields[0] != "btime" {
		return time.Time{}, false
	}
	seconds, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil || seconds <= 0 {
		return time.Time{}, false
	}
	return time.Unix(seconds, 0).UTC(), true
}
