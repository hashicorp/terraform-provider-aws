---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_service"
description: |-
  Lists Resilience Hub V2 Service resources.
---

# List Resource: aws_resiliencehubv2_service

Lists Resilience Hub V2 Service resources.

## Example Usage

```terraform
list "aws_resiliencehubv2_service" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
