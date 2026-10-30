---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_account_custom_permission"
description: |-
  Manages an AWS QuickSight Account Custom Permission.
---

# Resource: aws_quicksight_account_custom_permission

Manages an AWS QuickSight Account Custom Permission.

## Example Usage

```terraform
resource "aws_quicksight_account_custom_permission" "example" {
  custom_permissions_name = aws_quicksight_custom_permissions.example.custom_permissions_name
}
```

## Argument Reference

The following arguments are required:

* `custom_permissions_name` - (Required) Name of the custom permissions profile to apply to the account.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID. Defaults to the account ID determined from the provider configuration.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_quicksight_account_custom_permission.example
  identity = {
    aws_account_id = "012345678900"
  }
}

resource "aws_quicksight_account_custom_permission" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `aws_account_id` (String) AWS account ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import QuickSight Account Custom Permission using the `example_id_arg`. For example:

```terraform
import {
  to = aws_quicksight_account_custom_permission.example
  id = "account_custom_permission-id-012345678900"
}
```

Using `terraform import`, import QuickSight Account Custom Permission using the `example_id_arg`. For example:

```console
% terraform import aws_quicksight_account_custom_permission.example account_custom_permission-id-012345678900
```
