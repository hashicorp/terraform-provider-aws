---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_snapshot_copy"
description: |-
  Duplicates an existing Amazon snapshot
---

# Resource: aws_ebs_snapshot_copy

Creates a Snapshot of a snapshot.

## Example Usage

```terraform
resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size              = 40

  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ebs_snapshot" "example_snapshot" {
  volume_id = aws_ebs_volume.example.id

  tags = {
    Name = "HelloWorld_snap"
  }
}

resource "aws_ebs_snapshot_copy" "example_copy" {
  source_snapshot_id = aws_ebs_snapshot.example_snapshot.id
  source_region      = "us-west-2"

  tags = {
    Name = "HelloWorld_copy_snap"
  }
}
```

## Argument Reference

The following arguments are supported:

* `description` - (Optional) A description of what the snapshot is.
* `encrypted` - Whether the snapshot is encrypted.
* `kms_key_id` - The ARN for the KMS encryption key.
* `source_snapshot_id` The ARN for the snapshot to be copied.
* `source_region` The region of the source snapshot.
* `storage_tier` - (Optional) The name of the storage tier. Valid values are `archive` and `standard`. Default value is `standard`.
* `permanent_restore` - (Optional) Indicates whether to permanently restore an archived snapshot.
* `temporary_restore_days` - (Optional) Specifies the number of days for which to temporarily restore an archived snapshot. Required for temporary restores only. The snapshot will be automatically re-archived after this period.
* `tags` - A map of tags for the snapshot. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EBS Snapshot.
* `id` - The snapshot ID (e.g., snap-59fcb34e).
* `owner_id` - The AWS account ID of the snapshot owner.
* `owner_alias` - Value from an Amazon-maintained list (`amazon`, `aws-marketplace`, `microsoft`) of snapshot owners.
* `volume_size` - The size of the drive in GiBs.
* `data_encryption_key_id` - The data encryption key identifier for the snapshot.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)
