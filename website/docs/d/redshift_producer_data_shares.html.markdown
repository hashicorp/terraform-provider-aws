---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_producer_data_shares"
description: |-
  Terraform data source for managing AWS Redshift Producer Data Shares.
---

# Data Source: aws_redshift_producer_data_shares

Terraform data source for managing AWS Redshift Producer Data Shares.

## Example Usage

### Basic Usage

```terraform
data "aws_redshift_producer_data_shares" "example" {
  producer_arn = ""
}
```

## Argument Reference

The following arguments are required:

* `producer_arn` - (Required) Amazon Resource Name (ARN) of the producer namespace that returns in the list of datashares.

The following arguments are optional:

* `status` - (Optional) Status of a datashare in the producer. Valid values are `ACTIVE`, `AUTHORIZED`, `PENDING_AUTHORIZATION`, `DEAUTHORIZED`, and `REJECTED`. Omit this argument to return all statuses.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Producer ARN.
* `data_shares` - An array of all data shares in the producer. See [`data_shares`](#data_shares-attribute-reference) below.

### `data_shares` Attribute Reference

* `data_share_arn` - ARN (Amazon Resource Name) of the data share.
* `managed_by` - Identifier of a datashare to show its managing entity.
* `producer_arn` - ARN (Amazon Resource Name) of the producer.
