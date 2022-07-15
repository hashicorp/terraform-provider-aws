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

```hcl
data "aws_kendra_index" "example" {
  id = "12345678-1234-1234-1234-123456789123"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) Returns information on a specific Index by id.

## Attributes Reference

In addition to all of the arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Index.
* `capacity_units` - A block that sets the number of additional document storage and query capacity units that should be used by the index. Documented below.
* `created_at` - The Unix datetime that the index was created.
* `description` - The description of the Index.
* `document_metadata_configuration_updates` - One or more blocks that specify the configuration settings for any metadata applied to the documents in the index. Documented below.
* `edition` - The Amazon Kendra edition for the index.
* `error_message` - When the Status field value is `FAILED`, this contains a message that explains why.
* `id` - The identifier of the Index.
* `index_statistics` - A block that provides information about the number of FAQ questions and answers and the number of text documents indexed. Documented below.
* `name` - Specifies the name of the Index.
* `role_arn` - An AWS Identity and Access Management (IAM) role that gives Amazon Kendra permissions to access your Amazon CloudWatch logs and metrics. This is also the role you use when you call the `BatchPutDocument` API to index documents from an Amazon S3 bucket.
* `server_side_encryption_configuration` - A block that specifies the identifier of the AWS KMS customer managed key (CMK) that's used to encrypt data indexed by Amazon Kendra. Amazon Kendra doesn't support asymmetric CMKs. Documented below.
* `status` - The current status of the index. When the value is `ACTIVE`, the index is ready for use. If the Status field value is `FAILED`, the `error_message` field contains a message that explains why.
* `updated_at` - The Unix datetime that the index was last updated.
* `user_context_policy` - The user context policy. Valid values are `ATTRIBUTE_FILTER` or `USER_TOKEN`. For more information, refer to [UserContextPolicy](https://docs.aws.amazon.com/kendra/latest/dg/API_CreateIndex.
html#Kendra-CreateIndex-request-UserContextPolicy).
* `user_group_resolution_configuration` - A block that enables fetching access levels of groups and users from an AWS Single Sign-On identity source. Documented below.
* `user_token_configurations` - A block that specifies the user token configuration. Documented below.
* `tags` - Metadata that helps organize the Indices you create.

A `capacity_units` block supports the following attributes:

* `query_capacity_units` - The amount of extra query capacity for an index and GetQuerySuggestions capacity. For more information, refer to [QueryCapacityUnits](https://docs.aws.amazon.com/kendra/latest/dg/API_CapacityUnitsConfiguration.html#Kendra-Type-CapacityUnitsConfiguration-QueryCapacityUnits).
* `storage_capacity_units` - The amount of extra storage capacity for an index. A single capacity unit provides 30 GB of storage space or 100,000 documents, whichever is reached first. Minimum value of 0.

A `document_metadata_configuration_updates` block supports the following attributes:

* `name` - The name of the index field. Minimum length of 1. Maximum length of 30.
* `relevance` - A block that provides manual tuning parameters to determine how the field affects the search results. Documented below.
* `search` - A block that provides information about how the field is used during a search. Documented below.
* `type` - The data type of the index field. Valid values are `STRING_VALUE`, `STRING_LIST_VALUE`, `LONG_VALUE`, `DATE_VALUE`.

A `relevance` block supports the following attributes:

* `duration` - Specifies the time period that the boost applies to. For more information, refer to [Duration](https://docs.aws.amazon.com/kendra/latest/dg/API_Relevance.html#Kendra-Type-Relevance-Duration).
* `freshness` - Indicates that this field determines how "fresh" a document is. For more information, refer to [Freshness](https://docs.aws.amazon.com/kendra/latest/dg/API_Relevance.html#Kendra-Type-Relevance-Freshness).
* `importance` - The relative importance of the field in the search. Larger numbers provide more of a boost than smaller numbers. Minimum value of 1. Maximum value of 10.
* `rank_order` - Determines how values should be interpreted. For more information, refer to [RankOrder](https://docs.aws.amazon.com/kendra/latest/dg/API_Relevance.html#Kendra-Type-Relevance-RankOrder).
* `values_importance_map` - A list of values that should be given a different boost when they appear in the result list. For more information, refer to [ValueImportanceMap](https://docs.aws.amazon.com/kendra/latest/dg/API_Relevance.html#Kendra-Type-Relevance-ValueImportanceMap).

A `search` block supports the following attributes:

* `displayable` - Determines whether the field is returned in the query response. The default is `true`.
* `facetable` - Indicates that the field can be used to create search facets, a count of results for each value in the field. The default is `false`.
* `searchable` - Determines whether the field is used in the search. If the Searchable field is true, you can use relevance tuning to manually tune how Amazon Kendra weights the field in the search. The default is `true` for `string` fields and `false` for `number` and `date` fields.
* `sortable` - Determines whether the field can be used to sort the results of a query. If you specify sorting on a field that does not have Sortable set to true, Amazon Kendra returns an exception. The default is `false`.

A `index_statistics` block supports the following attributes:

* `faq_statistics` - A block that specifies the number of question and answer topics in the index. Documented below.
* `text_document_statistics` - A block that specifies the number of text documents indexed.

A `faq_statistics` block supports the following attributes:

* `indexed_question_answers_count` - The total number of FAQ questions and answers contained in the index.

A `text_document_statistics` block supports the following attributes:

* `indexed_text_bytes` - The total size, in bytes, of the indexed documents.
* `indexed_text_documents_count` - The number of text documents indexed.

A `server_side_encryption_configuration` block supports the following attributes:

* `kms_key_id` - The identifier of the AWS KMScustomer master key (CMK). Amazon Kendra doesn't support asymmetric CMKs.

A `user_group_resolution_configuration` block supports the following attributes:

* `user_group_resolution_mode` - The identity store provider (mode) you want to use to fetch access levels of groups and users. AWS Single Sign-On is currently the only available mode. Your users and groups must exist in an AWS SSO identity source in order to use this mode. Valid Values are `AWS_SSO` or `NONE`.

A `user_token_configurations` block supports the following attributes:

* `json_token_type_configuration` - A block that specifies the information about the JSON token type configuration.
* `jwt_token_type_configuration` - A block that specifies the information about the JWT token type configuration.

A `json_token_type_configuration` block supports the following attributes:

* `group_attribute_field` - The group attribute field.
* `user_name_attribute_field` - The user name attribute field.

A `jwt_token_type_configuration` block supports the following attributes:

* `claim_regex` - The regular expression that identifies the claim.
* `group_attribute_field` - The group attribute field.
* `issuer` - The issuer of the token.
* `key_location` - The location of the key. Valid values are `URL` or `SECRET_MANAGER`
* `secrets_manager_arn` - The Amazon Resource Name (ARN) of the secret.
* `url` - The signing key URL.
* `user_name_attribute_field` - The user name attribute field.
