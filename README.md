# gocov

[![Test](https://github.com/blck-snwmn/gocov/actions/workflows/test.yml/badge.svg)](https://github.com/blck-snwmn/gocov/actions/workflows/test.yml)

A Go tool that aggregates and displays test coverage by directory.

## Features

- Directory-level coverage aggregation
- Flexible aggregation by hierarchy level
- Coverage rate filtering
- Diff coverage (changed lines only)
- Configuration file support (`.gocov.yml`)
- Concurrent processing for performance
- JSON output support

## Installation

```bash
go install github.com/blck-snwmn/gocov@latest
```

## Usage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out -coverpkg=./... ./...

# Aggregate by directory
gocov -coverprofile=coverage.out
```

## Options

| Option | Description | Default |
|--------|-------------|----------|
| `-coverprofile` | Coverage profile file | Required |
| `-level` | Aggregation level (0:leaf, N:N levels, -1:top) | 0 |
| `-min` | Minimum coverage filter (0-100) | 0 |
| `-max` | Maximum coverage filter (0-100) | 100 |
| `-format` | Output format (table/json) | table |
| `-ignore` | Ignore patterns (comma-separated) | - |
| `-threshold` | Threshold check (for CI) | 0 |
| `-diff` | Diff coverage (HEAD~1, main, staged, etc.) | - |
| `-concurrent` | Enable concurrent processing | false |
| `-config` | Configuration file path | .gocov.yml |

## Output Examples

### Basic Usage
```
$ gocov -coverprofile=coverage.out
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
github.com/example/project/cmd/server                      7          5   71.4%
github.com/example/project/internal/service                7          6   85.7%
github.com/example/project/pkg/util                        7          5   71.4%
--------------------------------------------------------------------------------
TOTAL                                                     21         16   76.2%
```

### Level Aggregation (-level 4)
```
$ gocov -coverprofile=coverage.out -level 4
Directory                                          Statements    Covered Coverage
--------------------------------------------------------------------------------
github.com/example/project/cmd                             7          5   71.4%
github.com/example/project/internal                        7          6   85.7%
github.com/example/project/pkg                             7          5   71.4%
--------------------------------------------------------------------------------
TOTAL                                                     21         16   76.2%
```

### Diff Coverage
```
$ gocov -coverprofile=coverage.out -diff HEAD~1
Diff Coverage Report:
================================================================================
File                                                    Lines    Covered Coverage
--------------------------------------------------------------------------------
internal/service/user.go                                   45         38    84.4%
  Uncovered lines: [23 24 67 89 90]
--------------------------------------------------------------------------------
TOTAL DIFF                                                 45         38    84.4%
```

## Configuration File

Persist settings with `.gocov.yml`:

```yaml
level: 0
coverage:
  min: 0
  max: 100
format: table
ignore:
  - "*/vendor/*"
  - "*/test/*"
concurrent: true
threshold: 80
```

Command-line arguments override configuration file values.

## CI/CD Integration

### GitHub Actions
```yaml
- name: Run tests with coverage
  run: go test -coverprofile=coverage.out -coverpkg=./... ./...

- name: Check coverage threshold
  run: |
    go install github.com/blck-snwmn/gocov@latest
    gocov -coverprofile=coverage.out -threshold 80
```

## Requirements

- Go 1.25.0 or higher
- git (for diff coverage feature)