# snag - Backlog

Low priority tasks that are not currently scheduled for implementation.

**Last Updated**: 2025-10-19

## GitHub Actions Test Workflow

**Priority**: Low
**Effort**: 2-3 hours
**Status**: Not Started

**Description**: Automated testing workflow that runs on every PR and push to main.

**File**: `.github/workflows/test.yml`

**Triggers**:

- Push to main branch
- Pull requests to main
- Manual workflow dispatch

**Platforms**:

- ubuntu-latest (Chrome pre-installed)
- macos-latest (Chrome pre-installed)

**Workflow Steps**:

```yaml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  workflow_dispatch:

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
    runs-on: ${{ matrix.os }}

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.25.3"

      - name: Download dependencies
        run: go mod download

      - name: Build binary
        run: go build -o snag

      - name: Run tests
        run: go test -v -cover ./...

      - name: Generate coverage report
        run: |
          go test -coverprofile=coverage.out
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage
        uses: actions/upload-artifact@v4
        with:
          name: coverage-${{ matrix.os }}
          path: coverage.html
```

**Benefits**:

- Automated testing on every PR
- Cross-platform validation (Linux + macOS)
- Catches regressions early
- Coverage tracking over time
- Free for public repositories

**Why Low Priority**:

- All 71 tests passing locally (100% pass rate)
- Manual testing workflow is sufficient for current development pace
- Can be added later if team grows or contribution frequency increases
