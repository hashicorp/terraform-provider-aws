---
subcategory: "MWAA (Managed Workflows for Apache Airflow)"
layout: "aws"
page_title: "AWS: aws_mwaa_environment"
description: |-
  Creates a MWAA Environment
---

# Resource: aws_mwaa_environment

Creates a MWAA Environment resource.

## Example Usage

A MWAA Environment requires an IAM role (`aws_iam_role`), two subnets in the private zone (`aws_subnet`) and a versioned S3 bucket (`aws_s3_bucket`).

### Basic Usage

```terraform
resource "aws_mwaa_environment" "example" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.example.arn
  name               = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}
```

### Example with Airflow configuration options

```terraform
resource "aws_mwaa_environment" "example" {
  airflow_configuration_options = {
    "core.default_task_retries" = 16
    "core.parallelism"          = 1
  }

  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.example.arn
  name               = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}
```

### Example with logging configurations

Note that Airflow task logs are enabled by default with the `INFO` log level.

```terraform
resource "aws_mwaa_environment" "example" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.example.arn

  logging_configuration {
    dag_processing_logs {
      enabled   = true
      log_level = "DEBUG"
    }

    scheduler_logs {
      enabled   = true
      log_level = "INFO"
    }

    task_logs {
      enabled   = true
      log_level = "WARNING"
    }

    webserver_logs {
      enabled   = true
      log_level = "ERROR"
    }

    worker_logs {
      enabled   = true
      log_level = "CRITICAL"
    }
  }

  name = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn
}
```

### Example with tags

