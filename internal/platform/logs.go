// Package platform exposes recent host logs as optional status metadata.
// journalctl is used on systemd hosts, syslog files cover conservative
// FreeBSD/OPNsense deployments, and failures stay reportable instead of
// stopping the stub.
package platform

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const defaultLogLines = 50

func (p hostPlatform) Logs(ctx context.Context, cfg LogConfig) ([]LogEntry, []error) {
	source := normalizedSource(cfg.Source)
	if source == SourceOff {
		return nil, nil
	}
	switch source {
	case LogSourceJournalctl:
		return p.journalctlLogs(ctx, cfg)
	case LogSourceSyslog:
		return syslogFileLogs(cfg)
	default:
		return nil, []error{fmt.Errorf("unsupported log source %q", source)}
	}
}

func (p hostPlatform) journalctlLogs(ctx context.Context, cfg LogConfig) ([]LogEntry, []error) {
	unit := strings.TrimSpace(cfg.Unit)
	if unit == "" {
		unit = p.cfg.JournalUnit
	}
	lines := cfg.Lines
	if lines <= 0 {
		lines = defaultLogLines
	}
	out, err := commandContext(ctx, p.cfg.CommandTimeout, "journalctl", "--output=json", "--no-pager", "--lines", fmt.Sprintf("%d", lines), "--unit", unit)
	if err != nil {
		return nil, []error{err}
	}
	entries, parseErrs := parseJournalJSONLines(string(out))
	return entries, parseErrs
}

func parseJournalJSONLines(output string) ([]LogEntry, []error) {
	var entries []LogEntry
	var errs []error
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			errs = append(errs, fmt.Errorf("parse journal json line: %w", err))
			continue
		}
		entries = append(entries, LogEntry{
			Time:     stringField(raw, "__REALTIME_TIMESTAMP"),
			Unit:     firstStringField(raw, "_SYSTEMD_UNIT", "UNIT", "SYSLOG_IDENTIFIER"),
			Priority: stringField(raw, "PRIORITY"),
			Message:  stringField(raw, "MESSAGE"),
			Raw:      line,
		})
	}
	if err := scanner.Err(); err != nil {
		errs = append(errs, err)
	}
	return entries, errs
}

func syslogFileLogs(cfg LogConfig) ([]LogEntry, []error) {
	path := strings.TrimSpace(cfg.Path)
	if path == "" {
		path = defaultSyslogPath
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, []error{fmt.Errorf("open syslog %s: %w", path, err)}
	}
	defer func() {
		_ = file.Close()
	}()

	lines := cfg.Lines
	if lines <= 0 {
		lines = defaultLogLines
	}
	var ring []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ring = append(ring, scanner.Text())
		if len(ring) > lines {
			copy(ring, ring[1:])
			ring = ring[:lines]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, []error{fmt.Errorf("read syslog %s: %w", path, err)}
	}
	entries := make([]LogEntry, 0, len(ring))
	for _, line := range ring {
		entries = append(entries, parseSyslogLine(line))
	}
	return entries, nil
}

func parseSyslogLine(line string) LogEntry {
	entry := LogEntry{Raw: line, Message: strings.TrimSpace(line)}
	fields := strings.Fields(line)
	if len(fields) >= 5 {
		entry.Time = strings.Join(fields[0:3], " ")
		entry.Unit = strings.TrimSuffix(fields[4], ":")
		entry.Message = strings.Join(fields[5:], " ")
	}
	return entry
}

func firstStringField(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringField(values, key); value != "" {
			return value
		}
	}
	return ""
}

func stringField(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		if len(typed) == 0 {
			return ""
		}
		return fmt.Sprint(typed[0])
	default:
		return fmt.Sprint(typed)
	}
}
