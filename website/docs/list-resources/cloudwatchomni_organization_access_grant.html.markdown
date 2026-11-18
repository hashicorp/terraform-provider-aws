---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_organization_access_grant"
description: |-
  Lists CloudWatch Omni Organization Access Grant resources.
---

# List Resource: aws_cloudwatchomni_organization_access_grant

Lists CloudWatch Omni Organization Access Grant resources.

## Example Usage

```terraform
list "aws_cloudwatchomni_organization_access_grant" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
