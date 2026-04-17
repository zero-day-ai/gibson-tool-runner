# gibson-tool-runner

One microVM image. One Go binary. N parsers for CLI security + ops tools.

The Gibson daemon dispatches every sandboxed tool call into a Setec microVM
running this image. The binary reads a typed proto request from an env var,
shells out to the installed CLI, parses the output into taxonomy-aligned
`gibson.graphrag.DiscoveryResult` nodes, and emits the response via the
standard tool-runner ABI.

## Status

Early scaffold. v0.1.0 will ship the first three parsers (nmap, httpx,
nuclei). The daemon-side catalog refresher + removal of the Helm
`sandbox.tools.*` block lands as part of the `gibson-tool-runner` spec in
the main zero-day-ai repo.

## Tool-runner ABI

Input (env):
- `GIBSON_TOOL_NAME` — the registered parser name (e.g. `nmap`).
- `GIBSON_TOOL_INPUT_B64` — base64(protojson(request)).
- `GIBSON_TRACE_ID`, `GIBSON_SPAN_ID` — optional OpenTelemetry propagation.

Output (stdout):
- Arbitrary log lines may appear first.
- Last line prefixed with `===GIBSON_TOOL_OUTPUT===` followed by
  base64(protojson(response)).
- On error: `===GIBSON_TOOL_ERROR===<message>` then exit 2.

## Add a parser

1. Create a new directory under `parsers/<tool>/`.
2. Implement the `registry.Parser` interface in `<tool>.go`.
3. Register in its `init()`: `registry.Register(&myParser{})`.
4. Add the tool's apt/go/pip install step to the Dockerfile.
5. Add a `testdata/` directory with recorded tool output + a golden JSON file
   capturing the expected DiscoveryResult.
6. Add a blank import in `cmd/gibson-runner/main.go` so the parser's init runs.

Release flow: tag the runner image with semver + floating major. The Gibson
daemon's catalog refresher picks up the new tool set on its next tick (or
immediately via the `RefreshToolCatalog` admin RPC).

## Repo layout

```
cmd/gibson-runner/       Main binary (two modes: --list-tools, default execute)
internal/registry/       Parser interface + central registration map
parsers/<tool>/          One directory per tool: parser.go + testdata/
Dockerfile               debian:trixie-slim + curated tool set + binary
Makefile                 make bin | test | list-tools | image
```

## License

Apache 2.0.
