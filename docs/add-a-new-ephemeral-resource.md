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

Define attributes using `snake_case`, instead of the `CamelCase` format used by the AWS API.

### Implement Open Handler

`Open` will map the AWS API response to the ephemeral resource schema. You’ll also need to handle different response types, including errors, correctly. You will typically use `Autoflex` for mapping AWS API responses to Terraform models and vice versa. The [Error Handling Guide](error-handling.md) covers best practices for handling AWS API responses consistently.

### Register Ephemeral Resource to the provider

Ephemeral resources use a self-registration process that adds them to the provider via the `@EphemeralResource()` annotation in the resource's comments. To register the ephemeral resource, run `make gen`. This will generate an entry in the `service_package_gen.go` file located in the service package folder.

```go
package something

import (
	"context"
	
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @EphemeralResource("aws_something_example", name="Example")
func newEphemeralExample(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralExample{}, nil
}

type ephemeralExample struct {
	framework.EphemeralResourceWithConfigure
}
```

### Write Passing Acceptance Tests

To properly test the ephemeral resource, write a complete set of acceptance tests. An AWS account is required, which allows the provider to communicate with the AWS API. For a detailed guide on how to write and run these tests, refer to [Writing Acceptance Tests](running-and-writing-acceptance-tests.md).

You will need at a minimum:

- Basic Test - Tests full lifecycle (CRUD + Import) of a minimal configuration (all required fields, no optional).
- Per Attribute Tests - For each attribute a test should exist which tests that particular attribute in isolation alongside any required fields.

### Create Documentation for the Ephemeral Resource

Create a file documenting the use of the new ephemeral resource in `website/docs/ephemeral-resources/<service>_<name>.md` including a basic example. You may also want to include additional examples of the resource in use, especially if its usage is complex or depends on resources from another service. This documentation will appear on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) when the ephemeral resource is included in a provider release. It’s acceptable to link to AWS Documentation where appropriate, particularly for values that are likely to change.

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
