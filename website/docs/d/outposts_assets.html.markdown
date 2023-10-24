---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_assets"
description: |-
  Information about hardware assets in an Outpost.
---

# Data Source: aws_outposts_assets

Information about hardware assets in an Outpost.

## Example Usage

### Basic

```terraform
data "aws_outposts_assets" "example" {
  arn = data.aws_outposts_outpost.example.arn
}
```

### With Host ID Filter

```terraform
data "aws_outposts_assets" "example" {
  arn            = data.aws_outposts_outpost.example.arn
  host_id_filter = ["h-x38g5n0yd2a0ueb61"]
}
```

### With Status ID Filter

```terraform
data "aws_outposts_assets" "example" {
  arn              = data.aws_outposts_outpost.example.arn
  status_id_filter = ["ACTIVE"]
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Required) Outpost ARN.
* `host_id_filter` - (Optional) Filters by list of Host IDs of a Dedicated Host.
* `status_id_filter` - (Optional) Filters by list of state status. Valid values: "ACTIVE", "RETIRING".

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `asset_ids` - List of all the asset ids found. This data source will fail if none are found.
