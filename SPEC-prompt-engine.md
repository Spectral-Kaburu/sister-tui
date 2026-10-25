# Spec: SpectreTUI — prompt-engine

Module: `prompt-engine` (base of `CAPABILITY-MAP.md`) · Design doc: `docs/sister-tui-spec.md`

## Objective

The core of `SpectreTUI`: an agnostic, declarative terminal prompt engine. A consumer — `nudged`, a shell script, cron, a git hook — supplies a JSON payload (title, blocks, action bar); the user interacts with a centered, rounded card (width configurable via the payload, default 100); `SpectreTUI` returns a structured JSON answer and a process exit code. Zero application-domain logic lives here: the engine renders exactly what it is given.

Acceptance criteria:
- `cat payload.json | SpectreTUI` → user completes the card → result JSON on stdout, exit code = chosen action's `exit_code`.
- `SpectreTUI --input p.json --output r.json` → result JSON written to `r.json`, nothing on stdout, same exit-code contract.
- Esc/Ctrl+C anywhere → exit 130 and the `dismissed` response shape (§5) delivered to the same destination as a normal result.
- Navigation is **micro-style**: arrows, Tab, Space, Enter, PgUp/PgDn, plus full mouse support (click-to-focus/select, wheel scroll). Vim keys are banned.

## Tech Stack

- Go 1.22+, single binary `cmd/SpectreTUI`
- `charmbracelet/bubbletea` — TUI runtime
- `charmbracelet/lipgloss` — styling, borders
- `charmbracelet/glamour` — markdown rendering
- `charmbracelet/bubbles` — textinput, textarea, viewport
- Module path: `github.com/Spectral-Kaburu/SpectreTUI`

## Commands

```
Build:    go build -o SpectreTUI ./cmd/SpectreTUI
Test:     go test ./...
Vet:      go vet ./...
Run file: go run ./cmd/SpectreTUI --input examples/morning.json --output /tmp/result.json
Run pipe: cat examples/review.json | go run ./cmd/SpectreTUI
```

## Project Structure

```
cmd/SpectreTUI/     → main.go: flag parsing, transport selection (stdin/file), exit-code handler
internal/schema/    → input.go (payload unmarshal + validation), output.go (response marshal)
internal/theme/     → styles.go: lipgloss palette, rounded borders, severity accents, focus ring
internal/blocks/    → block.go (Block interface), markdown.go, singleselect.go, multiselect.go,
                      textinput.go, textarea.go, confirm.go
internal/model/     → prompt.go: root Bubble Tea model — layout, focus cycling, scrolling, submit
examples/           → morning.json, missed_class.json, review.json
docs/               → design doc (exists)
```

`*_test.go` files live next to the code they test (Go convention).

## Code Style

Standard Go, gofmt-clean, table-driven tests, no global widget state. Blocks implement one small interface:

```go
// internal/blocks/block.go
type Block interface {
    ID() string           // "" for layout-only blocks (markdown)
    Update(msg tea.Msg)  // handle keys and mouse when focused
    View(focused bool) string
    Answer() any         // string | []string | bool | nil
}
```

- Schema validation collects errors at the trust boundary with typed messages: `fmt.Errorf("block %d: missing %q", i, "id")`.
- Nothing writes to stdout except the final result JSON — diagnostics go to stderr.

## Testing Strategy

Go stdlib `testing` only; no new test dependencies. `go test ./...` and `go vet ./...` must pass before any commit.

- `internal/schema/input_test.go` — table of valid/invalid payloads: missing `id`, unknown block type, bad `default`, duplicate ids, action `exit_code` = 130 (reserved), payload `width` < 40.
- `internal/model/prompt_test.go` — focus cycling across blocks and action bar; Enter routing; required-input submit blocking; scroll follows focus; mouse click-to-focus/activate.
- `internal/schema/output_test.go` — answer map shapes (`string` / `[]string` / `bool`), `dismissed` shape.
- Exit-code mapping asserted on a pure `exitCode(resp)` helper in `cmd/SpectreTUI`.
- Visual rendering is verified manually with the three example payloads (TUI frames are not unit-asserted).

## Boundaries

- **Always:** validate the payload before starting the TUI — bad JSON → stderr + exit 1, no screen flash; honor the exit-code contract exactly (action `exit_code`; 130 reserved for dismissal); print the result JSON only after the TUI exits; keep every color/style in `internal/theme`; treat payload JSON as untrusted input (validate ids, option refs in `default`, types, `width`); clamp card width to `min(width, termWidth-4)` with floor 40.
- **Ask first:** adding a block type; changing the severity set (info/tip/warning/critical/roast); changing exit-code semantics; adding any dependency; new CLI flags; changing the default width (100) or floor (40).
- **Never:** hardcode consumer/domain concepts (nudged, school, mornings); vim-style h/j/k/l or modal states; write anything to stdout besides the result JSON; block Esc/Ctrl+C dismissal (deliberate deviation from micro's Ctrl+C=copy: abort must always work; no clipboard chords).

## Success Criteria

- All three examples render as a centered rounded card: width = `min(payload "width" (default 100), termWidth-4)` with floor 40 — a payload `width` < 40 is a validation error. Header banner, blocks, pinned action bar.
- Every block type works: markdown (all 5 severities with correct accent + badge), single_select (arrows, `default` preselected), multi_select (Space toggles, `default` list prechecked), text_input, textarea, confirm (Left/Right or Tab).
- Tab/Shift+Tab cycle focus across blocks then the action bar; digits `1–9` activate the n-th action when the action bar is focused; Enter on the highlighted button activates it.
- Mouse works everywhere: click a block to focus it, an option to select/toggle it, a button to activate it; the wheel scrolls the block area.
- Taller-than-terminal prompts: PgUp/PgDn and wheel scroll the block area; focus moves auto-scroll the focused block into view; the action bar never leaves the screen.
- A `required` empty input blocks submit with an inline error on that block until filled.
- Submitting returns the action's `exit_code` and a §5-shaped result JSON (`action_id`, `exit_code`, `completed_at` RFC3339 local, `answers` keyed by block id: `string`/`[]string`/`bool`; markdown blocks contribute no answer).
- Dismissal returns 130 + `dismissed` JSON to the same destination (stdout in pipeline mode, `--output` file in file mode).
- Invalid JSON payload: no TUI, one-line error on stderr, exit 1.

## Open Questions

— none remaining (module path settled: `github.com/Spectral-Kaburu/SpectreTUI`)
