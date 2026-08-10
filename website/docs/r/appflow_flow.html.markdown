---
subcategory: "AppFlow"
layout: "aws"
page_title: "AWS: aws_appflow_flow"
description: |-
  Provides an AppFlow Flow resource.
---

# Resource: aws_appflow_flow

Provides an AppFlow flow resource.

## Example Usage

```terraform
resource "aws_s3_bucket" "example_source" {
  bucket = "example-source"
}

data "aws_iam_policy_document" "example_source" {
  statement {
    sid    = "AllowAppFlowSourceActions"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["appflow.amazonaws.com"]
    }

    actions = [
      "s3:ListBucket",
      "s3:GetObject",
    ]

    resources = [
      "arn:aws:s3:::example-source",
      "arn:aws:s3:::example-source/*",
    ]
  }
}

resource "aws_s3_bucket_policy" "example_source" {
  bucket = aws_s3_bucket.example_source.id
  policy = data.aws_iam_policy_document.example_source.json
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket.example_source.id
  key    = "example_source.csv"
  source = "example_source.csv"
}

resource "aws_s3_bucket" "example_destination" {
  bucket = "example-destination"
}

data "aws_iam_policy_document" "example_destination" {
  statement {
    sid    = "AllowAppFlowDestinationActions"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["appflow.amazonaws.com"]
    }

    actions = [
      "s3:PutObject",
      "s3:AbortMultipartUpload",
      "s3:ListMultipartUploadParts",
      "s3:ListBucketMultipartUploads",
      "s3:GetBucketAcl",
      "s3:PutObjectAcl",
    ]

    resources = [
      "arn:aws:s3:::example-destination",
      "arn:aws:s3:::example-destination/*",
    ]
  }
}

resource "aws_s3_bucket_policy" "example_destination" {
  bucket = aws_s3_bucket.example_destination.id
  policy = data.aws_iam_policy_document.example_destination.json
}

resource "aws_appflow_flow" "example" {
  name = "example"

  source_flow_config {
    connector_type = "S3"
    source_connector_properties {
      s3 {
        bucket_name   = aws_s3_bucket_policy.example_source.bucket
        bucket_prefix = "example"
      }
    }
  }

  destination_flow_config {
    connector_type = "S3"
    destination_connector_properties {
      s3 {
        bucket_name = aws_s3_bucket_policy.example_destination.bucket

        s3_output_format_config {
          prefix_config {
            prefix_type = "PATH"
          }
        }
      }
    }
  }

  task {
    source_fields     = ["exampleField"]
    destination_field = "exampleField"
    task_type         = "Map"

    connector_operator {
      s3 = "NO_OP"
    }
  }

  trigger_config {
    trigger_type = "OnDemand"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) Description of the flow.
* `destination_flow_config` - (Required) Configuration that controls how Amazon AppFlow places data in the destination connector. See the `destination_flow_config` Block for details.
* `kms_arn` - (Optional) ARN of the Key Management Service (KMS) key you provide for encryption. Required if you do not want to use the Amazon AppFlow-managed KMS key. Uses the Amazon AppFlow-managed KMS key when not provided.
* `metadata_catalog_config` - (Optional) Configuration that determines how Amazon AppFlow catalogs the data that the flow transfers. See the `metadata_catalog_config` Block for details.
* `name` - (Required) Name of the flow.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_flow_config` - (Required) Configuration that controls how Amazon AppFlow retrieves data from the source connector. See the `source_flow_config` Block for details.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `task` - (Required) Tasks that Amazon AppFlow performs while transferring the data in the flow run. See the `task` Block for details.
* `trigger_config` - (Required) Configuration that determines how and when the flow runs. See the `trigger_config` Block for details.

### `destination_flow_config` Block

