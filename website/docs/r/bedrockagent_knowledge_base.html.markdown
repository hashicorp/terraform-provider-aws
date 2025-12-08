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

### Basic Usage

```terraform
resource "aws_bedrockagent_knowledge_base" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn
  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v2:0"
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

### Kendra Knowledge Base

```hcl
resource "aws_bedrockagent_knowledge_base" "kendra_example" {
  name     = "example-kendra-kb"
  role_arn = aws_iam_role.example.arn

  knowledge_base_configuration {
    type = "KENDRA"
    kendra_knowledge_base_configuration {
      kendra_index_arn = "arn:aws:kendra:us-east-1:123456789012:index/example-index-id"
    }
  }
}
```

### OpenSearch Managed Cluster Configuration

```terraform
resource "aws_bedrockagent_knowledge_base" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v2:0"
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "OPENSEARCH_MANAGED_CLUSTER"
    opensearch_managed_cluster_configuration {
      domain_arn        = "arn:aws:es:us-west-2:123456789012:domain/example-domain"
      domain_endpoint   = "https://search-example-domain.us-west-2.es.amazonaws.com"
      vector_index_name = "example_index"

      field_mapping {
        metadata_field = "metadata"
        text_field     = "chunks"
        vector_field   = "embedding"
      }
    }
  }
}
```

### With Supplemental Data Storage Configuration

```terraform
resource "aws_bedrockagent_knowledge_base" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn
  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v2:0"

      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 1024
          embedding_data_type = "FLOAT32"
        }
      }

      supplemental_data_storage_configuration {
        storage_location {
          type = "S3"

          s3_location {
            uri = "s3://my-bucket/chunk-processor/"
          }
        }
      }
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

### S3 Vectors Configuration

