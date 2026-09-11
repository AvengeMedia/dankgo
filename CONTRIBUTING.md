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

## Generative AI

Using an LLM to help write code, issues, or comments is fine. Submitting its output unread is not.

- You are responsible for every line you submit. You have read it, tested it, and can explain it in review.
- Say in the PR when a meaningful part of it was AI generated.
- Do not file issues or leave comments you have not verified yourself. Reports that do not reproduce get closed.
- PRs that read like unreviewed output, with narrating comments, invented APIs, or style that ignores the file they are in, get closed without review.
