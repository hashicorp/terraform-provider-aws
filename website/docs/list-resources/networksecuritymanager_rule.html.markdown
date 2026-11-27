---
subcategory: "Network Security Manager"
layout: "aws"
page_title: "AWS: aws_networksecuritymanager_rule"
description: |-
  Lists AWS Network Security Manager Rule resources.
---

# List Resource: aws_networksecuritymanager_rule

Lists AWS Network Security Manager Rule resources.

## Example Usage

```terraform
list "aws_networksecuritymanager_rule" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
