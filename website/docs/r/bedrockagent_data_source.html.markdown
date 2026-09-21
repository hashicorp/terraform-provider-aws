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

### Managed Knowledge Base Connector - S3

```terraform
resource "aws_bedrockagent_data_source" "example" {
  knowledge_base_id = aws_bedrockagent_knowledge_base.example.id
  name              = "example-s3-managed"

  data_source_configuration {
    type = "MANAGED_KNOWLEDGE_BASE_CONNECTOR"

    managed_knowledge_base_connector_configuration {
      connector_parameters = jsonencode({
        type    = "S3"
        version = "1"
        connectionConfiguration = {
          bucketName           = "my-documents-bucket"
          bucketOwnerAccountId = "123456789012"
        }
        aclEnabled = false
        filterConfiguration = {
          maxFileSizeInMegaBytes = "500"
        }
      })

      media_extraction_configuration {
        image_extraction_configuration {
          image_extraction_status = "ENABLED"
        }
      }
    }
  }

  vector_ingestion_configuration {
    parsing_configuration {
      parsing_strategy = "SMART_PARSING"
    }
  }
}
```

### Managed Knowledge Base Connector - SharePoint

```terraform
resource "aws_bedrockagent_data_source" "sharepoint" {
  knowledge_base_id = aws_bedrockagent_knowledge_base.example.id
  name              = "example-sharepoint"

  data_source_configuration {
    type = "MANAGED_KNOWLEDGE_BASE_CONNECTOR"

    managed_knowledge_base_connector_configuration {
      connector_parameters = jsonencode({
        type    = "SHAREPOINT"
        version = "1"
        connectionConfiguration = {
          tenantId  = "your-entra-tenant-id"
          authType  = "ENTRA_ID_APP_ONLY"
          secretArn = "arn:aws:secretsmanager:us-east-1:123456789012:secret:my-sharepoint-secret"
          certificateS3Path = {
            s3BucketName = "my-certs-bucket"
            s3KeyName    = "certs/sharepoint-cert.crt"
          }
        }
        dataEntityConfiguration = {
          type       = "DOCUMENT"
          crawlFiles = "true"
          crawlPages = "true"
          siteUrls   = ["https://company.sharepoint.com/sites/MySite"]
        }
      })
    }
  }
}
```

### Multimodal Parsing

