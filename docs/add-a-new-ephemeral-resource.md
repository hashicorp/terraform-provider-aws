<!-- markdownlint-configure-file { "code-block-style": false } -->
# Adding a New Ephemeral Resource

New ephemeral resources are required when AWS introduces a new service or adds features to an existing service that necessitate a new ephemeral resource. Ephemeral resources produce ephemeral values and are never stored in the state. Any resource that produces a sensitive value can be an ephemeral resource, though some are more useful than others.

Each ephemeral resource should be submitted for review individually. Pull requests containing multiple ephemeral resources or other resources are more difficult to review, and maintainers will typically request that they be split into separate submissions.

## Prerequisites

If an ephemeral resource is the first addition for a new service, ensure that the Service Client for the service has been created and merged first. Refer to [Adding a New Service](add-a-new-service.md) for detailed instructions.

## Steps to Add a Ephemeral Resource

### Fork the Provider and Create a Feature Branch

For a new ephemeral resource, use a branch name in the format `f-{ephemeral-resource-name}`, for example: `f-kms-secret`. See [Raising a Pull Request](raising-a-pull-request.md) for more details.

### Create and Name the Ephemeral Resource

See the [Naming Guide](naming.md#resources-and-data-sources) for details on how to name the new ephemeral resource and the ephemeral resource file. Not following the naming standards will cause extra delay as maintainers request that you make changes.

Use the [skaff](skaff.md) provider scaffolding tool to generate new ephemeral resource and test templates using your chosen name. Doing so will ensure that any boilerplate code, structural best practices and repetitive naming are done for you and always represent our most current standards.

### Fill out the Ephemeral Resource Schema

In the `internal/service/<service>/<service>_ephemeral.go` file, you'll find a `Schema` property, which is a map of `Schema` objects. This maps the AWS API data model to the Terraform resource. To make a property available in Terraform, add it as an attribute with the appropriate data type.

Attribute names are to be specified in `snake_case` as opposed to the AWS API which is `CamelCase`.

### Implement Open Handler

These will map the AWS API response to the ephemeral resource schema. You will also need to handle different response types (including errors correctly). For complex attributes you will need to implement Flattener or Expander functions. The [Data Handling and Conversion Guide](data-handling-and-conversion.md) covers everything you need to know for mapping AWS API responses to Terraform State and vice-versa. The [Error Handling Guide](error-handling.md) covers everything you need to know about handling AWS API responses consistently.

### Register Ephemeral Resource to the provider

Ephemeral Resources use a self-registration process that adds them to the provider using the `@EphemeralResource()` annotation in the ephemeral resource's comments. Run `make gen` to register the ephemeral resource. This will add an entry to the `service_package_gen.go` file located in the service package folder.

```go
package something

import (
	"context"
	
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @EphemeralResource(name="Example")
func newEphemeralExample(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralExample{}, nil
}

type ephemeralExample struct {
	framework.EphemeralResourceWithConfigure
}

func (r *ephemeralExample) Metadata(_ context.Context, request ephemeral.MetadataRequest, response *ephemeral.MetadataResponse) {
	response.TypeName = "aws_something_example"
}
```

### Write Passing Acceptance Tests

To adequately test the ephemeral resource we will need to write a complete set of Acceptance Tests. You will need an AWS account for this which allows the provider to read to state of the associated resource. See [Writing Acceptance Tests](running-and-writing-acceptance-tests.md) for a detailed guide on how to approach these.

You will need at a minimum:

- Basic Test - Tests full lifecycle (CRUD + Import) of a minimal configuration (all required fields, no optional).
- Per Attribute Tests - For each attribute a test should exist which tests that particular attribute in isolation alongside any required fields.

### Create Documentation for the Ephemeral Resource

Add a file covering the use of the new ephemeral resource in `website/docs/ephemeral-resources/<service>_<name>.md`. You may want to also add examples of the ephemeral resource in use particularly if its use is complex, or relies on resources in another service. This documentation will appear on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) when the ephemeral resource is made available in a provider release. It is fine to link out to AWS Documentation where appropriate, particularly for values which are likely to change.

### Ensure Format and Lint Checks are Passing Locally

Run `go fmt` to format your code, and install and run all linters to detect and resolve any structural issues with the implementation or documentation.

```sh
make fmt
make tools        # install linters and dependencies
make lint         # run provider linters
make docs-lint    # run documentation linters
make website-lint # run website documentation linters
```

### Raise a Pull Request

See [Raising a Pull Request](raising-a-pull-request.md).

### Wait for Prioritization

In general, pull requests are triaged within a few days of creation and are prioritized based on community reactions. Please view our [prioritization](prioritization.md) guide for full details of the process.
