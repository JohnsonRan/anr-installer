package main

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
	"anr-installer/internal/record"
)

// scoopURL is where Scoop says how to install it.
const scoopURL = "https://scoop.sh"

// maxRecordings is how many of a game's recordings its page lists, the newest first; the folder holds the rest.
const maxRecordings = 5

// recState is what the "Record and compare" section of a game's page shows, read off the main thread by
// refreshRecord.
type recState struct {
	pkg        *record.Package     // the nr-record package of the newest release that carries one, or nil
	tools      record.Tools        // FFmpeg, FFprobe and Scoop, as found
	recordings []*record.Recording // every recording, the newest first
	compared   map[string]bool     // by a recording's name: its comparison was made before
}

// refreshRecord looks again for the tools nr-record needs and for the recordings.
func (a *app) refreshRecord() {
	go func() {
		tools := record.FindTools()
		recs := record.Recordings(record.RecordingsDir())
		compared := map[string]bool{}
		for _, r := range recs {
			compared[r.Name] = (&record.Comparison{Recording: r}).Done()
		}
		a.update(func() { a.rec.tools, a.rec.recordings, a.rec.compared = tools, recs, compared })
	}()
}

// recordingsOf returns the recordings made in g.
func (a *app) recordingsOf(g *game) []*record.Recording {
	var recs []*record.Recording
	for _, r := range a.rec.recordings {
		if g.Scan != nil && r.Of(g.Scan.Exe) {
			recs = append(recs, r)
		}
	}
	return recs
}

// recordSection is the "Record and compare" group of a game's page: nr-record in the game, the programs its
// tools need, and the game's recordings, each with its comparison of ANR FG and DLSS-G. nr-record's tools
// for Windows are the ones driven, so the section shows on Windows only.
func (a *app) recordSection(c *ui.Context, t *ui.Theme, g *game) {
	if runtime.GOOS != "windows" {
		return
	}
	r := g.Scan
	recs := a.recordingsOf(g)
	idle := !a.busy && !g.Scanning
	// A row describes what it is; what the user has to read or do is said in an info bar.
	refresh := func() { ui.Button(c, T("rec.refresh")).Tooltip(T("rec.refresh.tip")).OnClick(a.refreshRecord) }
	section(c, t, T("rec.title"), refresh, func() {
		installed := g.Recorder
		switch {
		case installed && r.NRRecord:
			settingsRow(c, t, iconRecord, "nr-record", T("rec.nrrecord"), func() {
				statusValue(c, T("rec.installed"), t.Success)
				ui.Button(c, T("rec.uninstall")).Disabled(!idle).OnClick(func() { a.recordInstall(g, pkg.Uninstall) })
			})
			infoBar(c, t, apexRoyal, iconInfo, T("rec.nrrecord.on"), nil)
		case installed:
			settingsRow(c, t, iconRecord, "nr-record", T("rec.nrrecord"), func() {
				statusValue(c, T("rec.unused"), t.Warning)
				ui.Button(c, T("rec.uninstall")).Disabled(!idle).OnClick(func() { a.recordInstall(g, pkg.Uninstall) })
			})
			infoBar(c, t, t.Warning, iconWarning, T("rec.nrrecord.off"), func() {
				ui.Button(c, T("rec.reinstall")).Disabled(!idle).OnClick(func() { a.recordInstall(g, pkg.Install) })
			})
		case r.Installed != detect.KindFull:
			settingsRow(c, t, iconRecord, "nr-record", T("rec.nrrecord.needs"), func() { valueText(c, T("anr.none"), t.TextMuted) })
		default:
			settingsRow(c, t, iconRecord, "nr-record", T("rec.nrrecord"), func() {
				ui.Button(c, T("rec.install")).Disabled(!idle).OnClick(func() { a.recordInstall(g, pkg.Install) })
			})
		}
		if !installed && len(recs) == 0 {
			return
		}

		tools := a.rec.tools
		switch {
		case tools.Ready():
			settingsRow(c, t, iconFilm, T("rec.ffmpeg"), T("rec.ffmpeg.detail"), func() { statusValue(c, T("rec.found"), t.Success) })
		case tools.Scoop != "":
			settingsRow(c, t, iconFilm, T("rec.ffmpeg"), T("rec.ffmpeg.detail"), func() {
				statusValue(c, T("rec.missing"), t.Warning)
				ui.Button(c, T("rec.scoop.install")).Disabled(a.busy).OnClick(func() { a.installFFmpeg(g) })
			})
		default:
			settingsRow(c, t, iconFilm, T("rec.ffmpeg"), T("rec.ffmpeg.detail"), func() { statusValue(c, T("rec.missing"), t.Warning) })
			infoBar(c, t, t.Warning, iconWarning, T("rec.ffmpeg.noscoop"), func() {
				ui.Button(c, T("rec.scoop.get")).Tooltip(scoopURL).OnClick(func() { mygo.Shell.OpenExternal(scoopURL) })
				ui.Button(c, T("rec.check")).OnClick(a.refreshRecord)
			})
		}

		dir := record.RecordingsDir()
		count := T("rec.recordings.none")
		if len(a.rec.recordings) > 0 {
			count = T("rec.recordings.count", len(recs), len(a.rec.recordings))
		}
		settingsRow(c, t, iconFolder, T("rec.recordings"), count, func() {
			if len(a.rec.recordings) > 0 {
				ui.Button(c, T("open.folder")).OnClick(func() { mygo.Shell.OpenPath(dir) })
			}
			ui.Button(c, T("rec.browse")).Tooltip(T("rec.browse.tip")).Disabled(a.busy).OnClick(func() { a.openViewer(g, "--captures") })
		})
		why := a.compareBlocked(g)
		if why != "" && tools.Ready() && len(recs) > 0 { // without the tools, the bar above says what to do
			infoBar(c, t, t.Warning, iconWarning, why, nil)
		}
		if len(recs) > 0 {
			infoBar(c, t, apexRoyal, iconInfo, T("rec.slow"), nil)
		}
		for _, rec := range recs[:min(len(recs), maxRecordings)] {
			a.recordingRow(c.Key(rec.Name), t, g, rec, why != "" || a.busy)
		}
	})
}

