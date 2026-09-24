---
subcategory: "EventBridge Pipes"
layout: "aws"
page_title: "AWS: aws_pipes_pipe"
description: |-
  Terraform resource for managing an AWS EventBridge Pipes Pipe.
---

# Resource: aws_pipes_pipe

Terraform resource for managing an AWS EventBridge Pipes Pipe.

You can find out more about EventBridge Pipes in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes.html).

EventBridge Pipes are very configurable, and may require IAM permissions to work correctly. More information on the configuration options and IAM permissions can be found in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes.html).

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "main" {}

resource "aws_iam_role" "example" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "pipes.amazonaws.com"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.main.account_id
        }
      }
    }
  })
}

resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.example.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ReceiveMessage",
        ],
        Resource = [
          aws_sqs_queue.source.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "source" {}

resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.example.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
        ],
        Resource = [
          aws_sqs_queue.target.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "target" {}

resource "aws_pipes_pipe" "example" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = "example-pipe"
  role_arn   = aws_iam_role.example.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn
}
```

### Enrichment Usage

```terraform
resource "aws_pipes_pipe" "example" {
  name     = "example-pipe"
  role_arn = aws_iam_role.example.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  enrichment = aws_cloudwatch_event_api_destination.example.arn

  enrichment_parameters {
    http_parameters {
      path_parameter_values = ["example-path-param"]

      header_parameters = {
        "example-header"        = "example-value"
        "second-example-header" = "second-example-value"
      }

      query_string_parameters = {
        "example-query-string"        = "example-value"
        "second-example-query-string" = "second-example-value"
      }
    }
  }
}
```

### Filter Usage

```terraform
resource "aws_pipes_pipe" "example" {
  name     = "example-pipe"
  role_arn = aws_iam_role.example.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  source_parameters {
    filter_criteria {
      filter {
        pattern = jsonencode({
          source = ["event-source"]
        })
      }
    }
  }
}
```

### CloudWatch Logs Logging Configuration Usage

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "example-pipe-target"
}

resource "aws_pipes_pipe" "example" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]

  name     = "example-pipe"
  role_arn = aws_iam_role.example.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn
  log_configuration {
    include_execution_data = ["ALL"]
    level                  = "INFO"
    cloudwatch_logs_log_destination {
      log_group_arn = aws_cloudwatch_log_group.target.arn
    }
  }
}
```

### SQS Source and Target Configuration Usage

```terraform
resource "aws_pipes_pipe" "example" {
  name     = "example-pipe"
  role_arn = aws_iam_role.example.arn
  source   = aws_sqs_queue.source.arn
  target   = aws_sqs_queue.target.arn

  source_parameters {
    sqs_queue_parameters {
      batch_size                         = 1
      maximum_batching_window_in_seconds = 2
    }
  }

  target_parameters {
    sqs_queue_parameters {
      message_deduplication_id = "example-dedupe"
      message_group_id         = "example-group"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `role_arn` - (Required) ARN of the role that allows the pipe to send data to the target.
* `source` - (Required) Source resource of the pipe. This field typically requires an ARN. However, when using a self-managed Kafka cluster, you should use a different format. Instead of an ARN, use 'smk://' followed by the bootstrap server's address.
* `target` - (Required) Target resource of the pipe (typically an ARN).

The following arguments are optional:

* `description` - (Optional) Description of the pipe. At most 512 characters.
* `desired_state` - (Optional) State the pipe should be in. One of: `RUNNING`, `STOPPED`.
* `enrichment` - (Optional) Enrichment resource of the pipe (typically an ARN). Read more about enrichment in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes.html#pipes-enrichment).
* `enrichment_parameters` - (Optional) Parameters to configure enrichment for the pipe. See [`enrichment_parameters` Block](#enrichment_parameters-block) for details.
* `kms_key_identifier` - (Optional) Identifier of the AWS KMS customer managed key for EventBridge to use, if you choose to use a customer managed key to encrypt pipe data. The identifier can be the key ARN, KeyId, key alias, or key alias ARN. If not set, EventBridge uses an AWS owned key to encrypt pipe data.
* `log_configuration` - (Optional) Logging configuration settings for the pipe. See [`log_configuration` Block](#log_configuration-block) for details.
* `name` - (Optional) Name of the pipe. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_parameters` - (Optional) Parameters to configure a source for the pipe. See [`source_parameters` Block](#source_parameters-block) for details.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `target_parameters` - (Optional) Parameters to configure a target for the pipe. See [`target_parameters` Block](#target_parameters-block) for details.

### `enrichment_parameters` Block

You can find out more about EventBridge Pipes Enrichment in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/pipes-enrichment.html).

