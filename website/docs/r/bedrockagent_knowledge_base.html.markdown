---
subcategory: "Agents for Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrockagent_knowledge_base"
description: |-
  Terraform resource for managing an AWS Agents for Amazon Bedrock Knowledge Base.
---

# Resource: aws_bedrockagent_knowledge_base

Terraform resource for managing an AWS Agents for Amazon Bedrock Knowledge Base.

## Example Usage

```terraform
resource "aws_bedrockagent_knowledge_base" "test" {
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
      collection_arn    = "arn:aws:aoss:us-west-2:1234567890:collection/142bezjddq707i5stcrf"
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

This resource supports the following arguments:

* `description` - (Optional) A description of the knowledge base.
* `name` - (Required) A name for the knowledge base.
* `role_arn` - (Required) The ARN of the IAM role with permissions to create the knowledge base.
* `knowledge_base_configuration` - (Required) Contains details about the embeddings model used for the knowledge base.
* `storage_configuration` - (Required) Contains details about the configuration of the vector database used for the knowledge base.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

Knowledge Base Configuration supports the following:

* `type` – (Required) The type of data that the data source is converted into for the knowledge base.
* `vector_knowledge_base_configuration` – (Optional) Contains details about the embeddings model that'sused to   convert the data source.

Vector Knowledge Base Configuration supports the following:

* `embedding_model_arn` – (Required) The ARN of the model used to create vector embeddings for the knowledge base.

Storage Configuration supports the following:

* `type` – (Required) The vector store service in which the knowledge base is stored.Valid Values: OPENSEARCH_SERVERLESS | PINECONE | REDIS_ENTERPRISE_CLOUD | RDS | MONGO_DB_ATLAS
* `pinecone_configuration` – (Optional) Contains the storage configuration of the knowledge base in Pinecone.
* `rds_configuration` – (Optional) Contains details about the storage configuration of the knowledge base in Amazon RDS. For more information, see Create a vector index in Amazon RDS.
* `redis_enterprise_cloud_configuration` – (Optional) Contains the storage configuration of the knowledge base in Redis Enterprise Cloud.
* `opensearch_serverless_configuration` – (Optional) Contains the storage configuration of the knowledge base in Amazon OpenSearch Service.
* `mongo_db_atlas_configuration` – (Optional) Contains the storage configuration of the knowledge base in MongoDB Atlas.

Pinecone Configuration supports the following:

* `connection_string` – (Required) The endpoint URL for your index management page.
* `credentials_secret_arn` – (Required) The ARN of the secret that you created in AWS Secrets Manager that is linked to your Pinecone API key.
* `namespace` – (Optional) The namespace to be used to write new data to your database.
* `field_mapping` – (Required) Contains the names of the fields to which to map information about the vector store.
    * `metadata_field` – (Required) The name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` – (Required) The name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.

 RDS Configuration supports the following:

* `credentials_secret_arn` – (Required) The ARN of the secret that you created in AWS Secrets Manager that is linked to your Amazon RDS database.
* `database_name` – (Required) The name of your Amazon RDS database.
* `resource_arn` – (Required) The namespace to be used to write new data to your database.
* `table_name` – (Required) The name of the table in the database.
* `field_mapping` – (Required) Contains the names of the fields to which to map information about the vector store.
    * `metadata_field` – (Required) The name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `primary_key_field` – (Required) The name of the field in which Amazon Bedrock stores the ID for each entry.
    * `text_field` – (Required) The name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` – (Required) The name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.

Redis Enterprise Cloud Configuration supports the following:

* `credentials_secret_arn` – (Required) The ARN of the secret that you created in AWS Secrets Manager that is linked to your Redis Enterprise Cloud database.
* `endpoint` – (Required) The endpoint URL of the Redis Enterprise Cloud database.
* `resource_arn` – (Required) The namespace to be used to write new data to your database.
* `vector_index_name` – (Required) The name of the vector index.
* `field_mapping` – (Required) Contains the names of the fields to which to map information about the vector store.
    * `metadata_field` – (Required) The name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` – (Required) The name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` – (Required) The name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.

Opensearch Serverless Configuration supports the following:

* `collection_arn` – (Required) The ARN of the OpenSearch Service vector store.
* `vector_index_name` – (Required) The name of the vector store.
* `field_mapping` – (Required) Contains the names of the fields to which to map information about the vector store.
    * `metadata_field` – (Required) The name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` – (Required) The name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` – (Required) The name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.

MongoDB Atlas Configuration supports the following:

* `collection_name` – (Required) The name of the collection in the MongoDB Atlas database.
* `credentials_secret_arn` – (Required) The ARN of the secret that you created in AWS Secrets Manager that is linked to your MongoDB Atlas database.
* `database_name` – (Required) The name of the database in the MongoDB Atlas database.
* `field_mapping` – (Required) Contains the names of the fields to which to map information about the vector store.
    * `metadata_field` – (Required) The name of the field in which Amazon Bedrock stores metadata about the vector store.
    * `text_field` – (Required) The name of the field in which Amazon Bedrock stores the raw text from your data. The text is split according to the chunking strategy you choose.
    * `vector_field` – (Required) The name of the field in which Amazon Bedrock stores the vector embeddings for your data sources.
* `vectorIndexName` – (Required) The name of the vector index.
* `endpointServiceName` – (Required) The name of the service that hosts the MongoDB Atlas database.
* `endpoint` – (Required) The endpoint URL of the MongoDB Atlas database.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Knowledge Base. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Agents for Amazon Bedrock Knowledge Base using the `example_id_arg`. For example:

```terraform
import {
  to = aws_bedrockagent_knowledge_base.example
  id = "Q1IYMH6GQG"
}
```

Using `terraform import`, import Agents for Amazon Bedrock Knowledge Base using the `Q1IYMH6GQG`. For example:

```console
% terraform import aws_bedrockagent_knowledge_base.example Q1IYMH6GQG
```
