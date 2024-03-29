---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_snapshot"
description: |-
  Provides a Redshift Serverless Snapshot resource.
---

# Resource: aws_redshiftserverless_snapshot

Creates a new Amazon Redshift Serverless Snapshot.

## Example Usage

```terraform
resource "aws_redshiftserverless_snapshot" "example" {
  namespace_name = aws_redshiftserverless_workgroup.example.namespace_name
  snapshot_name  = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `namespace_name` - (Required) The namespace to create a snapshot for.
* `snapshot_name` - (Required) The name of the snapshot.
* `retention_period` - (Optional) How long to retain the created snapshot. Default value is `-1`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `accounts_with_provisioned_restore_access` - All of the Amazon Web Services accounts that have access to restore a snapshot to a provisioned cluster.
* `accounts_with_restore_access` - All of the Amazon Web Services accounts that have access to restore a snapshot to a namespace.
* `admin_username` - The username of the database within a snapshot.
* `arn` - The Amazon Resource Name (ARN) of the snapshot.
* `id` - The name of the snapshot.
* `kms_key_id` - The unique identifier of the KMS key used to encrypt the snapshot.
* `namespace_arn` - The Amazon Resource Name (ARN) of the namespace the snapshot was created from.
* `owner_account` - The owner Amazon Web Services; account of the snapshot.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Serverless Snapshots using the `snapshot_name`. For example:

```terraform
import {
  to = aws_redshiftserverless_snapshot.example
  id = "example"
}
```

Using `terraform import`, import Redshift Serverless Snapshots using the `snapshot_name`. For example:

```console
% terraform import aws_redshiftserverless_snapshot.example example
```
