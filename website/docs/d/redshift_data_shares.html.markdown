---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_data_shares"
description: |-
  Terraform data source for managing AWS Redshift Data Shares.
---

# Data Source: aws_redshift_data_shares

Terraform data source for managing AWS Redshift Data Shares.

## Example Usage

### Basic Usage

```terraform
data "aws_redshift_data_shares" "example" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS region.
* `data_shares` - An array of all data shares in the current region. See [`data_shares`](#data_shares-attribute-reference) below.

### `data_shares` Attribute Reference

* `data_share_arn` - ARN (Amazon Resource Name) of the data share.
* `managed_by` - Identifier of a datashare to show its managing entity.
* `producer_arn` - ARN (Amazon Resource Name) of the producer.
