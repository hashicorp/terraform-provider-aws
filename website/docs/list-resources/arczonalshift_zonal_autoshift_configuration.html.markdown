---
subcategory: "ARC (Application Recovery Controller) Zonal Shift"
layout: "aws"
page_title: "AWS: aws_arczonalshift_zonal_autoshift_configuration"
description: |-
  Lists ARC (Application Recovery Controller) Zonal Shift Zonal Autoshift Configuration resources.
---

# List Resource: aws_arczonalshift_zonal_autoshift_configuration

Lists ARC (Application Recovery Controller) Zonal Shift Zonal Autoshift Configuration resources.

## Example Usage

```terraform
list "aws_arczonalshift_zonal_autoshift_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
