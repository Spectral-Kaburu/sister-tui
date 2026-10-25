package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Spectral-Kaburu/SpectreTUI/internal/schema"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		resp     schema.Response
		wantCode int
	}{
		{
			name:     "success action",
			resp:     schema.NewResponse(schema.Action{ID: "submit", ExitCode: 0}, nil),
			wantCode: 0,
		},
		{
			name:     "defer action",
			resp:     schema.NewResponse(schema.Action{ID: "defer", ExitCode: 10}, nil),
			wantCode: 10,
		},
		{
			name:     "skip action",
			resp:     schema.NewResponse(schema.Action{ID: "skip", ExitCode: 20}, nil),
			wantCode: 20,
		},
		{
			name:     "dismissal",
			resp:     schema.Dismissed(),
			wantCode: schema.DismissedExitCode, // 130
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := exitCode(tt.resp)
			if got != tt.wantCode {
				t.Errorf("exitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestRunCLI_InvalidJSON(t *testing.T) {
	stdin := bytes.NewBufferString("{ invalid json ")
	var stdout, stderr bytes.Buffer

	called := false
	mockRunPrompt := func(p schema.Payload) (schema.Response, error) {
		called = true
		return schema.Response{}, nil
	}

	code := run([]string{}, stdin, &stdout, &stderr, mockRunPrompt)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if called {
		t.Error("runPrompt should not be called on invalid JSON")
	}
	if !strings.HasPrefix(stderr.String(), "error: ") {
		t.Errorf("expected stderr to start with 'error: ', got %q", stderr.String())
	}
	// Assert exactly one stderr line
	lines := strings.Split(strings.TrimSpace(stderr.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected exactly 1 line on stderr, got %d: %v", len(lines), lines)
	}
}

func TestRunCLI_InvalidSchema(t *testing.T) {
	// Missing title
	stdin := bytes.NewBufferString(`{"actions": [{"id": "ok", "label": "OK", "exit_code": 0}]}`)
	var stdout, stderr bytes.Buffer

	called := false
	mockRunPrompt := func(p schema.Payload) (schema.Response, error) {
		called = true
		return schema.Response{}, nil
	}

	code := run([]string{}, stdin, &stdout, &stderr, mockRunPrompt)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if called {
		t.Error("runPrompt should not be called on invalid schema")
	}
	if !strings.Contains(stderr.String(), "missing \"title\"") {
		t.Errorf("expected missing title error on stderr, got %q", stderr.String())
	}
}

func TestRunCLI_StdinStdout(t *testing.T) {
	payloadJSON := `{
		"title": "CLI Test",
		"actions": [
			{"id": "submit", "label": "Submit", "exit_code": 0, "is_primary": true}
		]
	}`
	stdin := bytes.NewBufferString(payloadJSON)
	var stdout, stderr bytes.Buffer

	mockRunPrompt := func(p schema.Payload) (schema.Response, error) {
		if p.Title != "CLI Test" {
			t.Errorf("unexpected payload title %q", p.Title)
		}
		return schema.NewResponse(p.Actions[0], map[string]any{"key": "val"}), nil
	}

	code := run([]string{}, stdin, &stdout, &stderr, mockRunPrompt)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Errorf("expected empty stderr, got %q", stderr.String())
	}

	var resp schema.Response
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("stdout did not contain valid JSON: %v, raw: %s", err, stdout.String())
	}
	if resp.ActionID != "submit" || resp.ExitCode != 0 {
		t.Errorf("unexpected response: %+v", resp)
	}
	if resp.Answers["key"] != "val" {
		t.Errorf("unexpected answers: %+v", resp.Answers)
	}
}

func TestRunCLI_FileInputOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	inPath := filepath.Join(tmpDir, "prompt.json")
	outPath := filepath.Join(tmpDir, "result.json")

	payloadJSON := `{
		"title": "File Test",
		"actions": [
			{"id": "defer", "label": "Defer", "exit_code": 10}
		]
	}`
	if err := os.WriteFile(inPath, []byte(payloadJSON), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	mockRunPrompt := func(p schema.Payload) (schema.Response, error) {
		return schema.NewResponse(p.Actions[0], nil), nil
	}

	args := []string{"--input", inPath, "--output", outPath}
	code := run(args, bytes.NewBuffer(nil), &stdout, &stderr, mockRunPrompt)
	if code != 10 {
		t.Fatalf("expected exit code 10, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be completely empty in file mode, got %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty, got %q", stderr.String())
	}

	outBytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	var resp schema.Response
	if err := json.Unmarshal(outBytes, &resp); err != nil {
		t.Fatalf("output file did not contain valid response JSON: %v", err)
	}
	if resp.ActionID != "defer" || resp.ExitCode != 10 {
		t.Errorf("unexpected response in file: %+v", resp)
	}
}

func TestRunCLI_Dismissal(t *testing.T) {
	payloadJSON := `{
		"title": "Dismiss Test",
		"actions": [
			{"id": "ok", "label": "OK", "exit_code": 0}
		]
	}`
	stdin := bytes.NewBufferString(payloadJSON)
	var stdout, stderr bytes.Buffer

	mockRunPrompt := func(p schema.Payload) (schema.Response, error) {
		return schema.Dismissed(), nil
	}

	code := run([]string{}, stdin, &stdout, &stderr, mockRunPrompt)
	if code != 130 {
		t.Fatalf("expected exit code 130, got %d", code)
	}

	var resp schema.Response
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("stdout did not contain valid JSON: %v", err)
	}
	if resp.ActionID != schema.DismissedActionID || resp.ExitCode != 130 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestRunCLI_Flags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	mockRunPrompt := func(p schema.Payload) (schema.Response, error) {
		return schema.Response{}, nil
	}

	// Help flag
	code := run([]string{"--help"}, bytes.NewBuffer(nil), &stdout, &stderr, mockRunPrompt)
	if code != 0 {
		t.Errorf("expected 0 on --help, got %d", code)
	}

	// Invalid flag
	stderr.Reset()
	code = run([]string{"--unknown-flag"}, bytes.NewBuffer(nil), &stdout, &stderr, mockRunPrompt)
	if code != 1 {
		t.Errorf("expected 1 on unknown flag, got %d", code)
	}
}

func TestExamplesValidation(t *testing.T) {
	exampleFiles := []string{
		"../../examples/morning.json",
		"../../examples/missed_class.json",
		"../../examples/review.json",
	}

	for _, file := range exampleFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read example file %s: %v", file, err)
			}
			p, err := schema.Parse(data)
			if err != nil {
				t.Fatalf("failed to parse example file %s: %v", file, err)
			}
			if p.Title == "" {
				t.Errorf("expected non-empty title in %s", file)
			}
			if len(p.Actions) == 0 {
				t.Errorf("expected at least one action in %s", file)
			}
		})
	}
}

