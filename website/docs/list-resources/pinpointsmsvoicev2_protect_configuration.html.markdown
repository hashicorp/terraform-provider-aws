---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_protect_configuration"
description: |-
  Lists AWS End User Messaging SMS Protect Configuration resources.
---

# List Resource: aws_pinpointsmsvoicev2_protect_configuration

Lists AWS End User Messaging SMS Protect Configuration resources.

## Example Usage

```terraform
list "aws_pinpointsmsvoicev2_protect_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
