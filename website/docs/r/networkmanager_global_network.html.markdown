---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_global_network"
description: |-
  Provides a global network resource.
---

# Resource: aws_networkmanager_global_network

Provides a global network resource.

## Example Usage

```terraform
resource "aws_networkmanager_global_network" "example" {
  description = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) Description of the Global Network.
* `tags` - (Optional) Key-value tags for the Global Network. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Global Network Amazon Resource Name (ARN)
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Import `aws_networkmanager_global_network` using the global network ID. For example:

```
$ terraform import aws_networkmanager_global_network.example global-network-0d47f6t230mz46dy4
```
