---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outpost"
description: |-
  Provides details about an Outposts Outpost
---

# Data Source: aws_outposts_outpost

Provides details about an Outposts Outpost.

## Example Usage

```terraform
data "aws_outposts_outpost" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) Identifier of the Outpost.
* `name` - (Optional) Name of the Outpost.
* `arn` - (Optional) ARN.
* `owner_id` - (Optional) AWS Account identifier of the Outpost owner.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `availability_zone` - Availability Zone name.
* `availability_zone_id` - Availability Zone identifier.
* `description` - The description of the Outpost.
* `lifecycle_status` - The life cycle status.
* `site_arn` - The Amazon Resource Name (ARN) of the site.
* `site_id` - The ID of the site.
* `supported_hardware_type` - The hardware type.
* `tags` - The Outpost tags.
