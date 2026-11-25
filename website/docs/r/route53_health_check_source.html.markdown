---
subcategory: "Route 53"
layout: "aws"
page_title: "AWS: aws_route53_health_check_source"
description: |-
  Manages a Route 53 health check source, the VPC location from which Route 53 sends private endpoint health checks.
---

# Resource: aws_route53_health_check_source

Manages a Route 53 health check source, the VPC location from which Route 53 sends private endpoint health checks. Route 53 creates network interfaces in each subnet of the source and sends health check requests from all of them. For more information, see [Health check sources](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/health-check-sources.html) in the Amazon Route 53 Developer Guide.

The health check source is in the Region of its VPC. Health checks that use the source, and their metrics and alarms, are in that Region too.

~> **Note:** Route 53 can't delete a health check source while a health check uses it. Move those health checks to another source or delete them first.

## Example Usage

### Basic Usage

```terraform
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  count = 3

  vpc_id               = aws_vpc.example.id
  availability_zone_id = data.aws_availability_zones.available.zone_ids[count.index]
  cidr_block           = cidrsubnet(aws_vpc.example.cidr_block, 8, count.index)
}

resource "aws_security_group" "example" {
  name   = "health-check-source"
  vpc_id = aws_vpc.example.id

  egress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [aws_vpc.example.cidr_block]
  }
}

resource "aws_route53_health_check_source" "example" {
  vpc_id = aws_vpc.example.id

  dynamic "subnet_configuration" {
    for_each = aws_subnet.example

    content {
      availability_zone_id = subnet_configuration.value.availability_zone_id
      security_group_ids   = [aws_security_group.example.id]
      subnet_id            = subnet_configuration.value.id
    }
  }

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `subnet_configuration` - (Required) Subnets from which Route 53 sends health check requests, each in a different Availability Zone. Route 53 sets the required number of subnets for each Region, which is two or three. For the number in your Region, see [Health check sources](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/health-check-sources.html) in the Amazon Route 53 Developer Guide. See [`subnet_configuration` Block](#subnet_configuration-block) below. Changing this forces a new resource.
* `tags` - (Optional) Map of tags to assign to the health check source. The `Name` tag is the name that the Route 53 console shows. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vpc_id` - (Required) ID of the VPC that contains the subnets. Changing this forces a new resource.

### `subnet_configuration` Block

* `availability_zone_id` - (Required) ID of the Availability Zone that contains the subnet, for example `use1-az1`. Use the Availability Zone ID, not the name. It must match the subnet's Availability Zone, or creation fails with `INVALID_SUBNET`.
* `security_group_ids` - (Required) IDs of the security groups that Route 53 associates with the network interfaces it creates in this subnet. Each subnet can have different security groups, but health checks are sent from every subnet, so the security groups of every subnet must allow outbound traffic to the endpoints that the health checks monitor. Using the same security groups for every subnet is simplest.
* `subnet_id` - (Required) ID of the subnet.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the health check source.
* `id` - ID of the health check source.
* `status` - Status of the health check source, for example `CREATE_COMPLETED`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `15m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_route53_health_check_source.example
  identity = {
    id = "health-check-source-id"
  }
}

resource "aws_route53_health_check_source" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `id` (String) ID of the health check source.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Route 53 health check sources using the `id`. For example:

```terraform
import {
  to = aws_route53_health_check_source.example
  id = "health-check-source-id"
}
```

Using `terraform import`, import Route 53 health check sources using the `id`. For example:

```console
% terraform import aws_route53_health_check_source.example health-check-source-id
```
