---
subcategory: "Verified Permissions"
layout: "aws"
page_title: "AWS: aws_verifiedpermissions_policy_store_alias"
description: |-
  Manages an AWS Verified Permissions Policy Store Alias.
---

# Resource: aws_verifiedpermissions_policy_store_alias

Manages an AWS Verified Permissions Policy Store Alias.

A policy store alias provides a user-defined name that can be used to reference an AWS Verified Permissions policy store.

## Example Usage

### Basic Usage

```terraform
resource "aws_verifiedpermissions_policy_store" "example" {
  validation_settings {
    mode = "OFF"
  }
}

resource "aws_verifiedpermissions_policy_store_alias" "example" {
  alias_name      = "policy-store-alias/example"
  policy_store_id = aws_verifiedpermissions_policy_store.example.policy_store_id
}
```

## Argument Reference

The following arguments are required:

* `alias_name` - (Required) Name of the policy store alias. The name must begin with `policy-store-alias/`. Changing this value forces replacement.
* `policy_store_id` - (Required) ID of the policy store associated with the alias. Changing this value forces replacement.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the policy store alias.
* `created_at` - Date and time when the policy store alias was created.
* `state` - Current state of the policy store alias.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_verifiedpermissions_policy_store_alias.example
  identity = {
    alias_name = "policy-store-alias/example"
  }
}

resource "aws_verifiedpermissions_policy_store_alias" "example" {
  alias_name      = "policy-store-alias/example"
  policy_store_id = "DxQg2j8xvXJQ1tQCYNWj9T"
}
```

### Identity Schema

#### Required

* `alias_name` - Name of the policy store alias.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Verified Permissions Policy Store Alias using its alias name. For example:

```terraform
import {
  to = aws_verifiedpermissions_policy_store_alias.example
  id = "policy-store-alias/example"
}
```

Using `terraform import`, import a Verified Permissions Policy Store Alias using its alias name. For example:

```console
% terraform import aws_verifiedpermissions_policy_store_alias.example policy-store-alias/example
```
