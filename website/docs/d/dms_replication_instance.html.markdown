---
subcategory: "DMS (Database Migration)"
layout: "aws"
page_title: "AWS: aws_dms_certificate"
description: |-
  Terraform data source for managing an AWS DMS (Database Migration) Replication Instance.
---

# Data Source: aws_dms_replication_instance

Terraform data source for managing an AWS DMS (Database Migration) Replication Instance.

## Example Usage

```terraform
data "aws_dms_replication_instance" "test" {
  replication_instance_id = aws_dms_replication_instance.test.replication_instance_id
}
```

## Argument Reference

The following arguments are required:

* `replication_instance_id` - (Required) The replication instance identifier. This parameter is stored as a lowercase string.

    - Must contain from 1 to 63 alphanumeric characters or hyphens.
    - First character must be a letter.
    - Cannot end with a hyphen
    - Cannot contain two consecutive hyphens.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `allocated_storage` - (Default: 50, Min: 5, Max: 6144) The amount of storage (in gigabytes) to be initially allocated for the replication instance.
* `allow_major_version_upgrade` - (Default: false) Indicates that major version upgrades are allowed.
* `apply_immediately` - (Default: false) Indicates whether the changes should be applied immediately or during the next maintenance window. Only used when updating an existing resource.
* `auto_minor_version_upgrade` - (Default: false) Indicates that minor engine upgrades will be applied automatically to the replication instance during the maintenance window.
* `availability_zone` - The EC2 Availability Zone that the replication instance will be created in.
* `engine_version` - The engine version number of the replication instance.
* `kms_key_arn` - The Amazon Resource Name (ARN) for the KMS key that will be used to encrypt the connection parameters. If you do not specify a value for `kms_key_arn`, then AWS DMS will use your default encryption key. AWS KMS creates the default encryption key for your AWS account. Your AWS account has a different default encryption key for each AWS region.
* `multi_az` - Specifies if the replication instance is a multi-az deployment. You cannot set the `availability_zone` parameter if the `multi_az` parameter is set to `true`.
* `preferred_maintenance_window` - The weekly time range during which system maintenance can occur, in Universal Coordinated Time (UTC).

    - Default: A 30-minute window selected at random from an 8-hour block of time per region, occurring on a random day of the week.
    - Format: `ddd:hh24:mi-ddd:hh24:mi`
    - Valid Days: `mon, tue, wed, thu, fri, sat, sun`
    - Constraints: Minimum 30-minute window.

* `publicly_accessible` - (Default: false) Specifies the accessibility options for the replication instance. A value of true represents an instance with a public IP address. A value of false represents an instance with a private IP address.
* `replication_instance_arn` - The Amazon Resource Name (ARN) of the replication instance.
* `replication_instance_class` - The compute and memory capacity of the replication instance as specified by the replication instance class. See [AWS DMS User Guide](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_ReplicationInstance.Types.html) for available instance sizes and advice on which one to choose.
* `replication_instance_private_ips` - A list of the private IP addresses of the replication instance.
* `replication_instance_public_ips` - A list of the public IP addresses of the replication instance.
* `replication_subnet_group_id` - (Optional) A subnet group to associate with the replication instance.
* `vpc_security_group_ids` - (Optional) A list of VPC security group IDs to be used with the replication instance. The VPC security groups must work with the VPC containing the replication instance.
