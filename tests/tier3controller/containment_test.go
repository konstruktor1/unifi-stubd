package tier3controller_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Invalid handshake namespaces must fail before Docker or result writes. The
// target stays inside this checkout so the regression cannot touch another lab.
func TestControllerGateRejectsSymlinkNamespacesBeforeDocker(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if root != "/workspace" {
		t.Skip("run through the canonical Make/Build Runner workspace")
	}
	dist := filepath.Join(root, "dist")
	if info, err := os.Lstat(dist); err == nil {
		if !info.IsDir() {
			t.Skip("preserve the existing output root; use a fresh CI checkout")
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, symlinkRole := range []string{"", "tier3-controller-input", "tier3-controller"} {
		name := symlinkRole
		if name == "" {
			name = "valid-namespaces"
		}
		t.Run(name, func(t *testing.T) {
			inputRole := filepath.Join(root, "dist", "tier3-controller-input")
			outputRole := filepath.Join(root, "dist", "tier3-controller")
			for _, role := range []string{inputRole, outputRole} {
				if _, err := os.Lstat(role); !os.IsNotExist(err) {
					if err != nil {
						t.Fatal(err)
					}
					t.Skip("preserve existing controller handshake state; use a fresh CI checkout")
				}
			}
			if err := os.Mkdir(dist, 0755); err == nil {
				t.Cleanup(func() {
					if err := os.Remove(dist); err != nil {
						t.Errorf("remove owned empty output root: %v", err)
					}
				})
			} else if !os.IsExist(err) {
				t.Fatal(err)
			}
			work := filepath.Join(root, "tests", ".work")
			if err := os.Mkdir(work, 0700); err == nil {
				t.Cleanup(func() {
					if err := os.Remove(work); err != nil {
						t.Errorf("remove owned empty verification root: %v", err)
					}
				})
			} else if !os.IsExist(err) {
				t.Fatal(err)
			}
			if physical, err := filepath.EvalSymlinks(work); err != nil || physical != work {
				t.Fatalf("verification root is not physical: %s, %v", physical, err)
			}
			fixture, err := os.MkdirTemp(work, "tier3guard-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.RemoveAll(fixture); err != nil {
					t.Errorf("remove owned fixture: %v", err)
				}
			})
			outsideRole := filepath.Join(fixture, "outside-role")
			if err := os.Mkdir(outsideRole, 0700); err != nil {
				t.Fatal(err)
			}
			sentinel := filepath.Join(outsideRole, "sentinel")
			if err := os.WriteFile(sentinel, []byte("preserve\n"), 0600); err != nil {
				t.Fatal(err)
			}
			for _, role := range []string{inputRole, outputRole} {
				var err error
				if filepath.Base(role) == symlinkRole {
					err = os.Symlink(outsideRole, role)
				} else {
					err = os.Mkdir(role, 0700)
				}
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := os.RemoveAll(role); err != nil {
						t.Errorf("remove owned handshake namespace: %v", err)
					}
				})
			}
			runID := filepath.Base(fixture)
			manifest, err := os.ReadFile("../fixtures/tier3controller/valid.env")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(inputRole, runID+".env"),
				[]byte(strings.ReplaceAll(string(manifest), "@RUN_ID@", runID)), 0600); err != nil {
				t.Fatal(err)
			}
			stub, err := os.ReadFile("../fixtures/tier3controller/docker.sh")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(fixture, "docker"), stub, 0755); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(fixture, "docker-calls")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, "bash", "lab/stub/scripts/run-tier3-controller-gate.sh")
			command.Dir = root
			command.Env = append(os.Environ(),
				"PATH="+fixture+string(os.PathListSeparator)+os.Getenv("PATH"),
				"UNIFI_STUB_TIER3_GATE_MANIFEST=dist/tier3-controller-input/"+runID+".env",
				"UNIFI_STUB_TEST_DOCKER_LOG="+marker)
			output, err := command.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatalf("expected the gate to terminate with an error: %v\n%s", err, output)
			}
			if symlinkRole == "" {
				if _, err := os.Lstat(marker); err != nil || exit.ExitCode() != 97 {
					t.Fatalf("valid fixture did not reach the isolated Docker boundary: %v\n%s", err, output)
				}
				return
			}
			if exit.ExitCode() != 2 {
				t.Fatalf("invalid namespace did not fail at the input boundary: %v\n%s", err, output)
			}
			for _, forbidden := range []string{marker, filepath.Join(outputRole, runID)} {
				if _, err := os.Lstat(forbidden); !os.IsNotExist(err) {
					t.Fatalf("invalid namespace reached Docker or result writes: %s, %v\n%s", forbidden, err, output)
				}
			}
			if data, err := os.ReadFile(sentinel); err != nil || string(data) != "preserve\n" {
				t.Fatalf("sentinel changed: %q, %v", data, err)
			}
		})
	}
}
