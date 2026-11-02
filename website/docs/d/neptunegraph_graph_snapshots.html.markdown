---
subcategory: "Neptune Analytics"
layout: "aws"
page_title: "AWS: aws_neptunegraph_graph_snapshots"
description: |-
  Terraform data source for listing Amazon Neptune Analytics graph snapshots.
---

# Data Source: aws_neptunegraph_graph_snapshots

Terraform data source for listing Amazon Neptune Analytics graph snapshots.

## Example Usage

### List all snapshots

```terraform
data "aws_neptunegraph_graph_snapshots" "example" {}
```

### List snapshots for a specific graph

```terraform
data "aws_neptunegraph_graph_snapshots" "example" {
  graph_identifier = aws_neptunegraph_graph.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `graph_identifier` - (Optional) Unique identifier of a Neptune Analytics graph. When set, only snapshots for that graph are returned.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `graph_snapshots` - List of graph snapshots. Each element is an object with the following attributes:
    * `arn` - ARN of the graph snapshot.
    * `id` - Unique identifier of the graph snapshot.
    * `kms_key_identifier` - ID of the KMS key used to encrypt and decrypt the snapshot.
    * `name` - Name of the snapshot.
    * `snapshot_create_time` - Time when the snapshot was created, in RFC3339 format.
    * `source_graph_id` - Unique identifier of the Neptune Analytics graph from which the snapshot was created.
    * `status` - Current status of the snapshot.
