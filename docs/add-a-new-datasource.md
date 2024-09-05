<!-- markdownlint-configure-file { "code-block-style": false } -->
# Adding a New Data Source

New data sources are required when AWS adds a new service, or adds new features within an existing service which would require a new data source to allow practitioners to query existing resources of that type for use in their configurations. Anything with a Describe or Get endpoint could make a data source, but some are more useful than others.

Each data source should be submitted for review in isolation, pull requests containing multiple data sources and/or resources are harder to review and the maintainers will normally ask for them to be broken apart.

## Prerequisites

If this is the first addition of a data source for a new service, please ensure the Service Client for the new service has been added and merged. See [Adding a new Service](add-a-new-service.md) for details.

Determine which version of the AWS SDK for Go the resource will be built upon. For more information and instructions on how to determine this choice, please read [AWS SDK for Go Versions](aws-go-sdk-versions.md)

## Steps to Add a Data Source

### Fork the Provider and Create a Feature Branch

For a new data source use a branch named `f-{datasource name}` for example: `f-ec2-vpc`. See [Raising a Pull Request](raising-a-pull-request.md) for more details.

### Create and Name the Data Source

See the [Naming Guide](naming.md#resources-and-data-sources) for details on how to name the new resource and the resource file. Not following the naming standards will cause extra delay as maintainers request that you make changes.

Use the [skaff](skaff.md) provider scaffolding tool to generate new resource and test templates using your chosen name ensuring you provide the `v1` flag if you are targeting version 1 of the `aws-go-sdk`. Doing so will ensure that any boilerplate code, structural best practices and repetitive naming are done for you and always represent our most current standards.

### Fill out the Data Source Schema

In the `internal/service/<service>/<service>_data_source.go` file you will see a `Schema` property which exists as a map of `Schema` objects. This relates the AWS API data model with the Terraform resource itself. For each property you want to make available in Terraform, you will need to add it as an attribute, and choose the correct data type.

Attribute names are to be specified in `snake_case` as opposed to the AWS API which is `CamelCase`.

### Implement Read Handler

These will map the AWS API response to the data source schema. You will also need to handle different response types (including errors correctly). For complex attributes you will need to implement Flattener or Expander functions. The [Data Handling and Conversion Guide](data-handling-and-conversion.md) covers everything you need to know for mapping AWS API responses to Terraform State and vice-versa. The [Error Handling Guide](error-handling.md) covers everything you need to know about handling AWS API responses consistently.

### Register Data Source to the provider

Data Sources use a self-registration process that adds them to the provider using the `@SDKDataSource()` annotation in the data source's comments. Run `make gen` to register the data source. This will add an entry to the `service_package_gen.go` file located in the service package folder.

=== "Terraform Plugin Framework (Preferred)"

    ```go
    package something

    import (
        "github.com/hashicorp/terraform-plugin-framework/datasource"
        "github.com/hashicorp/terraform-provider-aws/internal/framework"
    )

    // @FrameworkDataSource(name="Example")
    func newResourceExample(_ context.Context) (datasource.ResourceWithConfigure, error) {
    	return &dataSourceExample{}, nil
    }

    type dataSourceExample struct {
	    framework.DataSourceWithConfigure
    }

    func (r *dataSourceExample) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
    	response.TypeName = "aws_something_example"
    }
    ```

=== "Terraform Plugin SDK V2"

    ```go
    package something

    import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

    // @SDKDataSource("aws_something_example", name="Example")
    func DataSourceExample() *schema.Resource {
    	return &schema.Resource{
    	    // some configuration
    	}
    }
    ```

### Write Passing Acceptance Tests

To adequately test the data source we will need to write a complete set of Acceptance Tests. You will need an AWS account for this which allows the provider to read to state of the associated resource. See [Writing Acceptance Tests](running-and-writing-acceptance-tests.md) for a detailed guide on how to approach these.

You will need at a minimum:

- Basic Test - Tests full lifecycle (CRUD + Import) of a minimal configuration (all required fields, no optional).
- Disappears Test - Tests what Terraform does if a resource it is tracking can no longer be found.
- Per Attribute Tests - For each attribute a test should exist which tests that particular attribute in isolation alongside any required fields.

### Create Documentation for the Data Source

Add a file covering the use of the new data source in `website/docs/d/<service>_<name>.md`. You may want to also add examples of the data source in use particularly if its use is complex, or relies on resources in another service. This documentation will appear on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) when the data source is made available in a provider release. It is fine to link out to AWS Documentation where appropriate, particularly for values which are likely to change.

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
