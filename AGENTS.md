# Agent Instructions

## Communication

- Use English for all documentation and code comments.
- Challenge proposals when you have a better alternative.

## Project Context

- This project is an open-source, lightweight RSS reader and aggregator. Prioritize simplicity and maintainability over complexity.
- HTTP API facts live in code: routes in `backend/internal/handler/handler.go` (`SetupRouter`), response shapes in `backend/internal/model/model.go` and request bodies in the handler files (both mirrored by `frontend/src/lib/api/types.ts`); cross-cutting conventions in `docs/api-conventions.md`.

## Code Standards

- Default to no backward-compatibility work unless explicitly requested; if a change may break data formats, public APIs, or migrations, clearly state the impact.
- Docs answer why, code says what: never copy facts that live in a file (versions, routes, dependencies, env defaults) into documentation — link to the source instead. See `docs/README.md`.
- Comments explain why: complex logic, non-obvious decisions and trade-offs, exported APIs, and TODO/FIXME/NOTE markers with context.

## Verification

Run the check for each side you touched; these are what CI runs:

- Backend: `./scripts.sh test-backend`
- Frontend: `./scripts.sh test-frontend` (install, typecheck, unit tests, i18n completeness)

## Go Development

- The Go module lives in `backend/`. After modifying Go code, run `go tool goimports -w .` there.
- Use named SQL parameters (e.g., `:param_name` or `@param_name`).

## Frontend Development

- `routeTree.gen.ts` is generated and gitignored; `pnpm run typecheck` regenerates it.
- `pnpm lint` is currently broken (`typescript-eslint` lacks TypeScript 7 support); leave it out of verification until upstream lands.
- shadcn components use Base UI (`@base-ui/react`) on the `base-vega` style. Regenerate via CLI (`pnpm dlx shadcn@latest add <component> --overwrite`) instead of hand-editing source files in `frontend/src/components/ui/`.

## Agent skills

### Issue tracker

Agent specs and tickets live as local, gitignored markdown files under `.scratch/<feature>/`; user-reported issues live in GitHub Issues. See `docs/agents/issue-tracker.md`.

### Triage labels

Default five-role vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`), recorded as a `Status:` line in each issue file. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.
