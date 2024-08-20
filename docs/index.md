# Welcome

The Terraform AWS Provider is the work of thousands of contributors, and is maintained by a small team within HashiCorp. This site contains extensive instructions about how to contribute and how the AWS provider works.

!!! tip
    This documentation is intended for Terraform AWS Provider code developers. Typical operators writing and applying Terraform configurations do not need to read or understand this material.

## Contribute

Please follow the following steps to ensure your contribution goes smoothly.

### 1. Configure Development Environment

Install Terraform and Go. Clone the repository, compile the provider, and set up testing. Refer to [Configure Development Environment](development-environment.md).

### 2. Debug Code

If you are looking to _create or enhance code_, such as a new resource or adding an argument to an existing resource, skip to the next step.

Finding and fixing errors in the AWS Provider can be difficult. We have a [debugging guide](debugging.md) to help you get started.

### 3. Change Code

Follow the guide for your contribution type and refer to the Development Reference materials as needed for additional details about [provider design](provider-design.md), expected [naming conventions](naming.md), guidance for [error handling](error-handling.md), etc.

| Contribution Guide | Description |
|--------------------|-------------|
| [Small Changes](bugs-and-enhancements.md) | Requirements for small additions or bug-fixes on existing resources/data sources |
| [Resources](add-a-new-resource.md) | Allow the management of a logical resource within AWS by adding a new resource to the Terraform AWS Provider. |
| [Data Source](add-a-new-datasource.md) | Let your Terraform configurations use data from resources not under local management by creating ready only data sources. |
| [Services](add-a-new-service.md) | Allow Terraform (via the AWS Provider) to manage an entirely new AWS service by introducing the resources and data sources required to manage configuration of the service. |
| [AWS Region](add-a-new-region.md) | New regions are immediately usable with the provider with the caveat that a configuration workaround is required to skip validation of the region during cli operations. A small set of changes are required to makes this workaround necessary. |
| [Resource Name Generation](resource-name-generation.md) | Allow a resource to either fully, or partially, generate its own resource names. This can be useful in cases where the resource name uniquely identifes the resource and it needs to be recreated. It can also be used when a name is required, but the specific name is not important. |
| [Tagging Support](resource-tagging.md) | Many AWS resources allow assigning metadata via tags. However, frequently AWS services are launched without tagging support so this will often need to be added later. |
| [Import Support](add-import-support.md) | Adding import support allows `terraform import` to be run targeting an existing unmanaged resource and pulling its configuration into Terraform state. Typically import support is added during initial resource implementation but in some cases this will need to be added later. |
| [Documentation Changes](documentation-changes.md)| The provider documentation is displayed on the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) and is sourced and refreshed from the provider repository during the release process. |

### 4. Write Tests

We require all changes to be covered by [acceptance tests](running-and-writing-acceptance-tests.md) and/or [unit tests](unit-tests.md), depending on the situation. In the context of the Terraform AWS Provider, _acceptance tests_ are tests of interactions with AWS, such as creating, reading information about, and destroying AWS resources. In contrast, _unit tests_ test functionality wholly within the provider itself, such as function tests.

If you are unable to pay for acceptance tests for your contributions, mention this in your pull request. We will happily accept "best effort" acceptance tests implementations and run them for you on our side. Your PR may take longer to merge, but this is not a blocker for contributions.

### 5. Continuous Integration

When submitting a pull request, you'll notice that we run several automated processes on your proposed change.Some of these processes are tests to ensure your contribution aligns with our standards. While we strive for accuracy, some users may find these tests confusing. Check out [Continuous Integration](continuous-integration.md) for additional clarity.

### 6. Update the Changelog

HashiCorp's open-source projects have always maintained a user-friendly, readable CHANGELOG.md that allows users to tell at a glance whether a release should have any effect on them, and to gauge the risk of an upgrade. Not all changes require an entry in the changelog, refer to our [Changelog Process](changelog-process.md) for details about when and how to create a changelog.

### 7. Create a Pull Request

When your contribution is ready, Create a [Pull Request](raising-a-pull-request.md) in the AWS provider repository.

Pull requests are usually triaged within a few days of creation and are prioritized based on community reactions. Our [Prioritization Guides](prioritization.md) provide more details about the process.

## Submit an Issue

In addition to contributions, we welcome [bug reports](https://github.com/hashicorp/terraform-provider-aws/issues/new?assignees=&labels=&template=Bug_Report.md) and [feature requests](https://github.com/hashicorp/terraform-provider-aws/issues/new?assignees=&labels=enhancement&template=Feature_Request.md).

## Join the Contributors Slack

For frequent contributors, it's useful to join the contributors Slack channel hosted within the HashiCorp Slack workspace. This Slack channel is used to discuss topics such as general contribution questions, suggestions for improving the contribution process, coordinating on pair programming sessions, etc. The channel is not intended as a place to request status updates on open issues or pull requests. For prioritization questions, instead refer to the [prioritization guide](prioritization.md).

To request to join, fill out the [request form](https://forms.gle/Gf9ZAmUYXuzafkct6) and allow time for the request to be reviewed and processed.
