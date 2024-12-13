---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_delegated_administrator"
description: |-
  Provides a resource to manage an AWS Organizations Delegated Administrator.
---

# Resource: aws_organizations_delegated_administrator

Provides a resource to manage an [AWS Organizations Delegated Administrator](https://docs.aws.amazon.com/organizations/latest/APIReference/API_RegisterDelegatedAdministrator.html).

## Example Usage

```terraform
resource "aws_organizations_delegated_administrator" "example" {
  account_id        = "123456789012"
  service_principal = "principal"
}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Required) The account ID number of the member account in the organization to register as a delegated administrator.
* `service_principal` - (Required) The service principal of the AWS service for which you want to make the member account a delegated administrator.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique identifier (ID) of the delegated administrator.
* `arn` - The Amazon Resource Name (ARN) of the delegated administrator's account.
* `delegation_enabled_date` - The date when the account was made a delegated administrator.
* `email` - The email address that is associated with the delegated administrator's AWS account.
* `joined_method` - The method by which the delegated administrator's account joined the organization.
* `joined_timestamp` - The date when the delegated administrator's account became a part of the organization.
* `name` - The friendly name of the delegated administrator's account.
* `status` - The status of the delegated administrator's account in the organization.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_organizations_delegated_administrator` using the account ID and its service principal. For example:

```terraform
import {
  to = aws_organizations_delegated_administrator.example
  id = "123456789012/config.amazonaws.com"
}
```

Using `terraform import`, import `aws_organizations_delegated_administrator` using the account ID and its service principal. For example:

```console
% terraform import aws_organizations_delegated_administrator.example 123456789012/config.amazonaws.com
```
