# How ANR FG Installer works

Reference for what the installer reads, decides and changes. For using it, see the [README](../README.md).

## Finding games

- Steam's folder comes from the registry on Windows (`HKCU\Software\Valve\Steam`, `SteamPath`), and from
  `~/.steam/steam`, `~/.local/share/Steam` and the Flatpak folder elsewhere.
- Every Steam library is listed in `steamapps/libraryfolders.vdf`, and every installed game in an
  `appmanifest_<appid>.acf` there. Steam's own tools (Proton, the Linux runtimes, the redistributables) are left
  out.
- A game outside Steam is added by its folder. It is kept in `%APPDATA%\anr-installer\games.json` with any `.exe`
  folder chosen by hand, and is named after the product name in its `.exe`'s version resource until renamed.
  Removing it takes it off the list only; its files stay.

## Reading a game folder

- **DLSS frame generation**: the folder ships `nvngx_dlssg.dll` or `sl.dlss_g.dll` (Streamline), up to twelve
  folders down, as Unreal games keep them in `Engine/Plugins/Runtime/Nvidia/StreamlineCore/...`.
- **The `.exe` folder**, where the installers go: next to an earlier `OptiScaler.ini`; else Unreal's
  `*-Win64-Shipping.exe`; else next to the DLSS libraries; else the largest `.exe`. It can be changed by hand.
- **What is installed**: ANR FG's full package leaves `OptiScaler/nr_fg_install.json`; its add-on leaves
  `custom/Apex Neural Rendering FG/nr_fg_install.json` (`custom/nrfg` before V.B04) in an OptiScaler the user
  installed. `OptiScaler.ini` alone is an OptiScaler of the user's own. `FGNvngxReplacement=nr-record` means
  nr-record is in use.
- **Anti-cheat**: Easy Anti-Cheat, EA Javelin Anticheat, BattlEye, nProtect GameGuard, XIGNCODE3, Anti-Cheat
  Expert, mhyprot and PunkBuster are recognized by the folders and files they leave in the game's folder.
- **Pictures**: Steam's cached icon, library hero, logo and store header, from `appcache/librarycache/<appid>/`
  in the window's language, else the `.exe`'s icon.

## Versions

The provider DLL ANR FG installs carries no version. Each package's `nr_fg_package.json` gives its version and
the SHA-256 of that DLL (`provider_sha256`), so the installed DLL's hash names its version. The installer reads
`nr_fg_package.json` out of every release's packages with HTTP range requests, one or two 64 KB blocks each,
instead of downloading the packages, and keeps the result in `%LOCALAPPDATA%\anr-installer\versions.json`;
published packages do not change, so only new releases are read again.

## Installing

- The package is the newest `FG-V.*` release, or one chosen from V.B04 on (V.B01 to V.B03 installed into
  `custom/nrfg`), downloaded from GitHub and checked against the release's `SHA256SUMS.txt`; or a local zip.
- **full** brings its own OptiScaler; **addon** joins an OptiScaler the user installed. Auto picks by what the
  folder holds, and the combinations the packages' installers refuse are refused before they run.
- The package's own `install.ps1` and `uninstall.ps1` (`install.sh` and `uninstall.sh` under Proton) do the
  work, so backups and restores are exactly those of `install.cmd`.
- A newer version is installed over the installed one. An older one, or any but the newest when the installed
  version is unknown, is installed after uninstalling with the newest release's package: an installer copies
  files and removes none, so a newer version's files would stay beside an older one.
- `OptiScaler.ini`, which the full package replaces and its uninstaller restores, and `nr_fg.ini`, which an
  uninstall removes, are copied next to the `.exe` as `<name>.bak-<time>` first.
- Installing into a game that ships anti-cheat asks first; cancelling is the default.

## Files the installer keeps

| Path | Holds |
| --- | --- |
| `%APPDATA%\anr-installer\games.json` | games added by hand, and `.exe` folders chosen by hand |
| `%LOCALAPPDATA%\anr-installer\versions.json` | each release package's version and provider hash |
| `%LOCALAPPDATA%\anr-installer\downloads\` | downloaded packages |
| `%LOCALAPPDATA%\anr-installer\packages\` | unpacked packages |

The portable version, which a file named `portable` next to its executable marks, keeps `games.json` in its own
folder and the rest under `cache\` there.

`GITHUB_TOKEN`, when set, raises GitHub's API rate limit; it is sent to GitHub's API only.

## Updating the installer itself

The installer updates itself with MyGo's updater. It checks
`https://github.com/JohnsonRan/anr-installer/releases/latest/download/update-windows-amd64.json` 10 seconds after
launch and then once a day; the About dialog, from the version button in the title bar, checks at once and turns
the automatic checks on or off. A newer version is offered in a window with its release notes, to install, skip
or be reminded of later. The update archive, or a smaller delta from the running version, is installed only when
its signature matches the public key in `mygo.json`. The app can update itself where it can write its own folder:
the per-user install in `%LOCALAPPDATA%\Programs\ANR FG Installer`, or the portable version's folder.