```terraform
resource "aws_mwaa_environment" "example" {
  dag_s3_path        = "dags/"
  execution_role_arn = aws_iam_role.example.arn
  name               = "example"

  network_configuration {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = aws_subnet.private[*].id
  }

  source_bucket_arn = aws_s3_bucket.example.arn

  tags = {
    Name        = "example"
    Environment = "production"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `airflow_configuration_options` - (Optional) Airflow override options. Check the [Official documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-env-variables.html#configuring-env-variables-reference) for all possible configuration options.
* `airflow_version` - (Optional) Airflow version of your environment, will be set by default to the latest version that MWAA supports.
* `dag_s3_path` - (Required) Relative path to the DAG folder on your Amazon S3 storage bucket. For example, dags. For more information, see [Importing DAGs on Amazon MWAA](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-dag-import.html).
* `endpoint_management` - (Optional) Whether the VPC endpoints configured for the environment are created and managed by the customer or by AWS. If set to `SERVICE`, Amazon MWAA will create and manage the required VPC endpoints in your VPC. If set to `CUSTOMER`, you must create, and manage, the VPC endpoints for your VPC. Defaults to `SERVICE` if not set.
* `environment_class` - (Optional) Environment class for the cluster. Possible options are `mw1.micro`, `mw1.small`, `mw1.medium`, `mw1.large`. Will be set by default to `mw1.small`. Please check the [AWS Pricing](https://aws.amazon.com/de/managed-workflows-for-apache-airflow/pricing/) for more information about the environment classes.
* `execution_role_arn` - (Required) ARN of the task execution role that the Amazon MWAA and its environment can assume. Check the [official AWS documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/mwaa-create-role.html) for the detailed role specification.
* `kms_key` - (Optional) ARN of your KMS key that you want to use for encryption. Will be set to the ARN of the managed KMS key `aws/airflow` by default. Please check the [Official Documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/custom-keys-certs.html) for more information.
* `logging_configuration` - (Optional) Apache Airflow logs you want to send to Amazon CloudWatch Logs. See [`logging_configuration` Block](#logging_configuration-block) for details.
* `max_webservers` - (Optional) Maximum number of web servers that you want to run in your environment. Value need to be between `2` and `5` if `environment_class` is not `mw1.micro`, `1` otherwise.
* `max_workers` - (Optional) Maximum number of workers that can be automatically scaled up. Value need to be between `1` and `25`. Will be `10` by default.
* `min_webservers` - (Optional) Minimum number of web servers that you want to run in your environment. Value need to be between `2` and `5` if `environment_class` is not `mw1.micro`, `1` otherwise.
* `min_workers` - (Optional) Minimum number of workers that you want to run in your environment. Will be `1` by default.
* `name` - (Required) Name of the Apache Airflow Environment.
* `network_configuration` - (Required) Network configuration for your Apache Airflow Environment. This includes two private subnets as well as security groups for the Airflow environment. Each subnet requires internet connection, otherwise the deployment will fail. See [`network_configuration` Block](#network_configuration-block) for details.
* `plugins_s3_object_version` - (Optional) Version of the plugins.zip file you want to use.
* `plugins_s3_path` - (Optional) Relative path to the plugins.zip file on your Amazon S3 storage bucket. For example, plugins.zip. If a relative path is provided in the request, then plugins_s3_object_version is required. For more information, see [Importing DAGs on Amazon MWAA](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-dag-import.html).
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `requirements_s3_object_version` - (Optional) Version of the requirements.txt file you want to use.
* `requirements_s3_path` - (Optional) Relative path to the requirements.txt file on your Amazon S3 storage bucket. For example, requirements.txt. If a relative path is provided in the request, then requirements_s3_object_version is required. For more information, see [Importing DAGs on Amazon MWAA](https://docs.aws.amazon.com/mwaa/latest/userguide/configuring-dag-import.html).
* `schedulers` - (Optional) Number of schedulers that you want to run in your environment. v2.0.2 and above accepts `2` - `5`, default `2`. v1.10.12 accepts `1`.
* `source_bucket_arn` - (Required) ARN of your Amazon S3 storage bucket. For example, arn:aws:s3:::airflow-mybucketname.
* `startup_script_s3_object_version` - (Optional) Version of the startup shell script you want to use. You must specify the version ID that Amazon S3 assigns to the file every time you update the script.
* `startup_script_s3_path` - (Optional) Relative path to the script hosted in your bucket. The script runs as your environment starts before starting the Apache Airflow process. Use this script to install dependencies, modify configuration options, and set environment variables. See [Using a startup script](https://docs.aws.amazon.com/mwaa/latest/userguide/using-startup-script.html). Supported for environment versions 2.x and later.
* `tags` - (Optional) Map of resource tags to associate with the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `webserver_access_mode` - (Optional) Whether the webserver should be accessible over the internet or via your specified VPC. Possible options: `PRIVATE_ONLY` (default), `PUBLIC_ONLY` and `PUBLIC_AND_PRIVATE`.
* `weekly_maintenance_window_start` - (Optional) Start date for the weekly maintenance window.
* `worker_replacement_strategy` - (Optional) Worker replacement strategy. Valid values: `FORCED`, `GRACEFUL`.

### `logging_configuration` Block

The `logging_configuration` configuration block supports the following arguments.

* `dag_processing_logs` - (Optional) Log configuration options for processing DAGs. See [Module logging configuration blocks](#dag_processing_logs-scheduler_logs-task_logs-webserver_logs-and-worker_logs-block) for details. Disabled by default.
* `scheduler_logs` - (Optional) Log configuration options for the schedulers. See [Module logging configuration blocks](#dag_processing_logs-scheduler_logs-task_logs-webserver_logs-and-worker_logs-block) for details. Disabled by default.
* `task_logs` - (Optional) Log configuration options for DAG tasks. See [Module logging configuration blocks](#dag_processing_logs-scheduler_logs-task_logs-webserver_logs-and-worker_logs-block) for details. Enabled by default with `INFO` log level.
* `webserver_logs` - (Optional) Log configuration options for the webservers. See [Module logging configuration blocks](#dag_processing_logs-scheduler_logs-task_logs-webserver_logs-and-worker_logs-block) for details. Disabled by default.
* `worker_logs` - (Optional) Log configuration options for the workers. See [Module logging configuration blocks](#dag_processing_logs-scheduler_logs-task_logs-webserver_logs-and-worker_logs-block) for details. Disabled by default.

### `dag_processing_logs`, `scheduler_logs`, `task_logs`, `webserver_logs`, and `worker_logs` Block

Each module logging configuration block supports the following arguments.

* `cloud_watch_log_group_arn` - (Read-Only) ARN of the CloudWatch Logs log group where the logs are published.
* `enabled` - (Optional) Whether to enable the collection of logs.
* `log_level` - (Optional) Logging level. Valid values: `CRITICAL`, `ERROR`, `WARNING`, `INFO`, `DEBUG`. Will be `INFO` by default.

### `network_configuration` Block

The `network_configuration` configuration block supports the following arguments. More information about the required subnet and security group settings can be found in the [official AWS documentation](https://docs.aws.amazon.com/mwaa/latest/userguide/vpc-create.html).

* `security_group_ids` - (Required) Security groups IDs for the environment. At least one of the security group needs to allow MWAA resources to talk to each other, otherwise MWAA cannot be provisioned.
* `subnet_ids` - (Required) Private subnet IDs in which the environment should be created. MWAA requires two subnets.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the MWAA Environment.
* `created_at` - Created At date of the MWAA Environment.
* `database_vpc_endpoint_service` - VPC endpoint for the environment's Amazon RDS database.
* `last_updated` - Information about the last update of the environment. See [`last_updated` Block](#last_updated-block) for details.
* `service_role_arn` - Service Role ARN of the Amazon MWAA Environment.
* `status` - Status of the Amazon MWAA Environment.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `webserver_url` - Webserver URL of the MWAA Environment.
* `webserver_vpc_endpoint_service` - VPC endpoint for the environment's web server.

### `last_updated` Block

* `created_at` - Date and time the environment was last updated.
* `error` - Error that occurred during the last update, if any. See [`last_updated.error` Block](#last_updatederror-block) for details.
* `status` - Status of the last update.

### `last_updated.error` Block

* `error_code` - Error code of the last update error.
* `error_message` - Error message of the last update error.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `120m`)
- `update` - (Default `90m`)
- `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MWAA Environment using `Name`. For example:

```terraform
import {
  to = aws_mwaa_environment.example
  id = "MyAirflowEnvironment"
}
```

Using `terraform import`, import MWAA Environment using `Name`. For example:

```console
% terraform import aws_mwaa_environment.example MyAirflowEnvironment
```
