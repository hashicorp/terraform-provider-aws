---
subcategory: "Neptune Analytics"
layout: "aws"
page_title: "AWS: aws_neptunegraph_graph_snapshot"
description: |-
  Manages an Amazon Neptune Analytics graph snapshot.
---

# Resource: aws_neptunegraph_graph_snapshot

Manages an Amazon Neptune Analytics graph snapshot. A graph snapshot is a compacted, point-in-time copy of an entire Neptune Analytics graph that can be used as a backup.

## Example Usage

```terraform
resource "aws_neptunegraph_graph" "example" {
  graph_name          = "example-graph"
  provisioned_memory  = 16
  deletion_protection = false
}

resource "aws_neptunegraph_graph_snapshot" "example" {
  graph_identifier = aws_neptunegraph_graph.example.id
  snapshot_name    = "example-snapshot"

  tags = {
    Environment = "Development"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `graph_identifier` - (Required) Unique identifier of the Neptune Analytics graph to snapshot. Changing this forces a new resource to be created.
* `snapshot_name` - (Required) Name of the snapshot. Must contain from 1 to 63 lowercase letters, digits, or hyphens, begin with a letter, not end with a hyphen, not contain two consecutive hyphens, and not begin with the reserved prefix `gs-`. Changing this forces a new resource to be created.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the graph snapshot.
* `id` - Unique identifier of the graph snapshot (for example, `gs-0123456789`).
* `kms_key_identifier` - ID of the KMS key used to encrypt and decrypt the snapshot.
* `snapshot_create_time` - Time when the snapshot was created, in RFC3339 format.
* `source_graph_id` - Unique identifier of the Neptune Analytics graph from which the snapshot was created.
* `status` - Current status of the snapshot.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_neptunegraph_graph_snapshot` using the snapshot identifier. For example:

```terraform
import {
  to = aws_neptunegraph_graph_snapshot.example
  id = "gs-0123456789"
}
```

Using `terraform import`, import `aws_neptunegraph_graph_snapshot` using the snapshot identifier. For example:

```console
% terraform import aws_neptunegraph_graph_snapshot.example "gs-0123456789"
```
