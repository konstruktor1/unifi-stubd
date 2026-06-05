// Command inform-proxy records lab inform traffic while forwarding it upstream.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

func main() {
	var listen string
	var target string
	var captureDir string
	flag.StringVar(&listen, "listen", "0.0.0.0:8080", "listen address")
	flag.StringVar(&target, "target", "http://unifi-controller:8080", "upstream controller base URL")
	flag.StringVar(&captureDir, "capture-dir", envDefault("MITM_CAPTURE_DIR", "/captures"), "capture output directory")
	flag.Parse()

	proxy, err := newProxy(target, captureDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("writing inform captures to %s", captureDir)
	log.Printf("forwarding inform traffic to %s", target)
	if err := http.ListenAndServe(listen, proxy); err != nil {
		log.Fatal(err)
	}
}

type proxy struct {
	target   *url.URL
	captures string
	events   string
	client   *http.Client
	mu       sync.Mutex
	sequence uint64
}

func newProxy(target, captureDir string) (*proxy, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parse target URL: %w", err)
	}
	if err := os.MkdirAll(captureDir, 0o755); err != nil {
		return nil, fmt.Errorf("create capture directory: %w", err)
	}
	return &proxy{
		target:   targetURL,
		captures: captureDir,
		events:   filepath.Join(captureDir, "events.jsonl"),
		client:   &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := p.handle(w, r); err != nil {
		log.Printf("proxy request failed: %v", err)
		http.Error(w, "proxy request failed", http.StatusBadGateway)
	}
}

func (p *proxy) handle(w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read request body: %w", err)
	}
	if err := r.Body.Close(); err != nil {
		return fmt.Errorf("close request body: %w", err)
	}

	eventID := p.eventID()
	if r.URL.Path == "/inform" {
		if err := p.captureRequest(eventID, r, body); err != nil {
			return err
		}
	}

	upstreamURL := p.upstreamURL(r.URL)
	req, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create upstream request: %w", err)
	}
	req.Header = r.Header.Clone()
	req.Host = r.Host
	req.ContentLength = int64(len(body))

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("forward request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read upstream response: %w", err)
	}
	if r.URL.Path == "/inform" {
		if err := p.captureResponse(eventID, resp, responseBody); err != nil {
			return err
		}
	}

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(responseBody); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	return nil
}

func (p *proxy) upstreamURL(requestURL *url.URL) string {
	next := *p.target
	next.Path = joinPath(p.target.Path, requestURL.Path)
	next.RawQuery = requestURL.RawQuery
	return next.String()
}

func (p *proxy) captureRequest(eventID string, r *http.Request, body []byte) error {
	bodyPath := filepath.Join(p.captures, eventID+"-request.bin")
	if err := os.WriteFile(bodyPath, body, 0o644); err != nil {
		return fmt.Errorf("write request body: %w", err)
	}
	event := map[string]any{
		"event":        "request",
		"id":           eventID,
		"timestamp":    timestamp(),
		"client":       r.RemoteAddr,
		"method":       r.Method,
		"url":          requestURL(r),
		"http_version": r.Proto,
		"headers":      headerMap(r.Header),
		"body_path":    bodyPath,
		"body_bytes":   len(body),
		"body_sha256":  sha256Hex(body),
		"tnbu":         tnbuSummary(body),
	}
	return p.writeEvent(event)
}

func (p *proxy) captureResponse(eventID string, resp *http.Response, body []byte) error {
	bodyPath := filepath.Join(p.captures, eventID+"-response.bin")
	if err := os.WriteFile(bodyPath, body, 0o644); err != nil {
		return fmt.Errorf("write response body: %w", err)
	}
	event := map[string]any{
		"event":       "response",
		"id":          eventID,
		"timestamp":   timestamp(),
		"status_code": resp.StatusCode,
		"reason":      http.StatusText(resp.StatusCode),
		"headers":     headerMap(resp.Header),
		"body_path":   bodyPath,
		"body_bytes":  len(body),
		"body_sha256": sha256Hex(body),
		"tnbu":        tnbuSummary(body),
	}
	return p.writeEvent(event)
}

func (p *proxy) writeEvent(event map[string]any) error {
	encoded, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode event: %w", err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	file, err := os.OpenFile(p.events, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open events file: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	log.Printf("inform %s id=%s bytes=%v tnbu=%v", event["event"], event["id"], event["body_bytes"], event["tnbu"])
	return nil
}

func (p *proxy) eventID() string {
	seq := atomic.AddUint64(&p.sequence, 1)
	return fmt.Sprintf("%d-%08x", time.Now().UnixMilli(), seq)
}

func tnbuSummary(data []byte) map[string]any {
	if len(data) < 40 || string(data[:4]) != inform.Magic {
		return map[string]any{"present": false}
	}
	return map[string]any{
		"present":         true,
		"packet_version":  binary.BigEndian.Uint32(data[4:8]),
		"mac":             net.HardwareAddr(data[8:14]).String(),
		"flags":           binary.BigEndian.Uint16(data[14:16]),
		"iv_hex":          hex.EncodeToString(data[16:32]),
		"payload_version": binary.BigEndian.Uint32(data[32:36]),
		"payload_bytes":   binary.BigEndian.Uint32(data[36:40]),
	}
}

func requestURL(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.String()
	}
	return "http://" + r.Host + r.URL.RequestURI()
}

func headerMap(headers http.Header) map[string]string {
	out := make(map[string]string, len(headers))
	for key, values := range headers {
		out[key] = strings.Join(values, ", ")
	}
	return out
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func joinPath(basePath, requestPath string) string {
	if basePath == "" || basePath == "/" {
		return requestPath
	}
	if requestPath == "" || requestPath == "/" {
		return basePath
	}
	return strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(requestPath, "/")
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func timestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func envDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
