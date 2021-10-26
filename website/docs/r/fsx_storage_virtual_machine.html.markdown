---
subcategory: "File System (FSx)"
layout: "aws"
page_title: "AWS: aws_fsx_storage_virtual_machine"
description: |-
  Manages a storage virtual machine (SVM) for an Amazon FSx for ONTAP file system.
---

# Resource: aws_fsx_storage_virtual_machine

Manages a storage virtual machine (SVM) for an Amazon FSx for ONTAP file system.

## Example Usage

```terraform
resource "aws_fsx_storage_virtual_machine" "example" {
  name                       = "example"
  file_system_id             = aws_fsx_ontap_file_system.example.id
  root_volume_security_style = "UNIX"
}
```

## Argument Reference

The following arguments are supported:

* `file_system_id` - (Required) The ID of the file system, assigned by Amazon FSx.
* `name` - (Required) The name of the SVM.
* `root_volume_security_style` - (Optional) The security style of the root volume of the SVM. Valid values are `UNIX`, `NTFS`, and `MIXED`.
* `svm_admin_password` - (Optional) The password to use when managing the SVM using the NetApp ONTAP CLI or REST API. If you do not specify a password, you can still use the file system's fsxadmin user to manage the SVM.
* `tags` - (Optional) A map of tags to assign to the file system. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name of the file system.
* `endpoints` - The endpoints that are used to access data or to manage the file system using the NetApp ONTAP CLI, REST API, or NetApp SnapMirror. See [Endpoints](#endpoints) below.
* `id` - Identifier of the file system, e.g., `svm-12345678`.
* `uuid` - The SVM's UUID (universally unique identifier).
* `subtype` - Describes the SVM's subtype.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

### Endpoints

* `iscsi` - An endpoint for connecting using the Internet Small Computer Systems Interface (iSCSI) protocol. See [Endpoint](#endpoint).
* `management` - An endpoint for managing SVMs using the NetApp ONTAP CLI and NetApp ONTAP API. See [Endpoint](#endpoint).
* `nfs` - An endpoint for connecting using the Network File System (NFS) protocol. See [Endpoint](#endpoint).
* `smb` - An endpoint for connecting using the Server Message Block (SMB) protocol. See [Endpoint](#endpoint).

#### Endpoint

* `dns_name` - The Domain Name Service (DNS) name for the file system. You can mount your file system using its DNS name.
* `ip_addresses` - IP addresses of the file system endpoint.

## Timeouts

`aws_fsx_storage_virtual_machine` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

* `create` - (Default `60m`) How long to wait for the file system to be created.
* `update` - (Default `60m`) How long to wait for the file system to be updated.
* `delete` - (Default `60m`) How long to wait for the file system to be deleted.

## Import

FSx Storage Virtual Machines can be imported using the `id`, e.g.,

```
$ terraform import aws_fsx_storage_virtual_machine.example svm-543ab12b1ca672f33
```
