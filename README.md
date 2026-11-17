# ANR FG Installer

English | [简体中文](README.zh-CN.md) | [日本語](README.ja.md)

ANR FG Installer installs, upgrades, downgrades and removes [ANR FG](https://github.com/apex-compute/ANR)
(Apex Neural Rendering frame generation) in your games. It finds your Steam games, tells which ship DLSS frame
generation, picks the right package, and runs the package's own installer.

> **Unofficial.** This is a community tool. It is not made, endorsed or supported by Apex Compute, the makers
> of ANR FG. It installs ANR FG's own release packages from the
> [apex-compute/ANR releases](https://github.com/apex-compute/ANR/releases), unchanged. Report problems with
> ANR FG itself to [apex-compute/ANR](https://github.com/apex-compute/ANR/issues), and problems with this
> installer [here](https://github.com/JohnsonRan/anr-installer/issues).

## Requirements

- Windows 10 or 11 (x64)
- The Microsoft Visual C++ 2015-2022 Redistributable (x64), which OptiScaler needs
- A GPU that ANR FG supports: NVIDIA GeForce RTX 20 and newer or GTX 16; AMD Radeon RX 5000 and newer;
  Intel Arc A and B

Linux and the Steam Deck (Proton) are written for but not tested yet.

## Get it

Download `anr-fg-installer-setup-<version>-amd64.exe` from the
[latest release](https://github.com/JohnsonRan/anr-installer/releases/latest) and run it. It installs the app for
your Windows account only, in `%LOCALAPPDATA%\Programs\ANR FG Installer`, without administrator rights. Or install
it from PowerShell:

```
irm https://github.com/JohnsonRan/anr-installer/releases/latest/download/install.ps1 | iex
```

The portable version, `anr-fg-installer-<version>-windows-amd64-portable.zip`, needs no install: extract its
`ANR FG Installer` folder to a place you can write to and run `ANR FG Installer.exe`. It keeps its settings and
downloads in that folder.

The app and its installer are not code-signed, so Windows SmartScreen may warn before the first run: choose
**More info**, then **Run anyway**.

The app updates itself. It checks for a new version at launch and once a day, and the version button in its title
bar checks at once. It installs only updates signed with the project's key. To build it from source instead, see
[Developing](docs/development.md).

## Use it

1. Close the game.
2. Open ANR FG Installer and choose the game on the left. **Supported** lists the games that ship DLSS frame
   generation. To add a game from another store, select **Add game** and choose its folder.
3. Select **Install** at the top of the game's page. When ANR FG is already there, the button reads **Upgrade**,
   **Reinstall** or **Downgrade**, for the version chosen with the arrow beside it.
4. In the game, turn on frame generation (the NVIDIA option where the game offers several).

**Uninstall** puts back what the package's installer backed up. Before a step loses one of your settings files,
the installer copies it next to the game's `.exe` as `<name>.bak-<time>`: `OptiScaler.ini` before the full
package replaces it, and `nr_fg.ini` before an uninstall or a downgrade removes it.

When a new ANR FG version comes out and one of your games has an older one, the installer shows the release
notes once. **Update all** there upgrades those games one after another, each with the package it has. Games
with anti-cheat are left out: upgrade each of them from its own page. The list marks the games to upgrade, and
the bar above it shows only those.

## Record and compare

[nr-record](https://github.com/apex-compute/ANR#record-and-compare-frame-generation), from the same releases,
records what a game gives its frame generation, so that ANR FG and DLSS-G can be compared on the same frames. A
game's page shows **Record and compare** once ANR FG's full package is installed in it.

1. Click **Install nr-record**. Start the game and turn on DLSS frame generation. **F6** starts and stops a
   recording; **F9** records 8 frames.
2. Back in the installer, click **Refresh**. Each recording of the game has **Compare with DLSS-G**: it replays
   the recording through the installed ANR FG and through the game's own `nvngx_dlssg.dll`, then shows the two
   videos in the window. Replaying through DLSS-G needs an NVIDIA RTX 40 or 50 series GPU.
3. On the comparison page, play both videos together, drag the divider between them, or put them side by side.
   The videos are compressed and are not locked frame by frame: to compare a frame exactly, pause on it and
   choose **Stills**, which exports the frames around it as lossless PNG images. **Open in nr-record's viewer**
   opens the same comparison there, for pixel zoom, marks and exports.

The videos need FFmpeg and FFprobe. When they are missing, **Install with Scoop** installs them; without
[Scoop](https://scoop.sh), the installer asks you to install Scoop first. Installing ANR FG again switches frame
generation back to ANR FG, so install nr-record again afterwards to keep recording. Recordings of an upscaler's
inputs are listed, but replay from the command line only.

## Anti-cheat

OptiScaler and ANR FG hook into a game's rendering. Anti-cheat may take them for a cheat and ban the account,
in online modes above all. When a game ships anti-cheat (Easy Anti-Cheat, EA Javelin, BattlEye and others), the
installer marks it and asks before installing. Install into such a game only if you accept that risk; you bear
it yourself.

## How it works

[How it works](docs/how-it-works.md) covers what the installer reads and changes, how it tells the installed
version without downloading anything, and the files it keeps.

## License

[MIT](LICENSE). ANR FG and OptiScaler, which the installer downloads and installs, keep their own licenses.
