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

// TestFreeBSDPkgManifestUsesSimpleChecksumConfigEntries verifies native
// FreeBSD pkg packages avoid plist-derived owner/mode/mtime file objects while
// still marking runtime config as package-managed config. pkg-create normalizes
// direct manifest input back into file object entries, and pkg 2.3.1 on
// OPNsense 26.1 crashed when migrating a host that already had tarball-
// installed, unregistered files in place. The package builder must repack the
// generated archive with checksum-only +MANIFEST file entries.
func TestFreeBSDPkgManifestUsesSimpleChecksumConfigEntries(t *testing.T) {
	data, err := os.ReadFile("../../scripts/package-freebsd-pkg-repos.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(data)
	if strings.Contains(script, "cat >\"$stage/plist\"") ||
		strings.Contains(script, " -p \"$abi_dir/plist\"") {
		t.Fatal("FreeBSD pkg builder should use a direct manifest, not a plist")
	}
	if !strings.Contains(script, "pkg create -f txz -r \"$abi_dir/pkgroot\" -M \"$abi_dir/manifest.ucl\"") {
		t.Fatal("FreeBSD pkg builder does not create packages from manifest.ucl")
	}
	if !strings.Contains(script, "cp \"$abi_dir/manifest.ucl\" \"$repair_dir/+MANIFEST\"") ||
		!strings.Contains(script, "tar -cJf \"$tmp_pkg\" -P --no-recursion") ||
		!strings.Contains(script, "repair_manifest \"$abi_dir\" \"$pkg_file\"") {
		t.Fatal("FreeBSD pkg builder does not repack generated packages with the simple manifest")
	}
	if strings.Contains(script, "post-upgrade") {
		t.Fatal("FreeBSD pkg scripts must not use unsupported post-upgrade hooks")
	}
	for _, required := range []string{
		`flatsize = $flatsize`,
		`"/usr/local/bin/unifi-stubd" = "1\$`,
		`"/usr/local/etc/rc.d/unifi-stubd" = "1\$`,
		`"/usr/local/etc/unifi-stubd/config.yaml" = "1\$`,
		`config = [`,
		`  "/usr/local/etc/unifi-stubd/config.yaml"`,
		`scripts = {`,
		`post-install = <<EOS`,
		`/usr/local/bin/unifi-stubd -config-migrate -config /usr/local/etc/unifi-stubd/config.yaml || true`,
		`"/usr/local/share/doc/unifi-stubd/LICENSE" = "1\$`,
		`"/var/db/unifi-stubd" = "y"`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("FreeBSD pkg manifest does not contain %q", required)
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
