---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_nfs"
description: |-
  Manages an AWS DataSync NFS Location
---

# Resource: aws_datasync_location_nfs

Manages an NFS Location within AWS DataSync.

~> **NOTE:** The DataSync Agents must be available before creating this resource.

## Example Usage

```terraform
resource "aws_datasync_location_nfs" "example" {
  server_hostname = "nfs.example.com"
  subdirectory    = "/exported/path"

  on_prem_config {
    agent_arns = [aws_datasync_agent.example.arn]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `mount_options` - (Optional) Configuration block containing mount options used by DataSync to access the NFS Server.
* `on_prem_config` - (Required) Configuration block containing information for connecting to the NFS File System.
* `server_hostname` - (Required) Specifies the IP address or DNS name of the NFS server. The DataSync Agent(s) use this to mount the NFS server.
* `subdirectory` - (Required) Subdirectory to perform actions as source or destination. Should be exported by the NFS server.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### mount_options Argument Reference

The `mount_options` configuration block supports the following arguments:

* `version` - (Optional) The specific NFS version that you want DataSync to use for mounting your NFS share. Valid values: `AUTOMATIC`, `NFS3`, `NFS4_0` and `NFS4_1`. Default: `AUTOMATIC`

### on_prem_config Argument Reference

The `on_prem_config` configuration block supports the following arguments:

* `agent_arns` - (Required) List of Amazon Resource Names (ARNs) of the DataSync Agents used to connect to the NFS server.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Amazon Resource Name (ARN) of the DataSync Location.
* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_datasync_location_nfs` using the DataSync Task Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_datasync_location_nfs.example
  id = "arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567"
}
```

Using `terraform import`, import `aws_datasync_location_nfs` using the DataSync Task Amazon Resource Name (ARN). For example:

```console
% terraform import aws_datasync_location_nfs.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
