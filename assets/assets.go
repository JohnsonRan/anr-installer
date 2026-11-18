// Package assets holds the files built into the executable: nothing ships beside it.
package assets

import _ "embed"

// LogoSVG is the app's logo, drawn in the title bar, the empty state and the About dialog.
//
//go:embed logo.svg
var LogoSVG []byte

// IconPNG is the app icon, for the update window and the notification area.
//
//go:embed icon.png
var IconPNG []byte
