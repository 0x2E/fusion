# Agent Instructions

## Communication

- Use English for all documentation and code comments.
- Keep responses concise and actionable.
- Challenge proposals when you have a better alternative.

## Project Context

- This project is an open-source, lightweight RSS reader and aggregator.
- Prioritize simplicity and maintainability over complexity.
- HTTP API facts live in code: routes in `backend/internal/handler/handler.go` (`SetupRouter`), response shapes in `backend/internal/model/model.go` and request bodies in the handler files (both mirrored by `frontend/src/lib/api/types.ts`); cross-cutting conventions in `docs/api-conventions.md`.

## Code Standards

- Follow best practices without over-engineering.
- Default to no backward-compatibility work unless explicitly requested; if a change may break data formats, public APIs, or migrations, clearly state the impact.
- Docs answer why, code says what: never copy facts that live in a file (versions, routes, dependencies, env defaults) into documentation — link to the source instead. See `docs/README.md`.
- Write self-explanatory code with clear naming.
- Add comments in English only when they provide non-obvious value:
  - **DO write comments for:**
    - Complex business logic or algorithms
    - Non-obvious design decisions and trade-offs
    - Public APIs, exported functions, and package documentation
    - TODO/FIXME/NOTE markers with context
  - **DON'T write comments for:**
    - Self-evident code (e.g., getters/setters)
    - Repeating what the code already says
    - Implementation details that naming makes clear

## Go Development

- After modifying Go code, run `goimports -w .` before verification.
- Verify compilation with `go build -o /dev/null /path/to/file_or_dir`.
- Run related tests and ensure they pass.
- Use named SQL parameters (e.g., `:param_name` or `@param_name`).

## Frontend Development

- Verify TypeScript/TSX compilation with `npx tsc -b --noEmit`.
- Run frontend unit tests with `pnpm test` (vitest); check i18n completeness with `pnpm run check:i18n`.
- shadcn components use Base UI (`@base-ui/react`) on the `base-vega` style. Regenerate via CLI (`pnpm dlx shadcn@latest add <component> --overwrite`) instead of hand-editing source files in `frontend/src/components/ui/`.
