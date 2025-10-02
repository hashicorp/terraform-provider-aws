---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_user_custom_permission"
description: |-
  Manages the custom permissions profile for a user.
---

# Resource: aws_quicksight_user_custom_permission

Manages the custom permissions profile for a user.

## Example Usage

```terraform
resource "aws_quicksight_user_custom_permission" "example" {
  user_name               = aws_quicksight_user.example.user_name
  custom_permissions_name = aws_quicksight_custom_permissions.example.custom_permissions_name
}
```

## Argument Reference

The following arguments are required:

* `custom_permissions_name` - (Required, Forces new resource) Custom permissions profile name.
* `user_name` - (Required, Forces new resource) Username of the user.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `namespace` - (Optional, Forces new resource) Namespace that the user belongs to. Defaults to `default`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight user custom permissions using a comma-delimited string combining the `aws_account_id`, `namespace` and `user_name`. For example:

```terraform
import {
  to = aws_quicksight_user_custom_permission.example
  id = "012345678901,default,user1"
}
```

Using `terraform import`, import QuickSight user custom permissions using a comma-delimited string combining the `aws_account_id`, `namespace`, and `user_name`. For example:

```console
% terraform import aws_quicksight_user_custom_permission.example 012345678901,default,user1
```
