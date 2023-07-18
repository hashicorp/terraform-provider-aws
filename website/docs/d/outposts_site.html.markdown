---
subcategory: "Outposts"
layout: "aws"
page_title: "AWS: aws_outposts_site"
description: |-
  Provides details about an Outposts Site
---

# Data Source: aws_outposts_site

Provides details about an Outposts Site.

## Example Usage

```terraform
data "aws_outposts_site" "example" {
  name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Optional) Identifier of the Site.
* `name` - (Optional) Name of the Site.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `account_id` - AWS Account identifier.
* `description` - Description.
