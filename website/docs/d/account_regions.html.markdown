---
subcategory: "Account Management"
layout: "aws"
page_title: "AWS: aws_account_regions"
description: |-

 Terraform data source for managing AWS Account Management Regions. 
---

# Data Source: aws_account_regions

The `aws_account_regions` data source lets you query AWS region information for any account in your AWS Organization, not just the current account. It uses the AWS Account REST Service to show all regions, including those that are enabled, disabled, or in the process of being enabled or disabled. You can list regions for any organization account, see all possible region opt-in statuses (Enabled, Enabling, Disabled, Disabling, Enabled_By_Default), and check which regions are being enabled or disabled, not just those that are currently available.

This is more comprehensive than the [aws_regions](./region.html.markdown) data source, which only uses the EC2 REST service and is limited to the current account and a subset of region statuses.

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
