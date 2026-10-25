# Architecture

Calc is a Wails v3 desktop application: a Go host process owns the window and the calculator engines, and a SvelteKit single-page app renders the interface inside the native webview. The frontend is compiled to static assets and embedded into the Go binary, so the shipped application is one executable with no local server.

```
┌────────────────────────────── Wails host (Go) ──────────────────────────────┐
│ main.go                                                                     │
│   embed frontend/dist → application.New → register services → open window   │
│                                                                             │
│   CalcService ──► calc/scientific ──► calc/parser                           │
│                ├─► calc/developer                                           │
│                └─► calc/finance                                             │
│   HistoryService  ──► config (history.json)                                 │
│   SettingsService ──► config (settings.json)                                │
│                                                                             │
│   Events: history:changed · settings:changed                                │
└──────────────────────────────────▲──────────────────────────────────────────┘
                                   │ Wails runtime
                     Call.ByName(...) · Events.On(...)
┌──────────────────────────────────┴──────────────────────────────────────────┐
│ SvelteKit SPA (adapter-static, ssr = false)                                 │
│                                                                             │
│ routes/+page.svelte                                                         │
│   ├─ components/  Display, Keypad, Key, TabBar, HistoryPanel, Toast, …       │
│   ├─ modes/       Normal · Scientific · Developer · Finance                  │
│   ├─ stores/      settings · mode · history · memory · ui · toast (runes)    │
│   ├─ eval/        TypeScript parser/evaluator for Normal mode                │
│   └─ bindings/    typed wrappers over the Wails service calls                │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Process and window

[main.go](../main.go) is the only Go entry point. It:

1. Loads `HistoryService` and `SettingsService` from disk before the window exists, logging (not aborting) on failure.
2. Creates the Wails application with `Name: "Calc"`, embedded assets, and a `localhost:34115` server configuration used by the headless `server` build tag.
3. Registers the three services. Their exported methods become the binding surface.
4. Wires `OnChanged` callbacks to Wails events, so every persistence mutation reaches the UI.
5. Opens a 420×700 window (minimum 360×560) at `/` and calls `app.Run()`.

The window is a normal desktop window (`BackgroundTypeSolid`), with the dark glass treatment drawn entirely by CSS.

## Evaluation paths

There are two evaluation engines by design.

| Mode | Engine | Why |
| --- | --- | --- |
| Normal | TypeScript ([frontend/src/lib/eval/normal.ts](../frontend/src/lib/eval/normal.ts)) | Instant keystroke-level preview with no IPC. Simple arithmetic and percentage semantics are cheap to implement and test in JS. |
| Scientific | Go ([backend/calc/scientific](../backend/calc/scientific/scientific.go)) | Shared Pratt parser, `math/big` factorials, angle-unit handling, and one source of truth for expression errors. |
| Developer | Go ([backend/calc/developer](../backend/calc/developer/developer.go)) | Fixed-width word arithmetic, two's complement, shifts/rotates and base formatting need explicit unsigned semantics. |
| Finance | Go ([backend/calc/finance](../backend/calc/finance/finance.go)) | Amortization schedules and rounded monetary results belong with the domain logic and its tests. |

### Scientific parsing

The parser package ([backend/calc/parser](../backend/calc/parser/parser.go)) tokenizes numbers, identifiers, operators, parentheses and the postfix `!`, then builds an AST with a Pratt parser. It supports:

- `+ - * / % ^` with `^` right-associative and unary minus binding below it (`-2^2 = -4`).
- Implicit multiplication: `2(3+4)`, `2pi`, `3sin(0.5)`.
- Function calls with or without parentheses (`sin(30)` or `sin 30`), with arity validation.
- Constants and function names resolved through an environment supplied by the scientific engine.

The scientific engine registers trig, inverse trig, hyperbolic, logarithmic, power, root, rounding and utility functions, plus `pi`, `tau`, `e` and `phi`. Results that are `NaN` or infinite become user-facing errors. Factorials use `math/big` for values up to 1000!.

### Developer word model

Developer values are `uint64` words masked to 8, 16, 32 or 64 bits. A single value is carried as:

- `raw` — the unsigned bit pattern as a **decimal string**;
- `value` — the signed or unsigned numeric interpretation for display;
- `formatted` — `Hex`, `Dec`, `Oct`, `Bin` strings.

The decimal string matters: JavaScript numbers cannot represent every 64-bit integer. Keeping `raw` as a string across the JSON boundary preserves all 64 bits. The frontend only uses `BigInt` locally for flipping individual bits, then sends the resulting decimal string back to the service.

### Finance

`LoanPayment` uses the standard annuity formula, with a zero-rate branch, and builds a rounded amortization schedule whose final payment settles the remaining balance exactly. `SavingsFutureValue` treats deposits as end-of-month. `ROI`, `CAGR` and `PercentChange` validate their inputs and round to two decimals.

## Services and bindings

The service façade lives in [backend/services](../backend/services).

| Service | Methods | Purpose |
| --- | --- | --- |
| `CalcService` | `ScientificEvaluate`, `DevParse`, `DevBinary`, `DevUnary`, `DevFormat`, `CalcLoan`, `CalcSavings`, `CalcROI`, `CalcCAGR`, `CalcPercentChange` | Stateless calculation entry points. Inputs are validated by the engines; failures return Go errors. |
| `HistoryService` | `List`, `Add`, `Delete`, `Clear` | In-memory list backed by `history.json`; newest-first reads; capped at 200 entries; fires `OnChanged` after mutations. |
| `SettingsService` | `Get`, `Save` | Loads with defaults, merges unknown/missing fields, persists the whole record, fires `OnChanged`. |

The frontend calls services through [frontend/src/lib/bindings/index.ts](../frontend/src/lib/bindings/index.ts). It wraps `Call.ByName` with the fully qualified names:

```
calc-app/backend/services.CalcService.ScientificEvaluate
calc-app/backend/services.HistoryService.List
calc-app/backend/services.SettingsService.Save
```

TypeScript interfaces mirror the Go structs (`DevResult`, `LoanSummary`, `Payment`, `SavingsResult`, `HistoryEntry`, `Settings`). Go errors reject the returned promise; `errorMessage()` strips the runtime prefix so the UI can display the domain message.

`task generate:bindings` can emit Wails-generated TypeScript into `frontend/src/lib/bindings/generated` for inspection, but the committed application uses the hand-written wrapper because it provides one stable place for the FQN package, shared types and error normalization.

## Frontend structure

### Shell

[frontend/src/routes/+page.svelte](../frontend/src/routes/+page.svelte) owns the persistent frame: brand, theme toggle, history button, tab bar, animated mode stage, shortcut footer, history drawer and toast layer. The stage is keyed by the current mode, so switching modes runs a `fly`/`fade` transition while the accent variable transitions independently.

On mount it:

1. Loads persisted settings.
2. Selects the saved mode without re-persisting it.
3. Loads history.
4. Subscribes to `history:changed` when the Wails runtime is available.

### Stores

All shared state uses Svelte 5 runes in `.svelte.ts` modules.

| Store | File | State and behaviour |
| --- | --- | --- |
| Settings | [settings.svelte.ts](../frontend/src/lib/stores/settings.svelte.ts) | Holds `current` and `loaded`; merges backend values over defaults; applies the theme to `document.documentElement.dataset.theme`; `update(patch)` is optimistic and suppresses save failures so the browser preview still works. |
| Mode | [mode.svelte.ts](../frontend/src/lib/stores/mode.svelte.ts) | Current mode, mode metadata (label, accent, hint), index lookup, and persistence through settings. Accents are blue, violet, emerald and amber. |
| History | [history.svelte.ts](../frontend/src/lib/stores/history.svelte.ts) | Loads newest-first entries, adds optimistically, removes with rollback on failure, clears and reloads on failure. |
| Memory | [memory.svelte.ts](../frontend/src/lib/stores/memory.svelte.ts) | Shared MC/MR/M+/M− value, in memory only; non-finite updates are ignored. |
| UI | [ui.svelte.ts](../frontend/src/lib/stores/ui.svelte.ts) | History drawer visibility. |
| Toast | `toast.svelte.ts` | Transient error notifications used by stores and modes. |

Mode components keep their own expression/input state locally; only persisted preferences and shared history go through stores. This keeps each calculator surface independent and makes mode switches cheap.

### Components

- `Display` shows the expression line, dynamic-size result, optional badge (2nd function, developer base), optional parenthesis-depth chip, errors and a count-up animation after equals.
- `Keypad` is a CSS grid wrapper; `Key` implements press-scale/glow feedback, disabled states, variants (number, operator, function, memory, danger, equals) and accessible labels.
- `TabBar` renders the four mode tabs and a sliding accent indicator.
- `HistoryPanel` slides over the app, animates entries with `flip`, and supports per-entry delete and clear-all.
- `MemoryBar` exposes memory keys in Normal and Scientific modes.
- `Toast` renders the shared error queue.

## Persistence

[backend/config/config.go](../backend/config/config.go) resolves the data directory:

1. `XDG_CONFIG_HOME/calc-app` when the environment variable is set (used by tests and portable runs).
2. Otherwise `os.UserConfigDir()/calc-app`:
   - Linux: `~/.config/calc-app/`
   - macOS: `~/Library/Application Support/calc-app/`
   - Windows: `%AppData%\calc-app\`

`SaveJSON` writes indented JSON to `<file>.tmp` with mode `0600` and renames it over the target, so a crash cannot leave a half-written state file. `LoadJSON` treats a missing file as an empty default and reports decode errors to the caller.

`settings.json`:

```json
{
  "theme": "dark",
  "mode": "normal",
  "angleUnit": "deg",
  "devBits": 32,
  "devSigned": false,
  "currency": "USD"
}
```

`history.json` is an array of `{id, mode, expr, result, createdAt}`. IDs increase monotonically within the loaded file; the service trims the oldest entries beyond 200.

Settings merging overlays only non-zero loaded fields onto defaults, so adding a preference in a later version does not break an existing file. `DevSigned` is assigned directly because `false` is a meaningful value.

## Event flow

```
HistoryService.Add/Delete/Clear
        └─ OnChanged ─► app.Event.Emit("history:changed")
                              └─ shell listener ─► history.load() ─► UI refresh

