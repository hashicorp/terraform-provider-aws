---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_ontap_file_system"
description: |-
  Manages an Amazon FSx for NetApp ONTAP file system.
---

# Resource: aws_fsx_ontap_file_system

Manages an Amazon FSx for NetApp ONTAP file system.
See the [FSx ONTAP User Guide](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/what-is-fsx-ontap.html) for more information.

## Example Usage

```terraform
resource "aws_fsx_ontap_file_system" "test" {
  storage_capacity    = 1024
  subnet_ids          = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type     = "MULTI_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test1.id
}
```

```terraform
resource "aws_fsx_ontap_file_system" "testhapairs" {
  storage_capacity                = 2048
  subnet_ids                      = [aws_subnet.test1.id]
  deployment_type                 = "SINGLE_AZ_1"
  ha_pairs                        = 2
  throughput_capacity_per_ha_pair = 128
  preferred_subnet_id             = aws_subnet.test1.id
}
```

```terraform
resource "aws_fsx_ontap_file_system" "testsingleazgen2" {
  storage_capacity                = 4096
  subnet_ids                      = [aws_subnet.test1.id]
  deployment_type                 = "SINGLE_AZ_2"
  ha_pairs                        = 4
  throughput_capacity_per_ha_pair = 384
  preferred_subnet_id             = aws_subnet.test1.id
}
```

```terraform
resource "aws_fsx_ontap_file_system" "testmultiazgen2" {
  storage_capacity                = 1024
  subnet_ids                      = [aws_subnet.test1.id, aws_subnet.test2.id]
  deployment_type                 = "MULTI_AZ_2"
  ha_pairs                        = 1
  throughput_capacity_per_ha_pair = 384
  preferred_subnet_id             = aws_subnet.test1.id
}
```

## Argument Reference

This resource supports the following arguments:

