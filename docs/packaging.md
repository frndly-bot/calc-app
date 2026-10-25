# Packaging and release

The app ships as a single native executable with the frontend embedded. Packaging is split into platform Taskfiles under `build/` and orchestrated from the root `Taskfile.yml`.

| Platform | Build output | Installer / package |
| --- | --- | --- |
| Linux | `bin/calc-app` | AppImage, `.deb` (optional `.rpm`, Arch package) |
| Windows | `bin/calc-app.exe` | NSIS installer with WebView2 bootstrapper |
| macOS | `.app` bundle | `.app`, compressed `.dmg` (native or universal) |

## Metadata

Version and product metadata live in several places. Keep them in sync when releasing:

| File | Fields |
| --- | --- |
| [build/config.yml](../build/config.yml) | `productName`, `productIdentifier` (`com.zcode.calc`), `version`, company, copyright |
| [build/darwin/Info.plist](../build/darwin/Info.plist) | `CFBundleVersion`, `CFBundleShortVersionString`, bundle id, minimum macOS |
| [build/darwin/Info.dev.plist](../build/darwin/Info.dev.plist) | Development bundle id/name |
| [build/windows/info.json](../build/windows/info.json) | File/product version, company, description |
| [build/windows/nsis/project.nsi](../build/windows/nsis/project.nsi) | Installer version, publisher, display name |
| [build/linux/nfpm/nfpm.yaml](../build/linux/nfpm/nfpm.yaml) | Package version, maintainer, dependencies |
| [frontend/package.json](../frontend/package.json) | Frontend package version |
| `build/darwin/Taskfile.yml` | `APP_VERSION` used in the DMG filename |

The application identifier is `com.zcode.calc`; the Linux binary and package are named `calc-app`, the product name is `Calc`.

Icons are generated from `build/appicon.png`:

```bash
task generate:icons
```

This writes `build/windows/icon.ico` and `build/darwin/icons.icns`. The Linux desktop entry uses `build/appicon.png` directly.

## Common prerequisites

- Go, Node, pnpm, Task and the Wails v3 CLI (see [development.md](development.md)).
- Platform toolchains for the target OS. Cross-compiling installers is not supported by the Taskfiles; build each platform on its own machine or CI runner.
- A clean working tree before producing release artifacts.

`task package` always performs a production build first:

- Go flags: `-tags production -trimpath -buildvcs=false -ldflags="-w -s"`.
- Windows additionally sets `-H windowsgui` so no console window appears.
- The frontend is built by `pnpm run build` into `frontend/dist` and embedded.

## Linux

### Build

```bash
task build
task run
```

Output: `bin/calc-app`.

### Package

```bash
task linux:package       # AppImage + .deb
task linux:create:rpm    # optional RPM
task linux:create:aur    # optional Arch Linux package
```

Prerequisites for packaging:

- GTK4/WebKitGTK 6.0 development packages (for the cgo build).
- `wails3 generate appimage` downloads/manages its AppImage tooling.
- `nfpm` support is built into `wails3 tool package`; no separate install is required.
- RPM packaging may require `rpmbuild`; Arch packaging produces a PKGBUILD-style artifact.

Outputs are written to `bin/`. The AppImage task copies the binary and icon into a staging directory, then generates the AppImage. The Debian package uses [nfpm.yaml](../build/linux/nfpm/nfpm.yaml):

```
/usr/local/bin/calc-app
/usr/share/icons/hicolor/128x128/apps/appicon.png
/usr/share/applications/calc-app.desktop
```

with dependencies `libgtk-4-1` and `libwebkitgtk-6.0-4` (RPM/Arch equivalents are overridden in the same file).

### Install and smoke test

```bash
chmod +x bin/*.AppImage
./bin/*.AppImage

sudo apt install ./bin/calc-app_*.deb
calc-app
```

Run the installed binary once, switch modes, toggle the theme and confirm history persists in `~/.config/calc-app/`.

## Windows

Build on Windows (64-bit):

```powershell
task build
task package
```

Prerequisites:

- Go 1.27.2+, Node 24+, pnpm, Task, Wails CLI.
- WebView2 runtime (preinstalled on current Windows 10/11; the installer bootstrapper covers older machines).
- NSIS (`makensis` on `PATH`) for the installer.

`task windows:build`:

1. Regenerates `wails_windows_<arch>.syso` from the icon, manifest and `info.json`.
2. Builds `bin/calc-app.exe`.
3. Removes the intermediate `.syso`.

