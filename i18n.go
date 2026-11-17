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
	"title":               {"ANR FG Installer", "ANR FG 安装器"},
	"subtitle":            {"Apex Neural Rendering frame generation, for Steam games", "Apex Neural Rendering 帧生成，适用于 Steam 游戏"},
	"latest":              {"Latest: %s", "最新版本：%s"},
	"latest.loading":      {"Checking for the latest release…", "正在检查最新版本…"},
	"latest.error":        {"Couldn't reach GitHub: %s", "无法连接 GitHub：%s"},
	"refresh":             {"Rescan", "重新扫描"},
	"search":              {"Search games", "搜索游戏"},
	"only.supported":      {"Only games with DLSS frame generation", "只显示支持 DLSS 帧生成的游戏"},
	"add.folder":          {"Add a game folder…", "添加游戏文件夹…"},
	"scanning.steam":      {"Looking for Steam libraries…", "正在查找 Steam 游戏库…"},
	"no.steam":            {"No Steam library found. Add a game folder instead.", "没有找到 Steam 游戏库，可以手动添加游戏文件夹。"},
	"no.games":            {"No games match.", "没有符合条件的游戏。"},
	"games.count":         {"%d games, %d with DLSS frame generation", "共 %d 个游戏，其中 %d 个支持 DLSS 帧生成"},
	"pick.game":           {"Choose a game on the left.", "请在左侧选择一个游戏。"},
	"scanning":            {"Scanning…", "正在扫描…"},
	"badge.supported":     {"DLSS-G", "DLSS-G"},
	"badge.full":          {"ANR full", "ANR full"},
	"badge.addon":         {"ANR add-on", "ANR addon"},
	"badge.opti":          {"OptiScaler", "OptiScaler"},
	"badge.manual":        {"Added", "手动添加"},
	"exe.folder":          {"Folder of the .exe", ".exe 所在文件夹"},
	"change":              {"Change…", "更改…"},
	"open.folder":         {"Open", "打开"},
	"status":              {"Status", "状态"},
	"dlssg.yes":           {"DLSS frame generation found (%s)", "检测到 DLSS 帧生成（%s）"},
	"dlssg.no":            {"No DLSS frame generation found: ANR FG may not work in this game", "未检测到 DLSS 帧生成：ANR FG 在这个游戏里可能无法工作"},
	"inst.none":           {"ANR FG is not installed", "未安装 ANR FG"},
	"inst.full":           {"ANR FG installed (full package, OptiScaler as %s)", "已安装 ANR FG（full 版，OptiScaler 以 %s 加载）"},
	"inst.addon":          {"ANR FG installed (add-on to your OptiScaler)", "已安装 ANR FG（addon 版，加在你的 OptiScaler 上）"},
	"inst.version":        {"Installed version: %s", "已安装版本：%s"},
	"inst.version.unk":    {"unknown (check after a download)", "未知（下载安装包后可比对）"},
	"opti.user":           {"OptiScaler you installed is here: use the add-on package", "这里有你自己装的 OptiScaler：请用 addon 版"},
	"nrrecord.on":         {"nr-record is active (FGNvngxReplacement=nr-record)", "nr-record 正在使用（FGNvngxReplacement=nr-record）"},
	"package":             {"Package", "安装包类型"},
	"kind.auto":           {"Auto (%s)", "自动（%s）"},
	"source":              {"Source", "安装包来源"},
	"source.github0":      {"GitHub", "GitHub"},
	"source.local":        {"Local zip", "本地 zip"},
	"choose.zip":          {"Choose zip…", "选择 zip…"},
	"zip.none":            {"No zip chosen", "未选择 zip"},
	"zip.info":            {"%s, %s package", "%s，%s 版"},
	"install":             {"Install", "安装"},
	"upgrade":             {"Upgrade to %s", "升级到 %s"},
	"reinstall":           {"Reinstall", "重新安装"},
	"uninstall":           {"Uninstall", "卸载"},
	"why.user-optiscaler": {"This game has an OptiScaler you installed; the full package would refuse it. Use add-on.", "这个游戏已有你自己装的 OptiScaler，full 版会拒绝安装，请选 addon。"},
	"why.no-optiscaler":   {"The add-on package needs OptiScaler.ini next to the .exe. Use full, or install OptiScaler first.", "addon 版需要 .exe 旁边已有 OptiScaler.ini。请选 full，或先安装 OptiScaler。"},
	"why.full-installed":  {"The full package is installed here; upgrade it with the full package.", "这里装的是 full 版，请用 full 版升级。"},
	"why.zip-kind":        {"The chosen zip is a %s package, not %s.", "所选 zip 是 %s 版，不是 %s 版。"},
	"why.no-zip":          {"Choose a package zip first.", "请先选择安装包 zip。"},
	"why.no-release":      {"The latest release isn't known yet.", "还没有获取到最新版本。"},
	"note.ini.reset":      {"Upgrading the full package replaces OptiScaler.ini with the package's; the installer backs yours up first.", "full 版升级会用包里的 OptiScaler.ini 覆盖现有文件；安装器会先帮你备份。"},
	"note.nrrecord":       {"Installing sets FGNvngxReplacement back to ANR FG: run nr-record's installer again afterwards to keep recording.", "安装会把 FGNvngxReplacement 改回 ANR FG：之后如需继续录制，请重新运行 nr-record 的安装程序。"},
	"note.close":          {"Close the game before installing.", "安装前请先关闭游戏。"},
	"note.proton":         {"Proton: set the game's launch options to the line the installer prints.", "Proton：请把游戏的启动选项设为安装程序输出的那一行。"},
	"log":                 {"Log", "日志"},
	"clear":               {"Clear", "清空"},
	"downloading":         {"Downloading %s… %s", "正在下载 %s… %s"},
	"done.install":        {"Done: ANR FG %s is installed in %s.", "完成：ANR FG %s 已安装到 %s。"},
	"done.uninstall":      {"Done: ANR FG is removed from %s.", "完成：已从 %s 卸载 ANR FG。"},
	"failed":              {"Failed: %s", "失败：%s"},
	"warn.nrrecord":       {"nr-record was active before: run nr-record's install again to restore it.", "之前启用了 nr-record：如需恢复，请重新运行 nr-record 的安装程序。"},
	"pick.folder.title":   {"Folder of the game's .exe", "游戏 .exe 所在文件夹"},
	"pick.zip.title":      {"ANR FG package", "ANR FG 安装包"},
	"busy":                {"Working…", "正在处理…"},
}

