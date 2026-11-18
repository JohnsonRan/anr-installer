#!/bin/sh
# The repository's checks, by tier (REQUIREMENTS.md, "Commands"):
#   scripts/check.sh --fast     pre-commit: formatting and vet
#   scripts/check.sh            merge gate: the above, the tests and the requirements check
#   scripts/check.sh --release  release: the above, the real-package tests and the build
# Exit 1 when a check fails, 2 when one cannot run.
set -eu
cd "$(dirname "$0")/.."

tier=${1:-merge}
case $tier in
--fast) tier=fast ;;
--release) tier=release ;;
merge) ;;
*) echo "usage: scripts/check.sh [--fast|--release]" >&2; exit 2 ;;
esac

# Python is "py" on Windows, "python3" elsewhere.
python=
for p in python3 py python; do
	if "$p" -c 'import sys; sys.exit(sys.version_info < (3, 9))' >/dev/null 2>&1; then python=$p; break; fi
done

unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
	echo "gofmt: not formatted:" >&2
	echo "$unformatted" >&2
	exit 1
fi
# The interface files the scale checks look at: tracked Go files, not tests and not the scales themselves.
go_files() { git ls-files --cached --others --exclude-standard -- "*.go" ":!:*_test.go" ":!:$1"; }

# The type scale (gui/AGENTS.md, "Layout, type and loading"): text sizes and weights are set in
# gui/style/type.go only.
literals=$(grep -n -E '\.(FontSize|FontWeight)\(|\.Bold\(\)' $(go_files gui/style/type.go)) && found=0 || found=$?
[ "$found" -le 1 ] || { echo "type scale: grep failed" >&2; exit 2; }
if [ -n "$literals" ]; then
	echo "type scale: set text size and weight through a role in gui/style/type.go:" >&2
	echo "$literals" >&2
	exit 1
fi
# The spacing scale: paddings, gaps and margins take a name from gui/style/spacing.go, never a number
# but 0.
literals=$(grep -n -E '\.(Padding|Gap|Margin)[A-Za-z]*\(([^)]*[ ,+*/-])?[0-9]*[1-9]' $(go_files gui/style/spacing.go)) && found=0 || found=$?
[ "$found" -le 1 ] || { echo "spacing scale: grep failed" >&2; exit 2; }
if [ -n "$literals" ]; then
	echo "spacing scale: use a name from gui/style/spacing.go:" >&2
	echo "$literals" >&2
	exit 1
fi
go vet ./... || exit 1
go tool mygo vet . || exit 1
[ "$tier" = fast ] && exit 0

# The merge gate leaves the real-package tests to the release tier, whatever the environment holds.
env -u ANR_PKG_DIR -u ANR_OLD_PKG CGO_ENABLED=0 go test -count=1 ./... || exit 1
[ -n "$python" ] || { echo "check.sh: no Python 3.9 or later found" >&2; exit 2; }
"$python" scripts/test_check_requirements.py >/dev/null 2>&1 || { "$python" scripts/test_check_requirements.py; exit 1; }
"$python" scripts/check-requirements.py REQUIREMENTS.md || exit $?
[ "$tier" = merge ] && exit 0

# The real packages: install, upgrade, uninstall and downgrade them in a scratch game folder.
if [ -z "${ANR_PKG_DIR:-}" ] || [ -z "${ANR_OLD_PKG:-}" ]; then
	echo "check.sh --release: set ANR_PKG_DIR (a folder with the newest release's anr-fg-*-full.zip and anr-fg-*-addon.zip) and ANR_OLD_PKG (an older anr-fg-*-full.zip)" >&2
	exit 2
fi
# Only the summary lines are shown, but the gate takes the tests' own status: a pipe into grep would hide a
# failure.
real=$(CGO_ENABLED=0 go test -count=1 -run 'Real' -v ./internal/pkg ./internal/record 2>&1) && status=0 || status=$?
printf '%s\n' "$real" | grep -E '^(--- |ok|FAIL|PASS)' || true
if [ "$status" -ne 0 ]; then
	printf '%s\n' "$real" >&2
	exit 1
fi
go tool mygo build || exit 1
