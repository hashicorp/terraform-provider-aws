---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_integration"
description: |-
  Lists Continuum Integration resources.
---

# List Resource: aws_continuum_integration

Lists Continuum Integration resources.

## Example Usage

```terraform
list "aws_continuum_integration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
