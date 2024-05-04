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

* `snapshot_id` - (Required) The ID of the snapshot to lock.
* `lock_mode` - (Required) The mode in which to lock the snapshot. Specify one of the following. Valid values are `compliance` and `governance`.
* `cool_off_period` - (Optional) The cooling-off period during which you can unlock the snapshot or modify the lock settings after locking the snapshot in compliance mode, in hours. You can increase the lock duration after the cooling-off period expires. The cooling-off period is optional when locking a snapshot in compliance mode. If you are locking the snapshot in governance mode, omit this parameter. To lock the snapshot in compliance mode immediately without a cooling-off period, omit this parameter. If you are extending the lock duration for a snapshot that is locked in compliance mode after the cooling-off period has expired, omit this parameter. Allowed values: Min 1, max 72.
* `lock_duration` - (Optional) The period of time for which to lock the snapshot, in days. You must specify either this parameter or `expiration_date`, but not both. Allowed values: Min: 1, max 36500.
* `expiration_date` - (Optional) The date and time at which the snapshot lock is to automatically expire, in the UTC time zone ( YYYY-MM-DDThh:mm:ss.sssZ ). You must specify either this parameter or `lock_duration`, but not both.

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
% terraform import aws_ebs_snapshot_lock.example snap-049df61146c4d7901
```
