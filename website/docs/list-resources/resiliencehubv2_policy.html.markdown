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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
