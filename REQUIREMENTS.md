# REQUIREMENTS.md: numbers, gates and exceptions

Commands, numbers, gates and exceptions for this repository. The rules themselves are in `AGENTS.md`,
`UI.md` and `internal/AGENTS.md` and are not repeated here. When another file in this repository
disagrees with this one on a command, a number, a gate or an exception, this file wins.

`scripts/check-requirements.py` checks this file in the merge gate. It reads files and runs
nothing. Write a pipe inside a cell as `\|`. Write a command in backticks; the check confirms that
a script named in it exists.

## Rule files

Every `AGENTS.md` in this repository, by path, and `UI.md`, which holds the interface rules because
the interface code has no directory of its own. The check reads rule groups only from the files
marked `yes`, and fails on an `AGENTS.md` it finds that is not listed. It searches every directory
except `.git`, `node_modules`, `vendor`, `third_party` and hidden directories other than
`.github`. Mark a file `no`, with the reason in Note, when its rules are not this project's (a
vendored package, a generated tree).

| File | Checked | Note |
| --- | --- | --- |
| AGENTS.md | yes | |
| UI.md | yes | |
| internal/AGENTS.md | yes | |

## Commands

One row per tier. The repository has no CI: a person and an agent run the same command locally,
from the repository root, in Git Bash on Windows or a shell elsewhere. Fill in this order: the
command; then three runs on the same host, recorded as the baseline; then the limit, as the
baseline plus a margin (start at 20%).

| Tier | Command | Runs (where, when) | Wall-time limit | Baseline (date, host, load) |
| --- | --- | --- | --- | --- |
| Pre-commit | `scripts/check.sh --fast` | locally, before every commit | 3.5 s | 2.6 s, from 2.6 to 2.9 s over three runs (2026-11-18, dev PC, Windows 11, CPU load 1%) |
| Merge gate | `scripts/check.sh` | locally, before every merge into main | 6 s | 4.3 s (2026-11-18, dev PC, Windows 11, CPU load 1%) |
| Release | `scripts/check.sh --release` | locally, before every release, with ANR_PKG_DIR and ANR_OLD_PKG set | 55 s | 36 s, from 34 to 45 s over three runs, the NSIS installer included (2026-11-18, dev PC, Windows 11, CPU load 1%) |

## Numbers

Every limit has a measured baseline next to it. Record the date, the host and its load for each
baseline. Leave the limit empty until it has been measured, and list it under "Not yet enforced"
meanwhile.

