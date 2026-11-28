---
subcategory: "Lambda MicroVMs"
layout: "aws"
page_title: "AWS: aws_lambdamicrovms_image"
description: |-
  Lists Lambda MicroVM image resources.
---

# List Resource: aws_lambdamicrovms_image

Lists Lambda MicroVM image resources.

## Example Usage

```terraform
list "aws_lambdamicrovms_image" "example" {
  provider = aws

  config {
    name_filter = "example"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `name_filter` - (Optional) Filters images whose name contains the specified string.
* `region` - (Optional) Region to query. Defaults to provider region.
