# eis precheck hook (Claude Code PreToolUse)

Inject a module's **structural-debt context** into your coding agent right before
it writes a file — so the model reads the debt *before* it edits. It's a single
`eis` subcommand (no bash/jq/extra scripts), reads a local index, and stays
**silent on healthy code** and **fail-open on any error** (it never blocks a
write).

## How it works

```
eis write-index            →  .eis/write-index.json   (per-module structural facts)
Claude Code PreToolUse hook →  eis precheck-hook       (resolves file → module → directive)
```

`precheck-hook` reads the PreToolUse JSON on stdin, resolves the target
`file_path` to a module using the **same** module patterns the index was built
with (single source of truth), looks that module up in `.eis/write-index.json`,
and — only if the module is notable — prints an `additionalContext` directive.
Clean modules, unknown paths, and a missing index all produce **no output**.

The injected line carries **only structural facts** (tier, days idle/untouched,
recommendation). It never contains an owner's name.

## Setup

### 1. Generate the index (and keep it fresh)

```bash
eis write-index            # writes .eis/write-index.json (lean pipeline; cheap on a warm cache)
```

Refresh it when the picture changes — e.g. on a schedule, in CI, on a
post-merge/`post-checkout` git hook, or before a big agent session. A warm cache
makes the refresh near-instant. Commit `.eis/write-index.json` or `.gitignore`
it, your call.

### 2. Register the hook

Merge [`settings.json`](./settings.json) into your project's
`.claude/settings.json` (or `~/.claude/settings.json` for all projects):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [
          { "type": "command", "command": "eis precheck-hook" }
        ]
      }
    ]
  }
}
```

`eis` must be on `PATH`. To point at a non-default index location, use
`eis precheck-hook --index /abs/path/to/write-index.json`.

## What the agent sees

When a write targets an orphaned module, the hook injects (as
`additionalContext`, without blocking the tool):

```
[eis] svc/auth — Module has no active owner (left 243d ago, untouched 243d). Prefer a minimal diff; add a test to establish shared ownership; don't expand the public surface; flag to a human.
```

Directives by recommendation:

| recommendation   | directive |
|------------------|-----------|
| `orphaned_module`| minimal diff · add a test for shared ownership · don't expand the public surface · flag to a human |
| `dead_module`    | prefer deletion over extension · escalate any revive decision to a human |
| `bus_factor_1`   | don't concentrate more here · write so others can build on it |
| clean            | *(no output)* |

## Contract (PreToolUse)

- **Input** (stdin): `{ "hook_event_name": "PreToolUse", "tool_name": "Write", "cwd": "…", "tool_input": { "file_path": "…" } }`
- **Output** (stdout, exit 0): `{ "hookSpecificOutput": { "hookEventName": "PreToolUse", "additionalContext": "…" } }` — `permissionDecision` is omitted, so the write proceeds normally; the hook only adds context.
- **Silent** on clean/unknown/missing-index; **fail-open** on any error.
