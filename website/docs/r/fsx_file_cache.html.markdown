---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_file_cache"
description: |-
  Terraform resource for managing an Amazon File Cache cache.
---

# Resource: aws_fsx_file_cache

Terraform resource for managing an Amazon File Cache cache.
See the [Create File Cache](https://docs.aws.amazon.com/fsx/latest/APIReference/API_CreateFileCache.html) for more information.

## Example Usage

```terraform
resource "aws_fsx_file_cache" "example" {

  data_repository_association {
    data_repository_path           = "nfs://filer.domain.com"
    data_repository_subdirectories = ["test", "test2"]
    file_cache_path                = "/ns1"

    nfs {
      dns_ips = ["192.168.0.1", "192.168.0.2"]
      version = "NFS3"
    }
  }

  file_cache_type         = "LUSTRE"
  file_cache_type_version = "2.12"

  lustre_configuration {
    deployment_type = "CACHE_1"
    metadata_configuration {
      storage_capacity = 2400
    }
    per_unit_storage_throughput   = 1000
    weekly_maintenance_start_time = "2:05:00"
  }

  subnet_ids       = [aws_subnet.test1.id]
  storage_capacity = 1200
}
```

## Argument Reference

This resource supports the following arguments:

* `copy_tags_to_data_repository_associations` - (Optional) Whether to copy tags for the cache to data repository associations. Defaults to `false`.
* `data_repository_association` - (Optional) Configurations for up to 8 data repository associations (DRAs) to create during cache creation. All configurations must be of the same data repository type, either all S3 or all NFS. Maximum of 8. See [`data_repository_association` Block](#data_repository_association-block) below.
* `file_cache_type` - (Required) Type of cache to create. The only supported value is `LUSTRE`.
* `file_cache_type_version` - (Required) Version for the type of cache to create. The only supported value is `2.12`.
* `kms_key_id` - (Optional) ID of the AWS Key Management Service (KMS) key to use for encrypting data on the cache. Defaults to the Amazon FSx-managed KMS key for your account.
* `lustre_configuration` - (Optional) Configuration for the Lustre cache. Required when `file_cache_type` is `LUSTRE`. See [`lustre_configuration` Block](#lustre_configuration-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `security_group_ids` - (Optional) IDs of the security groups to apply to all network interfaces created for cache access.
* `storage_capacity` - (Required) Storage capacity of the cache in gibibytes (GiB). Valid values are `1200` GiB, `2400` GiB, and increments of `2400` GiB.
* `subnet_ids` - (Required) Subnet IDs that the cache is accessible from. You can specify only one subnet ID.
* `tags` - (Optional) Map of tags to assign to the file cache. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `data_repository_association` Block

The `data_repository_association` configuration block supports the following arguments:

* `data_repository_path` - (Required) Path to the S3 or NFS data repository that links to the cache.
* `data_repository_subdirectories` - (Optional) NFS exports linked with this data repository association, in the format `/exportpath1`. Configure `data_repository_path` as the domain name of the NFS file system to use this argument. Not supported for S3 data repositories. Maximum of 500.
* `file_cache_path` - (Required) Path on the cache that maps 1-1 with `data_repository_path`. Must begin with a forward slash and cannot overlap the cache path of another data repository association.
* `nfs` - (Optional) Configuration for a data repository association linked to an NFS file system. See [`nfs` Block](#nfs-block) below.

### `nfs` Block

The `nfs` configuration block supports the following arguments:

* `dns_ips` - (Optional) Up to 2 IP addresses of DNS servers used to resolve the NFS file system domain name.
* `version` - (Required) Version of the NFS protocol of the NFS data repository. The only supported value is `NFS3`.

### `lustre_configuration` Block

The `lustre_configuration` configuration block supports the following arguments:

* `deployment_type` - (Required) Cache deployment type. The only supported value is `CACHE_1`.
* `metadata_configuration` - (Required) Configuration for a Lustre MDT (Metadata Target) storage volume. See [`metadata_configuration` Block](#lustre_configurationmetadata_configuration-block) below.
* `per_unit_storage_throughput` - (Required) Throughput provisioned for each 1 tebibyte (TiB) of cache storage capacity, in MB/s/TiB. The only supported value is `1000`.
* `weekly_maintenance_start_time` - (Optional) Recurring weekly time to start maintenance, in the format `D:HH:MM`. `D` is the day of the week, where `1` represents Monday and `7` represents Sunday. `HH` is the zero-padded hour of the day (0-23), and `MM` is the zero-padded minute of the hour. See the [ISO week date](https://en.wikipedia.org/wiki/ISO_week_date) for more information.

### `lustre_configuration.metadata_configuration` Block

The `metadata_configuration` configuration block supports the following arguments:

* `storage_capacity` - (Required) Storage capacity of the Lustre MDT (Metadata Target) storage volume in gibibytes (GiB). The only supported value is `2400` GiB.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the cache.
* `data_repository_association_ids` - IDs of data repository associations that are associated with the cache.
* `dns_name` - Domain Name System (DNS) name for the cache.
* `file_cache_id` - System-generated, unique ID of the cache.
* `id` - System-generated, unique ID of the cache.
* `network_interface_ids` - IDs of the network interfaces.
* `owner_id` - AWS account that created the cache.
* `vpc_id` - ID of your virtual private cloud (VPC).

### `data_repository_association` Block

* `association_id` - System-generated, unique ID of the data repository association.
* `file_cache_id` - System-generated, unique ID of the cache.
* `file_system_id` - ID of the file system for an NFS data repository association.
* `file_system_path` - Path to the data repository on the file system.
* `imported_file_chunk_size` - Size, in mebibytes (MiB), of the data blocks used to represent imported files.
* `resource_arn` - Amazon Resource Name (ARN) of the data repository association.
* `tags` - Map of tags assigned to the data repository association.

### `lustre_configuration` Block

* `log_configuration` - Configuration for Lustre logging used to write the enabled logging events for the cache.
* `mount_name` - Mount name of the cache.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon File Cache cache using the resource `id`. For example:

```terraform
import {
  to = aws_fsx_file_cache.example
  id = "fc-8012925589"
}
```

Using `terraform import`, import Amazon File Cache cache using the resource `id`. For example:

```console
% terraform import aws_fsx_file_cache.example fc-8012925589
```
