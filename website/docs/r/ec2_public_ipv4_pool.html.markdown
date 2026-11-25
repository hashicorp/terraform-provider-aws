---
subcategory: "EC2 (Elastic Compute Cloud)"
layout: "aws"
page_title: "AWS: aws_ec2_public_ipv4_pool"
description: |-
  Manages an AWS EC2 Public IPv4 Pool.
---

# Resource: aws_ec2_public_ipv4_pool

Manages an AWS EC2 Public IPv4 Pool.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_public_ipv4_pool" "example" {}
```

### With Optional Arguments

```terraform
resource "aws_ec2_public_ipv4_pool" "example" {
  network_border_group = data.aws_region.current.region

  tags = {
    Name = "example"
  }
}

data "aws_region" "current" {}
```

## Argument Reference

This resource supports the following arguments:

* `network_border_group` - (Optional, Forces new resource) Availability Zone (AZ) or Local Zone (LZ) network border group that the resource that the IP address is assigned to is in. Defaults to an AZ network border group. For more information on available Local Zones, see [Local Zone availability](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-byoip.html#byoip-zone-avail) in the Amazon EC2 User Guide.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN, `arn:<partition>:ec2:<region>:<account>:ipv4pool-ec2/<pool_id>`.
* `description` - Description of the address pool.
* `pool_address_ranges` - Address ranges. See [`pool_address_ranges` Block](#pool_address_ranges-block) below.
* `pool_id` - ID of the address pool.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `total_address_count` - Total number of addresses.
* `total_available_address_count` - Total number of available addresses.

### `pool_address_ranges` Block

The `pool_address_ranges` block exports:

* `address_count` - Number of addresses in the range.
* `available_address_count` - Number of available addresses in the range.
* `first_address` - First IP address in the range.
* `last_address` - Last IP address in the range.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ec2_public_ipv4_pool.example
  identity = {
    pool_id = "ipv4pool-ec2-fbac04f087e582365"
  }
}

resource "aws_ec2_public_ipv4_pool" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `pool_id` (String) ID of the address pool.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 Public IPv4 Pool using the `pool_id`. For example:

```terraform
import {
  to = aws_ec2_public_ipv4_pool.example
  id = "ipv4pool-ec2-fbac04f087e582365"
}
```

Using `terraform import`, import EC2 Public IPv4 Pool using the `pool_id`. For example:

```console
% terraform import aws_ec2_public_ipv4_pool.example ipv4pool-ec2-fbac04f087e582365
```
