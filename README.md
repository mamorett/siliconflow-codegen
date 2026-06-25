# siliconflow-codegen

`siliconflow-codegen` is a tiny Go command-line tool that talks to the SiliconFlow model API and can generate provider configurations for OpenCode, Crush, or Qwencode.

By default, it prints the raw SiliconFlow model API response. With `--gen-opencode`, `--gen-crush`, or `--gen-qwencode`, it converts the discovered model IDs into a ready-to-use provider config.

The generated provider is intentionally SiliconFlow-specific:

- provider key: `siliconflow` for OpenCode and Crush
- provider type: `openai`
- base URL: `https://api.siliconflow.com/v1`
- API key placeholder: `${SILICONFLOW_API_KEY}` for OpenCode, `$SILICONFLOW_API_KEY` for Crush, or `SILICONFLOW_API_KEY` for Qwencode
- models: every model ID returned by the SiliconFlow API
- OpenCode model input modalities: `text`, `image`, `video`, `audio`
- OpenCode model output modalities: `text`
- Qwencode model generation config: `modalities.image = true`

## Requirements

You need:

- Go 1.22 or newer
- `make`
- A SiliconFlow API key exposed as `SILICONFLOW_API_KEY`
- Optional: `jq`, useful for inspecting generated JSON
- Optional: [claude-code-router](https://github.com/musistudio/claude-code-router), required when using the `--claude` option (see [installation](#installing-claude-code-router))

Check your Go version:

```bash
go version
```

Check that `make` is available:

```bash
make --version
```

## API key

The SiliconFlow API key is mandatory. Export it before running the tool:

```bash
export SILICONFLOW_API_KEY="your-siliconflow-api-key"
```

The generated OpenCode config will contain:

```json
"apiKey": "${SILICONFLOW_API_KEY}"
```

The generated Crush config will contain:

```json
"api_key": "$SILICONFLOW_API_KEY"
```

The generated Qwencode config will contain:

```json
"envKey": "SILICONFLOW_API_KEY"
```

The tool never writes your real API key into the generated config file.

## Fetch the raw model list

Without flags, the program prints the raw SiliconFlow API response:

```bash
go run .
```

You can also save the raw response:

```bash
make raw-models
```

This writes:

```text
siliconflow.models.json
```

The raw endpoint is:

```text
https://api.siliconflow.com/v1/models
```

The request uses:

```text
Accept: application/json
Authorization: Bearer ${SILICONFLOW_API_KEY}
```

## Generate the Qwencode config

To generate the Qwencode-compatible SiliconFlow provider config:

```bash
go run . --gen-qwencode
```

Or through Make:

```bash
make gen-qwencode
```

By default, this writes:

```text
siliconflow.qwencode.json
```

To choose another output path:

```bash
make gen-qwencode QWENCODE_CONFIG=qwencode-providers/siliconflow.json
```

## Generate the OpenCode config

To generate the OpenCode-compatible SiliconFlow provider config:

```bash
go run . --gen-opencode
```

Or through Make:

```bash
make gen-opencode
```

By default, this writes:

```text
siliconflow.opencode.json
```

To choose another output path:

```bash
make gen-opencode OPENCODE_CONFIG=openai-providers/siliconflow.json
```

## Generate the Crush config

To generate the Crush-compatible SiliconFlow provider config:

```bash
go run . --gen-crush
```

Or through Make:

```bash
make gen-crush
```

By default, this writes:

```text
siliconflow.crush.json
```

To choose another output path:

```bash
make gen-crush CRUSH_CONFIG=crush-providers/siliconflow.json
```

## Set the model for Claude Code CLI

You can use the `--claude` flag to interactively select a model from SiliconFlow and assign it directly to Claude Code's environment.

To make the selection persistent across sessions, the tool automatically updates your `~/.claude/settings.json` file, mapping:
* `ANTHROPIC_BASE_URL` to `"http://localhost:3456"` (the local Claude Code Router proxy)
* `ANTHROPIC_MODEL` to the selected model ID (e.g. `"deepseek-ai/DeepSeek-V3"`)
* `ANTHROPIC_API_KEY` to the value of your `$SILICONFLOW_API_KEY`

Additionally, it automatically configures `~/.claude-code-router/config.json` by adding or updating the `siliconflow` provider with:
* `api_base_url` set to the full chat completions endpoint: `"https://api.siliconflow.com/v1/chat/completions"`
* `transformer.use` set to `["OpenAI"]` (and dynamically appends `"reasoning"` if a reasoning model like R1 is selected)

> [!IMPORTANT]
> After updating the configuration, you must restart your Claude Code Router service for the changes to take effect:
> ```bash
> ccr restart
> ccr code # to start Claude Code with the router
> ```

Additionally, the tool outputs shell export statements on `stdout` so you can set these variables immediately in your current terminal session using `eval`:

```bash
eval $(./dist/siliconflow-codegen --claude)
```

Using `make`:

```bash
eval $(make -s claude)
```

### Installing claude-code-router

The `--claude` option requires [claude-code-router](https://github.com/musistudio/claude-code-router) to be installed globally:

```bash
npm install -g @musistudio/claude-code-router
```

Verify the installation:

```bash
ccr --version
```

### How it works

1. It fetches the latest model list from SiliconFlow.
2. It displays a clean, column-aligned grid of available models on `stderr`.
3. It prompts you to enter a number to make a selection.
4. It updates (or creates) `~/.claude/settings.json` with the environment variables.
5. It updates (or creates) `~/.claude-code-router/config.json` with SiliconFlow provider settings.
6. It prints the `export` shell commands to `stdout`, allowing `eval` to update the current shell session.

## Generated OpenCode config shape

The generated file has this structure:

```json
{
  "siliconflow": {
    "type": "openai",
    "baseURL": "https://api.siliconflow.com/v1",
    "apiKey": "${SILICONFLOW_API_KEY}",
    "models": {
      "ByteDance-Seed/Seed-OSS-36B-Instruct": {
        "name": "ByteDance-Seed/Seed-OSS-36B-Instruct",
        "modalities": {
          "input": [
            "audio",
            "image",
            "text",
            "video"
          ],
          "output": [
            "text"
          ]
        }
      }
    }
  }
}
```

Every model returned by the SiliconFlow API becomes an entry under `models`, keyed by its model ID.

Example model entry:

```json
"<model-id>": {
  "name": "<model-id>",
  "modalities": {
    "input": [
      "audio",
      "image",
      "text",
      "video"
    ],
    "output": [
      "text"
    ]
  }
}
```

The input modalities always include:

```text
text
image
video
audio
```

The output modality is currently:

```text
text
```

## Generated Crush config shape

The generated file has this structure:

```json
{
  "$schema": "https://charm.land/crush.json",
  "providers": {
    "siliconflow": {
      "type": "openai",
      "base_url": "https://api.siliconflow.com/v1",
      "api_key": "$SILICONFLOW_API_KEY",
      "models": [
        {
          "id": "ByteDance-Seed/Seed-OSS-36B-Instruct",
          "name": "ByteDance-Seed/Seed-OSS-36B-Instruct"
        }
      ]
    }
  }
}
```

Every model returned by the SiliconFlow API becomes an entry under `providers.siliconflow.models`, keyed by its model ID in the `id` field.

## Generated Qwencode config shape

The generated file has this structure:

```json
{
  "openai": [
    {
      "id": "ByteDance-Seed/Seed-OSS-36B-Instruct",
      "name": "ByteDance-Seed/Seed-OSS-36B-Instruct",
      "envKey": "SILICONFLOW_API_KEY",
      "baseUrl": "https://api.siliconflow.com/v1",
      "generationConfig": {
        "modalities": {
          "image": true
        }
      }
    }
  ]
}
```

Every model returned by the SiliconFlow API becomes an entry in the top-level `openai` array, sorted by model ID.

## Build locally

Build a native binary for your current machine:

```bash
make build
```

This creates:

```text
dist/siliconflow-codegen
```

Run it directly:

```bash
./dist/siliconflow-codegen
./dist/siliconflow-codegen --gen-opencode
./dist/siliconflow-codegen --gen-crush
./dist/siliconflow-codegen --gen-qwencode
```

## Cross-platform builds

Build all release binaries:

```bash
make dist
```

This creates:

```text
dist/siliconflow-codegen-linux-arm64
dist/siliconflow-codegen-linux-amd64
dist/siliconflow-codegen-darwin-arm64
```

The Makefile also exposes individual build targets.

### Linux ARM64

```bash
make dist/siliconflow-codegen-linux-arm64
```

or:

```bash
make gen-opencode-linux-arm64
```

### Linux AMD64

```bash
make dist/siliconflow-codegen-linux-amd64
```

or:

```bash
make gen-opencode-linux-amd64
```

### macOS / Darwin ARM64

```bash
make dist/siliconflow-codegen-darwin-arm64
```

or:

```bash
make gen-opencode-darwin-arm64
```

The platform-specific OpenCode generation targets first build the matching binary, then run it with `--gen-opencode` to create:

```text
siliconflow.opencode.json
```

## Makefile target reference

| Target | Description |
| --- | --- |
| `make build` | Build the native binary into `dist/`. |
| `make dist` | Build Linux ARM64, Linux AMD64 and macOS ARM64 binaries. |
| `make gen-opencode` | Generate `siliconflow.opencode.json` using `go run . --gen-opencode`. |
| `make gen-crush` | Generate `siliconflow.crush.json` using `go run . --gen-crush`. |
| `make gen-qwencode` | Generate `siliconflow.qwencode.json` using `go run . --gen-qwencode`. |
| `make claude` | Interactively select a SiliconFlow model, update Claude Code settings, and print exports. |
| `make gen-opencode-linux-arm64` | Build the Linux ARM64 binary, then generate the OpenCode config. |
| `make gen-opencode-linux-amd64` | Build the Linux AMD64 binary, then generate the OpenCode config. |
| `make gen-opencode-darwin-arm64` | Build the macOS ARM64 binary, then generate the OpenCode config. |
| `make raw-models` | Fetch and save the raw SiliconFlow API response to `siliconflow.models.json`. |
| `make test` | Run Go tests. |
| `make format` | Format the Go source with `gofmt`. |
| `make clean` | Remove generated binaries and generated JSON files. |
| `make help` | Print available Make targets. |

## Inspect the generated OpenCode config

After generation, inspect the top-level keys and model count with `jq`:

```bash
jq 'keys' siliconflow.opencode.json
jq '.siliconflow.models | length' siliconflow.opencode.json
```

Inspect one model:

```bash
jq '.siliconflow.models["ByteDance-Seed/Seed-OSS-36B-Instruct"]' siliconflow.opencode.json
```

Validate that the API key is a placeholder, not a real secret:

```bash
jq -r '.siliconflow.apiKey' siliconflow.opencode.json
```

Expected output:

```text
${SILICONFLOW_API_KEY}
```

## Inspect the generated Crush config

After generation, inspect the top-level keys and model count with `jq`:

```bash
jq 'keys' siliconflow.crush.json
jq '.providers.siliconflow.models | length' siliconflow.crush.json
```

Inspect one model:

```bash
jq '.providers.siliconflow.models[] | select(.id == "ByteDance-Seed/Seed-OSS-36B-Instruct")' siliconflow.crush.json
```

Validate that the API key is a placeholder, not a real secret:

```bash
jq -r '.providers.siliconflow.api_key' siliconflow.crush.json
```

Expected output:

```text
$SILICONFLOW_API_KEY
```

## Inspect the generated Qwencode config

After generation, inspect the model count with `jq`:

```bash
jq '.openai | length' siliconflow.qwencode.json
```

Inspect one model:

```bash
jq '.openai[] | select(.id == "ByteDance-Seed/Seed-OSS-36B-Instruct")' siliconflow.qwencode.json
```

Validate that the API key is a placeholder, not a real secret:

```bash
jq -r '.openai[0].envKey' siliconflow.qwencode.json
```

Expected output:

```text
SILICONFLOW_API_KEY
```

## Troubleshooting

### `ERROR: SILICONFLOW_API_KEY is not set`

Export your SiliconFlow API key before running the tool:

```bash
export SILICONFLOW_API_KEY="your-siliconflow-api-key"
```

Then retry:

```bash
make gen-opencode
make gen-crush
make gen-qwencode
```

### API returns a non-200 status

The program will print the API status and response body. Common causes:

- missing or invalid API key
- expired API key
- rate limiting
- temporary SiliconFlow service issue

Check the raw response:

```bash
make raw-models
```

### No models found

The generator expects the SiliconFlow API response to contain a `data` array with model objects that include an `id` field.

If the API shape changes, update this part of `siliconflow-codegen.go`:

```go
type apiResponse struct {
	Data []apiModel `json:"data"`
}

type apiModel struct {
	ID string `json:"id"`
}
```

### Cross-compilation fails on macOS

The Makefile uses `CGO_ENABLED=0` for all release builds, which avoids most cross-compilation issues.

If you see linker errors, make sure you are not overriding the build command or forcing CGO on.

### Generated JSON is huge

SiliconFlow returns many models. That is expected. The generator includes every discovered model ID.

To inspect only the keys:

```bash
jq '.siliconflow.models | keys' siliconflow.opencode.json
jq '.providers.siliconflow.models[].id' siliconflow.crush.json
```

## Clean generated artifacts

Remove binaries and generated JSON files:

```bash
make clean
```

This removes:

```text
dist/
siliconflow.opencode.json
siliconflow.crush.json
siliconflow.models.json
```

## Development workflow

A typical workflow:

```bash
export SILICONFLOW_API_KEY="your-siliconflow-api-key"

make format
make test
make build
make gen-opencode
make gen-crush
make dist
```

Then inspect the generated configs:

```bash
jq '.siliconflow.models | length' siliconflow.opencode.json
jq '.providers.siliconflow.models | length' siliconflow.crush.json