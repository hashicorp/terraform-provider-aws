---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_target_group"
description: |-
  Terraform resource for managing an AWS VPC Lattice Target Group.
---

# Resource: aws_vpclattice_target_group

Terraform resource for managing an AWS VPC Lattice Target Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_target_group" "example" {
  name = "example"
  type = "INSTANCE"

  config {
    vpc_identifier = aws_vpc.example.id

    port     = 443
    protocol = "HTTPS"
  }
}
```

### Basic usage with Health check

```terraform
resource "aws_vpclattice_target_group" "example" {
  name = "example"
  type = "IP"

  config {
    vpc_identifier = aws_vpc.example.id

    ip_address_type  = "IPV4"
    port             = 443
    protocol         = "HTTPS"
    protocol_version = "HTTP1"

    health_check {
      enabled                       = true
      health_check_interval_seconds = 20
      health_check_timeout_seconds  = 10
      healthy_threshold_count       = 7
      unhealthy_threshold_count     = 3

      matcher {
        value = "200-299"
      }

      path             = "/instance"
      port             = 80
      protocol         = "HTTP"
      protocol_version = "HTTP1"
    }
  }
}
```

### ALB

If the type is ALB, `health_check` block is not supported.

```terraform
resource "aws_vpclattice_target_group" "example" {
  name = "example"
  type = "ALB"

  config {
    vpc_identifier = aws_vpc.example.id

    port             = 443
    protocol         = "HTTPS"
    protocol_version = "HTTP1"
  }
}
```

### Lambda

If the type is Lambda, `config` block is not supported.

```terraform
resource "aws_vpclattice_target_group" "example" {
  name = "example"
  type = "LAMBDA"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the target group. The name must be unique within the account. The valid characters are a-z, 0-9, and hyphens (-). You can't use a hyphen as the first or last character, or immediately after another hyphen.
* `type` - (Required) Type of target group. Valid values are `IP`, `LAMBDA`, `INSTANCE`, or `ALB`.

The following arguments are optional:

* `config` - (Optional) Target group configuration. See [`config` Block](#config-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `config` Block

The `config` configuration block supports the following arguments:

* `health_check` - (Optional) Health check configuration. See [`health_check` Block](#health_check-block) below.
* `ip_address_type` - (Optional) Type of IP address used for the target group. Valid values: `IPV4` or `IPV6`.
* `lambda_event_structure_version` - (Optional) Version of the event structure that the Lambda function receives. Supported only if `type` is `LAMBDA`. Valid values are `V1` or `V2`.
* `port` - (Optional) Port on which the targets are listening.
* `protocol` - (Optional) Protocol to use for routing traffic to the targets. Valid values are `HTTP` or `HTTPS`.
* `protocol_version` - (Optional) Protocol version. Valid values are `HTTP1`, `HTTP2`, or `GRPC`. Default value is `HTTP1`.
* `vpc_identifier` - (Optional) ID of the VPC.

### `health_check` Block

The `health_check` configuration block supports the following arguments:

* `enabled` - (Optional) Whether health checking is enabled. Defaults to `true`.
* `health_check_interval_seconds` - (Optional) Approximate amount of time, in seconds, between health checks of an individual target. The range is 5–300 seconds. The default is 30 seconds.
* `health_check_timeout_seconds` - (Optional) Amount of time, in seconds, to wait before reporting a target as unhealthy. The range is 1–120 seconds. The default is 5 seconds.
* `healthy_threshold_count` - (Optional) Number of consecutive successful health checks required before considering an unhealthy target healthy. The range is 2–10. The default is 5.
* `matcher` - (Optional) Codes to use when checking for a successful response from a target. See [`matcher` Block](#matcher-block) below.
* `path` - (Optional) Destination for health checks on the targets. If the protocol version is HTTP/1.1 or HTTP/2, specify a valid URI (for example, /path?query). The default path is `/`. Health checks are not supported if the protocol version is gRPC, however, you can choose HTTP/1.1 or HTTP/2 and specify a valid URI.
* `port` - (Optional) Port used when performing health checks on targets. The default setting is the port that a target receives traffic on.
* `protocol` - (Optional) Protocol used when performing health checks on targets. The possible protocols are `HTTP` and `HTTPS`.
* `protocol_version` - (Optional) Protocol version used when performing health checks on targets. The possible protocol versions are `HTTP1` and `HTTP2`. The default is `HTTP1`.
* `unhealthy_threshold_count` - (Optional) Number of consecutive failed health checks required before considering a target unhealthy. The range is 2–10. The default is 2.

### `matcher` Block

The `matcher` configuration block supports the following argument:

* `value` - (Optional) HTTP codes to use when checking for a successful response from a target.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the target group.
* `id` - Unique identifier for the target group.
* `status` - Status of the target group.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Target Group using the `id`. For example:

```terraform
import {
  to = aws_vpclattice_target_group.example
  id = "tg-0c11d4dc16ed96bdb"
}
```

Using `terraform import`, import VPC Lattice Target Group using the `id`. For example:

```console
% terraform import aws_vpclattice_target_group.example tg-0c11d4dc16ed96bdb
```