SettingsService.Save
        └─ OnChanged ─► app.Event.Emit("settings:changed", Settings)
                              └─ available for future listeners
```

Mode components also update the history store optimistically after `Add` returns, so the drawer is current even before the event listener round-trip finishes.

## Error handling

- Engines return descriptive Go errors (`division by zero`, `word size must be …`, `principal must be positive`); services pass them through.
- The frontend normalizes rejected runtime errors with `errorMessage()` and displays them in the display area or a toast depending on context.
- Persistence failures are logged by Go and surfaced to the frontend where a user action can recover (history add/delete/clear); settings saves fail soft because the in-memory preference is still usable.
- Browser-preview mode catches missing bindings and events so the UI renders with defaults.

## Headless server build

`task build:server` compiles with `CGO_ENABLED=0 -tags server`, producing `bin/calc-server`. Wails serves the embedded frontend and the same services over HTTP at `http://localhost:34115`. The frontend detects the Wails runtime when present; in a plain browser, calls fail and stores fall back to defaults. This mode is used for CI, screenshots and smoke tests without a GTK/WebKit desktop session.

## Test seams

| Layer | Test style | Location |
| --- | --- | --- |
| Parser and engines | Table-driven Go tests with edge cases and error paths | `backend/calc/*/…_test.go` |
| Services and persistence | Go tests with `XDG_CONFIG_HOME` pointed at a temp directory | `backend/services/services_test.go`, `backend/config/config_test.go` |
| Normal evaluator and formatting | Vitest behaviour tests, including percentage semantics | `frontend/src/lib/eval/normal.test.ts` |
| UI flows | Headless server + browser automation (manual smoke test) | `task serve` |

The pure engine packages never import Wails, so they can be tested without a display server or the native toolchain.
