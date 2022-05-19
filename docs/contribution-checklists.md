# Contribution Types and Checklists

There are several different kinds of contribution, each of which has its own
standards for a speedy review. The following sections describe guidelines for
each type of contribution.

- [Adding Resource Import Support](#adding-resource-import-support)

- [Adding Resource Policy Support](#adding-resource-policy-support)

## Adding Resource Import Support

Adding import support for Terraform resources will allow existing infrastructure to be managed within Terraform. This type of enhancement generally requires a small to moderate amount of code changes.

Comprehensive code examples and information about resource import support can be found in the [Extending Terraform documentation](https://www.terraform.io/docs/extend/resources/import.html).

In addition to the below checklist and the items noted in the Extending Terraform documentation, please see the [Common Review Items](pullrequest-submission-and-lifecycle.md#common-review-items) sections for more specific coding and testing guidelines.

- [ ] _Resource Code Implementation_: In the resource code (e.g., `internal/service/{service}/{thing}.go`), implementation of `Importer` `State` function
- [ ] _Resource Acceptance Testing Implementation_: In the resource acceptance testing (e.g., `internal/service/{service}/{thing}_test.go`), implementation of `TestStep`s with `ImportState: true`
- [ ] _Resource Documentation Implementation_: In the resource documentation (e.g., `website/docs/r/service_thing.html.markdown`), addition of `Import` documentation section at the bottom of the page



## Adding Resource Policy Support

Some AWS components support [resource-based IAM policies](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_identity-vs-resource.html) to control permissions. When implementing this support in the Terraform AWS Provider, we typically prefer creating a separate resource, `aws_{SERVICE}_{THING}_policy` (e.g., `aws_s3_bucket_policy`). See the [New Resource section](#new-resource) for more information about implementing the separate resource and the [Provider Design page](provider-design.md) for rationale.

