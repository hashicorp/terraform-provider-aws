# Adding a New Function

!!! tip
    Provider-defined function support is in technical preview and offered without compatibility promises until Terraform 1.8 is generally available.

Provider-defined functions were introduced with Terraform 1.8, enabling provider developers to expose functions specific to a given cloud provider or use case.
Functions in the AWS provider provide a utility that is valuable when paired with AWS resources.

See the Terraform Plugin Framework [Function documentation](https://developer.hashicorp.com/terraform/plugin/framework/functions) for additional details.

## Prerequisites

The only prerequisite for creating a function is ensuring the desired functionality is appropriate for a provider-defined function.
Functions must be reproducible across executions ("pure" functions), where the same input always results in the same output.
This requirement precludes the use of network calls, so operations requiring an AWS API call should instead consider utilizing a [data source](add-a-new-datasource.md).
Data manipulation tasks tend to be the most common use cases.

## Steps to add a function

### Fork the provider and create a feature branch

For a new function use a branch named `f-{function name}`, for example, `f-arn_parse`.
See [Raising a Pull Request](raising-a-pull-request.md) for more details.

### Create and name the function

The function name should be descriptive of its functionality and succinct.
Existing examples include `arn_parse` for decomposing ARN strings and `arn_build` for constructing them.

At this time [skaff](skaff.md) does not support the creation of new functions.
New functions can be created by copying the format of an existing function inside `internal/functions`.

### Fill out the function parameter(s) and return value

The function struct's `Definition` method will document the expected parameters and return value.
Parameter names and return values should be specified in `camel_case`.

```go
func (f exampleFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Parameters: []function.Parameter{
			function.StringParameter{Name: "some_arg"},
		},
		Return: function.StringReturn{},
	}
}
```

The example above defines a function which accepts a string parameter, `some_arg`, and returns a string value.

### Implement the function logic

The function struct's `Run` method will contain the function logic.
This includes processing the arguments, setting the return value, and any data processing that needs to happen in between.

```go
func (f exampleFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var data string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &data))
	if resp.Error != nil {
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, data))
}
```

### Register function to the provider

Once the function is implemented, it must be registered to the provider to be used.
As only Terraform Plugin Framework supports provider-defined functions, registration occurs on the Plugin Framework provider inside `internal/provider/fwprovider/provider.go`.
Add the `New*` factory function in the `Functions` method to register it.

```go
func (p *fwprovider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
            // Append to list of existing functions here
            tffunction.NewExampleFunction,
	}
}
```

### Write passing acceptance tests

All functions should have corresponding acceptance tests.
For functions with variadic arguments, or which can potentially return an error, tests should be written to exercise those conditions.

An example outline is included below:

```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestExampleFunction_basic(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testExampleFunctionConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "foo"),
				),
			},
		},
	})
}

func testExampleFunctionConfig() string {
	return `
output "test" {
  value = provider::aws::example("foo")
}`
}
```

With Terraform 1.8+ installed, individual tests can be run like:

```console
go test -run='^TestExample' -v ./internal/function/
```

### Create documentation for the resource

Add a file covering the use of the new function in `website/docs/functions/<function name>.html.markdown`.
This documentation will appear on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) when the function is made available in a provider release.

### Raise a pull request

See [Raising a Pull Request](raising-a-pull-request.md).

### Wait for prioritization

In general, pull requests are triaged within a few days of creation and are prioritized based on community reactions.
Please view our [Prioritization Guide](prioritization.md) for full details of the process.