func init() {
	messages["win.minimize"] = [2]string{"Minimize", "最小化"}
	messages["win.maximize"] = [2]string{"Maximize", "最大化"}
	messages["win.restore"] = [2]string{"Restore", "还原"}
	messages["win.close"] = [2]string{"Close", "关闭"}
}

func init() {
	for k, v := range map[string][2]string{
		"version":              {"Version", "版本"},
		"version.latest":       {"%s (latest)", "%s（最新）"},
		"release.notes":        {"Release notes", "发布说明"},
		"install.v":            {"Install %s", "安装 %s"},
		"downgrade":            {"Downgrade to %s", "降级到 %s"},
		"switch":               {"Switch to %s", "切换到 %s"},
		"inst.version.probing": {"identifying…", "正在识别…"},
		"probe.failed":         {"Couldn't read the release versions: %s", "无法读取各版本信息：%s"},
		"note.downgrade": {
			"Going to an older version uninstalls the installed one first (with %s's uninstaller), then installs %s. OptiScaler.ini and nr_fg.ini are backed up next to the .exe; settings in nr_fg.ini are not carried over.",
			"降级会先卸载当前版本（使用 %s 的卸载程序），再安装 %s。OptiScaler.ini 和 nr_fg.ini 会先备份到 .exe 旁边；nr_fg.ini 中的自定义设置不会带到旧版本。",
		},
		"backup.file":       {"Backed up %s to %s", "已备份 %s 到 %s"},
		"running.install":   {"Installing %s (%s)…", "正在安装 %s（%s 版）…"},
		"running.uninstall": {"Uninstalling with %s's uninstaller (%s)…", "正在使用 %s 的卸载程序卸载（%s 版）…"},
	} {
		messages[k] = v
	}
}

func init() {
	for k, v := range map[string][2]string{
		"rename.tip":        {"Double-click to rename", "双击改名"},
		"remove":            {"Remove from list", "从列表移除"},
		"remove.tip":        {"Takes the game off the list; its files, ANR FG included, stay as they are", "只从列表中移除，游戏文件（包括已装的 ANR FG）不受影响"},
		"removed":           {"Removed %s (%s) from the list", "已从列表移除 %s（%s）"},
		"save.failed":       {"Couldn't save the list of games: %s", "无法保存游戏列表：%s"},
		"auto.exe":          {"Find automatically", "自动识别"},
		"auto.exe.tip":      {"Go back to the folder the installer found", "恢复为安装器自动找到的文件夹"},
		"pick.folder.title": {"The game's folder", "游戏所在文件夹"},
	} {
		messages[k] = v
	}
}
