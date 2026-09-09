---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_entitlement"
description: |-
  Lists Account Access Entitlement resources.
---

# List Resource: aws_accountaccess_entitlement

Lists Account Access Entitlement resources for an Application and target AWS account.

## Example Usage

```terraform
list "aws_accountaccess_entitlement" "example" {
  provider = aws

  config {
    account_id      = "123456789012"
    application_arn = "arn:aws:account-access:us-east-1:123456789012:application/aam-0123456789abcdef"

    filter {
      principal_role {
        account_id = "123456789012"
      }
    }
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `application_arn` - (Required) ARN of the application to list entitlements for.
* `filter` - (Required) Filter criteria to narrow the entitlements returned. You can filter by principal, IAM role, or account. See [`filter` Block](#filter-block) below.
* `region` - (Optional) Region to query. Defaults to provider region.

### `filter` Block

The `filter` block supports:

* `principal_role` - (Required) principal-to-role filter criteria for narrowing entitlement results. See [`filter.principal_role` Block](#filterprincipal_role-block) below.

### `filter.principal_role` Block

The `filter.principal_role` block supports:

* `account_id` - (Optional) AWS account ID to filter entitlements by.
* `principal` - (Optional) principal to filter entitlements by. See [`filter.principal_role.principal` Block](#filterprincipal_roleprincipal-block) below.
* `role_arn` - (Optional) IAM role ARN to filter entitlements by.

### `filter.principal_role.principal` Block

The `filter.principal_role.principal` block supports:

* `identity_center` - (Required) IAM Identity Center principal filter criteria. See [`filter.principal_role.principal.identity_center` Block](#filterprincipal_roleprincipalidentity_center-block) below.

### `filter.principal_role.principal.identity_center` Block

The `filter.principal_role.principal.identity_center` block requires exactly one of the following arguments:

* `group_id` - (Optional) Unique identifier of a group in IAM Identity Center to filter by.
* `user_id` - (Optional) Unique identifier of a user in IAM Identity Center to filter by.
