---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_asset"
description: |-
  Information about hardware assets in an Outpost.
---

# Data Source: aws_outposts_asset

Information about a specific hardware asset in an Outpost.

## Example Usage

```terraform
data "aws_outposts_assets" "example" {
  arn = data.aws_outposts_outpost.example.arn
}

data "aws_outposts_asset" "example" {
  count    = length(data.aws_outposts_assets.example.asset_ids)
  arn      = data.aws_outposts_outpost.example.arn
  asset_id = element(data.aws_outposts_assets.this.asset_ids, count.index)
}

```

## Argument Reference

The following arguments are required:

* `arn` - (Required) Outpost ARN.
* `asset_id` - (Required) ID of the asset.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `asset_type` - Type of the asset.
* `host_id` - Host ID of the Dedicated Hosts on the asset, if a Dedicated Host is provisioned.
* `rack_elevation` - Position of an asset in a rack measured in rack units.
* `rack_id` - Rack ID of the asset.
