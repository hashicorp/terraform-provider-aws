---
subcategory: "Resilience Hub V2"
layout: "aws"
page_title: "AWS: aws_resiliencehubv2_policy"
description: |-
  Lists Resilience Hub V2 Policy resources.
---

# List Resource: aws_resiliencehubv2_policy

Lists Resilience Hub V2 Policy resources.

## Example Usage

```terraform
list "aws_resiliencehubv2_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
