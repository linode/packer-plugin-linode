# AI Coding Agent Guidelines for packer-plugin-linode

## Project Overview
This is a HashiCorp Packer plugin for creating Linode images. It provides:
- **Builder** (`builder/linode/`): Creates Linode instances, provisions them, then snapshots to reusable images
- **Datasource** (`datasource/image/`): Queries existing Linode images for use in Packer templates

The plugin uses the [Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk) and [linodego](https://github.com/linode/linodego) client library.

## Build, Lint, and Test Commands
```bash
make dev          # Build plugin binary and install to Packer plugins dir
make build        # Build binary only (with format check)
make unit-test    # Run all unit tests with -race (builds first)
make acctest      # Run acceptance tests (requires PACKER_ACC=1, LINODE_TOKEN)
make lint         # Run golangci-lint (install via `make deps`)
make format       # Format all code with gofumpt
make generate     # Regenerate .hcl2spec.go files and documentation after codegen-input changes
make deps         # Install dev tools: golangci-lint, gofumpt, packer-sdc
```

### Running a Single Test
```bash
# Run a single test by name (regex match)
go test -race -count 1 -v -run TestFlattenDisk ./builder/linode/ -timeout=10m

# Run a single test file's tests by targeting the package
go test -race -count 1 -v -run 'TestResolveDiskLabel|TestFlattenInstanceConfig' ./builder/linode/

# Run a single acceptance test (requires real Linode credentials)
PACKER_ACC=1 go test -race -count 1 -v -run TestBuilderAcc_basic ./builder/linode/ -timeout=100m

# Run all tests in a specific package
go test -race -count 1 -v ./datasource/image/ -timeout=10m
```
### Testing Requirements
- **Unit tests**: No external dependencies, use `make unit-test` or `go test` directly
- **Acceptance tests**: Require `LINODE_TOKEN` env var and `PACKER_ACC=1`, create real Linode resources
- Test files: `*_test.go` for unit tests, `*_acc_test.go` for acceptance tests
- Debug with `PACKER_LOG=1` for detailed plugin logs

## Code Style Guidelines

### Formatting and Linting
- **Formatter**: `gofumpt` (stricter superset of `gofmt`); run `make format`
- **Linters** (via `.golangci.yml`): errcheck, govet, ineffassign, staticcheck, unconvert, unused
- **Import formatter**: `goimports` (enforced by golangci-lint)
- Files matching `*.hcl2spec.go` are excluded from linting (auto-generated)

### Import Ordering
Use `goimports` style with two groups separated by a blank line:
1. Standard library
2. All third-party and local packages together

Common package alias: `packersdk "github.com/hashicorp/packer-plugin-sdk/packer"`

### Naming Conventions
- **Exported types**: PascalCase — `Config`, `Builder`, `Artifact`, `Datasource`, `LinodeInterface`
- **Unexported step structs**: camelCase — `stepCreateLinode`, `stepShutdownLinode`, `stepCreateImage`
- **Flatten functions**: `flatten*` prefix for converting plugin types to linodego API types — `flattenDisk()`, `flattenInstanceConfig()`
- **Config struct tags**: `mapstructure:"field_name"` (snake_case), `required:"true"` for doc generation
- **State bag keys**: lowercase strings — `"config"`, `"instance"`, `"disk_label_to_id"`
- **Test variables**: `tt` for table-driven test cases, `want*` prefix for expected values (`wantID`, `wantError`, `wantNil`)
- **Short receivers**: single letter — `s` for steps, `c` for config, `a` for artifact, `d` for datasource

### Error Handling
- **In build steps** — define a local `handleError` closure: `handleError := func(prefix string, err error) multistep.StepAction { return helper.ErrorHelper(state, ui, prefix, err) }`
- **In config validation** — accumulate errors: `errs = packersdk.MultiErrorAppend(errs, errors.New("..."))`, return `errs` at the end
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` in helper/flatten functions. Use plain `errors.New()` for validation messages.

### Struct and Type Patterns
- **Config composition**: Embed shared types with `mapstructure:",squash"` tag
- **Optional fields**: Use pointer types (`*string`, `*int`, `*bool`) for truly optional config fields
- **Step structs**: Hold `*linodego.Client`, implement `multistep.Step` interface (`Run` + `Cleanup`)
- **Empty cleanup**: `func (s *stepName) Cleanup(state multistep.StateBag) {}`

### Test Patterns
- **No external assertion libraries** — use standard `testing` package only
- **Table-driven tests** are the dominant pattern:
```go
tests := []struct {
    name      string
    input     SomeType
    wantError bool
}{
    {name: "Valid input", input: validData, wantError: false},
    {name: "Missing field", input: badData, wantError: true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test body using tt
    })
}
```
- Use `t.Errorf()` for non-fatal failures, `t.Fatalf()` for fatal
- Use `reflect.DeepEqual` for struct comparison
- Use `strings.Contains(err.Error(), tt.errContains)` for error message validation
- Tests are in the **same package** (not `_test` suffix), allowing access to unexported symbols
- Acceptance tests use `acceptance.TestAccPreCheck(t)` guard and const HCL template strings

### Comments and Documentation
- Doc comments on all exported types and functions (these also generate plugin docs)
- Struct field comments become auto-generated documentation via `packer-sdc struct-markdown`; inline comments used sparingly

## Code Generation
Files ending in `.hcl2spec.go` are **auto-generated** — never edit manually. When modifying config structs:
1. Ensure `//go:generate packer-sdc struct-markdown` and `//go:generate packer-sdc mapstructure-to-hcl2 -type Type1,Type2` directives exist at file top
2. Run `make generate` to regenerate `.hcl2spec.go` files and docs
3. Documentation partials in `docs-partials/` are also generated from struct field comments

