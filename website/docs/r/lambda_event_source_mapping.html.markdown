---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_event_source_mapping"
description: |-
  Manages an AWS Lambda Event Source Mapping.
---

# Resource: aws_lambda_event_source_mapping

Manages an AWS Lambda Event Source Mapping. Use this resource to connect Lambda functions to event sources like Kinesis, DynamoDB, SQS, Amazon MQ, and Managed Streaming for Apache Kafka (MSK).

For information about Lambda and how to use it, see [What is AWS Lambda?](http://docs.aws.amazon.com/lambda/latest/dg/welcome.html). For information about event source mappings, see [CreateEventSourceMapping](http://docs.aws.amazon.com/lambda/latest/dg/API_CreateEventSourceMapping.html) in the API docs.

## Example Usage

### DynamoDB Stream

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_dynamodb_table.example.stream_arn
  function_name     = aws_lambda_function.example.arn
  starting_position = "LATEST"

  tags = {
    Name = "dynamodb-stream-mapping"
  }
}
```

### Kinesis Stream

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn                   = aws_kinesis_stream.example.arn
  function_name                      = aws_lambda_function.example.arn
  starting_position                  = "LATEST"
  batch_size                         = 100
  maximum_batching_window_in_seconds = 5
  parallelization_factor             = 2

  destination_config {
    on_failure {
      destination_arn = aws_sqs_queue.dlq.arn
    }
  }
}
```

### SQS Queue

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn = aws_sqs_queue.example.arn
  function_name    = aws_lambda_function.example.arn
  batch_size       = 10

  scaling_config {
    maximum_concurrency = 100
  }
}
```

### SQS with Event Filtering

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn = aws_sqs_queue.example.arn
  function_name    = aws_lambda_function.example.arn

  filter_criteria {
    filter {
      pattern = jsonencode({
        body = {
          Temperature = [{ numeric = [">", 0, "<=", 100] }]
          Location    = ["New York"]
        }
      })
    }
  }
}
```

### Amazon MSK

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_msk_cluster.example.arn
  function_name     = aws_lambda_function.example.arn
  topics            = ["orders", "inventory"]
  starting_position = "TRIM_HORIZON"
  batch_size        = 100

  amazon_managed_kafka_event_source_config {
    consumer_group_id = "lambda-consumer-group"
  }
}
```

### Self-Managed Apache Kafka

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  function_name     = aws_lambda_function.example.arn
  topics            = ["orders"]
  starting_position = "TRIM_HORIZON"

  self_managed_event_source {
    endpoints = {
      KAFKA_BOOTSTRAP_SERVERS = "kafka1.example.com:9092,kafka2.example.com:9092"
    }
  }

  self_managed_kafka_event_source_config {
    consumer_group_id = "lambda-consumer-group"
  }

  source_access_configuration {
    type = "VPC_SUBNET"
    uri  = "subnet:${aws_subnet.example1.id}"
  }

  source_access_configuration {
    type = "VPC_SUBNET"
    uri  = "subnet:${aws_subnet.example2.id}"
  }

  source_access_configuration {
    type = "VPC_SECURITY_GROUP"
    uri  = "security_group:${aws_security_group.example.id}"
  }

  provisioned_poller_config {
    maximum_pollers = 100
    minimum_pollers = 10
  }
}
```

### Amazon MQ (ActiveMQ)

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn = aws_mq_broker.example.arn
  function_name    = aws_lambda_function.example.arn
  queues           = ["orders"]
  batch_size       = 10

  source_access_configuration {
    type = "BASIC_AUTH"
    uri  = aws_secretsmanager_secret_version.example.arn
  }
}
```

### Amazon MQ (RabbitMQ)

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn = aws_mq_broker.example.arn
  function_name    = aws_lambda_function.example.arn
  queues           = ["orders"]
  batch_size       = 1

  source_access_configuration {
    type = "VIRTUAL_HOST"
    uri  = "/production"
  }

  source_access_configuration {
    type = "BASIC_AUTH"
    uri  = aws_secretsmanager_secret_version.example.arn
  }
}
```

