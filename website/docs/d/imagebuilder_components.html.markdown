---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_components"
description: |-
    Get information on Image Builder Components.
---

# Data Source: aws_imagebuilder_components

Use this data source to get the ARNs and names of Image Builder Components matching the specified criteria.

## Example Usage

```terraform
data "aws_imagebuilder_components" "example" {
  owner = "Self"

  filter {
    name   = "platform"
    values = ["Linux"]
  }
}
```

## Argument Reference

* `owner` - (Optional) Owner of the image recipes. Valid values are `Self`, `Shared` and `Amazon`. Defaults to `Self`.
* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) Name of the filter field. Valid values can be found in the [Image Builder ListComponents API Reference](https://docs.aws.amazon.com/imagebuilder/latest/APIReference/API_ListComponents.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

* `arns` - Set of ARNs of the matched Image Builder Components.
* `names` - Set of names of the matched Image Builder Components.
