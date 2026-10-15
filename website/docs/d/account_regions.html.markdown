---
subcategory: "Account Management"
layout: "aws"
page_title: "AWS: aws_account_regions"
description: |-

 Terraform data source for managing AWS Account Management Regions. 
---

# Data Source: aws_account_regions

The `aws_account_regions` data source lets you query AWS region information for any account in your AWS Organization. It uses the AWS Account REST Service to show all regions, including those that are enabled, disabled, or in the process of being enabled or disabled. You can list regions for any organization account, see all possible region opt-in statuses (`ENABLED`, `ENABLING`, `DISABLING`, `DISABLED`, `ENABLED_BY_DEFAULT`), and check which regions are being enabled or disabled.

This is more comprehensive than the [aws_regions](./region.html.markdown) data source, which only uses the EC2 REST service and is limited to the current account and a subset of region statuses.

## Example Usage

### Basic Usage

```terraform
data "aws_account_regions" "example" {}
```

## Argument Reference

The following arguments are optional:

* `account_id` - (Optional) AWS account ID. Must be a member account in the same organization.
* `region_opt_status_contains` - (Optional) A list of region opt-in statuses to filter the results. Valid values are `ENABLED`, `ENABLING`, `DISABLING`, `DISABLED`, and `ENABLED_BY_DEFAULT`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `regions` - The [regions](#region) for a given account

### region

* `region_name` - The Region code of a given Region
* `region_opt_status` - The opt-in status of the region. Possible values are `ENABLED`, `ENABLING`, `DISABLING`, `DISABLED`, and `ENABLED_BY_DEFAULT`.
