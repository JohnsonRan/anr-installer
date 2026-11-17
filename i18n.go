package main

import (
	"fmt"
	"strings"
)

// lang is the language the window shows: "zh" or "en".
var lang = "en"

func setLanguage(preferred []string) {
	for _, l := range preferred {
		l = strings.ToLower(l)
		switch {
		case strings.HasPrefix(l, "zh"):
			lang = "zh"
			return
		case strings.HasPrefix(l, "en"):
			lang = "en"
			return
		}
	}
}

// T returns the text of key in the window's language, formatted with args.
func T(key string, args ...any) string {
	m, ok := messages[key]
	if !ok {
		return key
	}
	s := m[0]
	if lang == "zh" {
		s = m[1]
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

// messages holds each text in English and Simplified Chinese.
var messages = map[string][2]string{
	// The window.
	"title":              {"ANR FG Installer", "ANR FG 安装器"},
	"subtitle":           {"Install Apex Neural Rendering frame generation in your games", "为游戏安装 Apex Neural Rendering 帧生成"},
	"about.open":         {"About and updates", "关于和更新"},
	"about.version":      {"Version %s", "版本 %s"},
	"about.installed":    {"Installed", "安装版"},
	"about.portable":     {"Portable", "便携版"},
	"about.updates":      {"Updates", "更新"},
	"about.checkNow":     {"Check for updates", "检查更新"},
	"about.auto":         {"Check for updates automatically", "自动检查更新"},
	"about.autoHint":     {"When the app starts and once a day", "在启动时和每天检查一次"},
	"about.lastCheck":    {"Last checked: %s", "上次检查时间: %s"},
	"about.neverChecked": {"Not checked yet", "尚未检查"},
	"about.available":    {"Version %s is available", "新版本 %s 可用"},
	"about.cantUpdate":   {"This copy can't update itself because it's a development build or its folder is read-only.", "此副本无法自行更新: 它是开发版本，或所在文件夹为只读。"},
	"about.github":       {"GitHub", "GitHub"},
	"about.changelog":    {"Changelog", "更新日志"},
	"about.issues":       {"Report a problem", "报告问题"},
	"about.license":      {"License", "许可证"},
	"about.anrfg":        {"ANR FG", "ANR FG"},
	"about.notice":       {"This is an unofficial tool. It isn't made or endorsed by Apex Compute. If a game uses anti-cheat, you install at your own risk.", "这是非官方工具，与 Apex Compute 无关。在使用反作弊的游戏中安装，风险由你自行承担。"},
	"about.credits":      {"ANR FG and OptiScaler come unchanged from the apex-compute/ANR release packages. Built with MyGo. Released under the MIT License.", "ANR FG 和 OptiScaler 直接取自 apex-compute/ANR 的发布包，未经修改。使用 MyGo 构建，按 MIT 许可证发布。"},
	"about.close":        {"Close", "关闭"},
	"win.minimize":       {"Minimize", "最小化"},
	"win.maximize":       {"Maximize", "最大化"},
	"win.restore":        {"Restore", "还原"},
	"win.close":          {"Close", "关闭"},

	// The list of games.
	"search":           {"Search games", "搜索游戏"},
	"show":             {"Filter", "筛选"},
	"show.all":         {"All", "全部"},
	"show.supported":   {"Supported", "受支持"},
	"show.installed":   {"Installed", "已安装"},
	"row.upgrade":      {"upgrade available", "可升级"},
	"updates.count":    {"ANR FG %[2]s available: %[1]d to upgrade", "%d 款游戏可升级到 %s"},
	"updates.show":     {"Show only", "只显示这些"},
	"updates.all":      {"Show all", "显示全部"},
	"update.available": {"ANR FG %s is available. This game has %s.", "ANR FG %s 已发布，这个游戏安装的是 %s。"},
	"games.count":      {"Games: %d · With DLSS-G: %d · With ANR FG: %d", "%d 款游戏 · %d 款支持 DLSS-G · %d 款已安装 ANR FG"},
	"add.game":         {"Add game", "添加游戏"},
	"add.game.tip":     {"Add a game from another store by choosing its folder.", "选择文件夹，添加 Steam 以外的游戏。"},
	"refresh":          {"Rescan", "重新扫描"},
	"refresh.tip":      {"Look for Steam games again.", "重新查找 Steam 游戏。"},
	"scanning.steam":   {"Looking for games…", "正在查找游戏…"},
	"scanning":         {"Scanning…", "正在扫描…"},
	"no.steam":         {"We couldn't find a Steam library. Add a game by choosing its folder.", "找不到 Steam 游戏库。可以选择文件夹来添加游戏。"},
	"no.games":         {"No games to show.", "没有可显示的游戏。"},
	"manual":           {"added manually", "手动添加"},
	"manual.long":      {"Added manually", "手动添加"},
	"row.supported":    {"Supports DLSS frame generation", "支持 DLSS 帧生成"},
	"row.unsupported":  {"No DLSS frame generation detected", "未检测到 DLSS 帧生成"},
	"unofficial":       {"Unofficial tool, not affiliated with Apex Compute", "非官方工具，与 Apex Compute 无关"},
	"unofficial.tip":   {"Report problems with ANR FG at apex-compute/ANR, and problems with this installer at JohnsonRan/anr-installer.", "ANR FG 本身的问题请报告到 apex-compute/ANR，安装器的问题请报告到 JohnsonRan/anr-installer。"},
	"empty.title":      {"No games yet", "还没有游戏"},
	"empty.body":       {"Steam games appear here automatically. To add a game from another store, choose its folder.", "Steam 游戏会自动显示在这里。要添加其他平台的游戏，请选择游戏文件夹。"},

	// A game.
	"rename.tip":  {"Double-click to rename", "双击以重命名"},
	"remove":      {"Remove from list", "从列表中移除"},
	"remove.tip":  {"Removes the game from this list only. Its files, including ANR FG, aren't changed.", "只从列表中移除，不会更改游戏文件(包括 ANR FG)。"},
	"removed":     {"Removed %s (%s) from the list.", "已从列表中移除 %s (%s)。"},
	"save.failed": {"Couldn't save the game list: %s", "无法保存游戏列表: %s"},
	"busy":        {"Working…", "正在处理…"},

	"anr.full":             {"Full package · OptiScaler loads as %s", "full 版 · OptiScaler 作为 %s 加载"},
	"anr.addon":            {"Addon package · uses your OptiScaler", "addon 版 · 使用你的 OptiScaler"},
	"anr.none":             {"Not installed", "未安装"},
	"dlss.title":           {"DLSS frame generation", "DLSS 帧生成"},
	"dlss.found":           {"Detected", "已检测到"},
	"dlss.root":            {"In the game folder", "位于游戏根目录"},
	"dlss.none":            {"Not detected", "未检测到"},
	"dlss.none.detail":     {"ANR FG works by replacing it", "ANR FG 通过替换它来工作"},
	"opti.ours":            {"Included with ANR FG", "ANR FG 自带"},
	"opti.yours":           {"Installed by you", "你自己安装的"},
	"opti.none":            {"Not installed", "未安装"},
	"opti.fg":              {"Frame generation: %s", "帧生成: %s"},
	"opti.none.detail":     {"The full package includes it", "full 版自带 OptiScaler"},
	"inst.version.unk":     {"Unknown version", "未知版本"},
	"inst.version.probing": {"Identifying version…", "正在识别版本…"},

	"dlssg.no":       {"No DLSS frame generation was detected in this game, so ANR FG might not work in it.", "未在这个游戏中检测到 DLSS 帧生成，ANR FG 可能无法工作。"},
	"note.ini.reset": {"The full package replaces OptiScaler.ini. The installer backs up your copy next to the .exe first.", "full 版会替换 OptiScaler.ini。安装器会先将原文件备份到 .exe 所在文件夹。"},
	"note.nrrecord":  {"nr-record is in use. Installing switches frame generation back to ANR FG, so to keep recording, run the nr-record installer again afterward.", "nr-record 正在使用。安装后帧生成会切换回 ANR FG；要继续录制，请在安装后重新运行 nr-record 安装程序。"},
	"note.proton":    {"On Proton, set the game's launch options to the line the installer shows.", "使用 Proton 时，请将游戏的启动选项设置为安装程序显示的那一行。"},
	"note.downgrade": {
		"To install an older version, the installer first uninstalls the current one with the %s uninstaller and then installs %s. OptiScaler.ini and nr_fg.ini are backed up next to the .exe. Your nr_fg.ini settings aren't carried over.",
		"安装旧版本时，会先用 %s 的卸载程序卸载当前版本，再安装 %s。OptiScaler.ini 和 nr_fg.ini 会备份到 .exe 所在文件夹，nr_fg.ini 中的设置不会保留。",
	},

	// Installing.
	"version":          {"Version", "版本"},
	"version.choose":   {"Choose a version to install", "选择要安装的版本"},
	"notes.title":      {"ANR FG %s release notes", "ANR FG %s 发行说明"},
	"notes.date":       {"Released %s", "发布日期: %s"},
	"notes.none":       {"This release has no notes.", "此版本没有发行说明。"},
	"notes.github":     {"View on GitHub", "在 GitHub 上查看"},
	"notes.whatsnew":   {"What's new", "查看更新内容"},
	"notes.updateAll":  {"Update all (%d)", "全部更新(%d)"},
	"notes.anticheat":  {"Games with anti-cheat aren't included. Update them one at a time from their pages: %s", "带反作弊的游戏不会一并更新，请在各自的页面中单独更新: %s"},
	"queue.start":      {"Updating games to %[2]s: %[1]d", "正在将 %d 款游戏更新到 %s"},
	"release.notes.of": {"Release notes for %s", "%s 发行说明"},
	"status":           {"Status", "状态"},
	"state.installed":  {"Installed %s · %s", "已安装 %s · %s"},
	"advanced.desc":    {"Package type and source", "安装包类型和来源"},
	"anr.none.detail":  {"Replaces the game's DLSS frame generation when installed", "安装后接替游戏的 DLSS 帧生成"},
	"version.latest":   {"%s (latest)", "%s (最新版本)"},
	"latest.loading":   {"Getting the version list…", "正在获取版本列表…"},
	"latest.error":     {"Couldn't connect to GitHub: %s", "无法连接到 GitHub: %s"},
	"retry":            {"Retry", "重试"},
	"advanced":         {"Advanced options", "高级选项"},
	"package":          {"Package", "安装包"},
	"package.help":     {"The full package includes OptiScaler. The addon package works with an OptiScaler you installed.", "full 版自带 OptiScaler；addon 版配合你自己安装的 OptiScaler 使用。"},
	"kind.auto":        {"Automatic (%s)", "自动(%s)"},
	"source":           {"Source", "来源"},
	"source.github0":   {"GitHub", "GitHub"},
	"source.local":     {"Local zip file", "本地 zip 文件"},
	"choose.zip":       {"Choose zip file…", "选择 zip 文件…"},
	"zip.none":         {"Choose a package zip file in Advanced options", "请在“高级选项”中选择安装包 zip 文件"},
	"zip.info":         {"%s · %s, %s package", "%s · %s，%s 版"},
	"pick.zip.title":   {"Choose an ANR FG package", "选择 ANR FG 安装包"},

	"install":   {"Install", "安装"},
	"install.v": {"Install %s", "安装 %s"},
	"upgrade":   {"Upgrade to %s", "升级到 %s"},
	"reinstall": {"Reinstall", "重新安装"},
	"downgrade": {"Downgrade to %s", "降级到 %s"},
	"switch":    {"Switch to %s", "切换到 %s"},
	"uninstall": {"Uninstall", "卸载"},

	"why.user-optiscaler": {"This game already has an OptiScaler you installed, and the full package won't install over it. In Advanced options, choose addon.", "这个游戏已有你自己安装的 OptiScaler，full 版无法覆盖安装。请在“高级选项”中选择 addon。"},
	"why.no-optiscaler":   {"The addon package needs OptiScaler.ini next to the .exe. Choose full, or install OptiScaler first.", "addon 版需要 .exe 所在文件夹中已有 OptiScaler.ini。请选择 full，或先安装 OptiScaler。"},
	"why.full-installed":  {"The full package is installed here, so keep using the full package.", "这里安装的是 full 版，请继续使用 full 版。"},
	"why.zip-kind":        {"The selected zip file is a %s package, not %s.", "所选 zip 文件是 %s 版，不是 %s 版。"},
	"why.no-zip":          {"Choose a package zip file first.", "请先选择安装包 zip 文件。"},
	"why.no-release":      {"The version list hasn't loaded yet.", "版本列表尚未加载。"},

	// Where it goes.
	"exe.folder":        {"Location", "位置"},
	"change":            {"Change…", "更改…"},
	"auto.exe":          {"Detect automatically", "自动检测"},
	"auto.exe.tip":      {"Go back to the folder the installer detected.", "恢复为自动检测到的文件夹。"},
	"open.folder":       {"Open folder", "打开文件夹"},
	"pick.folder.title": {"Choose the game folder", "选择游戏文件夹"},

	// Activity.
	"log":               {"Activity", "活动"},
	"log.empty":         {"No activity yet.", "暂无活动。"},
	"clear":             {"Clear", "清除"},
	"downloading":       {"Downloading %s… %s", "正在下载 %s… %s"},
	"running.install":   {"Installing %s (%s package)…", "正在安装 %s (%s 版)…"},
	"running.uninstall": {"Uninstalling with the %s uninstaller (%s package)…", "正在使用 %s 的卸载程序卸载(%s 版)…"},
	"backup.file":       {"Backed up %s to %s.", "已将 %s 备份到 %s。"},
	"done.install":      {"ANR FG %s is installed in %s.", "已将 ANR FG %s 安装到 %s。"},
	"done.uninstall":    {"ANR FG is uninstalled from %s.", "已从 %s 卸载 ANR FG。"},
	"failed":            {"Couldn't complete: %s", "无法完成: %s"},
	"warn.nrrecord":     {"nr-record was in use. To restore it, run the nr-record installer again.", "之前在使用 nr-record。要恢复它，请重新运行 nr-record 安装程序。"},
	"probe.failed":      {"Couldn't read version information: %s", "无法读取版本信息: %s"},

	// Anti-cheat.
	"anticheat.short":    {"anti-cheat", "反作弊"},
	"anticheat.found":    {"Anti-cheat detected: %s. It might treat OptiScaler and ANR FG as cheats and ban your account. Install at your own risk.", "检测到反作弊: %s。它可能把 OptiScaler 和 ANR FG 视为作弊程序并封禁账号。安装风险由你自行承担。"},
	"anticheat.title":    {"Install in a game with anti-cheat?", "要在带有反作弊的游戏中安装吗?"},
	"anticheat.detail":   {"%s protects %s. OptiScaler and ANR FG hook into the game's rendering, so anti-cheat might treat them as cheats and ban your account, especially in online modes. Install only if you accept this risk.", "%s 正在保护 %s。OptiScaler 和 ANR FG 会接入游戏渲染，反作弊可能将它们视为作弊程序并封禁账号，在线模式下尤其如此。只有在你愿意承担这个风险时才安装。"},
	"anticheat.cancel":   {"Cancel", "取消"},
	"anticheat.accept":   {"Install anyway", "仍然安装"},
	"anticheat.accepted": {"Installing in spite of %s. You accepted the risk.", "检测到 %s，你已确认自行承担风险，继续安装。"},
}
