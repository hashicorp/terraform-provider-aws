---
subcategory: "Meta Data Sources"
layout: "aws"
page_title: "AWS: aws_region"
description: |-
    Provides details about a specific AWS Region
---

# Data Source: aws_region

`aws_region` provides details about a specific AWS Region.

As well as validating a given Region name this resource can be used to
discover the name of the Region configured within the provider. The latter
can be useful in a child module which is inheriting an AWS provider
configuration from its parent module.

## Example Usage

The following example shows how the resource might be used to obtain
the name of the AWS Region configured on the provider.

```terraform
data "aws_region" "current" {}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
regions. The given filters must match exactly one region whose data will be
exported as attributes.

* `endpoint` - (Optional) EC2 endpoint of the region to select.
* `name` - (Optional, **Deprecated**) Full name of the region to select. Use `region` instead.
* `region` - (Optional) Full name of the region to select (e.g. `us-east-1`)

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Region's description in this format: "Location (Region name)".
