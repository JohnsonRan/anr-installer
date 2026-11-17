# ANR FG 安装器

[English](README.md) | 简体中文 | [日本語](README.ja.md)

ANR FG 安装器用来在你的游戏中安装、升级、降级和卸载 [ANR FG](https://github.com/apex-compute/ANR)
（Apex Neural Rendering 帧生成）。它会找到你的 Steam 游戏，标出哪些支持 DLSS 帧生成，选择合适的安装包，
并调用安装包自带的安装程序。

> **非官方工具。** 本项目是社区工具，并非由 ANR FG 的开发者 Apex Compute 制作、认可或提供支持。它安装的是
> [apex-compute/ANR 发布页](https://github.com/apex-compute/ANR/releases)上未经修改的官方安装包。ANR FG 本身的问题请反馈到
> [apex-compute/ANR](https://github.com/apex-compute/ANR/issues)，安装器的问题请反馈到
> [这里](https://github.com/JohnsonRan/anr-installer/issues)。

## 系统要求

- Windows 10 或 11 (x64)
- Microsoft Visual C++ 2015-2022 Redistributable (x64)，OptiScaler 需要它
- ANR FG 支持的显卡：NVIDIA GeForce RTX 20 及更新或 GTX 16 系列；AMD Radeon RX 5000 及更新；Intel Arc A 和
  B 系列

Linux 和 Steam Deck (Proton) 的支持已经写好，但尚未测试。

## 获取

从[最新版本](https://github.com/JohnsonRan/anr-installer/releases/latest)下载
`anr-fg-installer-setup-<版本>-amd64.exe` 并运行。它只为当前 Windows 账户安装，位置是
`%LOCALAPPDATA%\Programs\ANR FG Installer`，不需要管理员权限。也可以在 PowerShell 中安装：

```
irm https://github.com/JohnsonRan/anr-installer/releases/latest/download/install.ps1 | iex
```

便携版 `anr-fg-installer-<版本>-windows-amd64-portable.zip` 无需安装：把其中的 `ANR FG Installer` 文件夹解压到可写入的位置，
运行 `ANR FG Installer.exe` 即可。它会把设置和下载的文件保存在这个文件夹里。

程序和安装程序都没有代码签名，首次运行时 Windows SmartScreen 可能会发出警告：选择“更多信息”，再选择“仍要运行”。

程序会自动更新：启动时和之后每天检查一次新版本，选择标题栏上的版本号按钮可以立即检查。它只安装用本项目密钥签名的更新。
如果想从源码构建，请参阅《[开发说明](docs/development.md)》（英文）。

## 使用

1. 关闭游戏。
2. 打开 ANR FG 安装器，在左侧选择游戏。“受支持”列出带有 DLSS 帧生成的游戏。要添加其他平台的游戏，请选择
   “添加游戏”，然后选择游戏文件夹。
3. 选择游戏页面顶部的“安装”。如果已经装过 ANR FG，按钮会根据旁边箭头菜单中选择的版本显示为“升级”“重新安装”
   或“降级”。
4. 进入游戏，打开帧生成（游戏提供多个选项时选 NVIDIA 的）。

“卸载”会还原安装程序备份的文件。在某一步会丢失你的配置文件之前，安装器会先把它复制到游戏 `.exe` 所在文件夹，
命名为 `<文件名>.bak-<时间>`：full 版替换 `OptiScaler.ini` 之前备份它，卸载或降级删除 `nr_fg.ini` 之前备份它。

ANR FG 发布新版本、而你有游戏装的是旧版本时，安装器会显示一次发行说明。选择其中的“全部更新”，会按顺序升级这些
游戏，各自保持原来的安装包类型。带反作弊的游戏不会一并更新，请在各自的页面中单独升级。列表会标出可升级的游戏，
列表上方的提示条可以只显示这些游戏。

## 反作弊

OptiScaler 和 ANR FG 会接入游戏的渲染。反作弊可能把它们当作作弊程序并封禁账号，联机模式下尤其如此。
游戏带有反作弊（Easy Anti-Cheat、EA Javelin、BattlEye 等）时，安装器会标出来并在安装前询问。只有在你
接受这个风险时才安装，风险由你自行承担。

## 工作原理

《[工作原理](docs/how-it-works.md)》（英文）介绍安装器读取和修改哪些内容、如何在不下载的情况下识别已装版本，
以及它保存的文件。

## 许可证

[MIT](LICENSE)。安装器下载并安装的 ANR FG 和 OptiScaler 各自适用其本身的许可证。
