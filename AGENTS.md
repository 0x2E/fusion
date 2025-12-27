# Agent Instructions

## Communication

- All documentation and code comments in English.
- Keep responses concise.
- Challenge ideas when you have better insights.

## Code Standards

- Follow best practices without over-engineering.
- Write self-explanatory code with clear naming.
- Comments in English only when necessary:
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

- After modifying Go code, verify compilation: `go build -o /dev/null /path/to/file_or_dir`.
- Run related tests to ensure they pass.
- Use named parameters in SQL queries (e.g., `:param_name` or `@param_name`).
