---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_admin_account"
description: |-
  Terraform resource for managing an AWS Network Security Manager Admin Account.
---

# Resource: aws_networksecuritymanager_admin_account

Terraform resource for managing an AWS Network Security Manager Admin Account. An administrator account can centrally create and manage AWS WAF and Shield Advanced protection across the accounts and organizational units (OUs) of an AWS Organization. The resource must be managed from the organization's management account, and the administrator's authority can be narrowed with an `admin_scope`.

~> Setting an administrator account again shortly after it was removed, or while the service is still creating its service-linked role, is rejected by AWS with a `ConflictException`. The resource retries the request until it is accepted or the `create` timeout expires.

## Example Usage

### Basic Usage

```terraform
resource "aws_networksecuritymanager_admin_account" "example" {
  admin_account_id = "123456789012"
  priority         = 1
}
```

### Scoped Administrator

```terraform
resource "aws_networksecuritymanager_admin_account" "example" {
  admin_account_id = "123456789012"
  priority         = 2

  admin_scope {
    firewall_type_scope {
      firewall_types = ["WAF"]
    }

    scope_filter {
      include_only {
        organizational_units = ["ou-abcd-12345678"]
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `admin_account_id` - (Required, Forces new resource) AWS account ID to set as a Network Security Manager administrator account.
* `priority` - (Required) Priority of the administrator account, between `1` and `10`.

The following arguments are optional:

* `admin_scope` - (Optional) Scope of what the administrator account can manage. If omitted, AWS applies its default scope. See [`admin_scope` Block](#admin_scope-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#region).

### `admin_scope` Block

The `admin_scope` block supports the following arguments:

* `firewall_type_scope` - (Optional) Firewall types that the administrator can create and manage. See [`firewall_type_scope` Block](#firewall_type_scope-block) below.
* `scope_filter` - (Optional) Accounts and OUs that are in the administrator's scope. See [`scope_filter` Block](#scope_filter-block) below.

### `firewall_type_scope` Block

The `firewall_type_scope` block supports the following arguments:

* `all_firewall_types_enabled` - (Optional) Whether the administrator can manage all firewall types.
* `firewall_types` - (Optional) Set of firewall types that the administrator can manage. Valid values: `WAF`, `SHIELD_ADVANCED`.

### `scope_filter` Block

The `scope_filter` block supports the following arguments. Exactly one of them must be configured:

* `exclude_only` - (Optional) Accounts and OUs to exclude from the administrator's scope; everything else is in scope. See [`exclude_only` Block](#exclude_only-block) below.
* `include_all` - (Optional) Set to `true` to put all accounts and OUs of the organization in scope.
* `include_only` - (Optional) Only the specified accounts and OUs are in the administrator's scope. See [`include_only` Block](#include_only-block) below.

### `exclude_only` Block

The `exclude_only` block supports the following arguments:

* `accounts` - (Optional) Set of AWS account IDs.
* `organizational_units` - (Optional) Set of AWS Organizations OU IDs (`ou-...`). Root IDs are not accepted.

### `include_only` Block

The `include_only` block supports the following arguments:

* `accounts` - (Optional) Set of AWS account IDs.
* `organizational_units` - (Optional) Set of AWS Organizations OU IDs (`ou-...`). Root IDs are not accepted.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `status` - Status of the administrator account. Valid values: `ONBOARDED`, `OFFBOARDED`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_networksecuritymanager_admin_account.example
  identity = {
    admin_account_id = "123456789012"
  }
}
```

### Identity Schema

#### Required

* `admin_account_id` - (String) AWS account ID of the administrator account.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Network Security Manager Admin Account using the `admin_account_id`. For example:

```terraform
import {
  to = aws_networksecuritymanager_admin_account.example
  id = "123456789012"
}
```

Using `terraform import`, import Network Security Manager Admin Account using the `admin_account_id`. For example:

```console
% terraform import aws_networksecuritymanager_admin_account.example 123456789012
```
