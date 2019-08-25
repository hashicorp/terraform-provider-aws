---
layout: "aws"
page_title: "AWS: aws_quicksight_group"
sidebar_current: "docs-aws-resource-quicksight-group"
description: |-
  Manages a Resource Quick Sight Group.
---

# Resource: aws_quicksight_group

Resource for managing Quick Sight Group

## Example Usage

```hcl
resource "aws_quicksight_group" "example" {
	group_name = "tf-example"
}
```

## Argument Reference

The following arguments are supported:

* `group_name` - (Required) A name for the group.
* `aws_account_id` - (Optional) The ID for the AWS account that the group is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `description` - (Optional) A description for the group.
* `namespace` - (Optional) The namespace. Currently, you should set this to `default`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of group

## Import

Quick Sight Group can be imported using the aws account id, namespace and group name separated by `/`.

```
$ terraform import aws_quicksight_group.example 123456789123/default/tf-example
```