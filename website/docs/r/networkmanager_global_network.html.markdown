---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_global_network"
description: |-
  Manages a Network Manager Global Network.
---

# Resource: aws_networkmanager_global_network

Manages a Network Manager Global Network.

Use this resource to create and manage a global network, which is a single private network that acts as the high-level container for your network objects.

## Example Usage

```terraform
resource "aws_networkmanager_global_network" "example" {
  description = "example"
}
```

## Argument Reference

The following arguments are optional:

* `description` - (Optional) Description of the Global Network.
* `tags` - (Optional) Key-value tags for the Global Network. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Global Network ARN.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)
* `update` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_global_network` using the global network ID. For example:

```terraform
import {
  to = aws_networkmanager_global_network.example
  id = "global-network-0d47f6t230mz46dy4"
}
```

Using `terraform import`, import `aws_networkmanager_global_network` using the global network ID. For example:

```console
% terraform import aws_networkmanager_global_network.example global-network-0d47f6t230mz46dy4
```