* `api_version` - (Optional) API version that the destination connector uses.
* `connector_profile_name` - (Optional) Name of the connector profile. Must be unique for each connector profile in the AWS account.
* `connector_type` - (Required) Type of connector, such as Salesforce, Amplitude, and so on. Valid values are `Salesforce`, `Singular`, `Slack`, `Redshift`, `S3`, `Marketo`, `Googleanalytics`, `Zendesk`, `Servicenow`, `Datadog`, `Trendmicro`, `Snowflake`, `Dynatrace`, `Infornexus`, `Amplitude`, `Veeva`, `EventBridge`, `LookoutMetrics`, `Upsolver`, `Honeycode`, `CustomerProfiles`, `SAPOData`, and `CustomConnector`.
* `destination_connector_properties` - (Required) Information required to query a particular connector. See the `destination_flow_config.destination_connector_properties` Block for details.

### `destination_flow_config.destination_connector_properties` Block

* `custom_connector` - (Optional) Properties required to query the custom connector. See the `destination_flow_config.destination_connector_properties.custom_connector` Block for details.
* `customer_profiles` - (Optional) Properties required to query Amazon Connect Customer Profiles. See the `destination_flow_config.destination_connector_properties.customer_profiles` Block for details.
* `event_bridge` - (Optional) Properties required to query Amazon EventBridge. See the `destination_flow_config.destination_connector_properties.event_bridge` Block for details.
* `honeycode` - (Optional) Properties required to query Amazon Honeycode. See the `destination_flow_config.destination_connector_properties.honeycode` Block for details.
* `marketo` - (Optional) Properties required to query Marketo. See the `destination_flow_config.destination_connector_properties.marketo` Block for details.
* `redshift` - (Optional) Properties required to query Amazon Redshift. See the `destination_flow_config.destination_connector_properties.redshift` Block for details.
* `s3` - (Optional) Properties required to query Amazon S3. See the `destination_flow_config.destination_connector_properties.s3` Block for details.
* `salesforce` - (Optional) Properties required to query Salesforce. See the `destination_flow_config.destination_connector_properties.salesforce` Block for details.
* `sapo_data` - (Optional) Properties required to query SAPOData. See the `destination_flow_config.destination_connector_properties.sapo_data` Block for details.
* `snowflake` - (Optional) Properties required to query Snowflake. See the `destination_flow_config.destination_connector_properties.snowflake` Block for details.
* `upsolver` - (Optional) Properties required to query Upsolver. See the `destination_flow_config.destination_connector_properties.upsolver` Block for details.
* `zendesk` - (Optional) Properties required to query Zendesk. See the `destination_flow_config.destination_connector_properties.zendesk` Block for details.

### `destination_flow_config.destination_connector_properties.custom_connector` Block

* `custom_properties` - (Optional) Custom properties specific to the connector when it's used as a destination in the flow. Maximum of 50 items.
* `entity_name` - (Required) Entity specified in the custom connector as a destination in the flow.
* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the custom connector as destination. See the `destination_flow_config.destination_connector_properties.custom_connector.error_handling_config` Block for details.
* `id_field_names` - (Optional) Name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update, delete, or upsert.
* `write_operation_type` - (Optional) Type of write operation to be performed in the custom connector when it's used as destination. Valid values are `INSERT`, `UPSERT`, `UPDATE`, and `DELETE`.

### `destination_flow_config.destination_connector_properties.custom_connector.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `destination_flow_config.destination_connector_properties.customer_profiles` Block

* `domain_name` - (Required) Unique name of the Amazon Connect Customer Profiles domain.
* `object_type_name` - (Optional) Object specified in the Amazon Connect Customer Profiles flow destination.

### `destination_flow_config.destination_connector_properties.event_bridge` Block

* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the destination. See the `destination_flow_config.destination_connector_properties.event_bridge.error_handling_config` Block for details.
* `object` - (Required) Object specified in the flow destination.

### `destination_flow_config.destination_connector_properties.event_bridge.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `destination_flow_config.destination_connector_properties.honeycode` Block

* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the destination. See the `destination_flow_config.destination_connector_properties.honeycode.error_handling_config` Block for details.
* `object` - (Required) Object specified in the flow destination.

### `destination_flow_config.destination_connector_properties.honeycode.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `destination_flow_config.destination_connector_properties.marketo` Block

* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the destination. See the `destination_flow_config.destination_connector_properties.marketo.error_handling_config` Block for details.
* `object` - (Required) Object specified in the flow destination.

