# Contributing to Fusion

Thanks for contributing.

Fusion values simple, maintainable changes over complex abstractions.

## 1. Before you start

- Search existing issues/PRs first.
- For bugs, include reproducible steps, expected behavior, and actual behavior.
- For features, explain the concrete use case and user value.

## 2. Local setup

Requirements:

- Go `1.26+`
- Node.js `24+`
- pnpm

Install dependencies:

```shell
# frontend
cd frontend && pnpm install
```

Backend uses Go modules and installs dependencies automatically via `go` tooling.

## 3. Validation checklist (required before PR)

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
npx tsc -b --noEmit
pnpm lint
pnpm build
```

If your change is scoped, you can run a smaller test subset first, but run the relevant final checks before requesting review.

## 4. Pull request expectations

- Keep PRs focused. One PR should solve one clear problem.
- Explain why the change is needed, not only what changed.
- Include screenshots/GIFs for UI changes.
- Link related issue(s).
- Mark as Draft if not ready for review.
- If you use AI tools, review and validate outputs carefully before submission.

## 5. Code style guidelines

- Prefer readable, self-explanatory naming.
- Avoid over-engineering.
- Add comments only for non-obvious logic or decisions.
- Keep docs and API behavior aligned when changing contracts.
