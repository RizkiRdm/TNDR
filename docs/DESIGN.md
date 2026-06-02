# DESIGN.md — TENDR
# Terminal UI Design System

Document Version: 1.0.0
Project: TENDR — AI Gateway
Interface Target: Terminal (TUI via Bubble Tea + Lip Gloss)
Aesthetic Direction: Industrial Terminal — raw, dense, honest

---

## 1. Design Philosophy

TENDR runs in a terminal. The UI must feel native to that environment.

Design principles, in strict priority order:

1. **Information density** — terminal real estate is limited, every character earns its place
2. **Clarity of state** — gateway status, provider health, cost must be readable at a glance
3. **Terminal-native aesthetics** — no pretending to be a web UI, embrace the constraints
4. **Motion as signal** — spinners and updates communicate state, not decoration
5. **Keyboard-first** — every action reachable without a mouse

Aesthetic reference: htop, k9s, lazygit, OpenCode.
NOT: web dashboards ported to terminal, excessive box-drawing, ASCII art logos.

---

## 2. Color System

Terminal colors use Lip Gloss color definitions.
Support both true color (16M) terminals and 256-color fallback.

### Base Palette

| Token | True Color | 256-color | Role |
|---|---|---|---|
| `ColorBg` | `#0D0D0D` | `232` | Root background |
| `ColorSurface` | `#141414` | `233` | Panel backgrounds |
| `ColorSurface2` | `#1C1C1C` | `234` | Elevated surfaces |
| `ColorBorder` | `#2A2A2A` | `235` | Structural borders |
| `ColorBorder2` | `#3A3A3A` | `237` | Medium emphasis |
| `ColorBorder3` | `#525252` | `239` | Active/selected |

### Content Palette

| Token | True Color | Role |
|---|---|---|
| `ColorTextPrimary` | `#E8E8E8` | Primary labels, values |
| `ColorTextSecondary` | `#A0A0A0` | Secondary labels, metadata |
| `ColorTextTertiary` | `#606060` | Disabled, timestamps |
| `ColorTextMuted` | `#404040` | Decorative separators |

### Semantic Palette

| Token | True Color | Role |
|---|---|---|
| `ColorSuccess` | `#4CAF7D` | Request success, cache hit |
| `ColorFailed` | `#E05252` | Request failed, provider down |
| `ColorWarning` | `#F5A623` | Fallback triggered, high cost |
| `ColorRetrying` | `#7B68EE` | Fallback in progress |
| `ColorIdle` | `#3A3A3A` | Provider idle, no traffic |
| `ColorActive` | `#F5A623` | Active request processing |

### Accent Palette

| Token | True Color | Role |
|---|---|---|
| `ColorAccent` | `#F5A623` | Selected tabs, active elements |
| `ColorAccentDim` | `#7A5210` | Accent backgrounds |

---

## 3. Typography

Terminal typography = character selection + spacing + case conventions.

### Font Rules

- TENDR MUST NOT specify fonts — terminal inherits user's monospace font
- All text is monospace by nature of terminal
- Visual hierarchy via: weight (bold), case (UPPER/lower), color, and indentation

### Text Hierarchy

| Level | Style | Usage |
|---|---|---|
| Header | BOLD + UPPER + ColorTextPrimary | Tab names, section headers |
| Subheader | BOLD + ColorTextSecondary | Panel titles, column headers |
| Body | Normal + ColorTextPrimary | Values, names, content |
| Secondary | Normal + ColorTextSecondary | Labels, units, metadata |
| Muted | Normal + ColorTextTertiary | Timestamps, IDs, hints |
| Accent | BOLD + ColorAccent | Selected state, current value |
| Success | Normal + ColorSuccess | Positive states |
| Error | BOLD + ColorFailed | Error states |
| Warning | Normal + ColorWarning | Warning states |

### Numeric Display Rules

