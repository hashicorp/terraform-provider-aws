---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_ontap_storage_virtual_machine"
description: |-
  Retrieve information on FSx ONTAP Storage Virtual Machine (SVM).
---

# Data Source: aws_fsx_ontap_storage_virtual_machine

Retrieve information on FSx ONTAP Storage Virtual Machine (SVM).

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

The arguments of this data source act as filters for querying the available ONTAP Storage Virtual Machines in the current region. The given filters must match exactly one Storage Virtual Machine whose data will be exported as attributes.

### filter

This block allows for complex filters.

The following arguments are required:

* `name` - (Required) Name of the field to filter by, as defined by [the underlying AWS API](https://docs.aws.amazon.com/fsx/latest/APIReference/API_StorageVirtualMachineFilter.html).
* `values` - (Required) Set of values that are accepted for the given field. An SVM will be selected if any one of the given values matches.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name of the SVM.
* `active_directory_configuration` - The Microsoft Active Directory configuration to which the SVM is joined, if applicable. See [Active Directory Configuration](#active-directory-configuration) below.
* `creation_time` - The time that the SVM was created.
* `endpoints` - The endpoints that are used to access data or to manage the SVM using the NetApp ONTAP CLI, REST API, or NetApp CloudManager. They are the Iscsi, Management, Nfs, and Smb endpoints. See [SVM Endpoints](#svm-endpoints) below.
* `file_system_id` - Identifier of the file system (e.g. `fs-12345678`).
* `id` - The SVM's system generated unique ID.
* `lifecycle_status` - The SVM's lifecycle status.
* `lifecycle_transition_reason` - Describes why the SVM lifecycle state changed. See [Lifecycle Transition Reason](#lifecycle-transition-reason) below.
* `name` - The name of the SVM, if provisioned.
* `subtype` - The SVM's subtype.
* `uuid` - The SVM's UUID.

### Active Directory Configuration

The following arguments are supported for `active_directory_configuration` configuration block:

* `netbios_name` - The NetBIOS name of the AD computer object to which the SVM is joined.
* `self_managed_active_directory` - The configuration of the self-managed Microsoft Active Directory (AD) directory to which the Windows File Server or ONTAP storage virtual machine (SVM) instance is joined. See [Self Managed Active Directory](#self-managed-active-directory) below.

### Self Managed Active Directory

* `dns_ips` - A list of up to three IP addresses of DNS servers or domain controllers in the self-managed AD directory.
* `domain_name` - The fully qualified domain name of the self-managed AD directory.
* `file_system_administrators_group` - The name of the domain group whose members have administrative privileges for the FSx file system.
* `organizational_unit_distinguished_name` - The fully qualified distinguished name of the organizational unit within the self-managed AD directory to which the Windows File Server or ONTAP storage virtual machine (SVM) instance is joined.
* `username` - The user name for the service account on your self-managed AD domain that FSx uses to join to your AD domain.

### Lifecycle Transition Reason

* `message` - A detailed message.

### SVM Endpoints

* `Iscsi` - An endpoint for connecting using the Internet Small Computer Systems Interface (iSCSI) protocol. See [SVM Endpoint](#svm-endpoint) below.
* `management` - An endpoint for managing SVMs using the NetApp ONTAP CLI, NetApp ONTAP API, or NetApp CloudManager. See [SVM Endpoint](#svm-endpoint) below.
* `nfs` - An endpoint for connecting using the Network File System (NFS) protocol. See [SVM Endpoint](#svm-endpoint) below.
* `smb` - An endpoint for connecting using the Server Message Block (SMB) protocol. See [SVM Endpoint](#svm-endpoint) below.

### SVM Endpoint

* `DNSName` - The file system's DNS name. You can mount your file system using its DNS name.
* `IpAddresses` - The SVM endpoint's IP addresses.
