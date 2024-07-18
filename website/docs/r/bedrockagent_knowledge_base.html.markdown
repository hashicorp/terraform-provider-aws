---
subcategory: "Bedrock Agents"
layout: "aws"
page_title: "AWS: aws_bedrockagent_knowledge_base"
description: |-
  Terraform resource for managing an AWS Agents for Amazon Bedrock Knowledge Base.
---

# Resource: aws_bedrockagent_knowledge_base

Terraform resource for managing an AWS Agents for Amazon Bedrock Knowledge Base.

## Example Usage

```terraform
resource "aws_bedrockagent_knowledge_base" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn
  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v1"
    }
    type = "VECTOR"
  }
  storage_configuration {
    type = "OPENSEARCH_SERVERLESS"
    opensearch_serverless_configuration {
      collection_arn    = "arn:aws:aoss:us-west-2:123456789012:collection/142bezjddq707i5stcrf"
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
        vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `knowledge_base_configuration` - (Required) Details about the embeddings configuration of the knowledge base. See [`knowledge_base_configuration` block](#knowledge_base_configuration-block) for details.
* `name` - (Required, Forces new resource) Name of the knowledge base.
* `role_arn` - (Required) ARN of the IAM role with permissions to invoke API operations on the knowledge base.
* `storage_configuration` - (Required) Details about the storage configuration of the knowledge base. See [`storage_configuration` block](#storage_configuration-block) for details.

The following arguments are optional:

* `description` - (Optional) Description of the knowledge base.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `knowledge_base_configuration` block

The `knowledge_base_configuration` configuration block supports the following arguments:

* `type` – (Required) Type of data that the data source is converted into for the knowledge base. Valid Values: `VECTOR`.
* `vector_knowledge_base_configuration` – (Optional) Details about the embeddings model that'sused to convert the data source. See [`vector_knowledge_base_configuration` block](#vector_knowledge_base_configuration-block) for details.

### `vector_knowledge_base_configuration` block

The `vector_knowledge_base_configuration` configuration block supports the following arguments:

* `embedding_model_arn` – (Required) ARN of the model used to create vector embeddings for the knowledge base.

### `storage_configuration` block

The `storage_configuration` configuration block supports the following arguments:

* `type` – (Required) Vector store service in which the knowledge base is stored. Valid Values: `OPENSEARCH_SERVERLESS`, `PINECONE`, `REDIS_ENTERPRISE_CLOUD`, `RDS`.
* `opensearch_serverless_configuration` – (Optional) The storage configuration of the knowledge base in Amazon OpenSearch Service. See [`opensearch_serverless_configuration` block](#opensearch_serverless_configuration-block) for details.
* `pinecone_configuration` – (Optional)  The storage configuration of the knowledge base in Pinecone. See [`pinecone_configuration` block](#pinecone_configuration-block) for details.
* `rds_configuration` – (Optional) Details about the storage configuration of the knowledge base in Amazon RDS. For more information, see [Create a vector index in Amazon RDS](https://docs.aws.amazon.com/bedrock/latest/userguide/knowledge-base-setup.html). See [`rds_configuration` block](#rds_configuration-block) for details.
* `redis_enterprise_cloud_configuration` – (Optional) The storage configuration of the knowledge base in Redis Enterprise Cloud. See [`redis_enterprise_cloud_configuration` block](#redis_enterprise_cloud_configuration-block) for details.

### `opensearch_serverless_configuration` block

The `opensearch_serverless_configuration` configuration block supports the following arguments:

* `collection_arn` – (Required) ARN of the OpenSearch Service vector store.
* `field_mapping` – (Required) The names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `metadata_field` – (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` – (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` – (Required) Name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `vector_index_name` – (Required) Name of the vector store.

### `pinecone_configuration` block

The `pinecone_configuration` configuration block supports the following arguments:

* `connection_string` – (Required) Endpoint URL for your index management page.
* `credentials_secret_arn` – (Required) ARN of the secret that you created in AWS Secrets Manager that is linked to your Pinecone API key.
* `field_mapping` – (Required) The names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `metadata_field` – (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` – (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
* `namespace` – (Optional) Namespace to be used to write new data to your database.

### `rds_configuration` block

 The `rds_configuration` configuration block supports the following arguments:

* `credentials_secret_arn` – (Required) ARN of the secret that you created in AWS Secrets Manager that is linked to your Amazon RDS database.
* `database_name` – (Required) Name of your Amazon RDS database.
* `field_mapping` – (Required) Names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `metadata_field` – (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `primary_key_field` – (Required) Name of the field in which Amazon Bedrock stores the ID for each entry.
    * `text_field` – (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` – (Required) Name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `resource_arn` – (Required) ARN of the vector store.
* `table_name` – (Required) Name of the table in the database.

### `redis_enterprise_cloud_configuration` block

The `redis_enterprise_cloud_configuration` configuration block supports the following arguments:

* `credentials_secret_arn` – (Required) ARN of the secret that you created in AWS Secrets Manager that is linked to your Redis Enterprise Cloud database.
* `endpoint` – (Required) Endpoint URL of the Redis Enterprise Cloud database.
* `field_mapping` – (Required) The names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `metadata_field` – (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` – (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` – (Required) Name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `vector_index_name` – (Required) Name of the vector index.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the knowledge base.
* `created_at` - Time at which the knowledge base was created.
* `id` - Unique identifier of the knowledge base.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Time at which the knowledge base was last updated.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Knowledge Base using the knowledge base ID. For example:

```terraform
import {
  to = aws_bedrockagent_knowledge_base.example
  id = "EMDPPAYPZI"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Knowledge Base using the knowledge base ID. For example:

```console
% terraform import aws_bedrockagent_knowledge_base.example EMDPPAYPZI
```
