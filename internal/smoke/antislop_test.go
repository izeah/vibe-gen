//go:build smoke

package smoke

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	vibegen "github.com/izeah01/vibe-gen"
)

// TestAntislopRules proves the slop lint (templates/common/.semgrep/antislop.yml.tmpl) works: every
// rule fires on testdata/antislop, while the clean, nosemgrep-suppressed, token-file and e2e/ files stay silent.
// Requires semgrep on PATH.
func TestAntislopRules(t *testing.T) {
	rules, err := vibegen.Templates.ReadFile("templates/common/.semgrep/antislop.yml.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	err = filepath.WalkDir("testdata/antislop", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel("testdata/antislop", path)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, content, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".semgrep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".semgrep", "antislop.yml"), rules, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("semgrep", "scan", "--config", ".semgrep/antislop.yml", "--metrics=off", "--disable-version-check", "--quiet", "--json")
	cmd.Dir, cmd.Stdout, cmd.Stderr = dir, &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("semgrep: %v\n%s", err, stderr.String())
	}
	var res struct {
		Results []struct {
			CheckID string `json:"check_id"`
			Path    string `json:"path"`
		} `json:"results"`
		Errors []json.RawMessage `json:"errors"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("semgrep output: %v", err)
	}
	if len(res.Errors) > 0 {
		t.Fatalf("semgrep reported errors: %s", res.Errors)
	}

	silent := []string{"src/ok.", "src/suppressed.", "src/styles/", "e2e/"}
	fired := map[string]bool{}
	for _, r := range res.Results {
		fired[r.CheckID[strings.LastIndex(r.CheckID, ".")+1:]] = true
		for _, prefix := range silent {
			if strings.HasPrefix(r.Path, prefix) {
				t.Errorf("%s fired on %s, which must stay silent", r.CheckID, r.Path)
			}
		}
	}
	ids := regexp.MustCompile(`(?m)^\s*- id: (antislop-[a-z-]+)`).FindAllSubmatch(rules, -1)
	if len(ids) == 0 {
		t.Fatal("no rules found in antislop.yml.tmpl")
	}
	for _, m := range ids {
		if id := string(m[1]); !fired[id] {
			t.Errorf("rule %s never fired on testdata/antislop", id)
		}
	}
}
