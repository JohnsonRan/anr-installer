// genicon renders assets/logo.svg into assets/icon.png, the 1024×1024 app icon mygo build packages:
//
//	go run ./tools/genicon
package main

import (
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo/ui"
)

func main() {
	data, err := os.ReadFile("assets/logo.svg")
	if err != nil {
		log.Fatal(err)
	}
	svg, err := ui.ParseSVG(data)
	if err != nil {
		log.Fatal(err)
	}
	img := ui.Render(func(c *ui.Context) {
		c.Root().Background(ui.Transparent)
		ui.Image(c, svg).Size(1024, 1024)
	}, 1024, 1024, 1)
	f, err := os.Create("assets/icon.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
