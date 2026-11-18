package main

import (
	"context"
	"path/filepath"

	"github.com/egoist/mygo/ui"

	"anr-installer/internal/record"
)

// compareView is the comparison page of one of a game's recordings, shown in place of the game's page.
type compareView struct {
	game *game
	*compareState
}

// comparison describes rec replayed through the ANR FG installed in g and through g's DLSS-G.
func (a *app) comparison(g *game, rec *record.Recording) *record.Comparison {
	return &record.Comparison{Recording: rec, A: record.ANR(g.Scan.ANRDll), B: record.DLSSG(dlssgDLL(g.Scan)), FFmpeg: a.rec.tools.FFmpeg}
}

// openCompare shows the comparison of rec in the window. The videos open paused at their start.
func (a *app) openCompare(g *game, rec *record.Recording) {
	a.closeCompare()
	a.cmp = &compareView{game: g, compareState: openCompare(a.comparison(g, rec), a.update)}
}

// closeCompare leaves the comparison page.
func (a *app) closeCompare() {
	if a.cmp == nil {
		return
	}
	a.cmp.close()
	a.cmp = nil
}

func (a *app) comparePage(c *ui.Context, t *ui.Theme) {
	p := a.cmp
	p.page(c, t, compareActions{
		subject: p.game.Name,
		busy:    a.busy,
		blocked: func() string { return a.compareBlocked(p.game) },
		back:    a.closeCompare,
		viewer:  func(args ...string) { a.openViewer(p.game, args...) },
		export:  func() { a.exportStills(p) },
	})
}

// exportStills replays the frames around the moment shown through both sides and shows them as stills.
func (a *app) exportStills(p *compareView) {
	shown, cmp := p.stillsJob()
	real := shown.Real
	a.recordWork(p.game, gameTask(p.game, T("cmp.still.make", real)), func(ctx context.Context, r *reporter, tools func() (string, error)) error {
		dir, err := tools()
		if err != nil {
			return err
		}
		cmp.Tools = dir
		stills, err := cmp.Stills(ctx, real, func(label string) { r.say("%s", T("cmp.still.running", label)) }, r.out)
		if err != nil {
			return err
		}
		r.say("%s", T("cmp.still.done", len(stills), filepath.Dir(stills[0].A)))
		a.win.Update(func() { p.setStills(real, stills, shown) })
		return nil
	})
}
