# Agent Instructions

## Communication

- Use English for all documentation and code comments.
- Keep responses concise and actionable.
- Challenge proposals when you have a better alternative.

## Project Context

- This project is an open-source, lightweight RSS reader and aggregator.
- Prioritize simplicity and maintainability over complexity.

## Code Standards

- Follow best practices without over-engineering.
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
- Do not modify shadcn component source files directly.
