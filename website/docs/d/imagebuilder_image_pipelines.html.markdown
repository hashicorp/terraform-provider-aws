---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_image_pipelines"
description: |-
    Get information on Image Builder Image Pipelines.
---

# Data Source: aws_imagebuilder_image_pipelines

Use this data source to get the ARNs and names of Image Builder Image Pipelines matching the specified criteria.

## Example Usage

```terraform
data "aws_imagebuilder_image_pipelines" "example" {
  filter {
    name   = "name"
    values = ["example"]
  }
}
```

## Argument Reference

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

### filter Configuration Block

The `filter` configuration block supports the following arguments:

* `name` - (Required) Name of the filter field. Valid values can be found in the [Image Builder ListImagePipelines API Reference](https://docs.aws.amazon.com/imagebuilder/latest/APIReference/API_ListImagePipelines.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of ARNs of the matched Image Builder Image Pipelines.
* `names` - Set of names of the matched Image Builder Image Pipelines.
