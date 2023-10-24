---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_outposts"
description: |-
  Provides details about multiple Outposts
---

# Data Source: aws_outposts_outposts

Provides details about multiple Outposts.

## Example Usage

```terraform
data "aws_outposts_outposts" "example" {
  site_id = data.aws_outposts_site.id
}
```

## Argument Reference

This data source supports the following arguments:

* `availability_zone` - (Optional) Availability Zone name.
* `availability_zone_id` - (Optional) Availability Zone identifier.
* `site_id` - (Optional) Site identifier.
* `owner_id` - (Optional) AWS Account identifier of the Outpost owner.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of Amazon Resource Names (ARNs).
* `id` - AWS Region.
* `ids` - Set of identifiers.
