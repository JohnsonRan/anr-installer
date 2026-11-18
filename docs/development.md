# Developing ANR FG Installer

The app is a single Go program with a native UI drawn by [MyGo](https://mygo.run). `main.go` only starts it.
The window is under `gui/`, by feature:

- `gui`: the app's state and pages, the window and the notification area icon;
- `gui/i18n`: the texts in English, Simplified Chinese, Traditional Chinese and Japanese;
- `gui/style`: the colors, icons, type and spacing scales, and the shared components;
- `gui/activity`: the activity log and its panel;
- `gui/compare`: the comparison page.

The logic is under `internal/` (`steam`, `detect`, `pkg`, `record`, `store`, `paths`), and the files built into
the executable are in `assets/`.

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
go run ./tools/genicon                # render assets/logo.svg into the app icon, assets/icon.png
go test -tags screenshots -run TestScreenshots ./gui   # redraw the README screenshots in docs/images
```

The release tier installs, upgrades, uninstalls and downgrades real packages in scratch folders: set
`ANR_PKG_DIR` to a folder holding the newest release's `anr-fg-*-full.zip` and `anr-fg-*-addon.zip`, and
`ANR_OLD_PKG` to an older `anr-fg-*-full.zip`. With an `nr-record-*-windows.zip` in `ANR_PKG_DIR` as well,
it also installs nr-record into a scratch folder and removes it; without one, that test is skipped.

The README screenshots are drawn from made-up games in `gui/screenshots_test.go`, in English, Simplified Chinese
and Japanese. Draw them again when a view they show changes.

## Rules

The rules for changing this repository are in `AGENTS.md`, `gui/AGENTS.md` (the interface) and
`internal/AGENTS.md`; reviews follow `REVIEW.md`; commands, numbers and gates are in `REQUIREMENTS.md`. Text
sizes come from the type scale in `gui/style/type.go`, and paddings and gaps from the spacing scale in
`gui/style/spacing.go`; the pre-commit tier fails on a size or a spacing set anywhere else.

## Releasing

A tag releases. `.github/workflows/release.yml` checks the tag against `mygo.json` and `CHANGELOG.md`, runs the
release tier against the two newest ANR FG releases, builds and signs the app and its updates with
`scripts/release.sh`, and publishes the release, whose notes are the version's section of `CHANGELOG.md`.

1. Raise `version` in `mygo.json` and add its section, `## <version>`, to `CHANGELOG.md`: it becomes the release
   notes and the text of the update window. Commit and push.
2. Tag the commit and push the tag:

   ```
   git tag -s v<version> -m v<version>
   git push origin v<version>
   ```

The workflow signs updates with the repository secret `MYGO_UPDATER_PRIVATE_KEY`, the contents of
`mygo-update.key`, set once with `gh secret set MYGO_UPDATER_PRIVATE_KEY < path/to/mygo-update.key`. If a run
fails, fix the cause and run the failed job again: it keeps the draft and replaces its files.

To release from your machine instead, run the release tier, `scripts/check.sh --release`, with `ANR_PKG_DIR` (a
folder with the newest release's full and addon zips) and `ANR_OLD_PKG` (an older full zip); then
`MYGO_UPDATER_PRIVATE_KEY="$(cat path/to/mygo-update.key)" scripts/release.sh`, which uploads the installer,
`install.ps1`, the signed update files and the portable zip to a draft; then publish the draft with
`gh release edit v<version> --draft=false`.

The secret key signs every update; installed apps accept only updates signed with it, so keep it out of the
repository. Losing it strands the installed apps; anyone who has it can ship code to their users, so whoever can
change the workflow or the secret can too.
