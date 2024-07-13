---
subcategory: "DRS (Elastic Disaster Recovery)"
layout: "aws"
page_title: "AWS: drs_replication_configuration_template"
description: |-
  Provides an Elastic Disaster Recovery replication configuration template resource.
---

# Resource: aws_drs_replication_configuration_template

Provides an Elastic Disaster Recovery replication configuration template resource. Before using DRS, your account must be [initialized](https://docs.aws.amazon.com/drs/latest/userguide/getting-started-initializing.html).

~> **NOTE:** Your configuration must use the PIT policy shown in the [basic configuration](#basic-configuration) due to AWS rules. The only value that you can change is the `retention_duration` of `rule_id` 3.

## Example Usage

### Basic configuration

```terraform
resource "aws_drs_replication_configuration_template" "example" {
  associate_default_security_group        = false
  bandwidth_throttling                    = 12
  create_public_ip                        = false
  data_plane_routing                      = "PRIVATE_IP"
  default_large_staging_disk_type         = "GP2"
  ebs_ecryption                           = "DEFAULT"
  ebs_encryption_key_arn                  = "arn:aws:kms:us-east-1:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"
  replication_server_instance_type        = "t3.small"
  replication_servers_security_groups_ids = aws_security_group.example[*].id
  staging_area_subnet_id                  = aws_subnet.example.id
  use_dedicated_replication_server        = false

  pit_policy {
    enabled            = true
    interval           = 10
    retention_duration = 60
    units              = "MINUTE"
    rule_id            = 1
  }

  pit_policy {
    enabled            = true
    interval           = 1
    retention_duration = 24
    units              = "HOUR"
    rule_id            = 2
  }

  pit_policy {
    enabled            = true
    interval           = 1
    retention_duration = 3
    units              = "DAY"
    rule_id            = 3
  }
}
```

## Argument Reference

The following arguments are required:

* `associate_default_security_group` - (Required) Whether to associate the default Elastic Disaster Recovery Security group with the Replication Configuration Template.
* `bandwidth_throttling` - (Required) Configure bandwidth throttling for the outbound data transfer rate of the Source Server in Mbps.
* `create_public_ip` - (Required) Whether to create a Public IP for the Recovery Instance by default.
* `data_plane_routing` - (Required) Data plane routing mechanism that will be used for replication. Valid values are `PUBLIC_IP` and `PRIVATE_IP`.
* `default_large_staging_disk_type` - (Required) Staging Disk EBS volume type to be used during replication. Valid values are `GP2`, `GP3`, `ST1`, or `AUTO`.
* `ebs_encryption` - (Required) Type of EBS encryption to be used during replication. Valid values are `DEFAULT` and `CUSTOM`.
* `ebs_encryption_key_arn` - (Required) ARN of the EBS encryption key to be used during replication.
* `pit_policy` - (Required) Configuration block for Point in time (PIT) policy to manage snapshots taken during replication. [See below](#pit_policy).
* `replication_server_instance_type` - (Required) Instance type to be used for the replication server.
* `replication_servers_security_groups_ids` - (Required) Security group IDs that will be used by the replication server.
* `staging_area_subnet_id` - (Required) Subnet to be used by the replication staging area.
* `staging_area_tags` - (Required) Set of tags to be associated with all resources created in the replication staging area: EC2 replication server, EBS volumes, EBS snapshots, etc.
* `use_dedicated_replication_server` - (Required) Whether to use a dedicated Replication Server in the replication staging area.

The following arguments are optional:

* `auto_replicate_new_disks` - (Optional) Whether to allow the AWS replication agent to automatically replicate newly added disks.
* `tags` - (Optional) Set of tags to be associated with the Replication Configuration Template resource.

### `pit_policy`

The PIT policies _must_ be specified as shown in the [basic configuration example](#basic-configuration) above. The only value that you can change is the `retention_duration` of `rule_id` 3.

* `enabled` - (Optional) Whether this rule is enabled or not.
* `interval` - (Required) How often, in the chosen units, a snapshot should be taken.
* `retention_duration` - (Required) Duration to retain a snapshot for, in the chosen `units`.
* `rule_id` - (Optional) ID of the rule. Valid values are integers.
* `units` - (Required) Units used to measure the `interval` and `retention_duration`. Valid values are `MINUTE`, `HOUR`, and `DAY`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Replication configuration template ARN.
* `id` - Replication configuration template ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `20m`)
- `update` - (Default `20m`)
- `delete` - (Default `20m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DRS Replication Configuration Template using the `id`. For example:

```terraform
import {
  to = aws_drs_replication_configuration_template.example
  id = "templateid"
}
```

Using `terraform import`, import DRS Replication Configuration Template using the `id`. For example:

```console
% terraform import aws_drs_replication_configuration_template.example templateid
```
