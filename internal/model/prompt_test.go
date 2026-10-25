package model

import (
	"strings"
	"testing"

	"github.com/Spectral-Kaburu/SpectreTUI/internal/blocks"
	"github.com/Spectral-Kaburu/SpectreTUI/internal/schema"
	"github.com/Spectral-Kaburu/SpectreTUI/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
)

func TestPromptModel_FocusCycling(t *testing.T) {
	p := schema.Payload{
		Title: "Focus Test",
		Blocks: []schema.Block{
			{Type: schema.TypeMarkdown, Content: "Block 0"},
			{Type: schema.TypeMarkdown, Content: "Block 1"},
		},
		Actions: []schema.Action{
			{ID: "act0", Label: "Action 0", ExitCode: 0, Primary: true},
			{ID: "act1", Label: "Action 1", ExitCode: 10},
		},
	}
	m := NewPromptModel(p)
	// Initial focus: because there are no interactive blocks, focus starts on primary action act0 (index 2)
	if m.FocusIndex() != 2 {
		t.Fatalf("expected initial focus index 2, got %d", m.FocusIndex())
	}

	// Tab forward: 2 -> 3 (act1)
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(PromptModel)
	if m.FocusIndex() != 3 {
		t.Fatalf("after Tab, expected focus 3, got %d", m.FocusIndex())
	}

	// Tab forward: 3 -> 0 (block 0)
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(PromptModel)
	if m.FocusIndex() != 0 {
		t.Fatalf("after Tab, expected focus 0, got %d", m.FocusIndex())
	}

	// Tab forward: 0 -> 1 (block 1)
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(PromptModel)
	if m.FocusIndex() != 1 {
		t.Fatalf("after Tab, expected focus 1, got %d", m.FocusIndex())
	}

	// Tab forward: 1 -> 2 (act0)
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = res.(PromptModel)
	if m.FocusIndex() != 2 {
		t.Fatalf("after Tab, expected focus 2, got %d", m.FocusIndex())
	}

	// Shift+Tab backward: 2 -> 1
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = res.(PromptModel)
	if m.FocusIndex() != 1 {
		t.Fatalf("after Shift+Tab, expected focus 1, got %d", m.FocusIndex())
	}

	// Shift+Tab backward: 1 -> 0
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = res.(PromptModel)
	if m.FocusIndex() != 0 {
		t.Fatalf("after Shift+Tab, expected focus 0, got %d", m.FocusIndex())
	}

	// Shift+Tab backward: 0 -> 3
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = res.(PromptModel)
	if m.FocusIndex() != 3 {
		t.Fatalf("after Shift+Tab, expected focus 3, got %d", m.FocusIndex())
	}
}

func TestPromptModel_Submit(t *testing.T) {
	p := schema.Payload{
		Title: "Submit Test",
		Blocks: []schema.Block{
			{Type: schema.TypeMarkdown, Content: "Read me"},
		},
		Actions: []schema.Action{
			{ID: "submit", Label: "Submit", ExitCode: 0, Primary: true},
			{ID: "defer", Label: "Defer", ExitCode: 10},
		},
	}
	m := NewPromptModel(p)
	// Focus starts on submit (index 1)
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = res.(PromptModel)
	if cmd == nil {
		t.Fatal("expected tea.Quit command, got nil")
	}
	resp := m.Response()
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.ActionID != "submit" {
		t.Errorf("expected ActionID 'submit', got %q", resp.ActionID)
	}
	if resp.ExitCode != 0 {
		t.Errorf("expected ExitCode 0, got %d", resp.ExitCode)
	}

	// Test submitting second action
	m2 := NewPromptModel(p)
	res2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRight})
	m2 = res2.(PromptModel)
	if m2.FocusIndex() != 2 {
		t.Fatalf("expected focus index 2, got %d", m2.FocusIndex())
	}
	res2, cmd2 := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m2 = res2.(PromptModel)
	if cmd2 == nil {
		t.Fatal("expected tea.Quit, got nil")
	}
	resp2 := m2.Response()
	if resp2 == nil {
		t.Fatal("expected response, got nil")
	}
	if resp2.ActionID != "defer" || resp2.ExitCode != 10 {
		t.Errorf("expected defer/10, got %s/%d", resp2.ActionID, resp2.ExitCode)
	}
}

