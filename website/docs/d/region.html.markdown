---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_region"
description: |-
    Provides details about a specific service region
---

# Data Source: aws_region

`aws_region` provides details about a specific AWS region.

As well as validating a given region name this resource can be used to
discover the name of the region configured within the provider. The latter
can be useful in a child module which is inheriting an AWS provider
configuration from its parent module.

## Example Usage

The following example shows how the resource might be used to obtain
the name of the AWS region configured on the provider.

```terraform
data "aws_region" "current" {}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Optional) Full name of the region to select.
* `endpoint` - (Optional) EC2 endpoint of the region to select.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Region's description in this format: "Location (Region name)".
