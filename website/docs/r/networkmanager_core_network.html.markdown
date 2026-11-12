---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_core_network"
description: |-
  Manages a Network Manager Core Network.
---

# Resource: aws_networkmanager_core_network

Manages a Network Manager Core Network.

Use this resource to create and manage a core network within a global network.

## Example Usage

### Basic

Manage the core network policy directly with `policy_document`.

```terraform
resource "aws_networkmanager_global_network" "example" {}

data "aws_networkmanager_core_network_policy_document" "example" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
    }
  }

  segments {
    name = "segment"
  }
}

resource "aws_networkmanager_core_network" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
  policy_document   = data.aws_networkmanager_core_network_policy_document.example.json
}
```

### With Base Policy Document

A base policy is only needed when the final policy references resources that depend on the core network, which would otherwise create a dependency cycle in Terraform. The main case is a policy that references a [prefix list association](networkmanager_prefix_list_association.html) alias; static routes to attachment IDs are the other. Create the core network with `create_base_policy` and `base_policy_document`, then apply the final policy with the [`aws_networkmanager_core_network_policy_attachment` resource](networkmanager_core_network_policy_attachment.html) once those resources exist. Do not combine `policy_document` with that resource on the same core network.

```terraform
resource "aws_networkmanager_global_network" "example" {}

data "aws_networkmanager_core_network_policy_document" "base" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
    }
  }

  segments {
    name = "segment"
  }
}

resource "aws_networkmanager_core_network" "example" {
  global_network_id    = aws_networkmanager_global_network.example.id
  create_base_policy   = true
  base_policy_document = data.aws_networkmanager_core_network_policy_document.base.json
}

resource "aws_ec2_managed_prefix_list" "example" {
  name           = "example"
  address_family = "IPv4"
  max_entries    = 5

  entry {
    cidr = "10.0.0.0/8"
  }
}

resource "aws_networkmanager_prefix_list_association" "example" {
  core_network_id   = aws_networkmanager_core_network.example.id
  prefix_list_arn   = aws_ec2_managed_prefix_list.example.arn
  prefix_list_alias = "exampleprefixlist"
}

data "aws_networkmanager_core_network_policy_document" "example" {
  version = "2025.11"

  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
    }
  }

  segments {
    name = "segment"
  }

  routing_policies {
    routing_policy_name      = "allowprefixlist"
    routing_policy_direction = "inbound"
    routing_policy_number    = 100

    routing_policy_rules {
      rule_number = 1

      rule_definition {
        match_conditions {
          type  = "prefix-in-prefix-list"
          value = aws_networkmanager_prefix_list_association.example.prefix_list_alias
        }

        action {
          type = "allow"
        }
      }
    }
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}
```

## Argument Reference

The following arguments are required:

* `global_network_id` - (Required) ID of the global network that a core network will be a part of.

The following arguments are optional:

* `base_policy_document` - (Optional, conflicts with `base_policy_regions` and `policy_document`) Base policy document, applied and set to `LIVE` on create when `create_base_policy` is `true`. See [With Base Policy Document](#with-base-policy-document).
* `base_policy_regions` - (Optional, conflicts with `base_policy_document` and `policy_document`) List of regions to add to the base policy. The base policy created by setting the `create_base_policy` argument to `true` requires one or more regions to be set in the `edge-locations`, `location` key. If `base_policy_regions` is not specified, the region used in the base policy defaults to the region specified in the `provider` block.
* `create_base_policy` - (Optional, conflicts with `policy_document`) Whether to create and apply a base policy on create or update, using `base_policy_document` or `base_policy_regions`. Only needed when the final policy references prefix list associations or attachment IDs. See [With Base Policy Document](#with-base-policy-document).
* `description` - (Optional) Description of the core network.
* `policy_document` - (Optional, conflicts with `base_policy_document`, `base_policy_regions`, and `create_base_policy`) Core network policy document, applied and set to `LIVE`. Preferred way to manage the core network policy. Removing this argument stops managing the policy with this resource; the current `LIVE` policy is not changed. Refer to the [Core network policies documentation](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-policy-change-sets.html) for more information.
* `tags` - (Optional) Key-value tags for the core network. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Core Network ARN.
* `created_at` - Timestamp when a core network was created.
* `edges` - One or more blocks detailing the edges within a core network. [Detailed below](#edges-block).
* `id` - Core Network ID.
* `segments` - One or more blocks detailing the segments within a core network. [Detailed below](#segments-block).
* `state` - Current state of a core network.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `edges` Block

The `edges` configuration block supports the following arguments:

* `asn` - ASN of a core network edge.
* `edge_location` - Region where a core network edge is located.
* `inside_cidr_blocks` - Inside IP addresses used for core network edges.

### `segments` Block

The `segments` configuration block supports the following arguments:

* `edge_locations` - Regions where the edges are located.
* `name` - Name of a core network segment.
* `shared_segments` - Shared segments of a core network.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_core_network` using the core network ID. For example:

```terraform
import {
  to = aws_networkmanager_core_network.example
  id = "core-network-0d47f6t230mz46dy4"
}
```

Using `terraform import`, import `aws_networkmanager_core_network` using the core network ID. For example:

```console
% terraform import aws_networkmanager_core_network.example core-network-0d47f6t230mz46dy4
```
