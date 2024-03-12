---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_drt_access_role_arn_association"
description: |-
  Terraform resource for managing an AWS Shield DRT Access Role ARN Association.
---

# Resource: aws_shield_drt_access_role_arn_association

Authorizes the Shield Response Team (SRT) using the specified role, to access your AWS account to assist with DDoS attack mitigation during potential attacks.
For more information see [Configure AWS SRT Support](https://docs.aws.amazon.com/waf/latest/developerguide/authorize-srt.html)

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "test" {
  name = var.aws_shield_drt_access_role_arn
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        "Sid" : "",
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "drt.shield.amazonaws.com"
        },
        "Action" : "sts:AssumeRole"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSShieldDRTAccessPolicy"
}

resource "aws_shield_drt_access_role_arn_association" "test" {
  role_arn = aws_iam_role.test.arn
}
```

## Argument Reference

The following arguments are required:

* `role_arn` - (Required) The Amazon Resource Name (ARN) of the role the SRT will use to access your AWS account. Prior to making the AssociateDRTRole request, you must attach the `AWSShieldDRTAccessPolicy` managed policy to this role.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Shield DRT access role ARN association using the AWS account ID. For example:

```terraform
import {
  to = aws_shield_drt_access_role_arn_association.example
  id = "123456789012"
}
```

Using `terraform import`, import Shield DRT access role ARN association using the AWS account ID. For example:

```console
% terraform import aws_shield_drt_access_role_arn_association.example 123456789012
```
