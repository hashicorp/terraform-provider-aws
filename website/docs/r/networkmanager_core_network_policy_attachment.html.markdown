---
subcategory: "Network Manager"
layout: "aws"
page_title: "AWS: aws_networkmanager_core_network_policy_attachment"
description: |-
  Manages a Network Manager Core Network Policy Attachment.
---

# Resource: aws_networkmanager_core_network_policy_attachment

Manages a Network Manager Core Network Policy Attachment.

Use this resource to attach a Core Network Policy to an existing Core Network and execute the change set, which deploys changes globally based on the policy submitted (sets the policy to `LIVE`).

~> **NOTE:** Deleting this resource will not delete the current policy defined in this resource. Deleting this resource will also not revert the current `LIVE` policy to the previous version.

## Example Usage

### Basic

```terraform
resource "aws_networkmanager_core_network" "example" {
  global_network_id = aws_networkmanager_global_network.example.id
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}
```

### With VPC Attachment (Single Region)

The example below illustrates the scenario where your policy document has static routes pointing to VPC attachments and you want to attach your VPCs to the core network before applying the desired policy document. Set the `create_base_policy` argument of the [`aws_networkmanager_core_network` resource](/docs/providers/aws/r/networkmanager_core_network.html) to `true` if your core network does not currently have any `LIVE` policies (e.g. this is the first `terraform apply` with the core network resource), since a `LIVE` policy is required before VPCs can be attached to the core network. Otherwise, if your core network already has a `LIVE` policy, you may exclude the `create_base_policy` argument. There are 2 options to implement this:

- Option 1: Use the `base_policy_document` argument in the [`aws_networkmanager_core_network` resource](/docs/providers/aws/r/networkmanager_core_network.html) that allows the most customizations to a base policy. Use this to customize the `edge_locations` `asn`. In the example below, `us-west-2` and ASN `65500` are used in the base policy.
- Option 2: Use the `create_base_policy` argument only. This creates a base policy in the region specified in the `provider` block.

#### Option 1 - using base_policy_document

```terraform
resource "aws_networkmanager_global_network" "example" {}

data "aws_networkmanager_core_network_policy_document" "base" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
      asn      = "65500"
    }
  }

  segments {
    name = "segment"
  }
}

resource "aws_networkmanager_core_network" "example" {
  global_network_id    = aws_networkmanager_global_network.example.id
  base_policy_document = data.aws_networkmanager_core_network_policy_document.base.json
  create_base_policy   = true
}

data "aws_networkmanager_core_network_policy_document" "example" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
      asn      = "65500"
    }
  }

  segments {
    name = "segment"
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "0.0.0.0/0"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.example.id,
    ]
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}

resource "aws_networkmanager_vpc_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = aws_subnet.example[*].arn
  vpc_arn         = aws_vpc.example.arn
}
```

#### Option 2 - create_base_policy only

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

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "0.0.0.0/0"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.example.id,
    ]
  }
}

resource "aws_networkmanager_core_network" "example" {
  global_network_id  = aws_networkmanager_global_network.example.id
  create_base_policy = true
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}

resource "aws_networkmanager_vpc_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = aws_subnet.example[*].arn
  vpc_arn         = aws_vpc.example.arn
}
```

### With VPC Attachment (Multi-Region)

The example below illustrates the scenario where your policy document has static routes pointing to VPC attachments and you want to attach your VPCs to the core network before applying the desired policy document. Set the `create_base_policy` argument of the [`aws_networkmanager_core_network` resource](/docs/providers/aws/r/networkmanager_core_network.html) to `true` if your core network does not currently have any `LIVE` policies (e.g. this is the first `terraform apply` with the core network resource), since a `LIVE` policy is required before VPCs can be attached to the core network. Otherwise, if your core network already has a `LIVE` policy, you may exclude the `create_base_policy` argument. For multi-region in a core network that does not yet have a `LIVE` policy, there are 2 options:

- Option 1: Use the `base_policy_document` argument that allows the most customizations to a base policy. Use this to customize the `edge_locations` `asn`. In the example below, `us-west-2`, `us-east-1` and specific ASNs are used in the base policy.
- Option 2: Pass a list of regions to the [`aws_networkmanager_core_network` resource](/docs/providers/aws/r/networkmanager_core_network.html) `base_policy_regions` argument. In the example below, `us-west-2` and `us-east-1` are specified in the base policy.

#### Option 1 - using base_policy_document

```terraform
resource "aws_networkmanager_global_network" "example" {}

