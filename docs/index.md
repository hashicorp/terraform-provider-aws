# Welcome

The Terraform AWS Provider is the work of thousands of contributors, and is maintained by a small team within HashiCorp. This site contains extensive instructions about how to contribute and how the AWS provider works.

_Please Note: This documentation is intended for Terraform AWS Provider code developers. Typical operators writing and applying Terraform configurations do not need to read or understand this material._

## Contribute

Please follow the following steps to ensure your contribution goes smoothly.

### 1. Configure Development Environment

Install Terraform and Go. Clone the repository, compile the provider, and set up testing. Refer to [Configure Development Environment](development-environment.md).

### 2. Write Code

Follow the guide for your contribution type and refer to the Development Reference materials as needed for additional details about [provider design](provider-design.md), expected [naming conventions](naming.md), guidance for [error handling](error-handling.md), etc.

| Contribution Guide | Description |
|--------------------|-------------|
| [Resources](add-a-new-resource.md) | Allow the management of a logical resource within AWS by adding a new resource to the Terraform AWS Provider. |
| [Data Source](add-a-new-datasource.md) | Let your Terraform configurations use data from resources not under local management by creating ready only data sources. |
| [Services](add-a-new-service.md) | Allow Terraform (via the AWS Provider) to manage an entirely new AWS service by introducing the resources and data sources required to manage configuration of the service. |
| [AWS Region](add-a-new-region.md) | New regions are immediately usable with the provider with the caveat that a configuration workaround is required to skip validation of the region during cli operations. A small set of changes are required to makes this workaround necessary. |
| [Bug Fix or Enhancement](bugs-and-enhancements.md) | These constitute the majority of pull requests submitted, many of which we address and merge regardless of priority in our regular internal gardening days. |
| [Resource Name Generation](resource-name-generation.md) | Allow a resource to either fully, or partially, generate its own resource names. This can be useful in cases where the resource name uniquely identifes the resource and it needs to be recreated. It can also be used when a name is required, but the specific name is not important. |
| [Tagging Support](resource-tagging.md) | Many AWS resources allow assigning metadata via tags. However, frequently AWS services are launched without tagging support so this will often need to be added later. |
| [Import Support](add-import-support.md) | Adding import support allows `terraform import` to be run targeting an existing unmanaged resource and pulling its configuration into Terraform state. Typically import support is added during initial resource implementation but in some cases this will need to be added later. |
| [Documentation Changes](documentation-changes.md)| The provider documentation is displayed on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) and is sourced and refreshed from the provider repository during the release process. |

### 3. Write Tests

We require changes to be covered by [acceptance tests](running-and-writing-acceptance-tests.md) for all contributions. If you are unable to pay for acceptance tests for your contributions, mention this in your pull request. We will happily accept "best effort" acceptance tests implementations and run them for you on our side. Your PR may take longer to merge, but this is not a blocker for contributions.

### 4. Update the Changelog

HashiCorp's open-source projects have always maintained a user-friendly, readable CHANGELOG.md that allows users to tell at a glance whether a release should have any effect on them, and to gauge the risk of an upgrade. Not all changes require an entry in the changelog, refer to our [Changelog Process](changelog-process.md) for details about when and how to create a changelog.

### 5. Create a Pull Request

When your contribution is ready, Create a [Pull Request](raising-a-pull-request.md) in the AWS provider repository.

Pull requests are usually triaged within a few days of creation and are prioritized based on community reactions. Our [Prioritization Guides](prioritization.md) provides more details about the process.

### Submit an Issue

In addition to contributions, we welcome [bug reports](https://github.com/hashicorp/terraform-provider-aws/issues/new?assignees=&labels=&template=Bug_Report.md) and [feature requests](https://github.com/hashicorp/terraform-provider-aws/issues/new?assignees=&labels=enhancement&template=Feature_Request.md).