* `http_parameters` - (Optional) HTTP parameters to use when the target is an API Gateway REST endpoint or EventBridge ApiDestination. If you specify an API Gateway REST API or EventBridge ApiDestination as a target, you can use this parameter to specify headers, path parameters, and query string keys/values as part of your target invoking request. If you're using ApiDestinations, the corresponding Connection can also have these values configured. In case of any conflicting keys, values from the Connection take precedence. See [`enrichment_parameters.http_parameters` Block](#enrichment_parametershttp_parameters-block) for details.
* `input_template` - (Optional) Valid JSON text passed to the target. In this case, nothing from the event itself is passed to the target. Maximum length of 8192 characters.

### `enrichment_parameters.http_parameters` Block

* `header_parameters` - (Optional) Key-value mapping of the headers that need to be sent as part of request invoking the API Gateway REST API or EventBridge ApiDestination.
* `path_parameter_values` - (Optional) Path parameter values used to populate API Gateway REST API or EventBridge ApiDestination path wildcards ("*").
* `query_string_parameters` - (Optional) Key-value mapping of the query strings that need to be sent as part of request invoking the API Gateway REST API or EventBridge ApiDestination.

### `log_configuration` Block

You can find out more about EventBridge Pipes Log Configuration in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes-logs.html).

* `cloudwatch_logs_log_destination` - (Optional) Amazon CloudWatch Logs logging configuration settings for the pipe. See [`cloudwatch_logs_log_destination` Block](#cloudwatch_logs_log_destination-block) for details.
* `firehose_log_destination` - (Optional) Amazon Kinesis Data Firehose logging configuration settings for the pipe. See [`firehose_log_destination` Block](#firehose_log_destination-block) for details.
* `include_execution_data` - (Optional) String list that specifies whether the execution data (specifically, the `payload`, `awsRequest`, and `awsResponse` fields) is included in the log messages for this pipe. This applies to all log destinations for the pipe. Valid values `ALL`.
* `level` - (Required) Level of logging detail to include. Valid values `OFF`, `ERROR`, `INFO` and `TRACE`.
* `s3_log_destination` - (Optional) Amazon S3 logging configuration settings for the pipe. See [`s3_log_destination` Block](#s3_log_destination-block) for details.

### `cloudwatch_logs_log_destination` Block

* `log_group_arn` - (Required) ARN for the CloudWatch log group to which EventBridge sends the log records.

### `firehose_log_destination` Block

* `delivery_stream_arn` - (Required) ARN of the Kinesis Data Firehose delivery stream to which EventBridge delivers the pipe log records.

### `s3_log_destination` Block

* `bucket_name` - (Required) Name of the Amazon S3 bucket to which EventBridge delivers the log records for the pipe.
* `bucket_owner` - (Required) Amazon Web Services account that owns the Amazon S3 bucket to which EventBridge delivers the log records for the pipe.
* `output_format` - (Optional) EventBridge format for the log records. Valid values `json`, `plain` and `w3c`.
* `prefix` - (Optional) Prefix text with which to begin Amazon S3 log object names.

### `source_parameters` Block

You can find out more about EventBridge Pipes Sources in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes-event-source.html).

* `activemq_broker_parameters` - (Optional) Parameters for using an Active MQ broker as a source. See [`activemq_broker_parameters` Block](#activemq_broker_parameters-block) for details.
* `dynamodb_stream_parameters` - (Optional) Parameters for using a DynamoDB stream as a source. See [`dynamodb_stream_parameters` Block](#dynamodb_stream_parameters-block) for details.
* `filter_criteria` - (Optional) Collection of event patterns used to [filter events](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes-event-filtering.html). See [`filter_criteria` Block](#filter_criteria-block) for details.
* `kinesis_stream_parameters` - (Optional) Parameters for using a Kinesis stream as a source. See [`source_parameters.kinesis_stream_parameters` Block](#source_parameterskinesis_stream_parameters-block) for details.
* `managed_streaming_kafka_parameters` - (Optional) Parameters for using an MSK stream as a source. See [`managed_streaming_kafka_parameters` Block](#managed_streaming_kafka_parameters-block) for details.
* `rabbitmq_broker_parameters` - (Optional) Parameters for using a Rabbit MQ broker as a source. See [`rabbitmq_broker_parameters` Block](#rabbitmq_broker_parameters-block) for details.
* `self_managed_kafka_parameters` - (Optional) Parameters for using a self-managed Apache Kafka stream as a source. See [`self_managed_kafka_parameters` Block](#self_managed_kafka_parameters-block) for details.
* `sqs_queue_parameters` - (Optional) Parameters for using an Amazon SQS stream as a source. See [`source_parameters.sqs_queue_parameters` Block](#source_parameterssqs_queue_parameters-block) for details.

### `activemq_broker_parameters` Block

* `batch_size` - (Optional) Maximum number of records to include in each batch. Maximum value of 10000.
* `credentials` - (Required) Credentials needed to access the resource. See [`source_parameters.activemq_broker_parameters.credentials` Block](#source_parametersactivemq_broker_parameterscredentials-block) for details.
* `maximum_batching_window_in_seconds` - (Optional) Maximum length of a time to wait for events. Maximum value of 300.
* `queue_name` - (Required) Name of the destination queue to consume. Maximum length of 1000.

### `source_parameters.activemq_broker_parameters.credentials` Block

* `basic_auth` - (Required) ARN of the Secrets Manager secret containing the basic auth credentials.

### `dynamodb_stream_parameters` Block

* `batch_size` - (Optional) Maximum number of records to include in each batch. Maximum value of 10000.
* `dead_letter_config` - (Optional) Define the target queue to send dead-letter queue events to. See [`dead_letter_config` Block](#dead_letter_config-block) for details.
* `maximum_batching_window_in_seconds` - (Optional) Maximum length of a time to wait for events. Maximum value of 300.
* `maximum_record_age_in_seconds` - (Optional) Discard records older than the specified age. The default value is -1, which sets the maximum age to infinite. When the value is set to infinite, EventBridge never discards old records. Maximum value of 604,800.
* `maximum_retry_attempts` - (Optional) Discard records after the specified number of retries. The default value is -1, which sets the maximum number of retries to infinite. When MaximumRetryAttempts is infinite, EventBridge retries failed records until the record expires in the event source. Maximum value of 10,000.
* `on_partial_batch_item_failure` - (Optional) Define how to handle item process failures. AUTOMATIC_BISECT halves each batch and retry each half until all the records are processed or there is one failed message left in the batch. Valid values: AUTOMATIC_BISECT.
* `parallelization_factor` - (Optional) Number of batches to process concurrently from each shard. The default value is 1. Maximum value of 10.
* `starting_position` - (Required) Position in a stream from which to start reading. Valid values: TRIM_HORIZON, LATEST.

### `dead_letter_config` Block

* `arn` - (Optional) ARN of the Amazon SQS queue specified as the target for the dead-letter queue.

### `filter_criteria` Block

* `filter` - (Optional) Array of up to 5 event patterns. See [`filter` Block](#filter-block) for details.

### `filter` Block

* `pattern` - (Required) Event pattern. At most 4096 characters.

### `source_parameters.kinesis_stream_parameters` Block

* `batch_size` - (Optional) Maximum number of records to include in each batch. Maximum value of 10000.
* `dead_letter_config` - (Optional) Define the target queue to send dead-letter queue events to. See [`dead_letter_config` Block](#dead_letter_config-block) for details.
* `maximum_batching_window_in_seconds` - (Optional) Maximum length of a time to wait for events. Maximum value of 300.
* `maximum_record_age_in_seconds` - (Optional) Discard records older than the specified age. The default value is -1, which sets the maximum age to infinite. When the value is set to infinite, EventBridge never discards old records. Maximum value of 604,800.
* `maximum_retry_attempts` - (Optional) Discard records after the specified number of retries. The default value is -1, which sets the maximum number of retries to infinite. When MaximumRetryAttempts is infinite, EventBridge retries failed records until the record expires in the event source. Maximum value of 10,000.
* `on_partial_batch_item_failure` - (Optional) Define how to handle item process failures. AUTOMATIC_BISECT halves each batch and retry each half until all the records are processed or there is one failed message left in the batch. Valid values: AUTOMATIC_BISECT.
* `parallelization_factor` - (Optional) Number of batches to process concurrently from each shard. The default value is 1. Maximum value of 10.
* `starting_position` - (Required) Position in a stream from which to start reading. Valid values: TRIM_HORIZON, LATEST, AT_TIMESTAMP.
* `starting_position_timestamp` - (Optional) With StartingPosition set to AT_TIMESTAMP, the time from which to start reading, in Unix time seconds.

### `managed_streaming_kafka_parameters` Block

* `batch_size` - (Optional) Maximum number of records to include in each batch. Maximum value of 10000.
* `consumer_group_id` - (Optional) Name of the destination queue to consume. Maximum value of 200.
* `credentials` - (Optional) Credentials needed to access the resource. See [`source_parameters.managed_streaming_kafka_parameters.credentials` Block](#source_parametersmanaged_streaming_kafka_parameterscredentials-block) for details.
* `maximum_batching_window_in_seconds` - (Optional) Maximum length of a time to wait for events. Maximum value of 300.
* `starting_position` - (Optional) Position in a stream from which to start reading. Valid values: TRIM_HORIZON, LATEST.
* `topic_name` - (Required) Name of the topic that the pipe will read from. Maximum length of 249.

### `source_parameters.managed_streaming_kafka_parameters.credentials` Block

* `client_certificate_tls_auth` - (Optional) ARN of the Secrets Manager secret containing the credentials.
* `sasl_scram_512_auth` - (Optional) ARN of the Secrets Manager secret containing the credentials.

### `rabbitmq_broker_parameters` Block

* `batch_size` - (Optional) Maximum number of records to include in each batch. Maximum value of 10000.
* `credentials` - (Required) Credentials needed to access the resource. See [`source_parameters.rabbitmq_broker_parameters.credentials` Block](#source_parametersrabbitmq_broker_parameterscredentials-block) for details.
* `maximum_batching_window_in_seconds` - (Optional) Maximum length of a time to wait for events. Maximum value of 300.
* `queue_name` - (Required) Name of the destination queue to consume. Maximum length of 1000.
* `virtual_host` - (Optional) Name of the virtual host associated with the source broker. Maximum length of 200.

### `source_parameters.rabbitmq_broker_parameters.credentials` Block

* `basic_auth` - (Required) ARN of the Secrets Manager secret containing the credentials.

### `self_managed_kafka_parameters` Block

* `additional_bootstrap_servers` - (Optional) Array of server URLs. Maximum number of 2 items, each of maximum length 300.
* `batch_size` - (Optional) Maximum number of records to include in each batch. Maximum value of 10000.
* `consumer_group_id` - (Optional) Name of the destination queue to consume. Maximum value of 200.
* `credentials` - (Optional) Credentials needed to access the resource. See [`source_parameters.self_managed_kafka_parameters.credentials` Block](#source_parametersself_managed_kafka_parameterscredentials-block) for details.
* `maximum_batching_window_in_seconds` - (Optional) Maximum length of a time to wait for events. Maximum value of 300.
* `server_root_ca_certificate` - (Optional) ARN of the Secrets Manager secret used for certification.
* `starting_position` - (Optional) Position in a stream from which to start reading. Valid values: TRIM_HORIZON, LATEST.
* `topic_name` - (Required) Name of the topic that the pipe will read from. Maximum length of 249.
* `vpc` - (Optional) VPC subnets and security groups for the stream, and whether a public IP address is to be used. See [`vpc` Block](#vpc-block) for details.

### `source_parameters.self_managed_kafka_parameters.credentials` Block

* `basic_auth` - (Optional) ARN of the Secrets Manager secret containing the credentials.
* `client_certificate_tls_auth` - (Optional) ARN of the Secrets Manager secret containing the credentials.
* `sasl_scram_256_auth` - (Optional) ARN of the Secrets Manager secret containing the credentials.
* `sasl_scram_512_auth` - (Optional) ARN of the Secrets Manager secret containing the credentials.

### `vpc` Block

* `security_groups` - (Optional) List of security groups associated with the stream. These security groups must all be in the same VPC. You can specify as many as five security groups. If you do not specify a security group, the default security group for the VPC is used.
* `subnets` - (Optional) List of the subnets associated with the stream. These subnets must all be in the same VPC. You can specify as many as 16 subnets.

### `source_parameters.sqs_queue_parameters` Block

* `batch_size` - (Optional) Maximum number of records to include in each batch. Maximum value of 10000.
* `maximum_batching_window_in_seconds` - (Optional) Maximum length of a time to wait for events. Maximum value of 300.

### `target_parameters` Block

You can find out more about EventBridge Pipes Targets in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes-event-target.html).

* `batch_job_parameters` - (Optional) Parameters for using an AWS Batch job as a target. See [`batch_job_parameters` Block](#batch_job_parameters-block) for details.
* `cloudwatch_logs_parameters` - (Optional) Parameters for using a CloudWatch Logs log stream as a target. See [`cloudwatch_logs_parameters` Block](#cloudwatch_logs_parameters-block) for details.
* `ecs_task_parameters` - (Optional) Parameters for using an Amazon ECS task as a target. See [`ecs_task_parameters` Block](#ecs_task_parameters-block) for details.
* `eventbridge_event_bus_parameters` - (Optional) Parameters for using an EventBridge event bus as a target. See [`eventbridge_event_bus_parameters` Block](#eventbridge_event_bus_parameters-block) for details.
* `http_parameters` - (Optional) Custom parameters used when the target is an API Gateway REST API or EventBridge ApiDestination. See [`target_parameters.http_parameters` Block](#target_parametershttp_parameters-block) for details.
* `input_template` - (Optional) Valid JSON text passed to the target. In this case, nothing from the event itself is passed to the target. Maximum length of 8192 characters.
* `kinesis_stream_parameters` - (Optional) Parameters for using a Kinesis stream as a target. See [`target_parameters.kinesis_stream_parameters` Block](#target_parameterskinesis_stream_parameters-block) for details.
* `lambda_function_parameters` - (Optional) Parameters for using a Lambda function as a target. See [`lambda_function_parameters` Block](#lambda_function_parameters-block) for details.
* `redshift_data_parameters` - (Optional) Custom parameters used when the target is an Amazon Redshift cluster to invoke the Amazon Redshift Data API BatchExecuteStatement. See [`redshift_data_parameters` Block](#redshift_data_parameters-block) for details.
* `sagemaker_pipeline_parameters` - (Optional) Parameters for using a SageMaker AI pipeline as a target. See [`sagemaker_pipeline_parameters` Block](#sagemaker_pipeline_parameters-block) for details.
* `sqs_queue_parameters` - (Optional) Parameters for using an Amazon SQS stream as a target. See [`target_parameters.sqs_queue_parameters` Block](#target_parameterssqs_queue_parameters-block) for details.
* `step_function_state_machine_parameters` - (Optional) Parameters for using a Step Functions state machine as a target. See [`step_function_state_machine_parameters` Block](#step_function_state_machine_parameters-block) for details.

### `batch_job_parameters` Block

* `array_properties` - (Optional) Array properties for the submitted job, such as the size of the array. The array size can be between 2 and 10,000. If you specify array properties for a job, it becomes an array job. This parameter is used only if the target is an AWS Batch job. See [`array_properties` Block](#array_properties-block) for details.
* `container_overrides` - (Optional) Overrides that are sent to a container. See [`container_overrides` Block](#container_overrides-block) for details.
* `depends_on` - (Optional) List of dependencies for the job. A job can depend upon a maximum of 20 jobs. You can specify a SEQUENTIAL type dependency without specifying a job ID for array jobs so that each child array job completes sequentially, starting at index 0. You can also specify an N_TO_N type dependency with a job ID for array jobs. In that case, each index child of this job must wait for the corresponding index child of each dependency to complete before it can begin. See [`depends_on` Block](#depends_on-block) for details.
* `job_definition` - (Required) Job definition used by this job. This value can be one of name, name:revision, or the ARN for the job definition. If name is specified without a revision then the latest active revision is used.
* `job_name` - (Required) Name of the job. It can be up to 128 letters long.
* `parameters` - (Optional) Additional parameters passed to the job that replace parameter substitution placeholders that are set in the job definition. Parameters are specified as a key and value pair mapping. Parameters included here override any corresponding parameter defaults from the job definition.
* `retry_strategy` - (Optional) Retry strategy to use for failed jobs. When a retry strategy is specified here, it overrides the retry strategy defined in the job definition. See [`retry_strategy` Block](#retry_strategy-block) for details.

### `array_properties` Block

* `size` - (Optional) Size of the array, if this is an array batch job. Minimum value of 2. Maximum value of 10,000.

### `container_overrides` Block

* `command` - (Optional) List of commands to send to the container that overrides the default command from the Docker image or the task definition.
* `environment` - (Optional) Environment variables to send to the container. You can add new environment variables, which are added to the container at launch, or you can override the existing environment variables from the Docker image or the task definition. Environment variables cannot start with " AWS Batch ". This naming convention is reserved for variables that AWS Batch sets. See [`target_parameters.batch_job_parameters.container_overrides.environment` Block](#target_parametersbatch_job_parameterscontainer_overridesenvironment-block) for details.
* `instance_type` - (Optional) Instance type to use for a multi-node parallel job. This parameter isn't applicable to single-node container jobs or jobs that run on Fargate resources, and shouldn't be provided.
* `resource_requirement` - (Optional) Type and amount of resources to assign to a container. This overrides the settings in the job definition. The supported resources include GPU, MEMORY, and VCPU. See [`target_parameters.batch_job_parameters.container_overrides.resource_requirement` Block](#target_parametersbatch_job_parameterscontainer_overridesresource_requirement-block) for details.

### `target_parameters.batch_job_parameters.container_overrides.environment` Block

* `name` - (Optional) Name of the key-value pair. For environment variables, this is the name of the environment variable.
* `value` - (Optional) Value of the key-value pair. For environment variables, this is the value of the environment variable.

### `target_parameters.batch_job_parameters.container_overrides.resource_requirement` Block

* `type` - (Optional) Type of resource to assign to a container. The supported resources include GPU, MEMORY, and VCPU.
* `value` - (Optional) Quantity of the specified resource to reserve for the container. [The values vary based on the type specified](https://docs.aws.amazon.com/eventbridge/latest/pipes-reference/API_BatchResourceRequirement.html).

### `depends_on` Block

* `job_id` - (Optional) Job ID of the AWS Batch job that's associated with this dependency.
* `type` - (Optional) Type of the job dependency. Valid Values: N_TO_N, SEQUENTIAL.

### `retry_strategy` Block

* `attempts` - (Optional) Number of times to move a job to the RUNNABLE status. If the value of attempts is greater than one, the job is retried on failure the same number of attempts as the value. Maximum value of 10.

### `cloudwatch_logs_parameters` Block

* `log_stream_name` - (Optional) Name of the log stream.
* `timestamp` - (Optional) Time the event occurred, expressed as the number of milliseconds after Jan 1, 1970 00:00:00 UTC. This is the JSON path to the field in the event e.g. $.detail.timestamp

### `ecs_task_parameters` Block

* `capacity_provider_strategy` - (Optional) List of capacity provider strategies to use for the task. If a capacityProviderStrategy is specified, the launchType parameter must be omitted. If no capacityProviderStrategy or launchType is specified, the defaultCapacityProviderStrategy for the cluster is used. See [`capacity_provider_strategy` Block](#capacity_provider_strategy-block) for details.
* `enable_ecs_managed_tags` - (Optional) Whether to enable Amazon ECS managed tags for the task. Valid values: true, false.
* `enable_execute_command` - (Optional) Whether to enable the execute command functionality for the containers in this task. If true, this enables execute command functionality on all containers in the task. Valid values: true, false.
* `group` - (Optional) Amazon ECS task group for the task. The maximum length is 255 characters.
* `launch_type` - (Optional) Launch type on which your task is running. The launch type that you specify here must match one of the launch type (compatibilities) of the target task. The FARGATE value is supported only in the Regions where AWS Fargate with Amazon ECS is supported. Valid Values: EC2, FARGATE, EXTERNAL
* `network_configuration` - (Optional) Use this structure if the Amazon ECS task uses the awsvpc network mode. This structure specifies the VPC subnets and security groups associated with the task, and whether a public IP address is to be used. This structure is required if LaunchType is FARGATE because the awsvpc mode is required for Fargate tasks. If you specify NetworkConfiguration when the target ECS task does not use the awsvpc network mode, the task fails. See [`network_configuration` Block](#network_configuration-block) for details.
* `overrides` - (Optional) Overrides that are associated with a task. See [`overrides` Block](#overrides-block) for details.
* `placement_constraint` - (Optional) Array of placement constraint objects to use for the task. You can specify up to 10 constraints per task (including constraints in the task definition and those specified at runtime). See [`placement_constraint` Block](#placement_constraint-block) for details.
* `placement_strategy` - (Optional) Placement strategy objects to use for the task. You can specify a maximum of five strategy rules per task. See [`placement_strategy` Block](#placement_strategy-block) for details.
* `platform_version` - (Optional) Platform version for the task. Specify only the numeric portion of the platform version, such as 1.1.0. This structure is used only if LaunchType is FARGATE.
* `propagate_tags` - (Optional) Whether to propagate the tags from the task definition to the task. If no value is specified, the tags are not propagated. Tags can only be propagated to the task during task creation. To add tags to a task after task creation, use the TagResource API action. Valid Values: TASK_DEFINITION
* `reference_id` - (Optional) Reference ID to use for the task. Maximum length of 1,024.
* `tags` - (Optional) Key-value map of tags that you apply to the task to help you categorize and organize them.
* `task_count` - (Optional) Number of tasks to create based on TaskDefinition. The default is 1.
* `task_definition_arn` - (Required) ARN of the task definition to use if the event target is an Amazon ECS task.

### `capacity_provider_strategy` Block

* `base` - (Optional) Base value designates how many tasks, at a minimum, to run on the specified capacity provider. Only one capacity provider in a capacity provider strategy can have a base defined. If no value is specified, the default value of 0 is used. Maximum value of 100,000.
* `capacity_provider` - (Required) Short name of the capacity provider. Maximum value of 255.
* `weight` - (Optional) Weight value designates the relative percentage of the total number of tasks launched that should use the specified capacity provider. The weight value is taken into consideration after the base value, if defined, is satisfied. Maximum value of 1,000.

### `network_configuration` Block

* `aws_vpc_configuration` - (Optional) Use this structure to specify the VPC subnets and security groups for the task, and whether a public IP address is to be used. This structure is relevant only for ECS tasks that use the awsvpc network mode. See [`aws_vpc_configuration` Block](#aws_vpc_configuration-block) for details.

### `aws_vpc_configuration` Block

* `assign_public_ip` - (Optional) Whether the task's elastic network interface receives a public IP address. You can specify ENABLED only when LaunchType in EcsParameters is set to FARGATE. Valid Values: ENABLED, DISABLED.
* `security_groups` - (Optional) Security groups associated with the task. These security groups must all be in the same VPC. You can specify as many as five security groups. If you do not specify a security group, the default security group for the VPC is used.
* `subnets` - (Optional) Subnets associated with the task. These subnets must all be in the same VPC. You can specify as many as 16 subnets.

### `overrides` Block

* `container_override` - (Optional) One or more container overrides that are sent to a task. See [`container_override` Block](#container_override-block) for details.
* `cpu` - (Optional) CPU override for the task.
* `ephemeral_storage` - (Optional) Ephemeral storage setting override for the task. See [`ephemeral_storage` Block](#ephemeral_storage-block) for details.
* `execution_role_arn` - (Optional) ARN of the task execution IAM role override for the task.
* `inference_accelerator_override` - (Optional) List of Elastic Inference accelerator overrides for the task. See [`inference_accelerator_override` Block](#inference_accelerator_override-block) for details.
* `memory` - (Optional) Memory override for the task.
* `task_role_arn` - (Optional) ARN of the IAM role that containers in this task can assume. All containers in this task are granted the permissions that are specified in this role.

### `container_override` Block

* `command` - (Optional) List of commands to send to the container that overrides the default command from the Docker image or the task definition. You must also specify a container name.
* `cpu` - (Optional) Number of cpu units reserved for the container, instead of the default value from the task definition. You must also specify a container name.
* `environment` - (Optional) Environment variables to send to the container. You can add new environment variables, which are added to the container at launch, or you can override the existing environment variables from the Docker image or the task definition. You must also specify a container name. See [`target_parameters.ecs_task_parameters.overrides.container_override.environment` Block](#target_parametersecs_task_parametersoverridescontainer_overrideenvironment-block) for details.
* `environment_file` - (Optional) List of files containing the environment variables to pass to a container, instead of the value from the container definition. See [`environment_file` Block](#environment_file-block) for details.
* `memory` - (Optional) Hard limit (in MiB) of memory to present to the container, instead of the default value from the task definition. If your container attempts to exceed the memory specified here, the container is killed. You must also specify a container name.
* `memory_reservation` - (Optional) Soft limit (in MiB) of memory to reserve for the container, instead of the default value from the task definition. You must also specify a container name.
* `name` - (Optional) Name of the container that receives the override. This parameter is required if any override is specified.
* `resource_requirement` - (Optional) Type and amount of a resource to assign to a container, instead of the default value from the task definition. The only supported resource is a GPU. See [`target_parameters.ecs_task_parameters.overrides.container_override.resource_requirement` Block](#target_parametersecs_task_parametersoverridescontainer_overrideresource_requirement-block) for details.

### `target_parameters.ecs_task_parameters.overrides.container_override.environment` Block

* `name` - (Optional) Name of the key-value pair. For environment variables, this is the name of the environment variable.
* `value` - (Optional) Value of the key-value pair. For environment variables, this is the value of the environment variable.

### `environment_file` Block

* `type` - (Optional) File type to use. The only supported value is s3.
* `value` - (Optional) ARN of the Amazon S3 object containing the environment variable file.

### `target_parameters.ecs_task_parameters.overrides.container_override.resource_requirement` Block

* `type` - (Optional) Type of resource to assign to a container. The supported values are GPU or InferenceAccelerator.
* `value` - (Optional) Value for the specified resource type. If the GPU type is used, the value is the number of physical GPUs the Amazon ECS container agent reserves for the container. The number of GPUs that's reserved for all containers in a task can't exceed the number of available GPUs on the container instance that the task is launched on. If the InferenceAccelerator type is used, the value matches the deviceName for an InferenceAccelerator specified in a task definition.

### `ephemeral_storage` Block

* `size_in_gib` - (Required) Total amount, in GiB, of ephemeral storage to set for the task. The minimum supported value is 21 GiB and the maximum supported value is 200 GiB.

### `inference_accelerator_override` Block

* `device_name` - (Optional) Elastic Inference accelerator device name to override for the task. This parameter must match a deviceName specified in the task definition.
* `device_type` - (Optional) Elastic Inference accelerator type to use.

### `placement_constraint` Block

* `expression` - (Optional) Cluster query language expression to apply to the constraint. You cannot specify an expression if the constraint type is distinctInstance. Maximum length of 2,000.
* `type` - (Optional) Type of constraint. Use distinctInstance to ensure that each task in a particular group is running on a different container instance. Use memberOf to restrict the selection to a group of valid candidates. Valid Values: distinctInstance, memberOf.

### `placement_strategy` Block

* `field` - (Optional) Field to apply the placement strategy against. For the spread placement strategy, valid values are instanceId (or host, which has the same effect), or any platform or custom attribute that is applied to a container instance, such as attribute:ecs.availability-zone. For the binpack placement strategy, valid values are cpu and memory. For the random placement strategy, this field is not used. Maximum length of 255.
* `type` - (Optional) Type of placement strategy. The random placement strategy randomly places tasks on available candidates. The spread placement strategy spreads placement across available candidates evenly based on the field parameter. The binpack strategy places tasks on available candidates that have the least available amount of the resource that is specified with the field parameter. For example, if you binpack on memory, a task is placed on the instance with the least amount of remaining memory (but still enough to run the task). Valid Values: random, spread, binpack.

### `eventbridge_event_bus_parameters` Block

* `detail_type` - (Optional) Free-form string, with a maximum of 128 characters, used to decide what fields to expect in the event detail.
* `endpoint_id` - (Optional) URL subdomain of the endpoint. For example, if the URL for Endpoint is https://abcde.veo.endpoints.event.amazonaws.com, then the EndpointId is abcde.veo.
* `resources` - (Optional) List of AWS resources, identified by ARN, which the event primarily concerns. Any number, including zero, may be present.
* `source` - (Optional) Source of the event. Maximum length of 256.
* `time` - (Optional) Time stamp of the event, per RFC3339. If no time stamp is provided, the time stamp of the PutEvents call is used. This is the JSON path to the field in the event e.g. $.detail.timestamp

### `target_parameters.http_parameters` Block

* `header_parameters` - (Optional) Key-value mapping of the headers that need to be sent as part of request invoking the API Gateway REST API or EventBridge ApiDestination.
* `path_parameter_values` - (Optional) Path parameter values used to populate API Gateway REST API or EventBridge ApiDestination path wildcards ("*").
* `query_string_parameters` - (Optional) Key-value mapping of the query strings that need to be sent as part of request invoking the API Gateway REST API or EventBridge ApiDestination.

### `target_parameters.kinesis_stream_parameters` Block

* `partition_key` - (Required) Value used to determine which shard in the stream the data record is assigned to. Partition keys are Unicode strings with a maximum length limit of 256 characters for each key. Amazon Kinesis Data Streams uses the partition key as input to a hash function that maps the partition key and associated data to a specific shard. Specifically, an MD5 hash function is used to map partition keys to 128-bit integer values and to map associated data records to shards. As a result of this hashing mechanism, all data records with the same partition key map to the same shard within the stream.

### `lambda_function_parameters` Block

* `invocation_type` - (Required) Whether to invoke the function synchronously or asynchronously. Valid Values: REQUEST_RESPONSE, FIRE_AND_FORGET.

### `redshift_data_parameters` Block

* `database` - (Required) Name of the database. Required when authenticating using temporary credentials.
* `db_user` - (Optional) Database user name. Required when authenticating using temporary credentials.
* `secret_manager_arn` - (Optional) Name or ARN of the secret that enables access to the database. Required when authenticating using Secrets Manager.
* `sqls` - (Required) List of SQL statements text to run, each of maximum length of 100,000.
* `statement_name` - (Optional) Name of the SQL statement. You can name the SQL statement when you create it to identify the query.
* `with_event` - (Optional) Whether to send an event back to EventBridge after the SQL statement runs.

### `sagemaker_pipeline_parameters` Block

* `pipeline_parameter` - (Optional) List of Parameter names and values for SageMaker AI Model Building Pipeline execution. See [`pipeline_parameter` Block](#pipeline_parameter-block) for details.

### `pipeline_parameter` Block

* `name` - (Required) Name of parameter to start execution of a SageMaker AI Model Building Pipeline. Maximum length of 256.
* `value` - (Required) Value of parameter to start execution of a SageMaker AI Model Building Pipeline. Maximum length of 1024.

### `target_parameters.sqs_queue_parameters` Block

* `message_deduplication_id` - (Optional) Token used for deduplication of sent messages. This parameter applies only to FIFO (first-in-first-out) queues.
* `message_group_id` - (Optional) FIFO message group ID to use as the target.

### `step_function_state_machine_parameters` Block

* `invocation_type` - (Required) Whether to invoke the function synchronously or asynchronously. Valid Values: REQUEST_RESPONSE, FIRE_AND_FORGET.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of this pipe.
* `id` - Same as `name`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import pipes using the `name`. For example:

```terraform
import {
  to = aws_pipes_pipe.example
  id = "my-pipe"
}
```

Using `terraform import`, import pipes using the `name`. For example:

```console
% terraform import aws_pipes_pipe.example my-pipe
```
