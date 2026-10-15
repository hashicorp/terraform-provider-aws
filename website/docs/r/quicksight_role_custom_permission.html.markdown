---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_role_custom_permission"
description: |-
  Manages the custom permissions that are associated with a role.
---

# Resource: aws_quicksight_role_custom_permission

Manages the custom permissions that are associated with a role.

## Example Usage

```terraform
resource "aws_quicksight_role_custom_permission" "example" {
  role                    = "READER"
  custom_permissions_name = aws_quicksight_custom_permissions.example.custom_permissions_name
}
```

## Argument Reference

The following arguments are required:

* `custom_permissions_name` - (Required, Forces new resource) Custom permissions profile name.
* `role` - (Required, Forces new resource) Role. Valid values are `ADMIN`, `AUTHOR`, `READER`, `ADMIN_PRO`, `AUTHOR_PRO`, and `READER_PRO`.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) AWS account ID. Defaults to automatically determined account ID of the Terraform AWS provider.
* `namespace` - (Optional, Forces new resource) Namespace containing the role. Defaults to `default`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight role custom permissions using a comma-delimited string combining the `aws_account_id`, `namespace` and `role`. For example:

```terraform
import {
  to = aws_quicksight_role_custom_permission.example
  id = "012345678901,default,READER"
}
```

Using `terraform import`, import QuickSight role custom permissions using a comma-delimited string combining the `aws_account_id`, `namespace`, and `role`. For example:

```console
% terraform import aws_quicksight_role_custom_permission.example 012345678901,default,READER
```
