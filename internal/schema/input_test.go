package schema

import (
	"strings"
	"testing"
)

const validPayload = `{
  "title": "Quick Prompt",
  "width": 96,
  "blocks": [
    {"type": "markdown", "severity": "roast", "content": "### Attention\nYou are idle."},
    {"type": "single_select", "id": "status", "label": "Doing?", "options": [
      {"id": "coding", "label": "Writing code"},
      {"id": "slacking", "label": "Procrastinating"}
    ], "default": "coding"},
    {"type": "multi_select", "id": "goals", "label": "Goals:", "options": [
      {"id": "g1", "label": "Spec"}, {"id": "g2", "label": "Bug"}
    ], "default": ["g1"]},
    {"type": "text_input", "id": "note", "label": "Notes:", "required": true},
    {"type": "textarea", "id": "detail", "label": "Detail:"},
    {"type": "confirm", "id": "ok", "label": "Sure?", "default": false}
  ],
  "actions": [
    {"id": "submit", "label": "[Enter] Done", "exit_code": 0, "is_primary": true},
    {"id": "defer", "label": "[Tab+1] Ask Later", "exit_code": 10}
  ]
}`

func TestParse(t *testing.T) {
	minAction := `{"id": "a", "label": "L", "exit_code": 0}`
	cases := []struct {
		name    string
		payload string
		wantErr string // "" = must pass; else substring of the error
	}{
		{"valid", validPayload, ""},
		{"width 96 ok", `{"title": "t", "width": 96, "actions": [` + minAction + `]}`, ""},
		{"width too small", `{"title": "t", "width": 30, "actions": [` + minAction + `]}`, `payload: "width" must be >= 40`},
		{"bad json", `{`, "invalid JSON"},
		{"missing title", `{"actions": [` + minAction + `]}`, `payload: missing "title"`},
		{"no actions", `{"title": "t"}`, "at least one action"},
		{"unknown block type", `{"title": "t", "blocks": [{"type": "spinner", "id": "x", "label": "L"}], "actions": [` + minAction + `]}`, `unknown type "spinner"`},
		{"markdown severity default", `{"title": "t", "blocks": [{"type": "markdown", "content": "hi"}], "actions": [` + minAction + `]}`, ""},
		{"unknown severity", `{"title": "t", "blocks": [{"type": "markdown", "severity": "meh", "content": "hi"}], "actions": [` + minAction + `]}`, `unknown severity "meh"`},
		{"select missing id", `{"title": "t", "blocks": [{"type": "single_select", "label": "L", "options": [{"id": "o", "label": "O"}]}], "actions": [` + minAction + `]}`, `block 0: missing "id"`},
		{"select no options", `{"title": "t", "blocks": [{"type": "single_select", "id": "x", "label": "L"}], "actions": [` + minAction + `]}`, "at least one option"},
		{"select bad default", `{"title": "t", "blocks": [{"type": "single_select", "id": "x", "label": "L", "options": [{"id": "o", "label": "O"}], "default": "nope"}], "actions": [` + minAction + `]}`, `default "nope" is not an option`},
		{"multi default not array", `{"title": "t", "blocks": [{"type": "multi_select", "id": "x", "label": "L", "options": [{"id": "o", "label": "O"}], "default": "o"}], "actions": [` + minAction + `]}`, "default must be an array"},
		{"multi default bad id", `{"title": "t", "blocks": [{"type": "multi_select", "id": "x", "label": "L", "options": [{"id": "o", "label": "O"}], "default": ["zap"]}], "actions": [` + minAction + `]}`, `default "zap" is not an option`},
		{"confirm default not bool", `{"title": "t", "blocks": [{"type": "confirm", "id": "x", "label": "L", "default": "yes"}], "actions": [` + minAction + `]}`, "default must be a bool"},
		{"text default not string", `{"title": "t", "blocks": [{"type": "text_input", "id": "x", "label": "L", "default": 5}], "actions": [` + minAction + `]}`, "default must be a string"},
		{"duplicate block ids", `{"title": "t", "blocks": [{"type": "text_input", "id": "x", "label": "L"}, {"type": "textarea", "id": "x", "label": "L"}], "actions": [` + minAction + `]}`, `duplicate id "x"`},
		{"duplicate option ids", `{"title": "t", "blocks": [{"type": "single_select", "id": "s", "label": "L", "options": [{"id": "o", "label": "A"}, {"id": "o", "label": "B"}]}], "actions": [` + minAction + `]}`, `option 1: duplicate id "o"`},
		{"action exit code 130", `{"title": "t", "actions": [{"id": "a", "label": "L", "exit_code": 130}]}`, "reserved for dismissal"},
		{"action missing label", `{"title": "t", "actions": [{"id": "a", "exit_code": 0}]}`, `action "a": missing "label"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.payload))
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("want error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("want error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestSeverityNormalized(t *testing.T) {
	p, err := Parse([]byte(`{"title": "t", "blocks": [{"type": "markdown", "content": "hi"}], "actions": [{"id": "a", "label": "L", "exit_code": 0}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if p.Blocks[0].Severity != SevInfo {
		t.Fatalf("want severity %q after normalization, got %q", SevInfo, p.Blocks[0].Severity)
	}
}
