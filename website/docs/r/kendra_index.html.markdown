---
subcategory: "Kendra"
layout: "aws"
page_title: "AWS: aws_kendra_index"
description: |-
  Provides an Amazon Kendra Index resource.
---

# Resource: aws_kendra_index

Provides an Amazon Kendra Index resource.

## Example Usage

### Basic

```terraform
resource "aws_kendra_index" "example" {
  name        = "example"
  description = "example"
  edition     = "DEVELOPER_EDITION"
  role_arn    = aws_iam_role.this.arn

  tags = {
    "Key1" = "Value1"
  }
}
```

### With capacity units

```terraform
resource "aws_kendra_index" "example" {
  name     = "example"
  edition  = "DEVELOPER_EDITION"
  role_arn = aws_iam_role.this.arn

  capacity_units {
    query_capacity_units   = 2
    storage_capacity_units = 2
  }
}
```

### With server side encryption configuration

```terraform
resource "aws_kendra_index" "example" {
  name     = "example"
  role_arn = aws_iam_role.this.arn

  server_side_encryption_configuration {
    kms_key_id = data.aws_kms_key.this.arn
  }
}
```

### With user group resolution configuration

```terraform
resource "aws_kendra_index" "example" {
  name     = "example"
  role_arn = aws_iam_role.this.arn

  user_group_resolution_configuration {
    user_group_resolution_mode = "AWS_SSO"
  }
}
```

### With Document Metadata Configuration Updates

#### Specifying the predefined elements

