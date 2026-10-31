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

// Invalid native context and incomplete domain inputs must never enter a
// provider, VM, controller or evidence-writing operation.
func TestQualificationRejectsInvalidContextAndMissingInputs(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, script := range []string{"qualify-candidate.sh", "qualify-candidate-matrix.sh"} {
		for _, check := range []struct {
			name, root, legacy, message string
		}{
			{"missing-context", "", "", "inherit the native Core consumer context"},
			{"different-root", filepath.Join(root, "tests"), "", "inherit the native Core consumer context"},
			{"obsolete-root", root, root, "UNIFI_STUBD_CORE_ROOT is obsolete"},
			{"missing-data-locator", root, "", "set LAB_DATA_ROOT"},
		} {
			t.Run(script+"/"+check.name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				command := exec.CommandContext(ctx, "bash", filepath.Join("lab/stub/scripts", script))
				command.Dir = root
				for _, entry := range os.Environ() {
					key, _, _ := strings.Cut(entry, "=")
					if key == "PROVIDER_CONSUMER_ROOT" || key == "PROVIDER_CONSUMER_CONTEXT" ||
						key == "UNIFI_STUBD_CORE_ROOT" || key == "LAB_DATA_ROOT" || strings.HasPrefix(key, "TIER3_") {
						continue
					}
					command.Env = append(command.Env, entry)
				}
				command.Env = append(command.Env, "PROVIDER_CONSUMER_ROOT="+check.root,
					`PROVIDER_CONSUMER_CONTEXT={"consumer_id":"unifi-stubd","persistence":".consumer-data","resources":[]}`)
				if check.legacy != "" {
					command.Env = append(command.Env, "UNIFI_STUBD_CORE_ROOT="+check.legacy)
				}
				output, err := command.CombinedOutput()
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 1 || !strings.Contains(string(output), check.message) {
					t.Fatalf("qualification did not reject the invalid input at its boundary: %v\n%s", err, output)
				}
			})
		}
	}
}

func TestQualificationManifestIsCompleteAndNeverExecuted(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(root, "tests", ".work")
	if err := os.Mkdir(work, 0700); err == nil {
		t.Cleanup(func() {
			if err := os.Remove(work); err != nil {
				t.Errorf("remove owned empty verification directory: %v", err)
			}
		})
	} else if !os.IsExist(err) {
		t.Fatal(err)
	}
	if physical, err := filepath.EvalSymlinks(work); err != nil || physical != work {
		t.Fatalf("verification directory is not physical: %s, %v", physical, err)
	}
	fixture, err := os.MkdirTemp(work, "qualification-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(fixture); err != nil {
			t.Errorf("remove owned manifest fixture: %v", err)
		}
	})
	sentinel := filepath.Join(fixture, "executed")
	for _, check := range []struct {
		name, manifest string
		valid          bool
	}{
		{"complete", "SCHEMA_VERSION=1\nRUN_ID=fixture\n", true},
		{"missing", "SCHEMA_VERSION=1\n", false},
		{"duplicate", "SCHEMA_VERSION=1\nRUN_ID=one\nRUN_ID=two\n", false},
		{"unknown", "SCHEMA_VERSION=1\nRUN_ID=one\nEXTRA=value\n", false},
		{"whitespace", "SCHEMA_VERSION=1\nRUN_ID=two words\n", false},
		{"literal-shell", "SCHEMA_VERSION=1\nRUN_ID=$(touch${IFS}" + sentinel + ")\n", true},
	} {
		t.Run(check.name, func(t *testing.T) {
			manifest := filepath.Join(fixture, check.name+".env")
			if err := os.WriteFile(manifest, []byte(check.manifest), 0600); err != nil {
				t.Fatal(err)
			}
			command := exec.Command("bash", "-c", `source "$1"; manifest_validate "$2" 'SCHEMA_VERSION RUN_ID'`,
				"manifest-test", filepath.Join(root, "lab/stub/scripts/lib/manifest.sh"), manifest)
			output, err := command.CombinedOutput()
			if (err == nil) != check.valid {
				t.Fatalf("manifest contract valid=%v: %v\n%s", check.valid, err, output)
			}
			if _, err := os.Lstat(sentinel); !os.IsNotExist(err) {
				t.Fatalf("manifest data was executed: %v", err)
			}
		})
	}
}
