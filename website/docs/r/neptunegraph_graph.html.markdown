---
subcategory: "Neptune Analytics"
layout: "aws"
page_title: "AWS: aws_neptunegraph_graph"
description: |-
  Provides an Amazon Neptune Analytics Graph
---

# Resource: aws_neptunegraph_graph

The `aws_neptunegraph_graph` resource creates an Amazon Analytics Graph.

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

The following arguments are required:

* `provisioned_memory` - (Required) Provisioned memory-optimized Neptune Capacity Units (m-NCUs) to use for the graph.

The following arguments are optional:

* `account` - (Optional) AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `deletion_protection` - (Optional) Value that indicates whether the Graph has deletion protection enabled. The graph can't be deleted when deletion protection is enabled. Default: `true`.
* `graph_name` - (Optional, Forces new resource) User-supplied name for the Graph. If omitted, Terraform will assign a random, unique identifier. Conflicts with `graph_name_prefix`.
* `graph_name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `graph_name`.
* `kms_key_identifier` - (Optional, Forces new resource) ARN for the KMS encryption key. By Default, Neptune Analytics will use an AWS provided key ("AWS_OWNED_KEY"). This parameter is used if you want to encrypt the graph using a KMS Customer Managed Key (CMK).
* `public_connectivity` - (Optional) Whether the Graph can be reached over the internet. Access to all graphs requires IAM authentication. When the Graph is publicly reachable, its DNS endpoint resolves to the public IP address from the internet. When the Graph isn't publicly reachable, you need to create a PrivateGraphEndpoint in a given VPC to ensure the DNS name resolves to a private IP address that is reachable from the VPC. Default: `false`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `replica_count` - (Optional, Forces new resource) Number of replicas you want when finished. All replicas will be provisioned in different availability zones. Replica Count should always be less than or equal to 2. Default: `1`.
* `tags` - (Optional) Key-value tags for the graph. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vector_search_configuration` - (Optional, Forces new resource) Vector search configuration for the graph. See [`vector_search_configuration` Block](#vector_search_configuration-block) for details.

### `vector_search_configuration` Block

The `vector_search_configuration` block supports the following:

* `vector_search_dimension` - (Optional, Forces new resource) Number of dimensions for vector embeddings. Value must be between 1 and 65,535.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Graph resource ARN.
* `endpoint` - Connection endpoint for the graph. For example: `g-12a3bcdef4.us-east-1.neptune-graph.amazonaws.com`
* `id` - Auto-generated ID assigned by the service.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `30m`)
- `update` - (Default `30m`)
- `delete` - (Default `30m`)

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
