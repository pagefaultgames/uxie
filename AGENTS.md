# uxie

## Repository structure
- Discord bot written in Go using [tempest](https://github.com/amatsagu/tempest)
- Stores available help topics inside an SQLite database using [go-sqlite3](https://github.com/mattn/go-sqlite3)

## Important commands

- `task lint`/`task fmt` - Run linters/formatting
  - Should be done after EVERY commit!
  - **Do NOT use `gofmt`**; it does not support additional linting directives
- `task test` - Run automated tests