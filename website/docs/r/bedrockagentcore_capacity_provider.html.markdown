---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_capacity_provider"
description: |-
  Manages a Bedrock AgentCore Capacity Provider.
---

# Resource: aws_bedrockagentcore_capacity_provider

Manages a Bedrock AgentCore Capacity Provider for [Runtime Instances](https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/runtime-instances-how-it-works.html). Capacity providers define the EC2 infrastructure that AgentCore manages in your account. Instances are launched when a runtime session is invoked.

~> **Note:** Changing the name, compute configuration, or permissions configuration replaces the capacity provider. Remove associated runtimes before deleting it. Deleting a capacity provider also deletes its sessions and persistent storage.

## Example Usage

```terraform
resource "aws_bedrockagentcore_capacity_provider" "example" {
  name = "example_capacity_provider"

  compute_configuration {
    ec2_configuration {
      launch_template_source {
        launch_parameters {
          operating_system = "LINUX_ARM64"
          instance_requirements {
            allowed_instance_types = ["c6g.medium"]
          }
        }
      }
      vpc_configuration {
        subnets         = [aws_subnet.example.id]
        security_groups = [aws_security_group.example.id]
      }
    }
  }

  permissions_configuration {
    capacity_provider_operator_role_arn = aws_iam_role.example.arn
  }
}
```

