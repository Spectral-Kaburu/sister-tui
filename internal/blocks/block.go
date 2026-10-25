package blocks

import tea "github.com/charmbracelet/bubbletea"

// Block is one interactive or layout element of a prompt card.
type Block interface {
	ID() string                          // empty string for layout-only blocks (markdown)
	Update(msg tea.Msg) (Block, tea.Cmd) // handle keys/mouse when focused
	View(focused bool) string
	Answer() any // string | []string | bool | nil
}
