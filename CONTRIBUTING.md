# Contributing

## Verify changes

Run the build and test suites from the repository root:

```bash
go build ./...
go test ./...
go test -bench=.
```

## Pull requests

Keep changes focused and include tests for new behavior. Ensure the build,
tests, and benchmarks pass before opening a pull request. Use clear,
Conventional Commits-style messages and describe the motivation and testing
performed in the pull request.
