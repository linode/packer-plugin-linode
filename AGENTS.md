# AI Coding Agent Guidelines for packer-plugin-linode

## Project Overview

This is a HashiCorp Packer plugin for creating Linode images. It provides:
- **Builder** (`builder/linode/`): Creates Linode instances, provisions them, then snapshots to reusable images
- **Datasource** (`datasource/image/`): Queries existing Linode images for use in Packer templates

The plugin uses the [Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk) and [linodego](https://github.com/linode/linodego) client library.

## Architecture

### Build Flow
1. `main.go` registers the builder and datasource with the Packer plugin system
2. Builder executes as a series of **steps** (`step_*.go`) via `multistep.Runner`:
   - `StepCreateSSHKey` → `stepCreateLinode` → `stepCreateDiskConfig` → `StepConnect` → `StepProvision` → `stepShutdownLinode` → `stepCreateImage`
3. Configuration is defined in `config.go` with HCL2 specs auto-generated via `//go:generate packer-sdc`

### Instance Creation Modes
The builder supports two modes:

**Standard Mode** (with `image` specified):
- Linode API automatically creates disk and config from the image
- Instance starts booted and ready for SSH connection
- Used in most scenarios

**Custom Mode** (with `disk` and `config` blocks):
- `stepCreateLinode` creates unbooted instance without image
- `stepCreateDiskConfig` creates custom disks and configuration profiles
- Disk label → ID resolution: configs reference disks by label, which are resolved after disk creation
- At most one config may have `booted = true`; if none do, the first config is used as the boot configuration
- Instance is manually booted after configuration
- Enables fine-grained control over disk layout, kernel, helpers, and interfaces

### Key Patterns
- **Configuration structs** use `mapstructure` tags for HCL parsing and embed `helper.LinodeCommon` for shared auth config
- **Step pattern**: Each step implements `multistep.Step` interface with `Run()` and `Cleanup()` methods, storing state in `multistep.StateBag`
- **Flatten functions** in `step_create_linode.go` and `step_create_disk_config.go` convert config structs to linodego API types
- **Disk label resolution**: `resolveDiskLabel()` in `step_create_disk_config.go` maps user-provided disk labels in configs to actual disk IDs after creation
- **Two interface systems**: Legacy `interface` blocks and newer `linode_interface` blocks (see `linode_interfaces.go`)

### Adding a New Step
1. Create `step_<name>.go` implementing `multistep.Step` with `Run(ctx, state)` and `Cleanup(state)` methods
2. Retrieve config/ui from state: `c := state.Get("config").(*Config)`, `ui := state.Get("ui").(packersdk.Ui)`
3. Use `helper.ErrorHelper(state, ui, "prefix", err)` for error handling (returns `multistep.ActionHalt`)
4. Store results in state for subsequent steps: `state.Put("key", value)`
5. Add step to the `steps` slice in `builder.go` `Run()` method in correct order
6. Implement `Cleanup()` for resource teardown on failure (e.g., delete created resources)

## Developer Commands

```bash
make dev          # Build plugin and install to Packer plugins dir
make unit-test    # Run unit tests with race detection
make acctest      # Run acceptance tests (requires PACKER_ACC=1, LINODE_TOKEN)
make generate     # Regenerate HCL2 specs and documentation
make lint         # Run golangci-lint (install via `make deps`)
make format       # Format code with gofumpt
```

### Testing Requirements
- **Unit tests**: No external dependencies, run with `make unit-test`
- **Acceptance tests**: Require `LINODE_TOKEN` env var and `PACKER_ACC=1`, create real Linode resources
- Test files follow `*_test.go` naming; acceptance tests use `*_acc_test.go` suffix
- Debug with `PACKER_LOG=1` environment variable to see detailed plugin logs

## Code Generation

Files ending in `.hcl2spec.go` are **auto-generated** – do not edit manually. When modifying config structs:
1. Add/update `//go:generate packer-sdc struct-markdown` and `//go:generate packer-sdc mapstructure-to-hcl2` directives
2. Run `make generate` to regenerate specs and docs
3. Documentation partials in `docs-partials/` are generated from struct field comments

### Adding New Config Fields
1. Add field to struct in `config.go` or `linode_interfaces.go` with `mapstructure` tag
2. Add doc comment above field (becomes auto-generated documentation)
3. Mark required fields with `required:"true"` tag
4. Add validation in `Config.Prepare()` method if needed
5. Run `make generate` to update `.hcl2spec.go` and docs

## File Structure Reference

| Path | Purpose |
|------|---------|
| `builder/linode/config.go` | Builder configuration with validation |
| `builder/linode/step_*.go` | Build step implementations |
| `builder/linode/step_create_disk_config.go` | Custom disk and config profile creation step |
| `builder/linode/linode_interfaces.go` | Newer network interface configuration structs |
| `helper/common.go` | Shared config (LinodeCommon embedded struct) |
| `helper/client.go` | Linode API client initialization |
| `helper/error.go` | ErrorHelper for consistent step error handling |
| `datasource/image/data.go` | Image datasource implementation |
| `docs/` | MDX documentation source files |
| `example/` | Sample Packer templates (HCL and JSON) |
