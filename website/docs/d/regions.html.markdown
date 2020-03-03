---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_regions"
description: |-
    Provides list of all enabled AWS regions
---

# Data Source: aws_region

`aws_regions` provides list of all enabled AWS regions.

The data source provides list of AWS regions available.
Can be used to filter regions by Opt-In status or list only regions enabled for current account.
To get details like endpoint and description of each region the data source can be combined with `aws_region`.

## Example Usage

The following example shows how the resource might be used to obtain
the list of the AWS regions configured on the provider.

Regions enabled for the user

```hcl
data "aws_regions" "current" {}
```

All the regions regardless of the availability

```hcl
data "aws_regions" "current" {
    all_regions = true
}
```

To see regions that are `"not-opted-in"` `"all_regions"` need to be set to true 
or nothing will be displayed.

```hcl
data "aws_regions" "current" {
    all_regions = true
    opt_in_status = "not-opted-in"
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
regions. The given filters must match exactly one region whose data will be
exported as attributes.

* `all_regions` - (Optional) If true the source will query all regions regardless of availability.

* `opt_in_status` - (Optional) Filter the list of regions according to op-in-status filter. Can be one of: `"opt-in-not-required"`, `"opted-in"` or `"not-opted-in"`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `names` - Names of regions that meets the criteria.

