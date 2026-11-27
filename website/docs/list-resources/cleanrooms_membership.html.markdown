---
subcategory: "Clean Rooms"
layout: "aws"
page_title: "AWS: aws_cleanrooms_membership"
description: |-
  Lists Clean Rooms Membership resources.
---

# List Resource: aws_cleanrooms_membership

Lists Clean Rooms Membership resources.

## Example Usage

```terraform
list "aws_cleanrooms_membership" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
