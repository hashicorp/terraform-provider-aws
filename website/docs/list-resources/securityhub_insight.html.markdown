---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_insight"
description: |-
  Lists Security Hub Insight resources.
---

# List Resource: aws_securityhub_insight

Lists Security Hub Insight resources.

## Example Usage

```terraform
list "aws_securityhub_insight" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
