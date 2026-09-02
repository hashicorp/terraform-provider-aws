---
subcategory: "Account Access"
layout: "aws"
page_title: "AWS: aws_accountaccess_entitlements"
description: |-
  Lists AWS Account Access Entitlements matching a required filter.
---

# Data Source: aws_accountaccess_entitlements

Lists AWS Account Access [Entitlements](../r/accountaccess_entitlement.html.markdown) for a given Application, filtered by principal, role, or target account.

## Example Usage

### Filter by Principal

```terraform
data "aws_accountaccess_entitlements" "example" {
  application_arn = aws_accountaccess_application.example.arn

  principal {
    identity_center {
      user_id = "11111111-2222-3333-4444-555555555555"
    }
  }
}
```

### Filter by Role

```terraform
data "aws_accountaccess_entitlements" "example" {
  application_arn = aws_accountaccess_application.example.arn
  role_arn        = "arn:aws:iam::123456789012:role/example-role"
}
```

### Filter by Target Account

```terraform
data "aws_accountaccess_entitlements" "example" {
  application_arn = aws_accountaccess_application.example.arn
  account_id      = "123456789012"
}
```

## Argument Reference

The following arguments are required:

* `application_arn` - (Required) ARN of the parent Application to list Entitlements within.

The following arguments are optional:

* `account_id` - (Optional) 12-digit AWS account ID to filter by. At least one of `account_id`, `role_arn`, or `principal` must be configured.
* `principal` - (Optional) IAM Identity Center principal to filter by. See [`principal` Block](#principal-block) below. At least one of `account_id`, `role_arn`, or `principal` must be configured.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Optional) Target IAM role ARN to filter by. At least one of `account_id`, `role_arn`, or `principal` must be configured.

### `principal` Block

The `principal` block supports:

* `identity_center` - (Required) IAM Identity Center principal filter. See [`principal.identity_center` Block](#principalidentity_center-block) below.

### `principal.identity_center` Block

The `principal.identity_center` block requires exactly one of the following arguments:

* `group_id` - (Optional) IAM Identity Center group ID to filter by.
* `user_id` - (Optional) IAM Identity Center user ID to filter by.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `entitlements` - List of matching Entitlements. See [`entitlements` Block](#entitlements-block) below.

### `entitlements` Block

The `entitlements` block contains:

* `created_at` - Date and time when the Entitlement was created in RFC 3339 format.
* `entitlement` - Principal-role entitlement configuration. See [`entitlements.entitlement` Block](#entitlementsentitlement-block) below.
* `entitlement_id` - Service-assigned unique identifier for the Entitlement.

### `entitlements.entitlement` Block

The `entitlements.entitlement` block contains:

* `principal_role` - Principal-role entitlement configuration. See [`entitlements.entitlement.principal_role` Block](#entitlementsentitlementprincipal_role-block) below.

### `entitlements.entitlement.principal_role` Block

The `entitlements.entitlement.principal_role` block contains:

* `account_id` - 12-digit AWS account ID for the target role.
* `account_name` - Human-readable name of the target account.
* `principal` - IAM Identity Center principal granted access. See [`entitlements.entitlement.principal_role.principal` Block](#entitlementsentitlementprincipal_roleprincipal-block) below.
* `role_arn` - Target IAM role ARN.

### `entitlements.entitlement.principal_role.principal` Block

The `entitlements.entitlement.principal_role.principal` block contains:

* `identity_center` - IAM Identity Center principal. See [`entitlements.entitlement.principal_role.principal.identity_center` Block](#entitlementsentitlementprincipal_roleprincipalidentity_center-block) below.

### `entitlements.entitlement.principal_role.principal.identity_center` Block

The `entitlements.entitlement.principal_role.principal.identity_center` block contains one of the following attributes:

* `group_id` - IAM Identity Center group ID.
* `user_id` - IAM Identity Center user ID.
