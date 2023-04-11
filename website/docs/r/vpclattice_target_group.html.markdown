---
subcategory: 'VPC Lattice'
layout: 'aws'
page_title: 'AWS: aws_vpclattice_target_group'
description: |-
  Terraform resource for managing an AWS VPC Lattice Service.
---

# Resource: aws_vpclattice_target_group

Terraform resource for managing an AWS VPC Lattice Target Group.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_target_group" "example" {
  name = "example"
  type = "INSTANCE|ALB|IP"
  config {
    port           = 443
    protocol       = "HTTPS"
    vpc_identifier = aws_vpc.example.id
  }
}
```

Basic usage with Health check

```terraform
resource "aws_vpclattice_target_group" "example" {
  name         = "example"
  type         = "INSTANCE/ALB/IP"
  client_token = "tstclienttoken"
  config {
    port             = 443
    protocol         = "HTTPS"
    vpc_identifier   = aws_vpc.example.id
    protocol_version = "GRPC|HPPT1|HTTP2"
    health_check {
      enabled             = true
      interval            = 20
      timeout             = 10
      healthy_threshold   = 2
      unhealthy_threshold = 2
      matcher             = "200-299"
      path                = "/instance"
      port                = 80
      protocol            = "HTTP"
      protocol_version    = "HTTP1|HTTP2"
    }
  }
}
```

If the type is Lambda config block is not required

```terraform
resource "aws_vpclattice_target_group" "example" {
  name = "example"
  type = "LAMBDA"
}
```

## Argument Reference

The following arguments are required:

- `name` - (Required) The name of the target group. The name must be unique within the account. The valid characters are a-z, 0-9, and hyphens (-). You can't use a hyphen as the first or last character, or immediately after another hyphen.

- `type` - (Required) The type of target group. Valid Values are IP | LAMBDA | INSTANCE | ALB

- `config` - The target group configuration. If type is set to LAMBDA, this parameter doesn't apply.

- `protocol` - (Required) The protocol to use for routing traffic to the targets. Default is the protocol of a target group.

- `vpc_identifier` - (Required) The ID of the VPC.

The following arguments are optional:

- `client_token` - (Optional) A unique, case-sensitive identifier that you provide to ensure the idempotency of the request. If you retry a request that completed successfully using the same client token and parameters, the retry succeeds without performing any actions. If the parameters aren't identical, the retry fails.

- `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

Config (`config`) support the following:

- `ip_address_type` - (Optional) The type of IP address used for the target group. The possible values are ipv4 and ipv6. This is an optional parameter. If not specified, the IP address type defaults to ipv4.
- `port` - (Optional) The port on which the targets are listening. For HTTP, the default is 80. For HTTPS, the default is 443.
- `protocol_version` - (Optional) The protocol version. Default value is HTTP1. Valid Values are HTTP1 | HTTP2 | GRPC
- `health_check` - (Optional) The health check configuration.

Health Check (`health_check`) supports the following:

- `enable` - (Optional) Indicates whether health checking is enabled.
- `interval` - (Optional) The approximate amount of time, in seconds, between health checks of an individual target. The range is 5–300 seconds. The default is 30 seconds.
- `timeout` - (Optional) The amount of time, in seconds, to wait before reporting a target as unhealthy. The range is 1–120 seconds. The default is 5 seconds.
- `healthy_threshold ` - (Optional) The number of consecutive successful health checks required before considering an unhealthy target healthy. The range is 2–10. The default is 5.
- `unhealthy_threshold` - (Optional) The number of consecutive failed health checks required before considering a target unhealthy. The range is 2–10. The default is 2.
- `matcher` - (Optional) The codes to use when checking for a successful response from a target. These are called Success codes in the console.
- `path` - (Optional) The destination for health checks on the targets. If the protocol version is HTTP/1.1 or HTTP/2, specify a valid URI (for example, /path?query). The default path is /. Health checks are not supported if the protocol version is gRPC, however, you can choose HTTP/1.1 or HTTP/2 and specify a valid URI.
- `port` - (Optional) The port used when performing health checks on targets. The default setting is the port that a target receives traffic on.
- `protocol` - (Optional) The protocol used when performing health checks on targets. The possible protocols are HTTP and HTTPS. The default is HTTP.
- `protocol_version` - (Optional) The protocol version used when performing health checks on targets. The possible protocol versions are HTTP1 and HTTP2.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `arn` - ARN of the target group.
- `id` - Unique identifier for the service.
- `status` - Status of the service.
- `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `5m`)
- `delete` - (Default `5m`)

## Import

VPC Lattice Target Group can be imported using the `id`, e.g.,

```
$ terraform import aws_vpclattice_service.example tg-0c11d4dc16ed96bdb
```
