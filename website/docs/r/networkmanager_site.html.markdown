---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_site"
description: |-
  Creates a site in a global network.
---

# Resource: aws_networkmanager_site

Creates a site in a global network.

## Example Usage

```terraform
resource "aws_networkmanager_global_network" "example" {
}

resource "aws_networkmanager_site" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `global_network_id` - (Required) The ID of the Global Network to create the site in.
* `description` - (Optional) Description of the Site.
* `location` - (Optional) The site location as documented below.
* `tags` - (Optional) Key-value tags for the Site. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `location` object supports the following:

* `address` - (Optional) Address of the location.
* `latitude` - (Optional) Latitude of the location.
* `longitude` - (Optional) Longitude of the location.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Site Amazon Resource Name (ARN)
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_site` using the site ARN. For example:

```terraform
import {
  to = aws_networkmanager_site.example
  id = "arn:aws:networkmanager::123456789012:site/global-network-0d47f6t230mz46dy4/site-444555aaabbb11223"
}
```

Using `terraform import`, import `aws_networkmanager_site` using the site ARN. For example:

```console
% terraform import aws_networkmanager_site.example arn:aws:networkmanager::123456789012:site/global-network-0d47f6t230mz46dy4/site-444555aaabbb11223
```
