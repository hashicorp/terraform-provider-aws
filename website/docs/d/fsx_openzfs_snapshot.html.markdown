---
subcategory: "FSx"
layout: "aws"
page_title: "AWS: aws_fsx_openzfs_snapshot"
description: |-
  Get information on an Amazon FSx for OpenZFS snapshot.
---

# Data Source: aws_fsx_openzfs_snapshot

Use this data source to get information about an Amazon FSx for OpenZFS Snapshot for use when provisioning new Volumes.

## Example Usage

### Root volume Example

```terraform
data "aws_fsx_openzfs_snapshot" "example" {
  most_recent = true

  filter {
    name   = "volume-id"
    values = ["fsvol-073a32b6098a73feb"]
  }
}
```

## Argument Reference

This data source supports the following arguments:

* `most_recent` - (Optional) If more than one result is returned, use the most recent snapshot.

* `snapshot_ids` - (Optional) Returns information on a specific snapshot_id.

* `filter` - (Optional) One or more name/value pairs to filter off of. The
supported names are file-system-id or volume-id.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name of the snapshot.
* `creation_time` - Time that the resource was created.
* `id` - Identifier of the snapshot, e.g., `fsvolsnap-12345678`
* `name` - Name of the snapshot.
* `snapshot_id` - ID of the snapshot.
* `tags` - List of Tag values, with a maximum of 50 elements.
* `volume_id` - ID of the volume that the snapshot is of.
