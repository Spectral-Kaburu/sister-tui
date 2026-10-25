package model

import (
	"fmt"
	"os"
	"strings"

	"github.com/Spectral-Kaburu/SpectreTUI/internal/blocks"
	"github.com/Spectral-Kaburu/SpectreTUI/internal/schema"
	"github.com/Spectral-Kaburu/SpectreTUI/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PromptModel is the root Bubble Tea model managing card layout, focus,
// keyboard/mouse interaction, and response submission.
type PromptModel struct {
	payload    schema.Payload
	blocks     []blocks.Block
	actions    []schema.Action
	focusIndex int
	width      int
	termWidth  int
	termHeight int
	response   *schema.Response
	err        error
	quitting   bool
}

type blockHitZone struct {
	index int
	minY  int
	maxY  int
}

type actionHitZone struct {
	index int
	minX  int
	maxX  int
	minY  int
	maxY  int
}

// NewPromptModel initializes a PromptModel from a validated schema.Payload.
func NewPromptModel(p schema.Payload) PromptModel {
	effWidth := p.Width
	if effWidth <= 0 {
		effWidth = theme.DefaultCardWidth
	}
	contentW := effWidth - 4
	if contentW < theme.MinCardWidth-4 {
		contentW = theme.MinCardWidth - 4
	}

	var blks []blocks.Block
	for _, b := range p.Blocks {
		switch b.Type {
		case schema.TypeMarkdown:
			blks = append(blks, blocks.NewMarkdownBlock(b, contentW))
		// Phase 2 will instantiate single_select, multi_select, text_input, etc.
		}
	}

	hasInteractive := false
	firstInteractive := -1
	for i, b := range blks {
		if b.ID() != "" {
			hasInteractive = true
			if firstInteractive == -1 {
				firstInteractive = i
			}
		}
	}

	primaryActionIndex := 0
	for i, a := range p.Actions {
		if a.Primary {
			primaryActionIndex = i
			break
		}
	}

	focusIdx := 0
	if !hasInteractive {
		focusIdx = len(blks) + primaryActionIndex
	} else {
		focusIdx = firstInteractive
	}

	return PromptModel{
		payload:    p,
		blocks:     blks,
		actions:    p.Actions,
		focusIndex: focusIdx,
		width:      effWidth,
	}
}

// RunPrompt executes the prompt model using /dev/tty for I/O.
// Returns the completed Response or an error.
func RunPrompt(payload schema.Payload) (schema.Response, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return schema.Response{}, fmt.Errorf("opening /dev/tty: %w", err)
	}
	defer tty.Close()

	m := NewPromptModel(payload)
	p := tea.NewProgram(
		m,
		tea.WithInput(tty),
		tea.WithOutput(tty),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	finalModel, err := p.Run()
	if err != nil {
		return schema.Response{}, fmt.Errorf("running prompt: %w", err)
	}

	switch v := finalModel.(type) {
	case PromptModel:
		if v.err != nil {
			return schema.Response{}, v.err
		}
		if v.response == nil {
			return schema.Dismissed(), nil
		}
		return *v.response, nil
	case *PromptModel:
		if v.err != nil {
			return schema.Response{}, v.err
		}
		if v.response == nil {
			return schema.Dismissed(), nil
		}
		return *v.response, nil
	default:
		return schema.Response{}, fmt.Errorf("unexpected model type %T", finalModel)
	}
}

// Init implements tea.Model.
func (m PromptModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m PromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			resp := schema.Dismissed()
			m.response = &resp
			m.quitting = true
			return m, tea.Quit

		case tea.KeyTab:
			m.nextFocus()
			return m, nil

		case tea.KeyShiftTab:
			m.prevFocus()
			return m, nil

		case tea.KeyEnter:
			if m.isActionBarFocused() {
				actIdx := m.focusIndex - len(m.blocks)
				action := m.actions[actIdx]
				answers := m.collectAnswers()
				resp := schema.NewResponse(action, answers)
				m.response = &resp
				m.quitting = true
				return m, tea.Quit
			}
			if m.focusIndex < len(m.blocks) {
				b, cmd := m.blocks[m.focusIndex].Update(msg)
				m.blocks[m.focusIndex] = b
				return m, cmd
			}

		case tea.KeyLeft:
			if m.isActionBarFocused() {
				actIdx := m.focusIndex - len(m.blocks)
				numActs := len(m.actions)
				if numActs > 0 {
					actIdx = (actIdx - 1 + numActs) % numActs
					m.focusIndex = len(m.blocks) + actIdx
					return m, nil
				}
			}
			if m.focusIndex < len(m.blocks) {
				b, cmd := m.blocks[m.focusIndex].Update(msg)
				m.blocks[m.focusIndex] = b
				return m, cmd
			}

		case tea.KeyRight:
			if m.isActionBarFocused() {
				actIdx := m.focusIndex - len(m.blocks)
				numActs := len(m.actions)
				if numActs > 0 {
					actIdx = (actIdx + 1) % numActs
					m.focusIndex = len(m.blocks) + actIdx
					return m, nil
				}
			}
			if m.focusIndex < len(m.blocks) {
				b, cmd := m.blocks[m.focusIndex].Update(msg)
				m.blocks[m.focusIndex] = b
				return m, cmd
			}

		case tea.KeyRunes:
			s := msg.String()
			if len(s) == 1 && s[0] >= '1' && s[0] <= '9' {
				digit := int(s[0] - '0')
				if digit >= 1 && digit <= len(m.actions) {
					if m.isActionBarFocused() || (m.focusIndex < len(m.blocks) && m.blocks[m.focusIndex].ID() == "") {
						action := m.actions[digit-1]
						answers := m.collectAnswers()
						resp := schema.NewResponse(action, answers)
						m.response = &resp
						m.quitting = true
						return m, tea.Quit
					}
				}
			}
		}

		// Delegate any other key to focused block
		if m.focusIndex < len(m.blocks) {
			b, cmd := m.blocks[m.focusIndex].Update(msg)
			m.blocks[m.focusIndex] = b
			return m, cmd
		}

	case tea.MouseMsg:
		if (msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft) || msg.Type == tea.MouseLeft {
			return m.handleMouseClick(msg)
		}
		if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown || tea.MouseEvent(msg).IsWheel() {
			if m.focusIndex < len(m.blocks) {
				b, cmd := m.blocks[m.focusIndex].Update(msg)
				m.blocks[m.focusIndex] = b
				return m, cmd
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.width = theme.CardWidth(m.payload.Width, m.termWidth)
		contentW := m.contentWidth()
		for _, b := range m.blocks {
			if ws, ok := b.(interface{ SetWidth(int) }); ok {
				ws.SetWidth(contentW)
			}
		}
		return m, nil
	}

	return m, nil
}

// View implements tea.Model.
func (m PromptModel) View() string {
	if m.quitting {
		return ""
	}
	rendered, _, _ := m.layout()
	return rendered
}

// Response returns the user response, or nil if still interactive.
func (m PromptModel) Response() *schema.Response {
	return m.response
}

// FocusIndex returns current focus index.
func (m PromptModel) FocusIndex() int {
	return m.focusIndex
}

// Width returns the computed card width.
func (m PromptModel) Width() int {
	return m.width
}

// Blocks returns the blocks list.
func (m PromptModel) Blocks() []blocks.Block {
	return m.blocks
}

// Actions returns the actions list.
func (m PromptModel) Actions() []schema.Action {
	return m.actions
}

// Err returns any runtime error encountered.
func (m PromptModel) Err() error {
	return m.err
}

func (m PromptModel) totalItems() int {
	return len(m.blocks) + len(m.actions)
}

func (m *PromptModel) nextFocus() {
	tot := m.totalItems()
	if tot > 0 {
		m.focusIndex = (m.focusIndex + 1) % tot
	}
}

func (m *PromptModel) prevFocus() {
	tot := m.totalItems()
	if tot > 0 {
		m.focusIndex = (m.focusIndex - 1 + tot) % tot
	}
}

func (m PromptModel) isActionBarFocused() bool {
	return m.focusIndex >= len(m.blocks) && m.focusIndex < m.totalItems()
}

func (m PromptModel) contentWidth() int {
	w := m.width - 4
	if w < theme.MinCardWidth-4 {
		w = theme.MinCardWidth - 4
	}
	return w
}

func (m PromptModel) collectAnswers() map[string]any {
	answers := make(map[string]any)
	for _, b := range m.blocks {
		if b.ID() != "" && b.Answer() != nil {
			answers[b.ID()] = b.Answer()
		}
	}
	return answers
}

func (m PromptModel) layout() (string, []blockHitZone, []actionHitZone) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f8fafc"))
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#94a3b8"))

	var headerParts []string
	if m.payload.Title != "" {
		headerParts = append(headerParts, titleStyle.Render(m.payload.Title))
	}
	if m.payload.Subtitle != "" {
		headerParts = append(headerParts, subtitleStyle.Render(m.payload.Subtitle))
	}
	headerView := lipgloss.JoinVertical(lipgloss.Left, headerParts...)

	var blockViews []string
	for i, b := range m.blocks {
		isFocused := (m.focusIndex == i)
		blockViews = append(blockViews, b.View(isFocused))
	}

	var buttonViews []string
	for i, a := range m.actions {
		isFocused := (m.focusIndex == len(m.blocks)+i)
		btnLabel := a.Label
		if !strings.Contains(btnLabel, "[") {
			btnLabel = fmt.Sprintf("[%d] %s", i+1, a.Label)
		}
		var btnStyle lipgloss.Style
		if isFocused {
			if a.Primary {
				btnStyle = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#6366f1")).
					Background(lipgloss.Color("#6366f1")).
					Foreground(lipgloss.Color("#ffffff")).
					Bold(true).
					Padding(0, 1)
			} else {
				btnStyle = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#ffffff")).
					Background(lipgloss.Color("#334155")).
					Foreground(lipgloss.Color("#ffffff")).
					Bold(true).
					Padding(0, 1)
			}
		} else {
			if a.Primary {
				btnStyle = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#6366f1")).
					Foreground(lipgloss.Color("#a5b4fc")).
					Bold(true).
					Padding(0, 1)
			} else {
				btnStyle = lipgloss.NewStyle().
					BorderStyle(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("#475569")).
					Foreground(lipgloss.Color("#94a3b8")).
					Padding(0, 1)
			}
		}
		if i < len(m.actions)-1 {
			btnStyle = btnStyle.MarginRight(1)
		}
		buttonViews = append(buttonViews, btnStyle.Render(btnLabel))
	}
	actionBarView := lipgloss.JoinHorizontal(lipgloss.Top, buttonViews...)

	var sections []string
	if headerView != "" {
		sections = append(sections, headerView)
	}
	if len(blockViews) > 0 {
		sections = append(sections, "", strings.Join(blockViews, "\n\n"))
	}
	if actionBarView != "" {
		sections = append(sections, "", actionBarView)
	}
	cardContent := lipgloss.JoinVertical(lipgloss.Left, sections...)

	card := theme.CardBorder().Width(m.width).Render(cardContent)

	cardW := lipgloss.Width(card)
	cardH := lipgloss.Height(card)
	startX := 0
	startY := 0
	if m.termWidth > cardW {
		startX = (m.termWidth - cardW) / 2
	}
	if m.termHeight > cardH {
		startY = (m.termHeight - cardH) / 2
	}

	curY := startY + 1 // top border is 1 line
	if headerView != "" {
		curY += lipgloss.Height(headerView)
	}

	var bZones []blockHitZone
	if len(blockViews) > 0 {
		curY += 1 // blank line before blocks
		for i, bv := range blockViews {
			bH := lipgloss.Height(bv)
			bZones = append(bZones, blockHitZone{
				index: i,
				minY:  curY,
				maxY:  curY + bH,
			})
			curY += bH
			if i < len(blockViews)-1 {
				curY += 2 // \n\n between blocks is 2 lines
			}
		}
	}

	var aZones []actionHitZone
	if actionBarView != "" {
		curY += 1 // blank line before action bar
		actH := lipgloss.Height(actionBarView)
		curX := startX + 2 // 1 left border + 1 left padding
		for j, bv := range buttonViews {
			bW := lipgloss.Width(bv)
			aZones = append(aZones, actionHitZone{
				index: j,
				minX:  curX,
				maxX:  curX + bW,
				minY:  curY,
				maxY:  curY + actH,
			})
			curX += bW
		}
	}

	var finalView string
	if m.termWidth > 0 && m.termHeight > 0 {
		finalView = lipgloss.Place(m.termWidth, m.termHeight, lipgloss.Center, lipgloss.Center, card)
	} else {
		finalView = card
	}

	return finalView, bZones, aZones
}

func (m *PromptModel) handleMouseClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	_, bZones, aZones := m.layout()

	// 1. Check action buttons
	for _, az := range aZones {
		if msg.X >= az.minX && msg.X < az.maxX && msg.Y >= az.minY && msg.Y < az.maxY {
			m.focusIndex = len(m.blocks) + az.index
			action := m.actions[az.index]
			answers := m.collectAnswers()
			resp := schema.NewResponse(action, answers)
			m.response = &resp
			m.quitting = true
			return *m, tea.Quit
		}
	}

	// 2. Check blocks
	for _, bz := range bZones {
		if msg.Y >= bz.minY && msg.Y < bz.maxY {
			cardW := m.width
			startX := 0
			if m.termWidth > cardW {
				startX = (m.termWidth - cardW) / 2
			}
			if msg.X >= startX && msg.X < startX+cardW {
				m.focusIndex = bz.index
				b, cmd := m.blocks[bz.index].Update(msg)
				m.blocks[bz.index] = b
				return *m, cmd
			}
		}
	}

	return *m, nil
}
