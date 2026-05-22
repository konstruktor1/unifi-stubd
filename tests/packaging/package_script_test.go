package packaging_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestPackageScriptKeepsOtherTargetArtifacts verifies packaging cleanup is
// scoped to the requested target.
func TestPackageScriptKeepsOtherTargetArtifacts(t *testing.T) {
	dist := t.TempDir()
	version := "9.9.9"
	release := "77"
	linuxTGZ := filepath.Join(dist, "packages", "unifi-stubd_"+version+"-"+release+"_linux_amd64.tar.gz")
	freebsdTGZ := filepath.Join(dist, "packages", "unifi-stubd_"+version+"-"+release+"_freebsd_amd64.tar.gz")

	runPackageScript(t, dist, version, release, "linux", "amd64")
	if _, err := os.Stat(linuxTGZ); err != nil {
		t.Fatalf("linux tgz was not created: %v", err)
	}

	runPackageScript(t, dist, version, release, "freebsd", "amd64")
	for _, path := range []string{linuxTGZ, freebsdTGZ} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected package artifact %s: %v", path, err)
		}
	}
}

// runPackageScript executes the package helper against a temporary dist tree.
func runPackageScript(t *testing.T, dist, version, release, goos, goarch string) {
	t.Helper()
	cmd := exec.Command("sh", "../../scripts/package.sh", "tgz")
	cmd.Env = append(os.Environ(),
		"DIST_DIR="+dist,
		"PKG_VERSION="+version,
		"PKG_RELEASE="+release,
		"PKG_GOOS="+goos,
		"PKG_GOARCH="+goarch,
		"BUILD_LDFLAGS=-s -w -X main.version="+version,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("package %s/%s failed: %v\n%s", goos, goarch, err, output)
	}
}