```terraform
resource "aws_bedrockagent_data_source" "example" {
  knowledge_base_id = aws_bedrockagent_knowledge_base.example.id
  name              = "multimodal-example"

  data_source_configuration {
    type = "S3"
    s3_configuration {
      bucket_arn = aws_s3_bucket.example.arn
    }
  }

  vector_ingestion_configuration {
    chunking_configuration {
      chunking_strategy = "FIXED_SIZE"
      fixed_size_chunking_configuration {
        max_tokens         = 512
        overlap_percentage = 20
      }
    }

    parsing_configuration {
      parsing_strategy = "BEDROCK_FOUNDATION_MODEL"
      bedrock_foundation_model_configuration {
        model_arn        = "arn:aws:bedrock:us-east-1::foundation-model/anthropic.claude-3-sonnet-20240229-v1:0"
        parsing_modality = "MULTIMODAL"
        parsing_prompt {
          parsing_prompt_string = "Extract and transcribe all text and visual content from the document."
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `data_source_configuration` - (Required) Details about how the data source is stored. See [`data_source_configuration` Block](#data_source_configuration-block) for details.
* `knowledge_base_id` - (Required) Unique identifier of the knowledge base to which the data source belongs.
* `name` - (Required, Forces new resource) Name of the data source.

The following arguments are optional:

* `data_deletion_policy` - (Optional) Data deletion policy for a data source. Valid values: `RETAIN`, `DELETE`.
* `description` - (Optional) Description of the data source.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `server_side_encryption_configuration` - (Optional) Details about the configuration of the server-side encryption. See [`server_side_encryption_configuration` Block](#server_side_encryption_configuration-block) for details.
* `vector_ingestion_configuration` - (Optional, Forces new resource) Details about how to ingest the documents in the data source. See [`vector_ingestion_configuration` Block](#vector_ingestion_configuration-block) for details.

### `data_source_configuration` Block

The `data_source_configuration` configuration block supports the following arguments:

* `confluence_configuration` - (Optional) Configuration details for the Confluence data source. See [`data_source_configuration.confluence_configuration` Block](#data_source_configurationconfluence_configuration-block) for details.
* `managed_knowledge_base_connector_configuration` - (Optional) Configuration details for a Managed Knowledge Base connector data source. See [`managed_knowledge_base_connector_configuration` Block](#managed_knowledge_base_connector_configuration-block) for details.
* `s3_configuration` - (Optional) Configuration details for the S3 object that contains the data source. See [`s3_configuration` Block](#s3_configuration-block) for details.
* `salesforce_configuration` - (Optional) Configuration details for the Salesforce data source. See [`data_source_configuration.salesforce_configuration` Block](#data_source_configurationsalesforce_configuration-block) for details.
* `share_point_configuration` - (Optional) Configuration details for the SharePoint data source. See [`data_source_configuration.share_point_configuration` Block](#data_source_configurationshare_point_configuration-block) for details.
* `type` - (Required) Type of storage for the data source. Valid values: `S3`, `WEB`, `CONFLUENCE`, `SALESFORCE`, `SHAREPOINT`, `CUSTOM`, `REDSHIFT_METADATA`, `MANAGED_KNOWLEDGE_BASE_CONNECTOR`.
* `web_configuration` - (Optional) Configuration details for the web data source. See [`data_source_configuration.web_configuration` Block](#data_source_configurationweb_configuration-block) for details.

### `data_source_configuration.confluence_configuration` Block

For more details, see the [Amazon BedrockAgent Confluence documentation](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_agent_ConfluenceDataSourceConfiguration.html).

The `confluence_configuration` configuration block supports the following arguments:

* `crawler_configuration` - (Optional) Configuration for Confluence content. See [`data_source_configuration.confluence_configuration.crawler_configuration` Block](#data_source_configurationconfluence_configurationcrawler_configuration-block) for details.
* `source_configuration` - (Optional) Endpoint information to connect to your Confluence data source. See [`data_source_configuration.confluence_configuration.source_configuration` Block](#data_source_configurationconfluence_configurationsource_configuration-block) for details.

### `data_source_configuration.confluence_configuration.crawler_configuration` Block

The `crawler_configuration` configuration block supports the following arguments:

* `filter_configuration` - (Optional) Object configuration used to filter crawled content. See [`data_source_configuration.confluence_configuration.crawler_configuration.filter_configuration` Block](#data_source_configurationconfluence_configurationcrawler_configurationfilter_configuration-block) for details.

### `data_source_configuration.confluence_configuration.crawler_configuration.filter_configuration` Block

The `filter_configuration` configuration block supports the following arguments:

* `pattern_object_filter` - (Optional) Configuration for filtering objects or content types of the data source. See [`data_source_configuration.confluence_configuration.crawler_configuration.filter_configuration.pattern_object_filter` Block](#data_source_configurationconfluence_configurationcrawler_configurationfilter_configurationpattern_object_filter-block) for details.
* `type` - (Required) Type of filtering to apply to objects or content of the data source. For example, the `PATTERN` type uses regular expression patterns to filter content.

### `data_source_configuration.confluence_configuration.crawler_configuration.filter_configuration.pattern_object_filter` Block

The `pattern_object_filter` configuration block supports the following arguments:

* `filters` - (Required) Filters applied to your data source content. Minimum of 1 filter and maximum of 25 filters. See [`data_source_configuration.confluence_configuration.crawler_configuration.filter_configuration.pattern_object_filter.filters` Block](#data_source_configurationconfluence_configurationcrawler_configurationfilter_configurationpattern_object_filterfilters-block) for details.

### `data_source_configuration.confluence_configuration.crawler_configuration.filter_configuration.pattern_object_filter.filters` Block

The `filters` configuration block supports the following arguments:

* `exclusion_filters` - (Optional) One or more exclusion regular expression patterns to exclude object types that match the pattern.
* `inclusion_filters` - (Optional) One or more inclusion regular expression patterns to include object types that match the pattern.
* `object_type` - (Required) Object type or content type of the data source.

### `data_source_configuration.confluence_configuration.source_configuration` Block

The `source_configuration` configuration block supports the following arguments:

* `auth_type` - (Required) Supported authentication type to authenticate and connect to your Confluence instance. Valid values: `BASIC`, `OAUTH2_CLIENT_CREDENTIALS`.
* `credentials_secret_arn` - (Required) ARN of an AWS Secrets Manager secret that stores your authentication credentials for your Confluence instance URL. For more information on the key-value pairs that must be included in your secret, depending on your authentication type, see Confluence connection configuration. Pattern: `^arn:aws(|-cn|-us-gov):secretsmanager:[a-z0-9-]{1,20}:([0-9]{12}|):secret:[a-zA-Z0-9!/_+=.@-]{1,512}$`.
* `host_type` - (Required) Supported host type, whether online/cloud or server/on-premises. Valid values: `SAAS`.
* `host_url` - (Required) Confluence host URL or instance URL. Pattern: `^https://[A-Za-z0-9][^\s]*$`.

