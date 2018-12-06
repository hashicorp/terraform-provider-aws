---
layout: "aws"
page_title: "AWS: aws_securityhub_account"
sidebar_current: "docs-aws-resource-securityhub-account"
description: |-
  Enables Security Hub.
---

# aws_securityhub_account

-> **Note:** Destroying this resource will disable Security Hub.

Enables Security Hub.

## Example Usage

```hcl
resource "aws_securityhub_member" "example" {}
```

## Argument Reference

The resource does not support any arguments.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - Returns `securityhub-account`.

## Import

Security Hub account enablemenet can be imported using the word `securityhub-account`, e.g.

```
$ terraform import aws_securityhub_account.example securityhub-account
```
