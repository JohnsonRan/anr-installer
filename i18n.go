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
	"title":         {"ANR FG Installer", "ANR FG 安装器"},
	"subtitle":      {"Apex Neural Rendering frame generation for your games", "为你的游戏安装 Apex Neural Rendering 帧生成"},
	"check.updates": {"Check for updates", "检查更新"},
	"latest":        {"Latest ANR FG %s", "最新 ANR FG %s"},
	"win.minimize":  {"Minimize", "最小化"},
	"win.maximize":  {"Maximize", "最大化"},
	"win.restore":   {"Restore", "还原"},
	"win.close":     {"Close", "关闭"},

	// The list of games.
	"search":          {"Search games", "搜索游戏"},
	"show":            {"Show", "显示"},
	"show.all":        {"All", "全部"},
	"show.supported":  {"Supported", "支持"},
	"show.installed":  {"Installed", "已安装"},
	"games.count":     {"%d games · %d with DLSS-G · %d with ANR FG", "%d 款游戏 · %d 款支持 DLSS-G · %d 款已安装 ANR FG"},
	"add.game":        {"+ Add game", "+ 添加游戏"},
	"add.game.tip":    {"Add a game outside Steam by its folder", "按文件夹添加 Steam 以外的游戏"},
	"refresh":         {"Rescan", "重新扫描"},
	"refresh.tip":     {"Look for Steam games again", "重新查找 Steam 游戏"},
	"scanning.steam":  {"Looking for your games…", "正在查找游戏…"},
	"scanning":        {"Scanning…", "正在扫描…"},
	"no.steam":        {"No Steam library found. Add a game by its folder.", "没有找到 Steam 游戏库，可以按文件夹添加游戏。"},
	"no.games":        {"No games here.", "这里没有游戏。"},
	"manual":          {"added", "手动添加"},
	"manual.long":     {"Added by hand", "手动添加的游戏"},
	"row.supported":   {"DLSS frame generation", "支持 DLSS 帧生成"},
	"row.unsupported": {"No DLSS frame generation found", "未检测到 DLSS 帧生成"},
	"unofficial":      {"Unofficial tool, not made or endorsed by Apex Compute", "非官方工具，与 Apex Compute 无关"},
	"unofficial.tip":  {"Problems with ANR FG itself go to apex-compute/ANR; problems with this installer go to JohnsonRan/anr-installer", "ANR FG 本身的问题请反馈到 apex-compute/ANR，安装器的问题请反馈到 JohnsonRan/anr-installer"},
	"empty.title":     {"No game yet", "还没有游戏"},
	"empty.body":      {"Steam games show up here by themselves. Games from other stores can be added by their folder.", "Steam 游戏会自动出现在这里，其他平台的游戏可以按文件夹添加。"},

	// A game.
	"rename.tip":  {"Double-click to rename", "双击改名"},
	"remove":      {"Remove from list", "从列表移除"},
	"remove.tip":  {"Takes the game off the list; its files, ANR FG included, stay as they are", "只从列表中移除，游戏文件(包括已安装的 ANR FG)不受影响"},
	"removed":     {"Removed %s (%s) from the list", "已从列表移除 %s (%s)"},
	"save.failed": {"Couldn't save the list of games: %s", "无法保存游戏列表: %s"},
	"busy":        {"Working…", "正在处理…"},

	"anr.full":             {"Full package · OptiScaler as %s", "full 版 · OptiScaler 以 %s 加载"},
	"anr.addon":            {"Add-on to your OptiScaler", "addon 版 · 加在你的 OptiScaler 上"},
	"anr.none":             {"Not installed", "未安装"},
	"dlss.title":           {"DLSS FRAME GENERATION", "DLSS 帧生成"},
	"dlss.found":           {"Found", "已检测到"},
	"dlss.root":            {"next to the game's files", "在游戏根目录"},
	"dlss.none":            {"Not found", "未检测到"},
	"dlss.none.detail":     {"ANR FG takes the place of DLSS-G", "ANR FG 需要接管游戏的 DLSS-G"},
	"opti.ours":            {"ANR FG's own", "ANR FG 自带"},
	"opti.yours":           {"Yours", "你自己安装的"},
	"opti.none":            {"None", "未安装"},
	"opti.fg":              {"Frame generation by %s", "帧生成交给 %s"},
	"opti.none.detail":     {"The full package brings one", "full 版会自带"},
	"inst.version.unk":     {"Unknown version", "未知版本"},
	"inst.version.probing": {"Identifying…", "正在识别…"},

	"dlssg.no":       {"No DLSS frame generation was found in this game: ANR FG may not work in it.", "这个游戏里没有检测到 DLSS 帧生成，ANR FG 可能无法工作。"},
	"note.ini.reset": {"The full package replaces OptiScaler.ini; the installer backs yours up next to the .exe first.", "full 版会替换 OptiScaler.ini，安装器会先把原文件备份到 .exe 旁边。"},
	"note.nrrecord":  {"nr-record is in use: installing points frame generation back to ANR FG, so run nr-record's installer again afterwards to keep recording.", "nr-record 正在使用: 安装后帧生成会切回 ANR FG，如需继续录制，请重新运行 nr-record 的安装程序。"},
	"note.proton":    {"Proton: set the game's launch options to the line the installer prints.", "Proton: 请把游戏的启动选项设为安装程序输出的那一行。"},
	"note.downgrade": {
		"Going to an older version uninstalls the installed one first (with %s's uninstaller), then installs %s. OptiScaler.ini and nr_fg.ini are backed up next to the .exe; settings in nr_fg.ini are not carried over.",
		"降级会先卸载当前版本(使用 %s 的卸载程序)，再安装 %s。OptiScaler.ini 和 nr_fg.ini 会先备份到 .exe 旁边；nr_fg.ini 中的自定义设置不会带到旧版本。",
	},

	// Installing.
	"install.title":  {"INSTALL", "安装"},
	"version":        {"Version", "版本"},
	"version.latest": {"%s (latest)", "%s (最新)"},
	"release.notes":  {"Release notes", "发布说明"},
	"latest.loading": {"Looking for releases…", "正在获取版本…"},
	"latest.error":   {"Couldn't reach GitHub: %s", "无法连接 GitHub: %s"},
	"retry":          {"Retry", "重试"},
	"advanced":       {"Advanced", "高级选项"},
	"package":        {"Package", "安装包"},
	"package.help":   {"full brings its own OptiScaler; addon joins an OptiScaler you installed.", "full 版自带 OptiScaler；addon 版加在你自己安装的 OptiScaler 上。"},
	"kind.auto":      {"Auto (%s)", "自动(%s)"},
	"source":         {"Source", "来源"},
	"source.github0": {"GitHub", "GitHub"},
	"source.local":   {"Local zip", "本地 zip"},
	"choose.zip":     {"Choose zip…", "选择 zip…"},
	"zip.none":       {"Choose a package zip under Advanced", "请在“高级选项”中选择安装包 zip"},
	"zip.info":       {"%s · %s, %s package", "%s · %s，%s 版"},
	"pick.zip.title": {"ANR FG package", "ANR FG 安装包"},

	"install":   {"Install", "安装"},
	"install.v": {"Install %s", "安装 %s"},
	"upgrade":   {"Upgrade to %s", "升级到 %s"},
	"reinstall": {"Reinstall", "重新安装"},
	"downgrade": {"Downgrade to %s", "降级到 %s"},
	"switch":    {"Switch to %s", "切换到 %s"},
	"uninstall": {"Uninstall", "卸载"},

	"why.user-optiscaler": {"This game has an OptiScaler you installed, which the full package refuses: choose addon under Advanced.", "这个游戏已有你自己安装的 OptiScaler，full 版会拒绝安装: 请在“高级选项”中选择 addon。"},
	"why.no-optiscaler":   {"The add-on package needs OptiScaler.ini next to the .exe: choose full, or install OptiScaler first.", "addon 版需要 .exe 旁边已有 OptiScaler.ini: 请选择 full，或先安装 OptiScaler。"},
	"why.full-installed":  {"The full package is installed here: change it with the full package.", "这里装的是 full 版，请继续使用 full 版。"},
	"why.zip-kind":        {"The chosen zip is a %s package, not %s.", "所选 zip 是 %s 版，不是 %s 版。"},
	"why.no-zip":          {"Choose a package zip first.", "请先选择安装包 zip。"},
	"why.no-release":      {"The releases aren't known yet.", "还没有获取到版本信息。"},

	// Where it goes.
	"exe.folder":        {"FOLDER OF THE .EXE", ".EXE 所在文件夹"},
	"change":            {"Change…", "更改…"},
	"auto.exe":          {"Find automatically", "自动识别"},
	"auto.exe.tip":      {"Go back to the folder the installer found", "恢复为安装器自动找到的文件夹"},
	"open.folder":       {"Open folder", "打开文件夹"},
	"pick.folder.title": {"The game's folder", "游戏所在文件夹"},

	// Activity.
	"log":               {"Activity", "活动"},
	"log.empty":         {"Nothing yet.", "暂无记录。"},
	"clear":             {"Clear", "清空"},
	"downloading":       {"Downloading %s… %s", "正在下载 %s… %s"},
	"running.install":   {"Installing %s (%s)…", "正在安装 %s (%s 版)…"},
	"running.uninstall": {"Uninstalling with %s's uninstaller (%s)…", "正在使用 %s 的卸载程序卸载(%s 版)…"},
	"backup.file":       {"Backed up %s to %s", "已备份 %s 到 %s"},
	"done.install":      {"Done: ANR FG %s is installed in %s.", "完成: ANR FG %s 已安装到 %s。"},
	"done.uninstall":    {"Done: ANR FG is removed from %s.", "完成: 已从 %s 卸载 ANR FG。"},
	"failed":            {"Failed: %s", "失败: %s"},
	"warn.nrrecord":     {"nr-record was in use: run nr-record's installer again to restore it.", "之前启用了 nr-record: 如需恢复，请重新运行 nr-record 的安装程序。"},
	"probe.failed":      {"Couldn't read the release versions: %s", "无法读取各版本信息: %s"},

	// Anti-cheat.
	"anticheat.short":    {"anti-cheat", "反作弊"},
	"anticheat.found":    {"Anti-cheat found: %s. It may take OptiScaler and ANR FG for a cheat and ban the account. Installing is at your own risk.", "检测到反作弊: %s。反作弊可能把 OptiScaler 和 ANR FG 当作作弊程序并封禁账号，安装的风险由你自行承担。"},
	"anticheat.title":    {"This game uses anti-cheat", "这个游戏使用了反作弊"},
	"anticheat.detail":   {"%s protects %s. Anti-cheat may take OptiScaler and ANR FG, which hook into the game's rendering, for a cheat, and ban the account, in online modes above all. Install only if you accept that risk; you bear it yourself.", "%s 保护着 %s。OptiScaler 和 ANR FG 会接入游戏的渲染，反作弊可能把它们当作作弊程序并封禁账号，联机模式下尤其如此。只有在你接受这个风险时才安装，风险由你自行承担。"},
	"anticheat.cancel":   {"Cancel", "取消"},
	"anticheat.accept":   {"Install at my own risk", "我自行承担风险，继续安装"},
	"anticheat.accepted": {"Installing despite %s: the user accepted the risk.", "检测到 %s，你已确认自行承担风险，继续安装。"},
}
