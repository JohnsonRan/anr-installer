package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
	"anr-installer/internal/record"
)

// fetchRelease lists the releases, then reads the version of each package the cache lacks, so that an installed
// ANR FG shows its version without a download.
func (a *app) fetchRelease() {
	a.relLoading, a.relErr = true, ""
	go func() {
		rels, err := pkg.Releases(context.Background())
		a.win.Update(func() {
			a.relLoading = false
			if err != nil {
				a.relErr = err.Error()
				return
			}
			a.releasesFound(rels)
		})
	}()
}

// setReleases keeps the releases and the versions offered for install, from pkg.MinInstallable on.
func (a *app) setReleases(rels []*pkg.Release) {
	a.releases, a.rel = rels, rels[0]
	a.rec.pkg = record.Find(rels)
	a.versionRel = nil
	for _, r := range rels {
		if pkg.Number(r.Version) >= pkg.MinInstallable {
			a.versionRel = append(a.versionRel, r)
		}
	}
	a.versionLabels = a.labelVersions()
	if a.target() == nil && len(a.versionLabels) > 0 {
		a.versionSel = a.versionLabels[0]
	}
}

// labelVersions names the versions offered, each with its date, the newest marked in the window's language.
func (a *app) labelVersions() []string {
	var labels []string
	for _, r := range a.versionRel {
		label := r.Version + " · " + r.Published.Local().Format("2006-01-02")
		if r == a.rel {
			label = T("version.latest", label)
		}
		labels = append(labels, label)
	}
	return labels
}

// relabelVersions names the versions offered again after the language changed, keeping the one chosen, which
// versionSel holds by its name.
func (a *app) relabelVersions() {
	chosen := a.target()
	a.versionLabels = a.labelVersions()
	for i, r := range a.versionRel {
		if r == chosen {
			a.versionSel = a.versionLabels[i]
		}
	}
}

func (a *app) probe(rels []*pkg.Release) {
	a.probing = true
	cache := a.versions
	go func() {
		err := cache.Probe(context.Background(), rels, func(info *pkg.Info) {
			a.win.Update(func() { a.providers[info.ProviderSHA256] = info.Version })
		})
		a.win.Update(func() {
			a.probing = false
			if err != nil {
				a.note(levelWarning, T("probe.failed", err.Error()))
			}
		})
	}()
}

// target returns the release chosen to install from GitHub, or nil.
func (a *app) target() *pkg.Release {
	for i, l := range a.versionLabels {
		if l == a.versionSel {
			return a.versionRel[i]
		}
	}
	return nil
}

// targetVersion is the version an install would put in, or "".
func (a *app) targetVersion() string {
	if a.source == sourceLocal {
		if a.local != nil {
			return a.local.Version
		}
		return ""
	}
	if r := a.target(); r != nil {
		return r.Version
	}
	return ""
}

// versionOf returns the version of the ANR FG installed in r, or "" when it is not known.
func (a *app) versionOf(r *detect.Result) string {
	if r.ANRHash == "" {
		return ""
	}
	return a.providers[r.ANRHash]
}

// installedVersion describes the version of ANR FG installed in r.
func (a *app) installedVersion(r *detect.Result) string {
	if v := a.versionOf(r); v != "" {
		return v
	}
	if a.probing || a.relLoading {
		return T("inst.version.probing")
	}
	return T("inst.version.unk")
}

// Plan kinds: what the install button does.
type planKind int

const (
	planInstall   planKind = iota // nothing installed
	planUpgrade                   // a newer version over the installed one
	planReinstall                 // the same version again
	planDowngrade                 // an older version: uninstall, then install
	planSwitch                    // the installed version is unknown and the target is not the newest: uninstall, then install
	planUpdate                    // the installed version is unknown and the target is the newest: install over it
)

// plan returns what installing the chosen version into r does.
func (a *app) plan(r *detect.Result) planKind {
	if r.Installed == detect.KindNone {
		return planInstall
	}
	tv := a.targetVersion()
	iv := a.versionOf(r)
	if iv == "" {
		if a.rel != nil && tv == a.rel.Version {
			return planUpdate
		}
		return planSwitch
	}
	switch t, i := pkg.Number(tv), pkg.Number(iv); {
	case t > i:
		return planUpgrade
	case t == i:
		return planReinstall
	default:
		return planDowngrade
	}
}

// uninstallsFirst reports whether a plan removes the installed version before installing: going back to an older
// version over a newer one would leave the newer files the older installer does not know of.
func (p planKind) uninstallsFirst() bool { return p == planDowngrade || p == planSwitch }

