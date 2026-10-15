---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_link"
description: |-
  Manages a Network Manager link.
---

# Resource: aws_networkmanager_link

Manages a Network Manager link. Use this resource to create a link for a site.

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

The following arguments are required:

* `bandwidth` - (Required) Upload speed and download speed in Mbps. [See below](#bandwidth).
* `global_network_id` - (Required) ID of the global network.
* `site_id` - (Required) ID of the site.

The following arguments are optional:

* `description` - (Optional) Description of the link.
* `provider_name` - (Optional) Provider of the link.
* `tags` - (Optional) Key-value tags for the link. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `type` - (Optional) Type of the link.

### bandwidth

* `download_speed` - (Optional) Download speed in Mbps.
* `upload_speed` - (Optional) Upload speed in Mbps.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Link ARN.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)
* `update` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_link` using the link ARN. For example:

```terraform
import {
  to = aws_networkmanager_link.example
  id = "arn:aws:networkmanager::123456789012:link/global-network-0d47f6t230mz46dy4/link-444555aaabbb11223"
}
```

Using `terraform import`, import `aws_networkmanager_link` using the link ARN. For example:

```console
% terraform import aws_networkmanager_link.example arn:aws:networkmanager::123456789012:link/global-network-0d47f6t230mz46dy4/link-444555aaabbb11223
```
