---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_regions"
description: |-
    Provides list of all enabled AWS regions
---

# Data Source: aws_regions

`aws_regions` provides list of all enabled AWS regions.

The data source provides list of AWS regions available.
Can be used to filter regions i.e. by Opt-In status or list only regions enabled for current account.
To get details like endpoint and description of each region the data source can be combined with `aws_region`.

## Example Usage

The following example shows how the resource might be used to obtain
the list of the AWS regions configured on the provider.

To list regions enabled for the user:

```hcl
data "aws_regions" "current" {}
```

All the regions regardless of the availability

```hcl
data "aws_regions" "current" {
    all_regions = true
}
```

To see regions that are filtered by `"not-opted-in"` `"all_regions"` need to be set to true 
or probably nothing will be displayed.

```hcl
data "aws_regions" "current" {
    all_regions = true

    filter {
      name   = "opt-in-status"
      values = ["not-opted-in"]
    }
}
```

## Argument Reference

The arguments of this data source act as filters for querying the available
regions. The given filters must match exactly one region whose data will be
exported as attributes.

* `all_regions` - (Optional) If true the source will query all regions regardless of availability.

* `filter` - (Optional) One or more key/value pairs to use as filters. Full reference of valid keys 
can be found [describe-regions in the AWS CLI reference][1].

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `names` - Names of regions that meets the criteria.

[1]: https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-regions.html

