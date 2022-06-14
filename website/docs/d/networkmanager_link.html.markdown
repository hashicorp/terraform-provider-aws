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

* `global_network_id` - (Required) The ID of the Global Network of the link to retrieve.
* `link_id` - (Required) The id of the specific link to retrieve.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the link.
* `bandwidth` - The upload speed and download speed of the link as documented below
* `description` - The description of the link.
* `provider_name` - The provider of the link.
* `site_id` - The ID of the site.
* `tags` - Key-value tags for the link.
* `type` - The type of the link.

The `bandwidth` object supports the following:

* `download_speed` - Download speed in Mbps.
* `upload_speed` - Upload speed in Mbps.
