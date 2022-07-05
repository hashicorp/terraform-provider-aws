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

```terraform
data "aws_outposts_assets" "example" {
  arn = data.aws_outposts_outpost.example.arn
}

```

## Argument Reference

The following arguments are required:

* `arn` - (Required) Outpost ARN.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `asset_ids` - A list of all the subnet ids found. This data source will fail if none are found.
