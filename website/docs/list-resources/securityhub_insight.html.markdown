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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
