package main

import "github.com/egoist/mygo"

// closeHides reports whether the window's close button hides the window rather than closing it: only with the
// setting on and a tray icon to bring the window back by, and never while the app quits (Quit from the tray, an
// update's restart, the end of the session), which must close it.
func closeHides(toTray, hasTray, quitting bool) bool {
	return toTray && hasTray && !quitting
}

// showWindow brings the window forward, from the tray or a second launch.
func (a *app) showWindow() {
	a.win.Show()
	if a.win.IsMinimized() {
		a.win.Restore()
	}
	a.win.Focus()
}

// trayMenu pops up the tray icon's menu. The icon holds it only while it shows: an icon with a menu opens it on
// every click, and a click is to bring the window back, as Windows' own tray icons do.
func (a *app) trayMenu() {
	a.tray.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Label: T("tray.show"), Click: func(*mygo.MenuItem, *mygo.Window) { a.showWindow() }},
		mygo.Separator(),
		{Label: T("tray.quit"), Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	}))
	defer a.tray.SetMenu(nil)
	a.tray.PopUpMenu()
}

// syncTray adds the tray icon while the close button hides the window, and removes it otherwise. When the icon
// cannot be added, closing quits, as no way back to a hidden window would be left.
func (a *app) syncTray() {
	if a.win == nil { // tests: no MyGo app runs
		return
	}
	switch {
	case a.saved.CloseToTray && a.tray == nil:
		tray, err := mygo.NewTray(mygo.TrayOptions{ID: "main", Icon: iconPNG, ToolTip: T("title")})
		if err != nil {
			a.note(levelError, T("tray.failed", err.Error()))
			return
		}
		tray.OnClick(a.showWindow)
		tray.OnRightClick(a.trayMenu)
		a.tray = tray
	case !a.saved.CloseToTray && a.tray != nil:
		a.tray.Destroy()
		a.tray = nil
	}
}

// relabelTray says the tray icon's tooltip in the window's language; the menu is made as it pops up.
func (a *app) relabelTray() {
	if a.tray != nil {
		a.tray.SetToolTip(T("title"))
	}
}

// removeTray takes the icon out of the notification area as the app quits, where it would stay until the
// pointer passes over it.
func (a *app) removeTray() {
	if a.tray != nil {
		a.tray.Destroy()
		a.tray = nil
	}
}
