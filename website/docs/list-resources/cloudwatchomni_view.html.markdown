---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_view"
description: |-
  Lists CloudWatch Omni View resources.
---

# List Resource: aws_cloudwatchomni_view

Lists CloudWatch Omni View resources. Only `USER` views are listed; `MANAGED` views are provisioned by AWS and cannot be managed with the `aws_cloudwatchomni_view` resource.

## Example Usage

```terraform
list "aws_cloudwatchomni_view" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