### DocumentDB Change Stream

```terraform
resource "aws_lambda_event_source_mapping" "example" {
  event_source_arn  = aws_docdb_cluster.example.arn
  function_name     = aws_lambda_function.example.arn
  starting_position = "LATEST"

  document_db_event_source_config {
    database_name   = "orders"
    collection_name = "transactions"
    full_document   = "UpdateLookup"
  }

  source_access_configuration {
    type = "BASIC_AUTH"
    uri  = aws_secretsmanager_secret_version.example.arn
  }
}
```

## Argument Reference

The following arguments are required:

* `function_name` - (Required) Name or ARN of the Lambda function that will be subscribing to events.

The following arguments are optional:

* `amazon_managed_kafka_event_source_config` - (Optional) Additional configuration block for Amazon Managed Kafka sources. Incompatible with `self_managed_event_source` and `self_managed_kafka_event_source_config`. [See below](#amazon_managed_kafka_event_source_config-configuration-block).
* `batch_size` - (Optional) Largest number of records that Lambda will retrieve from your event source at the time of invocation. Defaults to `100` for DynamoDB, Kinesis, MQ and MSK, `10` for SQS.
* `bisect_batch_on_function_error` - (Optional) Whether to split the batch in two and retry if the function returns an error. Only available for stream sources (DynamoDB and Kinesis). Defaults to `false`.
* `destination_config` - (Optional) Amazon SQS queue, Amazon SNS topic or Amazon S3 bucket (only available for Kafka sources) destination for failed records. Only available for stream sources (DynamoDB and Kinesis) and Kafka sources (Amazon MSK and Self-managed Apache Kafka). [See below](#destination_config-configuration-block).
* `document_db_event_source_config` - (Optional) Configuration settings for a DocumentDB event source. [See below](#document_db_event_source_config-configuration-block).
* `enabled` - (Optional) Whether the mapping is enabled. Defaults to `true`.
* `event_source_arn` - (Optional) Event source ARN - required for Kinesis stream, DynamoDB stream, SQS queue, MQ broker, MSK cluster or DocumentDB change stream. Incompatible with Self Managed Kafka source.
* `filter_criteria` - (Optional) Criteria to use for [event filtering](https://docs.aws.amazon.com/lambda/latest/dg/invocation-eventfiltering.html) Kinesis stream, DynamoDB stream, SQS queue event sources. [See below](#filter_criteria-configuration-block).
* `function_response_types` - (Optional) List of current response type enums applied to the event source mapping for [AWS Lambda checkpointing](https://docs.aws.amazon.com/lambda/latest/dg/with-ddb.html#services-ddb-batchfailurereporting). Only available for SQS and stream sources (DynamoDB and Kinesis). Valid values: `ReportBatchItemFailures`.
* `kms_key_arn` - (Optional) ARN of the Key Management Service (KMS) customer managed key that Lambda uses to encrypt your function's filter criteria.
* `maximum_batching_window_in_seconds` - (Optional) Maximum amount of time to gather records before invoking the function, in seconds (between 0 and 300). Records will continue to buffer until either `maximum_batching_window_in_seconds` expires or `batch_size` has been met. For streaming event sources, defaults to as soon as records are available in the stream. Only available for stream sources (DynamoDB and Kinesis) and SQS standard queues.
* `maximum_record_age_in_seconds` - (Optional) Maximum age of a record that Lambda sends to a function for processing. Only available for stream sources (DynamoDB and Kinesis). Must be either -1 (forever, and the default value) or between 60 and 604800 (inclusive).
* `maximum_retry_attempts` - (Optional) Maximum number of times to retry when the function returns an error. Only available for stream sources (DynamoDB and Kinesis). Minimum and default of -1 (forever), maximum of 10000.
* `metrics_config` - (Optional) CloudWatch metrics configuration of the event source. Only available for stream sources (DynamoDB and Kinesis) and SQS queues. [See below](#metrics_config-configuration-block).
* `parallelization_factor` - (Optional) Number of batches to process from each shard concurrently. Only available for stream sources (DynamoDB and Kinesis). Minimum and default of 1, maximum of 10.
* `provisioned_poller_config` - (Optional) Event poller configuration for the event source. Only valid for Amazon MSK or self-managed Apache Kafka sources. [See below](#provisioned_poller_config-configuration-block).
* `queues` - (Optional) Name of the Amazon MQ broker destination queue to consume. Only available for MQ sources. The list must contain exactly one queue name.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `scaling_config` - (Optional) Scaling configuration of the event source. Only available for SQS queues. [See below](#scaling_config-configuration-block).
* `self_managed_event_source` - (Optional) For Self Managed Kafka sources, the location of the self managed cluster. If set, configuration must also include `source_access_configuration`. [See below](#self_managed_event_source-configuration-block).
* `self_managed_kafka_event_source_config` - (Optional) Additional configuration block for Self Managed Kafka sources. Incompatible with `event_source_arn` and `amazon_managed_kafka_event_source_config`. [See below](#self_managed_kafka_event_source_config-configuration-block).
* `source_access_configuration` - (Optional) For Self Managed Kafka sources, the access configuration for the source. If set, configuration must also include `self_managed_event_source`. [See below](#source_access_configuration-configuration-block).
* `starting_position` - (Optional) Position in the stream where AWS Lambda should start reading. Must be one of `AT_TIMESTAMP` (Kinesis only), `LATEST` or `TRIM_HORIZON` if getting events from Kinesis, DynamoDB, MSK or Self Managed Apache Kafka. Must not be provided if getting events from SQS. More information about these positions can be found in the [AWS DynamoDB Streams API Reference](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_streams_GetShardIterator.html) and [AWS Kinesis API Reference](https://docs.aws.amazon.com/kinesis/latest/APIReference/API_GetShardIterator.html#Kinesis-GetShardIterator-request-ShardIteratorType).
* `starting_position_timestamp` - (Optional) Timestamp in [RFC3339 format](https://tools.ietf.org/html/rfc3339#section-5.8) of the data record which to start reading when using `starting_position` set to `AT_TIMESTAMP`. If a record with this exact timestamp does not exist, the next later record is chosen. If the timestamp is older than the current trim horizon, the oldest available record is chosen.
* `tags` - (Optional) Map of tags to assign to the object. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `topics` - (Optional) Name of the Kafka topics. Only available for MSK sources. A single topic name must be specified.
* `tumbling_window_in_seconds` - (Optional) Duration in seconds of a processing window for [AWS Lambda streaming analytics](https://docs.aws.amazon.com/lambda/latest/dg/with-kinesis.html#services-kinesis-windows). The range is between 1 second up to 900 seconds. Only available for stream sources (DynamoDB and Kinesis).

### amazon_managed_kafka_event_source_config Configuration Block

* `consumer_group_id` - (Optional) Kafka consumer group ID between 1 and 200 characters for use when creating this event source mapping. If one is not specified, this value will be automatically generated. See [AmazonManagedKafkaEventSourceConfig Syntax](https://docs.aws.amazon.com/lambda/latest/dg/API_AmazonManagedKafkaEventSourceConfig.html).

### destination_config Configuration Block

* `on_failure` - (Optional) Destination configuration for failed invocations. [See below](#destination_config-on_failure-configuration-block).

#### destination_config on_failure Configuration Block

* `destination_arn` - (Required) ARN of the destination resource.

### document_db_event_source_config Configuration Block

* `collection_name` - (Optional) Name of the collection to consume within the database. If you do not specify a collection, Lambda consumes all collections.
* `database_name` - (Required) Name of the database to consume within the DocumentDB cluster.
* `full_document` - (Optional) Determines what DocumentDB sends to your event stream during document update operations. If set to `UpdateLookup`, DocumentDB sends a delta describing the changes, along with a copy of the entire document. Otherwise, DocumentDB sends only a partial document that contains the changes. Valid values: `UpdateLookup`, `Default`.

### filter_criteria Configuration Block

* `filter` - (Optional) Set of up to 5 filter. If an event satisfies at least one, Lambda sends the event to the function or adds it to the next batch. [See below](#filter_criteria-filter-configuration-block).

#### filter_criteria filter Configuration Block

* `pattern` - (Optional) Filter pattern up to 4096 characters. See [Filter Rule Syntax](https://docs.aws.amazon.com/lambda/latest/dg/invocation-eventfiltering.html#filtering-syntax).

### metrics_config Configuration Block

* `metrics` - (Required) List containing the metrics to be produced by the event source mapping. Valid values: `EventCount`.

### provisioned_poller_config Configuration Block

* `maximum_pollers` - (Optional) Maximum number of event pollers this event source can scale up to. The range is between 1 and 2000.
* `minimum_pollers` - (Optional) Minimum number of event pollers this event source can scale down to. The range is between 1 and 200.

### scaling_config Configuration Block

* `maximum_concurrency` - (Optional) Limits the number of concurrent instances that the Amazon SQS event source can invoke. Must be greater than or equal to 2. See [Configuring maximum concurrency for Amazon SQS event sources](https://docs.aws.amazon.com/lambda/latest/dg/with-sqs.html#events-sqs-max-concurrency). You need to raise a [Service Quota Ticket](https://docs.aws.amazon.com/general/latest/gr/aws_service_limits.html) to increase the concurrency beyond 1000.

### self_managed_event_source Configuration Block

* `endpoints` - (Required) Map of endpoints for the self managed source. For Kafka self-managed sources, the key should be `KAFKA_BOOTSTRAP_SERVERS` and the value should be a string with a comma separated list of broker endpoints.

### self_managed_kafka_event_source_config Configuration Block

* `consumer_group_id` - (Optional) Kafka consumer group ID between 1 and 200 characters for use when creating this event source mapping. If one is not specified, this value will be automatically generated. See [SelfManagedKafkaEventSourceConfig Syntax](https://docs.aws.amazon.com/lambda/latest/dg/API_SelfManagedKafkaEventSourceConfig.html).

### source_access_configuration Configuration Block

* `type` - (Required) Type of authentication protocol, VPC components, or virtual host for your event source. For valid values, refer to the [AWS documentation](https://docs.aws.amazon.com/lambda/latest/api/API_SourceAccessConfiguration.html).
* `uri` - (Required) URI for this configuration. For type `VPC_SUBNET` the value should be `subnet:subnet_id` where `subnet_id` is the value you would find in an aws_subnet resource's id attribute. For type `VPC_SECURITY_GROUP` the value should be `security_group:security_group_id` where `security_group_id` is the value you would find in an aws_security_group resource's id attribute.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Event source mapping ARN.
* `function_arn` - ARN of the Lambda function the event source mapping is sending events to. (Note: this is a computed value that differs from `function_name` above.)
* `last_modified` - Date this resource was last modified.
* `last_processing_result` - Result of the last AWS Lambda invocation of your Lambda function.
* `state` - State of the event source mapping.
* `state_transition_reason` - Reason the event source mapping is in its current state.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `uuid` - UUID of the created event source mapping.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lambda event source mappings using the `UUID` (event source mapping identifier). For example:

```terraform
import {
  to = aws_lambda_event_source_mapping.example
  id = "12345kxodurf3443"
}
```

Using `terraform import`, import Lambda event source mappings using the `UUID` (event source mapping identifier). For example:

```console
% terraform import aws_lambda_event_source_mapping.example 12345kxodurf3443
```
