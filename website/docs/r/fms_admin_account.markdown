---
layout: "aws"
page_title: "AWS: aws_fms_admin_account"
sidebar_current: "docs-aws-fms-admin-account"
description: |-
  Provides a resource to associate/disassociate an AWS Firewall Manager administrator account
---

# aws_fms_admin_account

-> **Note:** There is only a single Firewall Manager administator account allowed per AWS account. Any existing administrator account will be lost when using this resource as an effect of this limitation.

Provides a resource to associate/disassociate an AWS Firewall Manager administrator account.

```hcl
resource "aws_fms_admin_account" "example" {
  account_id = "123456789012" # Required
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) The AWS account ID to associate with AWS Firewall Manager as the AWS Firewall Manager administrator account. This can be an AWS Organizations master account or a member account.

## Import

Firewall Manager administrator account association can be imported using the account ID, e.g.

```
$ terraform import aws_fms_admin_account.example 123456789012
```
