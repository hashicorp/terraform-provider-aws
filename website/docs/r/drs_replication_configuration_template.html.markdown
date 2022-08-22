---
subcategory: "Elastic Disaster Recovery"
layout: "aws"
page_title: "AWS: drs_replication_configuration_template"
description: |-
  Provides an Elastic Disaster Recovery replication configuration template resource.
---

# Resource: aws_drs_replication_configuration_template

Provides an Elastic Disaster Recovery replication configuration template resource.

## Example Usage

### Basic configuration

```terraform
resource "aws_drs_replication_configuration_template" "example" {
  associate_default_security_group        = True or False
  bandwidth_throttling                    = 123
  create_public_ip                        = True or False
  data_plane_routing                      = "PRIVATE_IP" or "PUBLIC_IP"
  default_large_staging_disk_type         = "GP2"or "GP3 or "ST1 or "AUTO
  ebs_ecryption                           = "DEFAULT" or "CUSTOM"
  ebs_encryption_key_arn                  = "string"
  pit_policy                              = [{"enabled": True or False, "interval":123}]
  replication_server_instance_type        = "string"
  replication_servers_security_groups_ids = ["string"]
  staging_area_subnet_id                  = "string
  staging_area_tags                       = {"string": "string"}
  tags                                    = {"string": "string"}
  use_dedicated_replication_server        = True or False
}
```

## Argument Reference

The following arguments are required:

* `associate_default_security_group` - (Required)(boolean)  Whether to associate the default Elastic Disaster Recovery Security group with the Replication Configuration Template.
* `bandwidth_throttling` - (Required)(integer) Configure bandwidth throttling for the outbound data transfer rate of the Source Server in Mbps.
* `create_public_ip` (Required)(boolean) Whether to create a Public IP for the Recovery Instance by default.
* `data_plane_routing` (Required)(string) The data plane routing mechanism that will be used for replication.
* `default_large_staging_disk_type` (Required)(string) The Staging Disk EBS volume type to be used during replication.
* `ebs_encryption` (Required)(string) The type of EBS encryption to be used during replication.
* `ebs_encryption_key_arn` (Required)(string) The ARN of the EBS encryption key to be used during replication.
* `pit_policy`(Required)(list) The Point in time (PIT) policy to manage snapshots taken during replication.
* `replication_server_instance_type` (Required)(string) The instance type to be used for the replication server.
* `replication_servers_security_groups_ids` (Required)(list) The security group IDs that will be used by the replication server.
* `staging_area_subnet_id` (Required)(string) The subnet to be used by the replication staging area.
* `staging_area_tags` (Required)(dict) A set of tags to be associated with all resources created in the replication staging area: EC2 replication server, EBS volumes, EBS snapshots, etc.
* `use_dedicated_replication_server` (Required)(boolean) Whether to use a dedicated Replication Server in the replication staging area.

The following arguments are optional:

* `tags` (Optional)(dict) A set of tags to be associated with the Replication Configuration Template resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Replication Configuration Template ARN.
* `replication_configuration_template_id` - The Replication Configuration Template ID.



