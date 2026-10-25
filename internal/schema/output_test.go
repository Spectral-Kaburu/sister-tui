package schema

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewResponseShape(t *testing.T) {
	r := NewResponse(Action{ID: "submit", ExitCode: 0, Primary: true}, map[string]any{
		"status": "slacking",
		"goals":  []string{"g1", "g3"},
		"note":   "fix scheduler",
		"ok":     true,
	})
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"action_id", "exit_code", "completed_at", "answers"} {
		if _, ok := m[key]; !ok {
			t.Errorf("response JSON missing %q: %s", key, b)
		}
	}
	if m["action_id"] != "submit" {
		t.Errorf("action_id = %v, want submit", m["action_id"])
	}
	if _, err := time.Parse(time.RFC3339, r.CompletedAt); err != nil {
		t.Errorf("completed_at %q is not RFC3339: %v", r.CompletedAt, err)
	}
	answers := m["answers"].(map[string]any)
	if answers["status"] != "slacking" {
		t.Errorf("answers.status = %v, want slacking", answers["status"])
	}
	goals := answers["goals"].([]any)
	if len(goals) != 2 || goals[0] != "g1" || goals[1] != "g3" {
		t.Errorf("answers.goals = %v, want [g1 g3]", goals)
	}
	if answers["ok"] != true {
		t.Errorf("answers.ok = %v, want true", answers["ok"])
	}
}

func TestNewResponseNilAnswers(t *testing.T) {
	b, err := json.Marshal(NewResponse(Action{ID: "submit", ExitCode: 0}, nil))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"answers":{}`) {
		t.Errorf("nil answers must serialize as {}, got %s", b)
	}
}

func TestDismissedShape(t *testing.T) {
	b, err := json.Marshal(Dismissed())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"action_id":"dismissed"`, `"exit_code":130`, `"answers":{}`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("dismissed JSON missing %s: %s", want, b)
		}
	}
}
