---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_reserved_instance_offering"
description: |-
  Information about a single RDS Reserved Instance Offering.
---

# Data Source: aws_rds_reserved_instance_offering

Information about a single RDS Reserved Instance Offering.

## Example Usage

```terraform
data "aws_rds_reserved_instance_offering" "test" {
  db_instance_class   = "db.t2.micro"
  duration            = 31536000
  multi_az            = false
  offering_type       = "All Upfront"
  product_description = "mysql"
}
```

## Argument Reference

This data source supports the following arguments:

* `db_instance_class` - (Required) DB instance class for the reserved DB instance.
* `duration` - (Required) Duration of the reservation in years or seconds. Valid values are `1`, `3`, `31536000`, `94608000`
* `multi_az` - (Required) Whether the reservation applies to Multi-AZ deployments.
* `offering_type` - (Required) Offering type of this reserved DB instance. Valid values are `No Upfront`, `Partial Upfront`, `All Upfront`.
* `product_description` - (Required) Description of the reserved DB instance.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the reservation. Same as `offering_id`.
* `currency_code` - Currency code for the reserved DB instance.
* `fixed_price` - Fixed price charged for this reserved DB instance.
* `offering_id` - Unique identifier for the reservation.
