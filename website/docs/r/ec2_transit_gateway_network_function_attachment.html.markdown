---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_network_function_attachment"
description: |-
  Manages an EC2 Transit Gateway Network Function Attachment to a Gateway Load Balancer endpoint service
---

# Resource: aws_ec2_transit_gateway_network_function_attachment

Manages an EC2 Transit Gateway Network Function Attachment. The attachment connects a transit gateway directly to a Gateway Load Balancer through its VPC endpoint service: the transit gateway creates and manages one Gateway Load Balancer endpoint in each selected Availability Zone, and enables appliance mode automatically.

For more information, see [Network function attachments](https://docs.aws.amazon.com/vpc/latest/tgw/tgw-nf-fw.html) in the Amazon VPC Transit Gateways guide.

~> **NOTE:** A network function attachment supports transit gateway route table associations and static routes, but not route propagation. Use the [`aws_ec2_transit_gateway_route_table_association`](ec2_transit_gateway_route_table_association.html) and [`aws_ec2_transit_gateway_route`](ec2_transit_gateway_route.html) resources with the attachment `id`.

~> **NOTE:** The transit gateway must have Encryption Support disabled.

~> **NOTE:** When the VPC endpoint service requires acceptance, this resource does not wait for the attachment to become available: creation returns while the attachment is `pendingAcceptance`, and adding an Availability Zone returns while the attachment is `modifying`. The owner of the endpoint service must accept the connection of each Availability Zone.

## Example Usage

### Basic Usage

```terraform
resource "aws_lb" "example" {
  name               = "example"
  load_balancer_type = "gateway"
  subnets            = aws_subnet.example[*].id
}

resource "aws_vpc_endpoint_service" "example" {
  acceptance_required        = false
  gateway_load_balancer_arns = [aws_lb.example.arn]
}

resource "aws_ec2_transit_gateway" "example" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
}

resource "aws_ec2_transit_gateway_network_function_attachment" "example" {
  transit_gateway_id    = aws_ec2_transit_gateway.example.id
  service_name          = aws_vpc_endpoint_service.example.service_name
  availability_zone_ids = aws_subnet.example[*].availability_zone_id

  tags = {
    Name = "example"
  }
}
```

### Routing Traffic Through the Attachment

```terraform
resource "aws_ec2_transit_gateway_route_table" "inspection" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
}

resource "aws_ec2_transit_gateway_route_table_association" "inspection" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_network_function_attachment.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.inspection.id
}

resource "aws_ec2_transit_gateway_route_table" "spokes" {
  transit_gateway_id = aws_ec2_transit_gateway.example.id
}

resource "aws_ec2_transit_gateway_route" "spokes_to_inspection" {
  destination_cidr_block         = "0.0.0.0/0"
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_network_function_attachment.example.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.spokes.id
}
```

## Argument Reference

The following arguments are required:

* `service_name` - (Required, Forces new resource) Service name of the Gateway Load Balancer VPC endpoint service, for example `com.amazonaws.vpce.us-east-1.vpce-svc-0123456789abcdef0`. Only one attachment can exist for a given transit gateway and endpoint service.
* `transit_gateway_id` - (Required, Forces new resource) EC2 Transit Gateway identifier.

The following arguments are optional:

* `availability_zone_ids` - (Optional) Set of Availability Zone IDs, for example `use1-az1`, in which the transit gateway creates a Gateway Load Balancer endpoint. Each Availability Zone must be supported by both the transit gateway and the endpoint service. Exactly one of `availability_zone_ids` or `availability_zones` must be configured.
* `availability_zones` - (Optional) Set of Availability Zone names, for example `us-east-1a`, in which the transit gateway creates a Gateway Load Balancer endpoint. Exactly one of `availability_zone_ids` or `availability_zones` must be configured.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value tags for the EC2 Transit Gateway Network Function Attachment. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - EC2 Transit Gateway Attachment ARN.
* `id` - EC2 Transit Gateway Attachment identifier.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_network_function_attachment.example
  identity = {
    id = "tgw-attach-0123456789abcdef0"
  }
}

resource "aws_ec2_transit_gateway_network_function_attachment" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the EC2 Transit Gateway Attachment.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Transit Gateway Network Function Attachments using the EC2 Transit Gateway Attachment identifier. For example:

```terraform
import {
  to = aws_ec2_transit_gateway_network_function_attachment.example
  id = "tgw-attach-0123456789abcdef0"
}
```

Using `terraform import`, import EC2 Transit Gateway Network Function Attachments using the EC2 Transit Gateway Attachment identifier. For example:

```console
% terraform import aws_ec2_transit_gateway_network_function_attachment.example tgw-attach-0123456789abcdef0
```
