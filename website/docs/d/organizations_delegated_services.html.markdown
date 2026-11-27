---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_delegated_services"
description: |-
  Get a list the AWS services for which the specified account is a delegated administrator 
---

# Data Source: aws_organizations_delegated_services

Get a list the AWS services for which the specified account is a delegated administrator

## Example Usage

```terraform
data "aws_organizations_delegated_services" "example" {
  account_id = "AWS ACCOUNT ID"
}
```

## Argument Reference

This data source supports the following arguments:

* `account` - (Optional) AWS account ID where this data source will be read. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `account_id` - (Required) Account ID number of a delegated administrator account in the organization.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `delegated_services` - Services for which the account is a delegated administrator, which have the following attributes:
    * `delegation_enabled_date` - Date that the account became a delegated administrator for this service.
    * `service_principal` - Name of an AWS service that can request an operation for the specified service.