Refer to [Amazon Kendra documentation on built-in document fields](https://docs.aws.amazon.com/kendra/latest/dg/hiw-index.html#index-reserved-fields) for more information.

```terraform
resource "aws_kendra_index" "example" {
  name     = "example"
  role_arn = aws_iam_role.this.arn

  document_metadata_configuration_updates {
    name = "_authors"
    type = "STRING_LIST_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = false
    }
    relevance {
      importance = 1
    }
  }

  document_metadata_configuration_updates {
    name = "_category"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_created_at"
    type = "DATE_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      freshness  = false
      importance = 1
      duration   = "25920000s"
      rank_order = "ASCENDING"
    }
  }

  document_metadata_configuration_updates {
    name = "_data_source_id"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_document_title"
    type = "STRING_VALUE"
    search {
      displayable = true
      facetable   = false
      searchable  = true
      sortable    = true
    }
    relevance {
      importance            = 2
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_excerpt_page_number"
    type = "LONG_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = false
    }
    relevance {
      importance = 2
      rank_order = "ASCENDING"
    }
  }

  document_metadata_configuration_updates {
    name = "_faq_id"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_file_type"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_language_code"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_last_updated_at"
    type = "DATE_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      freshness  = false
      importance = 1
      duration   = "25920000s"
      rank_order = "ASCENDING"
    }
  }

  document_metadata_configuration_updates {
    name = "_source_uri"
    type = "STRING_VALUE"
    search {
      displayable = true
      facetable   = false
      searchable  = false
      sortable    = false
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_tenant_id"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_version"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_view_count"
    type = "LONG_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance = 1
      rank_order = "ASCENDING"
    }
  }
}
```

#### Appending additional elements

The example below shows additional elements with names, `example-string-value`, `example-long-value`, `example-string-list-value`, `example-date-value` representing the 4 types of `STRING_VALUE`, `LONG_VALUE`, `STRING_LIST_VALUE`, `DATE_VALUE` respectively.

```terraform
resource "aws_kendra_index" "example" {
  name     = "example"
  role_arn = aws_iam_role.this.arn

  document_metadata_configuration_updates {
    name = "_authors"
    type = "STRING_LIST_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = false
    }
    relevance {
      importance = 1
    }
  }

  document_metadata_configuration_updates {
    name = "_category"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_created_at"
    type = "DATE_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      freshness  = false
      importance = 1
      duration   = "25920000s"
      rank_order = "ASCENDING"
    }
  }

  document_metadata_configuration_updates {
    name = "_data_source_id"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_document_title"
    type = "STRING_VALUE"
    search {
      displayable = true
      facetable   = false
      searchable  = true
      sortable    = true
    }
    relevance {
      importance            = 2
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_excerpt_page_number"
    type = "LONG_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = false
    }
    relevance {
      importance = 2
      rank_order = "ASCENDING"
    }
  }

  document_metadata_configuration_updates {
    name = "_faq_id"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_file_type"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_language_code"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_last_updated_at"
    type = "DATE_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      freshness  = false
      importance = 1
      duration   = "25920000s"
      rank_order = "ASCENDING"
    }
  }

  document_metadata_configuration_updates {
    name = "_source_uri"
    type = "STRING_VALUE"
    search {
      displayable = true
      facetable   = false
      searchable  = false
      sortable    = false
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_tenant_id"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_version"
    type = "STRING_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "_view_count"
    type = "LONG_VALUE"
    search {
      displayable = false
      facetable   = false
      searchable  = false
      sortable    = true
    }
    relevance {
      importance = 1
      rank_order = "ASCENDING"
    }
  }

  document_metadata_configuration_updates {
    name = "example-string-value"
    type = "STRING_VALUE"
    search {
      displayable = true
      facetable   = true
      searchable  = true
      sortable    = true
    }
    relevance {
      importance            = 1
      values_importance_map = {}
    }
  }

  document_metadata_configuration_updates {
    name = "example-long-value"
    type = "LONG_VALUE"
    search {
      displayable = true
      facetable   = true
      searchable  = false
      sortable    = true
    }
    relevance {
      importance = 1
      rank_order = "ASCENDING"
    }
  }

  document_metadata_configuration_updates {
    name = "example-string-list-value"
    type = "STRING_LIST_VALUE"
    search {
      displayable = true
      facetable   = true
      searchable  = true
      sortable    = false
    }
    relevance {
      importance = 1
    }
  }

  document_metadata_configuration_updates {
    name = "example-date-value"
    type = "DATE_VALUE"
    search {
      displayable = true
      facetable   = true
      searchable  = false
      sortable    = false
    }
    relevance {
      freshness  = false
      importance = 1
      duration   = "25920000s"
      rank_order = "ASCENDING"
    }
  }
}
```

### With JSON token type configuration

```terraform
resource "aws_kendra_index" "example" {
  name     = "example"
  role_arn = aws_iam_role.this.arn

  user_token_configurations {
    json_token_type_configuration {
      group_attribute_field     = "groups"
      user_name_attribute_field = "username"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `capacity_units` - (Optional) Block that sets the number of additional document storage and query capacity units that should be used by the index. [Detailed below](#capacity_units-block).
* `description` - (Optional) Description of the Index.
* `document_metadata_configuration_updates` - (Optional) One or more blocks that specify the configuration settings for any metadata applied to the documents in the index. Minimum number of 0 items. Maximum number of 500 items. If specified, you must define all elements, including those that are provided by default. These index fields are documented at [Amazon Kendra Index documentation](https://docs.aws.amazon.com/kendra/latest/dg/hiw-index.html). For an example resource that defines these default index fields, refer to the [default example above](#specifying-the-predefined-elements). For an example resource that appends additional index fields, refer to the [append example above](#appending-additional-elements). All arguments for each block must be specified. Note that blocks cannot be removed since index fields cannot be deleted. This argument is [detailed below](#document_metadata_configuration_updates-block).
* `edition` - (Optional) Amazon Kendra edition to use for the index. Choose `DEVELOPER_EDITION` for indexes intended for development, testing, or proof of concept. Use `ENTERPRISE_EDITION` for your production databases. Use `GEN_AI_ENTERPRISE_EDITION` for creating generative AI applications. Once you set the edition for an index, it can't be changed. Defaults to `ENTERPRISE_EDITION`.
* `name` - (Required) Name of the Index.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `role_arn` - (Required) AWS Identity and Access Management (IAM) role that gives Amazon Kendra permissions to access your Amazon CloudWatch logs and metrics. This is also the role you use when you call the `BatchPutDocument` API to index documents from an Amazon S3 bucket.
* `server_side_encryption_configuration` - (Optional) Block that specifies the identifier of the AWS KMS customer managed key (CMK) that's used to encrypt data indexed by Amazon Kendra. Amazon Kendra doesn't support asymmetric CMKs. [Detailed below](#server_side_encryption_configuration-block).
* `tags` - (Optional) Map of tags to assign to the Index. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_context_policy` - (Optional) User context policy. Valid values are `ATTRIBUTE_FILTER` or `USER_TOKEN`. For more information, refer to [UserContextPolicy](https://docs.aws.amazon.com/kendra/latest/APIReference/API_CreateIndex.html#kendra-CreateIndex-request-UserContextPolicy). Defaults to `ATTRIBUTE_FILTER`.
* `user_group_resolution_configuration` - (Optional) Block that enables fetching access levels of groups and users from an AWS Single Sign-On identity source. To configure this, see [UserGroupResolutionConfiguration](https://docs.aws.amazon.com/kendra/latest/dg/API_UserGroupResolutionConfiguration.html). [Detailed below](#user_group_resolution_configuration-block).
* `user_token_configurations` - (Optional) Block that specifies the user token configuration. [Detailed below](#user_token_configurations-block).

### `capacity_units` Block

A `capacity_units` block supports the following arguments:

* `query_capacity_units` - (Optional) Amount of extra query capacity for an index and GetQuerySuggestions capacity. For more information, refer to [QueryCapacityUnits](https://docs.aws.amazon.com/kendra/latest/dg/API_CapacityUnitsConfiguration.html#Kendra-Type-CapacityUnitsConfiguration-QueryCapacityUnits).
* `storage_capacity_units` - (Optional) Amount of extra storage capacity for an index. A single capacity unit provides 30 GB of storage space or 100,000 documents, whichever is reached first. Minimum value of 0.

### `document_metadata_configuration_updates` Block

A `document_metadata_configuration_updates` block supports the following arguments:

* `name` - (Required) Name of the index field. Minimum length of 1. Maximum length of 30.
* `relevance` - (Required) Block that provides manual tuning parameters to determine how the field affects the search results. [Detailed below](#relevance-block)
* `search` - (Required) Block that provides information about how the field is used during a search. [Detailed below](#search-block)
* `type` - (Required) Data type of the index field. Valid values are `STRING_VALUE`, `STRING_LIST_VALUE`, `LONG_VALUE`, `DATE_VALUE`.

#### `relevance` Block

A `relevance` block supports the following arguments:

* `duration` - (Optional) Time period that the boost applies to. Required if `type` is `DATE_VALUE`. For more information, refer to [Duration](https://docs.aws.amazon.com/kendra/latest/dg/API_Relevance.html#Kendra-Type-Relevance-Duration).
* `freshness` - (Optional) Whether this field determines how "fresh" a document is. Required if `type` is `DATE_VALUE`. For more information, refer to [Freshness](https://docs.aws.amazon.com/kendra/latest/dg/API_Relevance.html#Kendra-Type-Relevance-Freshness).
* `importance` - (Optional) Relative importance of the field in the search. Required for all types. Larger numbers provide more of a boost than smaller numbers. Minimum value of 1. Maximum value of 10.
* `rank_order` - (Optional) How values should be interpreted. Required if `type` is `DATE_VALUE` or `LONG_VALUE`. For more information, refer to [RankOrder](https://docs.aws.amazon.com/kendra/latest/dg/API_Relevance.html#Kendra-Type-Relevance-RankOrder).
* `values_importance_map` - (Optional) List of values that should be given a different boost when they appear in the result list. Required if `type` is `STRING_VALUE`. For more information, refer to [ValueImportanceMap](https://docs.aws.amazon.com/kendra/latest/dg/API_Relevance.html#Kendra-Type-Relevance-ValueImportanceMap).

#### `search` Block

A `search` block supports the following arguments:

* `displayable` - (Optional) Whether the field is returned in the query response. The default is `true`.
* `facetable` - (Optional) Whether the field can be used to create search facets, a count of results for each value in the field. The default is `false`.
* `searchable` - (Optional) Whether the field is used in the search. If the Searchable field is true, you can use relevance tuning to manually tune how Amazon Kendra weights the field in the search. The default is `true` for `string` fields and `false` for `number` and `date` fields.
* `sortable` - (Optional) Whether the field can be used to sort the results of a query. If you specify sorting on a field that does not have Sortable set to true, Amazon Kendra returns an exception. The default is `false`.

### `server_side_encryption_configuration` Block

A `server_side_encryption_configuration` block supports the following arguments:

* `kms_key_id` - (Optional) Identifier of the AWS KMS customer master key (CMK). Amazon Kendra doesn't support asymmetric CMKs.

### `user_group_resolution_configuration` Block

A `user_group_resolution_configuration` block supports the following arguments:

* `user_group_resolution_mode` - (Required) Identity store provider (mode) you want to use to fetch access levels of groups and users. AWS Single Sign-On is currently the only available mode. Your users and groups must exist in an AWS SSO identity source in order to use this mode. Valid Values are `AWS_SSO` or `NONE`.

### `user_token_configurations` Block

A `user_token_configurations` block supports the following arguments:

* `json_token_type_configuration` - (Optional) Block that specifies the information about the JSON token type configuration. [Detailed below](#json_token_type_configuration-block).
* `jwt_token_type_configuration` - (Optional) Block that specifies the information about the JWT token type configuration. [Detailed below](#jwt_token_type_configuration-block).

#### `json_token_type_configuration` Block

A `json_token_type_configuration` block supports the following arguments:

* `group_attribute_field` - (Required) Group attribute field. Minimum length of 1. Maximum length of 2048.
* `user_name_attribute_field` - (Required) User name attribute field. Minimum length of 1. Maximum length of 2048.

#### `jwt_token_type_configuration` Block

A `jwt_token_type_configuration` block supports the following arguments:

* `claim_regex` - (Optional) Regular expression that identifies the claim. Minimum length of 1. Maximum length of 100.
* `group_attribute_field` - (Optional) Group attribute field. Minimum length of 1. Maximum length of 100.
* `issuer` - (Optional) Issuer of the token. Minimum length of 1. Maximum length of 65.
* `key_location` - (Required) Location of the key. Valid values are `URL` or `SECRET_MANAGER`.
* `secrets_manager_arn` - (Optional) ARN of the secret.
* `url` - (Optional) Signing key URL. Valid pattern is `^(https?|ftp|file):\/\/([^\s]*)`.
* `user_name_attribute_field` - (Optional) User name attribute field. Minimum length of 1. Maximum length of 100.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Index.
* `created_at` - Unix datetime that the index was created.
* `error_message` - When the Status field value is `FAILED`, this contains a message that explains why.
* `id` - Identifier of the Index.
* `index_statistics` - Block that provides information about the number of FAQ questions and answers and the number of text documents indexed. [Detailed below](#index_statistics-block).
* `status` - Current status of the index. When the value is `ACTIVE`, the index is ready for use. If the Status field value is `FAILED`, the `error_message` field contains a message that explains why.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `updated_at` - Unix datetime that the index was last updated.

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

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `40m`)
* `delete` - (Default `40m`)
* `update` - (Default `40m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Kendra Indexes using its `id`. For example:

```terraform
import {
  to = aws_kendra_index.example
  id = "12345678-1234-5678-9123-123456789123"
}
```

Using `terraform import`, import Amazon Kendra Indexes using its `id`. For example:

```console
% terraform import aws_kendra_index.example 12345678-1234-5678-9123-123456789123
```
