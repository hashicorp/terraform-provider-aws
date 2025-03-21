---
subcategory: "Account Management"
layout: "aws"
page_title: "AWS: aws_account_regions"
description: |-
  Terraform data source for managing an AWS Account Management Regions.
---

# Data Source: aws_account_regions

Terraform data source for managing an AWS Account Management Regions.

## Example Usage

### Basic Usage

```terraform
data "aws_account_regions" "example" {}
```

## Argument Reference


The following arguments are optional:

* `account_id` - (Optional) AWS account ID. Must be a member account in the same organization.
* `region_opt_status_contains` - (Optional) A list of Region statuses (Enabling, Enabled, Disabling, Disabled, Enabled_by_default) to use to filter the list of Regions for a given account.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `regions` - The [regions](#region) for a given account

### region

* `region_name` - The Region code of a given Region
* `region_opt_status` - One of potential statuses a Region can undergo (Enabled, Enabling, Disabled, Disabling, Enabled_By_Default).

