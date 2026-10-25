package blocks

import (
	"strings"
	"testing"

	"github.com/Spectral-Kaburu/SpectreTUI/internal/schema"
	"github.com/Spectral-Kaburu/SpectreTUI/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func TestMarkdownBlock_Severities(t *testing.T) {
	tests := []struct {
		name      string
		severity  string
		wantBadge string
	}{
		{name: "info", severity: schema.SevInfo, wantBadge: "ℹ INFO"},
		{name: "tip", severity: schema.SevTip, wantBadge: "✔ TIP"},
		{name: "warning", severity: schema.SevWarning, wantBadge: "⚠ WARNING"},
		{name: "critical", severity: schema.SevCritical, wantBadge: "✖ CRITICAL"},
		{name: "roast", severity: schema.SevRoast, wantBadge: "🔥 SISTER ROAST"},
	}

	cardWidth := 60
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := schema.Block{
				Type:     schema.TypeMarkdown,
				Severity: tt.severity,
				Content:  "# Header\nThis is a test markdown line that will wrap appropriately.",
			}
			mb := NewMarkdownBlock(b, cardWidth)
			out := mb.View(false)

			if len(strings.TrimSpace(out)) == 0 {
				t.Fatalf("expected non-empty output for severity %q", tt.severity)
			}
			if !strings.Contains(out, tt.wantBadge) {
				t.Errorf("expected badge %q in output, got:\n%s", tt.wantBadge, out)
			}
			t.Logf("Rendered %s:\n%s", tt.severity, out)

			maxAllowed := cardWidth + 10
			lines := strings.Split(out, "\n")
			for i, line := range lines {
				w := lipgloss.Width(line)
				if w > maxAllowed {
					t.Errorf("line %d width %d exceeds max allowed %d (%q)", i, w, maxAllowed, line)
				}
			}
		})
	}
}

func TestMarkdownBlock_Cache(t *testing.T) {
	b := schema.Block{
		Type:     schema.TypeMarkdown,
		Severity: schema.SevInfo,
		Content:  "Cached content test",
	}
	mb := NewMarkdownBlock(b, 80)

	first := mb.View(false)
	if first == "" {
		t.Fatal("expected non-empty output on first View()")
	}
	if mb.renderCount != 1 {
		t.Fatalf("expected renderCount=1 after first View(), got %d", mb.renderCount)
	}

	second := mb.View(false)
	if first != second {
		t.Errorf("expected repeated View() to return identical string, got:\n%q\nvs\n%q", first, second)
	}
	if mb.renderCount != 1 {
		t.Fatalf("expected renderCount to remain 1 after second View(), got %d", mb.renderCount)
	}

	third := mb.View(true)
	if third != first {
		t.Errorf("expected View(true) to return identical string, got:\n%q\nvs\n%q", third, first)
	}
	if mb.renderCount != 1 {
		t.Fatalf("expected renderCount to remain 1 after View(true), got %d", mb.renderCount)
	}
}

func TestMarkdownBlock_SetWidth_InvalidatesCache(t *testing.T) {
	longText := "This is a long sentence of markdown text designed to wrap across different column boundaries when the block width is changed from narrow to wide."
	b := schema.Block{
		Type:     schema.TypeMarkdown,
		Severity: schema.SevWarning,
		Content:  longText,
	}
	mb := NewMarkdownBlock(b, 45)

	out1 := mb.View(false)
	if out1 == "" {
		t.Fatal("expected non-empty output on first View()")
	}
	if mb.renderCount != 1 {
		t.Fatalf("expected renderCount=1, got %d", mb.renderCount)
	}

	// Change width to a significantly different width
	mb.SetWidth(95)

	out2 := mb.View(false)
	if out2 == "" {
		t.Fatal("expected non-empty output after SetWidth and View()")
	}
	if mb.renderCount != 2 {
		t.Fatalf("expected renderCount=2 after SetWidth and View(), got %d", mb.renderCount)
	}
	if out1 == out2 {
		t.Errorf("expected output to differ between width 45 and width 95")
	}

	// Verify line widths are within limits
	for i, line := range strings.Split(out1, "\n") {
		w := lipgloss.Width(line)
		if w > 45+10 {
			t.Errorf("width 45: line %d width %d exceeds max allowed: %q", i, w, line)
		}
	}
	for i, line := range strings.Split(out2, "\n") {
		w := lipgloss.Width(line)
		if w > 95+10 {
			t.Errorf("width 95: line %d width %d exceeds max allowed: %q", i, w, line)
		}
	}

	// Setting identical width should NOT invalidate cache
	mb.SetWidth(95)
	out3 := mb.View(false)
	if mb.renderCount != 2 {
		t.Fatalf("expected renderCount to remain 2 when SetWidth called with same width, got %d", mb.renderCount)
	}
	if out3 != out2 {
		t.Errorf("expected identical output when width unchanged")
	}
}

