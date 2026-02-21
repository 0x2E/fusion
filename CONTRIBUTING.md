# Contributing to ReedMe

Thanks for contributing.

ReedMe values simple, maintainable changes over complex abstractions.

## 1. Before you start

- Search existing issues/PRs first.
- For bugs, include reproducible steps, expected behavior, and actual behavior.
- For features, explain the concrete use case and user value.

## 2. Local setup

Requirements:

- Go `1.25+`
- Node.js `24+`
- pnpm

Install dependencies and configure git hooks:

```shell
# git hooks (commit message linting)
./scripts.sh setup-hooks

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
go build -o /dev/null ./cmd/reedme
```

### Frontend

```shell
cd frontend
npx tsc -b --noEmit
pnpm lint
pnpm build
```

If your change is scoped, you can run a smaller test subset first, but run the relevant final checks before requesting review.

## 4. Commit message convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for automated releases.

Format: `<type>[optional scope]: <description>`

Common types:
- `feat`: New feature (minor version bump)
- `fix`: Bug fix (patch version bump)
- `docs`: Documentation changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `chore`: Other changes

Examples:
```
feat(frontend): add dark mode toggle
fix(backend): prevent race condition in feed puller
docs: update installation instructions
```

See [.github/COMMIT_CONVENTION.md](.github/COMMIT_CONVENTION.md) for full details.

## 5. Pull request expectations

- Keep PRs focused. One PR should solve one clear problem.
- Use conventional commit format for commit messages.
- Explain why the change is needed, not only what changed.
- Include screenshots/GIFs for UI changes.
- Link related issue(s).
- Mark as Draft if not ready for review.
- If you use AI tools, review and validate outputs carefully before submission.

## 6. Code style guidelines

- Prefer readable, self-explanatory naming.
- Avoid over-engineering.
- Add comments only for non-obvious logic or decisions.
- Keep docs and API behavior aligned when changing contracts.