// dlssgDLL returns the game's own nvngx_dlssg.dll, which a recording replays through as DLSS-G, or "".
func dlssgDLL(r *detect.Result) string {
	for _, p := range r.DLSSG {
		if strings.EqualFold(filepath.Base(p), "nvngx_dlssg.dll") {
			return p
		}
	}
	return ""
}

// compareBlocked returns why a recording of g cannot be compared now, or "".
func (a *app) compareBlocked(g *game) string {
	switch r := g.Scan; {
	case !a.rec.tools.Ready():
		return T("rec.why.tools")
	case r.ANRDll == "" || r.Installed == detect.KindNone:
		return T("rec.why.anr")
	case dlssgDLL(r) == "":
		return T("rec.why.dlssg")
	}
	return ""
}

// recordWork runs work in the background as the task in progress, named title in the activity log, and rescans g
// after it. work gets r, which writes the task's steps and output, and tools, which returns the unpacked
// nr-record package, downloading it when the cache lacks it. The log opens once a step has something to say.
func (a *app) recordWork(g *game, title string, work func(ctx context.Context, r *reporter, tools func() (string, error)) error) {
	if a.busy {
		return
	}
	a.busy, a.progress = true, -1
	p, listed := a.rec.pkg, a.rel != nil
	update := func(fn func()) { a.win.Update(fn) }
	r := &reporter{a: a, tk: a.begin(title)}
	go func() {
		ctx := context.Background()
		err := work(ctx, r, func() (string, error) {
			dir, err := record.Unpack(ctx, p, func(done, total int64) {
				update(func() {
					if total > 0 {
						a.progress = float64(done) / float64(total)
					}
					a.status = T("downloading", p.Asset(), sizeText(done)+" / "+sizeText(total))
				})
			})
			update(func() { a.progress = -1 })
			switch {
			case errors.Is(err, record.ErrNoPackage) && listed:
				err = errors.New(T("rec.why.no-asset"))
			case errors.Is(err, record.ErrNoPackage):
				err = errors.New(T("rec.why.no-package"))
			}
			return dir, err
		})
		update(func() {
			a.busy, a.progress = false, -2
			a.end(r.tk, failure(err))
			a.rescan(g)
			a.refreshRecord()
		})
	}()
}

// recordInstall runs the nr-record package's installer or uninstaller for g. The package's own scripts do the
// work, as with ANR FG: they keep and restore OptiScaler.ini.
func (a *app) recordInstall(g *game, action pkg.Action) {
	gameDir := g.Scan.ExeDir
	a.recordWork(g, gameTask(g, T("rec."+string(action))), func(ctx context.Context, r *reporter, tools func() (string, error)) error {
		dir, err := tools()
		if err != nil {
			return err
		}
		r.say("%s", T("rec.running."+string(action)))
		if err := pkg.Run(ctx, dir, action, gameDir, r.out); err != nil {
			return err
		}
		r.say("%s", T("rec.done."+string(action), gameDir))
		return nil
	})
}

// installFFmpeg installs FFmpeg and FFprobe with Scoop, for every game.
func (a *app) installFFmpeg(g *game) {
	scoop := a.rec.tools.Scoop
	a.recordWork(g, T("rec.ffmpeg.task"), func(ctx context.Context, r *reporter, _ func() (string, error)) error {
		r.say("%s", T("rec.scoop.running"))
		if err := record.InstallFFmpeg(ctx, scoop, r.out); err != nil {
			return err
		}
		if !record.FindTools().Ready() {
			return errors.New(T("rec.scoop.missing"))
		}
		r.say("%s", T("rec.scoop.done"))
		return nil
	})
}

// compare replays rec through the ANR FG installed in g and through g's DLSS-G, then shows the two videos
// on the comparison page. secs is how much of the recording's start to replay, 0 for all of it.
func (a *app) compare(g *game, rec *record.Recording, secs int) {
	cmp := a.comparison(g, rec)
	if secs > 0 {
		fps := rec.RealFPS()
		if fps <= 0 {
			fps = assumedFPS // the length is not known: the replay is still kept short
		}
		cmp.Frames = int(float64(secs)*fps + 0.5)
	}
	a.recordWork(g, gameTask(g, T("rec.compare")), func(ctx context.Context, r *reporter, tools func() (string, error)) error {
		dir, err := tools()
		if err != nil {
			return err
		}
		cmp.Tools = dir
		side := ""
		step := func(label string) {
			side = label
			r.say("%s", T("rec.replaying", label))
		}
		if err := cmp.Make(ctx, step, r.out); err != nil {
			if side == cmp.B.Label {
				r.warn(T("rec.hint.dlssg"))
			}
			return err
		}
		r.say("%s", T("rec.done.compare", cmp.Out()))
		a.win.Update(func() { a.openCompare(g, rec) })
		return nil
	})
}

// openViewer opens nr-record's viewer with args.
func (a *app) openViewer(g *game, args ...string) {
	t := a.rec.tools
	a.recordWork(g, gameTask(g, T("rec.browse")), func(ctx context.Context, _ *reporter, tools func() (string, error)) error {
		dir, err := tools()
		if err != nil {
			return err
		}
		return record.View(dir, t, args...)
	})
}
