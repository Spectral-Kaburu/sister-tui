# Implementation Plan: SpectreTUI

**Status: APPROVED — 2026-10-14** (revised after review: micro-style interaction + mouse, configurable card width, rename sister-tui → SpectreTUI)

Input: `CAPABILITY-MAP.md`, `SPEC-prompt-engine.md`, `SPEC-viewer.md`, `SPEC-daemon.md`, `docs/sister-tui-spec.md`. Tasks tracked in `tasks/todo.md`.

## Overview

Greenfield Go + Charm (bubbletea/lipgloss/glamour/bubbles) repo. Build the three modules in map order: prompt-engine (declarative JSON → centered card → structured JSON + exit codes), then viewer, then daemon. Micro-style interaction throughout (arrows/Tab/Enter + full mouse support). Each task is a vertical slice that leaves the binary runnable and the examples valid.

## Architecture Decisions

1. **TUI renders to `/dev/tty`, never stdout.** In pipeline mode stdout is a capture pipe — Bubble Tea frames written there would corrupt `RESULT=$(cat p.json | SpectreTUI)`. Open `/dev/tty` and pass it via `tea.WithInput`/`tea.WithOutput`; no tty → exit 1 with a stderr message. stdout carries only the result JSON; stderr only diagnostics. (The TUI must also *read* from the tty in pipeline mode, because stdin holds the payload.)
2. **One core `runPrompt(payload) (schema.Response, error)`** in `internal/model`, shared by pipeline/file CLI and the daemon; `main` and `daemon` are thin wrappers choosing the response destination (stdout / `--output` file / socket).
3. **Exit codes via a pure `exitCode(resp)` helper in `cmd/SpectreTUI`**; 130 is reserved for dismissal and schema validation rejects an action declaring it.
4. **Markdown is rendered once per block and cached** (glamour is slow); cache invalidated on terminal resize.
5. **Block interface** per `SPEC-prompt-engine.md` § Code Style; the root model owns only focus cycling, scrolling, and submit.
6. **Module path `github.com/Spectral-Kaburu/SpectreTUI`** (settled by the user); binary `SpectreTUI` from `cmd/SpectreTUI`.
7. **Micro-style interaction:** full mouse support (click-to-focus/select/activate, wheel scroll) via bubbletea mouse mode over the tty. Ctrl+C stays hard-wired dismiss — deliberate deviation from micro's Ctrl+C=copy, because abort must always work; no clipboard chords. Ctrl+Q is an additional viewer quit key.
8. **Configurable card width:** prompts may set `"width"` in the payload (default 100, validated ≥ 40); viewer takes `--width N`; both always clamped at render time to `min(width, termWidth-4)`, floor 40.

## Task List

### Phase 1: prompt-engine foundation
- [x] Task 1: Module scaffold + input/output schema
- [x] Task 2: Theme + markdown block
- [x] Task 3: Root model + CLI end-to-end (markdown card, action bar, mouse, /dev/tty, exit codes)

### Checkpoint A: end-to-end card works
- [ ] Tests pass, builds clean, examples run in both transports
- [ ] Review with human before proceeding

### Phase 2: remaining block types
- [ ] Task 4: single_select + multi_select
- [ ] Task 5: text_input + textarea (required blocking)
- [ ] Task 6: confirm block + review example

### Checkpoint B: all six block types usable
- [ ] Tests + build clean; human review

### Phase 3: prompt-engine polish
- [ ] Task 7: scrolling + resize adaptation

### Checkpoint C: prompt-engine spec criteria met
- [ ] SPEC-prompt-engine success criteria all verified

### Phase 4: viewer
- [ ] Task 8: viewer mode (`view` subcommand + `--view` + `--width`)

### Phase 5: daemon
- [ ] Task 9: Unix socket daemon
- [ ] Task 10: README + final sweep

### Checkpoint D: complete
- [ ] All three specs' success criteria verified; ready for review

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| TUI frames pollute piped stdout, breaking `RESULT=$(...)` | High | `/dev/tty` for TUI I/O from day 1 (Task 3); capture-stdout check at every checkpoint |
| Bubble Tea reads payload stdin instead of keystrokes | High | `tea.WithInput(tty)` whenever stdin is a pipe (Task 3) |
| Glamour re-renders every frame → laggy UI | Medium | render once per block, cache, invalidate on resize (Tasks 2, 7) |
| Terminal resize / narrow terminals | Medium | `WindowSizeMsg` → `min(width, termWidth-4)` floor 40 (Task 3); re-wrap on resize (Task 7) |
| Mouse reporting unavailable (weird terminals) | Low | enable bubbletea mouse mode on the tty; keys alone remain fully sufficient |
| Charm version churn (bubbletea API drift) | Low | pin via go.sum; stick to stable v1 APIs |

## Open Questions

— none remaining (module path settled: `github.com/Spectral-Kaburu/SpectreTUI`)

## Parallelization

Tasks 4–6 own distinct widget files but all touch `internal/model/prompt.go` and the examples — run sequentially (or assign explicit file ownership if parallel). Tasks 8 and 9 are independent of each other but both touch `cmd/SpectreTUI/main.go` — sequential is cheaper than coordinating.
