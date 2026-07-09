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
    origination_identity = "phone-abcdef0123456789abcdef0123456789"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `origination_identity` - (Required) Origination identity to list keywords for. Value is the ID or ARN of a phone number, pool, or sender ID.
* `region` - (Optional) Region to query. Defaults to provider region.
