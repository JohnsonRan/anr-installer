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

No binary release is published yet. Build it from source with Go 1.27 or later:

```
git clone https://github.com/JohnsonRan/anr-installer.git
cd anr-installer
go tool mygo build -skip-installer
```

The app is `build/windows-amd64/ANR FG Installer.exe`. [Developing](docs/development.md) has the other
commands.

## Use it

1. Close the game.
2. Open ANR FG Installer and choose the game on the left. "Supported" lists the games that ship DLSS frame
   generation; a game outside Steam can be added with **+ Add game** and its folder.
3. Click **Install** on the game's banner. The button reads **Upgrade**, **Reinstall** or **Downgrade** when
   ANR FG is already there, by the version chosen under **Install**.
4. In the game, turn on frame generation (the NVIDIA option where the game offers several).

**Uninstall** puts back what the package's installer backed up. Before a step loses one of your settings files,
the installer copies it next to the game's `.exe` as `<name>.bak-<time>`: `OptiScaler.ini` before the full
package replaces it, and `nr_fg.ini` before an uninstall or a downgrade removes it.

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
