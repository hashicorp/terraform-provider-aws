---
subcategory: "FMS (Firewall Manager)"
layout: "aws"
page_title: "AWS: aws_fms_admin_account"
description: |-
  Provides a resource to associate/disassociate an AWS Firewall Manager administrator account
---

# Resource: aws_fms_admin_account

Provides a resource to associate/disassociate an AWS Firewall Manager administrator account. This operation must be performed in the `us-east-1` region.

~> **NOTE:** You can have up to 10 Firewall Manager administrators per AWS Organization. The first administrator created (or any without `admin_scope`) becomes the default administrator with full permissions. Additional administrators can be created with `admin_scope` to restrict their permissions to specific accounts, organizational units, regions, or policy types. Only the default administrator can manage third-party firewalls.

~> **IMPORTANT:** You must create a default administrator (without `admin_scope`) before creating any scoped administrators. If you attempt to create a scoped administrator as the first administrator, the operation may fail.

## Example Usage

### Basic Usage - Default Administrator

```terraform
resource "aws_fms_admin_account" "example" {}
```

### Multiple Administrators - Default and Scoped

```terraform
# First, create the default administrator with full permissions
resource "aws_fms_admin_account" "default" {
  account_id = "123456789012"
}

# Then, create scoped administrators with restricted permissions
resource "aws_fms_admin_account" "security_team" {
  account_id = "111111111111"

  admin_scope {
    region_scope {
      regions             = ["us-east-1", "us-west-2"]
      all_regions_enabled = false
    }

    policy_type_scope {
      policy_types             = ["WAF", "WAFV2", "SHIELD_ADVANCED"]
      all_policy_types_enabled = false
    }
  }

  depends_on = [aws_fms_admin_account.default]
}
```

### Scoped Administrator with Full Configuration

```terraform
resource "aws_fms_admin_account" "example" {
  account_id = "123456789012"

  admin_scope {
    region_scope {
      regions             = ["us-east-1", "us-west-2"]
      all_regions_enabled = false
    }

    policy_type_scope {
      policy_types             = ["WAF", "WAFV2"]
      all_policy_types_enabled = false
    }

    account_scope {
      accounts                    = ["111111111111", "222222222222"]
      all_accounts_enabled        = false
      exclude_specified_accounts  = false
    }

    organizational_unit_scope {
      organizational_units                     = ["ou-xxxx-yyyyyyyy"]
      all_organizational_units_enabled         = false
      exclude_specified_organizational_units  = false
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID to associate with AWS Firewall Manager as the AWS Firewall Manager administrator account. This can be an AWS Organizations master account or a member account. Defaults to the current account. Must be configured to perform drift detection.
* `admin_scope` - (Optional) Configuration block for defining the scope of the Firewall Manager administrator. When omitted, creates a default administrator with full permissions. When specified, creates a scoped administrator with restricted permissions. You can have up to 10 Firewall Manager administrators per organization (1 default + 9 scoped). Only the default administrator can manage third-party firewalls. See [`admin_scope`](#admin_scope) below.

### admin_scope

The `admin_scope` configuration block supports the following arguments:

* `account_scope` - (Optional) Configuration block for defining the accounts within the administrator scope. See [`account_scope`](#account_scope) below.
* `organizational_unit_scope` - (Optional) Configuration block for defining the organizational units within the administrator scope. See [`organizational_unit_scope`](#organizational_unit_scope) below.
* `policy_type_scope` - (Optional) Configuration block for defining the policy types within the administrator scope. See [`policy_type_scope`](#policy_type_scope) below.
* `region_scope` - (Optional) Configuration block for defining the regions within the administrator scope. See [`region_scope`](#region_scope) below.

### account_scope

The `account_scope` configuration block supports the following arguments:

* `accounts` - (Optional) Set of AWS account IDs. The behavior depends on `exclude_specified_accounts`. Cannot be used with `all_accounts_enabled` set to `true`.
* `all_accounts_enabled` - (Optional) Whether to include all accounts in the scope. When `true`, the `accounts` list cannot be specified. Default is `false`.
* `exclude_specified_accounts` - (Optional) Whether to exclude the specified accounts from the scope. When `true`, the accounts in the `accounts` list are excluded from the scope. When `false`, only the accounts in the `accounts` list are included. Only applies when `accounts` is specified. Default is `false`.

**Examples:**

Include all accounts:
```terraform
account_scope {
  all_accounts_enabled = true
}
```

Include only specific accounts:
```terraform
account_scope {
  accounts                   = ["111111111111", "222222222222"]
  all_accounts_enabled       = false
  exclude_specified_accounts = false
}
```

Include all accounts except specific ones:
```terraform
account_scope {
  accounts                   = ["111111111111", "222222222222"]
  all_accounts_enabled       = false
  exclude_specified_accounts = true
}
```

### organizational_unit_scope

The `organizational_unit_scope` configuration block supports the following arguments:

* `organizational_units` - (Optional) Set of AWS Organizations organizational unit IDs. The behavior depends on `exclude_specified_organizational_units`. Cannot be used with `all_organizational_units_enabled` set to `true`.
* `all_organizational_units_enabled` - (Optional) Whether to include all organizational units in the scope. When `true`, the `organizational_units` list cannot be specified. Default is `false`.
* `exclude_specified_organizational_units` - (Optional) Whether to exclude the specified organizational units from the scope. When `true`, the OUs in the `organizational_units` list are excluded from the scope. When `false`, only the OUs in the `organizational_units` list are included. Only applies when `organizational_units` is specified. Default is `false`.

### policy_type_scope

The `policy_type_scope` configuration block supports the following arguments:

* `policy_types` - (Optional) Set of policy types to include in the scope. Valid values: `WAF`, `WAFV2`, `SHIELD_ADVANCED`, `SECURITY_GROUPS_COMMON`, `SECURITY_GROUPS_CONTENT_AUDIT`, `SECURITY_GROUPS_USAGE_AUDIT`, `NETWORK_FIREWALL`, `DNS_FIREWALL`, `THIRD_PARTY_FIREWALL`, `IMPORT_NETWORK_FIREWALL`, `NETWORK_ACL_COMMON`. Cannot be used with `all_policy_types_enabled` set to `true`.
* `all_policy_types_enabled` - (Optional) Whether to include all policy types in the scope. Default is `false`.

### region_scope

The `region_scope` configuration block supports the following arguments:

* `regions` - (Optional) Set of AWS regions to include in the scope. Cannot be used with `all_regions_enabled` set to `true`.
* `all_regions_enabled` - (Optional) Whether to include all regions in the scope. Default is `false`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS account ID of the AWS Firewall Manager administrator account.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `30m`)
- `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Firewall Manager administrator account association using the account ID. For example:

```terraform
import {
  to = aws_fms_admin_account.example
  id = "123456789012"
}
```

Using `terraform import`, import Firewall Manager administrator account association using the account ID. For example:

```console
% terraform import aws_fms_admin_account.example 123456789012
```