- Cost values: ALWAYS 6 decimal places (`$0.000042`)
- Token counts: ALWAYS formatted with comma separator (`1,234`)
- Latency: ALWAYS show unit (`142ms`, `1.2s`)
- Percentages: ALWAYS one decimal (`98.4%`)
- Counts: plain integer, no formatting under 10,000

---

## 4. Layout System

### Global Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ TOPBAR — 1 line                                                 │
├─────────────────────────────────────────────────────────────────┤
│ TAB BAR — 1 line                                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ CONTENT AREA — fills remaining height                           │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ STATUSBAR — 1 line                                              │
└─────────────────────────────────────────────────────────────────┘
```

### Topbar

```
 TENDR v0.1.0   ●  running :4821   OpenAI ✓  Anthropic ✓  Groq ✗
```

- Left: `TENDR` + version, dimmed
- Center: gateway status + port
- Right: provider health indicators
- Height: 1 line
- Border-bottom: `─` separator line in ColorBorder

### Tab Bar

```
 [Dashboard]  [Cost]  [Cache]  [Config]  [Logs]
```

- Active tab: ColorAccent + bold + underline
- Inactive tab: ColorTextTertiary
- Separator: `  ` (two spaces between tabs)
- Keyboard hint: `1-5` or arrow keys shown in statusbar

### Statusbar

```
 ↑↓ navigate   tab switch   enter select   q quit   ? help
```

- Always visible
- ColorTextTertiary
- Updates contextually per active tab

### Minimum Terminal Size

- REQUIRED: 80 columns × 24 rows
- Below 80×24: show message `Terminal too small. Resize to continue.`

---

## 5. Tab Specifications

### Tab 1 — Dashboard

```
┌─ GATEWAY STATUS ──────────────────────────────────────────────┐
│  Status    ● RUNNING       Port    4821                       │
│  Uptime    2h 14m          Requests  1,847 total              │
└───────────────────────────────────────────────────────────────┘

┌─ PROVIDER HEALTH ─────────────────────────────────────────────┐
│  OpenAI      ● ok    142ms avg    2,341 req    $0.0421        │
│  Anthropic   ● ok    891ms avg      412 req    $0.0211        │
│  Gemini      ● ok     88ms avg      103 req    $0.0012        │
│  Groq        ✗ down    —             —           —            │
└───────────────────────────────────────────────────────────────┘

┌─ LAST 10 REQUESTS ────────────────────────────────────────────┐
│  TIME        MODEL         PROVIDER    TOKENS   COST    STATUS│
│  11:42:01    coding        openai       1,204   $0.0021  ✓    │
│  11:41:58    fast          groq           341   $0.0002  ✗ fb │
│  11:41:44    default       anthropic      892   $0.0091  ✓    │
│  ...                                                          │
└───────────────────────────────────────────────────────────────┘
```

`✗ fb` = failed, fallback used
Live updates every 1 second via Bubble Tea tick command.

---

### Tab 2 — Cost

```
┌─ COST SUMMARY ────────────────────────────────────────────────┐
│  Today         $0.0421     This Week    $0.3142               │
│  This Month    $1.2441     All Time     $4.8821               │
└───────────────────────────────────────────────────────────────┘

┌─ BY PROVIDER ─────────────────────────────────────────────────┐
│  OpenAI        $0.0282   ████████████░░░░  67%               │
│  Anthropic     $0.0112   █████░░░░░░░░░░░  27%               │
│  Gemini        $0.0027   █░░░░░░░░░░░░░░░   6%               │
└───────────────────────────────────────────────────────────────┘

┌─ BY MODEL ALIAS ──────────────────────────────────────────────┐
│  coding        $0.0211   requests: 142   avg: $0.000149      │
│  default       $0.0142   requests:  89   avg: $0.000160      │
│  fast          $0.0068   requests: 203   avg: $0.000033      │
└───────────────────────────────────────────────────────────────┘

