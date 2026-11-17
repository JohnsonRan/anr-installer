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

目前还没有发布可直接下载的程序，需要用 Go 1.27 或更新版本从源码构建：

```
git clone https://github.com/JohnsonRan/anr-installer.git
cd anr-installer
go tool mygo build -skip-installer
```

程序位于 `build/windows-amd64/ANR FG Installer.exe`。其他命令见《[开发说明](docs/development.md)》（英文）。

## 使用

1. 关闭游戏。
2. 打开 ANR FG 安装器，在左侧选择游戏。“支持”列出带有 DLSS 帧生成的游戏；Steam 以外的游戏可以用
   “+ 添加游戏”按文件夹添加。
3. 选择游戏横幅上的“安装”。如果已经装过 ANR FG，按钮会根据“安装”卡片里选择的版本显示为
   “升级”“重新安装”或“降级”。
4. 进入游戏，打开帧生成（游戏提供多个选项时选 NVIDIA 的）。

“卸载”会还原安装程序备份的文件。在某一步会丢失你的配置文件之前，安装器会先把它复制到游戏 `.exe` 旁边，
命名为 `<文件名>.bak-<时间>`：full 版替换 `OptiScaler.ini` 之前备份它，卸载或降级删除 `nr_fg.ini` 之前备份它。

## 反作弊

OptiScaler 和 ANR FG 会接入游戏的渲染。反作弊可能把它们当作作弊程序并封禁账号，联机模式下尤其如此。
游戏带有反作弊（Easy Anti-Cheat、EA Javelin、BattlEye 等）时，安装器会标出来并在安装前询问。只有在你
接受这个风险时才安装，风险由你自行承担。

## 工作原理

《[工作原理](docs/how-it-works.md)》（英文）介绍安装器读取和修改哪些内容、如何在不下载的情况下识别已装版本，
以及它保存的文件。

## 许可证

[MIT](LICENSE)。安装器下载并安装的 ANR FG 和 OptiScaler 各自适用其本身的许可证。
