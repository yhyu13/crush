## Summary

This PR introduces three built-in agent skills — **Critic**, **Replacer**, and **Toolcoach** — that work together as a middleware stack around the primary `SessionAgent`. The goal is to improve output quality, prevent regressions, and coach the agent toward better tool usage patterns without adding perceivable latency.

## What Changed

### 1. Critic Skill (`internal/skills/critic/`)

A secondary-LLM review layer that inspects the agent's diff after each turn and returns one of:
- **`approve`** — changes look good, turn completes normally
- **`revise`** — rollback changes, inject feedback, re-drive the primary agent
- **`halt`** — rollback changes, return error to user

**Key features:**
- SQLite-backed review persistence with sqlc-generated queries
- SHA-256 keyed LRU cache for deduplicated reviews
- Circuit breaker + timeout for resilient LLM calls
- LSP diagnostic fetching for richer context
- Configurable via `crush.json` (`options.critic`)
- CLI commands: `crush critic list`, `crush critic show`, `crush critic stats`
- Per-session disable via `ReplacerEnabled` flag

### 2. Replacer Skill (`internal/skills/replacer/`)

A conversation coach that evaluates whether the primary agent's response is complete or needs a follow-up. Uses a small/fast model to decide `stop` vs `continue`.

**Key features:**
- Configurable max iterations (default 3)
- Duplicate-prompt guard to avoid repetitive follow-ups
- Timeout handling (deadline exceeded = treat as `stop`)
- Coach spinner indicators ("Coach is ready" → "Coach is evaluating")
- **`/skipcoach`** command to interrupt current evaluation one-time
- Commands dialog integration ("Skip Coach")
- Auto-interrupt on user typing or Enter press

### 3. Toolcoach Skill (`internal/skills/toolcoach/`)

A zero-LLM, heuristic-based skill that detects anti-patterns in real-time tool usage and injects coaching tips into tool results.

**Key features:**
- **~2.7µs overhead per tool call** (benchmarked)
- Detects patterns: `destructive_bash`, `write_over_existing`, `edit_without_view`, `broad_grep`, `repeated_view`, `missing_multiedit`
- SQLite effectiveness tracking with adaptive severity
- Progressive coaching (hint → warning → critical) based on repetition
- Guided retry mechanism for destructive operations
- Configurable via `crush.json` (`options.toolcoach`)

### 4. Middleware Stack Order

```
outermost: ToolcoachMiddleware
            ↓ wraps
           ReplacerMiddleware
            ↓ wraps
           CriticMiddleware
            ↓ wraps
inner-most: SessionAgent (primary)
```

`SkipCoach` propagates through the chain via interface type-assertion at each layer.

### 5. UI & Integration

- New commands dialog items: "Skip Coach"
- `/skipcoach` (or `skipcoach`) intercepted in textarea
- `AgentSkipCoach(sessionID)` wired through workspace → coordinator → middleware chain
- Coach evaluation no longer blocks user commands (typing, new session, model selection, etc.)
- Visual busy indicators preserved (placeholder, progress bar, cancel key, todo spinner)

### 6. Database & Schema

New migrations:
- `critic_reviews` table (verdict, confidence, concerns, diff snapshot)
- `written_files` table (tracks files written per session for diff baseline)
- `toolcoach_effectiveness` table (pattern hit/miss tracking)

## Testing

```bash
# Critic tests
go test ./internal/skills/critic/... -v -race

# Replacer tests
go test ./internal/skills/replacer/... -v -race

# Toolcoach tests
go test ./internal/skills/toolcoach/... -v -race

# Standalone critic + replacer demo (no API keys)
go run ./cmd/critic-demo
```

## Configuration Example

```json
{
  "options": {
    "critic": {
      "enabled": true,
      "model": "anthropic/claude-sonnet-4",
      "max_iterations": 3,
      "auto_approve": false,
      "threshold": 0.85
    },
    "replacer": {
      "enabled": true,
      "max_iterations": 3
    },
    "toolcoach": {
      "enabled": true,
      "max_patterns": 100
    }
  }
}
```

## Checklist

- [ ] I have read [`CONTRIBUTING.md`](https://github.com/charmbracelet/.github/blob/main/CONTRIBUTING.md).
- [ ] I have created a discussion that was approved by a maintainer (for new features).