Treat `make generate` as required whenever you modify config structs, struct field comments, or any other codegen inputs. Before concluding any task that may affect generated artifacts, run it to keep generated source and docs synchronized with handwritten source.

## Architecture Quick Reference
`StepCreateSSHKey` → `stepCreateLinode` → `stepCreateDiskConfig` → `StepConnect` → `StepProvision` → `StepCleanupTempKeys` → `stepShutdownLinode` → `stepCreateImage`

### Key Files
| Path | Purpose |
|------|---------|
| `builder/linode/config.go` | Builder configuration with HCL2 parsing and validation |
| `builder/linode/builder.go` | Builder entry point, step pipeline assembly |
| `builder/linode/step_*.go` | Individual build step implementations |
| `builder/linode/linode_interfaces.go` | Network interface configuration structs |
| `helper/common.go` | Shared `LinodeCommon` config (auth, API URL) |
| `helper/client.go` | Linode API client initialization |
| `helper/error.go` | `ErrorHelper` for consistent step error handling |
| `datasource/image/data.go` | Image datasource implementation |
| `version/version.go` | Plugin version constants |

### Adding a New Step
1. Create `step_<name>.go` with an unexported struct holding `*linodego.Client`
2. Implement `Run(ctx context.Context, state multistep.StateBag) multistep.StepAction`
3. Implement `Cleanup(state multistep.StateBag)`
4. Retrieve config/ui from state: `c := state.Get("config").(*Config)`, `ui := state.Get("ui").(packersdk.Ui)`
5. Use `handleError` closure pattern (see Error Handling above)
6. Add step to the `steps` slice in `builder.go` `Run()` method

### Adding New Config Fields
1. Add field to struct in `config.go` or `linode_interfaces.go` with `mapstructure` tag
2. Add doc comment above field (used for generated documentation)
3. Mark required fields with `required:"true"` tag
4. Add validation in `Config.Prepare()` if needed
5. Run `make generate` to update `.hcl2spec.go` and docs

## Maintaining This File

**When your changes affect any of the following, update this AGENTS.md to stay in sync:**
- Build targets or commands in the `Makefile`
- Linter configuration in `.golangci.yml`
- New directories, key files, or architectural patterns
- Changes to the step pipeline in `builder.go`
- New conventions for naming, error handling, or testing
- Code generation workflow or `go:generate` directives
- Dependencies in `go.mod` (notably `replace` directives — see note below)

### go.mod Replace Directives
The project has `replace` directives in `go.mod` for compatibility fixes. Do not remove them:
- `github.com/ulikunitz/xz` — pinned to v0.5.15 for a release issue workaround
- `github.com/zclconf/go-cty` — replaced with a fork for packer-sdc compatibility
