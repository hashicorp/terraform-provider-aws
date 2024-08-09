---
subcategory: "Amazon Q Business"
layout: "aws"
page_title: "AWS: aws_qbusiness_datasource"
description: |-
  Provides a Q Business Datasource resource.
---

# Resource: aws_qbusiness_datasource

Provides a Q Business Datasource resource.

## Example Usage

```terraform
resource "aws_qbusiness_datasource" "example" {
  application_id       = aws_qbusiness_app.test.application_id
  index_id             = aws_qbusiness_index.test.index_id
  display_name         = "Datasource"
  iam_service_role_arn = aws_iam_role.test.arn
  configuration = jsonencode({
    type = "S3"
    connectionConfiguration = {
      repositoryEndpointMetadata = {
        BucketName = aws_s3_bucket.test.bucket
      }
    }
    syncMode = "FULL_CRAWL"
    repositoryConfigurations = {
      document = {
        fieldMappings = []
      }
    }
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) Identifier of the Amazon Q application the data source will be attached to.
* `configuration` - (Required) Configuration information (JSON) to connect to your data source repository.
* `description` - (Optional) Description for the data source connector.
* `display_name` - (Required) Name for the data source connector.
* `document_enrichment_configuration` - (Optional) Configuration information for altering document metadata and content during the document ingestion process.
* `iam_service_role_arn` - (Required) ARN of an IAM role with permission to access the data source and required resources.
* `index_id` - (Required) Identifier of the index that you want to use with the data source connector.
* `sync_schedule` - (Optional) Frequency for Amazon Q to check the documents in your data source repository and update your index. In `cron` format.
* `vpc_config` - (Optional) Information for an VPC to connect to your data source.

`document_enrichment_configuration` supports the following:

* `inline_configuration` - (Optional) Information to alter document attributes or metadata fields and content when ingesting documents into Amazon Q.
* `post_extraction_hook_configuration` - (Optional) Provides the configuration information for invoking a Lambda function in AWS Lambda to alter document metadata and content when ingesting documents into Amazon Q.
* `pre_extraction_hook_configuration` - (Optional) Provides the configuration information for invoking a Lambda function in AWS Lambda to alter document metadata and content when ingesting documents into Amazon Q.

`inline_configuration` supports the following:

* `condition` - (Optional) The condition used for the target document attribute or metadata field when ingesting documents into Amazon Q. You use this with DocumentAttributeTarget to apply the condition.
* `document_content_operator` - (Optional) `DELETE` to delete content if the condition used for the target attribute is met.
* `target` - (Optional) The target document attribute or metadata field you want to alter when ingesting documents into Amazon Q.

`condition` supports the following:

* `key` - (Required) The identifier of the document attribute used for the condition.
* `operator` - (Required) The identifier of the document attribute used for the condition. Valid Values: `GREATER_THAN`, `GREATER_THAN_OR_EQUALS`, `LESS_THAN`, `LESS_THAN_OR_EQUALS`, `EQUALS`, `NOT_EQUALS`, `CONTAINS`, `NOT_CONTAINS`, `EXISTS`, `NOT_EXISTS`, `BEGINS_WITH`
* `value` - (Optional) The value of a document attribute. You can only provide one value for a document attribute.

`target` supports the following:

* `key` - (Required) The identifier of the target document attribute or metadata field. For example, 'Department' could be an identifier for the target attribute or metadata field that includes the department names associated with the documents.
* `attribute_value_operator` - (Optional) `DELETE` to delete content if the condition used for the target attribute is met.
* `value` - (Optional) The value of a document attribute. You can only provide one value for a document attribute.

`value` supports the following:

* `date_value` - (Optional) A date expressed as an ISO 8601 string. Must be in UTC timezone
* `long_value` - (Optional) A long integer value.
* `string_list_value` - (Optional) A list of strings.
* `strings_value` - (Optional) A string.

`post_extraction_hook_configuration` supports the following:

* `invocation_condition` - (Optional) The condition used for when a Lambda function should be invoked. Type `condition`
* `lambda_arn` - (Optional) ARN of a role with permission to run a Lambda function during ingestion.
* `role_arn` - (Optional) The Amazon Resource Name (ARN) of a role with permission to run `PreExtractionHookConfiguration` and `PostExtractionHookConfiguration` for altering document metadata and content during the document ingestion process.
* `s3_bucket_name` - (Optional) Stores the original, raw documents or the structured, parsed documents before and after altering them.

`pre_extraction_hook_configuration` supports the following:

* `invocation_condition` - (Optional) The condition used for when a Lambda function should be invoked. Type `condition`
* `lambda_arn` - (Optional) ARN of a role with permission to run a Lambda function during ingestion.
* `role_arn` - (Optional) The Amazon Resource Name (ARN) of a role with permission to run `PreExtractionHookConfiguration` and `PostExtractionHookConfiguration` for altering document metadata and content during the document ingestion process.
* `s3_bucket_name` - (Optional) Stores the original, raw documents or the structured, parsed documents before and after altering them.

`vpc_config` supports the following

* `vpc_security_group_ids` - (Required) List of security group ids.
* `subnet_ids` - (Required) List of subnet ids.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `datasource_id` - Datasource identifier.
* `arn` - ARN of the Amazon Q datasource.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
