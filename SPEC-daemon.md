# Spec: SpectreTUI — daemon

Module: `daemon` · Depends on: `prompt-engine` · Parent: `CAPABILITY-MAP.md`

## Objective

Run the prompt engine persistently in a dedicated terminal: `SpectreTUI daemon --socket /tmp/SpectreTUI.sock` listens on a Unix domain socket. A client (`nudged` or a script) connects, writes one JSON payload, waits for the user to finish the card in the daemon's terminal, and reads one JSON response on the same connection.

Acceptance criteria:
- `cat examples/review.json | socat - UNIX-CONNECT:/tmp/SpectreTUI.sock` prints the response JSON.
- One prompt renders at a time; concurrent connections are handled FIFO.
- Ctrl+C shuts the daemon down cleanly and removes the socket file.
- The socket is created with 0600 perms — a local same-user trust boundary; no auth layer.

## Tech Stack

Go stdlib `net` (`UnixListener`) + the `prompt-engine` runner. No new dependencies.

## Commands

```
Run daemon:  go run ./cmd/SpectreTUI daemon --socket /tmp/SpectreTUI.sock
Test client: cat examples/review.json | socat - UNIX-CONNECT:/tmp/SpectreTUI.sock
Test:        go test ./...
```

## Project Structure

```
cmd/SpectreTUI/main.go         → `daemon` subcommand wiring
internal/daemon/server.go      → accept loop, FIFO queue, per-connection handler
internal/daemon/server_test.go → integration test with a stubbed prompt runner
```

## Code Style

As `SPEC-prompt-engine.md` § Code Style. The server takes its prompt runner as a `runPrompt func(schema.Payload) (schema.Response, error)` parameter so tests inject a stub instead of needing a TTY.

## Testing Strategy

Stdlib integration test: start the server on a socket in `t.TempDir()`, `net.Dial`, send an example payload, assert one JSON response with an `action_id` arrives and the connection closes. A two-connection test asserts FIFO ordering. The real TUI path is verified manually (`socat` client against a live daemon).

## Boundaries

- **Always:** replace a stale socket file on startup; remove the socket file on shutdown; render prompts strictly one at a time; 0600 socket perms.
- **Ask first:** changing the wire protocol (length-prefixing, keep-alive multi-prompt connections), adding auth.
- **Never:** render prompts concurrently, leave the socket file behind after exit, add TLS/auth layers (the local socket is the trust boundary).

## Success Criteria

- Payload in → card in the daemon terminal → §5 response JSON (including the `dismissed` shape on Esc/Ctrl+C) back on the same connection, then close.

## Open Questions

— none
