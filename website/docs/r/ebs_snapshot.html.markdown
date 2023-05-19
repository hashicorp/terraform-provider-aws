---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_snapshot"
description: |-
  Provides an elastic block storage snapshot resource.
---

# Resource: aws_ebs_snapshot

Creates a Snapshot of an EBS Volume.

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
```

## Argument Reference

The following arguments are supported:

* `volume_id` - (Required) The Volume ID of which to make a snapshot.
* `description` - (Optional) A description of what the snapshot is.
* `outpost_arn` - (Optional) The Amazon Resource Name (ARN) of the Outpost on which to create a local snapshot.
* `storage_tier` - (Optional) The name of the storage tier. Valid values are `archive` and `standard`. Default value is `standard`.
* `permanent_restore` - (Optional) Indicates whether to permanently restore an archived snapshot.
* `temporary_restore_days` - (Optional) Specifies the number of days for which to temporarily restore an archived snapshot. Required for temporary restores only. The snapshot will be automatically re-archived after this period.
* `tags` - (Optional) A map of tags to assign to the snapshot. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the EBS Snapshot.
* `id` - The snapshot ID (e.g., snap-59fcb34e).
* `owner_id` - The AWS account ID of the EBS snapshot owner.
* `owner_alias` - Value from an Amazon-maintained list (`amazon`, `aws-marketplace`, `microsoft`) of snapshot owners.
* `encrypted` - Whether the snapshot is encrypted.
* `volume_size` - The size of the drive in GiBs.
* `kms_key_id` - The ARN for the KMS encryption key.
* `data_encryption_key_id` - The data encryption key identifier for the snapshot.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `10m`)
- `delete` - (Default `10m`)

## Import

EBS Snapshot can be imported using the `id`, e.g.,

```
$ terraform import aws_ebs_snapshot.id snap-049df61146c4d7901
```
