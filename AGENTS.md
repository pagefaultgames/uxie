<!--
SPDX-FileCopyrightText: 2025-2026 Pagefault Games

SPDX-License-Identifier: CC-BY-NC-SA-4.0
-->

# uxie

## Repository structure
- Discord bot written in Go using [tempest](https://github.com/amatsagu/tempest)
- Stores available help topics inside a MySQL database using [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql)

## Important commands

- `task lint`/`task fmt` - Run linters/formatting
  - Should be done after EVERY commit!
  - **Do NOT use `gofmt`**; it does not support additional linting directives
<!-- - `task test` - Run automated tests -->