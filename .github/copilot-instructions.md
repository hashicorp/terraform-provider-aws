# GitHub Copilot Instructions for terraform-provider-aws

## Build, Test, and Lint

```bash
# Build
make build

# Unit tests (full codebase)
make test

# Unit tests for a single service
make test PKG=s3

# Acceptance tests (require AWS credentials, TF_ACC=1)
make testacc PKG=s3 T=TestAccS3Bucket_basic

# Run all CI checks
make ci

# Quick fixes (formatting, imports, copyright, semgrep)
make quick-fix PKG=<service>

# Regenerate code from annotations
make gen

# Formatting only
make fmt

# Linting
make golangci-lint
make provider-lint
```

**Variable shorthands**: `PKG=` or `K=` for service name; `T=` or `TESTS=` for test pattern.

## Architecture

The provider is organized around 156+ AWS services, each in `internal/service/<service>/`. The top-level structure:

- **`internal/service/<service>/`** — All resource, data source, and sweep files for a given AWS service
- **`internal/provider/`** — Provider registration and configuration; resources auto-register via code generation
- **`internal/acctest/`** — Shared acceptance test helpers (`PreCheck`, config builders, etc.)
- **`internal/conns/`** — AWS SDK client initialization and connection management
- **`internal/flex/`** — Type conversion helpers between AWS SDK types and Terraform state
- **`internal/tags/`** — Shared tagging logic and generated tag resources
- **`internal/tfresource/`** — Retry/waiter utilities (`RetryWhenNotFound`, `WaitForDeletion`, etc.)
- **`names/`** — Service naming metadata (`names_data.hcl`) and generated constants; source of truth for service identifiers
- **`skaff/`** — Scaffolding CLI tool that generates resource/datasource/function boilerplate
- **`website/docs/r/`**, **`website/docs/d/`** — Provider documentation (`.html.markdown`)
- **`docs/`** — Contributor guides including `add-a-new-resource.md`, `naming.md`, `error-handling.md`, `skaff.md`

### Two Plugin Frameworks

New resources use **Terraform Plugin Framework** (preferred). Legacy resources use **Plugin SDK v2** (being migrated). Both coexist in the same service package.

## Key Conventions

### Creating New Resources

Always use `skaff` — do not copy existing resource files:

```bash
cd internal/service/<service>
skaff resource --name ResourceName       # Plugin Framework (preferred)
skaff datasource --name ResourceName
```

### Resource Registration

Resources self-register via annotations processed by `make gen`. Add the annotation comment above the constructor:

```go
// @FrameworkResource("aws_service_thing", name="Thing")
func newThingResource(_ context.Context) (resource.ResourceWithConfigure, error) { ... }

// @SDKResource("aws_service_thing", name="Thing")
func ResourceThing() *schema.Resource { ... }
```

After adding annotations, run `make gen` to update `service_package_gen.go`.

### File Naming

| Type | Pattern | Example |
|---|---|---|
| Resource | `<resource_name>.go` | `bucket.go` |
| Data source | `<resource_name>_data_source.go` | `bucket_data_source.go` |
| Test | `<resource_name>_test.go` | `bucket_test.go` |
| Generated identity test | `<resource_name>_identity_gen_test.go` | auto-generated |
| Resource docs | `website/docs/r/<service>_<resource>.html.markdown` | `s3_bucket.html.markdown` |
| Data source docs | `website/docs/d/<service>_<resource>.html.markdown` | `s3_bucket.html.markdown` |

### Naming Rules

- Terraform resource names: `aws_<serviceidentifier>_<resource_name>` (e.g., `aws_imagebuilder_image_pipeline`)
- Service identifier = AWS Go SDK v2 package name or AWS CLI command, whichever is shorter, lowercase, no underscores
- Main resource function: `ResourceThingName()` — no service name in function name
- Main data source function: `DataSourceThingName()`
- Attribute names in schema: `snake_case` (AWS API uses `CamelCase`)

### Copyright Header

Every `.go` file must begin with:

```go
// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0
```

Run `make copyright-fix` to add missing headers.

### Acceptance Test Structure

```go
func TestAccServiceThing_basic(t *testing.T) {
    ctx := acctest.Context(t)
    resourceName := "aws_service_thing.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { acctest.PreCheck(ctx, t) },
        ErrorCheck:               acctest.ErrorCheck(t, names.ServiceEndpointID),
        ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
        CheckDestroy:             testAccCheckThingDestroy(ctx),
        Steps: []resource.TestStep{
            {
                Config: testAccThingConfig_basic(rName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckThingExists(ctx, resourceName),
                ),
            },
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

Test config helpers are named `testAcc<Resource>Config_<scenario>` and defined at the bottom of the test file.

### EC2 Special Case

EC2, EBS, VPC, Transit Gateway, IPAM, VPN, and Wavelength resources all live in `internal/service/ec2/`. Use `PKG=ec2` (or aliases like `PKG=vpc`, `PKG=ebs`) — the Makefile normalizes them.

### Tags

Most resources support AWS tags. Use the `@Tags()` annotation and `tftags` helpers. Tag-only resources are generated via `make gen`. See `docs/resource-tagging.md`.

### Error Handling

Use helpers from `internal/tfresource` and `internal/errs`:

```go
if tfresource.NotFound(err) { ... }
if errs.IsA[*awstypes.ResourceNotFoundException](err) { ... }
```

See `docs/error-handling.md` for full guidance.