The operator role must trust `bedrock-agentcore.amazonaws.com` and have permission to manage the instance infrastructure. AWS provides the [BedrockAgentCoreRuntimeInstancesOperatorRolePolicy managed policy](https://docs.aws.amazon.com/aws-managed-policy/latest/reference/BedrockAgentCoreRuntimeInstancesOperatorRolePolicy.html) for this role. Ensure role policy attachments are created before the capacity provider, using `depends_on` when necessary.

## Argument Reference

This resource supports the following arguments:

* `compute_configuration` - (Required) Compute infrastructure. See [`compute_configuration`](#compute_configuration-block) below. Forces replacement when changed.
* `description` - (Optional) Description, between 1 and 4096 characters.
* `name` - (Required) Capacity provider name, 1 to 48 characters. Must begin with a letter and contain only letters, numbers, and underscores. Forces replacement when changed.
* `permissions_configuration` - (Required) Infrastructure IAM role. See [`permissions_configuration`](#permissions_configuration-block) below. Forces replacement when changed.
* `region` - (Optional) Region where this resource will be managed. Defaults to the provider region.
* `tags` - (Optional) Map of tags to assign to the resource. Tags with matching keys override provider [`default_tags`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `capacity_reservation_specification` Block

The `capacity_reservation_specification` block supports the following:

* `capacity_reservation_preference` - (Optional) Capacity Reservation preference. Valid values are `capacity-reservations-only`, `none`, and `open`.
* `capacity_reservation_target` - (Optional) Target Capacity Reservation or resource group. See [`capacity_reservation_target`](#capacity_reservation_target-block) below.

### `capacity_reservation_target` Block

The `capacity_reservation_target` block supports the following:

* `capacity_reservation_id` - (Optional) Capacity Reservation ID.
* `capacity_reservation_resource_group_arn` - (Optional) Capacity Reservation resource group ARN.

### `compute_configuration` Block

The `compute_configuration` block supports the following:

* `ec2_configuration` - (Required) EC2 infrastructure configuration. See [`ec2_configuration`](#ec2_configuration-block) below.

### `ebs_configuration` Block

The `ebs_configuration` block supports the following:

* `encrypted` - (Optional) Whether to encrypt the EBS volume.
* `iops` - (Optional) Provisioned IOPS, subject to the selected EBS volume type.
* `kms_key_id` - (Optional) KMS key identifier for volume encryption.
* `name` - (Required) Logical volume name, referenced by an agent runtime volume mount.
* `size_gib` - (Required) Volume size in GiB.
* `snapshot_id` - (Optional) EBS snapshot ID used to initialize the volume.
* `throughput` - (Optional) EBS throughput in MiB/s, for `gp3` volumes.
* `volume_type` - (Optional) EBS volume type. See the [AWS API reference](https://docs.aws.amazon.com/bedrock-agentcore-control/latest/APIReference/API_EbsVolumeConfiguration.html) for supported values. AWS defaults persistent volumes to `gp3`.

### `ec2_configuration` Block

The `ec2_configuration` block supports the following:

* `launch_template_source` - (Required) Source of the instance launch parameters. See [`launch_template_source`](#launch_template_source-block) below.
* `lifecycle_configuration` - (Optional) Instance idle timeout and maximum lifetime in seconds. AWS defaults are 900 and 28800 seconds, respectively. See [`lifecycle_configuration`](#lifecycle_configuration-block) below. Configure as a single-element list of objects.
* `root_volume` - (Optional) Root volume performance, encryption, and free space configuration. See [`root_volume`](#root_volume-block) below. Configure as a single-element list of objects.
* `volume` - (Optional) Named persistent EBS volume. Supports up to five volumes. See [`volume`](#volume-block) below.
* `vpc_configuration` - (Required) Subnets and security groups for managed instances. See [`vpc_configuration`](#vpc_configuration-block) below.

### `ephemeral_volume` Block

The `ephemeral_volume` block supports the following:

* `device_name` - (Optional) Device name for the mapping.
* `virtual_name` - (Optional) Instance store virtual device name, such as `ephemeral0`.

### `instance_requirements` Block

The `instance_requirements` block supports the following:

* `allowed_instance_types` - (Required) Set of allowed EC2 instance types. AWS supports up to 30 types.

### `launch_parameters` Block

The `launch_parameters` block supports the following:

* `capacity_reservation_specification` - (Optional) Capacity Reservation targeting configuration. See [`capacity_reservation_specification`](#capacity_reservation_specification-block) below.
* `ephemeral_volume` - (Optional) Instance store device mapping. Supports up to five mappings. See [`ephemeral_volume`](#ephemeral_volume-block) below.
* `instance_profile_arn` - (Optional) IAM instance profile ARN for the managed instances.
* `instance_requirements` - (Required) Allowed EC2 instance types. See [`instance_requirements`](#instance_requirements-block) below.
* `license_specification` - (Optional) License configuration to associate with managed instances. Supports up to five configurations. See [`license_specification`](#license_specification-block) below.
* `monitoring` - (Optional) EC2 monitoring level. Valid values are `BASIC` and `DETAILED`.
* `operating_system` - (Required) Operating system and architecture. Valid values are `LINUX_ARM64` and `LINUX_X86_64`.
* `propagated_tags` - (Optional) Map of tags to propagate to managed EC2 resources.
* `ssh_key_name` - (Optional) EC2 key pair name for SSH access.

### `launch_template_source` Block

The `launch_template_source` block supports the following:

* `launch_parameters` - (Required) Operating system and instance launch settings. See [`launch_parameters`](#launch_parameters-block) below.

### `license_specification` Block

The `license_specification` block supports the following:

* `license_configuration_arn` - (Required) License configuration ARN.

### `lifecycle_configuration` Block

The `lifecycle_configuration` configuration supports the following:

* `idle_instance_timeout` - (Optional) Idle timeout in seconds.
* `max_lifetime` - (Optional) Maximum instance lifetime in seconds, up to 1209600 (14 days).

### `permissions_configuration` Block

The `permissions_configuration` block supports the following:

* `capacity_provider_operator_role_arn` - (Required) IAM role ARN that AgentCore assumes to manage instances.

### `root_volume` Block

The `root_volume` configuration supports the following:

* `encrypted` - (Optional) Whether to encrypt the EBS volume.
* `free_space_gib` - (Optional) Free space to guarantee on the root volume, in GiB. AWS defaults to 8 GiB.
* `iops` - (Optional) Provisioned IOPS, subject to the selected EBS volume type.
* `kms_key_id` - (Optional) KMS key identifier for volume encryption.
* `throughput` - (Optional) EBS throughput in MiB/s, for `gp3` volumes.
* `volume_type` - (Optional) EBS volume type. See the [AWS API reference](https://docs.aws.amazon.com/bedrock-agentcore-control/latest/APIReference/API_EbsVolumeConfiguration.html) for supported values. AWS defaults persistent volumes to `gp3`.

### `volume` Block

The `volume` block supports the following:

* `ebs_configuration` - (Required) Persistent EBS volume configuration. See [`ebs_configuration`](#ebs_configuration-block) below.

### `vpc_configuration` Block

The `vpc_configuration` block supports the following:

* `security_groups` - (Required) Set of security group IDs.
* `subnets` - (Required) Set of subnet IDs.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Capacity provider ARN.
* `id` - Capacity provider ID.
* `tags_all` - Map of tags assigned to the resource, including provider default tags.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute:

```terraform
import {
  to = aws_bedrockagentcore_capacity_provider.example
  identity = {
    id = "example_capacity_provider-abc1234567"
  }
}
```

### Identity Schema

#### Required

* `id` (String) Capacity provider ID.

#### Optional

* `account_id` (String) AWS account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a capacity provider using its ID:

```terraform
import {
  to = aws_bedrockagentcore_capacity_provider.example
  id = "example_capacity_provider-abc1234567"
}
```

Using `terraform import`, import a capacity provider using its ID:

```console
% terraform import aws_bedrockagentcore_capacity_provider.example example_capacity_provider-abc1234567
```
