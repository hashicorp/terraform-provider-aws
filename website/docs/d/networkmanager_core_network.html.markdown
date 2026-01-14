---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_core_network"
description: |-
  Provides details about an AWS Network Manager Core Network.
---

# Data Source: aws_networkmanager_core_network

Provides details about an AWS Network Manager Core Network.

## Example Usage

### Basic Usage

```terraform
data "aws_networkmanager_core_network" "example" {
  core_network_id = "core-network-0123456789abcdef0"
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required) ID of the core network.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the core network.
* `created_at` - Timestamp when the core network was created.
* `description` - Description of the core network.
* `edges` - Edges within a core network. See [`edges` Block](#edges-block) for details.
* `global_network_id` - ID of the global network that the core network is a part of.
* `id` - ID of the core network.
* `network_function_groups` - Network function groups associated with the core network. See [`network_function_groups` Block](#network_function_groups-block) for details.
* `segments` - Segments within a core network. See [`segments` Block](#segments-block) for details.
* `state` - Current state of the core network.
* `tags` - Map of tags assigned to the resource.

### `edges` Block

The `edges` configuration block exports the following attributes:

* `asn` - ASN of the core network edge.
* `edge_location` - AWS region where the edge is located.
* `inside_cidr_blocks` - Inside IP addresses used for core network edges.

### `network_function_groups` Block

The `network_function_groups` configuration block exports the following attributes:

* `edge_locations` - Core network edge locations.
* `name` - Name of the network function group.
* `segments` - Segments associated with the network function group. See [`network_function_groups.segments` Block](#network_function_groupssegments-block) for details.

### `network_function_groups.segments` Block

The `network_function_groups.segments` configuration block exports the following attributes:

* `send_to` - List of segments associated with the `send-to` action.
* `send_via` - List of segments associated with the `send-via` action.

### `segments` Block

The `segments` configuration block exports the following attributes:

* `edge_locations` -  AWS regions where the edges are located.
* `name` - Name of the core network segment.
* `shared_segments` - Shared segments of the core network.
