---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_folder"
description: |-
  Manages a QuickSight Folder.
---

# Resource: aws_quicksight_folder

Resource for managing a QuickSight Folder.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_folder" "example" {
  folder_id = "example-id"
  name      = "example-name"
}
```

### With Permissions

```terraform
resource "aws_quicksight_folder" "example" {
  folder_id = "example-id"
  name      = "example-name"

  permissions {
    actions = [
      "quicksight:CreateFolder",
      "quicksight:DescribeFolder",
      "quicksight:UpdateFolder",
      "quicksight:DeleteFolder",
      "quicksight:CreateFolderMembership",
      "quicksight:DeleteFolderMembership",
      "quicksight:DescribeFolderPermissions",
      "quicksight:UpdateFolderPermissions",
    ]
    principal = aws_quicksight_user.example.arn
  }
}
```

### With Parent Folder

```terraform
resource "aws_quicksight_folder" "parent" {
  folder_id = "parent-id"
  name      = "parent-name"
}

resource "aws_quicksight_folder" "example" {
  folder_id = "example-id"
  name      = "example-name"

  parent_folder_arn = aws_quicksight_folder.parent.arn
}
```

## Argument Reference

The following arguments are required:

* `folder_id` - (Required, Forces new resource) Identifier for the folder.
* `name` - (Required) Display name for the folder.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID.
* `folder_type` - (Optional) The type of folder. By default, it is `SHARED`. Valid values are: `SHARED`.
* `parent_folder_arn` - (Optional) The Amazon Resource Name (ARN) for the parent folder. If not set, creates a root-level folder.
* `permissions` - (Optional) A set of resource permissions on the folder. Maximum of 64 items. See [permissions](#permissions).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### permissions

* `actions` - (Required) List of IAM actions to grant or revoke permissions on.
* `principal` - (Required) ARN of the principal. See the [ResourcePermission documentation](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ResourcePermission.html) for the applicable ARN values.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the folder.
* `created_time` - The time that the folder was created.
* `folder_path` - An array of ancestor ARN strings for the folder. Empty for root-level folders.
* `id` - A comma-delimited string joining AWS account ID and folder ID.
* `last_updated_time` - The time that the folder was last updated.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `read`   - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a QuickSight folder using the AWS account ID and folder ID name separated by a comma (`,`). For example:

```terraform
import {
  to = aws_quicksight_folder.example
  id = "123456789012,example-id"
}
```

Using `terraform import`, import a QuickSight folder using the AWS account ID and folder ID name separated by a comma (`,`). For example:

```console
% terraform import aws_quicksight_folder.example 123456789012,example-id
```