`task windows:package` runs the production build, generates the WebView2 bootstrapper, then calls `makensis` with [project.nsi](../build/windows/nsis/project.nsi). Output: `bin/calc-app-installer.exe`.

The installer:

- requires 64-bit Windows and admin rights;
- installs to `%ProgramFiles%\Calc`;
- creates Start Menu and desktop shortcuts;
- registers an uninstall entry under `com.zcode.calc`;
- runs the app from the finish page.

Architecture selection uses the Taskfile's `ARCH` variable (`amd64` by default, `arm64` supported). The NSIS script accepts the matching binary define.

## macOS

Build on macOS:

```bash
task build              # native architecture
task darwin:build:universal
task package            # production .app + .dmg
task darwin:package:universal
```

Prerequisites: Xcode Command Line Tools, `hdiutil`, `lipo` and `codesign` (all system tools), plus the Wails CLI.

`task darwin:build` sets:

```
GOOS=darwin
GOARCH=<arch>
CGO_ENABLED=1
MACOSX_DEPLOYMENT_TARGET=12.0
```

Packaging creates `bin/Calc.app` with:

```
Contents/
├── Info.plist
├── MacOS/calc-app
└── Resources/icons.icns
```

and signs it ad hoc (`codesign --force --deep --sign -`). The DMG task then produces `bin/Calc-0.1.0.dmg` with `hdiutil` (UDZO compression, volume name `Calc`). Update `APP_VERSION` in `build/darwin/Taskfile.yml` for the filename.

`task darwin:run` builds a development `.app` using `Info.dev.plist`, signs it ad hoc and launches it — useful for testing bundle behaviour outside the Wails dev loop.

### Distribution signing

The committed configuration is suitable for local builds and internal distribution. For public macOS distribution:

1. Replace the ad-hoc signing identity with a Developer ID Application certificate.
2. Enable hardened runtime and entitlements as required.
3. Sign the `.app`, then sign the `.dmg`.
4. Submit for notarization with `notarytool` and staple the ticket.

These steps require Apple credentials and are intentionally not automated in the repository.

## CI

[.github/workflows/ci.yml](../.github/workflows/ci.yml) validates the project on Ubuntu 24.04:

- installs GTK4/WebKitGTK development packages;
- runs Go tests;
- runs `svelte-check`, Vitest, Playwright and the production frontend build;
- builds the Linux desktop binary with `CGO_ENABLED=1`.

CI validates the project but does not publish packages. Pushing a `v*` tag triggers [.github/workflows/release.yml](../.github/workflows/release.yml), which builds the Linux binary, AppImage, `.deb`, `.rpm` and Arch package, writes `SHA256SUMS`, and creates the GitHub release with generated notes. The package version comes from the tag, and the hardcoded versions in `build/config.yml` and `build/linux/nfpm/nfpm.yaml` are rewritten in the runner only.

Windows and macOS artifacts require their own runners (for example a GitHub Actions matrix with `windows-latest` and `macos-latest`), because the platform Taskfiles call native tools (`makensis`, `hdiutil`, `lipo`, `codesign`).

## Release checklist

1. **Version bump** — update every file in the metadata table above (`0.1.0` in this repository).
2. **Dependencies** — `go mod tidy`, `pnpm --dir frontend install`, commit lockfiles.
3. **Assets** — if the icon changed, run `task generate:icons` and commit the regenerated `.ico`/`.icns`.
4. **Checks** — `task lint`, `task test`, then `pnpm --dir frontend build`.
5. **Browser smoke test** — `task serve`, exercise all four modes, history, theme and persistence (see [development.md](development.md)).
6. **Packages** — run `task package` on each target OS; verify the artifacts launch from a clean install.
7. **Signing** — sign Windows/macOS artifacts where distribution requires it; document checksums.
8. **Evidence** — refresh `docs/screenshots/` if the UI changed, and update the README feature list.
9. **Tag and publish** — push an annotated tag matching the version (`git tag -a v1.2.3 -m "v1.2.3" && git push origin v1.2.3`). The release workflow publishes the Linux artifacts; attach Windows/macOS builds manually.

## What is not included

- Automatic updates.
- Code signing certificates or notarization credentials.
- A cross-platform release pipeline that builds every installer in one workflow. The tag workflow publishes Linux packages only.
- Package repository publishing (APT/RPM/Homebrew/Chocolatey).

These require external accounts, secrets and infrastructure and should be added deliberately when the project has a distribution target.
