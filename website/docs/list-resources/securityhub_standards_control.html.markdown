---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_standards_control"
description: |-
  Lists Security Hub Standards Control resources.
---

# List Resource: aws_securityhub_standards_control

Lists Security Hub Standards Control resources.

## Example Usage

```terraform
list "aws_securityhub_standards_control" "example" {
  provider = aws

  config {
    standards_subscription_arn = "arn:aws:securityhub:us-west-2:1234567890:subscription/cis-aws-foundations-benchmark/v/1.2.0"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
* `standards_subscription_arn` - (Required) ARN that represents your subscription to a supported standard. Use the `aws_securityhub_enabled_standards` data source to get the subscription ARNs of the standards you have enabled.
