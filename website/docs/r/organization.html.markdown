---
layout: "aws"
page_title: "AWS: aws_organization
sidebar_current: "docs-aws-resource-organization|"
description: |-
  Provides a resource to create an organization.
---

# aws_organization

Provides a resource to create an organization.

## Example Usage:

```hcl
resource "aws_organization" "org" {
  feature_set = "ALL"
}
```

## Argument Reference

The following arguments are supported:

* `feature_set` - (Optional) Specify "ALL" (default) or "CONSOLIDATED_BILLING".

## Attributes Reference

The following additional attributes are exported:

* `arn` - ARN of the organization
* `id` - Identifier of the organization
* `master_account_arn` - ARN of the master account
* `master_account_email` - Email address of the master account
* `master_account_id` - Identifier of the master account

## Import

The AWS organization can be imported by using the `account_id`, e.g.

```
$ terraform import aws_organization.my_org o-1234567
```
