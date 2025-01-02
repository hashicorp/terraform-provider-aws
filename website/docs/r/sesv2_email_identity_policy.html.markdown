---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_email_identity_policy"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Email Identity Policy.
---
# Resource: aws_sesv2_email_identity_policy

Terraform resource for managing an AWS SESv2 (Simple Email V2) Email Identity Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_email_identity" "example" {
  email_identity = "testing@example.com"
}

resource "aws_sesv2_email_identity_policy" "example" {
  email_identity = aws_sesv2_email_identity.example.email_identity
  policy_name    = "example"

  policy = <<EOF
{
  "Id":"ExampleAuthorizationPolicy",
  "Version":"2012-10-17",
  "Statement":[
    {
      "Sid":"AuthorizeIAMUser",
      "Effect":"Allow",
      "Resource":"${aws_sesv2_email_identity.example.arn}",
      "Principal":{
        "AWS":[
          "arn:aws:iam::123456789012:user/John",
          "arn:aws:iam::123456789012:user/Jane"
        ]
      },
      "Action":[
        "ses:DeleteEmailIdentity",
        "ses:PutEmailIdentityDkimSigningAttributes"
      ]
    }
  ]
}
EOF
}
```

## Argument Reference

The following arguments are required:

* `email_identity` - (Required) The email identity.
* `policy_name` - (Required) - The name of the policy.
* `policy` - (Required) - The text of the policy in JSON format.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SESv2 (Simple Email V2) Email Identity Policy using the `id` (`email_identity|policy_name`). For example:

```terraform
import {
  to = aws_sesv2_email_identity_policy.example
  id = "example_email_identity|example_policy_name"
}
```

Using `terraform import`, import SESv2 (Simple Email V2) Email Identity Policy using the `example_id_arg`. For example:

```console
% terraform import aws_sesv2_email_identity_policy.example example_email_identity|example_policy_name
```
