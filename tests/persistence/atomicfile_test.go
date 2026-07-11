// Persistence tests verify that interrupted writes cannot expose partial state.
package persistence_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/atomicfile"
)

// TestInterruptedWriteLeavesCompleteAdoptionState simulates a crash before
// rename and verifies that the old state remains readable and replaceable.
func TestInterruptedWriteLeavesCompleteAdoptionState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adoption.env")
	want := adoption.Store{
		State:     adoption.StateConnected,
		InformURL: "http://192.0.2.10:8080/inform",
		AuthKey:   "test-auth-key-placeholder",
	}
	if err := adoption.SaveEnv(path, want); err != nil {
		t.Fatal(err)
	}

	interruptedPath := filepath.Join(dir, ".adoption.env.tmp-interrupted")
	if err := os.WriteFile(interruptedPath, []byte("STATE=provisioning\nAUTHKEY="), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := adoption.LoadEnv(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("adoption state = %+v, want %+v", got, want)
	}

	next := adoption.Store{
		State:      adoption.StateConnected,
		InformURL:  "http://192.0.2.11:8080/inform",
		AuthKey:    "replacement-auth-key-placeholder",
		CFGVersion: "cfg-replacement",
	}
	if err := adoption.SaveEnv(path, next); err != nil {
		t.Fatal(err)
	}
	got, err = adoption.LoadEnv(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != next {
		t.Fatalf("replacement adoption state = %+v, want %+v", got, next)
	}
}

// TestAtomicWriteReplacesContentAndMode verifies a committed update has the
// complete new content, requested permissions, and no owned temporary file.
func TestAtomicWriteReplacesContentAndMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "status.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := atomicfile.WriteFile(path, []byte("new\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new\n" {
		t.Fatalf("content = %q, want new content", data)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("mode = %#o, want 0600", got)
	}
	assertNoOwnedTempFiles(t, filepath.Dir(path), filepath.Base(path))
}

// TestAtomicWriteCleansTempFileWhenRenameFails verifies a failed commit keeps
// the target intact and cleans up the temporary file created for that attempt.
func TestAtomicWriteCleansTempFileWhenRenameFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "status.json")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}

	err := atomicfile.WriteFile(path, []byte("incomplete\n"), 0o600)
	if err == nil || !strings.Contains(err.Error(), "replace") {
		t.Fatalf("write error = %v, want replace failure", err)
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatal(statErr)
	}
	if !info.IsDir() {
		t.Fatal("failed replacement changed the existing target")
	}
	assertNoOwnedTempFiles(t, dir, filepath.Base(path))
}

func assertNoOwnedTempFiles(t *testing.T, dir, base string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "."+base+".tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary files remain: %v", matches)
	}
}
