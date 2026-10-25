# sister-tui — Universal Declarative Terminal UI & Interactive Prompt Engine

> **Target:** Standalone Go Repository (`sister-tui`)  
> **Framework:** Charm CLI ([Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), [Glamour](https://github.com/charmbracelet/glamour), [Bubbles](https://github.com/charmbracelet/bubbles))  
> **Role:** Agnostic, universal interactive prompt engine. Any desktop daemon (`nudged`), CLI script, cron job, or git hook can invoke it with a declarative JSON payload and receive structured user responses.

---

## 1. Design Philosophy

`sister-tui` does **not** hardcode application domain concepts (no built-in "morning mode" or "school class" logic).  
Instead, it is a **declarative terminal widget runner** (akin to a styled, modern terminal equivalent of `zenity` or interactive agent harnesses like `agy` and `claude-code`).

### Core Features:
- **90–100 Column Centered Container:** The UI renders as a focused, beautifully centered card of 90–100 columns width with rounded borders (adapting down gracefully if terminal width is smaller).
- **Dual Mode (Prompts + Pure Markdown Viewer):**
  1. *Prompt Mode:* Interactive questionnaires outputting structured JSON answers.
  2. *Viewer Mode:* Specialized read-only Markdown viewer with zero output, styled severity callouts, and clean exit.
- **Rich Markdown with Severity Styling:** Renders Markdown blocks with colored callout boxes based on severity level (`info`, `tip`, `warning`, `critical`, `roast`).
- **Nano-Style Navigation (Zero Vim Keys):** Pure intuitive arrow keys, Tab, Space, PageUp/PageDown, and Enter. No modal states or `h`/`j`/`k`/`l` bindings.
- **Multiple Transport Interfaces:** Supports standard Unix pipeline (stdin $\rightarrow$ stdout), file arguments (`--input` / `--output`), and Unix Domain Sockets (`--socket`).
- **Flexible Exit & Action Codes:** Custom buttons (e.g. "Done", "Ask Later", "Ignore Today") map to distinct exit codes and output JSON statuses.

---

## 2. Invocation Contracts & Modes

### 2.1 Mode A: Interactive Unix Pipeline (Stdin $\rightarrow$ Stdout)
Perfect for shell scripts, piping, and quick CLI invocations:
```bash
echo '{"title": "Quick Prompt", ...}' | sister-tui
```
User interacts with the TUI in the terminal, and upon submission, `sister-tui` prints the result JSON directly to stdout:
```bash
RESULT=$(cat prompt.json | sister-tui)
```

### 2.2 Mode B: File-Based CLI Runner
Ideal for launching inside terminal emulators (`tilix`, `tmux`, `gnome-terminal`):
```bash
sister-tui --input /tmp/prompt.json --output /tmp/result.json
```
- **Exit Code 0:** User submitted/confirmed.
- **Exit Code 10:** User clicked secondary action 1 (e.g., "Ask Later").
- **Exit Code 20:** User clicked secondary action 2 (e.g., "Ignore Today").
- **Exit Code 130:** Abrupt user cancellation (`Esc`, `Ctrl+C`, or closed terminal).

### 2.3 Mode C: Pure Markdown Viewer Mode (Zero Output)
For reading markdown files or piped text without prompting for answers:
```bash
# View a local markdown file
sister-tui view notes.md

# Or pipe directly into the viewer
curl -s https://example.com/changelog.md | sister-tui --view
```
- Renders full Glamour-styled Markdown inside the centered 90–100 col viewport.
- Scroll with Up/Down arrows and PageUp/PageDown.
- Press `Enter`, `Esc`, or `q` to exit.
- Produces **zero output** to stdout/stderr (clean exit code 0).

### 2.4 Mode D: Unix Domain Socket Server (Daemon Mode)
Allows `sister-tui` to run persistently in a dedicated terminal or window and receive prompt events over a socket:
```bash
sister-tui daemon --socket /tmp/sister-tui.sock
```
Any client connects, streams the JSON payload, waits for the user to complete the UI, and receives the response JSON over the socket connection.

---

## 3. Declarative Input Schema

Every `sister-tui` screen is composed of a header, an array of **Blocks**, and an **Action Bar**.

```json
{
  "title": "Modal Title (Header Banner)",
  "subtitle": "Optional contextual subtitle",
  "blocks": [
    {
      "type": "markdown",
      "severity": "roast",
      "content": "### Attention Required\nYou are currently 35 minutes idle while your projects sit untouched."
    },
    {
      "type": "single_select",
      "id": "status_choice",
      "label": "What are you working on right now?",
      "options": [
        {"id": "coding", "label": "Writing code in annoying-sister"},
        {"id": "research", "label": "Reading documentation"},
        {"id": "slacking", "label": "Just procrastinating on YouTube"}
      ],
      "default": "coding"
    },
    {
      "type": "multi_select",
      "id": "goals_selected",
      "label": "Select goals to tackle today:",
      "options": [
        {"id": "g1", "label": "Finish v2 spec", "description": "Priority 1"},
        {"id": "g2", "label": "Fix DBus reader bug", "description": "Priority 2"},
        {"id": "g3", "label": "Study DSA trees", "description": "Semester 3"}
      ],
      "default": ["g1"]
    },
    {
      "type": "text_input",
      "id": "custom_note",
      "label": "Additional notes for today:",
      "placeholder": "Type here...",
      "required": false
    }
  ],
  "actions": [
    {"id": "submit", "label": "[Enter] Done", "exit_code": 0, "is_primary": true},
    {"id": "defer", "label": "[Tab+1] Ask Later", "exit_code": 10},
    {"id": "skip", "label": "[Tab+2] Ignore Today", "exit_code": 20}
  ]
}
```

---

## 4. Supported Block Types

### 4.1 Markdown Block (`"type": "markdown"`)
Rendered via Lip Gloss + Glamour with colored left-accent border and title pill:
- `"severity": "info"` $\rightarrow$ Cyan accent border, `ℹ INFO` badge.
- `"severity": "tip"` $\rightarrow$ Green accent border, `✔ SUCCESS / TIP` badge.
- `"severity": "warning"` $\rightarrow$ Yellow/Amber accent border, `⚠ WARNING` badge.
- `"severity": "critical"` $\rightarrow$ Bright Red bold accent border, `✖ CRITICAL` badge.
- `"severity": "roast"` $\rightarrow$ Magenta / Deep Purple border, `🔥 SISTER ROAST` badge.

### 4.2 Single Select (`"type": "single_select"`)
Radio buttons navigated via Up/Down arrow keys or `j`/`k`:
```
( ) Option A
(•) Option B  (Selected)
( ) Option C
```

### 4.3 Multi Select (`"type": "multi_select"`)
Checkboxes toggled via `Space`:
```
[ ] 1. Finish the v2 spec
[x] 2. Review DSA notes
[ ] 3. Clean git workspace
```

### 4.4 Text Input (`"type": "text_input"`)
Single-line text input using `bubbles/textinput`.

### 4.5 Text Area (`"type": "textarea"`)
Multi-line text input using `bubbles/textarea`.

### 4.6 Confirm (`"type": "confirm"`)
Binary prompt (`[ Yes ]  [ No ]`) navigated via Left/Right arrows or Tab.

---

## 5. Output Response Schema

When the user activates an action button, `sister-tui` produces:

```json
{
  "action_id": "submit",
  "exit_code": 0,
  "completed_at": "2026-10-13T15:45:00+03:00",
  "answers": {
    "status_choice": "slacking",
    "goals_selected": ["g1", "g3"],
    "custom_note": "Need to fix the scheduler first"
  }
}
```

If dismissed abruptly:
```json
{
  "action_id": "dismissed",
  "exit_code": 130,
  "completed_at": "2026-10-13T15:45:10+03:00",
  "answers": {}
}
```

---

## 6. How `nudged` Translates Its Domain into `sister-tui` Payloads

Since `sister-tui` is completely agnostic, `nudged` creates payloads on the fly:

### 6.1 Morning Session Example
`nudged` constructs:
1. `markdown` block (`severity: "info"`): Formatted Tech, AI, and Kenyan news summary.
2. `multi_select` block: Populated from `goals.json` active/pending items.
3. `text_input` block: "Add new daily tasks".
4. `markdown` block: Today's class schedule overview.
5. Action buttons: `[Done]`, `[Ask Later]`, `[Ignore Today]`.

### 6.2 Missed Class Example
`nudged` constructs:
1. `markdown` block (`severity: "roast"`): *"You're still on your laptop while DSA is ongoing at LT3!"*
2. `single_select` block: Excuse choices (Overslept, Sick, Procrastinating, etc.).
3. `text_input` block: "Other explanation".
4. Action buttons: `[Save Excuse]`, `[Ask Tonight]`.

---

## 7. Package Layout for `sister-tui` Repo

```
sister-tui/
├── cmd/
│   └── sister-tui/
│       └── main.go              (CLI flags, stdin/socket reader, exit handler)
├── internal/
│   ├── model/                  (Root Bubble Tea model, navigation, focus management)
│   ├── blocks/
│   │   ├── markdown.go         (Glamour markdown renderer + severity borders)
│   │   ├── singleselect.go     (Radio button component)
│   │   ├── multiselect.go      (Checkbox tag component)
│   │   ├── textinput.go        (Bubbles textinput wrapper)
│   │   └── textarea.go         (Bubbles textarea wrapper)
│   ├── schema/
│   │   ├── input.go            (JSON unmarshaling and validation)
│   │   └── output.go           (JSON response serialization)
│   └── theme/
│       └── styles.go           (Lip Gloss color palettes, borders, severity styles)
├── examples/
│   ├── morning.json
│   ├── missed_class.json
│   └── review.json
├── go.mod
├── go.sum
└── README.md
```

## 8. Keyboard Controls & Navigation (Nano-Style, Zero Vim)

| Key | Context | Action |
|---|---|---|
| `↑` / `↓` (Up / Down) | Lists / Radios / Checkboxes | Move selection up / down |
| `↑` / `↓` or `PgUp` / `PgDn` | Markdown Viewport | Scroll content up / down smoothly |
| `Space` | Checkboxes / Multi-select | Toggle item check/uncheck |
| `Tab` | Any modal | Move focus forward to next block or action button |
| `Shift + Tab` | Any modal | Move focus backward to previous block |
| `Enter` | Text Input / Buttons | Confirm input or activate highlighted button |
| `Esc` / `Ctrl + C` | Anywhere | Cancel / Dismiss modal (Exit code 130) |
| `q` | Viewer Mode only | Cleanly close markdown viewer (Exit code 0) |

> [!IMPORTANT]
> **Vim modal navigation is strictly banned.** No `h`/`j`/`k`/`l` cursor intercepting, no modal Normal/Insert states. Text typing and arrow keys behave naturally like standard desktop apps and GNU nano.

