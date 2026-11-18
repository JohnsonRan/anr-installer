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
  in the window's language, else the `.exe`'s icon. Pictures already shown keep their language when the
  language setting changes, until the next scan.

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
- After the full package replaced `OptiScaler.ini`, the user's values are merged into the new file three ways, by
  section and name, against the `game/OptiScaler.ini` of the package the user's file came with (read from the
  downloaded package, else with range requests). A value only the user changed is kept; a value only the new file
  changed takes the new one; a value both changed is asked about, or kept in an **Update all**, which logs it as a
  warning. When the installed version is unknown, every value that differs is asked about. `[FrameGen]`
  `FGInput` and `FGNvngxReplacement` always take the new file's values, which load ANR FG. Values are written
  into the new file's own lines, so its comments and layout stay; a merge that fails leaves the new file and the
  backup.
- Installing into a game that ships anti-cheat asks first; cancelling is the default.

## Recording and comparing

On Windows, a game's page drives nr-record, whose package ships in the same releases
(`nr-record-<version>-windows.zip`, checked against `nr-record-<version>-SHA256SUMS.txt`).

- **Installing**: the package's own `install.ps1` and `uninstall.ps1` do the work, as with ANR FG. They need the
  full package's OptiScaler, keep the original `OptiScaler.ini` in `OptiScaler\nr_record_backup` and restore it.
  nr-record is in the game when `OptiScaler\custom\nr-record` exists, and in use when `FGNvngxReplacement` is
  `nr-record`.
- **Recordings**: the folders with a `record.json` in `%LOCALAPPDATA%\nr-record`. A recording belongs to the
  game whose `.exe` has the name in its `game` field. Those with `feature` `fg`, or none, hold frame generation
  inputs.
- **Length**: a recording's length is the time between the first and the last frame in its `frames.idx`
  (64 bytes a frame, the timestamp at byte 32, in ticks of `record.json`'s `qpc_frequency`).
- **Comparing**: `nr-replay` replays the recording, or with `--frames 0:N` its first seconds, through the
  installed ANR FG (`--generator provider`) and
  through the game's own `nvngx_dlssg.dll` (`--generator nvidia`), each streamed into `nr-video`, which writes
  an `.mp4`; `nr-video project` writes the viewer's project of the two. The window plays the two videos with
  Windows' Media Foundation, on one clock but not locked frame by frame; they are HEVC, which needs the
  system's HEVC decoder. The package's viewer, `renderer.exe`, opens the same comparison with
  `--compare --project`, and browses the recordings with `--captures`.
- **Stills**: for the frame a video is paused at, `replay.json` names the recorded real frame behind it.
  `nr-replay --raw --frames` replays that frame and the one before it through both DLLs, and `nr-video png`
  writes each output frame as a PNG; the raw frames are removed afterwards.
- **FFmpeg and FFprobe**: looked for on the `PATH`, then in Scoop's shim folders (`%SCOOP%\shims`,
  `%USERPROFILE%\scoop\shims`, and the same for a global install). `scoop install ffmpeg` installs both. The
  viewer gets their paths in `NR_FFMPEG` and `NR_FFPROBE`.

## Files the installer keeps

| Path | Holds |
| --- | --- |
| `%APPDATA%\anr-installer\games.json` | games added by hand, `.exe` folders chosen by hand, what the close button does (`close_to_tray`), whether a new ANR FG release notifies (`release_notify_off`), the newest release whose notes opened by themselves (`seen_release`), and the newest release a check found (`notified_release`) |
| `%APPDATA%\anr-installer\locale` | the language chosen in Settings, kept by MyGo; absent while the app follows the system |
| `%APPDATA%\anr-installer\window-state.json` | the window's size, place and maximized state, kept by MyGo |
| `%APPDATA%\anr-installer\updater.json` | whether the installer checks for its own updates, and when it last did, kept by MyGo's updater |
| `%LOCALAPPDATA%\anr-installer\versions.json` | each release package's version and provider hash |
| `%LOCALAPPDATA%\anr-installer\downloads\` | downloaded packages |
| `%LOCALAPPDATA%\anr-installer\packages\` | unpacked packages |
| `%LOCALAPPDATA%\anr-installer\compare\<recording>\` | a comparison's two videos, the viewer's project, and its stills under `stills` |

The portable version, which a file named `portable` next to its executable marks, keeps `games.json` and MyGo's files in
its own folder and the rest under `cache\` there. Settings shows the folder of `games.json` as
**Data folder**.

**Delete downloads** in Settings removes `downloads\` and `packages\` only. An install, an uninstall and nr-record
download and unpack a missing package again, which needs GitHub. `versions.json` stays, as rebuilding it reads
every package again, and so do the comparisons.

## Running in the background

- **New releases**: the release list is read at launch and every 6 hours after. A check every 6 hours that fails
  keeps the list already read and adds a warning to the activity log, but shows no error on a game's page and no
  spinner. When a check finds a newest release with a higher build number than `notified_release`, a Windows
  notification tells of it while the window is hidden or minimized, unless `release_notify_off` is set; the
  release is recorded either way, once the games are scanned. A release withdrawn upstream leaves an older one the
  newest, which neither notifies nor is recorded. In the window, its notes open instead (`seen_release`).
  Notifications are notification-area balloons, which Windows 10 and later show as toasts.
- **Start with Windows**: MyGo writes the value `io.github.johnsonran.anrinstaller` under
  `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, the executable's path followed by
  `--mygo-opened-at-login`. Task Manager's Startup page can turn it off. Started so, the window stays hidden behind
  the tray icon when the close button hides to it, and starts minimized otherwise. The setting shows on only while
  the value names the running executable, so a portable folder moved elsewhere shows it off until it is turned
  on again. Uninstalling the installed version deletes the value, and its entry under
  `HKCU\Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run`, when the value starts the copy
  being uninstalled (`assets/installer.nsh`); a value that starts a portable copy stays.
- **Quitting while a task runs**: the close button, the tray's Quit and an update's restart ask first. A quit
  stops what the installer does itself; a program it started, such as a package's `install.ps1`, is not stopped
  by it. An update's restart answered with **Quit anyway** quits without starting the app again: the update is
  installed already and applies at the next launch. The end of a Windows session quits without asking.

`GITHUB_TOKEN`, when set, raises GitHub's API rate limit; it is sent to GitHub's API only.

## Updating the installer itself

The installer updates itself with MyGo's updater. It checks
`https://github.com/JohnsonRan/anr-installer/releases/latest/download/update-windows-amd64.json` 10 seconds after
launch and then once a day; the About dialog, from the version button in the title bar, checks at once, and
Settings turns the automatic checks on or off. A newer version is offered in a window with its release notes, to install, skip
or be reminded of later. The update archive, or a smaller delta from the running version, is installed only when
its signature matches the public key in `mygo.json`. The app can update itself where it can write its own folder:
the per-user install in `%LOCALAPPDATA%\Programs\ANR FG Installer`, or the portable version's folder.
