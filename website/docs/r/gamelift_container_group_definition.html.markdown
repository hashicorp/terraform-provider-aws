---
subcategory: "GameLift"
layout: "aws"
page_title: "AWS: aws_gamelift_container_group_definition"
description: |-
  Provides a GameLift Container Group Definition resource.
---

# Resource: aws_gamelift_container_group_definition

Provides a GameLift Container Group Definition resource for managed containers.

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

  support_container_definitions {
    container_name = "metrics"
    image_uri      = "123456789012.dkr.ecr.us-west-2.amazonaws.com/metrics:latest"
    essential      = true
  }

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

This resource supports the following arguments:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `name` - (Required) Name of the container group definition. Must be unique per Region.
- `container_group_type` - (Optional) Container group type. Valid values: `GAME_SERVER`, `PER_INSTANCE`. Defaults to `GAME_SERVER`.
- `operating_system` - (Optional) Container operating system. Valid values: `AMAZON_LINUX_2023`. Defaults to `AMAZON_LINUX_2023`.
- `total_memory_limit_mib` - (Required) Total memory limit (MiB) for the container group.
- `total_vcpu_limit` - (Required) Total vCPU limit for the container group.
- `game_server_container_definition` - (Optional) Game server container definition. Required when `container_group_type` is `GAME_SERVER`.
- `support_container_definitions` - (Optional) List of support container definitions.
- `version_description` - (Optional) Description for the container group definition version.
- `tags` - (Optional) Key-value tags to assign.

### game_server_container_definition

- `container_name` - (Required) Container name.
- `image_uri` - (Required) ECR image URI.
- `port_configuration` - (Required) Port configuration.
- `server_sdk_version` - (Required) GameLift server SDK version.
- `depends_on` - (Optional) Container dependencies.
- `environment_override` - (Optional) Environment overrides.
- `mount_points` - (Optional) Container mount points.

### support_container_definitions

- `container_name` - (Required) Container name.
- `image_uri` - (Required) ECR image URI.
- `port_configuration` - (Optional) Port configuration.
- `depends_on` - (Optional) Container dependencies.
- `environment_override` - (Optional) Environment overrides.
- `essential` - (Optional) Whether the container is essential.
- `health_check` - (Optional) Health check configuration.
- `memory_hard_limit_mib` - (Optional) Memory limit (MiB).
- `mount_points` - (Optional) Container mount points.
- `vcpu` - (Optional) vCPU allocation.

### port_configuration

- `container_port_ranges` - (Required) List of container port ranges.

### container_port_ranges

- `from_port` - (Required) Start port.
- `to_port` - (Required) End port.
- `protocol` - (Required) Protocol (`TCP` or `UDP`).

### depends_on

- `container_name` - (Required) Dependency container name.
- `condition` - (Required) Dependency condition (`START`, `COMPLETE`, `SUCCESS`, `HEALTHY`).

### environment_override

- `name` - (Required) Environment variable name.
- `value` - (Required) Environment variable value.

### mount_points

- `instance_path` - (Required) Host path.
- `access_level` - (Optional) Access level (`READ_ONLY`, `READ_AND_WRITE`).
- `container_path` - (Optional) Container path.

### health_check

- `command` - (Required) List of command arguments.
- `interval` - (Optional) Interval in seconds.
- `retries` - (Optional) Number of retries.
- `start_period` - (Optional) Grace period in seconds.
- `timeout` - (Optional) Timeout in seconds.

## Attribute Reference

This resource exports the following attributes:

- `id` - Container group definition ID in the format `name,version`.
- `arn` - Container group definition ARN.
- `version_number` - Container group definition version.
- `status` - Container group definition status.
- `status_reason` - Status details when `status` is `FAILED`.
- `creation_time` - Creation timestamp in RFC3339 format.

## Import

In Terraform v1.5.0 and later, use an [`import` block`](https://developer.hashicorp.com/terraform/language/import) to import container group definitions using the `name,version` ID format. For example:

```terraform
import {
  to = aws_gamelift_container_group_definition.example
  id = "example-game-server,1"
}
```

Using `terraform import`, import container group definitions using the `name,version` ID format. For example:

```console
% terraform import aws_gamelift_container_group_definition.example example-game-server,1
```
