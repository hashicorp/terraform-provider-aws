---
subcategory: "Lambda MicroVMs"
layout: "aws"
page_title: "AWS: aws_lambdamicrovms_microvm"
description: |-
  Lists Lambda MicroVMs Micro VM resources.
---

# List Resource: aws_lambdamicrovms_microvm

Lists Lambda MicroVMs Micro VM resources.

## Example Usage

```terraform
list "aws_lambdamicrovms_microvm" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `image_identifier` - (Optional) Filter to list only MicroVMs running the specified image.
* `image_version` - (Optional) Filter to list only MicroVMs running the specified image version.
* `region` - (Optional) Region to query. Defaults to provider region.