┌─ PRICING SOURCE ──────────────────────────────────────────────┐
│  Active source: fetched (updated 2h ago)                      │
│  gpt-4o        input $2.50/1M    output $10.00/1M            │
│  claude-3-5    input $3.00/1M    output $15.00/1M            │
│  press [r] to refresh pricing                                 │
└───────────────────────────────────────────────────────────────┘
```

Inline bar chart uses block characters: `█` for filled, `░` for empty.

---

### Tab 3 — Cache

```
┌─ CACHE STATUS ────────────────────────────────────────────────┐
│  Type       In-Memory + Disk                                  │
│  Entries    1,247 exact   12 semantic                         │
│  Hit Rate   34.2% today   41.8% all time                      │
│  Saved      $0.0142 today                                     │
└───────────────────────────────────────────────────────────────┘

┌─ CACHE ENTRIES ───────────────────────────────────────────────┐
│  MODEL ALIAS   TYPE      HITS   CREATED      EXPIRES          │
│  coding        exact       14   11:30:00     12:30:00         │
│  coding        exact        2   11:28:44     12:28:44         │
│  default       exact        1   10:14:22     never            │
│  ...                                                          │
└───────────────────────────────────────────────────────────────┘

 [c] clear all   [d] clear selected   [enter] view entry
```

---

### Tab 4 — Config

```
┌─ ACTIVE CONFIG ───────────────────────────────────────────────┐
│  File: ~/.tendr/config.yaml                [e] open in $EDITOR│
│                                                               │
│  tendr:                                                       │
│    port: 4821                                                 │
│    log_level: info                                            │
│                                                               │
│  providers:                                                   │
│    openai:                                                     │
│      api_key: sk-...••••••••                                  │
│    anthropic:                                                  │
│      api_key: sk-ant-...••••••••                              │
│                                                               │
│  models:                                                      │
│    - alias: coding                                            │
│      fallback_mode: smart                                     │
│      ...                                                      │
└───────────────────────────────────────────────────────────────┘

 [e] edit in $EDITOR   [r] reload config   [v] validate
