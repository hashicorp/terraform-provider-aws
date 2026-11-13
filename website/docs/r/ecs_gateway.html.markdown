---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_gateway"
description: |-
  Manages an Amazon ECS Gateway.
---

# Resource: aws_ecs_gateway

Manages an Amazon ECS Gateway. A gateway is a shared ingress path backed by an Application Load Balancer that Amazon ECS manages on your behalf. Services connect to a gateway with a `gateway_connector` block on [`aws_ecs_service`](ecs_service.html), after the gateway has been associated with the service's cluster using [`aws_ecs_gateway_cluster_association`](ecs_gateway_cluster_association.html).

~> **NOTE:** Gateways are created with deletion protection enabled by default. Set `deletion_protection = false` and apply before destroying the resource.

## Example Usage

### Basic Usage

```terraform
data "aws_partition" "current" {}

resource "aws_iam_role" "gateway" {
  name = "example-ecs-gateway"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "gateway" {
  role       = aws_iam_role.gateway.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRoleForGatewayHostSharedALB"
}

resource "aws_ecs_gateway" "example" {
  name                    = "example"
  infrastructure_role_arn = aws_iam_role.gateway.arn

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = [aws_subnet.a.id, aws_subnet.b.id]
      }
    }
  }

  depends_on = [aws_iam_role_policy_attachment.gateway]
}
```

### Internal Gateway with Access Logs

```terraform
resource "aws_ecs_gateway" "example" {
  name                    = "example-internal"
  infrastructure_role_arn = aws_iam_role.gateway.arn
  propagate_tags          = "GATEWAY"

  configuration {
    host_shared_alb {
      scheme        = "internal"
      http_desync   = "strictest"
      max_alb_count = 2

      network_configuration {
        inbound_security_group = aws_security_group.gateway.id
        subnets                = [aws_subnet.a.id, aws_subnet.b.id]
      }

      access_logs {
        destination {
          s3 {
            bucket = aws_s3_bucket.logs.bucket
            prefix = "ecs-gateway"
          }
        }
      }
    }
  }

  tags = {
    Environment = "production"
  }
}
```

## Argument Reference

The following arguments are required:

* `configuration` - (Required) Gateway configuration. See [`configuration`](#configuration-block) below.
* `infrastructure_role_arn` - (Required) ARN of the IAM role that Amazon ECS assumes to manage the gateway's load balancing resources. The role must trust `ecs.amazonaws.com` and have the `AmazonECSInfrastructureRoleForGatewayHostSharedALB` managed policy (or equivalent permissions) attached.
* `name` - (Required, Forces new resource) Name of the gateway. Up to 255 characters.

The following arguments are optional:

* `deletion_protection` - (Optional) Whether deletion protection is enabled. Defaults to `true`.
* `propagate_tags` - (Optional) Whether to propagate the gateway's tags to the resources Amazon ECS creates for it. Valid values: `GATEWAY`, `NONE`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `configuration` Block

* `host_shared_alb` - (Required) Configuration for a gateway backed by a shared Application Load Balancer. See [`host_shared_alb`](#host_shared_alb-block) below.

### `host_shared_alb` Block

* `access_logs` - (Optional) Access log configuration. See [`access_logs`](#access_logs-block) below.
* `http_desync` - (Optional) HTTP desync mitigation mode. Valid values: `monitor`, `defensive`, `strictest`.
* `max_alb_count` - (Optional) Maximum number of Application Load Balancers the gateway can use.
* `network_configuration` - (Required) Network configuration. See [`network_configuration`](#network_configuration-block) below.
* `port` - (Optional, Forces new resource if configured) Listener port.
* `protocol` - (Optional, Forces new resource if configured) Listener protocol. Valid values: `HTTP`.
* `scheme` - (Optional, Forces new resource if configured) Load balancer scheme. Valid values: `internet-facing`, `internal`.
* `test_traffic_rules` - (Optional) Rules that route test traffic during blue/green deployments. See [`test_traffic_rules`](#test_traffic_rules-block) below.
* `web_acl_arn` - (Optional) ARN of an AWS WAF web ACL to associate with the gateway.

### `access_logs` Block

* `destination` - (Required) Destination for the access logs.
    * `s3` - (Required) Amazon S3 destination.
        * `bucket` - (Required) Name of the S3 bucket.
        * `prefix` - (Optional) Key prefix for the log objects.

### `network_configuration` Block

* `inbound_security_group` - (Optional) ID of the security group that controls inbound traffic to the gateway.
* `ip_address_type` - (Optional, Forces new resource if configured) IP address type. Valid values: `ipv4`, `dualstack`.
* `subnets` - (Required, Forces new resource) IDs of the subnets for the gateway. At least two subnets in different Availability Zones are required.

### `test_traffic_rules` Block

* `header` - (Required) Header-based rule.
    * `name` - (Required) Name of the HTTP header.
    * `values` - (Required) Header values that identify test traffic.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_type` - Whether the gateway is reachable from the internet (`PUBLIC`) or only from within the VPC (`PRIVATE`).
* `arn` - ARN of the gateway.
* `configuration[0].host_shared_alb[0].current_alb_count` - Number of Application Load Balancers the gateway currently uses.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `20m`)
* `update` - (Default `20m`)
* `delete` - (Default `20m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecs_gateway.example
  identity = {
    arn = "arn:aws:ecs:us-east-1:123456789012:gateway/example/0123456789abcdef"
  }
}

resource "aws_ecs_gateway" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `arn` (String) ARN of the ECS Gateway.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import ECS Gateways using the ARN. For example:

```terraform
import {
  to = aws_ecs_gateway.example
  id = "arn:aws:ecs:us-east-1:123456789012:gateway/example/0123456789abcdef"
}
```

Using `terraform import`, import ECS Gateways using the ARN. For example:

```console
% terraform import aws_ecs_gateway.example arn:aws:ecs:us-east-1:123456789012:gateway/example/0123456789abcdef
```