func (a *app) planLabel(p planKind) string {
	v := a.targetVersion()
	switch p {
	case planUpgrade:
		return T("upgrade", v)
	case planReinstall:
		return T("reinstall")
	case planDowngrade:
		return T("downgrade", v)
	case planSwitch:
		return T("switch", v)
	case planUpdate:
		return T("install.v", v)
	}
	if v == "" {
		return T("install")
	}
	return T("install.v", v)
}

// sourceBlocked returns why the chosen source cannot install kind k, or "".
func (a *app) sourceBlocked(k detect.Kind) string {
	if a.source == sourceLocal {
		if a.local == nil {
			return T("why.no-zip")
		}
		if a.local.Kind != k {
			return T("why.zip-kind", a.local.Kind, k)
		}
		return ""
	}
	if a.target() == nil {
		return T("why.no-release")
	}
	return ""
}

// uninstallBlocked returns why ANR FG of kind k cannot be uninstalled, or "". Uninstalling takes the newest
// release's package, whose uninstaller knows every layout (custom/nrfg too); a local zip of that kind serves when
// GitHub cannot be reached.
func (a *app) uninstallBlocked(k detect.Kind) string {
	if a.rel != nil {
		return ""
	}
	if a.local != nil && a.local.Kind == k {
		return ""
	}
	return T("why.no-release")
}

// step is one run of a package's script.
type step struct {
	action pkg.Action
	kind   detect.Kind
	rel    *pkg.Release // the package from GitHub, or
	zip    string       // a local package
}

func (a *app) uninstallStep(k detect.Kind) step {
	if a.rel != nil {
		return step{pkg.Uninstall, k, a.rel, ""}
	}
	return step{pkg.Uninstall, k, nil, a.localZip}
}

func (a *app) installStep(k detect.Kind) step {
	if a.source == sourceLocal {
		return step{pkg.Install, k, nil, a.localZip}
	}
	return step{pkg.Install, k, a.target(), ""}
}

// install installs the chosen version of kind k into g, removing the installed one first when the plan says so.
func (a *app) install(g *game, k detect.Kind) {
	p := a.plan(g.Scan)
	var steps []step
	if p.uninstallsFirst() {
		steps = append(steps, a.uninstallStep(g.Scan.Installed))
	}
	steps = append(steps, a.installStep(k))
	a.run(g, a.planLabel(p), steps)
}

func (a *app) uninstall(g *game) {
	a.run(g, T("uninstall"), []step{a.uninstallStep(g.Scan.Installed)})
}

// backups copies what the steps would lose: OptiScaler.ini, which the full package replaces or its uninstaller
// restores, and nr_fg.ini, which an uninstall removes. The copies go next to the game's .exe.
func backups(r *detect.Result, steps []step, say func(string, ...any)) error {
	uninstalls, fullInstall := false, false
	for _, s := range steps {
		uninstalls = uninstalls || s.action == pkg.Uninstall
		fullInstall = fullInstall || (s.action == pkg.Install && s.kind == detect.KindFull)
	}
	stamp := time.Now().Format("20060102-150405")
	var files []string
	if r.Installed == detect.KindFull && r.OptiINI && (fullInstall || uninstalls) {
		files = append(files, filepath.Join(r.ExeDir, "OptiScaler.ini"))
	}
	if uninstalls && r.ANRDll != "" {
		files = append(files, filepath.Join(filepath.Dir(r.ANRDll), "nr_fg.ini"))
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		bak := filepath.Join(r.ExeDir, filepath.Base(f)+".bak-"+stamp)
		if err := os.WriteFile(bak, data, 0o644); err != nil {
			return err
		}
		say("%s", T("backup.file", filepath.Base(f), filepath.Base(bak)))
	}
	return nil
}

