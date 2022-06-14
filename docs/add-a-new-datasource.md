New datasources are required when AWS adds a new service, or adds new features within an existing service which would require a new datasource to allow practitioners to query existing resources of that type for use in their configurations. Anything with a Describe or Get endpoint could make a datasource, but some are more useful than others.

Each datasource should be submitted for review in isolation, pull requests containing multiple datasources and/or resources are harder to review and the maintainers will normally ask for them to be broken apart.

Please use the `skaff` tool to generate new datasource and test templates for any new resource. Doing so will ensure that any boilerplate code, structural best practices and repetitive naming is done for you and always represents our most current standards.

### Prerequisites

If this is the first addition of a resource or datasource for a new service, please ensure the Service Client for the new service has been added and merged. See [Adding a new Service](add-a-new-service.md) for details.

### Steps to Add a Datasource

#### Fork the provider and create a feature branch

For a new resources use a branch named `f-{datasource name}` for example: `f-ec2-vpc`. See [Raising a Pull Request](raising-a-pull-request.md) for more details.

#### Name the datasource

Either by creating the file manually, or using `skaff` to generate a template.

All resources should be named with the following pattern: `aws_<service>_<name>`

Where `<service>` is the AWS short service name that matches the key in the `serviceData` map in the `conns` package (created via the [Adding a new Service](add-a-new-service.md) )

Where `<name>` represents the conceptual infrastructure represented by the create, read, update, and delete methods of the service API. It should be a singular noun. For example, in an API that has methods such as `CreateThing`, `DeleteThing`, `DescribeThing`, and `ModifyThing` the name of the resource would end in `_thing`.

#### Fill out the Datasource Schema

In the `internal/service/<service>/<service>.go` file you will see a `Schema` property which exists as a map of `Schema` objects. This relates the AWS API data model with the Terraform resource itself. For each property you want to make available in Terraform, you will need to add it as an attribute, and choose the correct data type.

Attribute names are to specified in `camel_case` as opposed to the AWS API which is `CamelCase`

#### Implement Read Handler

These will map the AWS API response to the datasource schema. You will also need to handle different response types (including errors correctly). For complex attributes you will need to implement Flattener or Expander functions. The [Data Handling and Conversion Guide](data-handling-and-conversion.md) covers everything you need to know for mapping AWS API responses to Terraform State and vice-versa. The [Error Handling Guide](error-handling.md) covers everything you need to know about handling AWS API responses consistently.

#### Write passing Acceptance Tests
In order to adequately test the resource we will need to write a complete set of Acceptance Tests. You will need an AWS account for this which allows the creation of that resource. See [Writing Acceptance Tests](#writing-acceptance-tests) below for a detailed guide on how to approach these.

You will need at minimum:

- Basic Test - Tests full lifecycle (CRUD + Import) of a minimal configuration (all required fields, no optional).
- Disappears Test - Tests what Terraform does if a resource it is tracking can no longer be found.
- Per Attribute Tests - For each attribute a test should exists which tests that particular attribute in isolation alongside any required fields.

#### Create documentation for the resource

Add a file covering the use of the new resource in `website/docs/r/<service>_<name>.md`. You may want to also add examples of the resource in use particularly if its use is complex, or relies on resources in another service. This documentation will appear on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) when the resource is made available in a provider release. It is fine to link out to AWS Documentation where appropriate, particularly for values which are likely to change.

#### Ensure format and lint checks are passing locally

Run `go fmt` to format your code, and install and run all linters to detect and resolve any structural issues with the implementation or documentation.

```bash
go fmt
make tools     # install linters and dependencies
make lint      # run provider linters
make docs-lint # run documentation linters
```

#### Raise a Pull Request

See [Raising a Pull Request](raising-a-pull-request.md).

#### Wait for prioritization

In general, pull requests are triaged within a few days of creation and are prioritized based on community reactions. Please view our [prioritization](prioritization.md) guide for full details of the process.

