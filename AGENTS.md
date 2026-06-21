# AGENTS.md

## Repository role

This repository is a small Go command-line tool, `siliconflow-codegen`, that fetches the SiliconFlow model list and emits either:

- the raw SiliconFlow API response
- an OpenCode-compatible provider config
- a Crush-compatible provider config
- a Qwencode-compatible provider config
- an interactive configuration and shell environment exporter for Claude Code (`--claude`)

It is intentionally narrow: one package, one source file, one test file, and a Makefile that wraps build/test/generation commands.

## Directory layout

- `siliconflow-codegen.go` — only source file; contains CLI parsing, HTTP fetch, config generation, JSON parsing, and output helpers.
- `siliconflow-codegen_test.go` — package-level tests for config generation and model ID parsing.
- `Makefile` — primary command surface for build, dist, generation, tests, formatting, and cleanup.
- `README.md` — user-facing documentation for API key setup, generated config shapes, and troubleshooting.
- `go.mod` — Go module definition; currently no external dependencies.

## Essential commands

Use `make` as the command entry point unless there is a specific reason to call Go directly.

```bash
make test      # Run all Go tests
make build     # Build dist/siliconflow-codegen
make dist      # Build Linux ARM64, Linux AMD64, and Darwin ARM64 binaries
make format    # Run gofmt on siliconflow-codegen.go
make clean     # Remove dist/ and generated JSON files
make help      # Print Make targets
```

Generation and raw model fetching require `SILICONFLOW_API_KEY`:

```bash
make raw-models
make gen-opencode
make gen-crush
```

Equivalent direct Go commands:

```bash
go run .
go run . --gen-opencode
go run . --gen-crush
```

Generated files:

- `siliconflow.models.json` from `make raw-models`
- `siliconflow.opencode.json` from `make gen-opencode`
- `siliconflow.crush.json` from `make gen-crush`

The Makefile also supports output-path overrides:

```bash
make gen-opencode OPENCODE_CONFIG=path/to/file.json
make gen-crush CRUSH_CONFIG=path/to/file.json
```

## API key and secret handling

`SILICONFLOW_API_KEY` is required for every run path, including raw model fetching.

The tool must not write the real API key into generated configs:

- OpenCode uses `${SILICONFLOW_API_KEY}` in `siliconflow.apiKey`.
- Crush uses `$SILICONFLOW_API_KEY` in `providers.siliconflow.api_key`.

The fetch request uses:

- URL: `https://api.siliconflow.com/v1/models`
- `Accept: application/json`
- `Authorization: Bearer ${SILICONFLOW_API_KEY}`
- 30-second HTTP client timeout

If the API returns a non-200 status, the program prints the status and response body. Treat that as a troubleshooting signal, not as a generated-config failure.

## Generated config conventions

OpenCode generation:

- Top-level key: `siliconflow`
- Provider `type`: `openai`
- `baseURL`: `https://api.siliconflow.com/v1`
- Models are stored as a map keyed by model ID.
- Each model has `name` equal to the model ID.
- Each model has fixed modalities:
  - input: `text`, `image`, `video`, `audio`
  - output: `text`

Crush generation:

- Top-level `$schema`: `https://charm.land/crush.json`
- Providers are stored under `providers.siliconflow`.
- Provider `type`: `openai`
- `base_url`: `https://api.siliconflow.com/v1`
- Models are stored as an array.
- Each model has `id` and `name`, both equal to the model ID.
- Optional pricing/context fields exist in the struct but are currently not populated.

## Source patterns

The program is a single `package main` binary. Keep changes local and simple unless there is a clear reason to split packages.

Important functions:

- `main()` — parses flags, rejects simultaneous use of multiple generator/action flags, checks the API key, fetches models, and dispatches the requested output mode.
- `fetchModels(apiKey string)` — performs the HTTP GET and validates status/body.
- `generateOpenCodeConfig(body []byte)` — parses IDs, builds the OpenCode JSON shape, and writes JSON to stdout.
- `generateCrushConfig(body []byte)` — parses IDs, builds the Crush JSON shape, and writes JSON to stdout.
- `parseModelIDs(body []byte)` — extracts, trims, deduplicates, and sorts model IDs.
- `printRawResponse(body []byte)` — writes the raw API response to stdout and ensures a trailing newline.

Model ID handling is shared by both generators:

- `strings.TrimSpace` is applied before validation.
- Empty IDs are skipped.
- Duplicates are removed.
- IDs are sorted before output.

This means tests and generated output should not depend on the API response order.

## Testing approach

Run:

```bash
make test
```

Current tests are table-free package tests in `siliconflow-codegen_test.go`. They focus on:

- OpenCode config shape and placeholder API key
- Crush config shape and placeholder API key
- rejection of empty model lists
- deduplication, trimming, and sorting of model IDs

When adding tests, prefer small focused tests around generator behavior and parsing edge cases. Avoid tests that require network access or a real API key.

## Style and conventions

- Use standard Go formatting via `gofmt`.
- Keep generated JSON deterministic: model IDs are sorted, and OpenCode modality slices are sorted before marshaling.
- Error messages emitted by the CLI currently include an `ERROR:` prefix. Preserve that user-visible style for new command-line errors.
- Use `fmt.Errorf("ERROR: ...: %w", err)` for wrapped internal errors.
- Keep the public behavior aligned with `README.md`; if behavior changes, update both code and README.
- The Makefile is the command contract. If adding a new command path, add a target and document it.

## Gotchas

- `--gen-opencode`, `--gen-crush`, `--gen-qwencode`, and `--claude` are mutually exclusive; running more than one exits with an error.
- `make gen-opencode`, `make gen-crush`, and `make gen-qwencode` overwrite their target files through shell redirection.
- `make clean` removes `dist/` plus generated `siliconflow.opencode.json`, `siliconflow.crush.json`, `siliconflow.qwencode.json`, and `siliconflow.models.json`.
- `dist/` is ignored by Git.
- There is no lint target in the Makefile. Use `gofmt` and `go test ./...` as the baseline checks.
- The source file currently has an unused `opencodeProviderKey` constant. Do not rely on it for behavior; remove it only if making a cleanup change.

## Existing rule files

No `.cursor/rules/*.md`, `.cursorrules`, `.github/copilot-instructions.md`, `claude.md`, or prior `AGENTS.md` file was present when this file was created.
