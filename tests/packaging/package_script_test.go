package packaging_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// TestFreeBSDPkgPlistUsesPrefixRelativePaths verifies native FreeBSD pkg
// packages follow plist semantics. Absolute /usr/local plist entries caused
// pkg 2.3.1 to crash when migrating a host that already had tarball-installed,
// unregistered files in place.
func TestFreeBSDPkgPlistUsesPrefixRelativePaths(t *testing.T) {
	data, err := os.ReadFile("../../scripts/package-freebsd-pkg-repos.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	start := strings.Index(script, "cat >\"$stage/plist\" <<'EOF'\n")
	if start == -1 {
		t.Fatal("FreeBSD pkg plist heredoc not found")
	}
	plist := script[start:]
	end := strings.Index(plist, "\nEOF\n")
	if end == -1 {
		t.Fatal("FreeBSD pkg plist heredoc end not found")
	}
	plist = plist[:end]
	for _, forbidden := range []string{
		"/usr/local/bin/unifi-stubd",
		"/usr/local/etc/rc.d/unifi-stubd",
		"/usr/local/etc/unifi-stubd/config.yaml",
		"/usr/local/share/doc/unifi-stubd/LICENSE",
	} {
		if strings.Contains(plist, forbidden) {
			t.Fatalf("FreeBSD pkg plist contains prefix-absolute path %q", forbidden)
		}
	}
	for _, required := range []string{
		"bin/unifi-stubd",
		"etc/rc.d/unifi-stubd",
		"etc/unifi-stubd/config.yaml",
		"share/doc/unifi-stubd/LICENSE",
		"@dir /var/db/unifi-stubd",
	} {
		if !strings.Contains(plist, required) {
			t.Fatalf("FreeBSD pkg plist does not contain %q", required)
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
