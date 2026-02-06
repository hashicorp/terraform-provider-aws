---
subcategory: "GameLift"
layout: "aws"
page_title: "AWS: aws_gamelift_container_fleet"
description: |-
  Provides a GameLift Container Fleet resource.
---

# Resource: aws_gamelift_container_fleet

Provides a GameLift Container Fleet resource for managed containers.

## Example Usage

```terraform
resource "aws_gamelift_container_group_definition" "game_server" {
  name                   = "example-game-server"
  container_group_type   = "GAME_SERVER"
  operating_system       = "AMAZON_LINUX_2023"
  total_memory_limit_mib = 4096
  total_vcpu_limit       = 2

  game_server_container_definition {
    container_name     = "game-server"
    image_uri          = "123456789012.dkr.ecr.us-west-2.amazonaws.com/game-server:latest"
    server_sdk_version = "5.2.0"

    port_configuration {
      container_port_ranges {
        from_port = 7777
        to_port   = 7777
        protocol  = "UDP"
      }
    }
  }
}

resource "aws_gamelift_container_fleet" "example" {
  fleet_role_arn = aws_iam_role.gamelift.arn
  billing_type   = "ON_DEMAND"

  game_server_container_group_definition_name = aws_gamelift_container_group_definition.game_server.arn
  game_server_container_groups_per_instance   = 1

  instance_type = "c5.large"

  instance_connection_port_range {
    from_port = 4192
    to_port   = 4200
  }

  instance_inbound_permission {
    from_port = 4192
    to_port   = 4200
    protocol  = "UDP"
    ip_range  = "0.0.0.0/0"
  }

  log_configuration {
    log_destination = "CLOUDWATCH"
    log_group_arn   = aws_cloudwatch_log_group.gamelift.arn
  }

  metric_groups = ["example"]

  new_game_session_protection_policy = "NoProtection"

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `fleet_role_arn` - (Required) IAM role ARN with the `GameLiftContainerFleetPolicy` managed policy attached.
- `billing_type` - (Optional) Billing type for fleet instances. Valid values: `ON_DEMAND`, `SPOT`. Defaults to `ON_DEMAND`.
- `description` - (Optional) Fleet description.
- `game_server_container_group_definition_name` - (Required) Name or ARN of the game server container group definition to deploy.
- `game_server_container_groups_per_instance` - (Optional) Number of game server container groups per instance.
- `game_session_creation_limit_policy` - (Optional) Game session creation limit policy.
- `instance_connection_port_range` - (Optional) Connection port range for instances.
- `instance_inbound_permission` - (Optional) Inbound permissions for fleet instances.
- `instance_type` - (Optional) EC2 instance type to use for the fleet.
- `locations` - (Optional) Locations to deploy fleet instances to.
- `log_configuration` - (Optional) Log configuration.
- `metric_groups` - (Optional) CloudWatch metric group names.
- `new_game_session_protection_policy` - (Optional) Game session protection policy for new sessions. Defaults to `NoProtection`.
- `per_instance_container_group_definition_name` - (Optional) Name or ARN of the per-instance container group definition to deploy.
- `remove_per_instance_container_group_definition` - (Optional) When `true`, removes the per-instance container group definition.
- `deployment_configuration` - (Optional) Deployment configuration settings.
- `tags` - (Optional) Key-value tags to assign.

### game_session_creation_limit_policy

- `new_game_sessions_per_creator` - (Optional) Number of game sessions a creator can create during the period.
- `policy_period_in_minutes` - (Optional) Period in minutes.

### instance_connection_port_range

- `from_port` - (Required) Start port.
- `to_port` - (Required) End port.

### instance_inbound_permission

- `from_port` - (Required) Start port.
- `to_port` - (Required) End port.
- `protocol` - (Required) Protocol (`TCP` or `UDP`).
- `ip_range` - (Required) CIDR range.

### locations

- `location` - (Required) AWS Region or Local Zone name.

### log_configuration

- `log_destination` - (Optional) Log destination (`CLOUDWATCH`, `S3`, `NONE`).
- `log_group_arn` - (Optional) CloudWatch Log Group ARN.
- `s3_bucket_name` - (Optional) S3 bucket name for logs.

### deployment_configuration

- `impairment_strategy` - (Optional) Deployment impairment strategy (`MAINTAIN`, `ROLLBACK`).
- `minimum_healthy_percentage` - (Optional) Minimum healthy percentage (0-100).
- `protection_strategy` - (Optional) Deployment protection strategy (`WITH_PROTECTION`, `IGNORE_PROTECTION`).

## Attribute Reference

This resource exports the following attributes:

- `id` - Container fleet ID.
- `arn` - Container fleet ARN.
- `status` - Container fleet status.
- `game_server_container_group_definition_arn` - ARN of the game server container group definition in use.
- `per_instance_container_group_definition_arn` - ARN of the per-instance container group definition in use.
- `maximum_game_server_container_groups_per_instance` - Maximum container groups per instance.
- `location_attributes` - Location attributes for the fleet.
- `deployment_details` - Deployment details for the fleet.

## Import

In Terraform v1.5.0 and later, use an [`import` block`](https://developer.hashicorp.com/terraform/language/import) to import container fleets using the fleet ID. For example:

```terraform
import {
  to = aws_gamelift_container_fleet.example
  id = "fleet-a1234567-b8c9-0d1e-2fa3-b45c6d7e8912"
}
```

Using `terraform import`, import container fleets using the fleet ID. For example:

```console
% terraform import aws_gamelift_container_fleet.example fleet-a1234567-b8c9-0d1e-2fa3-b45c6d7e8912
```
