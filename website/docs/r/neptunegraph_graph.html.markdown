---
subcategory: "Neptune Analytics"
layout: "aws"
page_title: "AWS: aws_neptunegraph_graph"
description: |-
  Provides an Amazon Neptune Analytics Graph
---

# Resource: aws_neptunegraph_graph

The AWS::NeptuneGraph::Graph resource creates an Amazon Analytics Graph.

## Example Usage

### Neptune Graph (with Vector Search configuration)

Creates a Neptune Graph with 16GB provisioned memory, vector search capability with 128 dimensions, and a single replica for high availability.

```terraform
# Create Neptune Graph
resource "aws_neptunegraph_graph" "example" {
  graph_name          = "example-graph-test-20250203"
  provisioned_memory  = 16
  deletion_protection = false
  public_connectivity = false
  replica_count       = 1
  kms_key_identifier  = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"

  vector_search_configuration {
    vector_search_dimension = 128
  }

  tags = {
    "Environment" = "Development"
    "ModifiedBy"  = "AWS"
  }
}
```

## Argument Reference

### Required

- `provisioned_memory` (Number, Forces new reource) The provisioned memory-optimized Neptune Capacity Units (m-NCUs) to use for the graph.

### Optional

- `deletion_protection` (Boolean, Default: `true`) Value that indicates whether the Graph has deletion protection enabled. The graph can't be deleted when deletion protection is enabled.

- `graph_name` (String, Forces new resource) Contains a user-supplied name for the Graph. If omitted, Terraform will assign a random, unique identifier.

- `public_connectivity` (Boolean, Default: `false`) Specifies whether the Graph can be reached over the internet. Access to all graphs requires IAM authentication.  When the Graph is publicly reachable, its Domain Name System (DNS) endpoint resolves to the public IP address from the internet.  When the Graph isn't publicly reachable, you need to create a PrivateGraphEndpoint in a given VPC to ensure the DNS name resolves to a private IP address that is reachable from the VPC.

- `replica_count` (Number, Default: `1`, Forces new resource) Specifies the number of replicas you want when finished. All replicas will be provisioned in different availability zones.  Replica Count should always be less than or equal to 2.

- `kms_key_identifier` (String) The ARN for the KMS encryption key. By Default, Neptune Analytics will use an AWS provided key ("AWS_OWNED_KEY"). This parameter is used if you want to encrypt the graph using a KMS Customer Managed Key (CMK).

- `vector_search_configuration` (Block, Forces new resource) Vector Search Configuration (see below for nested schema of vector_search_configuration)

- `tags` (Attributes Set) The tags associated with this graph. (see below for nested schema of tags)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `endpoint` (String) The connection endpoint for the graph. For example: `g-12a3bcdef4.us-east-1.neptune-graph.amazonaws.com`
- `arn` (String) Graph resource ARN
- `id` (String) The auto-generated id assigned by the service.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `30m`)
- `delete` - (Default `30m`)

### Nested Schema for `tags`

Optional:

- `key` (String) The key name of the tag. You can specify a value that is 1 to 128 Unicode characters in length and cannot be prefixed with aws:. You can use any of the following characters: the set of Unicode letters, digits, whitespace, _, ., /, =, +, and -.
- `value` (String) The value for the tag. You can specify a value that is 0 to 256 Unicode characters in length and cannot be prefixed with aws:. You can use any of the following characters: the set of Unicode letters, digits, whitespace, _, ., /, =, +, and -.

### Nested Schema for `vector_search_configuration`

Optional:

- `vector_search_dimension` (Number, Forces new resource) Specifies the number of dimensions for vector embeddings.  Value must be between 1 and 65,535.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_neptunegraph_graph` using the graph identifier. For example:

```terraform
import {
  to = aws_neptunegraph_graph.example
  id = "graph_id"
}
```

Using `terraform import`, import `aws_neptunegraph_graph` using the graph identifier. For example:

```console
% terraform import aws_neptunegraph_graph.example "graph_id"
```
