package schema

import "time"

// Response is what SpectreTUI emits when the user is done (design doc §5).
type Response struct {
	ActionID    string         `json:"action_id"`
	ExitCode    int            `json:"exit_code"`
	CompletedAt string         `json:"completed_at"`
	Answers     map[string]any `json:"answers"`
}

// NewResponse records a chosen action. answers may be nil for
// layout-only prompts; it still serializes as {}.
func NewResponse(a Action, answers map[string]any) Response {
	if answers == nil {
		answers = map[string]any{}
	}
	return Response{
		ActionID:    a.ID,
		ExitCode:    a.ExitCode,
		CompletedAt: time.Now().Format(time.RFC3339),
		Answers:     answers,
	}
}

// Dismissed is the abrupt-cancellation response (Esc/Ctrl+C → exit 130).
func Dismissed() Response {
	return Response{
		ActionID:    DismissedActionID,
		ExitCode:    DismissedExitCode,
		CompletedAt: time.Now().Format(time.RFC3339),
		Answers:     map[string]any{},
	}
}
