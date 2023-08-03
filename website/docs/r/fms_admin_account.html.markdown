---
subcategory: "FMS (Firewall Manager)"
layout: "aws"
page_title: "AWS: aws_fms_admin_account"
description: |-
  Provides a resource to associate/disassociate an AWS Firewall Manager administrator account
---

# Resource: aws_fms_admin_account

Provides a resource to associate/disassociate an AWS Firewall Manager administrator account. This operation must be performed in the `us-east-1` region.

## Example Usage

```terraform
resource "aws_fms_admin_account" "example" {}
```

## Argument Reference

This resource supports the following arguments:

* `account_id` - (Optional) The AWS account ID to associate with AWS Firewall Manager as the AWS Firewall Manager administrator account. This can be an AWS Organizations master account or a member account. Defaults to the current account. Must be configured to perform drift detection.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS account ID of the AWS Firewall Manager administrator account.

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
