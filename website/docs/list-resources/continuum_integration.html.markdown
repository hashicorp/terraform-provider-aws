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

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