### `managed_knowledge_base_connector_configuration` Block

The `managed_knowledge_base_connector_configuration` configuration block supports the following arguments:

* `connector_parameters` - (Optional) JSON-encoded string containing the connector-specific parameters. The structure depends on the connector type (S3, SharePoint, Google Drive, etc.). See [Managed Knowledge Base connector parameters](https://docs.aws.amazon.com/bedrock/latest/userguide/knowledge-base-connectors.html) for details on each connector type.
* `deletion_protection_configuration` - (Optional) Configuration for deletion protection on the data source. See [`deletion_protection_configuration` Block](#deletion_protection_configuration-block) for details.
* `media_extraction_configuration` - (Optional) Configuration for extracting media content (images, audio, video) from documents. See [`media_extraction_configuration` Block](#media_extraction_configuration-block) for details.

### `deletion_protection_configuration` Block

The `deletion_protection_configuration` configuration block supports the following arguments:

* `deletion_protection_status` - (Required) Enable or disable deletion protection for the connector. Valid values: `ENABLED`, `DISABLED`.
* `deletion_protection_threshold` - (Optional) Maximum percentage of documents that a sync job can delete from your index.

### `media_extraction_configuration` Block

The `media_extraction_configuration` configuration block supports the following arguments:

* `audio_extraction_configuration` - (Optional) Configuration for extracting audio content. See [`audio_extraction_configuration` Block](#audio_extraction_configuration-block) for details.
* `image_extraction_configuration` - (Optional) Configuration for extracting image content. See [`image_extraction_configuration` Block](#image_extraction_configuration-block) for details.
* `video_extraction_configuration` - (Optional) Configuration for extracting video content. See [`video_extraction_configuration` Block](#video_extraction_configuration-block) for details.

### `audio_extraction_configuration` Block

The `audio_extraction_configuration` configuration block supports the following arguments:

* `audio_extraction_status` - (Required) Whether audio extraction is enabled. Valid values: `ENABLED`, `DISABLED`.

### `image_extraction_configuration` Block

The `image_extraction_configuration` configuration block supports the following arguments:

* `image_extraction_status` - (Required) Whether image extraction is enabled. Valid values: `ENABLED`, `DISABLED`.

### `video_extraction_configuration` Block

The `video_extraction_configuration` configuration block supports the following arguments:

* `video_extraction_status` - (Required) Whether video extraction is enabled. Valid values: `ENABLED`, `DISABLED`.

### `s3_configuration` Block

The `s3_configuration` configuration block supports the following arguments:

* `bucket_arn` - (Required) ARN of the bucket that contains the data source.
* `bucket_owner_account_id` - (Optional) Bucket account owner ID for the S3 bucket.
* `inclusion_prefixes` - (Optional) List of S3 prefixes that define the object containing the data sources. For more information, see [Organizing objects using prefixes](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-prefixes.html).

### `data_source_configuration.salesforce_configuration` Block

For more details, see the [Amazon BedrockAgent Salesforce documentation](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_agent_SalesforceDataSourceConfiguration.html).

The `salesforce_configuration` configuration block supports the following arguments:

* `crawler_configuration` - (Optional) Configuration for Salesforce content. See [`data_source_configuration.salesforce_configuration.crawler_configuration` Block](#data_source_configurationsalesforce_configurationcrawler_configuration-block) for details.
* `source_configuration` - (Optional) Endpoint information to connect to your Salesforce data source. See [`data_source_configuration.salesforce_configuration.source_configuration` Block](#data_source_configurationsalesforce_configurationsource_configuration-block) for details.

### `data_source_configuration.salesforce_configuration.crawler_configuration` Block

The `crawler_configuration` configuration block supports the following arguments:

* `filter_configuration` - (Optional) Object configuration used to filter crawled content. See [`data_source_configuration.salesforce_configuration.crawler_configuration.filter_configuration` Block](#data_source_configurationsalesforce_configurationcrawler_configurationfilter_configuration-block) for details.

### `data_source_configuration.salesforce_configuration.crawler_configuration.filter_configuration` Block

The `filter_configuration` configuration block supports the following arguments:

* `pattern_object_filter` - (Optional) Configuration for filtering objects or content types of the data source. See [`data_source_configuration.salesforce_configuration.crawler_configuration.filter_configuration.pattern_object_filter` Block](#data_source_configurationsalesforce_configurationcrawler_configurationfilter_configurationpattern_object_filter-block) for details.
* `type` - (Required) Type of filtering to apply to objects or content of the data source. For example, the `PATTERN` type uses regular expression patterns to filter content.

### `data_source_configuration.salesforce_configuration.crawler_configuration.filter_configuration.pattern_object_filter` Block

The `pattern_object_filter` configuration block supports the following arguments:

* `filters` - (Required) Filters applied to your data source content. Minimum of 1 filter and maximum of 25 filters. See [`data_source_configuration.salesforce_configuration.crawler_configuration.filter_configuration.pattern_object_filter.filters` Block](#data_source_configurationsalesforce_configurationcrawler_configurationfilter_configurationpattern_object_filterfilters-block) for details.

### `data_source_configuration.salesforce_configuration.crawler_configuration.filter_configuration.pattern_object_filter.filters` Block

The `filters` configuration block supports the following arguments:

* `exclusion_filters` - (Optional) One or more exclusion regular expression patterns to exclude object types that match the pattern.
* `inclusion_filters` - (Optional) One or more inclusion regular expression patterns to include object types that match the pattern.
* `object_type` - (Required) Object type or content type of the data source.

### `data_source_configuration.salesforce_configuration.source_configuration` Block

The `source_configuration` configuration block supports the following arguments:

* `auth_type` - (Required) Supported authentication type to authenticate and connect to your Salesforce instance. Valid values: `OAUTH2_CLIENT_CREDENTIALS`.
* `credentials_secret_arn` - (Required) ARN of an AWS Secrets Manager secret that stores your authentication credentials for your Salesforce instance URL. For more information on the key-value pairs that must be included in your secret, depending on your authentication type, see Salesforce connection configuration. Pattern: `^arn:aws(|-cn|-us-gov):secretsmanager:[a-z0-9-]{1,20}:([0-9]{12}|):secret:[a-zA-Z0-9!/_+=.@-]{1,512}$`.
* `host_url` - (Required) Salesforce host URL or instance URL. Pattern: `^https://[A-Za-z0-9][^\s]*$`.

### `data_source_configuration.share_point_configuration` Block

For more details, see the [Amazon BedrockAgent SharePoint documentation](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_agent_SharePointDataSourceConfiguration.html).

The `share_point_configuration` configuration block supports the following arguments:

* `crawler_configuration` - (Optional) Configuration for SharePoint content. See [`data_source_configuration.share_point_configuration.crawler_configuration` Block](#data_source_configurationshare_point_configurationcrawler_configuration-block) for details.
* `source_configuration` - (Optional) Endpoint information to connect to your SharePoint data source. See [`data_source_configuration.share_point_configuration.source_configuration` Block](#data_source_configurationshare_point_configurationsource_configuration-block) for details.

### `data_source_configuration.share_point_configuration.crawler_configuration` Block

The `crawler_configuration` configuration block supports the following arguments:

* `filter_configuration` - (Optional) Object configuration used to filter crawled content. See [`data_source_configuration.share_point_configuration.crawler_configuration.filter_configuration` Block](#data_source_configurationshare_point_configurationcrawler_configurationfilter_configuration-block) for details.

### `data_source_configuration.share_point_configuration.crawler_configuration.filter_configuration` Block

The `filter_configuration` configuration block supports the following arguments:

* `pattern_object_filter` - (Optional) Configuration for filtering objects or content types of the data source. See [`data_source_configuration.share_point_configuration.crawler_configuration.filter_configuration.pattern_object_filter` Block](#data_source_configurationshare_point_configurationcrawler_configurationfilter_configurationpattern_object_filter-block) for details.
* `type` - (Required) Type of filtering to apply to objects or content of the data source. For example, the `PATTERN` type uses regular expression patterns to filter content.

### `data_source_configuration.share_point_configuration.crawler_configuration.filter_configuration.pattern_object_filter` Block

The `pattern_object_filter` configuration block supports the following arguments:

* `filters` - (Required) Filters applied to your data source content. Minimum of 1 filter and maximum of 25 filters. See [`data_source_configuration.share_point_configuration.crawler_configuration.filter_configuration.pattern_object_filter.filters` Block](#data_source_configurationshare_point_configurationcrawler_configurationfilter_configurationpattern_object_filterfilters-block) for details.

### `data_source_configuration.share_point_configuration.crawler_configuration.filter_configuration.pattern_object_filter.filters` Block

The `filters` configuration block supports the following arguments:

* `exclusion_filters` - (Optional) One or more exclusion regular expression patterns to exclude object types that match the pattern.
* `inclusion_filters` - (Optional) One or more inclusion regular expression patterns to include object types that match the pattern.
* `object_type` - (Required) Object type or content type of the data source.

### `data_source_configuration.share_point_configuration.source_configuration` Block

The `source_configuration` configuration block supports the following arguments:

* `auth_type` - (Required) Supported authentication type to authenticate and connect to your SharePoint site. Valid values: `OAUTH2_CLIENT_CREDENTIALS`, `OAUTH2_SHAREPOINT_APP_ONLY_CLIENT_CREDENTIALS`.
* `credentials_secret_arn` - (Required) ARN of an AWS Secrets Manager secret that stores your authentication credentials for your SharePoint site. For more information on the key-value pairs that must be included in your secret, depending on your authentication type, see SharePoint connection configuration. Pattern: `^arn:aws(|-cn|-us-gov):secretsmanager:[a-z0-9-]{1,20}:([0-9]{12}|):secret:[a-zA-Z0-9!/_+=.@-]{1,512}$`.
* `domain` - (Required) Domain of your SharePoint instance or site URL/URLs.
* `host_type` - (Required) Supported host type, whether online/cloud or server/on-premises. Valid values: `ONLINE`.
* `site_urls` - (Required) One or more SharePoint site URLs.
* `tenant_id` - (Optional) Identifier of your Microsoft 365 tenant.

### `data_source_configuration.web_configuration` Block

The `web_configuration` configuration block supports the following arguments:

* `crawler_configuration` - (Optional) Configuration for web content. See [`data_source_configuration.web_configuration.crawler_configuration` Block](#data_source_configurationweb_configurationcrawler_configuration-block) for details.
* `source_configuration` - (Optional) Endpoint information to connect to your web data source. See [`data_source_configuration.web_configuration.source_configuration` Block](#data_source_configurationweb_configurationsource_configuration-block) for details.

### `data_source_configuration.web_configuration.crawler_configuration` Block

The `crawler_configuration` configuration block supports the following arguments:

* `crawler_limits` - (Optional) Configuration of crawl limits for the web URLs. See [`crawler_limits` Block](#crawler_limits-block) for details.
* `exclusion_filters` - (Optional) List of one or more exclusion regular expression patterns to exclude object types that match the pattern.
* `inclusion_filters` - (Optional) List of one or more inclusion regular expression patterns to include object types that match the pattern.
* `scope` - (Optional) Scope of what is crawled for your URLs.
* `user_agent` - (Optional) String used to identify the crawler or bot when it accesses a web server. Default value is `bedrockbot_UUID`.

### `crawler_limits` Block

The `crawler_limits` configuration block supports the following arguments:

* `max_pages` - (Optional) Max number of web pages crawled from your source URLs, up to 25,000 pages.
* `rate_limit` - (Optional) Max rate at which pages are crawled, up to 300 per minute per host.

### `data_source_configuration.web_configuration.source_configuration` Block

The `source_configuration` configuration block supports the following arguments:

* `url_configuration` - (Required) URL configuration of your web data source. See [`url_configuration` Block](#url_configuration-block) for details.

### `url_configuration` Block

The `url_configuration` configuration block supports the following arguments:

* `seed_urls` - (Optional) List of one or more seed URLs to crawl. See [`seed_urls` Block](#seed_urls-block) for details.

### `seed_urls` Block

The `seed_urls` configuration block supports the following arguments:

* `url` - (Optional) Seed or starting point URL. Must match the pattern `^https?://[A-Za-z0-9][^\s]*$`.

### `server_side_encryption_configuration` Block

The `server_side_encryption_configuration` configuration block supports the following arguments:

* `kms_key_arn` - (Optional) ARN of the AWS KMS key used to encrypt the resource.

### `vector_ingestion_configuration` Block

The `vector_ingestion_configuration` configuration block supports the following arguments:

* `chunking_configuration` - (Optional, Forces new resource) Details about how to chunk the documents in the data source. A chunk refers to an excerpt from a data source that is returned when the knowledge base that it belongs to is queried. See [`chunking_configuration` Block](#chunking_configuration-block) for details.
* `custom_transformation_configuration` - (Optional, Forces new resource) Configuration for custom transformation of data source documents. See [`custom_transformation_configuration` Block](#custom_transformation_configuration-block) for details.
* `parsing_configuration` - (Optional, Forces new resource) Configuration for custom parsing of data source documents. See [`parsing_configuration` Block](#parsing_configuration-block) for details.

### `chunking_configuration` Block

The `chunking_configuration` configuration block supports the following arguments:

* `chunking_strategy` - (Required, Forces new resource) Option for chunking your source data, either in fixed-sized chunks or as one chunk. Valid values: `FIXED_SIZE`, `HIERARCHICAL`, `SEMANTIC`, `NONE`.
* `fixed_size_chunking_configuration` - (Optional, Forces new resource) Configurations for when you choose fixed-size chunking. Requires `chunking_strategy` as `FIXED_SIZE`. See [`fixed_size_chunking_configuration` Block](#fixed_size_chunking_configuration-block) for details.
* `hierarchical_chunking_configuration` - (Optional, Forces new resource) Configurations for when you choose hierarchical chunking. Requires `chunking_strategy` as `HIERARCHICAL`. See [`hierarchical_chunking_configuration` Block](#hierarchical_chunking_configuration-block) for details.
* `semantic_chunking_configuration` - (Optional, Forces new resource) Configurations for when you choose semantic chunking. Requires `chunking_strategy` as `SEMANTIC`. See [`semantic_chunking_configuration` Block](#semantic_chunking_configuration-block) for details.

### `fixed_size_chunking_configuration` Block

The `fixed_size_chunking_configuration` configuration block supports the following arguments:

* `max_tokens` - (Required, Forces new resource) Maximum number of tokens to include in a chunk.
* `overlap_percentage` - (Optional, Forces new resource) Percentage of overlap between adjacent chunks of a data source.

### `hierarchical_chunking_configuration` Block

The `hierarchical_chunking_configuration` configuration block supports the following arguments:

* `level_configuration` - (Required, Forces new resource) Token settings for each layer. Must contain two `level_configuration` blocks. See [`level_configuration` Block](#level_configuration-block) for details.
* `overlap_tokens` - (Required, Forces new resource) Number of tokens to repeat across chunks in the same layer.

### `level_configuration` Block

The `level_configuration` configuration block supports the following arguments:

* `max_tokens` - (Required) Maximum number of tokens that a chunk can contain in this layer.

### `semantic_chunking_configuration` Block

The `semantic_chunking_configuration` configuration block supports the following arguments:

* `breakpoint_percentile_threshold` - (Required, Forces new resource) Dissimilarity threshold for splitting chunks.
* `buffer_size` - (Required, Forces new resource) Buffer size.
* `max_token` - (Required, Forces new resource) Maximum number of tokens a chunk can contain.

### `custom_transformation_configuration` Block

The `custom_transformation_configuration` configuration block supports the following arguments:

* `intermediate_storage` - (Required, Forces new resource) Intermediate storage for custom transformation. See [`intermediate_storage` Block](#intermediate_storage-block) for details.
* `transformation` - (Required) Custom processing step for documents moving through the data source ingestion pipeline. See [`transformation` Block](#transformation-block) for details.

### `intermediate_storage` Block

The `intermediate_storage` configuration block supports the following arguments:

* `s3_location` - (Required, Forces new resource) Configuration block for intermediate S3 storage. See [`s3_location` Block](#s3_location-block) for details.

### `s3_location` Block

The `s3_location` configuration block supports the following arguments:

* `uri` - (Required, Forces new resource) S3 URI for intermediate storage.

### `transformation` Block

The `transformation` configuration block supports the following arguments:

* `step_to_apply` - (Required, Forces new resource) When the service applies the transformation. Currently only `POST_CHUNKING` is supported.
* `transformation_function` - (Required) Lambda function that processes documents. See [`transformation_function` Block](#transformation_function-block) for details.

### `transformation_function` Block

The `transformation_function` configuration block supports the following arguments:

* `transformation_lambda_configuration` - (Required, Forces new resource) Configuration of the Lambda function. See [`transformation_lambda_configuration` Block](#transformation_lambda_configuration-block) for details.

### `transformation_lambda_configuration` Block

The `transformation_lambda_configuration` configuration block supports the following arguments:

* `lambda_arn` - (Required, Forces new resource) ARN of the Lambda to use for custom transformation.

### `parsing_configuration` Block

The `parsing_configuration` configuration block supports the following arguments:

* `bedrock_data_automation_configuration` - (Optional) Settings for using Amazon Bedrock Data Automation to parse documents. See [`bedrock_data_automation_configuration` Block](#bedrock_data_automation_configuration-block) for details.
* `bedrock_foundation_model_configuration` - (Optional) Settings for a foundation model used to parse documents in a data source. See [`bedrock_foundation_model_configuration` Block](#bedrock_foundation_model_configuration-block) for details.
* `parsing_strategy` - (Required) Parsing strategy to use. Valid values: `BEDROCK_FOUNDATION_MODEL`, `BEDROCK_DATA_AUTOMATION`.

### `bedrock_data_automation_configuration` Block

The `bedrock_data_automation_configuration` configuration block supports the following arguments:

* `parsing_modality` - (Optional, Forces new resource) Whether to enable parsing of multimodal data, including both text and images. Valid value: `MULTIMODAL`.

### `bedrock_foundation_model_configuration` Block

The `bedrock_foundation_model_configuration` configuration block supports the following arguments:

* `model_arn` - (Required, Forces new resource) ARN of the model used to parse documents.
* `parsing_modality` - (Optional, Forces new resource) Whether to enable parsing of multimodal data, including both text and images. Valid values: `MULTIMODAL`.
* `parsing_prompt` - (Optional, Forces new resource) Instructions for interpreting the contents of the document. See [`parsing_prompt` Block](#parsing_prompt-block) for details.

### `parsing_prompt` Block

The `parsing_prompt` configuration block supports the following arguments:

* `parsing_prompt_string` - (Required) Instructions for interpreting the contents of the document.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `data_source_id` -  Unique identifier of the data source.
* `id` -  Identifier of the data source which consists of the data source ID and the knowledge base ID.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
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
