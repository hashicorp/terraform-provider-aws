---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_pool"
description: |-
  Lists AWS End User Messaging SMS Pool resources.
---

# List Resource: aws_pinpointsmsvoicev2_pool

Lists AWS End User Messaging SMS Pool resources.

## Example Usage

```terraform
list "aws_pinpointsmsvoicev2_pool" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