### `destination_flow_config.destination_connector_properties.marketo.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `destination_flow_config.destination_connector_properties.redshift` Block

* `bucket_prefix` - (Optional) Object key for the bucket in which Amazon AppFlow places the destination files.
* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the destination. See the `destination_flow_config.destination_connector_properties.redshift.error_handling_config` Block for details.
* `intermediate_bucket_name` - (Required) Intermediate bucket that Amazon AppFlow uses when moving data into Amazon Redshift.
* `object` - (Required) Object specified in the Amazon Redshift flow destination.

### `destination_flow_config.destination_connector_properties.redshift.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `destination_flow_config.destination_connector_properties.s3` Block

* `bucket_name` - (Required) Amazon S3 bucket name in which Amazon AppFlow places the transferred data.
* `bucket_prefix` - (Optional) Object key for the bucket in which Amazon AppFlow places the destination files.
* `s3_output_format_config` - (Optional) Configuration that determines how Amazon AppFlow formats the flow output data when Amazon S3 is used as the destination. See the `destination_flow_config.destination_connector_properties.s3.s3_output_format_config` Block for details.

### `destination_flow_config.destination_connector_properties.s3.s3_output_format_config` Block

* `aggregation_config` - (Optional) Aggregation settings that you can use to customize the output format of your flow data. See the `destination_flow_config.destination_connector_properties.s3.s3_output_format_config.aggregation_config` Block for details.
* `file_type` - (Optional) File type that Amazon AppFlow places in the Amazon S3 bucket. Valid values are `CSV`, `JSON`, and `PARQUET`.
* `prefix_config` - (Optional) Prefix that Amazon AppFlow applies to the folder name in the Amazon S3 bucket. See the `destination_flow_config.destination_connector_properties.s3.s3_output_format_config.prefix_config` Block for details.
* `preserve_source_data_typing` - (Optional, Boolean) Whether to preserve the data types from the source system. Only valid for the `PARQUET` file type.

### `destination_flow_config.destination_connector_properties.s3.s3_output_format_config.aggregation_config` Block

* `aggregation_type` - (Optional) Whether Amazon AppFlow aggregates the flow records into a single file, or leaves them unaggregated. Valid values are `None` and `SingleFile`.
* `target_file_size` - (Optional) Desired file size, in MB, for each output file that Amazon AppFlow writes to the flow destination.

### `destination_flow_config.destination_connector_properties.s3.s3_output_format_config.prefix_config` Block

* `prefix_format` - (Optional) Level of granularity included in the prefix. Valid values are `YEAR`, `MONTH`, `DAY`, `HOUR`, and `MINUTE`.
* `prefix_hierarchy` - (Optional) Determines whether the destination file path includes either or both of the selected elements. Valid values are `EXECUTION_ID` and `SCHEMA_VERSION`.
* `prefix_type` - (Optional) Format of the prefix, and whether it applies to the file name, file path, or both. Valid values are `FILENAME`, `PATH`, and `PATH_AND_FILENAME`.

### `destination_flow_config.destination_connector_properties.salesforce` Block

* `data_transfer_api` - (Optional) Salesforce API used by Amazon AppFlow when the flow transfers data to Salesforce.
* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the destination. See the `destination_flow_config.destination_connector_properties.salesforce.error_handling_config` Block for details.
* `id_field_names` - (Optional) Name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update or delete.
* `object` - (Required) Object specified in the flow destination.
* `write_operation_type` - (Optional) Type of write operation to be performed in Salesforce. When the value is `UPSERT`, `id_field_names` is required. Valid values are `INSERT`, `UPSERT`, `UPDATE`, and `DELETE`.

### `destination_flow_config.destination_connector_properties.salesforce.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `destination_flow_config.destination_connector_properties.sapo_data` Block

* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the destination. See the `destination_flow_config.destination_connector_properties.sapo_data.error_handling_config` Block for details.
* `id_field_names` - (Optional) Name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update or delete.
* `object_path` - (Required) Object path specified in the SAPOData flow destination.
* `success_response_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles the success response it gets from the connector after placing data. See the `destination_flow_config.destination_connector_properties.sapo_data.success_response_handling_config` Block for details.
* `write_operation_type` - (Optional) Possible write operations in the destination connector. Defaults to `INSERT` when not provided. Valid values are `INSERT`, `UPSERT`, `UPDATE`, and `DELETE`.

### `destination_flow_config.destination_connector_properties.sapo_data.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `destination_flow_config.destination_connector_properties.sapo_data.success_response_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.

### `destination_flow_config.destination_connector_properties.snowflake` Block

* `bucket_prefix` - (Optional) Object key for the bucket in which Amazon AppFlow places the destination files.
* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the destination. See the `destination_flow_config.destination_connector_properties.snowflake.error_handling_config` Block for details.
* `intermediate_bucket_name` - (Required) Intermediate bucket that Amazon AppFlow uses when moving data into Amazon Snowflake.
* `object` - (Required) Object specified in the Amazon Snowflake flow destination.

### `destination_flow_config.destination_connector_properties.snowflake.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `destination_flow_config.destination_connector_properties.upsolver` Block

* `bucket_name` - (Required) Upsolver Amazon S3 bucket name in which Amazon AppFlow places the transferred data. Must begin with `upsolver-appflow`.
* `bucket_prefix` - (Optional) Object key for the Upsolver Amazon S3 bucket in which Amazon AppFlow places the destination files.
* `s3_output_format_config` - (Required) Configuration that determines how Amazon AppFlow formats the flow output data when Upsolver is used as the destination. See the `destination_flow_config.destination_connector_properties.upsolver.s3_output_format_config` Block for details.

### `destination_flow_config.destination_connector_properties.upsolver.s3_output_format_config` Block

* `aggregation_config` - (Optional) Aggregation settings that you can use to customize the output format of your flow data. See the `destination_flow_config.destination_connector_properties.upsolver.s3_output_format_config.aggregation_config` Block for details.
* `file_type` - (Optional) File type that Amazon AppFlow places in the Upsolver Amazon S3 bucket. Valid values are `CSV`, `JSON`, and `PARQUET`.
* `prefix_config` - (Required) Prefix that Amazon AppFlow applies to the folder name in the Amazon S3 bucket. See the `destination_flow_config.destination_connector_properties.upsolver.s3_output_format_config.prefix_config` Block for details.

### `destination_flow_config.destination_connector_properties.upsolver.s3_output_format_config.aggregation_config` Block

* `aggregation_type` - (Optional) Whether Amazon AppFlow aggregates the flow records into a single file, or leaves them unaggregated. Valid values are `None` and `SingleFile`.

### `destination_flow_config.destination_connector_properties.upsolver.s3_output_format_config.prefix_config` Block

* `prefix_format` - (Optional) Level of granularity included in the prefix. Valid values are `YEAR`, `MONTH`, `DAY`, `HOUR`, and `MINUTE`.
* `prefix_hierarchy` - (Optional) Determines whether the destination file path includes either or both of the selected elements. Valid values are `EXECUTION_ID` and `SCHEMA_VERSION`.
* `prefix_type` - (Required) Format of the prefix, and whether it applies to the file name, file path, or both. Valid values are `FILENAME`, `PATH`, and `PATH_AND_FILENAME`.

### `destination_flow_config.destination_connector_properties.zendesk` Block

* `error_handling_config` - (Optional) Settings that determine how Amazon AppFlow handles an error when placing data in the destination. See the `destination_flow_config.destination_connector_properties.zendesk.error_handling_config` Block for details.
* `id_field_names` - (Optional) Name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update or delete.
* `object` - (Required) Object specified in the flow destination.
* `write_operation_type` - (Optional) Type of write operation to be performed in Zendesk. When the value is `UPSERT`, `id_field_names` is required. Valid values are `INSERT`, `UPSERT`, `UPDATE`, and `DELETE`.

### `destination_flow_config.destination_connector_properties.zendesk.error_handling_config` Block

* `bucket_name` - (Optional) Name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, Boolean) Whether to fail the flow after the first instance of a failure when attempting to place data in the destination.

### `source_flow_config` Block

