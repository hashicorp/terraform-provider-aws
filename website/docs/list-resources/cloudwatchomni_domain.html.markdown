---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_domain"
description: |-
  Lists CloudWatch Omni Domain resources.
---

# List Resource: aws_cloudwatchomni_domain

Lists CloudWatch Omni Domain resources.

## Example Usage

```terraform
list "aws_cloudwatchomni_domain" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