func TestMarkdownBlock_EmptyContent(t *testing.T) {
	b := schema.Block{
		Type:     schema.TypeMarkdown,
		Severity: schema.SevTip,
		Content:  "",
	}
	mb := NewMarkdownBlock(b, 60)
	out := mb.View(false)
	if !strings.Contains(out, "✔ TIP") {
		t.Errorf("expected badge in empty content block, got %q", out)
	}
}

func TestMarkdownBlock_ID_Answer_Update(t *testing.T) {
	b := schema.Block{
		Type:     schema.TypeMarkdown,
		ID:       "should_be_ignored",
		Severity: schema.SevInfo,
		Content:  "Content",
	}
	mb := NewMarkdownBlock(b, 80)

	if id := mb.ID(); id != "" {
		t.Errorf("expected empty string ID for markdown block, got %q", id)
	}
	if ans := mb.Answer(); ans != nil {
		t.Errorf("expected nil Answer for markdown block, got %v", ans)
	}

	updatedBlock, cmd := mb.Update(nil)
	if updatedBlock != mb {
		t.Errorf("expected Update to return self")
	}
	if cmd != nil {
		t.Errorf("expected Update to return nil cmd")
	}
}

func TestTheme_CardWidth(t *testing.T) {
	tests := []struct {
		name         string
		payloadWidth int
		termWidth    int
		want         int
	}{
		{name: "default width when 0", payloadWidth: 0, termWidth: 120, want: 100},
		{name: "clamped by terminal width", payloadWidth: 100, termWidth: 80, want: 76},
		{name: "floor 40 when terminal is narrow", payloadWidth: 100, termWidth: 35, want: 40},
		{name: "floor 40 when payload width is small", payloadWidth: 30, termWidth: 100, want: 40},
		{name: "exact payload width fits", payloadWidth: 60, termWidth: 100, want: 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := theme.CardWidth(tt.payloadWidth, tt.termWidth)
			if got != tt.want {
				t.Errorf("CardWidth(%d, %d) = %d, want %d", tt.payloadWidth, tt.termWidth, got, tt.want)
			}
		})
	}
}

func TestTheme_Severities(t *testing.T) {
	expected := map[string]string{
		schema.SevInfo:     "ℹ INFO",
		schema.SevTip:      "✔ TIP",
		schema.SevWarning:  "⚠ WARNING",
		schema.SevCritical: "✖ CRITICAL",
		schema.SevRoast:    "🔥 SISTER ROAST",
	}

	for sev, wantBadge := range expected {
		style, ok := theme.Severities[sev]
		if !ok {
			t.Errorf("missing severity %q in theme.Severities", sev)
			continue
		}
		if style.Badge != wantBadge {
			t.Errorf("severity %q: badge = %q, want %q", sev, style.Badge, wantBadge)
		}
		if style.Accent == "" {
			t.Errorf("severity %q: accent color is empty", sev)
		}
	}
}

func TestTheme_CardBorder_FocusRing(t *testing.T) {
	cb := theme.CardBorder()
	renderedCard := cb.Render("content")
	if renderedCard == "" {
		t.Error("expected non-empty rendered CardBorder")
	}

	fr := theme.FocusRing()
	renderedFocus := fr.Render("focused content")
	if renderedFocus == "" {
		t.Error("expected non-empty rendered FocusRing")
	}
}

func TestTheme_SeverityBox(t *testing.T) {
	// Known severity
	box := theme.SeverityBox(schema.SevCritical, 50)
	rendered := box.Render("critical message")
	if !strings.Contains(rendered, "critical message") {
		t.Errorf("expected rendered text in box, got %q", rendered)
	}

	// Unknown severity should fallback gracefully
	boxUnknown := theme.SeverityBox("unknown", 50)
	renderedUnknown := boxUnknown.Render("unknown message")
	if !strings.Contains(renderedUnknown, "unknown message") {
		t.Errorf("expected rendered text in fallback box, got %q", renderedUnknown)
	}
}
