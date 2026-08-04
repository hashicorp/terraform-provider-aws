---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_secondary_network"
description: |-
  Provides an EC2 Secondary Network resource.
---

# Resource: aws_ec2_secondary_network

Provides an EC2 Secondary Network resource for RDMA networking.

Secondary networks are specialized network resources that enable high-performance RDMA (Remote Direct Memory Access) networking for compute-intensive workloads.
They provide dedicated network infrastructure with low latency and high bandwidth capabilities.

## Example Usage

```terraform
resource "aws_ec2_secondary_network" "example" {
  ipv4_cidr_block = "10.0.0.0/16"
  network_type    = "rdma"

  tags = {
    Name = "example-secondary-network"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `ipv4_cidr_block` - (Required) IPv4 CIDR block for the secondary network. The CIDR block size must be between `/12` and `/28`.
* `network_type` - (Required) Type of secondary network. Currently only `rdma` is supported.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the secondary network.
* `id` - ID of the secondary network.
* `ipv4_cidr_block_associations` - A list of IPv4 CIDR block associations for the secondary network.
* `secondary_network_id` - ID of the secondary network.
* `state` - Current state of the secondary network.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

The following attributes are exported in the `ipv4_cidr_block_associations` block:

* `association_id` - Association ID for the IPv4 CIDR block.
* `cidr_block` - IPv4 CIDR block.
* `state` - State of the IPv4 CIDR block association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_secondary_network.example
  identity = {
    id = "sn-0123456789abcdef0"
  }
}

resource "aws_ec2_secondary_network" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` - (String) ID of the secondary network.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Secondary Networks using the `id`. For example:

```terraform
import {
  to = aws_ec2_secondary_network.example
  id = "sn-0123456789abcdef0"
}
```

Using `terraform import`, import EC2 Secondary Networks using the `id`. For example:

```console
% terraform import aws_ec2_secondary_network.example sn-0123456789abcdef0
```