* `api_version` - (Optional) API version that the source connector uses.
* `connector_profile_name` - (Optional) Name of the connector profile. Must be unique for each connector profile in the AWS account.
* `connector_type` - (Required) Type of connector, such as Salesforce, Amplitude, and so on. Valid values are `Salesforce`, `Singular`, `Slack`, `Redshift`, `S3`, `Marketo`, `Googleanalytics`, `Zendesk`, `Servicenow`, `Datadog`, `Trendmicro`, `Snowflake`, `Dynatrace`, `Infornexus`, `Amplitude`, `Veeva`, `EventBridge`, `LookoutMetrics`, `Upsolver`, `Honeycode`, `CustomerProfiles`, `SAPOData`, and `CustomConnector`.
* `incremental_pull_config` - (Optional) Configuration for a scheduled incremental data pull. When a valid configuration is provided, the specified fields are used when querying for the incremental data pull. See the `source_flow_config.incremental_pull_config` Block for details.
* `source_connector_properties` - (Required) Information required to query a particular source connector. See the `source_flow_config.source_connector_properties` Block for details.

### `source_flow_config.incremental_pull_config` Block

* `datetime_type_field_name` - (Optional) Field that specifies the date time or timestamp field as the criteria to use when importing incremental records from the source.

### `source_flow_config.source_connector_properties` Block

* `amplitude` - (Optional) Information required to query Amplitude. See the `source_flow_config.source_connector_properties.amplitude` Block for details.
* `custom_connector` - (Optional) Properties applied when the custom connector is used as a source. See the `source_flow_config.source_connector_properties.custom_connector` Block for details.
* `datadog` - (Optional) Information required to query Datadog. See the `source_flow_config.source_connector_properties.datadog` Block for details.
* `dynatrace` - (Optional) Information required to query Dynatrace. See the `source_flow_config.source_connector_properties.dynatrace` Block for details.
* `google_analytics` - (Optional) Information required to query Google Analytics. See the `source_flow_config.source_connector_properties.google_analytics` Block for details.
* `infor_nexus` - (Optional) Information required to query Infor Nexus. See the `source_flow_config.source_connector_properties.infor_nexus` Block for details.
* `marketo` - (Optional) Information required to query Marketo. See the `source_flow_config.source_connector_properties.marketo` Block for details.
* `s3` - (Optional) Information required to query Amazon S3. See the `source_flow_config.source_connector_properties.s3` Block for details.
* `salesforce` - (Optional) Information required to query Salesforce. See the `source_flow_config.source_connector_properties.salesforce` Block for details.
* `sapo_data` - (Optional) Information required to query SAPOData as a flow source. See the `source_flow_config.source_connector_properties.sapo_data` Block for details.
* `service_now` - (Optional) Information required to query ServiceNow. See the `source_flow_config.source_connector_properties.service_now` Block for details.
* `singular` - (Optional) Information required to query Singular. See the `source_flow_config.source_connector_properties.singular` Block for details.
* `slack` - (Optional) Information required to query Slack. See the `source_flow_config.source_connector_properties.slack` Block for details.
* `trendmicro` - (Optional) Information required to query Trend Micro. See the `source_flow_config.source_connector_properties.trendmicro` Block for details.
* `veeva` - (Optional) Information required to query Veeva. See the `source_flow_config.source_connector_properties.veeva` Block for details.
* `zendesk` - (Optional) Information required to query Zendesk. See the `source_flow_config.source_connector_properties.zendesk` Block for details.

### `source_flow_config.source_connector_properties.amplitude` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.custom_connector` Block

* `custom_properties` - (Optional) Custom properties specific to the connector when it's used as a source in the flow. Maximum of 50 items.
* `entity_name` - (Required) Entity specified in the custom connector as a source in the flow.

### `source_flow_config.source_connector_properties.datadog` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.dynatrace` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.google_analytics` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.infor_nexus` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.marketo` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.s3` Block

