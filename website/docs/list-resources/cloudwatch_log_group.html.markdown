---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_group"
description: |-
  Lists CloudWatch Logs Log Group resources.
---

# List Resource: aws_cloudwatch_log_group

~> **Note:** The `aws_cloudwatch_log_group` List Resource is in beta. Its interface and behavior may change as the feature evolves, and breaking changes are possible. It is offered as a technical preview without compatibility guarantees until Terraform 1.14 is generally available.

Lists CloudWatch Logs Log Group resources.

## Example Usage

```terraform
list "aws_cloudwatch_log_group" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) [Region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) to query.
  Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
