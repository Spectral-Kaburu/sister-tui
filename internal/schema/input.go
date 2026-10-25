// Package schema defines SpectreTUI's declarative contract: the JSON
// payload a consumer sends and the JSON response it gets back.
package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Block types.
const (
	TypeMarkdown     = "markdown"
	TypeSingleSelect = "single_select"
	TypeMultiSelect  = "multi_select"
	TypeTextInput    = "text_input"
	TypeTextArea     = "textarea"
	TypeConfirm      = "confirm"
)

// Markdown severities.
const (
	SevInfo     = "info"
	SevTip      = "tip"
	SevWarning  = "warning"
	SevCritical = "critical"
	SevRoast    = "roast"
)

// Dismissal contract, reported when the user aborts (Esc/Ctrl+C).
// Actions may not declare the reserved exit code.
const (
	DismissedActionID = "dismissed"
	DismissedExitCode = 130
)

// MinWidth is the smallest card width a payload may request.
const MinWidth = 40

var severities = map[string]bool{
	SevInfo: true, SevTip: true, SevWarning: true, SevCritical: true, SevRoast: true,
}

// Payload is one prompt screen: header, blocks, action bar.
type Payload struct {
	Title    string   `json:"title"`
	Subtitle string   `json:"subtitle,omitempty"`
	Width    int      `json:"width,omitempty"` // card width; 0 = default 100; must be >= MinWidth
	Blocks   []Block  `json:"blocks"`
	Actions  []Action `json:"actions"`
}

// Block is one element of a screen. Which fields matter depends on
// Type; Validate enforces that per type.
type Block struct {
	Type        string   `json:"type"`
	ID          string   `json:"id,omitempty"`
	Label       string   `json:"label,omitempty"`
	Severity    string   `json:"severity,omitempty"`    // markdown
	Content     string   `json:"content,omitempty"`     // markdown
	Options     []Option `json:"options,omitempty"`     // selects
	Default     any      `json:"default,omitempty"`     // string | []string | bool
	Placeholder string   `json:"placeholder,omitempty"` // text_input
	Required    bool     `json:"required,omitempty"`    // text_input/textarea
}

// Option is one choice of a select block.
type Option struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// Action is one button of the action bar.
type Action struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	ExitCode int    `json:"exit_code"`
	Primary  bool   `json:"is_primary,omitempty"`
}

// Parse decodes and validates a payload. Payload JSON is untrusted
// input: every problem is reported before any TUI runs.
func Parse(data []byte) (Payload, error) {
	var p Payload
	if err := json.Unmarshal(data, &p); err != nil {
		return Payload{}, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := p.Validate(); err != nil {
		return Payload{}, err
	}
	return p, nil
}

// Validate checks a decoded payload and normalizes defaults (markdown
// severity → info). All problems are joined into one error.
func (p *Payload) Validate() error {
	var errs []error
	if strings.TrimSpace(p.Title) == "" {
		errs = append(errs, fmt.Errorf("payload: missing %q", "title"))
	}
	if p.Width != 0 && p.Width < MinWidth {
		errs = append(errs, fmt.Errorf("payload: %q must be >= %d (got %d)", "width", MinWidth, p.Width))
	}
	if len(p.Actions) == 0 {
		errs = append(errs, errors.New("payload: at least one action is required"))
	}

	blockIDs := map[string]bool{}
	for i := range p.Blocks {
		b := &p.Blocks[i]
		errs = append(errs, b.validate(i)...)
		if b.ID != "" {
			if blockIDs[b.ID] {
				errs = append(errs, fmt.Errorf("block %d: duplicate id %q", i, b.ID))
			}
			blockIDs[b.ID] = true
		}
	}

	actionIDs := map[string]bool{}
	for i := range p.Actions {
		a := &p.Actions[i]
		switch {
		case a.ID == "":
			errs = append(errs, fmt.Errorf("action %d: missing %q", i, "id"))
		case actionIDs[a.ID]:
			errs = append(errs, fmt.Errorf("action %d: duplicate id %q", i, a.ID))
		}
		actionIDs[a.ID] = true
		if strings.TrimSpace(a.Label) == "" {
			errs = append(errs, fmt.Errorf("action %q: missing %q", a.ID, "label"))
		}
		if a.ExitCode == DismissedExitCode {
			errs = append(errs, fmt.Errorf("action %q: exit_code %d is reserved for dismissal", a.ID, DismissedExitCode))
		}
	}
	return errors.Join(errs...)
}

func (b *Block) validate(i int) []error {
	var errs []error
	add := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("block %d: "+format, append([]any{i}, args...)...))
	}
	switch b.Type {
	case TypeMarkdown:
		switch {
		case b.Severity == "":
			b.Severity = SevInfo // normalize: absent severity renders as info
		case !severities[b.Severity]:
			add("unknown severity %q", b.Severity)
		}
	case TypeSingleSelect, TypeMultiSelect:
		errs = append(errs, b.validateInteractive(i)...)
		if len(b.Options) == 0 {
			add("at least one option is required")
		}
		optionIDs := map[string]bool{}
		for j, o := range b.Options {
			switch {
			case o.ID == "":
				add("option %d: missing %q", j, "id")
			case optionIDs[o.ID]:
				add("option %d: duplicate id %q", j, o.ID)
			}
			optionIDs[o.ID] = true
		}
		errs = append(errs, b.validateDefault(i)...)
	case TypeTextInput, TypeTextArea:
		errs = append(errs, b.validateInteractive(i)...)
		if b.Default != nil {
			if _, ok := b.Default.(string); !ok {
				add("default must be a string for %s", b.Type)
			}
		}
	case TypeConfirm:
		errs = append(errs, b.validateInteractive(i)...)
		if b.Default != nil {
			if _, ok := b.Default.(bool); !ok {
				add("default must be a bool for %s", b.Type)
			}
		}
	default:
		add("unknown type %q", b.Type)
	}
	return errs
}

func (b *Block) validateInteractive(i int) []error {
	var errs []error
	if strings.TrimSpace(b.ID) == "" {
		errs = append(errs, fmt.Errorf("block %d: missing %q", i, "id"))
	}
	if strings.TrimSpace(b.Label) == "" {
		errs = append(errs, fmt.Errorf("block %d: missing %q", i, "label"))
	}
	return errs
}

// validateDefault checks that default values reference real options.
func (b *Block) validateDefault(i int) []error {
	if b.Default == nil {
		return nil
	}
	known := map[string]bool{}
	for _, o := range b.Options {
		known[o.ID] = true
	}
	wrong := func(format string, args ...any) error {
		return fmt.Errorf("block %d: "+format, append([]any{i}, args...)...)
	}
	switch b.Type {
	case TypeSingleSelect:
		id, ok := b.Default.(string)
		if !ok {
			return []error{wrong("default must be an option id")}
		}
		if !known[id] {
			return []error{wrong("default %q is not an option", id)}
		}
	case TypeMultiSelect:
		var items []string
		switch v := b.Default.(type) {
		case []string:
			items = v
		case []any:
			for _, item := range v {
				id, ok := item.(string)
				if !ok {
					return []error{wrong("default must be an array of option ids")}
				}
				items = append(items, id)
			}
		default:
			return []error{wrong("default must be an array of option ids")}
		}
		for _, id := range items {
			if !known[id] {
				return []error{wrong("default %q is not an option", id)}
			}
		}
	}
	return nil
}
