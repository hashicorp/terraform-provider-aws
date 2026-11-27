---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_delegated_administrators"
description: |-
  Get a list of AWS accounts that are designated as delegated administrators in this organization 
---

# Data Source: aws_organizations_delegated_administrators

Get a list of AWS accounts that are designated as delegated administrators in this organization

## Example Usage

```terraform
data "aws_organizations_delegated_administrators" "example" {
  service_principal = "SERVICE PRINCIPAL"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `service_principal` - (Optional) Service principal name. If specified, then the operation lists the delegated administrators only for the specified service. If you don't specify a service principal, the operation lists all delegated administrators for all services in your organization.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `delegated_administrators` - List of delegated administrators in your organization, which have the following attributes:
    * `arn` - ARN of the delegated administrator's account.
    * `delegation_enabled_date` - Date when the account was made a delegated administrator.
    * `email` - Email address that is associated with the delegated administrator's AWS account.
    * `id` - Unique identifier (ID) of the delegated administrator's account.
    * `joined_method` - Method by which the delegated administrator's account joined the organization.
    * `joined_timestamp` - Date when the delegated administrator's account became a part of the organization.
    * `name` - Friendly name of the delegated administrator's account.
    * `status` - Status of the delegated administrator's account in the organization.
