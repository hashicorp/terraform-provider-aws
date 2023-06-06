---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_link"
description: |-
  Creates a link for a site.
---

# Resource: aws_networkmanager_link

Creates a link for a site.

## Example Usage

```terraform
resource "aws_networkmanager_link" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
  site_id           = aws_networkmanager_site.example.id

  bandwidth {
    upload_speed   = 10
    download_speed = 50
  }

  provider_name = "MegaCorp"
}
```

## Argument Reference

The following arguments are supported:

* `bandwidth` - (Required) The upload speed and download speed in Mbps. Documented below.
* `description` - (Optional) A description of the link.
* `global_network_id` - (Required) The ID of the global network.
* `provider_name` - (Optional) The provider of the link.
* `site_id` - (Required) The ID of the site.
* `tags` - (Optional) Key-value tags for the link. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) The type of the link.

The `bandwidth` object supports the following:

* `download_speed` - (Optional) Download speed in Mbps.
* `upload_speed` - (Optional) Upload speed in Mbps.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Link Amazon Resource Name (ARN).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

`aws_networkmanager_link` can be imported using the link ARN, e.g.

```
$ terraform import aws_networkmanager_link.example arn:aws:networkmanager::123456789012:link/global-network-0d47f6t230mz46dy4/link-444555aaabbb11223
```
