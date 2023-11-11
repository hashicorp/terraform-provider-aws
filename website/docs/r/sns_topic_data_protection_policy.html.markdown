---
subcategory: "SNS (Simple Notification)"
layout: "aws"
page_title: "AWS: aws_sns_topic_data_protection_policy"
description: |-
  Provides an SNS data protection topic policy resource.
---

# Resource: aws_sns_topic_data_protection_policy

Provides an SNS data protection topic policy resource

## Example Usage

```terraform
resource "aws_sns_topic" "example" {
  name = "example"
}

resource "aws_sns_topic_data_protection_policy" "example" {
  arn = aws_sns_topic.example.arn
  policy = jsonencode(
    {
      "Description" = "Example data protection policy"
      "Name"        = "__example_data_protection_policy"
      "Statement" = [
        {
          "DataDirection" = "Inbound"
          "DataIdentifier" = [
            "arn:aws:dataprotection::aws:data-identifier/EmailAddress",
          ]
          "Operation" = {
            "Deny" = {}
          }
          "Principal" = [
            "*",
          ]
          "Sid" = "__deny_statement_11ba9d96"
        },
      ]
      "Version" = "2021-06-01"
    }
  )
}
```

## Argument Reference

This resource supports the following arguments:

* `arn` - (Required) The ARN of the SNS topic
* `policy` - (Required) The fully-formed AWS policy as JSON. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SNS Data Protection Topic Policy using the topic ARN. For example:

```terraform
import {
  to = aws_sns_topic_data_protection_policy.example
  id = "arn:aws:sns:us-west-2:0123456789012:example"
}
```

Using `terraform import`, import SNS Data Protection Topic Policy using the topic ARN. For example:

```console
% terraform import aws_sns_topic_data_protection_policy.example arn:aws:sns:us-west-2:0123456789012:example
```
