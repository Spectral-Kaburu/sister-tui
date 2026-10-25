# Capability Map: SpectreTUI

Source design doc: `docs/sister-tui-spec.md` (historical — written under the original name `sister-tui`)

**Status: APPROVED — 2026-10-14** (revised after review: micro-style interaction + mouse, configurable card width, rename sister-tui → SpectreTUI, module path settled)

| Module id | Responsibility | Depends on |
|---|---|---|
| prompt-engine | Declarative prompt runner: JSON input/output schema, theme + severity styles, six block widgets, root Bubble Tea model with micro-style navigation (keys + full mouse), whole-card scrolling, configurable width, pipeline (stdin→stdout) + file (`--input`/`--output`) transports, exit-code contract | — |
| viewer | Read-only Glamour markdown viewer (`SpectreTUI view <file>` / `--view`): zero output, clean exit 0 | prompt-engine |
| daemon | Persistent Unix domain socket server: one payload → one response per connection, FIFO queue, runs the prompt engine in a dedicated terminal | prompt-engine |

Build order: **prompt-engine → viewer → daemon**

`viewer` and `daemon` are independent of each other; both reuse `prompt-engine`'s theme, markdown renderer, and root model.

Module specs: `SPEC-prompt-engine.md`, `SPEC-viewer.md`, `SPEC-daemon.md`
