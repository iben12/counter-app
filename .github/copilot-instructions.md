<!--
  Machine-generated guidance for AI coding agents working on this repo.
  This file was created because a quick search found no existing AI-agent docs
  (no README or copilot/AGENT files). If you merge with an existing file, keep
  human-written sections verbatim and add/replace the "Agent checklist" area.
-->

# Copilot / AI agent instructions — counter-app

Purpose
- Help an AI coding agent be immediately productive in this repository.

Repository scan (quick summary)
- At time of generation there was no README.md, no .github/copilot-instructions.md,
  and no AGENT/AGENTS/CLAUDE rule files found. Treat this as a new baseline.

What to do first (explicit, reproducible steps)
1. Run a filename scan to locate language and entry points. Look for these files (in order):
   - `package.json`, `yarn.lock`, `pnpm-lock.yaml` (Node.js)
   - `pyproject.toml`, `requirements.txt`, `Pipfile` (Python)
   - `go.mod` (Go), `Cargo.toml` (Rust), `pom.xml` (Java)
   - `Dockerfile`, `docker-compose.yml`
   - `.github/workflows/*` (CI clues)
   - `README.md`, `CHANGELOG.md`
   - Common entry folders: `src/`, `cmd/`, `app/`, `server/`, `web/`, `client/`

2. Open the first matching manifest you find and extract the canonical run/test commands:
   - Node: prefer `npm run` / `yarn` scripts from `package.json`.
   - Python: prefer `poetry`/`pip` commands from `pyproject.toml` or `requirements.txt`.
   - Go/Rust: use `go build`/`go test` or `cargo build`/`cargo test` as shown in manifests.

3. If no manifest is present, ask the human owner two short questions:
   - "What is the primary language/runtime for counter-app?"
   - "How do you normally run tests and start the app locally?"

Project-specific assumptions (explicit)
- Repo name: `counter-app`. Assume it's a small app with a service and/or frontend.
- These are assumptions only. Confirm with the maintainer before making architecture changes.

Agent checklist (what an AI may safely do now)
- Prefer read-only exploration: list top-level files and open manifests and any `src/` files.
- Extract concrete build/test commands from discovered files and record them in a short summary.
- When creating or editing files, keep changes minimal and isolated; add tests where small and fast.

What NOT to do without confirmation
- Don’t update CI workflows or version bump files without the maintainer’s approval.
- Don’t assume external services (DBs, caches) are available — find `docker-compose.yml` or CI jobs first.

Integration & infra hints to look for
- `docker-compose.yml`, `Dockerfile`: indicates local infra and service dependencies.
- `.env` or `.env.sample`: lists required environment variables.
- `.github/workflows/*`: reveals build matrix, test steps, and release automation.

How to produce a minimal PR
1. Create a branch `ai/<short-descr>`.
2. Make a minimal change with tests (happy path only) and run local tests.
3. In PR description include:
   - One-sentence summary.
   - Commands you ran and test results.
   - Any assumptions you made (list files that motivated them).

If you find an existing `.github/copilot-instructions.md` or `AGENT.md`
- Preserve human-written sections. Where you add content, annotate with `<!-- AI: added -->`.

Questions for the maintainer
- Is `counter-app` primarily a frontend web app, a backend service, or a fullstack project?
- Preferred commands to build, test, and run locally (if non-standard).

End
--