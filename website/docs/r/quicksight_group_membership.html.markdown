---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_group_membership"
description: |-
  Manages a Resource QuickSight Group Membership.
---

# Resource: aws_quicksight_group_membership

Resource for managing QuickSight Group Membership

## Example Usage

```terraform
resource "aws_quicksight_group_membership" "example" {
  group_name  = "all-access-users"
  member_name = "john_smith"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `group_name` - (Required) The name of the group in which the member will be added.
* `member_name` - (Required) The name of the member to add to the group.
* `aws_account_id` - (Optional) The ID for the AWS account that the group is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `namespace` - (Required) The namespace that you want the user to be a part of. Defaults to `default`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight Group membership using the AWS account ID, namespace, group name and member name separated by `/`. For example:

```terraform
import {
  to = aws_quicksight_group_membership.example
  id = "123456789123/default/all-access-users/john_smith"
}
```

Using `terraform import`, import QuickSight Group membership using the AWS account ID, namespace, group name and member name separated by `/`. For example:

```console
% terraform import aws_quicksight_group_membership.example 123456789123/default/all-access-users/john_smith
```
