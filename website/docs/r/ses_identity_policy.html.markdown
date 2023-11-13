---
subcategory: "SES (Simple Email)"
layout: "aws"
page_title: "AWS: aws_ses_identity_policy"
description: |-
  Manages a SES Identity Policy
---

# Resource: aws_ses_identity_policy

Manages a SES Identity Policy. More information about SES Sending Authorization Policies can be found in the [SES Developer Guide](https://docs.aws.amazon.com/ses/latest/DeveloperGuide/sending-authorization-policies.html).

## Example Usage

```terraform
resource "aws_ses_domain_identity" "example" {
  domain = "example.com"
}

data "aws_iam_policy_document" "example" {
  statement {
    actions   = ["SES:SendEmail", "SES:SendRawEmail"]
    resources = [aws_ses_domain_identity.example.arn]

    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_ses_identity_policy" "example" {
  identity = aws_ses_domain_identity.example.arn
  name     = "example"
  policy   = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `identity` - (Required) Name or Amazon Resource Name (ARN) of the SES Identity.
* `name` - (Required) Name of the policy.
* `policy` - (Required) JSON string of the policy. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SES Identity Policies using the identity and policy name, separated by a pipe character (`|`). For example:

```terraform
import {
  to = aws_ses_identity_policy.example
  id = "example.com|example"
}
```

Using `terraform import`, import SES Identity Policies using the identity and policy name, separated by a pipe character (`|`). For example:

```console
% terraform import aws_ses_identity_policy.example 'example.com|example'
```
