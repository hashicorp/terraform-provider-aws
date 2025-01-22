---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_snapshot_copy"
description: |-
  Terraform resource for managing an AWS Redshift Snapshot Copy.
---
# Resource: aws_redshift_snapshot_copy

Terraform resource for managing an AWS Redshift Snapshot Copy.

## Example Usage

### Basic Usage

```terraform
resource "aws_redshift_snapshot_copy" "example" {
  cluster_identifier = aws_redshift_cluster.example.id
  destination_region = "us-east-1"
}
```

## Argument Reference

The following arguments are required:

* `cluster_identifier` - (Required) Identifier of the source cluster.
* `destination_region` - (Required) AWS Region to copy snapshots to.

The following arguments are optional:

* `manual_snapshot_retention_period` - (Optional) Number of days to retain newly copied snapshots in the destination AWS Region after they are copied from the source AWS Region. If the value is `-1`, the manual snapshot is retained indefinitely.
* `retention_period` - (Optional) Number of days to retain automated snapshots in the destination region after they are copied from the source region.
* `snapshot_copy_grant_name` - (Optional) Name of the snapshot copy grant to use when snapshots of an AWS KMS-encrypted cluster are copied to the destination region.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the source cluster.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Snapshot Copy using the `id`. For example:

```terraform
import {
  to = aws_redshift_snapshot_copy.example
  id = "cluster-id-12345678"
}
```

Using `terraform import`, import Redshift Snapshot Copy using the `id`. For example:

```console
% terraform import aws_redshift_snapshot_copy.example cluster-id-12345678
```
