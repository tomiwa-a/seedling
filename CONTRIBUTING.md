# Contributing to Seedling

Thanks for your interest in contributing!

## Getting Started

1. Fork the repository.
2. Clone your fork.
3. Run `go mod tidy` to install dependencies.
4. Run `go build ./...` to verify the build.

## Development Workflow

- Create a feature branch from `main`.
- Write tests for new functionality.
- Ensure `go vet ./...` and `go test ./...` pass.
- Keep commits small and focused.
- Use conventional commit messages (`feat:`, `fix:`, `chore:`, `docs:`, etc.).

## Running Tests

```bash
go test -v -race -count=1 ./...
```

## Code Style

- Follow standard Go conventions (`gofmt`, `golint`).
- Avoid external dependencies where possible.
- Public API changes should be discussed in an issue first.

## Pull Request Process

1. Update documentation if needed.
2. Ensure CI passes (lint, vet, build, test).
3. Request a review.
4. Squash commits on merge.

## Reporting Issues

Open an issue with:
- A clear description of the problem.
- Steps to reproduce.
- Expected vs actual behavior.
- Go version and OS.
