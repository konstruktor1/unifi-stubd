// WAN health tests cover parsing and active probe result shaping without
// executing the host ping binary.
package observe_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/observe/wanhealth"
)

const testPingHost = "1.1.1.1"

type fakePingRunner struct {
	output  []byte
	elapsed time.Duration
	err     error
}

func (runner fakePingRunner) Ping(context.Context, string, time.Duration) ([]byte, time.Duration, error) {
	return runner.output, runner.elapsed, runner.err
}

func TestWANHealthParseLatencyFromLinuxAndBSDPing(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   int
	}{
		{
			name:   "linux",
			output: "64 bytes from " + testPingHost + ": icmp_seq=1 ttl=58 time=7.42 ms\n",
			want:   7,
		},
		{
			name:   "bsd",
			output: "64 bytes from 1.1.1.1: icmp_seq=0 ttl=58 time=0.741 ms\n",
			want:   1,
		},
		{
			name:   "summary",
			output: "round-trip min/avg/max/stddev = 6.211/6.733/7.255/0.522 ms\n",
			want:   7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := wanhealth.ParseLatencyMS([]byte(tt.output))
			if !ok {
				t.Fatal("ParseLatencyMS ok = false")
			}
			if got != tt.want {
				t.Fatalf("ParseLatencyMS = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestWANHealthMeasureShapesConnectedResult(t *testing.T) {
	results := wanhealth.MeasureWithRunner(context.Background(), wanhealth.Config{
		Source:   wanhealth.SourcePing,
		Interval: 10 * time.Second,
		Timeout:  time.Second,
		Targets: []wanhealth.Target{
			{Port: 3, Host: testPingHost},
		},
	}, fakePingRunner{
		output:  []byte("64 bytes from " + testPingHost + ": icmp_seq=1 ttl=58 time=7.42 ms\n"),
		elapsed: 8 * time.Millisecond,
	})
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	result := results[0]
	if result.Port != 3 || result.Host != testPingHost || !result.Connected {
		t.Fatalf("result identity = %+v", result)
	}
	if result.LatencyMS != 7 || result.UptimePercent != 100 || result.DowntimeSeconds != 0 || result.LastError != "" {
		t.Fatalf("result health = %+v", result)
	}
}

func TestWANHealthMeasureShapesFailedResult(t *testing.T) {
	results := wanhealth.MeasureWithRunner(context.Background(), wanhealth.Config{
		Source:   wanhealth.SourcePing,
		Interval: 10 * time.Second,
		Timeout:  time.Second,
		Targets: []wanhealth.Target{
			{Port: 3, Host: "unreachable.example"},
		},
	}, fakePingRunner{err: errors.New("exit status 1")})
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	result := results[0]
	if result.Connected || result.UptimePercent != 0 || result.DowntimeSeconds != 10 {
		t.Fatalf("failed result = %+v", result)
	}
	if result.LastError == "" {
		t.Fatal("LastError is empty")
	}
}

func TestWANHealthMeasureIgnoresOffSource(t *testing.T) {
	results := wanhealth.MeasureWithRunner(context.Background(), wanhealth.Config{
		Source: wanhealth.SourceOff,
		Targets: []wanhealth.Target{
			{Port: 3, Host: testPingHost},
		},
	}, fakePingRunner{})
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}
