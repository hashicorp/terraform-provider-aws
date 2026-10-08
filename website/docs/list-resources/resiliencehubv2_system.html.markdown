---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_system"
description: |-
  Lists Resilience Hub V2 System resources.
---

# List Resource: aws_resiliencehubv2_system

Lists Resilience Hub V2 System resources.

## Example Usage

```terraform
list "aws_resiliencehubv2_system" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
