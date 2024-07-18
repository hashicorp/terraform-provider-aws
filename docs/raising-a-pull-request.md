# Raising a Pull Request

1. [Fork the GitHub repository](https://help.github.com/en/articles/fork-a-repo) allowing you to make the changes in your own copy of the repository.

1. Create a branch using the following naming prefixes:

    - f = feature
    - b = bug fix
    - d = documentation
    - t = tests
    - td = technical debt
    - v = dependencies ("vendoring" previously)

    Some indicative example branch names would be `f-aws_emr_instance_group-refactor` or `td-staticcheck-st1008`

1. Make the changes you would like to include in the provider, add new tests as required, and make sure that all relevant existing tests are passing.

1. [Create a pull request](https://help.github.com/en/articles/creating-a-pull-request-from-a-fork). Please ensure (if possible) that the 'Allow edits from maintainers' checkbox is checked. This will allow the maintainers to make changes and merge the PR without requiring action from the contributor.
   You are welcome to submit your pull request for commentary or review before
   it is fully completed by creating a [draft pull request](https://help.github.com/en/articles/about-pull-requests#draft-pull-requests).
   Please include specific questions or items you'd like feedback on.

1. Create a changelog entry following the process outlined [here](changelog-process.md)

1. Once you believe your pull request is ready to be reviewed, ensure the
   pull request is not a draft pull request by [marking it ready for review](https://help.github.com/en/articles/changing-the-stage-of-a-pull-request)
   or removing `[WIP]` from the pull request title if necessary, and a
   maintainer will review it. Follow [the checklists below](#resource-contribution-guidelines)
   to help ensure that your contribution can be easily reviewed and potentially
   merged.

1. One of Terraform's provider team members will look over your contribution and
   either approve it or provide comments letting you know if there is anything
   left to do. We'll try to give you the opportunity to make the required changes yourself, but in some cases, we may perform the changes ourselves if it makes sense to (minor changes, or for urgent issues).  We do our best to keep up with the volume of PRs waiting for review, but it may take some time depending on the complexity of the work.

1. Once all outstanding comments and checklist items have been addressed, your
   contribution will be merged! Merged PRs will be included in the next
   Terraform release.

1. In some cases, we might decide that a PR should be closed without merging.
   We'll make sure to provide clear reasoning when this happens.

### Go Coding Style

All Go code is automatically checked for compliance with various linters, such as `gofmt`. These tools can be installed using the `GNUMakefile` in this repository.

```console
make tools
```

Check your code with the linters:

```console
make lint
```

We use [Semgrep](https://semgrep.dev/docs/) to check for other code standards.
This can be run directly on the command line, i.e.,

```console
semgrep
```

or it can be run using Docker via the Makefile, i.e.,

```console
make semgrep
```

`gofmt` will also fix many simple formatting issues for you. The Makefile includes a target for this:

```console
make fmt
```

The import statement in a Go file follows these rules (see [#15903](https://github.com/hashicorp/terraform-provider-aws/issues/15903)):

1. Import declarations are grouped into a maximum of three groups in the following order:
    - Standard packages (also called short import path or built-in packages)
    - Third-party packages (also called long import path packages)
    - Local packages
1. Groups are separated by a single blank line
1. Packages within each group are alphabetized

Check your imports:

```console
make import-lint
```

For greater detail, the following Go language resources provide common coding preferences that may be referenced during review, if not automatically handled by the project's linting tools.

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

### Resource Contribution Guidelines

The following resource checks need to be addressed before your contribution can be merged. The exclusion of any applicable check may result in a delayed time to merge. Some of these are not handled by the automated code testing that occurs during submission, so reviewers (even those outside the maintainers) are encouraged to reach out to contributors about any issues to save time.

This Contribution Guide also includes separate sections on topics such as [Error Handling](error-handling.md), which also applies to contributions.

- __Passes Testing__: All code and documentation changes must pass unit testing, code linting, and website link testing. Resource code changes must pass all acceptance testing for the resource.
- __Avoids API Calls Across Account, Region, and Service Boundaries__: Resources should not implement cross-account, cross-region, or cross-service API calls.
- __Does Not Set Optional or Required for Non-Configurable Attributes__: Resource schema definitions for read-only attributes must not include `Optional: true` or `Required: true`.
- __Avoids retry.RetryContext() without retry.RetryableError()__: Resource logic should only implement [`retry.Retry()`](https://godoc.org/github.com/hashicorp/terraform/helper/retry#Retry) if there is a retryable condition (e.g., `return retry.RetryableError(err)`).
- __Avoids Reusing Resource Read Function in Data Source Read Function__: Data sources should fully implement their own resource `Read` functionality.
- __Avoids Reading Schema Structure in Resource Code__: The resource `Schema` should not be read in resource `Create`/`Read`/`Update`/`Delete` functions to perform looping or otherwise complex attribute logic. Use [`d.Get()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Get) and [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) directly with individual attributes instead.
- __Avoids ResourceData.GetOkExists()__: Resource logic should avoid using [`ResourceData.GetOkExists()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.GetOkExists) as its expected functionality is not guaranteed in all scenarios.
- __Calls Read After Create and Update__: Except where API eventual consistency prohibits immediate reading of resources or updated attributes,  resource `Create` and `Update` functions should return the resource `Read` function.
- __Implements Immediate Resource ID Set During Create__: Immediately after calling the API creation function, the resource ID should be set with [`d.SetId()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.SetId) before other API operations or returning the `Read` function.
- __Implements Attribute Refreshes During Read__: All attributes available in the API should have [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) called their values in the Terraform state during the `Read` function.
- __Performs Error Checks with Non-Primitive Attribute Refreshes__: When using [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) with non-primitive types (`schema.TypeList`, `schema.TypeSet`, or `schema.TypeMap`), perform error checking to [prevent issues where the code is not properly able to refresh the Terraform state](https://www.terraform.io/plugin/sdkv2/best-practices/detecting-drift#error-checking-aggregate-types).
- __Implements Import Acceptance Testing and Documentation__: Support for resource import (`Importer` in resource schema) must include `ImportState` acceptance testing (see also the [Acceptance Testing Guidelines](running-and-writing-acceptance-tests.md)) and `## Import` section in resource documentation.
- __Implements Customizable Timeouts Documentation__: Support for customizable timeouts (`Timeouts` in resource schema) must include `## Timeouts` section in resource documentation.
- __Implements State Migration When Adding New Virtual Attribute__: For new "virtual" attributes (those only in Terraform and not in the API), the schema should implement [State Migration](https://www.terraform.io/plugin/sdkv2/resources#state-migrations) to prevent differences for existing configurations that upgrade.
- __Uses AWS Go SDK Constants__: Many AWS services provide string constants for value enumerations, error codes, and status types. See also the "Constants" sections under each of the service packages in the [AWS Go SDK documentation](https://docs.aws.amazon.com/sdk-for-go/api/).
- __Uses AWS Go SDK Pointer Conversion Functions__: Many APIs return pointer types and these functions return the zero value for the type if the pointer is `nil`. This prevents potential panics from unchecked `*` pointer dereferences and can eliminate boilerplate `nil` checking in many cases. See also the [`aws` package in the AWS Go SDK documentation](https://docs.aws.amazon.com/sdk-for-go/api/aws/).
- __Uses AWS Go SDK Types__: Use available SDK structs instead of implementing custom types with indirection.
- __Uses Existing Validation Functions__: Schema definitions including `ValidateFunc` for attribute validation should use available [Terraform `helper/validation` package](https://godoc.org/github.com/hashicorp/terraform/helper/validation) functions. `All()`/`Any()` can be used for combining multiple validation function behaviors.
- __Uses tfresource.TimedOut() with retry.Retry()__: Resource logic implementing [`retry.Retry()`](https://godoc.org/github.com/hashicorp/terraform/helper/retry#Retry) should error check with [`tfresource.TimedOut(err error)`](https://godoc.org/github.com/hashicorp/terraform-provider-aws/internal/tfresource#TimedOut) and potentially unset the error before returning the error. For example:

  ```go
  var output *kms.CreateKeyOutput
  err := retry.Retry(1*time.Minute, func() *retry.RetryError {
    var err error

    output, err = conn.CreateKey(input)

    /* ... */

    return nil
  })

  if tfresource.TimedOut(err) {
    output, err = conn.CreateKey(input)
  }

  if err != nil {
    return fmt.Errorf("creating KMS External Key: %s", err)
  }
  ```

- __Uses id.UniqueId()__: API fields for concurrency protection such as `CallerReference` and `IdempotencyToken` should use [`id.UniqueId()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#UniqueId). The implementation includes a monotonic counter which is safer for concurrent operations than solutions such as `time.Now()`.
- __Skips id Attribute__: The `id` attribute is implicit for all Terraform resources and does not need to be defined in the schema.

The below are style-based items that _may_ be noted during review and are recommended for simplicity, consistency, and quality assurance:

- __Implements arn Attribute__: APIs that return an ARN should implement `arn` as an attribute. Alternatively, the ARN can be synthesized using the AWS Go SDK [`arn.ARN`](https://docs.aws.amazon.com/sdk-for-go/api/aws/arn/#ARN) structure. For example:

  ```go
  // Direct Connect Virtual Interface ARN.
  // See https://docs.aws.amazon.com/directconnect/latest/UserGuide/security_iam_service-with-iam.html#security_iam_service-with-iam-id-based-policies-resources.
  arn := arn.ARN{
      Partition: meta.(*conns.AWSClient).Partition,
      Region:    meta.(*conns.AWSClient).Region,
      Service:   "directconnect",
      AccountID: meta.(*conns.AWSClient).AccountID,
      Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
  }.String()
  d.Set("arn", arn)
  ```

  When the `arn` attribute is synthesized this way, add the resource to the [list](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id) of those affected by the provider's `skip_requesting_account_id` attribute.

- __Implements Warning Logging With Resource State Removal__: If a resource is removed outside of Terraform (e.g., via a different tool, API, or web UI), `d.SetId("")` and `return nil` can be used in the resource `Read` function to trigger resource recreation. When this occurs, a warning log message should be printed beforehand: `log.Printf("[WARN] {SERVICE} {THING} (%s) not found, removing from state", d.Id())`
- __Uses American English for Attribute Naming__: For any ambiguity with attribute naming, prefer American English over British English. e.g., `color` without the British `u`.
- __Skips Timestamp Attributes__: Generally, creation and modification dates from the API should be omitted from the schema.
- __Uses Paginated AWS Go SDK Functions When Iterating Over a Collection of Objects__: When the API for listing a collection of objects provides a paginated function, use it instead of looping until the next page token is not set. For example, with the EC2 API, [`DescribeInstancesPages`](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeInstancesPages) should be used instead of [`DescribeInstances`](https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeInstances) when more than one result is expected.
- __Adds Paginated Functions Missing from the AWS Go SDK to Internal Service Package__: If the AWS Go SDK does not define a paginated equivalent for a function to list a collection of objects, it should be added to a per-service internal package using the [`listpages` generator](https://github.com/hashicorp/terraform-provider-aws/blob/main/internal/generate/listpages/README.md). A support case should also be opened with AWS to have the paginated functions added to the AWS Go SDK.
