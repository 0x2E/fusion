# Documentation

## How documentation is written here

- Code is the single source of truth for what it does. Docs answer
  **why**: design decisions, trade-offs, history, and cross-cutting rules
  that are scattered across the code.
- Never copy facts that already live in a file — versions, dependencies,
  route inventories, env-var lists and defaults, schema fields. Link to
  the file instead. The usual truth sources: `backend/go.mod`,
  `frontend/package.json`, `SetupRouter` in
  `backend/internal/handler/handler.go`,
  `backend/internal/store/migrations/`, `frontend/src/routes/`,
  [`.env.example`](../.env.example).
- Each fact lives in exactly one place. Breaking-change notes and other
  history are kept once, in the doc that owns the topic.
- Update docs together with any behavior or contract change they
  describe; if code and a doc disagree, the code wins and the doc is a
  bug.

## Index

- `api-conventions.md`: cross-cutting HTTP API rules (auth, envelopes,
  cursors, errors, status codes)
- `fever-api.md`: Fever protocol compatibility — an external contract
  with third-party clients
- `backend-design.md`: backend architecture, feed pull strategy, schema
  and cascade design
- `frontend-design.md`: frontend interaction model, URL state, data-flow
  and read-state undo design
- `old-database-schema.md`: legacy pre-migrations schema snapshot, kept
  for migration work only
