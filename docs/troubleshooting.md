# Troubleshooting

## Installation Fails

- Ensure `go` is in PATH and internet access for `go install`.
- Check logs from [`New`](wasmtest.go:19) for errors (e.g., timeout after 5min).
- Manually install: `go install github.com/agnivade/wasmbrowsertest@latest` and add to PATH.

## "total length of command line and environment variables exceeds limit"

Env vars (e.g., GITHUB_ in CI) too large. Use [cleanenv](https://github.com/agnivade/wasmbrowsertest/tree/main/cmd/cleanenv) to filter:
```
go install github.com/agnivade/wasmbrowsertest/cmd/cleanenv@latest
GOOS=js GOARCH=wasm cleanenv -remove-prefix GITHUB_ -- go test ./...
```

## No Progress Messages or Tests Skip

- Ensure tests have `//go:build js && wasm`.
- Verify in browser env (progress skips if no document for DOM tests).
- Check for `go.test` compilation errors in logs.

## Tests Fail in Headless Mode

Some DOM/JS tests may need `WASM_HEADLESS=off` for visual debugging.

For more, see underlying [wasmbrowsertest docs](docs/wasmbrowsertest.md).