# ANR FG Installer

A small desktop installer for [ANR FG](https://github.com/apex-compute/ANR) (Apex Neural Rendering frame generation), built with [MyGo](https://mygo.run) native UI: one Go binary, no WebView.

- Finds the games in every Steam library (`libraryfolders.vdf`, `appmanifest_*.acf`) and marks those that ship DLSS frame generation (`nvngx_dlssg.dll`, `sl.dlss_g.dll`). Other game folders can be added by hand.
- Finds the folder of the game's `.exe` (Unreal's `*-Win64-Shipping.exe`, an earlier OptiScaler, the DLSS libraries, else the largest `.exe`); it can be changed.
- Reads what is installed: ANR FG's full package (`OptiScaler/nr_fg_install.json`), its add-on (`custom/Apex Neural Rendering FG/nr_fg_install.json`), an OptiScaler of your own, and nr-record.
- Picks the **full** or **add-on** package by itself, or lets you choose; it refuses the combinations the packages' installers refuse.
- Gets the package from the latest `FG-V.*` GitHub release, checked against its `SHA256SUMS.txt`, or from a local zip (`nr_fg_package.json` tells its kind and version).
- Runs the package's own `install.ps1` / `uninstall.ps1` (`install.sh` / `uninstall.sh` under Proton), so backups and restores work exactly as with `install.cmd`.
- Before a full-package upgrade, backs up `OptiScaler.ini`, which the package replaces, to `OptiScaler.ini.bak-<time>`; warns when nr-record must be installed again.

Downloads and unpacked packages are cached in `%LOCALAPPDATA%\anr-installer`. Set `GITHUB_TOKEN` to raise GitHub's API rate limit.

## Build

Go 1.27 or later:

```
go tool mygo dev                      # run with live reload
go test ./...                         # tests (ANR_PKG_DIR=<folder with the zips> also runs the real packages)
go tool mygo build                    # build/windows-amd64/ANR FG Installer.exe and its NSIS installer
go tool mygo build -skip-installer    # the executable only
```
