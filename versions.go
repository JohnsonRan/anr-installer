package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"anr-installer/internal/detect"
	"anr-installer/internal/pkg"
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
			a.setReleases(rels)
			a.probe(rels)
		})
	}()
}

// setReleases keeps the releases and the versions offered for install, from pkg.MinInstallable on.
func (a *app) setReleases(rels []*pkg.Release) {
	a.releases, a.rel = rels, rels[0]
	a.versionRel, a.versionLabels = nil, nil
	for i, r := range rels {
		if pkg.Number(r.Version) < pkg.MinInstallable {
			continue
		}
		label := r.Version + " · " + r.Published.Local().Format("2006-01-02")
		if i == 0 {
			label = T("version.latest", label)
		}
		a.versionRel = append(a.versionRel, r)
		a.versionLabels = append(a.versionLabels, label)
	}
	if a.target() == nil && len(a.versionLabels) > 0 {
		a.versionSel = a.versionLabels[0]
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
				a.logf("%s", T("probe.failed", err.Error()))
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
	a.run(g, steps)
}

func (a *app) uninstall(g *game) {
	a.run(g, []step{a.uninstallStep(g.Scan.Installed)})
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

// run runs the steps for g in the background, one after the other, stopping at the first failure.
func (a *app) run(g *game, steps []step) {
	if a.busy || g.Scan == nil || len(steps) == 0 {
		return
	}
	a.busy, a.progress = true, -1
	scan := g.Scan
	gameDir := scan.ExeDir
	a.logf("── %s", g.Name)
	a.logf("%s", gameDir)

	update := func(fn func()) { a.win.Update(fn) }
	say := func(format string, args ...any) {
		line := fmt.Sprintf(format, args...)
		update(func() { a.status = line; a.logLines = append(a.logLines, line) })
	}
	go func() {
		err := func() error {
			ctx := context.Background()
			if err := backups(scan, steps, say); err != nil {
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
							a.status = T("downloading", name, fmt.Sprintf("%.1f / %.1f MB", float64(done)/1e6, float64(total)/1e6))
						})
					})
					if err != nil {
						return err
					}
					zipPath = p
					say("%s ✓ SHA-256", name)
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
				say("%s", T("running."+string(s.action), info.Version, s.kind))
				if err := pkg.Run(ctx, dir, s.action, gameDir, func(line string) {
					update(func() { a.logLines = append(a.logLines, "  "+line) })
				}); err != nil {
					return err
				}
				if s.action == pkg.Install {
					installed = info.Version
				}
			}
			if installed != "" {
				say("%s", T("done.install", installed, gameDir))
				if scan.NRRecord {
					say("%s", T("warn.nrrecord"))
				}
			} else {
				say("%s", T("done.uninstall", gameDir))
			}
			return nil
		}()
		update(func() {
			a.busy, a.progress = false, -2
			if err != nil {
				a.status = T("failed", err.Error())
				a.logLines = append(a.logLines, a.status)
			}
			a.rescan(g)
		})
	}()
}

func (a *app) logf(format string, args ...any) {
	a.logLines = append(a.logLines, fmt.Sprintf(format, args...))
}
