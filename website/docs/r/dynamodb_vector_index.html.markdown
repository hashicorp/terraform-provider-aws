---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_vector_index"
description: |-
  Provides a DynamoDB Vector Index resource
---

# Resource: aws_dynamodb_vector_index

Manages a vector index on a DynamoDB table, enabling similarity search over vector embeddings stored in table items via the `SearchVectors` API.

A vector index cannot be modified in place; any configuration change forces replacement of the index. Deleting a vector index does not delete the underlying table or its items.

## Example Usage

```terraform
resource "aws_dynamodb_vector_index" "example" {
  table_name = aws_dynamodb_table.example.name
  index_name = "DescriptionEmbeddingIndex"

  dimensions        = 1536
  distance_function = "COSINE"

  vector_attribute = {
    attribute_name = "description_embedding"
  }

  projection {
    projection_type = "ALL"
  }

  search_schema {
    attribute_name = "TenantId"
    attribute_type = "S"
    type           = "HASH"
  }
}

resource "aws_dynamodb_table" "example" {
  name         = "example"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TenantId"

  attribute {
    name = "TenantId"
    type = "S"
  }
}
```

## Argument Reference

The following arguments are required:

* `dimensions` - (Required) Number of dimensions in each vector. Must match the size of the vector embeddings written to `vector_attribute.attribute_name`. Changing this value will re-create the resource.
* `distance_function` - (Required) Distance function used to calculate similarity between vectors. Valid values are `COSINE`, `DOT_PRODUCT`, or `EUCLIDEAN`. Changing this value will re-create the resource.
* `index_name` - (Required) Name of the vector index. Must be unique within the table. Changing this value will re-create the resource.
* `projection` - (Required) Describes which attributes from the table are copied into the vector index. See [`projection` below](#projection).
* `table_name` - (Required) Name of the table this vector index belongs to. Changing this value will re-create the resource.
* `vector_attribute` - (Required) Attribute on the table that holds the vector embedding. See [`vector_attribute` below](#vector_attribute).

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `search_schema` - (Optional) Partition key and inline filter attributes for the vector index. Changing this value will re-create the resource. See [`search_schema` below](#search_schema).

### `projection`

* `non_key_attributes` - (Optional) Specifies which additional attributes to include in the index. Only valid when `projection_type` is `INCLUDE`.
* `projection_type` - (Required) Set of attributes represented in the index. One of `ALL`, `INCLUDE`, or `KEYS_ONLY`.

### `search_schema`

* `attribute_name` - (Required) Name of the attribute.
* `attribute_type` - (Required) Type of the attribute. Valid values are `S` (string), `N` (number), or `B` (binary).
* `type` - (Required) Role of the attribute in the search schema.
  Valid values are `HASH` (a partition key that partitions the vector index for independent scaling; its value must be supplied in `SearchConditionExpression` when searching) or `INLINE_FILTER` (an attribute projected into the vector index for filtering at the storage layer during search; optional in `SearchConditionExpression`).

### `vector_attribute`

* `attribute_name` - (Required) Name of the table attribute that contains the vector embedding, stored as a `List` of `Number`s.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the vector index.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_dynamodb_vector_index.example
  identity = {
    "table_name" = "example-table"
    "index_name" = "DescriptionEmbeddingIndex"
  }
}

resource "aws_dynamodb_vector_index" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `index_name` (String) Name of the vector index.
* `table_name` (String) Name of the table this vector index belongs to.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB vector indexes using the `table_name` and `index_name`, separated by a comma. For example:

```terraform
import {
  to = aws_dynamodb_vector_index.example
  id = "example-table,DescriptionEmbeddingIndex"
}
```

Using `terraform import`, import DynamoDB vector indexes using the `table_name` and `index_name`, separated by a comma. For example:

```console
% terraform import aws_dynamodb_vector_index.example 'example-table,DescriptionEmbeddingIndex'
```
