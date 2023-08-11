---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_link"
description: |-
  Retrieve information about a link.
---

# Data Source:  aws_networkmanager_link

Retrieve information about a link.

## Example Usage

```terraform
data "aws_networkmanager_link" "example" {
  global_network_id = var.global_network_id
  link_id           = var.link_id
}
```

## Argument Reference

* `global_network_id` - (Required) ID of the Global Network of the link to retrieve.
* `link_id` - (Required) ID of the specific link to retrieve.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the link.
* `bandwidth` - Upload speed and download speed of the link as documented below
* `description` - Description of the link.
* `provider_name` - Provider of the link.
* `site_id` - ID of the site.
* `tags` - Key-value tags for the link.
* `type` - Type of the link.

The `bandwidth` object supports the following:

* `download_speed` - Download speed in Mbps.
* `upload_speed` - Upload speed in Mbps.
