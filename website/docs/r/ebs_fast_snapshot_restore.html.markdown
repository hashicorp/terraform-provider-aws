---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_fast_snapshot_restore"
description: |-
  Terraform resource for managing an EBS (Elastic Block Storage) Fast Snapshot Restore.
---

# Resource: aws_ebs_fast_snapshot_restore

Terraform resource for managing an EBS (Elastic Block Storage) Fast Snapshot Restore.

## Example Usage

### Basic Usage

```terraform
resource "aws_ebs_fast_snapshot_restore" "example" {
  availability_zone = "us-west-2a"
  snapshot_id       = aws_ebs_snapshot.example.id
}
```

## Argument Reference

The following arguments are required:

* `availability_zone` - (Required) Availability zone in which to enable fast snapshot restores.
* `snapshot_id` - (Required) ID of the snapshot.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string concatenating `availability_zone` and `snapshot_id`.
* `state` - State of fast snapshot restores. Valid values are `enabling`, `optimizing`, `enabled`, `disabling`, `disabled`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 (Elastic Compute Cloud) EBS Fast Snapshot Restore using the `example_id_arg`. For example:

```terraform
import {
  to = aws_ebs_fast_snapshot_restore.example
  id = "us-west-2a,snap-abcdef123456"
}
```

Using `terraform import`, import EC2 (Elastic Compute Cloud) EBS Fast Snapshot Restore using the `id`. For example:

```console
% terraform import aws_ebs_fast_snapshot_restore.example us-west-2a,snap-abcdef123456
```
