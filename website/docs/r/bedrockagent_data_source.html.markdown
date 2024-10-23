---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_data_source"
description: |-
  Terraform resource for managing an AWS Agents for Amazon Bedrock Data Source.
---

# Resource: aws_bedrockagent_data_source

Terraform resource for managing an AWS Agents for Amazon Bedrock Data Source.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagent_data_source" "example" {
  knowledge_base_id = "EMDPPAYPZI"
  name              = "example"
  data_source_configuration {
    type = "S3"
    s3_configuration {
      bucket_arn = "arn:aws:s3:::example-bucket"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `data_source_configuration` - (Required) Details about how the data source is stored. See [`data_source_configuration` block](#data_source_configuration-block) for details.
* `knowledge_base_id` - (Required) Unique identifier of the knowledge base to which the data source belongs.
* `name` - (Required, Forces new resource) Name of the data source.

The following arguments are optional:

* `data_deletion_policy` - (Optional) Data deletion policy for a data source. Valid values: `RETAIN`, `DELETE`.
* `description` - (Optional) Description of the data source.
* `server_side_encryption_configuration` - (Optional) Details about the configuration of the server-side encryption. See [`server_side_encryption_configuration` block](#server_side_encryption_configuration-block) for details.
* `vector_ingestion_configuration` - (Optional, Forces new resource) Details about the configuration of the server-side encryption. See [`vector_ingestion_configuration` block](#vector_ingestion_configuration-block) for details.

### `data_source_configuration` block

The `data_source_configuration` configuration block supports the following arguments:

* `type` - (Required) Type of storage for the data source. Valid values: `S3`.
* `s3_configuration` - (Optional) Details about the configuration of the S3 object containing the data source. See [`s3_data_source_configuration` block](#s3_data_source_configuration-block) for details.

### `s3_data_source_configuration` block

The `s3_data_source_configuration` configuration block supports the following arguments:

* `bucket_arn` - (Required) ARN of the bucket that contains the data source.
* `bucket_owner_account_id` - (Optional) Bucket account owner ID for the S3 bucket.
* `inclusion_prefixes` - (Optional) List of S3 prefixes that define the object containing the data sources. For more information, see [Organizing objects using prefixes](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-prefixes.html).

### `server_side_encryption_configuration` block

The `server_side_encryption_configuration` configuration block supports the following arguments:

* `kms_key_arn` - (Optional) ARN of the AWS KMS key used to encrypt the resource.

### `vector_ingestion_configuration` block

The `vector_ingestion_configuration` configuration block supports the following arguments:

* `chunking_configuration` - (Optional, Forces new resource) Details about how to chunk the documents in the data source. A chunk refers to an excerpt from a data source that is returned when the knowledge base that it belongs to is queried. See [`chunking_configuration` block](#chunking_configuration-block) for details.
* `custom_transformation_configuration`- (Optional, Forces new resource) Configuration for custom transformation of data source documents.
* `parsing_configuration` - (Optional, Forces new resource) Configuration for custom parsing of data source documents. See [`parsing_configuration` block](#parsing_configuration-block) for details.

### `chunking_configuration` block

 The `chunking_configuration` configuration block supports the following arguments:

* `chunking_strategy` - (Required, Forces new resource) Option for chunking your source data, either in fixed-sized chunks or as one chunk. Valid values: `FIXED_SIZE`, `HIERARCHICAL`, `SEMANTIC`, `NONE`.
* `fixed_size_chunking_configuration` - (Optional, Forces new resource) Configurations for when you choose fixed-size chunking. Requires chunking_strategy as `FIXED_SIZE`. See [`fixed_size_chunking_configuration`](#fixed_size_chunking_configuration-block) for details.
* `hierarchical_chunking_configuration` - (Optional, Forces new resource) Configurations for when you choose hierarchical chunking. Requires chunking_strategy as `HIERARCHICAL`. See [`hierarchical_chunking_configuration`](#hierarchical_chunking_configuration-block) for details.
* `semantic_chunking_configuration` - (Optional, Forces new resource) Configurations for when you choose semantic chunking. Requires chunking_strategy as `SEMANTIC`. See [`semantic_chunking_configuration`](#semantic_chunking_configuration-block) for details.

### `fixed_size_chunking_configuration` block

The `fixed_size_chunking_configuration` block supports the following arguments:

* `max_tokens` - (Required, Forces new resource) Maximum number of tokens to include in a chunk.
* `overlap_percentage` - (Optional, Forces new resource) Percentage of overlap between adjacent chunks of a data source.

### `hierarchical_chunking_configuration` block

The `hierarchical_chunking_configuration` block supports the following arguments:

* `level_configuration` - (Required, Forces new resource) Maximum number of tokens to include in a chunk. Must contain two `level_configurations`. See [`level_configurations`](#level_configuration-block) for details.
* `overlap_tokens` - (Required, Forces new resource) The number of tokens to repeat across chunks in the same layer.

### `level_configuration` block

The `level_configuration` block supports the following arguments:

* `max_tokens` - (Required) The maximum number of tokens that a chunk can contain in this layer.

### `semantic_chunking_configuration` block

The `semantic_chunking_configuration` block supports the following arguments:

* `breakpoint_percentile_threshold` - (Required, Forces new resource) The dissimilarity threshold for splitting chunks.
* `buffer_size` - (Required, Forces new resource) The buffer size.
* `max_tokens` - (Required, Forces new resource) The maximum number of tokens a chunk can contain.

### `custom_transformation_configuration` block

The `custom_transformation_configuration` block supports the following arguments:

* `intermediate_storage` - (Required, Forces new resource) The intermediate storage for custom transformation.
* `transformation_function` - (Required) The configuration of transformation function.

### `intermediate_storage` block

The `intermediate_storage` block supports the following arguments:

* `s3_location` - (Required, Forces new resource) Configuration block for intermedia S3 storage.

### `s3_location` block

The `s3_location` block supports the following arguments:

* `uri` - (Required, Forces new resource) S3 URI for intermediate storage.

### `transformation_function` block

The `transformation_function` block supports the following arguments:

* `step_to_apply` - (Required, Forces new resource) Currently only `POST_CHUNKING` is supported.
* `transformation_lambda_configuration` - (Required, Forces new resource) The lambda configuration for custom transformation.

### `transformation_lambda_configuration` block

The `transformation_lambda_configuration` block supports the following arguments:

* `lambda_arn` - (Required, Forces new resource) The ARN of the lambda to use for custom transformation.

### `parsing_configuration` block

The `parsing_configuration` configuration block supports the following arguments:

* `parsing_strategy` - (Required) Currently only `BEDROCK_FOUNDATION_MODEL` is supported
* `bedrock_foundation_model_configuration` - (Optional) Settings for a foundation model used to parse documents in a data source. See [`bedrock_foundation_model_configuration` block](#bedrock_foundation_model_configuration-block) for details.

### `bedrock_foundation_model_configuration` block

The `bedrock_foundation_model_configuration` configuration block supports the following arguments:

* `model_arn` - (Required) The ARN of the model used to parse documents
* `parsing_prompt` - (Optional) Instructions for interpreting the contents of the document. See [`parsing_prompt` block](#parsing_prompt-block) for details.

### `parsing_prompt` block

The `parsing_prompt` configuration block supports the following arguments:

* `parsing_prompt_string` - (Required) Instructions for interpreting the contents of the document.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `data_source_id` -  Unique identifier of the data source.
* `id` -  Identifier of the data source which consists of the data source ID and the knowledge base ID.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Data Source using the data source ID and the knowledge base ID. For example:

```terraform
import {
  to = aws_bedrockagent_data_source.example
  id = "GWCMFMQF6T,EMDPPAYPZI"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Data Source using the data source ID and the knowledge base ID. For example:

```console
% terraform import aws_bedrockagent_data_source.example GWCMFMQF6T,EMDPPAYPZI
```
