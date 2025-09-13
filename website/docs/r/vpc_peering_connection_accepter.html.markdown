---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_peering_connection_accepter"
description: |-
  Manage the accepter's side of a VPC Peering Connection.
---

# Resource: aws_vpc_peering_connection_accepter

Provides a resource to manage the accepter's side of a VPC Peering Connection.

When a cross-account (requester's AWS account differs from the accepter's AWS account) or an inter-region
VPC Peering Connection is created, a VPC Peering Connection resource is automatically created in the
accepter's account.
The requester can use the `aws_vpc_peering_connection` resource to manage its side of the connection
and the accepter can use the `aws_vpc_peering_connection_accepter` resource to "adopt" its side of the
connection into management.

## Example Usage

### Cross-Account Peering Or Cross-Region Peering Terraform AWS Provider v5 (and below)

```terraform
provider "aws" {
  region = "us-east-1"

  # Requester's credentials.
}

provider "aws" {
  alias  = "peer"
  region = "us-west-2"

  # Accepter's credentials.
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc" "peer" {
  provider   = aws.peer
  cidr_block = "10.1.0.0/16"
}

data "aws_caller_identity" "peer" {
  provider = aws.peer
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "peer" {
  vpc_id        = aws_vpc.main.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  peer_region   = "us-west-2"
  auto_accept   = false

  tags = {
    Side = "Requester"
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider                  = aws.peer
  vpc_peering_connection_id = aws_vpc_peering_connection.peer.id
  auto_accept               = true

  tags = {
    Side = "Accepter"
  }
}
```

### Cross-Region Peering (Same Account) Terraform AWS Provider v6 (and above)

```terraform
provider "aws" {
  region = "us-east-1"

  # Requester's credentials.
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc" "peer" {
  region     = "us-west-2"
  cidr_block = "10.1.0.0/16"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "peer" {
  vpc_id      = aws_vpc.main.id
  peer_vpc_id = aws_vpc.peer.id
  peer_region = "us-west-2"
  auto_accept = false

  tags = {
    Side = "Requester"
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  region                    = "us-west-2"
  vpc_peering_connection_id = aws_vpc_peering_connection.peer.id
  auto_accept               = true

  tags = {
    Side = "Accepter"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_peering_connection_id` - (Required) The VPC Peering Connection ID to manage.
* `auto_accept` - (Optional) Whether or not to accept the peering request. Defaults to `false`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Removing `aws_vpc_peering_connection_accepter` from your configuration

AWS allows a cross-account VPC Peering Connection to be deleted from either the requester's or accepter's side.
However, Terraform only allows the VPC Peering Connection to be deleted from the requester's side
by removing the corresponding `aws_vpc_peering_connection` resource from your configuration.
Removing a `aws_vpc_peering_connection_accepter` resource from your configuration will remove it
from your statefile and management, **but will not destroy the VPC Peering Connection.**

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the VPC Peering Connection.
* `accept_status` - The status of the VPC Peering Connection request.
* `vpc_id` - The ID of the accepter VPC.
* `peer_vpc_id` - The ID of the requester VPC.
* `peer_owner_id` - The AWS account ID of the owner of the requester VPC.
* `peer_region` - The region of the accepter VPC.
* `accepter` - A configuration block that describes [VPC Peering Connection]
(https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html) options set for the accepter VPC.
* `requester` - A configuration block that describes [VPC Peering Connection]
(https://docs.aws.amazon.com/vpc/latest/peering/what-is-vpc-peering.html) options set for the requester VPC.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

#### Accepter and Requester Attribute Reference

* `allow_remote_vpc_dns_resolution` - Indicates whether a local VPC can resolve public DNS hostnames to
private IP addresses when queried from instances in a peer VPC.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Peering Connection Accepters using the Peering Connection ID. For example:

```terraform
import {
  to = aws_vpc_peering_connection_accepter.example
  id = "pcx-12345678"
}
```

Using `terraform import`, import VPC Peering Connection Accepters using the Peering Connection ID. For example:

```console
% terraform import aws_vpc_peering_connection_accepter.example pcx-12345678
```

Certain resource arguments, like `auto_accept`, do not have an EC2 API method for reading the information after peering connection creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference. For example:

```terraform
resource "aws_vpc_peering_connection_accepter" "example" {
  # ... other configuration ...

  # There is no AWS EC2 API for reading auto_accept
  lifecycle {
    ignore_changes = [auto_accept]
  }
}
```
