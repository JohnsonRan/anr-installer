# Changelog

## 0.1.0

The first release. ANR FG Installer is an unofficial tool, not made or endorsed by Apex Compute.

- Finds the games in every Steam library and marks those that ship DLSS frame generation, wherever in the
  game's folder its libraries are. Games outside Steam can be added by their folder, renamed and removed.
- Installs, upgrades, reinstalls, downgrades and uninstalls ANR FG with its own release packages, full or
  add-on, from any release from V.B04 on, or from a local zip. Downloads are checked against each release's
  SHA-256 sums.
- Tells the installed ANR FG version without downloading a package.
- Warns before installing into a game that ships anti-cheat: the risk is the user's.
- Backs up `OptiScaler.ini` and `nr_fg.ini` next to the game's `.exe` before a step would lose them.
- Shows each game's Steam art, in the system's light or dark appearance and accent color, in English or
  Simplified Chinese.
- Updates itself: it checks for a new version at launch and daily, and installs it after asking. The About
  dialog, from the version button in the title bar, checks at once, turns the automatic checks on or off, and
  links to the project.
- Comes as an installer, which installs for the current user without administrator rights, or as a portable
  zip that keeps its settings and downloads in its own folder. Either is one executable, its icon built in.
