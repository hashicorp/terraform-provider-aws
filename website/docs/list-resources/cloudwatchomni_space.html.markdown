---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_space"
description: |-
  Lists CloudWatch Omni Space resources.
---

# List Resource: aws_cloudwatchomni_space

Lists CloudWatch Omni Space resources.

## Example Usage

```terraform
list "aws_cloudwatchomni_space" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
