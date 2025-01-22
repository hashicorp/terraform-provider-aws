# Adding a New Function

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

### Generate function scaffolding

The `skaff function` subcommand can be used to generate an outline for the new function.

First, install `skaff` and navigate to the directory where provider functions are defined (`internal/function`).

```console
make skaff
```

```console
cd internal/function
```

Next, run the `skaff function` subcommand.
The name and description flags are required.
The name argument should be [mixed caps](naming.md#mixed-caps) (ie. `FooBar`), and the generator will handle converting the name to snake case where appropriate.

```console
skaff function -n Example -d "Makes some output from an input."
```

This will generate files storing the function definition, unit tests, and registry documentation.
The following steps describe how to complete the function implementation.

### Fill out the function parameter(s) and return value

The function struct's `Definition` method will document the expected parameters and return value.
Parameter names and return values should be specified in `snake_case`.

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

	//
	// Function logic goes here
	//

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
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestExampleFunction_basic(t *testing.T) {
	t.Parallel()

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(version.Must(version.NewVersion("1.8.0"))),
		},
		Steps: []resource.TestStep{
			{
				Config: testExampleFunctionConfig("foo"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("test", "foo"),
				),
			},
		},
	})
}

func testExampleFunctionConfig(arg string) string {
	return fmt.Sprintf(`
output "test" {
  value = provider::aws::example(%[1]q)
}`, arg)
}
```

With Terraform 1.8+ installed, individual tests can be run like:

```console
go test -run='^TestExampleFunction' -v ./internal/function/
```

### Create documentation for the resource

`skaff` will have generated framed out registry documentation in `website/docs/functions/<function name>.html.markdown`.
The `Example Usage`, `Signature`, and `Arguments` sections should all be updated with the appropriate content.
Once released, this documentation will appear on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest).

### Raise a pull request

See [Raising a Pull Request](raising-a-pull-request.md).

### Wait for prioritization

In general, pull requests are triaged within a few days of creation and are prioritized based on community reactions.
Please view our [Prioritization Guide](prioritization.md) for full details of the process.
