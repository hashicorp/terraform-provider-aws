---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_snapshot_create_volume_permission"
description: |-
  Adds create volume permission to an EBS Snapshot
---

# Resource: aws_snapshot_create_volume_permission

Adds permission to create volumes off of a given EBS Snapshot.

## Example Usage

```terraform
resource "aws_snapshot_create_volume_permission" "example_perm" {
  snapshot_id = aws_ebs_snapshot.example_snapshot.id
  account_id  = "12345678"
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size              = 40
}

resource "aws_ebs_snapshot" "example_snapshot" {
  volume_id = aws_ebs_volume.example.id
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `snapshot_id` - (Required) A snapshot ID
* `account_id` - (Required) An AWS Account ID to add create volume permissions. The AWS Account cannot be the snapshot's owner

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A combination of "`snapshot_id`-`account_id`".
