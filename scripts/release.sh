#!/bin/sh
# Builds the version in mygo.json and uploads it to a draft GitHub release: the installer, install.ps1, the
# signed update files, and the portable zip. Publishing the draft stays a separate step (docs/development.md).
#   MYGO_UPDATER_PRIVATE_KEY="$(cat path/to/mygo-update.key)" scripts/release.sh
# Exit 1 when a step fails, 2 when the release cannot be made.
set -eu
cd "$(dirname "$0")/.."

[ -n "${MYGO_UPDATER_PRIVATE_KEY:-}" ] || { echo "release.sh: set MYGO_UPDATER_PRIVATE_KEY to the update signing key" >&2; exit 2; }
command -v gh >/dev/null || { echo "release.sh: the GitHub CLI (gh) is needed" >&2; exit 2; }
version=$(sed -n 's/^ *"version": *"\([^"]*\)".*/\1/p' mygo.json)
slug=$(sed -n 's/^ *"slug": *"\([^"]*\)".*/\1/p' mygo.json)
name=$(sed -n 's/^ *"name": *"\([^"]*\)".*/\1/p' mygo.json)
repo=$(sed -n 's/^ *"github": *"\([^"]*\)".*/\1/p' mygo.json)
[ -n "$version" ] && [ -n "$slug" ] && [ -n "$name" ] && [ -n "$repo" ] || { echo "release.sh: mygo.json lacks version, slug, name or updates.github" >&2; exit 2; }
grep -q "^## $version\$" CHANGELOG.md || { echo "release.sh: CHANGELOG.md has no \"## $version\" section" >&2; exit 2; }

rm -rf build
go tool mygo build -upload || exit 1

# The portable version: the executable, the license, and the "portable" marker that makes it keep its files
# in its own folder (internal/paths).
stage=build/portable
zip="$slug-$version-windows-amd64-portable.zip"
mkdir -p "$stage/$name"
cp "build/windows-amd64/$name.exe" LICENSE "$stage/$name/"
: > "$stage/$name/portable"
if command -v powershell >/dev/null; then
	(cd "$stage" && powershell -NoProfile -Command "Compress-Archive -Path '$name' -DestinationPath '../$zip' -Force") || exit 1
else
	(cd "$stage" && zip -qr "../$zip" "$name") || exit 1
fi
# SHA256SUMS, which mygo build uploaded, gets the portable zip too, so that every download can be checked. -t
# writes the two-space form mygo writes, where Git Bash would mark the file binary with "*".
(cd build && sha256sum -t "$zip" >> SHA256SUMS) || exit 1
gh release upload "v$version" "build/$zip" build/SHA256SUMS --repo "$repo" --clobber || exit 1
echo "Uploaded build/$zip to the draft release v$version. Publish it with:"
echo "  gh release edit v$version --repo $repo --draft=false"
