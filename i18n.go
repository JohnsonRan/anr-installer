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
	"note.ini.reset": {"Installing or upgrading ANR FG replaces OptiScaler.ini with a fresh one. Your current OptiScaler.ini is backed up to the game's .exe folder first.", "安装或升级 ANR FG 会把 OptiScaler.ini 换成全新的文件。你当前的 OptiScaler.ini 会先备份到游戏 .exe 所在的文件夹。"},
	"note.nrrecord":  {"Recording is on for this game. Installing or upgrading ANR FG replaces OptiScaler.ini, which turns recording off. To record again afterward, install nr-record again under “Record and compare” below.", "这款游戏已开启录制功能。安装或升级 ANR FG 会替换 OptiScaler.ini，录制功能随之关闭。之后要继续录制，请在下方的“录制和对比”中重新安装 nr-record。"},
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
	"warn.nrrecord":     {"Recording is now off for this game. To record again, install nr-record again under “Record and compare”.", "这款游戏的录制功能已关闭。要继续录制，请在“录制和对比”中重新安装 nr-record。"},
	"probe.failed":      {"Couldn't read version information: %s", "无法读取版本信息: %s"},

	// Anti-cheat.
	"anticheat.short":    {"anti-cheat", "反作弊"},
	"anticheat.found":    {"Anti-cheat detected: %s. It might treat OptiScaler and ANR FG as cheats and ban your account. Install at your own risk.", "检测到反作弊: %s。它可能把 OptiScaler 和 ANR FG 视为作弊程序并封禁账号。安装风险由你自行承担。"},
	"anticheat.title":    {"Install in a game with anti-cheat?", "要在带有反作弊的游戏中安装吗?"},
	"anticheat.detail":   {"%s protects %s. OptiScaler and ANR FG hook into the game's rendering, so anti-cheat might treat them as cheats and ban your account, especially in online modes. Install only if you accept this risk.", "%s 正在保护 %s。OptiScaler 和 ANR FG 会接入游戏渲染，反作弊可能将它们视为作弊程序并封禁账号，在线模式下尤其如此。只有在你愿意承担这个风险时才安装。"},
	"anticheat.cancel":   {"Cancel", "取消"},
	"anticheat.accept":   {"Install anyway", "仍然安装"},
	"anticheat.accepted": {"Installing in spite of %s. You accepted the risk.", "检测到 %s，你已确认自行承担风险，继续安装。"},

	// Record and compare.
	"rec.title":             {"Record and compare", "录制和对比"},
	"rec.nrrecord":          {"Records the game's frame generation inputs", "录制游戏的帧生成输入"},
	"rec.nrrecord.needs":    {"Needs ANR FG's full package", "需要 ANR FG 的 full 版"},
	"rec.nrrecord.on":       {"With DLSS frame generation on in the game, F6 starts and stops a recording, and F9 records 8 frames. A beep marks each start and stop. Recordings are large: about 26 MB a frame at 3840 × 2160, tens of GB a minute.", "在游戏中开启 DLSS 帧生成后，按 F6 开始或停止录制，按 F9 录制 8 帧，开始和停止时各有一声提示音。录制内容很大: 3840 × 2160 下每帧约 26 MB，一分钟可达数十 GB。"},
	"rec.nrrecord.off":      {"OptiScaler.ini no longer sends frame generation to nr-record, so F6 records nothing. To record, install nr-record again.", "OptiScaler.ini 不再把帧生成交给 nr-record，按 F6 不会录制。要录制，请重新安装 nr-record。"},
	"rec.installed":         {"Installed", "已安装"},
	"rec.unused":            {"Not in use", "未启用"},
	"rec.install":           {"Install nr-record", "安装 nr-record"},
	"rec.reinstall":         {"Install nr-record again", "重新安装 nr-record"},
	"rec.uninstall":         {"Uninstall nr-record", "卸载 nr-record"},
	"rec.ffmpeg":            {"FFmpeg and FFprobe", "FFmpeg 和 FFprobe"},
	"rec.ffmpeg.detail":     {"For making and playing the videos", "用于生成和播放视频"},
	"rec.ffmpeg.noscoop":    {"FFmpeg and FFprobe weren't found, and neither was Scoop, which installs them. Install Scoop, then check again.", "未检测到 FFmpeg 和 FFprobe，也未检测到用来安装它们的 Scoop。请先安装 Scoop，然后重新检测。"},
	"rec.found":             {"Found", "已检测到"},
	"rec.missing":           {"Not found", "未检测到"},
	"rec.scoop.install":     {"Install with Scoop", "通过 Scoop 安装"},
	"rec.scoop.get":         {"Get Scoop", "获取 Scoop"},
	"rec.check":             {"Check again", "重新检测"},
	"rec.recordings":        {"Recordings", "录制内容"},
	"rec.recordings.none":   {"None yet", "还没有录制内容"},
	"rec.recordings.count":  {"%d of this game · %d in all", "这款游戏 %d 段 · 共 %d 段"},
	"rec.refresh":           {"Refresh", "刷新"},
	"rec.refresh.tip":       {"Look again for recordings, FFmpeg and Scoop", "重新检测录制内容、FFmpeg 和 Scoop"},
	"rec.browse":            {"Open viewer", "打开查看器"},
	"rec.browse.tip":        {"Browse the recordings in nr-record's viewer", "在 nr-record 的查看器中浏览录制内容"},
	"rec.item":              {"%d × %d · %d frames · %s", "%d × %d · %d 帧 · %s"},
	"rec.item.dropped":      {" · %d dropped", " · 丢弃 %d 帧"},
	"rec.item.upscaler":     {"Upscaler inputs", "超分辨率输入"},
	"rec.item.cli":          {"Command line only", "仅支持命令行回放"},
	"rec.compare":           {"Compare with DLSS-G", "与 DLSS-G 对比"},
	"rec.compare.open":      {"Open comparison", "打开对比"},
	"rec.compare.again":     {"Replay again", "重新回放"},
	"rec.why.tools":         {"Comparing a recording needs FFmpeg and FFprobe.", "对比录制内容需要 FFmpeg 和 FFprobe。"},
	"rec.why.anr":           {"Comparing a recording needs ANR FG installed in this game.", "对比录制内容需要在这款游戏中安装 ANR FG。"},
	"rec.why.dlssg":         {"Comparing a recording needs the game's own nvngx_dlssg.dll, which wasn't found.", "对比录制内容需要游戏自带的 nvngx_dlssg.dll，但未检测到。"},
	"rec.why.no-package":    {"Couldn't find an nr-record package: the releases couldn't be read.", "找不到 nr-record 安装包: 无法读取版本信息。"},
	"rec.why.no-asset":      {"The releases carry no nr-record package.", "已发布的版本中没有 nr-record 安装包。"},
	"rec.running.install":   {"Installing nr-record…", "正在安装 nr-record…"},
	"rec.running.uninstall": {"Uninstalling nr-record…", "正在卸载 nr-record…"},
	"rec.done.install":      {"Done: nr-record is installed in %s.", "完成: nr-record 已安装到 %s。"},
	"rec.done.uninstall":    {"Done: nr-record is removed from %s. The recordings stay.", "完成: 已从 %s 卸载 nr-record，录制内容不受影响。"},
	"rec.scoop.running":     {"Installing FFmpeg with Scoop…", "正在通过 Scoop 安装 FFmpeg…"},
	"rec.scoop.done":        {"Done: FFmpeg and FFprobe are installed.", "完成: FFmpeg 和 FFprobe 已安装。"},
	"rec.scoop.missing":     {"Scoop finished, but FFmpeg and FFprobe still weren't found.", "Scoop 已运行完毕，但仍未检测到 FFmpeg 和 FFprobe。"},
	"rec.replaying":         {"Replaying through %s…", "正在通过 %s 回放…"},
	"rec.hint.dlssg":        {"Replaying through DLSS-G needs an NVIDIA RTX 40 or 50 series GPU.", "通过 DLSS-G 回放需要 NVIDIA RTX 40 或 50 系列显卡。"},
	"rec.done.compare":      {"Done: the comparison is in %s.", "完成: 对比结果已保存到 %s。"},

	// The comparison page.
	"cmp.back":          {"Back", "返回"},
	"cmp.title":         {"%s and %s", "%s 和 %s"},
	"cmp.viewer":        {"Open in nr-record's viewer", "在 nr-record 查看器中打开"},
	"cmp.error":         {"Couldn't play the videos: %s. They are HEVC videos: if Windows has no HEVC decoder, install “HEVC Video Extensions” from the Microsoft Store, or open the comparison in nr-record's viewer.", "无法播放视频: %s。视频采用 HEVC 编码: 如果 Windows 没有 HEVC 解码器，请从 Microsoft Store 安装“HEVC 视频扩展”，或者在 nr-record 查看器中打开对比。"},
	"cmp.mode":          {"Show", "显示"},
	"cmp.video":         {"Video", "视频"},
	"cmp.stills":        {"Stills", "静帧"},
	"cmp.layout":        {"View", "视图"},
	"cmp.divider":       {"Divider", "分隔线"},
	"cmp.divider.at":    {"Divider position", "分隔线位置"},
	"cmp.side":          {"Side by side", "并排"},
	"cmp.play":          {"Play", "播放"},
	"cmp.pause":         {"Pause", "暂停"},
	"cmp.prev":          {"Previous frame", "上一帧"},
	"cmp.next":          {"Next frame", "下一帧"},
	"cmp.time":          {"Time", "时间"},
	"cmp.rate":          {"Speed", "速度"},
	"cmp.note":          {"The two videos play on one clock but aren't locked frame by frame. To compare a frame, pause, or export its stills. Space plays and pauses, and the left and right arrow keys step by a frame.", "两段视频按同一时钟播放，但不是逐帧锁定。要对比某一帧，请先暂停，或导出这一帧的静帧。按空格键播放或暂停，按向左键或向右键逐帧移动。"},
	"cmp.still.none":    {"Stills are the frames themselves, as lossless PNG images, where the videos are compressed. Pause the video at a frame, then export it: the frames around it are replayed again.", "静帧是无损 PNG 格式的原始帧，而视频经过压缩。在视频中暂停到某一帧，然后导出，安装器会重新回放这一帧前后的画面。"},
	"cmp.still.error":   {"Couldn't show the stills: %s. Export them again.", "无法显示静帧: %s。请重新导出。"},
	"cmp.still.make":    {"Export the stills of frame %d", "导出第 %d 帧的静帧"},
	"cmp.still.loading": {"Loading the still…", "正在加载静帧…"},
	"cmp.still.prev":    {"Previous still", "上一张"},
	"cmp.still.next":    {"Next still", "下一张"},
	"cmp.still.name":    {"%d of %d · %s", "第 %d 张，共 %d 张 · %s"},
	"cmp.still.running": {"Exporting stills through %s…", "正在通过 %s 导出静帧…"},
	"cmp.still.done":    {"Done: %d stills are in %s.", "完成: %d 张静帧已保存到 %s。"},
}
