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

### Create Graph from Data in S3

Creates a Neptune Graph using source data hosted in an S3 bucket.  For more details: https://docs.aws.amazon.com/neptune-analytics/latest/userguide/bulk-import-create-from-s3.html

```terraform
resource "aws_neptunegraph_graph" "example" {
  graph_name          = "example-graph-test-20250203"
  deletion_protection = false
  public_connectivity = false
  replica_count       = 1

  import_task {
    source                 = "s3://example-bucket/prefix"
    role_arn               = "arn:aws:iam::123456789012:role/NeptuneGraphImportRole"
    format                 = "CSV"
    fail_on_error          = false
    max_provisioned_memory = 32
    min_provisioned_memory = 16
  }

  tags = {
    "Environment" = "Development"
    "ModifiedBy"  = "AWS"
  }
}
```

### Create Graph from a Neptune Database cluster

Creates a Neptune Graph using data from a Neptune Database cluster as the source.  For more details:  https://docs.aws.amazon.com/neptune-analytics/latest/userguide/bulk-import-create-from-neptune.html

```terraform
resource "aws_neptunegraph_graph" "example" {
  graph_name          = "example-graph-test-20250203"
  deletion_protection = false
  public_connectivity = false
  replica_count       = 0

  import_task {
    source   = "arn:aws:rds:us-east-1:123456789012:cluster:neptune-db-cluster-name"
    role_arn = "arn:aws:iam::123456789012:role/NeptuneGraphImportRole"

    import_options {
      neptune {
        s3_export_path                 = "s3://example-export-bucket/export/"
        s3_export_kms_key_id           = "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
        preserve_default_vertex_labels = true
        preserve_edge_ids              = true
      }
    }
  }

  tags = {
    "Environment" = "Development"
    "ModifiedBy"  = "AWS"
  }
}
```

## Argument Reference

The following arguments are required:

- `provisioned_memory` (Number, Forces new reource) The provisioned memory-optimized Neptune Capacity Units (m-NCUs) to use for the graph.

The following arguments are optional:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `deletion_protection` (Boolean, Default: `true`) Value that indicates whether the Graph has deletion protection enabled. The graph can't be deleted when deletion protection is enabled.
- `graph_name` (String, Forces new resource) Contains a user-supplied name for the Graph. If omitted, Terraform will assign a random, unique identifier.
- `public_connectivity` (Boolean, Default: `false`) Specifies whether the Graph can be reached over the internet. Access to all graphs requires IAM authentication.  When the Graph is publicly reachable, its Domain Name System (DNS) endpoint resolves to the public IP address from the internet.  When the Graph isn't publicly reachable, you need to create a PrivateGraphEndpoint in a given VPC to ensure the DNS name resolves to a private IP address that is reachable from the VPC.
- `replica_count` (Number, Default: `1`, Forces new resource) Specifies the number of replicas you want when finished. All replicas will be provisioned in different availability zones.  Replica Count should always be less than or equal to 2.
- `kms_key_identifier` (String) The ARN for the KMS encryption key. By Default, Neptune Analytics will use an AWS provided key ("AWS_OWNED_KEY"). This parameter is used if you want to encrypt the graph using a KMS Customer Managed Key (CMK).
- `vector_search_configuration` (Block, Forces new resource) Vector Search Configuration (see below for nested schema of vector_search_configuration)
- `import_task` (Block, Forces new source) Creates a new Neptune Analytics graph and imports data into it, either from Amazon Simple Storage Service (S3) or from a Neptune database or a Neptune database snapshot.
- `tags` - (Optional) Key-value tags for the graph. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `endpoint` (String) The connection endpoint for the graph. For example: `g-12a3bcdef4.us-east-1.neptune-graph.amazonaws.com`
- `arn` (String) Graph resource ARN
- `id` (String) The auto-generated id assigned by the service.
- `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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

### Nested Schema for `import_task`

Optional:

- `blankNodeHandling` (String) The method to handle blank nodes in the dataset. Currently, only `convertToIri` is supported, meaning blank nodes are converted to unique IRIs at load time. Must be provided when format is ntriples. For more information, see [Handling RDF values](https://docs.aws.amazon.com/neptune-analytics/latest/userguide/using-rdf-data.html#rdf-handling).
- `parquetType` (String) The parquet type of the import task.  Currently, only `COLUMNAR` is supported.
- `source` (String) A URL identifying to the location of the data to be imported. This can be an Amazon S3 URL, Neptune Database cluster resource ARN or Neptune Database cluster snapshot ARN.
- `role_arn` (String) The ARN of the IAM role that will allow access to the data that is to be imported.
- `format` (String) Specifies the format of S3 data to be imported. Valid values are CSV, which identifies the [Gremlin CSV format](https://docs.aws.amazon.com/neptune/latest/userguide/bulk-load-tutorial-format-gremlin.html), OPEN_CYPHER, which identifies the [openCypher load format](https://docs.aws.amazon.com/neptune/latest/userguide/bulk-load-tutorial-format-opencypher.html), NTRIPLES, which identifies the [RDF n-triples](https://docs.aws.amazon.com/neptune-analytics/latest/userguide/using-rdf-data.html) format, or PARQUET.
- `fail_on_error` (Boolean) If set to true, the task halts when an import error is encountered. If set to false, the task skips the data that caused the error and continues if possible.
- `min_provisioned_memory` (Integer, Default: 16) The minimum provisioned memory-optimized Neptune Capacity Units (m-NCUs) to use for the graph.  (Minimum value of 16. Maximum value of 24576.)
- `max_provisioned_memory` (Integer, Default: 1024, or the approved upper limit for your account.) The maximum provisioned memory-optimized Neptune Capacity Units (m-NCUs) to use for the graph.   If both the minimum and maximum values are specified, the final provisioned-memory will be chosen per the actual size of your imported data. If neither value is specified, 128 m-NCUs are used.  (Minimum value of 16. Maximum value of 24576.)
- `import_options` (Block) Contains options for controlling the import process. For example, if the failOnError key is set to false, the import skips problem data and attempts to continue (whereas if set to true, the default, or if omitted, the import operation halts immediately when an error is encountered.

#### Nested Schema for `import_options`.`neptune`

- `s3ExportKmsKeyId` (String) The ARN of the KMS key to use to encrypt data in the S3 bucket where the graph data is exported.
- `s3ExportPath` (String) The path (URL) to an S3 bucket from which to import data.
- `preserveDefaultVertexLabels` (Boolean) Neptune Analytics supports label-less vertices and no labels are assigned unless one is explicitly provided. Neptune assigns default labels when none is explicitly provided. When importing the data into Neptune Analytics, the default vertex labels can be omitted by setting preserveDefaultVertexLabels to false. Note that if the vertex only has default labels, and has no other properties or edges, then the vertex will effectively not get imported into Neptune Analytics when preserveDefaultVertexLabels is set to false.
- `preserveEdgeIds` (Boolean) Neptune Analytics currently does not support user defined edge ids. The edge ids are not imported by default. They are imported if preserveEdgeIds is set to true, and ids are stored as properties on the relationships with the property name neptuneEdgeId.

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