* `bucket_name` - (Required) Amazon S3 bucket name where the source files are stored.
* `bucket_prefix` - (Required) Object key for the Amazon S3 bucket in which the source files are stored.
* `s3_input_format_config` - (Optional) When you use Amazon S3 as the source, configuration format that you provide for the flow input data. See the `source_flow_config.source_connector_properties.s3.s3_input_format_config` Block for details.

### `source_flow_config.source_connector_properties.s3.s3_input_format_config` Block

* `s3_input_file_type` - (Optional) File type that Amazon AppFlow gets from your Amazon S3 bucket. Valid values are `CSV` and `JSON`.

### `source_flow_config.source_connector_properties.salesforce` Block

* `data_transfer_api` - (Optional) Salesforce API used by Amazon AppFlow when the flow transfers data from Salesforce.
* `enable_dynamic_field_update` - (Optional, Boolean) Whether to enable dynamic fetching of new (recently added) fields in the Salesforce objects while running a flow.
* `include_deleted_records` - (Optional, Boolean) Whether to include deleted files in the flow run.
* `object` - (Required) Object specified in the Salesforce flow source.

### `source_flow_config.source_connector_properties.sapo_data` Block

* `object_path` - (Required) Object path specified in the SAPOData flow source.
* `pagination_config` - (Optional) Page size for each concurrent process that transfers OData records from your SAP instance. See the `source_flow_config.source_connector_properties.sapo_data.pagination_config` Block for details.
* `parallelism_config` - (Optional) Number of concurrent processes that transfer OData records from your SAP instance. See the `source_flow_config.source_connector_properties.sapo_data.parallelism_config` Block for details.

### `source_flow_config.source_connector_properties.sapo_data.pagination_config` Block

* `max_page_size` - (Required) Maximum number of records that Amazon AppFlow receives in each page of the response from your SAP application.

### `source_flow_config.source_connector_properties.sapo_data.parallelism_config` Block

* `max_page_size` - (Required) Maximum number of processes that Amazon AppFlow runs at the same time when it retrieves your data from your SAP application.

### `source_flow_config.source_connector_properties.service_now` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.singular` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.slack` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.trendmicro` Block

* `object` - (Required) Object specified in the flow source.

### `source_flow_config.source_connector_properties.veeva` Block

* `document_type` - (Optional) Document type specified in the Veeva document extract flow.
* `include_all_versions` - (Optional, Boolean) Whether to include all versions of files in the Veeva document extract flow.
* `include_renditions` - (Optional, Boolean) Whether to include file renditions in the Veeva document extract flow.
* `include_source_files` - (Optional, Boolean) Whether to include source files in the Veeva document extract flow.
* `object` - (Required) Object specified in the Veeva flow source.

### `source_flow_config.source_connector_properties.zendesk` Block

* `object` - (Required) Object specified in the flow source.

### `task` Block

* `connector_operator` - (Optional) Operation to be performed on the provided source fields. See the `task.connector_operator` Block for details.
* `destination_field` - (Optional) Field in a destination connector, or a field value against which Amazon AppFlow validates a source field.
* `source_fields` - (Optional) Source fields to which a particular task is applied.
* `task_properties` - (Optional) Map used to store task-related information. The execution service looks for particular information based on the `TaskType`. Valid keys are `VALUE`, `VALUES`, `DATA_TYPE`, `UPPER_BOUND`, `LOWER_BOUND`, `SOURCE_DATA_TYPE`, `DESTINATION_DATA_TYPE`, `VALIDATION_ACTION`, `MASK_VALUE`, `MASK_LENGTH`, `TRUNCATE_LENGTH`, `MATH_OPERATION_FIELDS_ORDER`, `CONCAT_FORMAT`, `SUBFIELD_CATEGORY_MAP`, and `EXCLUDE_SOURCE_FIELDS_LIST`.
* `task_type` - (Required) Particular task implementation that Amazon AppFlow performs. Valid values are `Arithmetic`, `Filter`, `Map`, `Map_all`, `Mask`, `Merge`, `Passthrough`, `Truncate`, and `Validate`.

### `task.connector_operator` Block

