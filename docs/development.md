# Development

## Prerequisites

| Tool | Version | Notes |
| --- | --- | --- |
| Go | 1.27.2+ | Matches `go.mod`. The pure Go tests run without a native toolchain. |
| Node.js | 24+ | Frontend build and tests. |
| pnpm | 10+ | The repository commits `frontend/pnpm-lock.yaml`; CI pins pnpm 10. |
| Task | 3.x | Task runner used by every script. Install from <https://taskfile.dev/installation/>. |
| Wails v3 CLI | `v3.0.0-beta.34` | Required for `task dev`, native builds and packaging. |

Install the Wails CLI at the version used by the module:

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.34
export PATH="$PATH:$(go env GOPATH)/bin"
wails3 doctor
```

### Platform packages

**Linux** — Wails v3 beta uses GTK4 and WebKitGTK 6.0:

```bash
sudo apt-get update
sudo apt-get install -y libgtk-4-dev libwebkitgtk-6.0-dev
```

Build essentials (`gcc`, `pkg-config`) are also required for a native build. The headless server build (`CGO_ENABLED=0`) needs none of these.

**macOS** — Xcode Command Line Tools and macOS 12+ deployment target. The Taskfile sets `MACOSX_DEPLOYMENT_TARGET=12.0`.

**Windows** — Windows 10/11, the WebView2 runtime, and NSIS (`makensis`) only if you build the installer. Native Windows builds use `CGO_ENABLED=0`.

## First run

```bash
git clone <repository-url> calc-app
cd calc-app
pnpm --dir frontend install

