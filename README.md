# Calc

[![CI](https://github.com/frndly-bot/calc-app/actions/workflows/ci.yml/badge.svg)](https://github.com/frndly-bot/calc-app/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A native desktop calculator with four modes — **Normal**, **Scientific**, **Developer** and **Finance** — built with Go, Wails v3 and SvelteKit (Svelte 5). Dark glass by default, with a light theme, per-mode accent colours and a full keyboard workflow.

<p align="center">
  <img src="docs/screenshots/normal.png" alt="Normal mode" width="320">
  <br>
  <sub>Normal mode in the default dark glass theme</sub>
</p>

## Features

- **Normal** — instant arithmetic in TypeScript with `%`, `±`, parentheses, memory keys and an expression preview.
- **Scientific** — trig/log/power/factorial engine in Go, DEG/RAD/GRAD units and a 2nd-function layer.
- **Developer** — 8/16/32/64-bit words, signed/unsigned reinterpretation, simultaneous HEX/DEC/OCT/BIN readouts, bitwise operations and a clickable bit grid.
- **Finance** — loan payment + amortization schedule, savings future value, ROI, CAGR and percentage change with currency formatting.
- **History** — every confirmed result is persisted locally, searchable at a glance, and deletable individually or all at once.
- **Persistence** — theme, last mode, angle unit, word size, sign mode and currency survive restarts.
- **Keyboard-first** — digits, operators, Enter, Backspace and Escape work in every mode.
- **Reduced motion** — animations are disabled automatically under `prefers-reduced-motion`.

| Scientific | Developer |
| :---: | :---: |
| <img src="docs/screenshots/scientific.png" alt="Scientific mode" width="280"> | <img src="docs/screenshots/developer.png" alt="Developer mode" width="280"> |

| Finance | History |
| :---: | :---: |
| <img src="docs/screenshots/finance.png" alt="Finance mode" width="280"> | <img src="docs/screenshots/history.png" alt="History panel" width="280"> |

<p align="center">
  <img src="docs/screenshots/normal-light.png" alt="Light theme" width="320">
  <br>
  <sub>Normal mode in the light theme</sub>
</p>

## Stack

| Layer | Technology |
| --- | --- |
| Desktop shell | Wails v3 (`v3.0.0-beta.34`) |
| Backend | Go 1.27, pure calculation packages + Wails services |
| Frontend | SvelteKit 3, Svelte 5 runes, TypeScript, Vite 8 |
| Styling | Hand-written CSS, glass surfaces, per-mode CSS variables |
| Tests | `go test`, Vitest |
| Packaging | Task, Wails v3 platform taskfiles, NSIS, AppImage/deb, .app/dmg |

## Quick start

Prerequisites: Go 1.27+, Node 24+, [pnpm](https://pnpm.io/), [Task](https://taskfile.dev/), the [Wails v3 CLI](https://v3.wails.io/), and the platform webview development packages (see [docs/development.md](docs/development.md)).

```bash
git clone https://github.com/frndly-bot/calc-app.git
cd calc-app
pnpm --dir frontend install

task dev          # Wails hot-reload loop
# or
task build        # native binary in bin/
task run
```

### Browser preview

The same frontend and backend can run headlessly over HTTP — useful for CI, screenshots and UI checks without a desktop session:

```bash
task serve        # builds frontend + server binary, serves http://localhost:34115
```

### Tests and checks

```bash
task test         # Go tests + Vitest
task check        # go vet + svelte-check
task lint         # gofmt check + all of the above
```

## Task reference

| Task | Description |
| --- | --- |
| `task` | List all tasks |
| `task dev` | Wails hot-reload development loop |
| `task build` | Native build for the current OS |
| `task run` | Run the native build |
| `task package` | AppImage + deb (Linux), NSIS (Windows), .app + dmg (macOS) |
| `task serve` | Headless HTTP preview on port 34115 |
| `task test` | Go and frontend unit tests |
| `task check` | `go vet` and `svelte-check` |
| `task generate:icons` | Regenerate `.ico`/`.icns` from `build/appicon.png` |
| `task generate:bindings` | Optional Wails TS binding generation (see docs) |
| `task clean` | Remove build output |

## Project layout

```
.
├── main.go                 # Wails bootstrap, embedded frontend, service wiring
├── backend/
│   ├── calc/               # pure Go engines: parser, scientific, developer, finance
│   ├── services/           # CalcService, HistoryService, SettingsService
│   └── config/             # config-dir resolution and atomic JSON I/O
├── frontend/
│   └── src/
│       ├── lib/modes/      # Normal, Scientific, Developer, Finance views
│       ├── lib/components/ # Display, keys, keypad, history, toasts
│       ├── lib/stores/     # settings, mode, memory, history, toast (runes)
│       ├── lib/eval/       # TypeScript evaluator used by Normal mode
│       └── lib/bindings/   # typed wrappers over Wails runtime calls
├── build/                  # icons, platform taskfiles, packaging configs
├── docs/                   # architecture, modes, development, packaging
└── .github/workflows/ci.yml
```

## Data storage

Application state lives in the OS config directory:

- Linux: `$XDG_CONFIG_HOME/calc-app/` (defaults to `~/.config/calc-app/`)
- macOS: `~/Library/Application Support/calc-app/`
- Windows: `%AppData%\calc-app\`

`settings.json` holds preferences; `history.json` keeps the newest 200 calculations. Writes are atomic (temp file + rename).

## Documentation

- [Architecture](docs/architecture.md) — how the Go engines, Wails services and Svelte stores fit together.
- [Modes](docs/modes.md) — behaviour, grammar and keyboard reference for each mode.
- [Development](docs/development.md) — setup, dev loop, testing and troubleshooting.
- [Packaging](docs/packaging.md) — per-OS builds, installers and release checklist.

## License

[MIT](LICENSE)
