---
subcategory: "DataSync"
layout: "aws"
page_title: "AWS: aws_datasync_location_smb"
description: |-
  Manages an AWS DataSync SMB Location
---

# Resource: aws_datasync_location_smb

Manages a SMB Location within AWS DataSync.

~> **NOTE:** The DataSync Agents must be available before creating this resource.

## Example Usage

```terraform
resource "aws_datasync_location_smb" "example" {
  server_hostname = "smb.example.com"
  subdirectory    = "/exported/path"

  user     = "Guest"
  password = "ANotGreatPassword"

  agent_arns = [aws_datasync_agent.example.arn]
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `agent_arns` - (Required) A list of DataSync Agent ARNs with which this location will be associated.
* `domain` - (Optional) The name of the Windows domain the SMB server belongs to.
* `mount_options` - (Optional) Configuration block containing mount options used by DataSync to access the SMB Server. Can be `AUTOMATIC`, `SMB2`, or `SMB3`.
* `password` - (Required) The password of the user who can mount the share and has file permissions in the SMB.
* `server_hostname` - (Required) Specifies the IP address or DNS name of the SMB server. The DataSync Agent(s) use this to mount the SMB share.
* `subdirectory` - (Required) Subdirectory to perform actions as source or destination. Should be exported by the NFS server.
* `tags` - (Optional) Key-value pairs of resource tags to assign to the DataSync Location. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user` - (Required) The user who can mount the share and has file and folder permissions in the SMB share.

### mount_options Argument Reference

The `mount_options` configuration block supports the following arguments:

* `version` - (Optional) The specific SMB version that you want DataSync to use for mounting your SMB share. Valid values: `AUTOMATIC`, `SMB2`, and `SMB3`. Default: `AUTOMATIC`

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the DataSync Location.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_datasync_location_smb` using the Amazon Resource Name (ARN). For example:

```terraform
import {
  to = aws_datasync_location_smb.example
  id = "arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567"
}
```

Using `terraform import`, import `aws_datasync_location_smb` using the Amazon Resource Name (ARN). For example:

```console
% terraform import aws_datasync_location_smb.example arn:aws:datasync:us-east-1:123456789012:location/loc-12345678901234567
```
