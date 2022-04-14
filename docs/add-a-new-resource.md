## Adding a New Resource

New resources are required when AWS adds a new service, or adds new features within an existing service which would require a new resource to manage in Terraform. Typically anything with a new set of CRUD API endpoints is a great candidate for a new resource.

Each resource should be submitted for review in isolation, pull requests containing multiple resources are harder to review and the maintainers will normally ask for them to be broken apart.

Please use the `skaff` tool to generate new resource and test templates for any new resource. Doing so will ensure that any boilerplate code, structural best practices and repetitive naming is done for you and always represents our most current standards. For more general guidance on items commonly looked for during the code review process please review [Common Review
Items](pullrequest-submission-and-lifecycle.md#common-review-items).

### Prequisites

If this is the first resource for a new service, please ensure the Service Client for the new service has been added and merged. See [Adding a new Service](add-a-new-service.md) for details.

### Steps to Add a Resource

#### Name the resource: 

All resources should be named with the following pattern: `aws_<service>_<name>`

Where `<service>` is the AWS short service name that matches the key in the `serviceData` map in the `conns` package (created via the [Adding a new Service](add-a-new-service.md) )

Where `<name>` represents the conceptual infrastructure represented by the create, read, update, and delete methods of the service API. It should be a singular noun. For example, in an API that has methods such as `CreateThing`, `DeleteThing`, `DescribeThing`, and `ModifyThing` the name of the resource would end in `_thing`.

#### Fill out the Resource Schema

In the `internal/service/<service>/<service>.go` file you will see a `Schema` property which exists as a map of `Schema` objects. This relates the AWS API data model with the Terraform resource itself. For each property you want to make available in Terraform, you will need to add it as an attribute, choose the correct data type and supply the correct [Schema Behaviors](https://www.terraform.io/plugin/sdkv2/schemas/schema-behaviors) in order to ensure Terraform knows how to correctly handle the value.

Typically you will add arguments to represent the values that are under control by Terraform, and attributes in order to supply read-only values as references for Terraform. These are distinguished by Schema Behavior.

Attribute names are to specified in `camel_case` as opposed to the AWS API which is `CamelCase`

#### Implement CRUD handlers
These will map planned Terraform state to the AWS API call, or an AWS API response to an applied Terrafrom state. You will also need to handle different response types (including errors correctly). For complex attributes you will need to implement Flattener or Expander functions. The [Data Handling and Conversion Guide](data-handling-and-conversion.md) covers everything you need to know for mapping AWS API responses to Terraform State and vice-versa. The [Error Handling Guide](error-handling.md) covers everything you need to know about handling AWS API responses consistently.

#### Write passing Acceptance Tests
In order to adequately test the resource we will need to write a complete set of Acceptance Tests. You will need an AWS account for this which allows the creation of that resource. See [Writing Acceptance Tests](#writing-acceptance-tests) below for a detailed guide on how to approach these.

    You will need at minimum:

    - Basic Test - Tests full lifecycle (CRUD + Import) of a minimal configuration (all required fields, no optional).
    - Dissapears Test - Tests what Terraform does if a resource it is tracking can no longer be found.
    - Per Attribute Tests - For each attribute a test should exists which tests that particular attribute in isolation alongside any required fields.

#### Create documentation for the resource. 

Add a file covering the use of the new resource in `website/docs/r/<service>_<name>.md`. You may want to also add examples of the resource in use particularly if its use is complex, or relies on resources in another service. This documentation will appear on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) when the resource is made available in a provider release. It is fine to link out to AWS Documentation where appropriate, particularly for values which are likely to change.

#### Ensure format and lint checks are passing

Run `go fmt` to format your code, and install and run all linters to detect and resolve any structural issues with the implementation or documentation.

    ```
    go fmt
    make tools     # install linters and dependencies
    make lint      # run provider linters
    make docs-lint # run documentation linters
    ```

#### Raise a Pull Request

Create a pull request against the default branch of the provider detailing the resource you are adding and linking it to the related issue (create one if one does not exist). If possible make sure that the "Allow edits from maintainers" box is checked, this will greatly improve the speed at which we can get the Pull Request merged. Ensure that all linters are passing in the GitHub PR Check section.

8. Wait for prioritization.

In general, pull requests are triaged within a few days of creation and are prioritized based on community reactions. Please view our Prioritization guide TODO for full details of the process.

