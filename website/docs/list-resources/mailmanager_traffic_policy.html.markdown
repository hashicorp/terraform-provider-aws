---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_traffic_policy"
description: |-
  Lists SES Mail Manager Traffic Policy resources.
---

# List Resource: aws_mailmanager_traffic_policy

Lists SES Mail Manager Traffic Policy resources.

## Example Usage

```terraform
list "aws_mailmanager_traffic_policy" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
