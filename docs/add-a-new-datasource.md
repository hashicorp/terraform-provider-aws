<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

<!-- markdownlint-configure-file { "code-block-style": false } -->
# Adding a New Data Source Type

New data sources are required when AWS adds a new service, or adds new features within an existing service which would require practitioners to query existing resources of that type for use in their configurations.
Anything with a `Describe` or `Get` endpoint could make a data source, but some are more useful than others.

Each data source should be submitted for review in isolation.
Pull requests containing multiple data sources and/or resources are harder to review and the maintainers will normally ask for them to be broken apart.

## Prerequisites

If this is the first addition of a data source for a new service, please ensure the Service Client for the new service has been added and merged.
See [Adding a new Service](add-a-new-service.md) for details.

## Steps to Add a Data Source

### Fork the Provider and Create a Feature Branch

For a new data source use a branch named `f-{datasource name}` for example: `f-vpc_endpoint`.
See [Raising a Pull Request](raising-a-pull-request.md) for more details.

### Create and Name the Data Source

See the [Naming Guide](naming.md#resources-and-data-sources) for details on how to name the new data source and the data source file.

Use the [skaff](skaff.md) tool to generate new data source and test templates.
Doing so will ensure that any boilerplate code, structural best practices and repetitive naming are done for you and always represent our most current standards.

### Fill out the Schema

In `internal/service/<service>/<name>_data_source.go`, the `Schema` method defines the `Attributes` and `Blocks` for the data source.
The Schema maps the Terraform data model to the AWS API and should match exactly in most cases.

For each API property add a corresponding attribute or block and choose the correct data type.
For objects with only `Computed` attributes, always use the [`framework.DataSourceComputedListOfObjectAttribute`](https://github.com/hashicorp/terraform-provider-aws/blob/v6.37.0/internal/framework/data_source_list_of_object.go#L14-L24) helper. [^1]

Attribute names are specified in `snake_case` as opposed to the AWS API which is `CamelCase`.
A corresponding "model" struct, named `<name>DataSourceModel` by `skaff`, should match the Schema definition.

[^1]: Fully computed blocks are not supported by Terraform protocol V6, which the AWS provider will adopt in a future major version. See [this issue](https://github.com/hashicorp/terraform-provider-aws/issues/45338) for additional details.

### Implement the Read Handler

The `Read` method converts the AWS API response to the data source model.
You will also need to handle different response types (and errors).

In most cases [AutoFlex](data-handling-and-conversion.md#recommended-implementations) will convert between Terraform and AWS data types with no custom handling required.

The [Data Handling and Conversion Guide](data-handling-and-conversion.md) covers everything you need to know for mapping AWS API responses to Terraform State and vice-versa. The [Error Handling Guide](error-handling.md) covers everything you need to know about handling AWS API responses consistently.

### Register the Data Source

Data Sources use a self-registration process that adds them to the provider using the `@FrameworkDataSource()` annotation in the data source's comments.
Run `go generate ./internal/service/<service>` to register the data source.
This will add an entry to the `service_package_gen.go` file located in the service package folder.

```go
package something

import (
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @FrameworkDataSource("aws_something_example", name="Example")
func newExampleDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
  return &exampleDataSource{}, nil
}

type exampleDataSource struct {
  framework.DataSourceWithModel[exampleDataSourceModel]
}

type exampleDataSourceModel {
  // Fields corresponding to attributes in the Schema.
}
```

### Write Passing Acceptance Tests

To adequately test the data source, include a complete set of Acceptance Tests.
You will need an AWS account which allows the provider to read the associated resource.
See [Writing Acceptance Tests](running-and-writing-acceptance-tests.md) for a detailed guide on how to approach these.

At a minimum the data source should include a "Basic" test with a minimal configuration (all required fields, no optional).
If additional optional arguments are supported (e.g. filters), a test should be added to verify each.

### Fill Out the Documentation

`skaff` will generate documentation for the new data source in `website/docs/d/<service>_<name>.md`.
If the data source is particularly complex or relies on resources in another service, additional examples may be added.
The argument and attribute references should match the Schema definition.
It is fine to link out to AWS Documentation where appropriate, particularly for values which are likely to change.

This documentation will appear on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) when the data source is made available in a provider release.

### Ensure Format and Lint Checks are Passing Locally

Run `make fmt` to format your code, and install and run all linters to detect and resolve any structural issues with the implementation or documentation.

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

In general, pull requests are triaged within a few days of creation and are prioritized based on community reactions.
Please view our [prioritization](prioritization.md) guide for full details of the process.