```terraform
resource "aws_s3vectors_vector_bucket" "example" {
  vector_bucket_name = "example-bucket"
}

resource "aws_s3vectors_index" "example" {
  index_name         = "example-index"
  vector_bucket_name = aws_s3vectors_vector_bucket.example.vector_bucket_name

  data_type       = "float32"
  dimension       = 256
  distance_metric = "euclidean"
}

resource "aws_bedrockagent_knowledge_base" "example" {
  name     = "example-s3vectors-kb"
  role_arn = aws_iam_role.example.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:aws:bedrock:us-west-2::foundation-model/amazon.titan-embed-text-v2:0"
      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 256
          embedding_data_type = "FLOAT32"
        }
      }
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "S3_VECTORS"
    s3_vectors_configuration {
      index_arn = aws_s3vectors_index.example.index_arn
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `knowledge_base_configuration` - (Required, Forces new resource) Details about the embeddings configuration of the knowledge base. See [`knowledge_base_configuration` block](#knowledge_base_configuration-block) for details.
* `name` - (Required) Name of the knowledge base.
* `role_arn` - (Required) ARN of the IAM role with permissions to invoke API operations on the knowledge base.

The following arguments are optional:

* `description` - (Optional) Description of the knowledge base.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `storage_configuration` - (Optional, Forces new resource) Details about the storage configuration of the knowledge base. See [`storage_configuration` block](#storage_configuration-block) for details.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `knowledge_base_configuration` block

The `knowledge_base_configuration` configuration block supports the following arguments:

* `type` - (Required) Type of data that the data source is converted into for the knowledge base. Valid Values: `VECTOR`, `KENDRA`, `SQL`.
* `kendra_knowledge_base_configuration` - (Optional) Configuration for Kendra knowledge base. See [`kendra_knowledge_base_configuration` block](#kendra_knowledge_base_configuration-block) for details.
* `vector_knowledge_base_configuration` - (Optional) Details about the embeddings model that'sused to convert the data source. See [`vector_knowledge_base_configuration` block](#vector_knowledge_base_configuration-block) for details.

### `kendra_knowledge_base_configuration` block

The `kendra_knowledge_base_configuration` configuration block supports the following arguments:

* `kendra_index_arn` - (Required) ARN of the Amazon Kendra index.

### `vector_knowledge_base_configuration` block

The `vector_knowledge_base_configuration` configuration block supports the following arguments:

* `embedding_model_arn` - (Required) ARN of the model used to create vector embeddings for the knowledge base.
* `embedding_model_configuration` - (Optional) The embeddings model configuration details for the vector model used in Knowledge Base.  See [`embedding_model_configuration` block](#embedding_model_configuration-block) for details.
* `supplemental_data_storage_configuration` - (Optional) supplemental_data_storage_configuration.  See [`supplemental_data_storage_configuration` block](#supplemental_data_storage_configuration-block) for details.

### `embedding_model_configuration` block

The `embedding_model_configuration` configuration block supports the following arguments:

* `bedrock_embedding_model_configuration` - (Optional) The vector configuration details on the Bedrock embeddings model.  See [`bedrock_embedding_model_configuration` block](#bedrock_embedding_model_configuration-block) for details.

### `bedrock_embedding_model_configuration` block

The `bedrock_embedding_model_configuration` configuration block supports the following arguments:

* `dimensions` - (Optional) Dimension details for the vector configuration used on the Bedrock embeddings model.
* `embedding_data_type` - (Optional) Data type for the vectors when using a model to convert text into vector embeddings. The model must support the specified data type for vector embeddings.  Valid values are `FLOAT32` and `BINARY`.

### `supplemental_data_storage_configuration` block

The `supplemental_data_storage_configuration` configuration block supports the following arguments:

* `storage_location` - (Required) A storage location specification for images extracted from multimodal documents in your data source.  See [`storage_location` block](#storage_location-block) for details.

### `storage_location` block

The `storage_location` configuration block supports the following arguments:

* `type` - (Required) Storage service used for this location. `S3` is the only valid value.
* `s3_location` - (Optional) Contains information about the Amazon S3 location for the extracted images.  See [`s3_location` block](#s3_location-block) for details.

### `s3_location` block

The `s3_location` configuration block supports the following arguments:

* `uri` - (Required) URI of the location.

### `storage_configuration` block

The `storage_configuration` configuration block supports the following arguments:

* `type` - (Required) Vector store service in which the knowledge base is stored. Valid Values: `MONGO_DB_ATLAS`, `OPENSEARCH_SERVERLESS`, `OPENSEARCH_MANAGED_CLUSTER`, `PINECONE`, `REDIS_ENTERPRISE_CLOUD`, `RDS`, `S3_VECTORS`.
* `mongo_db_atlas_configuration` – (Optional) The storage configuration of the knowledge base in MongoDB Atlas. See [`opensearch_managed_cluster_comongo_db_atlas_configurationnfiguration` block](#mongo_db_atlas_configuration-block) for details.
* `opensearch_managed_cluster_configuration` - (Optional) The storage configuration of the knowledge base in Amazon OpenSearch Service Managed Cluster. See [`opensearch_managed_cluster_configuration` block](#opensearch_managed_cluster_configuration-block) for details.
* `opensearch_serverless_configuration` - (Optional) The storage configuration of the knowledge base in Amazon OpenSearch Service Serverless. See [`opensearch_serverless_configuration` block](#opensearch_serverless_configuration-block) for details.
* `pinecone_configuration` - (Optional)  The storage configuration of the knowledge base in Pinecone. See [`pinecone_configuration` block](#pinecone_configuration-block) for details.
* `rds_configuration` - (Optional) Details about the storage configuration of the knowledge base in Amazon RDS. For more information, see [Create a vector index in Amazon RDS](https://docs.aws.amazon.com/bedrock/latest/userguide/knowledge-base-setup.html). See [`rds_configuration` block](#rds_configuration-block) for details.
* `redis_enterprise_cloud_configuration` - (Optional) The storage configuration of the knowledge base in Redis Enterprise Cloud. See [`redis_enterprise_cloud_configuration` block](#redis_enterprise_cloud_configuration-block) for details.
* `s3_vectors_configuration` - (Optional) The storage configuration of the knowledge base in Amazon S3 Vectors. See [`s3_vectors_configuration` block](#s3_vectors_configuration-block) for details.

### `mongo_db_atlas_configuration` block

The `mongo_db_atlas_configuration` configuration block supports the following arguments:

* `collection_name` – (Required) The name of the collection in the MongoDB Atlas database.
* `credentials_secret_arn` – (Required) The ARN of the secret that you created in AWS Secrets Manager that is linked to your MongoDB Atlas database.
* `database_name` – (Required) The name of the database in the MongoDB Atlas database.
* `endpoint` – (Required) The endpoint URL of the MongoDB Atlas database.
* `field_mapping` – (Required) Contains the names of the fields to which to map information about the vector store.
    * `metadata_field` – (Required) The name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` – (Required) The name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` – (Required) The name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `vector_index_name` – (Required) The name of the vector index.
* `endpoint_service_name` – (Optional) The name of the service that hosts the MongoDB Atlas database.
* `text_index_name` – (Optional) The name of the vector index.

### `opensearch_managed_cluster_configuration` block

The `opensearch_managed_cluster_configuration` configuration block supports the following arguments:

* `domain_arn` - (Required) ARN of the OpenSearch domain.
* `domain_endpoint` - (Required) Endpoint URL of the OpenSearch domain.
* `field_mapping` - (Required) The names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `metadata_field` - (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` - (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` - (Required) Name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `vector_index_name` - (Required) Name of the vector store.

### `opensearch_serverless_configuration` block

The `opensearch_serverless_configuration` configuration block supports the following arguments:

* `collection_arn` - (Required) ARN of the OpenSearch Service vector store.
* `field_mapping` - (Required) The names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `metadata_field` - (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` - (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` - (Required) Name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `vector_index_name` - (Required) Name of the vector store.

### `pinecone_configuration` block

The `pinecone_configuration` configuration block supports the following arguments:

* `connection_string` - (Required) Endpoint URL for your index management page.
* `credentials_secret_arn` - (Required) ARN of the secret that you created in AWS Secrets Manager that is linked to your Pinecone API key.
* `field_mapping` - (Required) The names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `metadata_field` - (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` - (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
* `namespace` - (Optional) Namespace to be used to write new data to your database.

### `rds_configuration` block

 The `rds_configuration` configuration block supports the following arguments:

* `credentials_secret_arn` - (Required) ARN of the secret that you created in AWS Secrets Manager that is linked to your Amazon RDS database.
* `database_name` - (Required) Name of your Amazon RDS database.
* `field_mapping` - (Required) Names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `custom_metadata_field` - (Optional) Name for the universal metadata field where Amazon Bedrock will store any custom metadata from your data source.
    * `metadata_field` - (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `primary_key_field` - (Required) Name of the field in which Amazon Bedrock stores the ID for each entry.
    * `text_field` - (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` - (Required) Name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `resource_arn` - (Required) ARN of the vector store.
* `table_name` - (Required) Name of the table in the database.

### `redis_enterprise_cloud_configuration` block

The `redis_enterprise_cloud_configuration` configuration block supports the following arguments:

* `credentials_secret_arn` - (Required) ARN of the secret that you created in AWS Secrets Manager that is linked to your Redis Enterprise Cloud database.
* `endpoint` - (Required) Endpoint URL of the Redis Enterprise Cloud database.
* `field_mapping` - (Required) The names of the fields to which to map information about the vector store. This block supports the following arguments:
    * `metadata_field` - (Required) Name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` - (Required) Name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` - (Required) Name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `vector_index_name` - (Required) Name of the vector index.

### `s3_vectors_configuration` block

The `s3_vectors_configuration` configuration block supports the following arguments.
Either `index_arn`, or both `index_name` and `vector_bucket_arn` must be specified.

* `index_arn` - (Optional) ARN of the S3 Vectors index. Conflicts with `index_name` and `vector_bucket_arn`.
* `index_name` - (Optional) Name of the S3 Vectors index. Must be specified with `vector_bucket_arn`. Conflicts with `index_arn`.
* `vector_bucket_arn` - (Optional) ARN of the S3 Vectors vector bucket. Must be specified with `index_name`. Conflicts with `index_arn`.

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
