---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_access_grant"
description: |-
  Lists CloudWatch Omni Access Grant resources.
---

# List Resource: aws_cloudwatchomni_access_grant

Lists CloudWatch Omni Access Grant resources. Service-managed access grants are not listed.

## Example Usage

```terraform
list "aws_cloudwatchomni_access_grant" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