func TestPromptModel_Dismiss(t *testing.T) {
	p := schema.Payload{
		Title: "Dismiss Test",
		Actions: []schema.Action{
			{ID: "ok", Label: "OK", ExitCode: 0},
		},
	}

	// Test Esc
	m := NewPromptModel(p)
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = res.(PromptModel)
	if cmd == nil {
		t.Fatal("expected tea.Quit on Esc")
	}
	if m.Response() == nil || m.Response().ActionID != schema.DismissedActionID || m.Response().ExitCode != schema.DismissedExitCode {
		t.Fatalf("expected dismissed response on Esc, got %+v", m.Response())
	}

	// Test Ctrl+C
	m2 := NewPromptModel(p)
	res2, cmd2 := m2.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m2 = res2.(PromptModel)
	if cmd2 == nil {
		t.Fatal("expected tea.Quit on Ctrl+C")
	}
	if m2.Response() == nil || m2.Response().ActionID != schema.DismissedActionID || m2.Response().ExitCode != 130 {
		t.Fatalf("expected dismissed response on Ctrl+C, got %+v", m2.Response())
	}
}

func TestPromptModel_DigitShortcuts(t *testing.T) {
	p := schema.Payload{
		Title: "Digits Test",
		Actions: []schema.Action{
			{ID: "first", Label: "First", ExitCode: 0},
			{ID: "second", Label: "Second", ExitCode: 2},
			{ID: "third", Label: "Third", ExitCode: 5},
		},
	}

	// Press '2'
	m := NewPromptModel(p)
	res, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = res.(PromptModel)
	if cmd == nil {
		t.Fatal("expected tea.Quit on digit 2")
	}
	if m.Response() == nil || m.Response().ActionID != "second" || m.Response().ExitCode != 2 {
		t.Fatalf("expected second/2, got %+v", m.Response())
	}

	// Press '3'
	m3 := NewPromptModel(p)
	res3, cmd3 := m3.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m3 = res3.(PromptModel)
	if cmd3 == nil {
		t.Fatal("expected tea.Quit on digit 3")
	}
	if m3.Response() == nil || m3.Response().ActionID != "third" || m3.Response().ExitCode != 5 {
		t.Fatalf("expected third/5, got %+v", m3.Response())
	}

	// Press '9' (out of range, should not quit)
	m9 := NewPromptModel(p)
	res9, cmd9 := m9.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}})
	m9 = res9.(PromptModel)
	if cmd9 != nil || m9.Response() != nil {
		t.Fatalf("expected digit 9 to be ignored, got response %+v", m9.Response())
	}
}

func TestPromptModel_WindowSize(t *testing.T) {
	p := schema.Payload{
		Title: "WindowSize Test",
		Width: 90,
		Blocks: []schema.Block{
			{Type: schema.TypeMarkdown, Content: "Responsive text"},
		},
		Actions: []schema.Action{
			{ID: "ok", Label: "OK", ExitCode: 0},
		},
	}
	m := NewPromptModel(p)
	// Send WindowSizeMsg with width 70
	res, _ := m.Update(tea.WindowSizeMsg{Width: 70, Height: 30})
	m = res.(PromptModel)

	expectedCardWidth := theme.CardWidth(90, 70) // min(90, 70-4=66) -> 66
	if m.Width() != expectedCardWidth {
		t.Errorf("expected card width %d, got %d", expectedCardWidth, m.Width())
	}

	expectedContentWidth := expectedCardWidth - 4
	mb, ok := m.Blocks()[0].(*blocks.MarkdownBlock)
	if !ok {
		t.Fatalf("expected MarkdownBlock, got %T", m.Blocks()[0])
	}
	if mb.Width() != expectedContentWidth {
		t.Errorf("expected block width %d, got %d", expectedContentWidth, mb.Width())
	}

	// Narrow terminal below floor 40
	res, _ = m.Update(tea.WindowSizeMsg{Width: 30, Height: 30})
	m = res.(PromptModel)
	if m.Width() != theme.MinCardWidth {
		t.Errorf("expected floor width %d, got %d", theme.MinCardWidth, m.Width())
	}
}

