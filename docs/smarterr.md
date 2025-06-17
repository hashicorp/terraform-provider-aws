# smarterr

smarterr is config-driven error format management. Using these inputs:

1. HCL config files (i.e., `internal/smarterr.hcl` and `internal/service/cloudwatch/smarterr.hcl`)
2. Context (`context.Context`)
3. An error
4. A diagnostic
5. Arguments to the smarterr API functions
6. Call stacks (captured and live)

smarterr allows template-based formatting to output information in SDKv2 diagnostics, Framework diagnostics, and logs.

## smerr

`internal/smerr` is a wrapper for smarterr to inject private context (i.e., resource and service names) into arguments to smarterr functions.

## Configuration

Use `smarterr.hcl` files to configure smarterr using these configuration blocks: `token` (building blocks to use in `template`), `template` (Go `text/template` to format), `smarterr` (general configuration), `parameter` (static information), `stack_match` (expressions for matching the call stack), and `transform` (text transformations).

For more details, see [the documentation](https://github.com/YakDriver/smarterr/tree/main/docs).

## Migration Guide

Follow these steps to migrate SDKv2- and Framework-style resources and data sources to smarterr:
1. [Determine](#is-the-resource-or-data-source-sdkv2--or-framework-style) whether you're working with an SDKv2- or Framework-style resource or data source
2. If the resource or data source is SDKv2-style, follow the [SDKv2 steps](#steps-to-migrate-an-sdkv2-style-resource-or-data-source-to-smarterr)
3. If the resource or data source is Framework-style, follow the [Framework steps](#steps-to-migrate-a-framework-style-resource-or-data-source-to-smarterr)

### Is the resource or data source SDKv2- or Framework-style?

1. A **framework-style** resource/data source will have these characteristics:
    - Imports `"github.com/hashicorp/terraform-plugin-framework/resource"`
    - Has a `// @FrameworkResource(...)` or `// @FrameworkDataSource(...)` tag comment before the resource/data source function
    - Defines a resource or data source receiver with `Schema`, `Create`, `Read`, `Update`, and/or `Delete` methods
2. An **SDKv2-style** resource/data source will have these characteristics:
    - Imports `"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"`
    - Includes a `// @SDKResource(...)` tag comment before a function that returns the resource or data source

### Steps to Migrate an SDKv2-style resource or data source to smarterr

To migrate SDKv2 style resources or data sources to smarterr, follow these steps:

1. [Replace diagnostic appending functions with](#replace-diagnostic-appending-eg-sdkdiagappenderrorf) `smerr.Append()`
2. [Wrap Go errors](#wrap-errors-with-smarterr-errors) with `smarterr.NewError()`

#### Replace Diagnostic Appending (e.g., `sdkdiag.AppendErrorf()`)

Replace use of any diagnostic append function with `smerr.Append()`. This includes replacing these and any other similar calls:

- `sdkdiag.AppendErrorf` (from `"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"`)
- `sdkdiag.AppendFromErr`
- `create.AppendDiagError`

**Important** Use the same diagnostics variable (e.g., `diags`), ID (e.g., `name`, `d.Id()`, `id`), and error (e.g., `err`) as used in the original call.

Before:

```go
return sdkdiag.AppendErrorf(diags, "creating CloudWatch Metric Alarm (%s): %s", name, err)
```

After:

```go
return smerr.Append(ctx, diags, err, smerr.ID, name)
```

#### Wrap Go Errors with smarterr Errors

In places where errors are generated, passed, or returned as Go errors, wrap them with a smarterr `Error`. This often happens with find, wait, and status helpers.

Before:

```go
func findMetricAlarmByName(ctx context.Context, conn *cloudwatch.Client, name string) (*types.MetricAlarm, error) {
    // ...
	output, err := conn.DescribeAlarms(ctx, input)
	if err != nil {
		return nil, err // <===
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input) // <===
	}

	return tfresource.AssertSingleValueResult(output.MetricAlarms) // <===
}
```

After:

```go
func findMetricAlarmByName(ctx context.Context, conn *cloudwatch.Client, name string) (*types.MetricAlarm, error) {
    // ...
	output, err := conn.DescribeAlarms(ctx, input)
	if err != nil {
		return nil, smarterr.NewError(err) // <===
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(input)) // <===
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output.MetricAlarms)) // <===
}
```

### Steps to Migrate a Framework-style resource or data source to smarterr

To migrate Framework-style resources or data sources to smarterr, follow these steps:

1. [Replace `Diagnostics.AddError` with `smerr.AddError`](#replace-diagnosticsadderror-with-smerradderror)
2. [Replace `Diagnostics.Append` with `smerr.Append`](#replace-diagnosticsadderror-with-smerradderror)
2. [Wrap Go errors](#wrap-errors-with-smarterr-errors) with `smarterr.NewError()` (see above)

#### Replace `Diagnostics.AddError` with `smerr.AddError`

If code uses `Diagnostics.AddError` (`github.com/hashicorp/terraform-plugin-framework/diag.Diagnostics.AddError()`), replace with `smerr.AddError`.

**Important** Preserve these things from the original `Diagnostics.AddError()` call:

* Use a pointer to the same `Diagnostics` instance (e.g., `&resp.Diagnostics`)
* From the original `Diagnostics.AddError` call, use the same ID in the `smerr.AddError` call (e.g., `state.RuleName.String()`)
* Also, use the same error (e.g., `err`)

Everything else (e.g., `names.CloudWatch`, `create.ErrActionSetting`, `ResNameContributorInsightsRule`) can be discarded.

Before:

```go
resp.Diagnostics.AddError(
    create.ProblemStandardMessage(names.CloudWatch, create.ErrActionSetting, ResNameContributorInsightRule, state.RuleName.String(), err),
    err.Error(),
)
```

After:

```go
smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RuleName.String())
```

#### Replace `Diagnostics.Append` with `smerr.Append`

**Important** Take note of these points when migrating:

* Use a pointer to the same `Diagnostics` instance (e.g., `&resp.Diagnostics`)
* From the original `Diagnostics.Append` call, use the **same inner call** (e.g., `req.State.Get()`)
* _Add an ID_ if available in that part of the resource
    - When getting state at the top of a CRUD function, no ID may be available.
    - At the end of, e.g., Update or Read, an ID is available to use.

Before (No ID available):

```go
resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
```

After (No ID available):

```go
smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
```

Before (ID _available_):

```go
resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
```

After (ID _available_):

```go
smerr.EnrichAppend(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state), smerr.ID, state.RuleName.String())
```
