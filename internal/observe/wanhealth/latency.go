package wanhealth

import (
	"math"
	"regexp"
	"strconv"
	"time"
)

var (
	latencyPattern = regexp.MustCompile(`(?i)\btime[=<]([0-9]+(?:\.[0-9]+)?)\s*ms\b`)
	summaryPattern = regexp.MustCompile(`(?i)=\s*[0-9]+(?:\.[0-9]+)?/([0-9]+(?:\.[0-9]+)?)/[0-9]+(?:\.[0-9]+)?/[0-9]+(?:\.[0-9]+)?\s*ms\b`)
)

// ParseLatencyMS extracts a latency value from common Linux and BSD ping output.
func ParseLatencyMS(output []byte) (int, bool) {
	text := string(output)
	if match := latencyPattern.FindStringSubmatch(text); len(match) == 2 {
		return parseLatencyValue(match[1])
	}
	if match := summaryPattern.FindStringSubmatch(text); len(match) == 2 {
		return parseLatencyValue(match[1])
	}
	return 0, false
}

func parseLatencyValue(value string) (int, bool) {
	ms, err := strconv.ParseFloat(value, 64)
	if err != nil || ms < 0 {
		return 0, false
	}
	rounded := int(math.Round(ms))
	if rounded < 1 && ms > 0 {
		return 1, true
	}
	return rounded, true
}

func durationMS(value time.Duration) int {
	ms := int(math.Round(float64(value) / float64(time.Millisecond)))
	if ms < 1 {
		return 1
	}
	return ms
}

func downtimeSeconds(interval, timeout time.Duration) int {
	value := interval
	if value <= 0 {
		value = timeout
	}
	seconds := int(math.Ceil(value.Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}
