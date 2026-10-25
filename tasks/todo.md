# Task List: SpectreTUI

**Status: APPROVED — 2026-10-14** · Index: `tasks/plan.md` · Specs: `SPEC-prompt-engine.md`, `SPEC-viewer.md`, `SPEC-daemon.md`

---

## Task 1: Module scaffold + input/output schema

**Description:** `go mod init github.com/Spectral-Kaburu/SpectreTUI`, then `internal/schema`: Payload/Block/Option/Action types with unmarshal + validation (title required; ≥1 action; unknown block type; missing `id`/`label` on interactive blocks; duplicate block and option ids; `default` values present in options and correctly typed per block type; action `exit_code` = 130 rejected; payload `width` ≥ 40 when set; markdown severity normalized to `info` when absent, validated against the 5 known values), and Response marshaling (§5 shape: `action_id`, `exit_code`, `completed_at` RFC3339, `answers`; `dismissed` shape). Stdlib only — no Charm deps yet.

**Acceptance criteria:**
- [ ] Valid/invalid payload table tests pass (every rule listed above)
- [ ] Response marshal produces §5 field names; answer values are `string` / `[]string` / `bool`; dismissed answers serialize as `{}` (not `null`)

**Verification:**
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean

**Dependencies:** None
**Files:** `go.mod`, `internal/schema/input.go`, `internal/schema/input_test.go`, `internal/schema/output.go`, `internal/schema/output_test.go`
**Scope:** S

---

## Task 2: Theme + markdown block

**Description:** `internal/theme/styles.go` (card width rule `min(payloadWidth or 100, termWidth-4)` floor 40, rounded border, five severity accents + badges, focus ring) and `internal/blocks/`: `block.go` (Block interface) + `markdown.go` — glamour render wrapped in the severity-bordered box, rendered once and cached per block. Pulls the Charm deps (`bubbletea`, `lipgloss`, `glamour`, `bubbles`).

**Acceptance criteria:**
- [ ] All 5 severities render a styled box with the right badge: ℹ INFO, ✔ TIP, ⚠ WARNING, ✖ CRITICAL, 🔥 SISTER ROAST
- [ ] Repeated `View()` calls don't re-run glamour (cache); rendered width ≤ card width

**Verification:**
- [ ] `go test ./...` (badge presence, non-empty, width)
- [ ] `go vet ./...` clean

**Dependencies:** Task 1
**Files:** `internal/theme/styles.go`, `internal/blocks/block.go`, `internal/blocks/markdown.go`, `internal/blocks/markdown_test.go`
**Scope:** M

---

## Task 3: Root model + CLI end-to-end

**Description:** `internal/model/prompt.go`: card layout (header, block stack, pinned action bar), Tab/Shift+Tab focus cycling, Enter on highlighted button activates, digits 1–9 activate the n-th action when the action bar is focused, Esc/Ctrl+C → dismissed, `WindowSizeMsg` width adaptation, **mouse mode** (click a block to focus, click an option/button to select/activate, wheel scrolls). `cmd/SpectreTUI/main.go`: `--input`/`--output` flags, stdin payload when piped, **TUI I/O via `/dev/tty`** (plan decision 1), pure `exitCode(resp)` helper, result to stdout (pipeline) or `--output` file, invalid payload → stderr + exit 1. Three examples created with markdown + actions only (they grow in Tasks 4–6).

**Acceptance criteria:**
- [x] `--input examples/morning.json --output /tmp/r.json`: card renders, Enter → exit 0, `/tmp/r.json` matches §5; stdout stays empty
- [x] `cat examples/review.json | go run ./cmd/SpectreTUI > /tmp/r.json`: TUI on the terminal, /tmp/r.json contains ONLY the result JSON
- [x] Mouse: click focuses/selects/activates; wheel scrolls; keys alone still work everywhere
- [x] Esc or Ctrl+C → exit 130 + dismissed JSON to the same destination
- [x] Invalid JSON → one stderr line, exit 1, no TUI
- [x] Terminal narrowed below the payload width → card shrinks (floor 40); payload `"width"` is honored

**Verification:**
- [x] `go test ./...` (focus cycling, submit, dismiss, exitCode mapping, click-to-focus)
- [x] Manual: both transports × all examples; resize check; pipe-capture check; mouse check

**Dependencies:** Tasks 1, 2
**Files:** `internal/model/prompt.go`, `internal/model/prompt_test.go`, `cmd/SpectreTUI/main.go`, `cmd/SpectreTUI/main_test.go`, `examples/morning.json`, `examples/missed_class.json`, `examples/review.json`
**Scope:** M-L (4 code files + 3 data files)

---

## Checkpoint A: end-to-end card works

- [ ] All tests pass, build + vet clean
- [ ] Markdown card works in both transports with correct exit codes
- [ ] Review with human before proceeding

---

## Task 4: single_select + multi_select

**Description:** Radio list (Up/Down moves, `default` preselected) and checkbox list (Up/Down moves, Space toggles, `default` list prechecked); mouse click selects (single) / toggles (multi). Answers: option id string / []string. Wire both into the focus cycle. Add single_select to `missed_class.json`, multi_select to `morning.json`.

**Acceptance criteria:**
- [ ] Arrows move selection; Space toggles only multi_select; click does the same; Enter never submits from a list (only from the action bar)
- [ ] Defaults preselected; answers carry the selected ids only

**Verification:**
- [ ] `go test ./...` (move/toggle/default/click tests per widget)
- [ ] Manual: run both updated examples

