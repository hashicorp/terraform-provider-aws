---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_listener"
description: |-
  Terraform resource for managing an AWS VPC Lattice Listener.
---

# Resource: aws_vpclattice_listener

Terraform resource for managing an AWS VPC Lattice Listener.

## Example Usage

### Fixed response action

```terraform
resource "aws_vpclattice_service" "example" {
  name = "example"
}

resource "aws_vpclattice_listener" "example" {
  name               = "example"
  protocol           = "HTTPS"
  service_identifier = aws_vpclattice_service.example.id
  default_action {
    fixed_response {
      status_code = 404
    }
  }
}
```

### Forward action

```terraform
resource "aws_vpclattice_service" "example" {
  name = "example"
}

resource "aws_vpclattice_target_group" "example" {
  name = "example-target-group-1"
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.example.id
  }
}

resource "aws_vpclattice_listener" "example" {
  name               = "example"
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.example.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.example.id
      }
    }
  }
}
```

### Forward action with weighted target groups

```terraform
resource "aws_vpclattice_service" "example" {
  name = "example"
}

resource "aws_vpclattice_target_group" "example1" {
  name = "example-target-group-1"
  type = "INSTANCE"

  config {
    port           = 80
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.example.id
  }
}

resource "aws_vpclattice_target_group" "example2" {
  name = "example-target-group-2"
  type = "INSTANCE"

  config {
    port           = 8080
    protocol       = "HTTP"
    vpc_identifier = aws_vpc.example.id
  }
}

resource "aws_vpclattice_listener" "example" {
  name               = "example"
  protocol           = "HTTP"
  service_identifier = aws_vpclattice_service.example.id
  default_action {
    forward {
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.example1.id
        weight                  = 80
      }
      target_groups {
        target_group_identifier = aws_vpclattice_target_group.example2.id
        weight                  = 20
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `default_action` - (Required) Default action block for the default listener rule. Default action blocks are defined below.
* `name` - (Required, Forces new resource) Name of the listener. A listener name must be unique within a service. Valid characters are a-z, 0-9, and hyphens (-). You can't use a hyphen as the first or last character, or immediately after another hyphen.
* `port` - (Optional, Forces new resource) Listener port. You can specify a value from 1 to 65535. If `port` is not specified and `protocol` is HTTP, the value will default to 80. If `port` is not specified and `protocol` is HTTPS, the value will default to 443.
* `protocol` - (Required, Forces new resource) Protocol for the listener. Supported values are `HTTP`, `HTTPS` or `TLS_PASSTHROUGH`
* `service_arn` - (Optional) Amazon Resource Name (ARN) of the VPC Lattice service. You must include either the `service_arn` or `service_identifier` arguments.
* `service_identifier` - (Optional) ID of the VPC Lattice service. You must include either the `service_arn` or `service_identifier` arguments.
-> **NOTE:** You must specify one of the following arguments: `service_arn` or `service_identifier`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Default Action

Default action blocks (for `default_action`) must include at least one of the following argument blocks:

* `fixed-response` - (Optional) Configuration block for returning a fixed response. See Fixed Response blocks below.
* `forward` - (Optional) Route requests to one or more target groups. See Forward blocks below.

-> **NOTE:** You must specify exactly one of the following argument blocks: `fixed_response` or `forward`.

### Fixed Response

Fixed response blocks (for `fixed-response`) must include the following argument:

* `status_code` - (Required) Custom HTTP status code to return, e.g. a 404 response code. See [Listeners](https://docs.aws.amazon.com/vpc-lattice/latest/ug/listeners.html) in the AWS documentation for a list of supported codes.

### Forward

Forward blocks (for `forward`) must include the following arguments:

* `target_groups` - (Required) One or more target group blocks.

### Target Groups

Target group blocks (for `target_group`) must include the following arguments:

* `target_group_identifier` - (Required) ID or Amazon Resource Name (ARN) of the target group.
* `weight` - (Optional) Determines how requests are distributed to the target group. Only required if you specify multiple target groups for a forward action. For example, if you specify two target groups, one with a
weight of 10 and the other with a weight of 20, the target group with a weight of 20 receives twice as many requests as the other target group. See [Listener rules](https://docs.aws.amazon.com/vpc-lattice/latest/ug/listeners.html#listener-rules) in the AWS documentation for additional examples. Default: `100`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the listener.
* `created_at` - Date and time that the listener was created, specified in ISO-8601 format.
* `listener_id` - Standalone ID of the listener, e.g. `listener-0a1b2c3d4e5f6g`.
* `updated_at` - Date and time that the listener was last updated, specified in ISO-8601 format.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Listener using the `listener_id` of the listener and the `id` of the VPC Lattice service combined with a `/` character. For example:

```terraform
import {
  to = aws_vpclattice_listener.example
  id = "svc-1a2b3c4d/listener-987654321"
}
```

Using `terraform import`, import VPC Lattice Listener using the `listener_id` of the listener and the `id` of the VPC Lattice service combined with a `/` character. For example:

```console
% terraform import aws_vpclattice_listener.example svc-1a2b3c4d/listener-987654321
```