data "aws_networkmanager_core_network_policy_document" "base" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
      asn      = "65500"
    }

    edge_locations {
      location = "us-east-1"
      asn      = "65501"
    }
  }

  segments {
    name = "segment"
  }
}

resource "aws_networkmanager_core_network" "example" {
  global_network_id    = aws_networkmanager_global_network.example.id
  base_policy_document = data.aws_networkmanager_core_network_policy_document.base.json
  create_base_policy   = true
}

data "aws_networkmanager_core_network_policy_document" "example" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
      asn      = "65500"
    }

    edge_locations {
      location = "us-east-1"
      asn      = "65501"
    }
  }

  segments {
    name = "segment"
  }

  segments {
    name = "segment2"
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "10.0.0.0/16"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.example_us_west_2.id,
    ]
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "10.1.0.0/16"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.example_us_east_1.id,
    ]
  }
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}

resource "aws_networkmanager_vpc_attachment" "example_us_west_2" {
  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = aws_subnet.example_us_west_2[*].arn
  vpc_arn         = aws_vpc.example_us_west_2.arn
}

resource "aws_networkmanager_vpc_attachment" "example_us_east_1" {
  provider = "alternate"

  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = aws_subnet.example_us_east_1[*].arn
  vpc_arn         = aws_vpc.example_us_east_1.arn
}
```

#### Option 2 - using base_policy_regions

```terraform
resource "aws_networkmanager_global_network" "example" {}

data "aws_networkmanager_core_network_policy_document" "example" {
  core_network_configuration {
    asn_ranges = ["65022-65534"]

    edge_locations {
      location = "us-west-2"
    }

    edge_locations {
      location = "us-east-1"
    }
  }

  segments {
    name = "segment"
  }

  segments {
    name = "segment2"
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "10.0.0.0/16"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.example_us_west_2.id,
    ]
  }

  segment_actions {
    action  = "create-route"
    segment = "segment"
    destination_cidr_blocks = [
      "10.1.0.0/16"
    ]
    destinations = [
      aws_networkmanager_vpc_attachment.example_us_east_1.id,
    ]
  }
}

resource "aws_networkmanager_core_network" "example" {
  global_network_id   = aws_networkmanager_global_network.example.id
  base_policy_regions = ["us-west-2", "us-east-1"]
  create_base_policy  = true
}

resource "aws_networkmanager_core_network_policy_attachment" "example" {
  core_network_id = aws_networkmanager_core_network.example.id
  policy_document = data.aws_networkmanager_core_network_policy_document.example.json
}

resource "aws_networkmanager_vpc_attachment" "example_us_west_2" {
  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = aws_subnet.example_us_west_2[*].arn
  vpc_arn         = aws_vpc.example_us_west_2.arn
}

resource "aws_networkmanager_vpc_attachment" "example_us_east_1" {
  provider = "alternate"

  core_network_id = aws_networkmanager_core_network.example.id
  subnet_arns     = aws_subnet.example_us_east_1[*].arn
  vpc_arn         = aws_vpc.example_us_east_1.arn
}
```

## Argument Reference

The following arguments are required:

* `core_network_id` - (Required) ID of the core network that a policy will be attached to and made `LIVE`.
* `policy_document` - (Required) Policy document for creating a core network. Note that updating this argument will result in the new policy document version being set as the `LATEST` and `LIVE` policy document. Refer to the [Core network policies documentation](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-policy-change-sets.html) for more information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `state` - Current state of a core network.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `update` - (Default `30m`). If this is the first time attaching a policy to a core network then this timeout value is also used as the `create` timeout value.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_networkmanager_core_network_policy_attachment` using the core network ID. For example:

```terraform
import {
  to = aws_networkmanager_core_network_policy_attachment.example
  id = "core-network-0d47f6t230mz46dy4"
}
```

Using `terraform import`, import `aws_networkmanager_core_network_policy_attachment` using the core network ID. For example:

```console
% terraform import aws_networkmanager_core_network_policy_attachment.example core-network-0d47f6t230mz46dy4
```