**Dependencies:** Task 3
**Files:** `internal/blocks/singleselect.go`, `internal/blocks/multiselect.go`, their tests, `internal/model/prompt.go`, `examples/missed_class.json`, `examples/morning.json`
**Scope:** M

---

## Task 5: text_input + textarea

**Description:** `bubbles` textinput/textarea wrappers; `required` semantics: submit with an empty required input is blocked with an inline error on that block. Add inputs to `morning.json` and `missed_class.json`.

**Acceptance criteria:**
- [ ] Typing works when focused; Enter inserts a newline only in textarea; Tab leaves the field
- [ ] Empty required input blocks submit with a visible inline error; optional empty input answers `""`

**Verification:**
- [ ] `go test ./...` (required-blocking model test, widget wiring)
- [ ] Manual: example with a required field

**Dependencies:** Task 3
**Files:** `internal/blocks/textinput.go`, `internal/blocks/textarea.go`, their tests, `internal/model/prompt.go`, example updates
**Scope:** M

---

## Task 6: confirm block + review example

**Description:** Yes/No binary block navigated by Left/Right or Tab; click selects too. `Answer()` returns bool. Finalize `examples/review.json` (markdown + confirm + textarea) so the three examples together exercise all six block types.

**Acceptance criteria:**
- [ ] Left/Right flips the choice; answer is `true`/`false`
- [ ] Each block type appears in at least one example

**Verification:**
- [ ] `go test ./...`
- [ ] Manual: run all three examples

**Dependencies:** Tasks 4, 5
**Files:** `internal/blocks/confirm.go`, its test, `examples/review.json`
**Scope:** S

---

## Checkpoint B: all six block types usable

- [ ] Tests + build clean
- [ ] Human review before scrolling work

---

## Task 7: Scrolling + resize

**Description:** When blocks exceed terminal height: PgUp/PgDn and mouse wheel scroll the block area, focus moves auto-scroll the focused block into view, action bar always pinned. Markdown render cache invalidated on resize (re-wrap to new width).

**Acceptance criteria:**
- [ ] Tall prompt scrolls (keys + wheel); focused block stays visible; action bar never leaves the screen
- [ ] Resizing the terminal re-wraps markdown to the new card width

**Verification:**
- [ ] `go test ./...` (scroll-offset model tests)
- [ ] Manual: tall synthetic payload + live terminal resize + wheel

**Dependencies:** Task 6
**Files:** `internal/model/prompt.go`, `internal/model/prompt_test.go`, `internal/blocks/markdown.go`
**Scope:** M

---

## Checkpoint C: prompt-engine spec complete

- [ ] SPEC-prompt-engine success criteria all verified

---

## Task 8: Viewer mode

**Description:** `SpectreTUI view <file>` and `--view` (stdin) rendering one document in the same centered card via a bubbletea viewport: ↑/↓ line scroll, PgUp/PgDn page scroll, wheel scroll, Enter/Esc/`q`/Ctrl+Q → exit 0 with zero stdout/stderr output. `--width N` configures the card width (default 100, clamped to terminal, floor 40).

**Acceptance criteria:**
- [ ] File and stdin render identically; scrolling reaches document end and stops
- [ ] Enter/Esc/`q`/Ctrl+Q write no bytes, exit 0; missing file → stderr + exit 1, no TUI
- [ ] `--width` is respected after clamping

**Verification:**
- [ ] `go test ./...` (viewer model key mapping incl. Ctrl+Q, document load)
- [ ] Manual: `view docs/sister-tui-spec.md` and `cat README.md | SpectreTUI --view --width 96`; confirm silence with `> /dev/null` capture

**Dependencies:** Task 7 (per build order)
**Files:** `cmd/SpectreTUI/main.go`, `internal/model/viewer.go`, `internal/model/viewer_test.go`
**Scope:** S

---

## Task 9: Unix socket daemon

**Description:** `SpectreTUI daemon --socket <path>`: UnixListener with 0600 perms, stale socket file replaced on startup, removed on shutdown. Per connection: decode one JSON payload → `runPrompt` → write one response JSON → close. FIFO queue; prompts render strictly one at a time. Server takes `runPrompt` as a parameter (test stub).

**Acceptance criteria:**
- [ ] `cat examples/review.json | socat - UNIX-CONNECT:/tmp/SpectreTUI.sock` prints the response JSON (incl. dismissed shape on Esc/Ctrl+C)
- [ ] Two queued connections handled in order; Ctrl+C removes the socket file

**Verification:**
- [ ] `go test ./...` (stubbed-runPrompt integration: round-trip, FIFO, socket cleanup)
- [ ] Manual: live daemon + socat client

**Dependencies:** Task 8 (main.go ownership — sequential)
**Files:** `internal/daemon/server.go`, `internal/daemon/server_test.go`, `cmd/SpectreTUI/main.go`
**Scope:** M

---

## Task 10: README + final sweep

**Description:** README.md: install, the four modes with copy-paste examples, key table (micro-style), exit codes, payload schema summary (link to the design doc). Full verification pass over all specs.

**Acceptance criteria:**
- [ ] README matches actual behavior; every documented command runs as written
- [ ] `go test ./...`, `go vet ./...`, `go build` clean; all examples work in all transports

**Verification:**
- [ ] Full pass of every checkpoint's commands

**Dependencies:** Tasks 8, 9
**Files:** `README.md`
**Scope:** XS

---

## Checkpoint D: complete

- [ ] All three specs' success criteria verified
- [ ] Ready for review