task check
task test
task dev
```

`task dev` starts the Wails hot-reload loop:

1. Installs frontend dependencies once.
2. Starts Vite on `127.0.0.1:9245`.
3. Runs `go mod tidy` and a development build.
4. Launches the app; Go changes rebuild, frontend changes hot-reload.

Set `WAILS_VITE_PORT` to move the Vite port if 9245 is occupied.

## Headless browser preview

For UI work on a machine without a desktop session, CI, or screenshots:

```bash
task serve
# http://localhost:34115
```

`task serve` builds the frontend, compiles `bin/calc-server` with `CGO_ENABLED=0 -tags server`, and runs it. The server embeds the same assets and serves the same services over HTTP. Without the Wails runtime, persistence calls fail softly: settings use defaults, history stays empty, and the interface still renders and evaluates Normal mode locally.

Stop the process with `Ctrl+C`. To isolate state while testing, point the config directory somewhere disposable:

```bash
XDG_CONFIG_HOME=$(mktemp -d) ./bin/calc-server
```

## Everyday commands

| Command | What it does |
| --- | --- |
| `task` | List tasks |
| `task dev` | Wails hot-reload loop |
| `task build` | Native development build into `bin/` |
| `task run` | Run the native build |
| `task package` | Platform installer(s) |
| `task build:server` | Frontend build + headless server binary |
| `task serve` | Build and serve on port 34115 |
| `task test` | Go tests + Vitest |
| `task check` | `go vet` + `svelte-check` |
| `task lint` | `gofmt` check + `task check` |
| `task generate:bindings` | Optional Wails TS bindings into `frontend/src/lib/bindings/generated` |
| `task generate:icons` | Regenerate `.ico` / `.icns` from `build/appicon.png` |
| `task clean` | Remove `bin/`, `frontend/dist`, `.svelte-kit` and Windows `.syso` files |

The native build embeds `frontend/dist` at compile time. If you change the frontend and run an existing binary, rebuild — hot reload only exists in `task dev`.

## Testing

### Full suite

```bash
task test
```

This runs `CGO_ENABLED=0 go test -tags server ./...` and `pnpm test` (Vitest). Both `task test` and `task check` build `frontend/dist` first, because `main.go` embeds it; the same ordering is required in CI.

### Single Go package or test

```bash
CGO_ENABLED=0 go test -tags server ./backend/calc/scientific -v
CGO_ENABLED=0 go test -tags server ./backend/calc/finance -run TestLoanPayment -v
```

The `server` tag avoids linking the desktop webview. Service and config tests set `XDG_CONFIG_HOME` to a temporary directory, so they never touch your real preferences.

### Single frontend test file

```bash
pnpm --dir frontend test src/lib/eval/normal.test.ts
pnpm --dir frontend test:watch
```

Vitest runs in a Node environment (`frontend/vitest.config.ts`), so only pure TypeScript is unit-tested there. Component behaviour is verified through the headless server in a browser.

### Type checking and formatting

```bash
task check          # go vet + svelte-check
task lint           # gofmt + check
gofmt -w backend main.go
```

### UI smoke test

```bash
task serve
```

Then exercise each mode in a browser:

1. Normal: `200 + 10 % =` → `220`; keyboard `7 × 8 Enter` → `56`.
2. Scientific: switch to `DEG`, `sin ( 30 ) =` → `0.5`; check the `( n` chip while typing.
3. Developer: 32-bit unsigned, enter `F0`, `AND`, `0F`, `=` → `0`; toggle a bit.
4. Finance: loan defaults → `$1,580.17` per month; savings, ROI, CAGR and Change tabs.
5. History: confirm, open the drawer, delete an entry, clear all.
6. Toggle light theme, reload, and confirm theme and last mode persist.

## Bindings workflow

The application imports from [frontend/src/lib/bindings/index.ts](../frontend/src/lib/bindings/index.ts), a hand-written wrapper. When you add or rename a service method:

1. Update the Go method signature.
2. Update the matching TypeScript interface and wrapper in `index.ts`.
3. Run `task check` to catch type drift.
4. Optionally run `task generate:bindings` and compare the generated output in `frontend/src/lib/bindings/generated` (ignored by git) with the wrapper.

The wrapper deliberately centralizes the package FQN (`calc-app/backend/services`) and normalizes Go errors. Generated files are not committed.

## Conventions

- **Go** — keep engine packages (`backend/calc/...`) free of Wails imports; they must remain testable with the standard toolchain. Services are thin adapters. Return errors, do not panic on user input.
- **Svelte** — Svelte 5 runes only (`$state`, `$derived`, `$effect`, `$props`). Shared reactive state lives in `.svelte.ts` stores under `src/lib/stores`. Use the `#lib` import alias. Use CSS custom properties (`--accent`, `--surface`, `--muted`, …) rather than hard-coded theme colours.
- **SvelteKit** — configuration lives in `vite.config.ts`; there is no `svelte.config.js`. The app is a static SPA (`ssr = false`, `prerender = false`, fallback `index.html`, output `frontend/dist`).
- **Formatting** — tabs in Svelte/TypeScript/CSS per `.editorconfig`; `gofmt` for Go.
- **Assets** — UI screenshots live in `docs/screenshots/`; source icon is `build/appicon.png`, with generated platform icons committed.

## Troubleshooting

| Symptom | Fix |
| --- | --- |
| `package gtk4 was not found` / `webkitgtk-6.0` missing | Install the Linux packages above, or use `task build:server` / `task test` which avoid cgo. |
| `wails3: command not found` | Add `$(go env GOPATH)/bin` to `PATH`. |
| `wails3 doctor` complains about the CLI version | Install the pinned `v3.0.0-beta.34`. |
| App launches but shows stale UI | Rebuild with `task build`; `frontend/dist` is embedded at compile time. |
| Bindings reject with `RuntimeError:` | Use `errorMessage()` from the bindings module to strip the prefix. |
| History/settings look wrong during development | Point `XDG_CONFIG_HOME` at a temp directory, or delete `calc-app/history.json` and `settings.json` from the config dir. |
| Port 9245 in use | `WAILS_VITE_PORT=9250 task dev`. |
| Port 34115 in use | Stop the other server; the headless server uses a fixed default port. |
| Headless browser cannot reach the backend | Confirm you launched `bin/calc-server` (not `vite preview`) and that requests target `http://localhost:34115`. |
| WebView fails in a container | Use `task serve` for UI checks or run with a virtual display; the native window needs GTK/WebKit. |

## CI

[.github/workflows/ci.yml](../.github/workflows/ci.yml) runs on Ubuntu 24.04:

1. Install GTK4/WebKitGTK build dependencies.
2. Go tests.
3. `svelte-check`.
4. Vitest.
5. Frontend production build.
6. Linux desktop binary with `CGO_ENABLED=1`.

Cross-platform packaging is documented separately in [packaging.md](packaging.md); Windows and macOS artifacts are built on their target operating systems.
