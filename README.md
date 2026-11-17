# ANR FG Installer

A small desktop installer for [ANR FG](https://github.com/apex-compute/ANR) (Apex Neural Rendering frame generation), built with [MyGo](https://mygo.run) native UI: one Go binary, no WebView.

- Finds the games in every Steam library (`libraryfolders.vdf`, `appmanifest_*.acf`) and marks those that ship DLSS frame generation (`nvngx_dlssg.dll`, `sl.dlss_g.dll`). Other game folders can be added by hand.
- Finds the folder of the game's `.exe` (Unreal's `*-Win64-Shipping.exe`, an earlier OptiScaler, the DLSS libraries, else the largest `.exe`); it can be changed.
- Keeps games outside Steam, added by their folder, across launches in `%APPDATA%\anr-installer\games.json`, with the .exe folders chosen by hand. Such a game is named after its `.exe`'s product name until renamed, and can be taken off the list (its files stay).
- Reads what is installed: ANR FG's full package (`OptiScaler/nr_fg_install.json`), its add-on (`custom/Apex Neural Rendering FG/nr_fg_install.json`, or `custom/nrfg` before V.B04), an OptiScaler of your own, and nr-record.
- Picks the **full** or **add-on** package by itself, or lets you choose; it refuses the combinations the packages' installers refuse.
- Tells the installed version without downloading a package: it reads `nr_fg_package.json` (whose `provider_sha256` is the hash of that version's `nvngx_dlssg.dll`) out of each release's packages with HTTP range requests, about 64 KB each, once, and compares the installed DLL with it.
- Installs any `FG-V.*` GitHub release from V.B04 on (V.B01 to V.B03 used another folder), checked against its `SHA256SUMS.txt`, or a local zip (`nr_fg_package.json` tells its kind and version). A newer version goes over the installed one; an older one, or any but the newest when the installed version is unknown, is installed after uninstalling with the newest release's package, so that no newer files stay behind. `OptiScaler.ini` and `nr_fg.ini` are backed up next to the `.exe` first.
- Runs the package's own `install.ps1` / `uninstall.ps1` (`install.sh` / `uninstall.sh` under Proton), so backups and restores work exactly as with `install.cmd`.
- The app icon is `resources/logo.svg`, rendered to the `resources/icon.png` the build packages by `go run ./tools/genicon`.
- Draws its own title bar and window buttons in place of the system's.
- Shows each game's Steam icon and header art (Steam's `appcache/librarycache`), else the icon of its `.exe`.
- Before a full-package upgrade, backs up `OptiScaler.ini`, which the package replaces, to `OptiScaler.ini.bak-<time>`; warns when nr-record must be installed again.

Downloads, unpacked packages and the versions read from the releases (`versions.json`) are cached in `%LOCALAPPDATA%\anr-installer`. Set `GITHUB_TOKEN` to raise GitHub's API rate limit.

## Build

Go 1.27 or later:

```
go tool mygo dev                      # run with live reload
go test ./...                         # tests (ANR_PKG_DIR=<folder with the zips> also runs the real packages)
go tool mygo build                    # build/windows-amd64/ANR FG Installer.exe and its NSIS installer
go tool mygo build -skip-installer    # the executable only
```