* `amplitude` - (Optional) Operation to be performed on the provided Amplitude source fields. The only valid value is `BETWEEN`.
* `custom_connector` - (Optional) Operators supported by the custom connector. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `datadog` - (Optional) Operation to be performed on the provided Datadog source fields. Valid values are `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `dynatrace` - (Optional) Operation to be performed on the provided Dynatrace source fields. Valid values are `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `google_analytics` - (Optional) Operation to be performed on the provided Google Analytics source fields. Valid values are `PROJECTION` and `BETWEEN`.
* `infor_nexus` - (Optional) Operation to be performed on the provided Infor Nexus source fields. Valid values are `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `marketo` - (Optional) Operation to be performed on the provided Marketo source fields. Valid values are `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `s3` - (Optional) Operation to be performed on the provided Amazon S3 source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `salesforce` - (Optional) Operation to be performed on the provided Salesforce source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `sapo_data` - (Optional) Operation to be performed on the provided SAPOData source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `service_now` - (Optional) Operation to be performed on the provided ServiceNow source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `singular` - (Optional) Operation to be performed on the provided Singular source fields. Valid values are `PROJECTION`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `slack` - (Optional) Operation to be performed on the provided Slack source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `trendmicro` - (Optional) Operation to be performed on the provided Trend Micro source fields. Valid values are `PROJECTION`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `veeva` - (Optional) Operation to be performed on the provided Veeva source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `zendesk` - (Optional) Operation to be performed on the provided Zendesk source fields. Valid values are `PROJECTION`, `GREATER_THAN`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.

### `trigger_config` Block

* `trigger_properties` - (Optional) Configuration details of a schedule-triggered flow as defined by the user. Currently, these settings only apply to the `Scheduled` trigger type. See the `trigger_config.trigger_properties` Block for details.
* `trigger_type` - (Required) Type of flow trigger. Valid values are `Scheduled`, `Event`, and `OnDemand`.

### `trigger_config.trigger_properties` Block

* `scheduled` - (Optional) Configuration details of a schedule-triggered flow. See the `trigger_config.trigger_properties.scheduled` Block for details.

### `trigger_config.trigger_properties.scheduled` Block

* `data_pull_mode` - (Optional) Whether a scheduled flow has an incremental data transfer or a complete data transfer for each flow run. Valid values are `Incremental` and `Complete`.
* `first_execution_from` - (Optional) Date range for the records to import from the connector in the first flow run. Must be a valid RFC3339 timestamp.
* `schedule_end_time` - (Optional) Scheduled end time for a schedule-triggered flow. Must be a valid RFC3339 timestamp.
* `schedule_expression` - (Required) Scheduling expression that determines the rate at which the schedule runs, for example `rate(5minutes)`.
* `schedule_offset` - (Optional) Offset that is added to the time interval for a schedule-triggered flow. Maximum value of 36000.
* `schedule_start_time` - (Optional) Scheduled start time for a schedule-triggered flow. Must be a valid RFC3339 timestamp.
* `timezone` - (Optional) Time zone used when referring to the date and time of a scheduled-triggered flow, such as `America/New_York`.

### `metadata_catalog_config` Block

* `glue_data_catalog` - (Optional) Configuration that determines how Amazon AppFlow catalogs data with the AWS Glue Data Catalog. See the `metadata_catalog_config.glue_data_catalog` Block for details.

### `metadata_catalog_config.glue_data_catalog` Block

* `database_name` - (Required) Name of an existing Glue database to store the metadata tables that Amazon AppFlow creates.
* `role_arn` - (Required) ARN of the IAM role that grants Amazon AppFlow the permissions it needs to create Data Catalog tables, databases, and partitions.
* `table_prefix` - (Required) Naming prefix for each Data Catalog table that Amazon AppFlow creates.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Flow's ARN.
* `flow_status` - Current status of the flow.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_appflow_flow.example
  identity = {
    name = "example-flow"
  }
}

resource "aws_appflow_flow" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the AppFlow flow.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppFlow flows using the `name`. For example:

```terraform
import {
  to = aws_appflow_flow.example
  id = "example-flow"
}
```

Using `terraform import`, import AppFlow flows using the `name`. For example:

```console
% terraform import aws_appflow_flow.example example-flow
```
