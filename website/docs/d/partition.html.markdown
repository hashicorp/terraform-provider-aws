---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_partition"
description: |-
  Get AWS partition identifier
---

# Data Source: aws_partition

Use this data source to lookup information about the current AWS partition in
which Terraform is working.

## Example Usage

```terraform
data "aws_partition" "current" {}

data "aws_iam_policy_document" "s3_policy" {
  statement {
    sid = "1"

    actions = [
      "s3:ListBucket",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::my-bucket",
    ]
  }
}
```

## Argument Reference

This data source does not support any arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `dns_suffix` - Base DNS domain name for the current partition (e.g., `amazonaws.com` in AWS Commercial, `amazonaws.com.cn` in AWS China).
* `id` - Identifier of the current partition (e.g., `aws` in AWS Commercial, `aws-cn` in AWS China).
* `partition` - Identifier of the current partition (e.g., `aws` in AWS Commercial, `aws-cn` in AWS China).
* `reverse_dns_prefix` - Prefix of service names (e.g., `com.amazonaws` in AWS Commercial, `cn.com.amazonaws` in AWS China).
