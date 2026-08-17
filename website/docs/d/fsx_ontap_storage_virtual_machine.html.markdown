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

* `active_directory_configuration` - Microsoft Active Directory configuration to which the SVM is joined, if applicable. See [Active Directory Configuration](#active-directory-configuration) below.
* `arn` - Amazon Resource Name of the SVM.
* `creation_time` - Time that the SVM was created.
* `endpoints` - Endpoints that are used to access data or to manage the SVM using the NetApp ONTAP CLI, REST API, or NetApp CloudManager. They are the Iscsi, Management, Nfs, and Smb endpoints. See [SVM Endpoints](#svm-endpoints) below.
* `file_system_id` - Identifier of the file system (e.g. `fs-12345678`).
* `id` - SVM's system generated unique ID.
* `lifecycle_status` - SVM's lifecycle status.
* `lifecycle_transition_reason` - Reason why the SVM lifecycle state changed. See [Lifecycle Transition Reason](#lifecycle-transition-reason) below.
* `name` - Name of the SVM, if provisioned.
* `subtype` - SVM's subtype.
* `tags` - Map of tags assigned to the resource.
* `uuid` - SVM's UUID.

### Active Directory Configuration

The following arguments are supported for `active_directory_configuration` configuration block:

* `netbios_name` - NetBIOS name of the AD computer object to which the SVM is joined.
* `self_managed_active_directory` - Configuration of the self-managed Microsoft Active Directory (AD) directory to which the Windows File Server or ONTAP storage virtual machine (SVM) instance is joined. See [Self Managed Active Directory](#self-managed-active-directory) below.

### Self Managed Active Directory

* `dns_ips` - List of up to three IP addresses of DNS servers or domain controllers in the self-managed AD directory.
* `domain_name` - Fully qualified domain name of the self-managed AD directory.
* `file_system_administrators_group` - Name of the domain group whose members have administrative privileges for the FSx file system.
* `organizational_unit_distinguished_name` - Fully qualified distinguished name of the organizational unit within the self-managed AD directory to which the Windows File Server or ONTAP storage virtual machine (SVM) instance is joined.
* `username` - User name for the service account on your self-managed AD domain that FSx uses to join to your AD domain.

### Lifecycle Transition Reason

* `message` - Detailed message.

### SVM Endpoints

* `Iscsi` - Endpoint for connecting using the Internet Small Computer Systems Interface (iSCSI) protocol. See [SVM Endpoint](#svm-endpoint) below.
* `management` - Endpoint for managing SVMs using the NetApp ONTAP CLI, NetApp ONTAP API, or NetApp CloudManager. See [SVM Endpoint](#svm-endpoint) below.
* `nfs` - Endpoint for connecting using the Network File System (NFS) protocol. See [SVM Endpoint](#svm-endpoint) below.
* `smb` - Endpoint for connecting using the Server Message Block (SMB) protocol. See [SVM Endpoint](#svm-endpoint) below.

### SVM Endpoint

* `DNSName` - File system's DNS name. You can mount your file system using its DNS name.
* `IpAddresses` - SVM endpoint's IP addresses.
