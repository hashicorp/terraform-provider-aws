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

## Argument Reference

The following arguments are supported:

* `storage_capacity` - (Optional) The storage capacity (GiB) of the file system. Valid values between `1024` and `196608`.
* `subnet_ids` - (Required) A list of IDs for the subnets that the file system will be accessible from. Upto 2 subnets can be provided.
* `preferred_subnet_id` - (Required) The ID for a subnet. A subnet is a range of IP addresses in your virtual private cloud (VPC).
* `security_group_ids` - (Optional) A list of IDs for the security groups that apply to the specified network interfaces created for file system access. These security groups will apply to all network interfaces.
* `weekly_maintenance_start_time` - (Optional) The preferred start time (in `d:HH:MM` format) to perform weekly maintenance, in the UTC time zone.
* `deployment_type` - (Optional) - The filesystem deployment type. Supports `MULTI_AZ_1` and `SINGLE_AZ_1`.
* `kms_key_id` - (Optional) ARN for the KMS Key to encrypt the file system at rest, Defaults to an AWS managed KMS Key.
* `automatic_backup_retention_days` - (Optional) The number of days to retain automatic backups. Setting this to 0 disables automatic backups. You can retain automatic backups for a maximum of 90 days.
* `daily_automatic_backup_start_time` - (Optional) A recurring daily time, in the format HH:MM. HH is the zero-padded hour of the day (0-23), and MM is the zero-padded minute of the hour. For example, 05:00 specifies 5 AM daily. Requires `automatic_backup_retention_days` to be set.
* `disk_iops_configuration` - (Optional) The SSD IOPS configuration for the Amazon FSx for NetApp ONTAP file system. See [Disk Iops Configuration](#disk-iops-configuration) Below.
* `endpoint_ip_address_range` - (Optional) Specifies the IP address range in which the endpoints to access your file system will be created. By default, Amazon FSx selects an unused IP address range for you from the 198.19.* range.
* `storage_type` - (Optional) - The filesystem storage type. defaults to `SSD`.
* `fsx_admin_password` - (Optional) The ONTAP administrative password for the fsxadmin user that you can use to administer your file system using the ONTAP CLI and REST API.
* `route_table_ids` - (Optional) Specifies the VPC route tables in which your file system's endpoints will be created. You should specify all VPC route tables associated with the subnets in which your clients are located. By default, Amazon FSx selects your VPC's default route table.
* `tags` - (Optional) A map of tags to assign to the file system. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `throughput_capacity` - (Required) Sets the throughput capacity (in MBps) for the file system that you're creating. Valid values are `128`, `256`, `512`, `1024`, and `2048`.

### Disk Iops Configuration

* `iops` - (Optional) - The total number of SSD IOPS provisioned for the file system.
* `mode` - (Optional) - Specifies whether the number of IOPS for the file system is using the system. Valid values are `AUTOMATIC` and `USER_PROVISIONED`. Default value is `AUTOMATIC`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the file system.
* `dns_name` - DNS name for the file system, e.g., `fs-12345678.fsx.us-west-2.amazonaws.com`
* `endpoints` - The endpoints that are used to access data or to manage the file system using the NetApp ONTAP CLI, REST API, or NetApp SnapMirror. See [Endpoints](#endpoints) below.
* `id` - Identifier of the file system, e.g., `fs-12345678`
* `network_interface_ids` - Set of Elastic Network Interface identifiers from which the file system is accessible The first network interface returned is the primary network interface.
* `owner_id` - AWS account identifier that created the file system.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).
* `vpc_id` - Identifier of the Virtual Private Cloud for the file system.

### Endpoints

* `intercluster` - An endpoint for managing your file system by setting up NetApp SnapMirror with other ONTAP systems. See [Endpoint](#endpoint).
* `management` - An endpoint for managing your file system using the NetApp ONTAP CLI and NetApp ONTAP API. See [Endpoint](#endpoint).

#### Endpoint

* `dns_name` - The Domain Name Service (DNS) name for the file system. You can mount your file system using its DNS name.
* `ip_addresses` - IP addresses of the file system endpoint.

## Timeouts

`aws_fsx_ontap_file_system` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

* `create` - (Default `60m`) How long to wait for the file system to be created.
* `update` - (Default `60m`) How long to wait for the file system to be updated.
* `delete` - (Default `60m`) How long to wait for the file system to be deleted.

## Import

FSx File Systems can be imported using the `id`, e.g.,

```
$ terraform import aws_fsx_ontap_file_system.example fs-543ab12b1ca672f33
```

Certain resource arguments, like `security_group_ids`, do not have a FSx API method for reading the information after creation. If the argument is set in the Terraform configuration on an imported resource, Terraform will always show a difference. To workaround this behavior, either omit the argument from the Terraform configuration or use [`ignore_changes`](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html#ignore_changes) to hide the difference, e.g.,

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
