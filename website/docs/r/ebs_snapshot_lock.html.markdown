---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_snapshot_lock"
description: |-
  Provides an elastic block storage snapshot resource.
---

# Resource: aws_ebs_snapshot_lock

Locks a Snapshot of an EBS Volume.

## Example Usage

```terraform
resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size              = 40
}

resource "aws_ebs_snapshot" "example" {
  volume_id = aws_ebs_volume.example.id
}

aresource "aws_ebs_snapshot_lock" "example" {
  snapshot_id = aws_ebs_snapshot.example.id
  lock_mode   = "governance"
}

```

## Argument Reference

This resource supports the following arguments:

* `snapshot_id` - (Required)
* `lock_mode` - (Required)
* `cool_off_period` - (Optional)
* `lock_duration` - (Optional)
* `expiration_date` - (Optional)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `lock_created_on` -
* `cool_off_period_expires_on` -
* `lock_duration_start_time` -
* `lock_state` - 

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an EBS Snapshot Lock using the Snapshot `id`. For example:

```terraform
import {
  to = aws_ebs_snapshot_lock.example
  id = "snap-049df61146c4d7901"
}
```

Using `terraform import`, import an EBS Snapshot Lock using the Snapshot `id`. For example:

```console
% terraform import aws_ebs_snapshot.example snap-049df61146c4d7901
```
