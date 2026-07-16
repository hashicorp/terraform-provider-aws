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

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The description for the custom action target.
* `identifier` - (Required) The ID for the custom action target.
* `description` - (Required) The name of the custom action target.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the Security Hub custom action target.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_securityhub_action_target.example
  identity = {
    arn = "arn:aws:securityhub:eu-west-1:123456789012:action/custom/a"
  }
}

resource "aws_securityhub_action_target" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) Security Hub custom action ARN.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub custom actions using `arn`. For example:

```terraform
import {
  to = aws_securityhub_action_target.example
  id = "arn:aws:securityhub:eu-west-1:123456789012:action/custom/a"
}
```

Using `terraform import`, import Security Hub custom actions using `arn`. For example:

```console
% terraform import aws_securityhub_action_target.example arn:aws:securityhub:eu-west-1:123456789012:action/custom/a
```
