# Contributing

## Setup

```
make deps
prek install
```

The hooks run golangci-lint (fmt + full + config-verify), `go test`, and `go build`. If they fail, the commit fails. That's the point.

## Code style

- Self-documenting code. Comments are for constraints the code can't express — if you're narrating what the next line does, delete the comment.
- Guard clauses and early returns. No nested happy paths.
- `switch` over `if/else` chains whenever it applies.
- New exported API needs a test.

## Checks by hand

```
make fmt vet test lint
prek run -a
```

## Compatibility

The `ipc` wire format is consumed by QML shells and CLI clients across the suite. Don't add, remove, or rename JSON fields on `Request`, `Response`, `Capabilities`, or the event envelope without checking every consumer first.
