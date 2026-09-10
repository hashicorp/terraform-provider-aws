---
subcategory: "SES Mail Manager"
layout: "aws"
page_title: "AWS: aws_mailmanager_ingress_point"
description: |-
  Lists SES Mail Manager Ingress Point resources.
---

# List Resource: aws_mailmanager_ingress_point

Lists SES Mail Manager Ingress Point resources.

## Example Usage

```terraform
list "aws_mailmanager_ingress_point" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
