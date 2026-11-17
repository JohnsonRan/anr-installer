# internal/AGENTS.md: backend rules

Rules for every change to the files under this directory. The root `AGENTS.md` (rule priority,
workflow, CI, writing, commits) applies as well; numbers are in `REQUIREMENTS.md`.

## Backend

### APIs and modularity

- Clear API abstraction and module boundaries.
- No hardcoded URLs, paths or config values. Upstream fees, limits and rules are data, not code.
- No broken endpoints.

### Compatibility

- If a config format, a database or storage schema, or API behaviour changes, old data and old
  clients keep working, or a migration is provided.
- Migrations (if present) only move forward, can be rerun safely, and apply on a database that
  already holds rows.

### Logging

- One log format, with levels used for their real meaning: debug, info, warn, error.
- Do not log expected cases as errors.
- Do not swallow errors silently.
- No secrets in logs (root `AGENTS.md`, "Secrets").

### Background work (if present)

- A failing or panicking background round is contained: it is logged with its job name, it raises
  an alert, and the service and the other jobs keep running.

### Performance

- Watch storage size, CPU and memory against the baseline.
- No unnecessary dependencies.

### Verification

- When needed, run and test in an isolated environment: a VM, or a network namespace with
  disposable data.
- Verify against a real database. Builds, vet and unit tests pass through defects that only a real
  database catches.
