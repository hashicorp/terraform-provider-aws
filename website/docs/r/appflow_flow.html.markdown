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
  bucket = "example_source"
}

resource "aws_s3_bucket_policy" "example_source" {
  bucket = aws_s3_bucket.example_source.id
  policy = <<EOF
{
    "Statement": [
        {
            "Effect": "Allow",
            "Sid": "AllowAppFlowSourceActions",
            "Principal": {
                "Service": "appflow.amazonaws.com"
            },
            "Action": [
                "s3:ListBucket",
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::example_source",
                "arn:aws:s3:::example_source/*"
            ]
        }
    ],
	"Version": "2012-10-17"
}
EOF
}

resource "aws_s3_object" "example" {
  bucket = aws_s3_bucket.example_source.id
  key    = "example_source.csv"
  source = "example_source.csv"
}

resource "aws_s3_bucket" "example_destination" {
  bucket = "example_destination"
}

resource "aws_s3_bucket_policy" "example_destination" {
  bucket = aws_s3_bucket.example_destination.id
  policy = <<EOF

{
    "Statement": [
        {
            "Effect": "Allow",
            "Sid": "AllowAppFlowDestinationActions",
            "Principal": {
                "Service": "appflow.amazonaws.com"
            },
            "Action": [
                "s3:PutObject",
                "s3:AbortMultipartUpload",
                "s3:ListMultipartUploadParts",
                "s3:ListBucketMultipartUploads",
                "s3:GetBucketAcl",
                "s3:PutObjectAcl"
            ],
            "Resource": [
                "arn:aws:s3:::example_destination",
                "arn:aws:s3:::example_destination/*"
            ]
        }
    ],
	"Version": "2012-10-17"
}
EOF
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

The following arguments are supported:

* `name` - (Required) The name of the flow.
* `destination_flow_config` - (Required) A [Destination Flow Config](#destination-flow-config) that controls how Amazon AppFlow places data in the destination connector.
* `source_flow_config` - (Required) The [Source Flow Config](#source-flow-config) that controls how Amazon AppFlow retrieves data from the source connector.
* `task` - (Required) A [Task](#task) that Amazon AppFlow performs while transferring the data in the flow run.
* `trigger_config` - (Required) A [Trigger](#trigger-config) that determine how and when the flow runs.
* `description` - (Optional) A description of the flow you want to create.
* `kms_arn` - (Optional) The ARN (Amazon Resource Name) of the Key Management Service (KMS) key you provide for encryption. This is required if you do not want to use the Amazon AppFlow-managed KMS key. If you don't provide anything here, Amazon AppFlow uses the Amazon AppFlow-managed KMS key.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

### Destination Flow Config

* `connector_type` - (Required) The type of connector, such as Salesforce, Amplitude, and so on. Valid values are `Salesforce`, `Singular`, `Slack`, `Redshift`, `S3`, `Marketo`, `Googleanalytics`, `Zendesk`, `Servicenow`, `Datadog`, `Trendmicro`, `Snowflake`, `Dynatrace`, `Infornexus`, `Amplitude`, `Veeva`, `EventBridge`, `LookoutMetrics`, `Upsolver`, `Honeycode`, `CustomerProfiles`, `SAPOData`, and `CustomConnector`.
* `destination_connector_properties` - (Required) This stores the information that is required to query a particular connector. See [Destination Connector Properties](#destination-connector-properties) for more information.
* `api_version` - (Optional) The API version that the destination connector uses.
* `connector_profile_name` - (Optional) The name of the connector profile. This name must be unique for each connector profile in the AWS account.

#### Destination Connector Properties

* `custom_connector` - (Optional) The properties that are required to query the custom Connector. See [Custom Connector Destination Properties](#custom-connector-destination-properties) for more details.
* `customer_profiles` - (Optional) The properties that are required to query Amazon Connect Customer Profiles. See [Customer Profiles Destination Properties](#customer-profiles-destination-properties) for more details.
* `event_bridge` - (Optional) The properties that are required to query Amazon EventBridge. See [Generic Destination Properties](#generic-destination-properties) for more details.
* `honeycode` - (Optional) The properties that are required to query Amazon Honeycode. See [Generic Destination Properties](#generic-destination-properties) for more details.
* `marketo` - (Optional) The properties that are required to query Marketo. See [Generic Destination Properties](#generic-destination-properties) for more details.
* `redshift` - (Optional) The properties that are required to query Amazon Redshift. See [Redshift Destination Properties](#redshift-destination-properties) for more details.
* `s3` - (Optional) The properties that are required to query Amazon S3. See [S3 Destination Properties](#s3-destination-properties) for more details.
* `salesforce` - (Optional) The properties that are required to query Salesforce. See [Salesforce Destination Properties](#salesforce-destination-properties) for more details.
* `sapo_data` - (Optional) The properties that are required to query SAPOData. See [SAPOData Destination Properties](#sapodata-destination-properties) for more details.
* `snowflake` - (Optional) The properties that are required to query Snowflake. See [Snowflake Destination Properties](#snowflake-destination-properties) for more details.
* `upsolver` - (Optional) The properties that are required to query Upsolver. See [Upsolver Destination Properties](#upsolver-destination-properties) for more details.
* `zendesk` - (Optional) The properties that are required to query Zendesk. See [Zendesk Destination Properties](#zendesk-destination-properties) for more details.

##### Generic Destination Properties

EventBridge, Honeycode, and Marketo destination properties all support the following attributes:

* `object` - (Required) The object specified in the flow destination.
* `error_handling_config` - (Optional) The settings that determine how Amazon AppFlow handles an error when placing data in the destination. See [Error Handling Config](#error-handling-config) for more details.

##### Custom Connector Destination Properties

* `entity_name` - (Required) The entity specified in the custom connector as a destination in the flow.
* `custom_properties` - (Optional) The custom properties that are specific to the connector when it's used as a destination in the flow. Maximum of 50 items.
* `error_handling_config` - (Optional) The settings that determine how Amazon AppFlow handles an error when placing data in the custom connector as destination. See [Error Handling Config](#error-handling-config) for more details.
* `id_field_names` - (Optional) The name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update, delete, or upsert.
* `write_operation_type` - (Optional) Specifies the type of write operation to be performed in the custom connector when it's used as destination. Valid values are `INSERT`, `UPSERT`, `UPDATE`, and `DELETE`.

##### Customer Profiles Destination Properties

* `domain_name` - (Required) The unique name of the Amazon Connect Customer Profiles domain.
* `object_type_name` - (Optional) The object specified in the Amazon Connect Customer Profiles flow destination.

##### Redshift Destination Properties

* `intermediate_bucket_name` - (Required) The intermediate bucket that Amazon AppFlow uses when moving data into Amazon Redshift.
* `object` - (Required) The object specified in the Amazon Redshift flow destination.
* `bucket_prefix` - (Optional) The object key for the bucket in which Amazon AppFlow places the destination files.
* `error_handling_config` - (Optional) The settings that determine how Amazon AppFlow handles an error when placing data in the destination. See [Error Handling Config](#error-handling-config) for more details.

##### S3 Destination Properties

* `bucket_name` - (Required) The Amazon S3 bucket name in which Amazon AppFlow places the transferred data.
* `bucket_prefix` - (Optional) The object key for the bucket in which Amazon AppFlow places the destination files.
* `s3_output_format_config` - (Optional) The configuration that determines how Amazon AppFlow should format the flow output data when Amazon S3 is used as the destination. See [S3 Output Format Config](#s3-output-format-config) for more details.

###### S3 Output Format Config

* `aggregation_config` - (Optional) The aggregation settings that you can use to customize the output format of your flow data. See [Aggregation Config](#aggregation-config) for more details.
* `file_type` - (Optional) Indicates the file type that Amazon AppFlow places in the Amazon S3 bucket. Valid values are `CSV`, `JSON`, and `PARQUET`.
* `prefix_config` - (Optional) Determines the prefix that Amazon AppFlow applies to the folder name in the Amazon S3 bucket. You can name folders according to the flow frequency and date. See [Prefix Config](#prefix-config) for more details.

##### Salesforce Destination Properties

* `object` - (Required) The object specified in the flow destination.
* `error_handling_config` - (Optional) The settings that determine how Amazon AppFlow handles an error when placing data in the destination. See [Error Handling Config](#error-handling-config) for more details.
* `id_field_names` - (Optional) The name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update or delete.
* `write_operation_type` - (Optional) This specifies the type of write operation to be performed in Salesforce. When the value is `UPSERT`, then `id_field_names` is required. Valid values are `INSERT`, `UPSERT`, `UPDATE`, and `DELETE`.

##### SAPOData Destination Properties

* `object_path` - (Required) The object path specified in the SAPOData flow destination.
* `error_handling_config` - (Optional) The settings that determine how Amazon AppFlow handles an error when placing data in the destination. See [Error Handling Config](#error-handling-config) for more details.
* `id_field_names` - (Optional) The name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update or delete.
* `success_response_handling_config` - (Optional) Determines how Amazon AppFlow handles the success response that it gets from the connector after placing data. See [Success Response Handling Config](#success-response-handling-config) for more details.
* `write_operation` - (Optional) The possible write operations in the destination connector. When this value is not provided, this defaults to the `INSERT` operation. Valid values are `INSERT`, `UPSERT`, `UPDATE`, and `DELETE`.

###### Success Response Handling Config

* `bucket_name` - (Optional) The name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) The Amazon S3 bucket prefix.

##### Snowflake Destination Properties

* `intermediate_bucket_name` - (Required) The intermediate bucket that Amazon AppFlow uses when moving data into Amazon Snowflake.
* `object` - (Required) The object specified in the Amazon Snowflake flow destination.
* `bucket_prefix` - (Optional) The object key for the bucket in which Amazon AppFlow places the destination files.
* `error_handling_config` - (Optional) The settings that determine how Amazon AppFlow handles an error when placing data in the destination. See [Error Handling Config](#error-handling-config) for more details.

##### Upsolver Destination Properties

* `bucket_name` - (Required) The Upsolver Amazon S3 bucket name in which Amazon AppFlow places the transferred data. This must begin with `upsolver-appflow`.
* `bucket_prefix` - (Optional) The object key for the Upsolver Amazon S3 Bucket in which Amazon AppFlow places the destination files.
* `s3_output_format_config` - (Optional) The configuration that determines how Amazon AppFlow should format the flow output data when Upsolver is used as the destination. See [Upsolver S3 Output Format Config](#upsolver-s3-output-format-config) for more details.

###### Upsolver S3 Output Format Config

* `aggregation_config` - (Optional) The aggregation settings that you can use to customize the output format of your flow data. See [Aggregation Config](#aggregation-config) for more details.
* `file_type` - (Optional) Indicates the file type that Amazon AppFlow places in the Upsolver Amazon S3 bucket. Valid values are `CSV`, `JSON`, and `PARQUET`.
* `prefix_config` - (Optional) Determines the prefix that Amazon AppFlow applies to the folder name in the Amazon S3 bucket. You can name folders according to the flow frequency and date. See [Prefix Config](#prefix-config) for more details.

###### Aggregation Config

* `aggregation_type` - (Optional) Specifies whether Amazon AppFlow aggregates the flow records into a single file, or leave them unaggregated. Valid values are `None` and `SingleFile`.

###### Prefix Config

* `prefix_format` - (Optional) Determines the level of granularity that's included in the prefix. Valid values are `YEAR`, `MONTH`, `DAY`, `HOUR`, and `MINUTE`.
* `prefix_type` - (Optional) Determines the format of the prefix, and whether it applies to the file name, file path, or both. Valid values are `FILENAME`, `PATH`, and `PATH_AND_FILENAME`.

##### Zendesk Destination Properties

* `object` - (Required) The object specified in the flow destination.
* `error_handling_config` - (Optional) The settings that determine how Amazon AppFlow handles an error when placing data in the destination. See [Error Handling Config](#error-handling-config) for more details.
* `id_field_names` - (Optional) The name of the field that Amazon AppFlow uses as an ID when performing a write operation such as update or delete.
* `write_operation_type` - (Optional) This specifies the type of write operation to be performed in Zendesk. When the value is `UPSERT`, then `id_field_names` is required. Valid values are `INSERT`, `UPSERT`, `UPDATE`, and `DELETE`.

###### Error Handling Config

* `bucket_name` - (Optional) Specifies the name of the Amazon S3 bucket.
* `bucket_prefix` - (Optional) Specifies the Amazon S3 bucket prefix.
* `fail_on_first_destination_error` - (Optional, boolean) Specifies if the flow should fail after the first instance of a failure when attempting to place data in the destination.

### Source Flow Config

* `connector_type` - (Required) The type of connector, such as Salesforce, Amplitude, and so on. Valid values are `Salesforce`, `Singular`, `Slack`, `Redshift`, `S3`, `Marketo`, `Googleanalytics`, `Zendesk`, `Servicenow`, `Datadog`, `Trendmicro`, `Snowflake`, `Dynatrace`, `Infornexus`, `Amplitude`, `Veeva`, `EventBridge`, `LookoutMetrics`, `Upsolver`, `Honeycode`, `CustomerProfiles`, `SAPOData`, and `CustomConnector`.
* `source_connector_properties` - (Required) Specifies the information that is required to query a particular source connector. See [Source Connector Properties](#source-connector-properties) for details.
* `api_version` - (Optional) The API version that the destination connector uses.
* `connector_profile_name` - (Optional) The name of the connector profile. This name must be unique for each connector profile in the AWS account.
* `incremental_pull_config` - (Optional) Defines the configuration for a scheduled incremental data pull. If a valid configuration is provided, the fields specified in the configuration are used when querying for the incremental data pull. See [Incremental Pull Config](#incremental-pull-config) for more details.

#### Source Connector Properties

* `amplitude` - (Optional) Specifies the information that is required for querying Amplitude. See [Generic Source Properties](#generic-source-properties) for more details.
* `custom_connector` - (Optional) The properties that are applied when the custom connector is being used as a source. See [Custom Connector Source Properties](#custom-connector-source-properties).
* `datadog` - (Optional) Specifies the information that is required for querying Datadog. See [Generic Source Properties](#generic-source-properties) for more details.
* `dynratrace` - (Optional) Specifies the information that is required for querying Dynatrace. See [Generic Source Properties](#generic-source-properties) for more details.
* `infor_nexus` - (Optional) Specifies the information that is required for querying Infor Nexus. See [Generic Source Properties](#generic-source-properties) for more details.
* `marketo` - (Optional) Specifies the information that is required for querying Marketo. See [Generic Source Properties](#generic-source-properties) for more details.
* `s3` - (Optional) Specifies the information that is required for querying Amazon S3. See [S3 Source Properties](#s3-source-properties) for more details.
* `salesforce` - (Optional) Specifies the information that is required for querying Salesforce. See [Salesforce Source Properties](#s3-source-properties) for more details.
* `sapo_data` - (Optional) Specifies the information that is required for querying SAPOData as a flow source. See [SAPO Source Properties](#sapodata-source-properties) for more details.
* `service_now` - (Optional) Specifies the information that is required for querying ServiceNow. See [Generic Source Properties](#generic-source-properties) for more details.
* `singular` - (Optional) Specifies the information that is required for querying Singular. See [Generic Source Properties](#generic-source-properties) for more details.
* `slack` - (Optional) Specifies the information that is required for querying Slack. See [Generic Source Properties](#generic-source-properties) for more details.
* `trend_micro` - (Optional) Specifies the information that is required for querying Trend Micro. See [Generic Source Properties](#generic-source-properties) for more details.
* `veeva` - (Optional) Specifies the information that is required for querying Veeva. See [Veeva Source Properties](#veeva-source-properties) for more details.
* `zendesk` - (Optional) Specifies the information that is required for querying Zendesk. See [Generic Source Properties](#generic-source-properties) for more details.

##### Generic Source Properties

Amplitude, Datadog, Dynatrace, Google Analytics, Infor Nexus, Marketo, ServiceNow, Singular, Slack, Trend Micro, and Zendesk source properties all support the following attributes:

* `object` - (Required) The object specified in the flow source.

##### Custom Connector Source Properties

* `entity_name` - (Required) The entity specified in the custom connector as a source in the flow.
* `custom_properties` - (Optional) The custom properties that are specific to the connector when it's used as a source in the flow. Maximum of 50 items.

##### S3 Source Properties

* `bucket_name` - (Required) The Amazon S3 bucket name where the source files are stored.
* `bucket_prefix` - (Optional) The object key for the Amazon S3 bucket in which the source files are stored.
* `s3_input_format_config` - (Optional) When you use Amazon S3 as the source, the configuration format that you provide the flow input data. See [S3 Input Format Config](#s3-input-format-config) for details.

###### S3 Input Format Config

* `s3_input_file_type` - (Optional) The file type that Amazon AppFlow gets from your Amazon S3 bucket. Valid values are `CSV` and `JSON`.

##### Salesforce Source Properties

* `object` - (Required) The object specified in the Salesforce flow source.
* `enable_dynamic_field_update` - (Optional, boolean) The flag that enables dynamic fetching of new (recently added) fields in the Salesforce objects while running a flow.
* `include_deleted_records` - (Optional, boolean) Indicates whether Amazon AppFlow includes deleted files in the flow run.

##### SAPOData Source Properties

* `object_path` - (Optional) The object path specified in the SAPOData flow source.

##### Veeva Source Properties

* `object` - (Required) The object specified in the Veeva flow source.
* `document_type` - (Optional) The document type specified in the Veeva document extract flow.
* `include_all_versions` - (Optional, boolean) Boolean value to include All Versions of files in Veeva document extract flow.
* `include_renditions` - (Optional, boolean) Boolean value to include file renditions in Veeva document extract flow.
* `include_source_files` - (Optional, boolean) Boolean value to include source files in Veeva document extract flow.

#### Incremental Pull Config

* `datetime_type_field_name` - (Optional) A field that specifies the date time or timestamp field as the criteria to use when importing incremental records from the source.

### Task

* `source_fields` - (Required) The source fields to which a particular task is applied.
* `task_type` - (Required) Specifies the particular task implementation that Amazon AppFlow performs. Valid values are `Arithmetic`, `Filter`, `Map`, `Map_all`, `Mask`, `Merge`, `Passthrough`, `Truncate`, and `Validate`.
* `connector_operator` - (Optional) The operation to be performed on the provided source fields. See [Connector Operator](#connector-operator) for details.
* `destination_field` - (Optional) A field in a destination connector, or a field value against which Amazon AppFlow validates a source field.
* `task_properties` - (Optional) A map used to store task-related information. The execution service looks for particular information based on the `TaskType`. Valid keys are `VALUE`, `VALUES`, `DATA_TYPE`, `UPPER_BOUND`, `LOWER_BOUND`, `SOURCE_DATA_TYPE`, `DESTINATION_DATA_TYPE`, `VALIDATION_ACTION`, `MASK_VALUE`, `MASK_LENGTH`, `TRUNCATE_LENGTH`, `MATH_OPERATION_FIELDS_ORDER`, `CONCAT_FORMAT`, `SUBFIELD_CATEGORY_MAP`, and `EXCLUDE_SOURCE_FIELDS_LIST`.

#### Connector Operator

* `amplitude` - (Optional) The operation to be performed on the provided Amplitude source fields. The only valid value is `BETWEEN`.
* `custom_connector` - (Optional) Operators supported by the custom connector. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `datadog` - (Optional) The operation to be performed on the provided Datadog source fields. Valid values are `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `dynatrace` - (Optional) The operation to be performed on the provided Dynatrace source fields. Valid values are `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `google_analytics` - (Optional) The operation to be performed on the provided Google Analytics source fields. Valid values are `PROJECTION` and `BETWEEN`.
* `infor_nexus` - (Optional) The operation to be performed on the provided Infor Nexus source fields. Valid values are `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `marketo` - (Optional) The operation to be performed on the provided Marketo source fields. Valid values are `PROJECTION`, `BETWEEN`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `s3` - (Optional) The operation to be performed on the provided Amazon S3 source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `salesforce` - (Optional) The operation to be performed on the provided Salesforce source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `sapo_data` - (Optional) The operation to be performed on the provided SAPOData source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `service_now` - (Optional) The operation to be performed on the provided ServiceNow source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `singular` - (Optional) The operation to be performed on the provided Singular source fields. Valid values are `PROJECTION`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `slack` - (Optional) The operation to be performed on the provided Slack source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `trendmicro` - (Optional) The operation to be performed on the provided Trend Micro source fields. Valid values are `PROJECTION`, `EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `veeva` - (Optional) The operation to be performed on the provided Veeva source fields. Valid values are `PROJECTION`, `LESS_THAN`, `GREATER_THAN`, `CONTAINS`, `BETWEEN`, `LESS_THAN_OR_EQUAL_TO`, `GREATER_THAN_OR_EQUAL_TO`, `EQUAL_TO`, `NOT_EQUAL_TO`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.
* `zendesk` - (Optional) The operation to be performed on the provided Zendesk source fields. Valid values are `PROJECTION`, `GREATER_THAN`, `ADDITION`, `MULTIPLICATION`, `DIVISION`, `SUBTRACTION`, `MASK_ALL`, `MASK_FIRST_N`, `MASK_LAST_N`, `VALIDATE_NON_NULL`, `VALIDATE_NON_ZERO`, `VALIDATE_NON_NEGATIVE`, `VALIDATE_NUMERIC`, and `NO_OP`.

### Trigger Config

* `trigger_type` - (Required) Specifies the type of flow trigger. Valid values are `Scheduled`, `Event`, and `OnDemand`.
* `trigger_properties` - (Optional) Specifies the configuration details of a schedule-triggered flow as defined by the user. Currently, these settings only apply to the `Scheduled` trigger type. See [Scheduled Trigger Properties](#scheduled-trigger-properties) for details.

#### Scheduled Trigger Properties

The `trigger_properties` block only supports one attribute: `scheduled`, a block which in turn supports the following:

* `schedule_expression` - (Required) The scheduling expression that determines the rate at which the schedule will run, for example `rate(5minutes)`.
* `data_pull_mode` - (Optional) Specifies whether a scheduled flow has an incremental data transfer or a complete data transfer for each flow run. Valid values are `Incremental` and `Complete`.
* `first_execution_from` - (Optional) Specifies the date range for the records to import from the connector in the first flow run. Must be a valid RFC3339 timestamp.
* `schedule_end_time` - (Optional) Specifies the scheduled end time for a schedule-triggered flow. Must be a valid RFC3339 timestamp.
* `schedule_offset` - (Optional) Specifies the optional offset that is added to the time interval for a schedule-triggered flow. Maximum value of 36000.
* `schedule_start_time` - (Optional) Specifies the scheduled start time for a schedule-triggered flow. Must be a valid RFC3339 timestamp.
* `timezone` - (Optional) Specifies the time zone used when referring to the date and time of a scheduled-triggered flow, such as `America/New_York`.

```terraform
resource "aws_appflow_flow" "example" {
  # ... other configuration ...

  trigger_config {
    scheduled {
      schedule_expression = "rate(1minutes)"
    }
  }
}
```

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The flow's Amazon Resource Name (ARN).

## Import

AppFlow flows can be imported using the `arn`, e.g.:

```
$ terraform import aws_appflow_flow.example arn:aws:appflow:us-west-2:123456789012:flow/example-flow
```