func TestPromptModel_ClickFocus(t *testing.T) {
	p := schema.Payload{
		Title: "Mouse Test",
		Width: 80,
		Blocks: []schema.Block{
			{Type: schema.TypeMarkdown, Content: "Clickable block content"},
		},
		Actions: []schema.Action{
			{ID: "first", Label: "First", ExitCode: 0, Primary: true},
			{ID: "second", Label: "Second", ExitCode: 20},
		},
	}
	m := NewPromptModel(p)
	res, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = res.(PromptModel)

	_, bZones, aZones := m.layout()
	if len(bZones) != 1 {
		t.Fatalf("expected 1 block hit zone, got %d", len(bZones))
	}
	if len(aZones) != 2 {
		t.Fatalf("expected 2 action hit zones, got %d", len(aZones))
	}

	// 1. Click on block 0: should focus block 0
	bz := bZones[0]
	clickBlockMsg := tea.MouseMsg{
		X:      (100 - m.Width())/2 + 5,
		Y:      (bz.minY + bz.maxY) / 2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	res, cmd := m.Update(clickBlockMsg)
	m = res.(PromptModel)
	if cmd != nil {
		t.Errorf("clicking block should not quit")
	}
	if m.FocusIndex() != 0 {
		t.Errorf("expected focus on block 0, got %d", m.FocusIndex())
	}

	// 2. Click on action button 1 ("second"): should activate action 1
	az := aZones[1]
	clickActionMsg := tea.MouseMsg{
		X:      (az.minX + az.maxX) / 2,
		Y:      (az.minY + az.maxY) / 2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	res, cmd = m.Update(clickActionMsg)
	m = res.(PromptModel)
	if cmd == nil {
		t.Fatal("expected tea.Quit on action click")
	}
	if m.Response() == nil || m.Response().ActionID != "second" || m.Response().ExitCode != 20 {
		t.Fatalf("expected second/20 response, got %+v", m.Response())
	}
}

func TestPromptModel_ActionArrowNavigation(t *testing.T) {
	p := schema.Payload{
		Title: "Arrow Test",
		Actions: []schema.Action{
			{ID: "a0", Label: "Action 0", ExitCode: 0},
			{ID: "a1", Label: "Action 1", ExitCode: 1},
			{ID: "a2", Label: "Action 2", ExitCode: 2},
		},
	}
	m := NewPromptModel(p)
	if m.FocusIndex() != 0 {
		t.Fatalf("expected focus 0, got %d", m.FocusIndex())
	}

	// Right arrow -> 1
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = res.(PromptModel)
	if m.FocusIndex() != 1 {
		t.Errorf("expected focus 1, got %d", m.FocusIndex())
	}

	// Right arrow -> 2
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = res.(PromptModel)
	if m.FocusIndex() != 2 {
		t.Errorf("expected focus 2, got %d", m.FocusIndex())
	}

	// Right arrow wraps -> 0
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = res.(PromptModel)
	if m.FocusIndex() != 0 {
		t.Errorf("expected focus 0, got %d", m.FocusIndex())
	}

	// Left arrow wraps -> 2
	res, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = res.(PromptModel)
	if m.FocusIndex() != 2 {
		t.Errorf("expected focus 2, got %d", m.FocusIndex())
	}
}

func TestPromptModel_View(t *testing.T) {
	p := schema.Payload{
		Title:    "Morning Briefing",
		Subtitle: "Check today's status",
		Width:    80,
		Blocks: []schema.Block{
			{Type: schema.TypeMarkdown, Severity: schema.SevTip, Content: "Everything looks good!"},
		},
		Actions: []schema.Action{
			{ID: "ok", Label: "Acknowledge", ExitCode: 0, Primary: true},
		},
	}
	m := NewPromptModel(p)
	res, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = res.(PromptModel)

	view := m.View()
	if !strings.Contains(view, "Morning Briefing") {
		t.Errorf("expected title in view")
	}
	if !strings.Contains(view, "Check today's status") {
		t.Errorf("expected subtitle in view")
	}
	if !strings.Contains(view, "Acknowledge") {
		t.Errorf("expected action label in view")
	}
}
