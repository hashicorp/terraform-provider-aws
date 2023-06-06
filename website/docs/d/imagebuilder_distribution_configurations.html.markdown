---
subcategory: "EC2 Image Builder"
layout: "aws"
page_title: "AWS: aws_imagebuilder_distribution_configurations"
description: |-
    Get information on Image Builder Distribution Configurations.
---

# Data Source: aws_imagebuilder_distribution_configurations

Use this data source to get the ARNs and names of Image Builder Distribution Configurations matching the specified criteria.

## Example Usage

```terraform
data "aws_imagebuilder_distribution_configurations" "example" {
  filter {
    name   = "name"
    values = ["example"]
  }
}
```

## Argument Reference

* `filter` - (Optional) Configuration block(s) for filtering. Detailed below.

## filter Configuration Block

The following arguments are supported by the `filter` configuration block:

* `name` - (Required) Name of the filter field. Valid values can be found in the [Image Builder ListDistributionConfigurations API Reference](https://docs.aws.amazon.com/imagebuilder/latest/APIReference/API_ListDistributionConfigurations.html).
* `values` - (Required) Set of values that are accepted for the given filter field. Results will be selected if any given value matches.

## Attributes Reference

* `arns` - Set of ARNs of the matched Image Builder Distribution Configurations.
* `names` - Set of names of the matched Image Builder Distribution Configurations.
