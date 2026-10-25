package theme

import (
	"github.com/Spectral-Kaburu/SpectreTUI/internal/schema"
	"github.com/charmbracelet/lipgloss"
)

// DefaultCardWidth is the fallback card width if payload does not specify one.
const DefaultCardWidth = 100

// MinCardWidth is the minimum allowed card width.
const MinCardWidth = 40

// SeverityStyle defines the accent color and badge text for a severity level.
type SeverityStyle struct {
	Accent lipgloss.Color
	Badge  string
}

// Severities maps severity identifiers to their visual styles.
var Severities = map[string]SeverityStyle{
	schema.SevInfo: {
		Accent: lipgloss.Color("#06b6d4"), // Cyan
		Badge:  "ℹ INFO",
	},
	schema.SevTip: {
		Accent: lipgloss.Color("#10b981"), // Emerald / Green
		Badge:  "✔ TIP",
	},
	schema.SevWarning: {
		Accent: lipgloss.Color("#f59e0b"), // Amber / Yellow
		Badge:  "⚠ WARNING",
	},
	schema.SevCritical: {
		Accent: lipgloss.Color("#ef4444"), // Bright Red
		Badge:  "✖ CRITICAL",
	},
	schema.SevRoast: {
		Accent: lipgloss.Color("#d946ef"), // Magenta / Deep Purple
		Badge:  "🔥 SISTER ROAST",
	},
}

// CardWidth calculates the effective card width given payload width and terminal width:
// min(payloadWidth, termWidth-4) with floor 40.
// A payloadWidth of 0 uses DefaultCardWidth (100).
func CardWidth(payloadWidth, termWidth int) int {
	w := payloadWidth
	if w <= 0 {
		w = DefaultCardWidth
	}
	avail := termWidth - 4
	if avail < w {
		w = avail
	}
	if w < MinCardWidth {
		w = MinCardWidth
	}
	return w
}

// CardBorder returns a rounded border with left/right padding 1 and no top/bottom padding.
func CardBorder() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)
}

// FocusRing returns a subtle border to show which block has focus.
func FocusRing() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6366f1")).
		Padding(0, 1)
}

// SeverityBox returns a lipgloss Style with the correct accent color left border,
// with width bounded to the card width.
func SeverityBox(sev string, width int) lipgloss.Style {
	style, ok := Severities[sev]
	if !ok {
		style = Severities[schema.SevInfo]
	}

	s := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(style.Accent).
		PaddingLeft(1)

	if width > 0 {
		s = s.Width(width)
	}

	return s
}
