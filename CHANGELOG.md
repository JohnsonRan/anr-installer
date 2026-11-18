# Changelog

## Unreleased

- When the list of ANR FG releases cannot be loaded, a game's page says why, as "Couldn't connect to GitHub" with
  the error, instead of "The version list hasn't loaded yet." An installed game no longer shows "Switch to" or a
  downgrade note without a version meanwhile.

## 0.3.0

- Adds a settings page, from the gear button in the title bar. The language follows the system by default, or
  is English, Simplified Chinese, Traditional Chinese or Japanese; the update window follows it too. The close button quits the app by
  default, or hides the window to the notification area, whose icon brings it back.
- Starting the app again while it runs brings its window forward instead of opening a second one.
- The first window fits the display: 1280 × 900 where there is room, else 90% of the display's work area, so
  its title bar stays on screen on a small laptop display. Later launches keep the size the user left it at.
- MyGo's files (the language chosen, the window's size and place, the update preference) move to
  `%APPDATA%\anr-installer` with `games.json`, and the portable version keeps them in its own folder. The
  window's size and the update preference start from their defaults once; `%APPDATA%\ANR FG Installer` can be
  deleted.
- The automatic checks for a new version of the installer are turned on or off in Settings. The About dialog
  keeps the update status and **Check for updates**.
- Closing or quitting while an install, an uninstall, a replay or another task runs asks first, as quitting can
  leave the task unfinished and a game's files half changed. The end of a Windows session is never held up.
- Settings shows the data folder, and the size of the downloaded ANR FG and nr-record packages. **Delete
  downloads** removes them; they download again when needed. Comparisons and the list of package versions stay.
- While the window is hidden or minimized, a Windows notification tells of a new ANR FG version, once a release.
  The app looks for one at launch and every 6 hours. It can be turned off in Settings.
- **Start with Windows** in Settings starts the app when the user signs in: hidden in the notification area when
  the close button minimizes to the tray, else minimized. Uninstalling the installed version removes its entry;
  the entry of a portable copy stays.
- The activity log lists each install, uninstall, upgrade, replay, export and deletion as a task, with its start
  time, its duration and whether it succeeded or failed, and its steps under it. What the programs printed is
  folded under each task, and a task that fails unfolds it. **Copy log** copies the whole log as text. A failed
  background check for a new ANR FG version is logged as a warning.
- Deleting a recording no longer leaves the window busy without deleting it.
- In a narrow window, a game's buttons, a page's actions, an info bar's button and a card's controls go under
  their text instead of cutting it off, and the filter of the game list shows its labels whole in every language.

## 0.2.0

- Records and compares frame generation with nr-record: installs it into a game, lists the recordings, and
  replays one through ANR FG and the game's DLSS-G. The two videos play together in the window, with a divider
  or side by side, and a paused frame can be exported as lossless stills; nr-record's own viewer opens from
  there too. A comparison takes the first 5 seconds of a recording unless more is chosen, and a recording can
  be deleted from its row. FFmpeg and FFprobe, when missing, are installed with Scoop.
- Starts a game from its page.

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
- Lays out a game as Windows 11's Settings does, under its Steam library art, in the dark colors of
  apexcompute.com, in English or Simplified Chinese.
- Tells when a newer ANR FG is out: the list marks the games to upgrade, and the release notes open once a
  release, in the window, with **Update all** for the games without anti-cheat.
- Updates itself: it checks for a new version at launch and daily, and installs it after asking. The About
  dialog, from the version button in the title bar, checks at once, turns the automatic checks on or off, and
  links to the project.
- Comes as an installer, which installs for the current user without administrator rights, or as a portable
  zip that keeps its settings and downloads in its own folder. Either is one executable, its icon built in.