| Item | Limit | Baseline (date, host, load) | Enforced by |
| --- | --- | --- | --- |
| Type scale (role: size / line height) | title 1.6× / 1.2, heading 1.15× / 1.25, body 1× / the theme's, small 0.86× / 1.35, caption 0.79× / 1.3, of the theme's FontSize (22, 16, 14, 12 and 11 DIP at 14) | a setting, not measured | `type.go`; `scripts/check.sh --fast` fails on a size or weight set elsewhere |
| Spacing scale (section, card, inside card, label to field) | | | design system |
| Window size (minimum; default) | 760 × 520 DIP; 1040 × 760 DIP | a setting, not measured | `main.go`, WindowOptions |
| Scan of every game in the Steam libraries | 1 s | 0.62 s for 17 games in 3 libraries (2026-11-18, dev PC, CPU load 0%) | review |
| Slowest test package | 0.3 s | 0.24 s, the main package (2026-11-18, dev PC, CPU load 0%) | review of `go test -v` |
| Release executable size | 16 MB | 13.9 MB (2026-11-18, windows/amd64) | review |
| Source file size, p90 of this repository (pass bar) | 330 lines | 329 lines over 30 Go files (2026-11-18) | review |
| Wait timeout before a stuck step is recorded and left | 30 min | a setting, not measured | agent |
| Context handoff limit (share of the model's window) | 60% | a setting, not measured | orchestrator |
| Status check interval while several jobs run | 10 min | a setting, not measured | orchestrator |
| Commit subject length | 72 characters | a setting, not measured | review |

## Enforcement

One row per rule group of every file checked under "Rule files": each `###` heading, named
`path: heading`, and each rule tagged `[rule: <id>]`, named `path#id`. The check reports a group
with no row, and a row whose group no longer exists, so renaming a heading means renaming its row.

Fill one or both of:

- Gate command: the command, in backticks, that turns red on a violation, run by a tier in
  "Commands";
- Review criterion: the `REVIEW.md` section that checks it, with the file name in backticks.

Or fill only "Does not apply because" with what the project lacks (no admin panel, no i18n). A
tagged rule marked as not applying needs the owner's approval, quoted in the commit body.

| Rule group | Gate command | Review criterion | Does not apply because |
| --- | --- | --- | --- |
| AGENTS.md: Change risk | | `REVIEW.md`, "Scope" | |
| AGENTS.md: Workflow | | `REVIEW.md`, "Pass bar" | |
| AGENTS.md: When unsure | | `REVIEW.md`, "Before you start" | |
| AGENTS.md: Waiting | | `AGENTS.md`, "Waiting": the owner reads the agent's run | |
| AGENTS.md: Reporting | | `REVIEW.md`, "Every finding has evidence" | |
| AGENTS.md: Code quality | `scripts/check.sh --fast` | `REVIEW.md`, "Pass bar" | |
| AGENTS.md: Impact and edge cases | | `REVIEW.md`, "Scope" | |
| AGENTS.md: Secrets | | `REVIEW.md`, "Checking against the rules" | |
| AGENTS.md: Performance baseline | | `REVIEW.md`, "Checking against the rules" | |
| AGENTS.md: Tests (if present) | `scripts/check.sh` | `REVIEW.md`, "Every finding has evidence" | |
| AGENTS.md: Merging | | `REVIEW.md`, "Report" | |
| AGENTS.md: Model usage | | `AGENTS.md`, "Model usage": the orchestrator's brief and run record | |
| AGENTS.md: Documentation | | `REVIEW.md`, "CI, docs and commits" | |
| AGENTS.md: Writing (docs, comments, reports, PR and issue text) | | `REVIEW.md`, "CI, docs and commits" | |
| AGENTS.md: Commits and pull requests | | `REVIEW.md`, "CI, docs and commits" | |
| UI.md: Components and modularity | | `REVIEW.md`, "Checking against the rules" | |
| UI.md: Consistency | | `REVIEW.md`, "Checking against the rules" | |
| UI.md: Layout, type and loading | `scripts/check.sh --fast` | `REVIEW.md`, "Checking against the rules", Layout, type and loading | |
| UI.md: i18n (if present) | | `REVIEW.md`, "Checking against the rules", Copy and i18n | |
| UI.md: Usability | | `REVIEW.md`, "Checking against the rules" | |
| UI.md: Accessibility | | `REVIEW.md`, "Checking against the rules", Accessibility | |
| UI.md: Performance | | `REVIEW.md`, "Checking against the rules" | |
| internal/AGENTS.md: APIs and modularity | | `REVIEW.md`, "Pass bar" | |
| internal/AGENTS.md: Compatibility | | `REVIEW.md`, "Checking against the rules" | |
| internal/AGENTS.md: Logging | | `REVIEW.md`, "Checking against the rules" | |
| internal/AGENTS.md: Background work (if present) | | `REVIEW.md`, "Checking against the rules" | |
| internal/AGENTS.md: Performance | | `REVIEW.md`, "Checking against the rules" | |
| internal/AGENTS.md: Verification | `scripts/check.sh --release` | `REVIEW.md`, "Every finding has evidence" | |

## Exceptions

Backend endpoints with no frontend caller:

| Endpoint | Caller | Reason |
| --- | --- | --- |

Intentional designs that look like defects:

| What | Why | Owner ruling (date, quote) |
| --- | --- | --- |

## Not yet enforced

A row of "Commands", "Numbers" or "Enforcement" that is not filled yet goes here, named
`Section: row`, for example `Numbers: Slowest test package`. The merge gate fails on:

- an unfilled row that is not listed here;
- a listed row that is already filled, or that does not exist;
- a Due date that has passed.

Due holds one date (YYYY-MM-DD). Moving it later needs the owner's approval, quoted in the commit
body.

| Row | Current state | Plan | Due |
| --- | --- | --- | --- |
| Numbers: Spacing scale (section, card, inside card, label to field) | view.go sets literal paddings and gaps | name the gaps in one scale and replace the literals | 2026-12-31 |

## Example (illustrative values; measure your own)

| Tier | Command | Runs (where, when) | Wall-time limit | Baseline (date, host, load) |
| --- | --- | --- | --- | --- |
| Pre-commit | `scripts/check-all.sh --fast` | locally, every commit | 5 s | 3.8 s (2026-11-12, build host, load 0.4) |
| Merge gate | `scripts/check-all.sh --strict` | CI, every merge | 15 min | 11 min (2026-11-12, build host, load 1.2) |

| Rule group | Gate command | Review criterion | Does not apply because |
| --- | --- | --- | --- |
| web/AGENTS.md: i18n (if present) | `scripts/check-i18n.sh` | | |
| server/AGENTS.md#authz | `scripts/check-authz.sh` | `REVIEW.md`, "Checking against the rules", Authorization | |
| server/AGENTS.md: Admin panel (if present) | | | the project has no admin panel |
| AGENTS.md: Writing (docs, comments, reports, PR and issue text) | | `REVIEW.md`, "CI, docs and commits" | |

| Row | Current state | Plan | Due |
| --- | --- | --- | --- |
| Numbers: Slowest test package | not measured | measure after the race suite lands | 2027-01-31 |
