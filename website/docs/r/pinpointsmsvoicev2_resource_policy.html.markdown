---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_resource_policy"
description: |-
  Manages an AWS End User Messaging SMS Resource Policy.
---

# Resource: aws_pinpointsmsvoicev2_resource_policy

Manages an AWS End User Messaging SMS Resource Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_pinpointsmsvoicev2_phone_number" "example" {
  iso_country_code    = "US"
  message_type        = "TRANSACTIONAL"
  number_type         = "SIMULATOR"
  number_capabilities = ["SMS"]
}

data "aws_iam_policy_document" "example" {
  statement {
    effect    = "Allow"
    actions   = ["sms-voice:SendTextMessage"]
    resources = [aws_pinpointsmsvoicev2_phone_number.example.arn]
    principals {
      type        = "AWS"
      identifiers = ["123456789012"]
    }
  }
}

resource "aws_pinpointsmsvoicev2_resource_policy" "example" {
  resource_arn = aws_pinpointsmsvoicev2_phone_number.example.arn
  policy       = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) Resource-based policy document in JSON format.
* `resource_arn` - (Required) ARN of the End User Messaging SMS resource — phone number, opt-out list, pool, or sender ID — to attach the policy to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-options.html#cli-configure-options-region). Defaults to the region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_resource_policy.example
  identity = {
    resource_arn = "arn:aws:sms-voice:us-east-1:123456789012:phone-number/phone-abcdef0123456789abcdef0123456789"
  }
}

resource "aws_pinpointsmsvoicev2_resource_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` (String) ARN of the End User Messaging SMS resource the policy is attached to.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the resource policy using the parent resource ARN. For example:

```terraform
import {
  to = aws_pinpointsmsvoicev2_resource_policy.example
  id = "arn:aws:sms-voice:us-east-1:123456789012:phone-number/phone-abcdef0123456789abcdef0123456789"
}
```

Using `terraform import`, import the resource policy using the parent resource ARN. For example:

```console
% terraform import aws_pinpointsmsvoicev2_resource_policy.example arn:aws:sms-voice:us-east-1:123456789012:phone-number/phone-abcdef0123456789abcdef0123456789
```
