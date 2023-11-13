---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_site"
description: |-
  Retrieve information about a site.
---

# Data Source:  aws_networkmanager_site

Retrieve information about a site.

## Example Usage

```terraform
data "aws_networkmanager_site" "example" {
  global_network_id = var.global_network_id
  site_id           = var.site_id
}
```

## Argument Reference

* `global_network_id` - (Required) ID of the Global Network of the site to retrieve.
* `site_id` - (Required) ID of the specific site to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the site.
* `description` - Description of the site.
* `location` - Site location as documented below.
* `tags` - Key-value tags for the Site.

The `location` object supports the following:

* `address` - Address of the location.
* `latitude` - Latitude of the location.
* `longitude` - Longitude of the location.
