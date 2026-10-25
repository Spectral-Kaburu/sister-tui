# Spec: SpectreTUI — viewer

Module: `viewer` · Depends on: `prompt-engine` (theme, markdown renderer) · Parent: `CAPABILITY-MAP.md`

## Objective

A read-only markdown reader using the same centered card (width configurable via `--width`, default 100): read notes, changelogs, or piped text with Glamour styling, zero interaction cost, zero output.

Acceptance criteria:
- `SpectreTUI view notes.md` renders the file styled in the card viewport.
- `cat file.md | SpectreTUI --view` renders piped stdin the same way.
- ↑/↓ scroll line-by-line, PgUp/PgDn page-by-page, mouse wheel scrolls; Enter, Esc, `q`, or Ctrl+Q (micro's quit key) exits with code 0 and zero bytes written to stdout/stderr.

## Tech Stack

Same as `SPEC-prompt-engine.md` — bubbletea `viewport` + the shared Glamour renderer. No new dependencies.

## Commands

```
Build:  go build -o SpectreTUI ./cmd/SpectreTUI
Test:   go test ./...
Run:    go run ./cmd/SpectreTUI view docs/sister-tui-spec.md
Pipe:   cat README.md | go run ./cmd/SpectreTUI --view --width 96
```

## Project Structure

```
cmd/SpectreTUI/main.go    → `view` subcommand + `--view`/`--width` flag wiring (stdin fallback)
internal/model/viewer.go  → viewer model: viewport, scroll keys, wheel, quit keys
```

Reuses `internal/theme` and the markdown rendering path from `internal/blocks/markdown.go`.

## Code Style

As `SPEC-prompt-engine.md` § Code Style.

## Testing Strategy

Stdlib test: a temp `.md` file and piped stdin both produce non-empty styled content; key→quit/scroll mapping (incl. Ctrl+Q) asserted on the model's `Update`; wheel scrolls. Visual check is manual.

## Boundaries

- **Always:** exit 0 with zero output on Enter/Esc/`q`/Ctrl+Q; accept either a file argument or stdin (a missing/unreadable file → stderr + exit 1, no TUI); clamp `--width` to the terminal (`min(width, termWidth-4)`, floor 40).
- **Ask first:** adding keys or navigation in viewer mode.
- **Never:** prompt for answers, emit output, block Esc/`q`/Ctrl+Q, or render anything beyond the one document.

## Success Criteria

- File and stdin paths render identically; scrolling (keys and wheel) reaches the document end and stops; quitting writes no bytes and exits 0.
- `--width N` is respected, clamped to the terminal width.

## Open Questions

— none
