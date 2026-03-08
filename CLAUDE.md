# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

journal-fmt is a formatter for hledger journal files, similar to `gofmt`. It normalizes dates, aligns columns, sorts transactions by date, and formats comments.

## Commands

```bash
# Build and install
go install ./cmd/jf

# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/parser
go test ./internal/formatter

# Run with verbose output
go test -v ./...

# Format a file (output to stdout)
go run ./cmd/jf file.journal

# Format a file (write in place)
go run ./cmd/jf -w file.journal
```

## Architecture

The codebase follows a classic compiler pipeline: **Parse → Format**

```
cmd/jf/main.go          # CLI entry point, file I/O
    ↓
internal/parser/        # Lexer/parser converts journal text → AST
    ↓
internal/types/         # AST types (Journal, Transaction, Posting, etc.)
    ↓
internal/formatter/     # AST → formatted text (column alignment, sorting)
```

### Key Data Flow

1. `Parser.Parse()` reads raw journal text and produces a `*types.Journal`
2. `Formatter.Format()` takes the journal, sorts transactions by date, calculates column widths, and outputs formatted text

### Type Hierarchy

- `Journal` contains a slice of `Item` interface types
- `Item` implementations: `Transaction`, `Posting`, `Comment`, `PriceDecl`, `EmptyLine`, `RawLine`
- `Transaction` contains `[]Posting`

### Formatting Rules

- Dates normalized to `YYYY-MM-DD` (accepts `/` and `.` separators)
- Normal comments: `; text` (single space after semicolon)
- Tag comments: `;  tag:` (two spaces, ends with colon)
- Columns right-aligned based on max width across all transactions
- Transactions sorted by date while preserving non-transaction items in place
- Unrecognized lines (directives, etc.) preserved as `RawLine`
