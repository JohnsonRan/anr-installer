# ANR FG Installer

[English](README.md) | [简体中文](README.zh-CN.md) | 日本語

ANR FG Installer は、ゲームに [ANR FG](https://github.com/apex-compute/ANR) (Apex Neural Rendering のフレーム生成) を
インストール、アップグレード、ダウングレード、アンインストールするツールです。Steam のゲームを見つけて DLSS
フレーム生成に対応したゲームを示し、適切なパッケージを選んで、パッケージに付属のインストーラーを実行します。

> **非公式のツールです。** ANR FG の開発元である Apex Compute が作成、承認、サポートしているものではありません。
> インストールされるのは [apex-compute/ANR のリリース](https://github.com/apex-compute/ANR/releases) にある
> ANR FG のパッケージで、内容は変更されません。ANR FG 自体の問題は
> [apex-compute/ANR](https://github.com/apex-compute/ANR/issues) に、このインストーラーの問題は
> [JohnsonRan/anr-installer](https://github.com/JohnsonRan/anr-installer/issues) にお知らせください。

## 動作環境

- Windows 10 または 11 (x64)
- Microsoft Visual C++ 2015-2022 Redistributable (x64)。OptiScaler に必要です。
- ANR FG に対応した GPU: NVIDIA GeForce RTX 20 以降または GTX 16、AMD Radeon RX 5000 以降、Intel Arc A と B

Linux と Steam Deck (Proton) に対応するコードはありますが、まだテストされていません。

## 入手方法

バイナリはまだリリースされていません。Go 1.27 以降を使って、ソースからビルドします:

```
git clone https://github.com/JohnsonRan/anr-installer.git
cd anr-installer
go tool mygo build -skip-installer
```

アプリは `build/windows-amd64/ANR FG Installer.exe` に作成されます。そのほかのコマンドについては、
[開発者向けドキュメント](docs/development.md) (英語) をご覧ください。

## 使い方

アプリの画面は英語で表示されます。以下では、画面の表記に日本語訳を添えています。

1. ゲームを終了します。
2. ANR FG Installer を起動して、左側でゲームを選びます。[Supported] (対応) には、DLSS フレーム生成に対応した
   ゲームが表示されます。Steam 以外のゲームは、[+ Add game] (ゲームを追加) からフォルダーを指定して追加します。
3. ゲームのバナーにある [Install] (インストール) をクリックします。ANR FG がすでにインストールされている場合、
   ボタンは [Install] カードで選んだバージョンに応じて [Upgrade] (アップグレード)、[Reinstall] (再インストール)、
   [Downgrade] (ダウングレード) になります。
4. ゲームを起動して、フレーム生成をオンにします。複数の選択肢がある場合は、NVIDIA のものを選びます。

[Uninstall] (アンインストール) を実行すると、パッケージのインストーラーがバックアップしたファイルが元に戻されます。
設定ファイルが失われる操作の前には、そのファイルがゲームの `.exe` と同じフォルダーに `<名前>.bak-<日時>` として
コピーされます。full パッケージが `OptiScaler.ini` を置き換える前と、アンインストールやダウングレードで
`nr_fg.ini` が削除される前です。

## アンチチート

OptiScaler と ANR FG は、ゲームの描画処理に組み込まれます。アンチチートがこれらを不正なツールと判断して、
アカウントを停止する場合があります。特にオンライン モードではご注意ください。ゲームにアンチチート
(Easy Anti-Cheat、EA Javelin、BattlEye など) が含まれている場合、そのゲームには印が付き、インストールの前に
確認のダイアログ ボックスが表示されます。このリスクを受け入れられる場合にのみ、インストールしてください。
リスクはご自身で負うことになります。

## 動作のしくみ

[動作のしくみ](docs/how-it-works.md) (英語) では、インストーラーが読み取る内容と変更する内容、パッケージを
ダウンロードせずにインストール済みのバージョンを判別する方法、保存されるファイルについて説明しています。

## ライセンス

[MIT](LICENSE)。インストーラーがダウンロードしてインストールする ANR FG と OptiScaler には、それぞれのライセンスが
適用されます。
