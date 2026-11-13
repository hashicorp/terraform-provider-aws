---
subcategory: "ACM (Certificate Manager)"
layout: "aws"
page_title: "AWS: aws_acm_acme_endpoint"
description: |-
  Lists ACM (Certificate Manager) ACME Endpoint resources.
---

# List Resource: aws_acm_acme_endpoint

Lists ACM (Certificate Manager) ACME Endpoint resources.

## Example Usage

```terraform
list "aws_acm_acme_endpoint" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
