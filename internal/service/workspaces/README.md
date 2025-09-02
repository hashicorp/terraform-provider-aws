# Terraform AWS Provider WorkSpaces Package

This area is primarily for AWS provider contributors and maintainers. For information on _using_ Terraform and the AWS provider, see the links below.

Acceptance tests for the following resource types are bundled into the `TestAccWorkSpaces_serial` test:

* `aws_workspaces_directory`
* `aws_workspaces_ip_group`
* `aws_workspaces_workspace`

Acceptance tests for the following data sources are bundled into the `TestAccWorkSpacesDataSource_serial` test:

* `aws_workspaces_bundle`
* `aws_workspaces_directory`
* `aws_workspaces_image`
* `aws_workspaces_workspace`

To invoke specific tests in a bundle, use the subtest specification syntax (`/<Package>` or `/<Package>/<test>`).

## Handy Links

* [Find out about contributing](https://hashicorp.github.io/terraform-provider-aws/#contribute) to the AWS provider!
* AWS Provider Docs: [Home](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
* AWS Provider Docs: [One of the WorkSpaces resources](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/workspaces_directory)
* AWS Docs: [AWS SDK for Go WorkSpaces](https://docs.aws.amazon.com/sdk-for-go/api/service/workspaces/)
