---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_account_suppression_attributes"
description: |-
  Manages AWS SESv2 (Simple Email V2) account-level suppression attributes.
---

# Resource: aws_sesv2_account_suppression_attributes

Manages AWS SESv2 (Simple Email V2) account-level suppression attributes.

## Example Usage

```terraform
resource "aws_sesv2_account_suppression_attributes" "example" {
  suppressed_reasons = ["COMPLAINT"]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `suppressed_reasons` - (Required) A list that contains the reasons that email addresses will be automatically added to the suppression list for your account. Valid values: `COMPLAINT`, `BOUNCE`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import account-level suppression attributes using the account ID. For example:

```terraform
import {
  to = aws_sesv2_account_suppression_attributes.example
  id = "123456789012"
}
```

Using `terraform import`, import account-level suppression attributes using the account ID. For example:

```console
% terraform import aws_sesv2_account_suppression_attributes.example 123456789012
```
