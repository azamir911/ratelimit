# Contributing

Contributions are welcome through focused issues and pull requests.

Before submitting a change, run:

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
```

Public API changes should include documentation, tests, and a changelog entry. Keep the core package dependency-free unless a dependency has a clear, measured benefit that cannot reasonably be achieved with the standard library.
