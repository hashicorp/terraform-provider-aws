---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_index"
description: |-
  Provides details about a specific Amazon Kendra Index.
---

# Data Source: aws_kendra_index

Provides details about a specific Amazon Kendra Index.

## Example Usage

```terraform
data "aws_kendra_index" "example" {
  id = "12345678-1234-1234-1234-123456789123"
}
```

## Argument Reference

This data source supports the following arguments:

* `id` - (Required) Returns information on a specific Index by id.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Index.
* `capacity_units` - Block that sets the number of additional document storage and query capacity units that should be used by the index. [Detailed below](#capacity_units-block).
* `created_at` - Unix datetime that the index was created.
* `description` - Description of the Index.
* `document_metadata_configuration_updates` - One or more blocks that specify the configuration settings for any metadata applied to the documents in the index. [Detailed below](#document_metadata_configuration_updates-block).
* `edition` - Amazon Kendra edition for the index.
* `error_message` - When the Status field value is `FAILED`, this contains a message that explains why.
* `index_statistics` - Block that provides information about the number of FAQ questions and answers and the number of text documents indexed. [Detailed below](#index_statistics-block).
* `name` - Name of the Index.
* `role_arn` - AWS Identity and Access Management (IAM) role that gives Amazon Kendra permissions to access your Amazon CloudWatch logs and metrics. This is also the role you use when you call the `BatchPutDocument` API to index documents from an Amazon S3 bucket.
* `server_side_encryption_configuration` - Block that specifies the identifier of the AWS KMS customer managed key (CMK) that's used to encrypt data indexed by Amazon Kendra. Amazon Kendra doesn't support asymmetric CMKs. [Detailed below](#server_side_encryption_configuration-block).
* `status` - Current status of the index. When the value is `ACTIVE`, the index is ready for use. If the Status field value is `FAILED`, the `error_message` field contains a message that explains why.
* `tags` - Metadata that helps organize the Indices you create.
* `updated_at` - Unix datetime that the index was last updated.
* `user_context_policy` - User context policy. Valid values are `ATTRIBUTE_FILTER` or `USER_TOKEN`. For more information, refer to [UserContextPolicy](https://docs.aws.amazon.com/kendra/latest/APIReference/API_CreateIndex.html#kendra-CreateIndex-request-UserContextPolicy).
* `user_group_resolution_configuration` - Block that enables fetching access levels of groups and users from an AWS Single Sign-On identity source. [Detailed below](#user_group_resolution_configuration-block).
* `user_token_configurations` - Block that specifies the user token configuration. [Detailed below](#user_token_configurations-block).

### `capacity_units` Block

A `capacity_units` block supports the following attributes:

* `query_capacity_units` - Amount of extra query capacity for an index and GetQuerySuggestions capacity. For more information, refer to [QueryCapacityUnits](https://docs.aws.amazon.com/kendra/latest/APIReference/API_CapacityUnitsConfiguration.html#Kendra-Type-CapacityUnitsConfiguration-QueryCapacityUnits).
* `storage_capacity_units` - Amount of extra storage capacity for an index. A single capacity unit provides 30 GB of storage space or 100,000 documents, whichever is reached first. Minimum value of 0.

### `document_metadata_configuration_updates` Block

A `document_metadata_configuration_updates` block supports the following attributes:

* `name` - Name of the index field. Minimum length of 1. Maximum length of 30.
* `relevance` - Block that provides manual tuning parameters to determine how the field affects the search results. [Detailed below](#relevance-block).
* `search` - Block that provides information about how the field is used during a search. [Detailed below](#search-block).
* `type` - Data type of the index field. Valid values are `STRING_VALUE`, `STRING_LIST_VALUE`, `LONG_VALUE`, `DATE_VALUE`.

#### `relevance` Block

A `relevance` block supports the following attributes:

* `duration` - Time period that the boost applies to. For more information, refer to [Duration](https://docs.aws.amazon.com/kendra/latest/APIReference/API_Relevance.html#Kendra-Type-Relevance-Duration).
* `freshness` - How "fresh" a document is. For more information, refer to [Freshness](https://docs.aws.amazon.com/kendra/latest/APIReference/API_Relevance.html#Kendra-Type-Relevance-Freshness).
* `importance` - Relative importance of the field in the search. Larger numbers provide more of a boost than smaller numbers. Minimum value of 1. Maximum value of 10.
* `rank_order` - How values should be interpreted. For more information, refer to [RankOrder](https://docs.aws.amazon.com/kendra/latest/APIReference/API_Relevance.html#Kendra-Type-Relevance-RankOrder).
* `values_importance_map` - List of values that should be given a different boost when they appear in the result list. For more information, refer to [ValueImportanceMap](https://docs.aws.amazon.com/kendra/latest/APIReference/API_Relevance.html#Kendra-Type-Relevance-ValueImportanceMap).

#### `search` Block

A `search` block supports the following attributes:

* `displayable` - Whether the field is returned in the query response. The default is `true`.
* `facetable` - Whether the field can be used to create search facets, a count of results for each value in the field. The default is `false`.
* `searchable` - Whether the field is used in the search. If the Searchable field is true, you can use relevance tuning to manually tune how Amazon Kendra weights the field in the search. The default is `true` for `string` fields and `false` for `number` and `date` fields.
* `sortable` - Whether the field can be used to sort the results of a query. If you specify sorting on a field that does not have Sortable set to true, Amazon Kendra returns an exception. The default is `false`.

### `index_statistics` Block

A `index_statistics` block supports the following attributes:

* `faq_statistics` - Block that specifies the number of question and answer topics in the index. [Detailed below](#faq_statistics-block).
* `text_document_statistics` - Block that specifies the number of text documents indexed. [Detailed below](#text_document_statistics-block).

#### `faq_statistics` Block

A `faq_statistics` block supports the following attributes:

* `indexed_question_answers_count` - Total number of FAQ questions and answers contained in the index.

#### `text_document_statistics` Block

A `text_document_statistics` block supports the following attributes:

* `indexed_text_bytes` - Total size, in bytes, of the indexed documents.
* `indexed_text_documents_count` - Number of text documents indexed.

### `server_side_encryption_configuration` Block

A `server_side_encryption_configuration` block supports the following attributes:

* `kms_key_id` - Identifier of the AWS KMS customer master key (CMK). Amazon Kendra doesn't support asymmetric CMKs.

### `user_group_resolution_configuration` Block

A `user_group_resolution_configuration` block supports the following attributes:

* `user_group_resolution_mode` - Identity store provider (mode) you want to use to fetch access levels of groups and users. AWS Single Sign-On is currently the only available mode. Your users and groups must exist in an AWS SSO identity source in order to use this mode. Valid Values are `AWS_SSO` or `NONE`.

### `user_token_configurations` Block

A `user_token_configurations` block supports the following attributes:

* `json_token_type_configuration` - Block that specifies the information about the JSON token type configuration. [Detailed below](#json_token_type_configuration-block).
* `jwt_token_type_configuration` - Block that specifies the information about the JWT token type configuration. [Detailed below](#jwt_token_type_configuration-block).

#### `json_token_type_configuration` Block

A `json_token_type_configuration` block supports the following attributes:

* `group_attribute_field` - Group attribute field.
* `user_name_attribute_field` - User name attribute field.

#### `jwt_token_type_configuration` Block

A `jwt_token_type_configuration` block supports the following attributes:

* `claim_regex` - Regular expression that identifies the claim.
* `group_attribute_field` - Group attribute field.
* `issuer` - Issuer of the token.
* `key_location` - Location of the key. Valid values are `URL` or `SECRET_MANAGER`.
* `secrets_manager_arn` - ARN of the secret.
* `url` - Signing key URL.
* `user_name_attribute_field` - User name attribute field.
