---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_action_target"
description: |-
  Creates Security Hub custom action.
---

# Resource: aws_securityhub_action_target

Creates Security Hub custom action.

## Example Usage

```terraform
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_action_target" "example" {
  depends_on  = [aws_securityhub_account.example]
  name        = "Send notification to chat"
  identifier  = "SendToChat"
  description = "This is custom action sends selected findings to chat"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The description for the custom action target.
* `identifier` - (Required) The ID for the custom action target.
* `description` - (Required) The name of the custom action target.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Security Hub custom action target.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub custom action using the action target ARN. For example:

```terraform
import {
  to = aws_securityhub_action_target.example
  id = "arn:aws:securityhub:eu-west-1:312940875350:action/custom/a"
}
```

Using `terraform import`, import Security Hub custom action using the action target ARN. For example:

```console
% terraform import aws_securityhub_action_target.example arn:aws:securityhub:eu-west-1:312940875350:action/custom/a
```
