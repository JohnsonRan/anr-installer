# Developing ANR FG Installer

The app is a single Go program with a native UI drawn by [MyGo](https://mygo.run): the window is the main
package at the repository root, and the logic is under `internal/` (`steam`, `detect`, `pkg`, `store`).

## Requirements

- Go 1.27 or later
- Python 3.9 or later, for the requirements check
- Git Bash on Windows, or a shell elsewhere, for `scripts/check.sh`

## Commands

```
go tool mygo dev                      # run with live reload
scripts/check.sh --fast               # before a commit: formatting, vet, the type and spacing scales
scripts/check.sh                      # before a merge: also the tests and the requirements check
scripts/check.sh --release            # before a release: also the real packages and the build
go tool mygo build                    # build/windows-amd64/ANR FG Installer.exe and its NSIS installer
go tool mygo build -skip-installer    # the executable only
go run ./tools/genicon                # render resources/logo.svg into the app icon, resources/icon.png
```

The release tier installs, upgrades, uninstalls and downgrades real packages in scratch folders: set
`ANR_PKG_DIR` to a folder holding the newest `anr-fg-*-full.zip` and `ANR_OLD_PKG` to an older one.

## Rules

The rules for changing this repository are in `AGENTS.md`, `UI.md` (the interface) and `internal/AGENTS.md`;
reviews follow `REVIEW.md`; commands, numbers and gates are in `REQUIREMENTS.md`. Text sizes come from the type
scale in `type.go`, and paddings and gaps from the spacing scale in `spacing.go`; the pre-commit tier fails on
a size or a spacing set anywhere else.
