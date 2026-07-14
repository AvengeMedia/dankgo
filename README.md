# dankgo

Common Go modules used by DMS & the Dank Linux suite (dankcalendar, dgop, danksearch, and whatever comes next).

Every app in the suite was carrying its own copy of the same logger, unix-socket RPC, chi/huma scaffolding, and XDG path helpers. This is that code, extracted once.

## Packages

| Package | What it is |
|---|---|
| `log` | charmbracelet logger singleton. `log.SetEnvPrefix("DANKCAL")` → reads `DANKCAL_LOG_LEVEL` / `DANKCAL_LOG_FILE`. File sink strips ANSI. |
| `paths` | XDG dirs + per-process socket paths, keyed by app name. `paths.New("dankcal").SocketPath()` |
| `errdefs` | `CustomError`/`ErrorType` with HTTP status mapping. Apps register their own types from `AppErrorBase` up. stdlib-only. |
| `errdefs/humaerr` | huma error envelope + `HumaErrorFunc`. Split out so non-HTTP apps never pull huma. |
| `ipc` | JSON-over-unix-socket RPC: server (ping/subscribe/unsubscribe built in), client, event bus, mux. Stale sockets get reaped by PID. |
| `ipc/params` | Typed extraction from `map[string]any` request params. |
| `httpapi` | huma/chi scaffold: `NewHumaConfig`, `/docs` page, router with health + recoverer, graceful `Server`. Use à la carte or via `NewRouter`. |
| `httpapi/middleware` | The chi-ported logger/recoverer/request-id set. |
| `netutil` | `GetIPAddress` header dance. |
| `app` | Thin bootstrap: `app.New(info, rootCmd)` wires logging + version, `app.Serve(ctx, runners...)` handles signals and shutdown. Not a framework. |

## Using it

```
go get github.com/AvengeMedia/dankgo@v0.1.0
```

## Minimal daemon

```go
a := app.New(app.Info{Name: "Dank Thing", ID: "dankthing", Version: version}, rootCmd)

srv := ipc.NewServer(ipc.Config{AppName: "dankthing", APIVersion: 1}, route)
if err := srv.Listen(); err != nil {
    return err
}
return app.Serve(ctx, srv)
```
