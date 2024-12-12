---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_group"
description: |-
  Manages a Resource QuickSight Group.
---

# Resource: aws_quicksight_group

Resource for managing QuickSight Group

## Example Usage

```terraform
resource "aws_quicksight_group" "example" {
  group_name = "tf-example"
}
```

## Argument Reference

This resource supports the following arguments:

* `group_name` - (Required) A name for the group.
* `aws_account_id` - (Optional) The ID for the AWS account that the group is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `description` - (Optional) A description for the group.
* `namespace` - (Optional) The namespace. Currently, you should set this to `default`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of group

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight Group using the aws account id, namespace and group name separated by `/`. For example:

```terraform
import {
  to = aws_quicksight_group.example
  id = "123456789123/default/tf-example"
}
```

Using `terraform import`, import QuickSight Group using the aws account id, namespace and group name separated by `/`. For example:

```console
% terraform import aws_quicksight_group.example 123456789123/default/tf-example
```
