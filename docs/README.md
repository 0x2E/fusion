# Documentation Index

- `openapi.yaml`: current HTTP API contract (source of truth)
- `fever-api.md`: Fever API compatibility contract (`/fever*` endpoints)
- `backend-design.md`: backend architecture and operational notes
- `frontend-design.md`: frontend architecture and interaction model
- `old-database-schema.md`: legacy schema snapshot kept for migration work

Release note: API and design docs should be updated together with any behavior or contract changes.

Current breaking API note: feed runtime fields are now under `feed.fetch_state.*` (top-level `last_build/last_failure_at/failure/failures` removed).