```

API keys MUST be masked: show prefix 6 chars + `••••••••`.
Config is read-only in TUI — editing opens `$EDITOR`.

---

### Tab 5 — Logs

```
┌─ REQUEST LOG ─────────────────────────────────────────────────┐
│  [live] auto-scroll ON   filter: all   [f] filter  [p] pause  │
├───────────────────────────────────────────────────────────────┤
│  11:42:01 INFO  gateway   request_received  req_abc123        │
│  11:42:01 INFO  router    provider_selected openai/gpt-4o     │
│  11:42:01 INFO  gateway   request_success   142ms $0.0021     │
│  11:41:58 WARN  router    fallback_triggered groq→openai      │
│  11:41:58 ERROR provider  provider_error    groq timeout      │
│  11:41:58 INFO  gateway   request_success   891ms $0.0091 fb  │
└───────────────────────────────────────────────────────────────┘
```

Level color coding:
- `INFO` → ColorTextTertiary
- `WARN` → ColorWarning
- `ERROR` → ColorFailed

Auto-scroll MUST pause when user scrolls up.
Auto-scroll MUST resume when user presses `G` (go to bottom).

---

## 6. Component Patterns

### Status Indicator

```
● RUNNING     (ColorSuccess + bold)
● STOPPED     (ColorFailed + bold)
◌ STARTING    (ColorWarning, animated spinner)
```

Spinner frames for Bubble Tea: `◐ ◓ ◑ ◒` cycling at 100ms.

### Provider Health Badge

```
● ok      (ColorSuccess)
✗ down    (ColorFailed)
⚡ slow   (ColorWarning)  — latency > 2x average
```

### Inline Bar Chart (Cost/Usage)

```
OpenAI   ████████████░░░░  67%
```

- Full block: `█` (U+2588)
- Empty block: `░` (U+2591)
- Bar width: 16 chars fixed
- Percentage: right-aligned, 4 chars

### Table Layout

```
COLUMN_A    COLUMN_B    COLUMN_C
────────    ────────    ────────
value_1     value_2     value_3
value_1     value_2     value_3
```

- Column headers: UPPER + ColorTextSecondary
- Separator line: `─` in ColorBorder
- Values: ColorTextPrimary
- Alternating row highlight: PROHIBITED (visual noise)
- Fixed column widths per table — MUST NOT reflow on data change

### Border Boxes

Use Lip Gloss `Border(lipgloss.RoundedBorder())` — EXCEPTION to flat border rule because terminal box-drawing requires it for readability.

Title format: `─ SECTION TITLE ─────`

---

## 7. Keyboard Navigation

### Global Keys

| Key | Action |
|---|---|
| `1` | Switch to Dashboard tab |
| `2` | Switch to Cost tab |
| `3` | Switch to Cache tab |
| `4` | Switch to Config tab |
| `5` | Switch to Logs tab |
| `tab` | Next tab |
| `shift+tab` | Previous tab |
| `q` | Quit |
| `?` | Toggle help overlay |

### Dashboard Tab

| Key | Action |
|---|---|
| `r` | Refresh provider health |

### Cost Tab

| Key | Action |
|---|---|
| `r` | Refresh pricing from GitHub |
| `e` | Export cost log to CSV |
| `d` | Toggle day/week/month/all view |

### Cache Tab

| Key | Action |
|---|---|
| `c` | Clear all cache |
| `d` | Delete selected entry |
| `enter` | View entry detail |

### Config Tab

| Key | Action |
|---|---|
| `e` | Open config in `$EDITOR` |
| `r` | Reload config from disk |
| `v` | Validate current config |

### Logs Tab

| Key | Action |
|---|---|
| `f` | Open filter input |
| `p` | Pause/resume auto-scroll |
| `G` | Jump to latest (resume auto-scroll) |
| `c` | Clear log view (not log file) |

---

## 8. Motion Rules

Terminal animation is minimal. Use only for state communication.

| Trigger | Animation | Implementation |
|---|---|---|
| Gateway starting | Spinner on status indicator | Bubble Tea spinner component |
| Live request incoming | New row flash in log | Brief ColorAccent on new row, 300ms |
| Provider health check | Spinner per provider | Bubble Tea ticker |
| Auto-scroll new entry | Scroll to bottom | Bubble Tea viewport |
| Tab switch | Instant — no transition | N/A |
| Cost value update | None — just update | N/A |

### Animation Constraints

- MUST NOT animate more than 2 elements simultaneously
- MUST NOT use animation for purely decorative purposes
- MUST respect `NO_COLOR` environment variable — disable all color if set
- MUST respect `TERM=dumb` — disable all styling if set

---

## 9. CLI Output Style

When running CLI commands (non-TUI), output follows these rules:

### Success Output

```
$ tendr status
  status   running
  port     4821
  uptime   2h 14m
  requests 1,847
```

Key-value format. Left-aligned keys, 2-space indent, value aligned.

### Error Output

```
$ tendr start
  error  config file not found: ~/.tendr/config.yaml
  hint   run 'tendr init' to create a default config
```

Always include `hint` line on recoverable errors.

### Verbose Flag

`--json` flag on any command outputs raw JSON to stdout.
Useful for scripting.

```bash
tendr cost --json | jq '.today'
```

---

## 10. Forbidden Patterns

| Pattern | Reason |
|---|---|
| Animated ASCII art logo on startup | Noise, slows startup feel |
| Progress bars for individual requests | Too much visual noise in logs |
| Confirmation prompts for non-destructive actions | Slows workflow |
| Color without semantic meaning | Every color must mean something |
| Mouse-dependent interactions | Terminal-native = keyboard-first |
| Emoji in output | Breaks non-unicode terminals |
| Gradient text effects | Not supported consistently across terminals |
| Blocking TUI renders | All I/O must be async via Bubble Tea commands |
| Raw provider error strings shown to user | Must be normalized to typed errors |
| Showing full API keys anywhere | Always mask, always |
