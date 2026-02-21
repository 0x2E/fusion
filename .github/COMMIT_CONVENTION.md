# Commit Message Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/) for automated versioning and changelog generation via [release-please](https://github.com/googleapis/release-please).

## Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Types

The following commit types will appear in the changelog:

- **feat**: A new feature (triggers minor version bump)
- **fix**: A bug fix (triggers patch version bump)
- **perf**: Performance improvements (triggers patch version bump)
- **docs**: Documentation changes
- **refactor**: Code refactoring without behavior changes
- **build**: Changes to build system or dependencies
- **revert**: Reverts a previous commit

The following types are valid but hidden from changelog:

- **style**: Code style changes (formatting, whitespace, etc.)
- **test**: Adding or updating tests
- **ci**: CI/CD configuration changes
- **chore**: Other changes that don't modify src or test files

## Breaking Changes

Add `!` after the type/scope or include `BREAKING CHANGE:` in the footer to trigger a major version bump:

```
feat!: remove support for SQLite < 3.35

BREAKING CHANGE: SQLite version 3.35 or higher is now required
```

## Scope

Optional scope indicates the area of change:

- `backend`: Backend Go code
- `frontend`: Frontend TypeScript/React code
- `api`: API changes
- `db`: Database schema or migrations
- `docker`: Docker configuration
- `deps`: Dependency updates

## Examples

### Feature commits

```
feat(frontend): add dark mode toggle

feat(backend): implement PostgreSQL connection pooling

feat!: migrate to new config format

BREAKING CHANGE: FUSION_* environment variables renamed to REEDME_*
```

### Bug fix commits

```
fix(backend): prevent race condition in feed puller

fix(frontend): correct article pagination offset
```

### Other commits

```
docs: update installation instructions

refactor(backend): simplify error handling in store

perf(frontend): optimize feed list rendering

build(deps): update Go dependencies

chore: update .gitignore
```

## Tips

- Use imperative mood: "add feature" not "added feature"
- Don't capitalize first letter of description
- No period at the end of description
- Keep description under 72 characters
- Use body to explain *what* and *why*, not *how*
- Reference issues in footer: `Fixes #123` or `Closes #456`

## Versioning

Based on commit types:

- `feat` → minor version bump (0.1.0 → 0.2.0)
- `fix`, `perf` → patch version bump (0.1.0 → 0.1.1)
- `BREAKING CHANGE` or `!` → major version bump (0.1.0 → 1.0.0)

Before 1.0.0:
- `feat` → patch bump (0.1.0 → 0.1.1)
- `BREAKING CHANGE` → minor bump (0.1.0 → 0.2.0)

## Release Process

1. Push commits to `main` branch
2. release-please bot creates/updates a release PR
3. Review the generated CHANGELOG.md in the PR
4. Merge the release PR when ready
5. release-please creates a GitHub release with tag
6. CI automatically builds and publishes artifacts
