---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_ontap_storage_virtual_machine"
description: |-
  Retrieve information on FSx ONTAP Storage Virtual Machine (SVM).
---

# Data Source: aws_fsx_ontap_storage_virtual_machine

Retrieve information on FSx ONTAP Storage Virtual Machine (SVM).

The arguments of this data source act as filters for querying the available ONTAP Storage Virtual Machines in the current region. The given filters must match exactly one Storage Virtual Machine whose data will be exported as attributes.

## Example Usage

### Basic Usage

```terraform
data "aws_fsx_ontap_storage_virtual_machine" "example" {
  id = "svm-12345678"
}
```

### Filter Example

```
data "aws_fsx_ontap_storage_virtual_machine" "example" {
  filter {
    name   = "file-system-id"
    values = ["fs-12345678"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `filter` - (Optional) Configuration block. Detailed below.
* `id` - (Optional) Identifier of the storage virtual machine (e.g. `svm-12345678`).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `filter` Block

This block allows for complex filters.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/fsx/latest/APIReference/API_StorageVirtualMachineFilter.html).
* `values` - (Required) Set of values that are accepted for the given field. An SVM will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `active_directory_configuration` - Microsoft Active Directory configuration to which the SVM is joined, if applicable. See [`active_directory_configuration` Block](#active_directory_configuration-block) below.
* `arn` - ARN of the SVM.
* `creation_time` - Time that the SVM was created.
* `endpoints` - Endpoints that are used to access data or to manage the SVM using the NetApp ONTAP CLI, REST API, or NetApp CloudManager. See [`endpoints` Block](#endpoints-block) below.
* `file_system_id` - Identifier of the file system (e.g. `fs-12345678`).
* `id` - SVM's system generated unique ID.
* `lifecycle_status` - SVM's lifecycle status.
* `lifecycle_transition_reason` - Reason why the SVM lifecycle state changed. See [`lifecycle_transition_reason` Block](#lifecycle_transition_reason-block) below.
* `name` - Name of the SVM, if provisioned.
* `subtype` - SVM's subtype.
* `tags` - Map of tags assigned to the resource.
* `uuid` - SVM's UUID.

### `active_directory_configuration` Block

* `netbios_name` - NetBIOS name of the AD computer object to which the SVM is joined.
* `self_managed_active_directory_configuration` - Configuration of the self-managed Microsoft Active Directory (AD) directory to which the SVM is joined. See [`self_managed_active_directory_configuration` Block](#self_managed_active_directory_configuration-block) below.

### `self_managed_active_directory_configuration` Block

* `dns_ips` - List of up to three IP addresses of DNS servers or domain controllers in the self-managed AD directory.
* `domain_name` - Fully qualified domain name of the self-managed AD directory.
* `file_system_administrators_group` - Name of the domain group whose members have administrative privileges for the FSx file system.
* `organizational_unit_distinguished_name` - Fully qualified distinguished name of the organizational unit within the self-managed AD directory to which the SVM is joined.
* `username` - User name for the service account on your self-managed AD domain that FSx uses to join to your AD domain.

### `endpoints` Block

* `iscsi` - Endpoint for connecting using the Internet Small Computer Systems Interface (iSCSI) protocol. See [`iscsi` Block](#iscsi-block) below.
* `management` - Endpoint for managing SVMs using the NetApp ONTAP CLI, NetApp ONTAP API, or NetApp CloudManager. See [`management` Block](#management-block) below.
* `nfs` - Endpoint for connecting using the Network File System (NFS) protocol. See [`nfs` Block](#nfs-block) below.
* `smb` - Endpoint for connecting using the Server Message Block (SMB) protocol. See [`smb` Block](#smb-block) below.

### `iscsi` Block

* `dns_name` - SVM endpoint's DNS name.
* `ip_addresses` - SVM endpoint's IP addresses.

### `management` Block

* `dns_name` - SVM endpoint's DNS name.
* `ip_addresses` - SVM endpoint's IP addresses.

### `nfs` Block

* `dns_name` - SVM endpoint's DNS name.
* `ip_addresses` - SVM endpoint's IP addresses.

### `smb` Block

* `dns_name` - SVM endpoint's DNS name.
* `ip_addresses` - SVM endpoint's IP addresses.

### `lifecycle_transition_reason` Block

* `message` - Detailed message.
