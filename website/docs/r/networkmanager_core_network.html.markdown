---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_core_network"
description: |-
  Provides a core network resource.
---

# Resource: aws_networkmanager_core_network

Provides a core network resource.

## Example Usage

### Basic

```terraform
resource "aws_networkmanager_core_network" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
}
```

### With description

```terraform
resource "aws_networkmanager_core_network" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
  description       = "example"
}
```

### With policy document

```terraform
resource "aws_networkmanager_core_network" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
  policy_document   = data.aws_networkmanager_core_network_policy_document.example.json
}
```

### With tags

```terraform
resource "aws_networkmanager_core_network" "example" {
  global_network_id = aws_networkmanager_global_network.example.id

  tags = {
    "hello" = "world"
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) Description of the Core Network.
* `global_network_id` - (Required) The ID of the global network that a core network will be a part of.
* `policy_document` - (Optional) Policy document for creating a core network. Note that updating this argument will result in the new policy document version being set as the `LATEST` and `LIVE` policy document. Refer to the [Core network policies documentation](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-policy-change-sets.html) for more information.
* `tags` - (Optional) Key-value tags for the Core Network. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)
* `update` - (Default `30m`)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Core Network Amazon Resource Name (ARN).
* `created_at` - Timestamp when a core network was created.
* `edges` - One or more blocks detailing the edges within a core network. [Detailed below](#edges).
* `id` - Core Network ID.
* `segments` - One or more blocks detailing the segments within a core network. [Detailed below](#segments).
* `state` - Current state of a core network.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `edges`

The `edges` configuration block supports the following arguments:

* `asn` - ASN of a core network edge.
* `edge_location` - Region where a core network edge is located.
* `inside_cidr_blocks` - Inside IP addresses used for core network edges.

### `segments`

The `segments` configuration block supports the following arguments:

* `edge_locations` - Regions where the edges are located.
* `name` - Name of a core network segment.
* `shared_segments` - Shared segments of a core network.

## Import

`aws_networkmanager_core_network` can be imported using the core network ID, e.g.

```
$ terraform import aws_networkmanager_core_network.example core-network-0d47f6t230mz46dy4
```
