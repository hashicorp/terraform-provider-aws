---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_folder_membership"
description: |-
  Terraform resource for managing an AWS QuickSight Folder Membership.
---

# Resource: aws_quicksight_folder_membership

Terraform resource for managing an AWS QuickSight Folder Membership.

## Example Usage

### Basic Usage

```terraform
resource "aws_quicksight_folder_membership" "example" {
  folder_id   = aws_quicksight_folder.example.folder_id
  member_type = "DATASET"
  member_id   = aws_quicksight_data_set.example.data_set_id
}
```

## Argument Reference

The following arguments are required:

* `folder_id` - (Required, Forces new resource) Identifier for the folder.
* `member_id` - (Required, Forces new resource) ID of the asset (the dashboard, analysis, or dataset).
* `member_type` - (Required, Forces new resource) Type of the member. Valid values are `ANALYSIS`, `DASHBOARD`, and `DATASET`.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string joining AWS account ID, folder ID, member type, and member ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight Folder Membership using the AWS account ID, folder ID, member type, and member ID separated by commas (`,`). For example:

```terraform
import {
  to = aws_quicksight_folder_membership.example
  id = "123456789012,example-folder,DATASET,example-dataset"
}
```

Using `terraform import`, import QuickSight Folder Membership using the AWS account ID, folder ID, member type, and member ID separated by commas (`,`). For example:

```console
% terraform import aws_quicksight_folder_membership.example 123456789012,example-folder,DATASET,example-dataset
```
