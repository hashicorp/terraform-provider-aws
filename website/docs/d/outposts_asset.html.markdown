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
* `asset_id` - (Required) The ID of the asset.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `asset_type` - The type of the asset.
* `host_id` - The host ID of the Dedicated Hosts on the asset, if a Dedicated Host is provisioned.
* `rack_id` - The rack ID of the asset.