// run runs the steps for g in the background, one after the other, stopping at the first failure, as the task
// named action in the log.
func (a *app) run(g *game, action string, steps []step) {
	if a.busy || g.Scan == nil || len(steps) == 0 {
		return
	}
	a.busy, a.progress, a.logOpen = true, -1, true
	scan := g.Scan
	gameDir := scan.ExeDir
	r := &reporter{a: a, tk: a.begin(gameTask(g, action)), opened: true}
	update := func(fn func()) { a.win.Update(fn) }
	go func() {
		err := func() error {
			ctx := context.Background()
			if err := backups(scan, steps, r.say); err != nil {
				return err
			}
			var installed string
			for _, s := range steps {
				zipPath := s.zip
				if s.rel != nil {
					name := s.rel.AssetName(s.kind)
					update(func() { a.status = T("downloading", name, ""); a.progress = 0 })
					p, err := pkg.Download(ctx, s.rel, s.kind, func(done, total int64) {
						update(func() {
							if total > 0 {
								a.progress = float64(done) / float64(total)
							}
							a.status = T("downloading", name, sizeText(done)+" / "+sizeText(total))
						})
					})
					if err != nil {
						return err
					}
					zipPath = p
					r.say("%s ✓ SHA-256", name)
				}
				update(func() { a.progress = -1 })
				info, err := pkg.ReadInfo(zipPath)
				if err != nil {
					return err
				}
				if info.Kind != s.kind {
					return fmt.Errorf("%s", T("why.zip-kind", info.Kind, s.kind))
				}
				update(func() { a.providers[info.ProviderSHA256] = info.Version })
				dir, err := pkg.Extract(zipPath)
				if err != nil {
					return err
				}
				r.say("%s", T("running."+string(s.action), info.Version, s.kind))
				if err := pkg.Run(ctx, dir, s.action, gameDir, r.out); err != nil {
					return err
				}
				if s.action == pkg.Install {
					installed = info.Version
				}
			}
			if installed != "" {
				r.say("%s", T("done.install", installed, gameDir))
				if scan.NRRecord {
					r.warn(T("warn.nrrecord"))
				}
			} else {
				r.say("%s", T("done.uninstall", gameDir))
			}
			return nil
		}()
		update(func() {
			a.busy, a.progress = false, -2
			a.end(r.tk, failure(err))
			a.rescan(g)
			a.next()
		})
	}()
}

// outdated returns the newest version when the version of ANR FG installed in r is known and older than it.
// An unknown version is not reported: it may be a build GitHub does not list.
func (a *app) outdated(r *detect.Result) (string, bool) {
	if r == nil || a.rel == nil || r.Installed == detect.KindNone {
		return "", false
	}
	iv := a.versionOf(r)
	if iv == "" || pkg.Number(iv) >= pkg.Number(a.rel.Version) {
		return "", false
	}
	return a.rel.Version, true
}

// outdatedCount is how many games in the list have an older ANR FG than the newest release.
func (a *app) outdatedCount() int {
	n := 0
	for _, g := range a.games {
		if _, ok := a.outdated(g.Scan); ok {
			n++
		}
	}
	return n
}

// upgradable returns the games whose ANR FG is older than the newest release and can be upgraded together, and
// the names of those left out because they ship anti-cheat: each of them needs its own confirmation.
func (a *app) upgradable() (games []*game, anticheat []string) {
	for _, g := range a.games {
		if _, ok := a.outdated(g.Scan); !ok || g.Scanning {
			continue
		}
		if len(g.Scan.AntiCheat) > 0 {
			anticheat = append(anticheat, g.Name)
			continue
		}
		if ok, _ := g.Scan.CanInstall(g.Scan.Installed); ok {
			games = append(games, g)
		}
	}
	return games, anticheat
}

// upgradeAll upgrades the games to the newest release one after the other, each with the kind it has.
func (a *app) upgradeAll() {
	games, _ := a.upgradable()
	if len(games) == 0 || a.busy {
		return
	}
	a.note(levelInfo, T("queue.start", len(games), a.rel.Version))
	a.queue = games
	a.next()
}

// next starts the upgrade of the next game in the queue.
func (a *app) next() {
	for len(a.queue) > 0 && !a.busy {
		g := a.queue[0]
		a.queue = a.queue[1:]
		if _, ok := a.outdated(g.Scan); !ok || a.rel == nil {
			continue
		}
		a.run(g, T("upgrade", a.rel.Version), []step{{pkg.Install, g.Scan.Installed, a.rel, ""}})
	}
}

// autoNotes opens the notes of a new ANR FG release by themselves, once a release, when a game runs an older one.
func (a *app) autoNotes() {
	if a.rel == nil || a.saved == nil || a.saved.SeenRelease == a.rel.Tag || a.notesRel != nil || a.aboutOpen || a.busy {
		return
	}
	if games, _ := a.upgradable(); len(games) == 0 && a.outdatedCount() == 0 {
		return
	}
	a.saved.SeenRelease = a.rel.Tag
	a.saveSettings()
	a.showNotes(a.rel)
}