* `storage_capacity` - (Required) The storage capacity (GiB) of the file system. Valid values between `1024` and `196608` for file systems with deployment_type `SINGLE_AZ_1` and `MULTI_AZ_1`. Valid values are between `1024` and `524288` for `MULTI_AZ_2`. Valid values between `1024` (`1024` per ha pair) and `1048576` for file systems with deployment_type `SINGLE_AZ_2`. For `SINGLE_AZ_2`, the `1048576` (1PB) maximum is only supported when using 2 or more ha_pairs, the maximum is `524288` (512TB) when using 1 ha_pair.
* `subnet_ids` - (Required) A list of IDs for the subnets that the file system will be accessible from. Up to 2 subnets can be provided.
* `preferred_subnet_id` - (Required) The ID for a subnet. A subnet is a range of IP addresses in your virtual private cloud (VPC).
* `security_group_ids` - (Optional) A list of IDs for the security groups that apply to the specified network interfaces created for file system access. These security groups will apply to all network interfaces.
* `weekly_maintenance_start_time` - (Optional) The preferred start time (in `d:HH:MM` format) to perform weekly maintenance, in the UTC time zone.
* `deployment_type` - (Optional) - The filesystem deployment type. Supports `MULTI_AZ_1`, `MULTI_AZ_2`, `SINGLE_AZ_1`, and `SINGLE_AZ_2`.
* `kms_key_id` - (Optional) ARN for the KMS Key to encrypt the file system at rest, Defaults to an AWS managed KMS Key.
* `automatic_backup_retention_days` - (Optional) The number of days to retain automatic backups. Setting this to 0 disables automatic backups. You can retain automatic backups for a maximum of 90 days.
* `daily_automatic_backup_start_time` - (Optional) A recurring daily time, in the format HH:MM. HH is the zero-padded hour of the day (0-23), and MM is the zero-padded minute of the hour. For example, 05:00 specifies 5 AM daily. Requires `automatic_backup_retention_days` to be set.
* `disk_iops_configuration` - (Optional) The SSD IOPS configuration for the Amazon FSx for NetApp ONTAP file system. See [Disk Iops Configuration](#disk-iops-configuration) below.
* `endpoint_ip_address_range` - (Optional) Specifies the IP address range in which the endpoints to access your file system will be created. By default, Amazon FSx selects an unused IP address range for you from the 198.19.* range.
* `ha_pairs` - (Optional) - The number of ha_pairs to deploy for the file system. Valid value is 1 for `SINGLE_AZ_1` or `MULTI_AZ_1` and `MULTI_AZ_2`. Valid values are 1 through 12 for `SINGLE_AZ_2`.
* `storage_type` - (Optional) - The filesystem storage type. defaults to `SSD`.
* `fsx_admin_password` - (Optional) The ONTAP administrative password for the fsxadmin user that you can use to administer your file system using the ONTAP CLI and REST API.
* `route_table_ids` - (Optional) Specifies the VPC route tables in which your file system's endpoints will be created. You should specify all VPC route tables associated with the subnets in which your clients are located. By default, Amazon FSx selects your VPC's default route table.
* `tags` - (Optional) A map of tags to assign to the file system. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `throughput_capacity` - (Optional) Sets the throughput capacity (in MBps) for the file system that you're creating. Valid values are `128`, `256`, `512`, `1024`, `2048`, and `4096`. This parameter is only supported when not using the ha_pairs parameter. Either throughput_capacity or throughput_capacity_per_ha_pair must be specified.
* `throughput_capacity_per_ha_pair` - (Optional) Sets the per-HA-pair throughput capacity (in MBps) for the file system that you're creating, as opposed to `throughput_capacity` which specifies the total throughput capacity for the file system. Valid value for `MULTI_AZ_1` and `SINGLE_AZ_1` are `128`, `256`, `512`, `1024`, `2048`, and `4096`. Valid values for deployment type `MULTI_AZ_2` and `SINGLE_AZ_2` are `384`,`768`,`1536`,`3072`,`6144` where `ha_pairs` is `1`. Valid values for deployment type `SINGLE_AZ_2` are `1536`, `3072`, and `6144` where `ha_pairs` is greater than 1. This parameter is only supported when specifying the ha_pairs parameter. Either throughput_capacity or throughput_capacity_per_ha_pair must be specified.

### Disk Iops Configuration

* `iops` - (Optional) - The total number of SSD IOPS provisioned for the file system.
* `mode` - (Optional) - Specifies whether the number of IOPS for the file system is using the system. Valid values are `AUTOMATIC` and `USER_PROVISIONED`. Default value is `AUTOMATIC`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name of the file system.
* `dns_name` - DNS name for the file system.

  **Note:** This attribute does not apply to FSx for ONTAP file systems and is consequently not set. You can access your FSx for ONTAP file system and volumes via a [Storage Virtual Machine (SVM)](fsx_ontap_storage_virtual_machine.html) using its DNS name or IP address.
* `endpoints` - The endpoints that are used to access data or to manage the file system using the NetApp ONTAP CLI, REST API, or NetApp SnapMirror. See [Endpoints](#endpoints) below.
* `id` - Identifier of the file system, e.g., `fs-12345678`
* `network_interface_ids` - Set of Elastic Network Interface identifiers from which the file system is accessible The first network interface returned is the primary network interface.
* `owner_id` - AWS account identifier that created the file system.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vpc_id` - Identifier of the Virtual Private Cloud for the file system.

### Endpoints

* `intercluster` - An endpoint for managing your file system by setting up NetApp SnapMirror with other ONTAP systems. See [Endpoint](#endpoint).
* `management` - An endpoint for managing your file system using the NetApp ONTAP CLI and NetApp ONTAP API. See [Endpoint](#endpoint).

#### Endpoint

* `dns_name` - The Domain Name Service (DNS) name for the file system. You can mount your file system using its DNS name.
* `ip_addresses` - IP addresses of the file system endpoint.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `60m`)
* `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import FSx File Systems using the `id`. For example:

```terraform
import {
  to = aws_fsx_ontap_file_system.example
  id = "fs-543ab12b1ca672f33"
}
```

Using `terraform import`, import FSx File Systems using the `id`. For example:

```console
% terraform import aws_fsx_ontap_file_system.example fs-543ab12b1ca672f33
```

Certain resource arguments, like `security_group_ids`, do not have a FSx API method for reading the information after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference. For example:

```terraform
resource "aws_fsx_ontap_file_system" "example" {
  # ... other configuration ...
  security_group_ids = [aws_security_group.example.id]

  # There is no FSx API for reading security_group_ids
  lifecycle {
    ignore_changes = [security_group_ids]
  }
}
```
