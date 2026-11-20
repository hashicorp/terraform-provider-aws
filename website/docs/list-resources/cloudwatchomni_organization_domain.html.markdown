---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_organization_domain"
description: |-
  Lists CloudWatch Omni Organization Domain resources.
---

# List Resource: aws_cloudwatchomni_organization_domain

Lists CloudWatch Omni Organization Domain resources.

## Example Usage

```terraform
list "aws_cloudwatchomni_organization_domain" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
