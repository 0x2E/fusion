# Contributing to Fusion

Thanks for your interest in Fusion.

Fusion is maintained by a small team working closely with coding agents, and
the sustainable way to contribute is [issues](https://github.com/0x2E/fusion/issues):
report bugs, propose features, or ask questions there. There is no need to
open pull requests.

Fusion values simple, maintainable changes over complex abstractions.

## 1. Reporting problems

- Search existing issues first to avoid duplicates.
- For bugs, include reproducible steps, expected behavior, and actual behavior.
- For features, explain the concrete use case and user value.

## 2. Local setup

Requirements:

- Go `1.25+`
- Node.js `24+`
- pnpm

Install dependencies:

```shell
# frontend
cd frontend && pnpm install
```

Backend uses Go modules and installs dependencies automatically via `go` tooling.

## 3. Validation checklist (for local changes)

### Backend

```shell
cd backend
go test ./...
goimports -w .
go build -o /dev/null ./cmd/fusion
```

### Frontend

```shell
cd frontend
pnpm typecheck
pnpm test
pnpm run check:i18n
pnpm build
```

`pnpm lint` currently cannot start: `typescript-eslint` does not support
TypeScript 7 yet (typescript-eslint/typescript-eslint#10940). Re-add it to
the checklist once upstream support lands.

If your change is scoped, you can run a smaller test subset first, but run
the relevant final checks before committing.

## 4. Code style guidelines

- Prefer readable, self-explanatory naming.
- Avoid over-engineering.
- Add comments only for non-obvious logic or decisions.
- Keep docs and API behavior aligned when changing contracts.
