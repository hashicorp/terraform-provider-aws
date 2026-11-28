---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_principal_association"
description: |-
  Lists RAM (Resource Access Manager) Principal Association resources.
---

# List Resource: aws_ram_principal_association

Lists RAM (Resource Access Manager) Principal Association resources.

## Example Usage

```terraform
list "aws_ram_principal_association" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
