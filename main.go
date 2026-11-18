// anr-installer installs, upgrades and removes ANR FG (Apex Neural Rendering frame generation) in Steam games:
// it finds the games, tells which ship DLSS frame generation, picks the full or add-on package, downloads and
// checks it, and runs the package's own installer.
package main

import (
	"log"

	"anr-installer/gui"
)

func main() {
	if err := gui.Run(); err != nil {
		log.Fatal(err)
	}
}
