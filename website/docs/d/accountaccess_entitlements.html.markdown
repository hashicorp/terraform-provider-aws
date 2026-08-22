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
  principal_id    = "11111111-2222-3333-4444-555555555555"
  principal_type  = "USER"
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

* `account_id` - (Optional) 12-digit AWS account ID to filter by. At least one of `account_id`, `role_arn`, or the `principal_id` and `principal_type` pair must be configured.
* `principal_id` - (Optional) IAM Identity Center user or group ID to filter by. Must be set with `principal_type`. At least one of `account_id`, `role_arn`, or the `principal_id` and `principal_type` pair must be configured.
* `principal_type` - (Optional) Type of principal. Valid values are `USER` and `GROUP`. Must be set with `principal_id`. At least one of `account_id`, `role_arn`, or the `principal_id` and `principal_type` pair must be configured.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Optional) Target IAM role ARN to filter by. At least one of `account_id`, `role_arn`, or the `principal_id` and `principal_type` pair must be configured.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `entitlements` - List of matching Entitlements. See [`entitlements` Block](#entitlements-block) below.
* `id` - Application ARN used to retrieve Entitlements.

### `entitlements` Block

The `entitlements` block contains:

* `account_id` - 12-digit AWS account ID for the target role.
* `account_name` - Human-readable name of the target account.
* `created_at` - Date and time when the Entitlement was created in RFC 3339 format.
* `entitlement_id` - Service-assigned unique identifier for the Entitlement.
* `principal_id` - IAM Identity Center user or group ID granted access.
* `principal_type` - Type of principal.
* `role_arn` - Target IAM role ARN.
