package blocks

import (
	"strings"

	"github.com/Spectral-Kaburu/SpectreTUI/internal/schema"
	"github.com/Spectral-Kaburu/SpectreTUI/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// MarkdownBlock renders markdown content with severity styling.
type MarkdownBlock struct {
	block       schema.Block
	width       int
	rendered    string
	computed    bool
	renderCount int
}

// NewMarkdownBlock creates a new MarkdownBlock.
func NewMarkdownBlock(b schema.Block, width int) *MarkdownBlock {
	return &MarkdownBlock{
		block: b,
		width: width,
	}
}

// ID returns an empty string as markdown blocks are layout-only.
func (m *MarkdownBlock) ID() string {
	return ""
}

// Update is a no-op since markdown blocks are not interactive.
func (m *MarkdownBlock) Update(msg tea.Msg) (Block, tea.Cmd) {
	return m, nil
}

// Answer returns nil as markdown blocks do not produce answers.
func (m *MarkdownBlock) Answer() any {
	return nil
}

// SetWidth updates the block's width and invalidates the cached rendering if changed.
func (m *MarkdownBlock) SetWidth(w int) {
	if m.width != w {
		m.width = w
		m.computed = false
		m.rendered = ""
	}
}

// Width returns the block's configured width.
func (m *MarkdownBlock) Width() int {
	return m.width
}

// View returns the rendered markdown wrapped in a severity-styled box.
func (m *MarkdownBlock) View(focused bool) string {
	if !m.computed {
		m.render()
	}
	return m.rendered
}

func (m *MarkdownBlock) render() {
	m.renderCount++
	m.computed = true

	sev := m.block.Severity
	if sev == "" {
		sev = schema.SevInfo
	}
	sevStyle, ok := theme.Severities[sev]
	if !ok {
		sevStyle = theme.Severities[schema.SevInfo]
	}

	badge := lipgloss.NewStyle().Bold(true).Foreground(sevStyle.Accent).Render(sevStyle.Badge)

	// Glamour word wrap width inside the box: subtract 2 for left border (1) and left padding (1)
	wrapWidth := m.width
	if wrapWidth <= 0 {
		wrapWidth = theme.DefaultCardWidth
	}
	wrapWidth -= 2
	if wrapWidth < 20 {
		wrapWidth = 20
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(wrapWidth),
	)

	var renderedMd string
	if err == nil {
		out, err := r.Render(m.block.Content)
		if err == nil {
			renderedMd = strings.Trim(out, "\n")
		} else {
			renderedMd = m.block.Content
		}
	} else {
		renderedMd = m.block.Content
	}

	var content string
	if renderedMd == "" {
		content = badge
	} else {
		content = badge + "\n" + renderedMd
	}

	box := theme.SeverityBox(sev, m.width)
	m.rendered = box.Render(content)
}
