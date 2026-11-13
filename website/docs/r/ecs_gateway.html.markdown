---
subcategory: "ECS (Elastic Container)"
layout: "aws"
page_title: "AWS: aws_ecs_gateway"
description: |-
  Manages an ECS Gateway.
---

# Resource: aws_ecs_gateway

Manages an ECS Gateway. A gateway manages shared Application Load Balancers that route traffic to ECS services in the clusters the gateway is associated with. AWS creates the load balancers only when a service in an associated cluster starts using the gateway.

~> **Note:** AWS enables deletion protection on a new gateway unless `deletion_protection` is `false`. To destroy a protected gateway, first apply `deletion_protection = false`.

~> **Note:** Every in-place change other than a change to `tags` sends the whole gateway configuration, including `infrastructure_role_arn`, to AWS. The caller needs `iam:PassRole` on the infrastructure role for each such change, also when the role itself does not change.

## Example Usage

### Basic Usage

```terraform
resource "aws_ecs_gateway" "example" {
  name                    = "example"
  infrastructure_role_arn = aws_iam_role.example.arn
  deletion_protection     = false

  configuration {
    host_shared_alb {
      network_configuration {
        subnets = aws_subnet.example[*].id
      }
    }
  }
}
```

### Internal Gateway with Access Logs and Test Traffic Rules

```terraform
resource "aws_ecs_gateway" "example" {
  name                    = "example"
  infrastructure_role_arn = aws_iam_role.example.arn
  deletion_protection     = false
  propagate_tags          = "NONE"

  configuration {
    host_shared_alb {
      http_desync   = "strictest"
      max_alb_count = 4
      scheme        = "internal"
      web_acl_arn   = aws_wafv2_web_acl.example.arn

      access_logs {
        destination {
          s3 {
            bucket = aws_s3_bucket.example.bucket
            prefix = "gateway"
          }
        }
      }

      network_configuration {
        inbound_security_group = aws_security_group.example.id
        ip_address_type        = "ipv4"
        subnets                = aws_subnet.example[*].id
      }

      test_traffic_rules {
        header {
          name   = "x-canary"
          values = ["true"]
        }
      }
    }
  }

  tags = {
    Team = "platform"
  }
}
```

## Argument Reference

The following arguments are required:

* `configuration` - (Required) Type-specific configuration of the gateway. See [`configuration` Block](#configuration-block) below.
* `infrastructure_role_arn` - (Required) ARN of the IAM role that the gateway assumes to manage infrastructure on your behalf. The role must trust `ecs.amazonaws.com`. The AWS managed policy `AmazonECSInfrastructureRoleForGatewayHostSharedALB` grants the permissions that a `host_shared_alb` gateway needs.
* `name` - (Required, Forces new resource) Name of the gateway. Must be unique within the account and up to 255 characters.

The following arguments are optional:

* `deletion_protection` - (Optional) Whether to enable deletion protection. AWS enables it when the argument is not set.
* `propagate_tags` - (Optional) Source of the tags that AWS adds to the infrastructure the gateway manages. Valid values are `GATEWAY` (the gateway's tags) and `NONE`. AWS uses `GATEWAY` when the argument is not set.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `access_logs` Block

The `access_logs` block supports:

* `destination` - (Required) Destination of the access logs. See [`destination` Block](#destination-block) below.

### `configuration` Block

The `configuration` block supports exactly one of:

* `host_shared_alb` - (Optional) Configuration of a gateway that routes traffic through shared Application Load Balancers with host-based routing. See [`host_shared_alb` Block](#host_shared_alb-block) below.

### `destination` Block

The `destination` block supports exactly one of:

* `s3` - (Optional) S3 destination. See [`s3` Block](#s3-block) below.

### `header` Block

The `header` block supports:

* `name` - (Required) Name of the header to match.
* `values` - (Required) Header values to match.

### `host_shared_alb` Block

The `host_shared_alb` block supports:

* `access_logs` - (Optional) Access log configuration of the load balancers. See [`access_logs` Block](#access_logs-block) below.
* `http_desync` - (Optional) Desync mitigation mode of the load balancers. Valid values are `defensive`, `monitor`, and `strictest`. AWS uses `defensive` when the argument is not set.
* `max_alb_count` - (Optional) Maximum number of Application Load Balancers that the gateway manages. AWS uses `10` when the argument is not set.
* `network_configuration` - (Required) Network configuration of the load balancers. See [`network_configuration` Block](#network_configuration-block) below.
* `scheme` - (Optional, Forces new resource) Scheme of the load balancers. Valid values are `internet-facing` and `internal`. AWS uses `internet-facing` when the argument is not set.
* `test_traffic_rules` - (Optional) Rules that route test traffic to the new service revision during progressive deployments. See [`test_traffic_rules` Block](#test_traffic_rules-block) below.
* `web_acl_arn` - (Optional) ARN of the WAF web ACL to associate with the load balancers.

### `network_configuration` Block

The `network_configuration` block supports:

* `inbound_security_group` - (Optional) ID of a security group that controls inbound traffic to the load balancers. When it is set, the gateway does not create inbound rules on the security group it manages.
* `ip_address_type` - (Optional, Forces new resource) IP address type of the load balancers. Valid values are `dualstack`, `dualstack-without-public-ipv4`, and `ipv4`. AWS uses `ipv4` when the argument is not set.
* `subnets` - (Required, Forces new resource) IDs of the subnets for the load balancers.

### `s3` Block

The `s3` block supports:

* `bucket` - (Required) Name of the S3 bucket that receives the access logs.
* `prefix` - (Optional) Object key prefix of the access logs.

### `test_traffic_rules` Block

The `test_traffic_rules` block supports:

* `header` - (Required) Header-based rule that identifies test traffic. See [`header` Block](#header-block) below.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `access_type` - Whether the gateway is publicly accessible (`PUBLIC`) or private (`PRIVATE`). AWS derives it from `scheme`.
* `arn` - ARN of the gateway.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `20m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ecs_gateway.example
  identity = {
    arn = "arn:aws:ecs:us-west-2:123456789012:gateway/example/a1b2c3d4-5678-90ab-cdef-11111example"
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
  id = "arn:aws:ecs:us-west-2:123456789012:gateway/example/a1b2c3d4-5678-90ab-cdef-11111example"
}
```

Using `terraform import`, import ECS Gateways using the ARN. For example:

```console
% terraform import aws_ecs_gateway.example arn:aws:ecs:us-west-2:123456789012:gateway/example/a1b2c3d4-5678-90ab-cdef-11111example
```
