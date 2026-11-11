---
name: verify
description: Check that TamiAI-TUI builds and passes its quality gates without spending tokens (gofmt, vet, race tests, lint, vulnerabilities, agents listing). Use after any Go change and before saying something is done or committing.
---

# Verify TamiAI-TUI

Run in this order and stop at the first failure:

```bash
gofmt -l .                       # must print nothing
go vet ./...
go test -race -count=1 ./...
go mod tidy -diff                # must print nothing
golangci-lint run ./...          # if installed
govulncheck ./...                # if installed
go run ./cmd/tami agents         # must list the agents without errors
```

Rules:
- If `golangci-lint` or `govulncheck` is missing, report it as **skipped** with its install command
  (`brew install golangci-lint`, `go install golang.org/x/vuln/cmd/govulncheck@latest`). Don't install
  tools yourself.
- Never run `go run ./cmd/tami run ...`, or `serve` with requests: they start real agent CLIs that
  spend tokens and can edit files. If a real run is needed, give the user the command with
  `-dir /tmp/tami-test`.
- Don't leave binaries in the repo.
- Report the result of each step; on failure show the error exactly as printed.
