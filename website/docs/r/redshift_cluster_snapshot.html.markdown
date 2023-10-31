---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_cluster_snapshot"
description: |-
  Creates a Redshift cluster snapshot
---

# Resource: aws_redshift_cluster_snapshot

Creates a Redshift cluster snapshot

## Example Usage

```terraform
resource "aws_redshift_cluster_snapshot" "example" {
  cluster_snapshot_name = "example"
  cluster_snapshot_content = jsonencode(
    {
      AllowDBUserOverride = "1"
      Client_ID           = "ExampleClientID"
      App_ID              = "example"
    }
  )
}
```

## Argument Reference

This resource supports the following arguments:

* `cluster_identifier` - (Required, Forces new resource) The cluster identifier for which you want a snapshot.
* `snapshot_identifier` - (Required, Forces new resource) A unique identifier for the snapshot that you are requesting. This identifier must be unique for all snapshots within the Amazon Web Services account.
* `manual_snapshot_retention_period` - (Optional) The number of days that a manual snapshot is retained. If the value is `-1`, the manual snapshot is retained indefinitely. Valid values are -1 and between `1` and `3653`.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the snapshot.
* `id` - A unique identifier for the snapshot that you are requesting. This identifiermust be unique for all snapshots within the Amazon Web Services account.
* `kms_key_id` - The Key Management Service (KMS) key ID of the encryption key that was used to encrypt data in the cluster from which the snapshot was taken.
* `owner_account` - For manual snapshots, the Amazon Web Services account used to create or copy the snapshot. For automatic snapshots, the owner of the cluster. The owner can perform all snapshot actions, such as sharing a manual snapshot.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Cluster Snapshots using `snapshot_identifier`. For example:

```terraform
import {
  to = aws_redshift_cluster_snapshot.test
  id = "example"
}
```

Using `terraform import`, import Redshift Cluster Snapshots using `snapshot_identifier`. For example:

```console
% terraform import aws_redshift_cluster_snapshot.test example
```
