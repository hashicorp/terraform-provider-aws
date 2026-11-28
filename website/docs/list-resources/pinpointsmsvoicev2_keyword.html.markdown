---
subcategory: "End User Messaging SMS"
layout: "aws"
page_title: "AWS: aws_pinpointsmsvoicev2_keyword"
description: |-
  Lists AWS End User Messaging SMS Keyword resources.
---

# List Resource: aws_pinpointsmsvoicev2_keyword

Lists AWS End User Messaging SMS Keyword resources.

## Example Usage

```terraform
list "aws_pinpointsmsvoicev2_keyword" "example" {
  provider = aws

  config {
    origination_identity_arn = "arn:aws:sms-voice:us-east-1:123456789012:phone-number/phone-abcdef0123456789abcdef0123456789"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `origination_identity_arn` - (Required) ARN of the origination identity (phone number or pool) to list keywords for.
* `region` - (Optional) Region to query. Defaults to provider region.
