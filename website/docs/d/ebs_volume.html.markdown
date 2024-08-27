---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_volume"
description: |-
  Get information on an EBS volume.
---

# Data Source: aws_ebs_volume

Use this data source to get information about an EBS volume for use in other
resources.

## Example Usage

```terraform
data "aws_ebs_volume" "ebs_volume" {
  most_recent = true

  filter {
    name   = "volume-type"
    values = ["gp2"]
  }

  filter {
    name   = "tag:Name"
    values = ["Example"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `most_recent` - (Optional) If more than one result is returned, use the most
recent Volume.
* `filter` - (Optional) One or more name/value pairs to filter off of. There are
several valid keys, for a full reference, check out
[describe-volumes in the AWS CLI reference][1].

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Volume ID (e.g., vol-59fcb34e).
* `volume_id` - Volume ID (e.g., vol-59fcb34e).
* `arn` - Volume ARN (e.g., arn:aws:ec2:us-east-1:0123456789012:volume/vol-59fcb34e).
* `availability_zone` - AZ where the EBS volume exists.
* `encrypted` - Whether the disk is encrypted.
* `iops` - Amount of IOPS for the disk.
* `multi_attach_enabled` - (Optional) Specifies whether Amazon EBS Multi-Attach is enabled.
* `size` - Size of the drive in GiBs.
* `snapshot_id` - Snapshot_id the EBS volume is based off.
* `outpost_arn` - ARN of the Outpost.
* `volume_type` - Type of EBS volume.
* `kms_key_id` - ARN for the KMS encryption key.
* `tags` - Map of tags for the resource.
* `throughput` - Throughput that the volume supports, in MiB/s.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `read` - (Default `20m`)

[1]: http://docs.aws.amazon.com/cli/latest/reference/ec2/describe-volumes.html
