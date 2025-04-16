---
subcategory: ""
layout: "aws"
page_title: "Terraform AWS Provider Version 6 Upgrade Guide"
description: |-
  Terraform AWS Provider Version 6 Upgrade Guide
---

# Terraform AWS Provider Version 6 Upgrade Guide

Version 6.0.0 of the AWS provider for Terraform is a major release and includes changes that you need to consider when upgrading. This guide will help with that process and focuses only on changes from version 5.x to version 6.0.0. See the [Version 5 Upgrade Guide](/docs/providers/aws/guides/version-5-upgrade.html) for information on upgrading from 4.x to version 5.0.0.

Upgrade topics:

<!-- TOC depthFrom:2 depthTo:2 -->

- [Provider Version Configuration](#provider-version-configuration)
- [Provider Arguments](#provider-arguments)
- [Dropping Support For Amazon SimpleDB](#dropping-support-for-amazon-simpledb)
- [Dropping Support For Amazon Worklink](#dropping-support-for-amazon-worklink)
- [AWS OpsWorks Stacks End of Life](#aws-opsworks-stacks-end-of-life)
- [data-source/aws_ami](#data-sourceaws_ami)
- [data-source/aws_batch_compute_environment](#data-sourceaws_batch_compute_environment)
- [data-source/aws_ecs_task_definition](#data-sourceaws_ecs_task_definition)
- [data-source/aws_ecs_task_execution](#data-sourceaws_ecs_task_execution)
- [data-source/aws_globalaccelerator_accelerator](#data-sourceaws_globalaccelerator_accelerator)
- [data-source/aws_launch_template](#data-sourceaws_launch_template)
- [data-source/aws_service_discovery_service](#data-sourceaws_service_discovery_service)
- [resource/aws_api_gateway_account](#resourceaws_api_gateway_account)
- [resource/aws_api_gateway_deployment](#resourceaws_api_gateway_deployment)
- [resource/aws_batch_compute_environment](#resourceaws_batch_compute_environment)
- [resource/aws_cloudfront_key_value_store](#resourceaws_cloudfront_key_value_store)
- [resource/aws_cloudfront_response_headers_policy](#resourceaws_cloudfront_response_headers_policy)
- [resource/aws_ecs_task_definition](#resourceaws_ecs_task_definition)
- [resource/aws_instance](#resourceaws_instance)
- [resource/aws_kinesis_analytics_application](#resourceaws_kinesis_analytics_application)
- [resource/aws_launch_template](#resourceaws_launch_template)
- [resource/aws_networkmanager_core_network](#resourceaws_networkmanager_core_network)
- [resource/aws_redshift_cluster](#resourceaws_redshift_cluster)
- [resource/aws_redshift_service_account](#resourceaws_redshift_service_account)
- [resource/aws_sagemaker_notebook_instance](#resourceaws_sagemaker_notebook_instance)
- [resource/aws_spot_instance_request](#resourceaws_spot_instance_request)
- [resource/aws_ssm_association](#resourceaws_ssm_association)

<!-- /TOC -->

## Provider Version Configuration

-> Before upgrading to version 6.0.0, upgrade to the most recent 5.X version of the provider and ensure that your environment successfully runs [`terraform plan`](https://www.terraform.io/docs/commands/plan.html). You should not see changes you don't expect, or deprecation notices for anything mentioned in this guide.

Use [version constraints when configuring Terraform providers](https://www.terraform.io/docs/configuration/providers.html#provider-versions). If you are following that recommendation, update the version constraints in your Terraform configuration and run [`terraform init -upgrade`](https://www.terraform.io/docs/commands/init.html) to download the new version.

For example, given this previous configuration:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.92"
    }
  }
}

provider "aws" {
  # Configuration options
}
```

Update to the latest 6.X version:

```terraform
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  # Configuration options
}
```

## Provider Arguments

Version 6.0.0 removes these `provider` arguments:

* `endpoints.opsworks` - Removed following AWS OpsWorks Stacks End of Life
* `endpoints.simpledb` and `endpoints.sdb` - Removed following dropping support for Amazon SimpleDB
* `endpoints.worklink` - Removed following dropping support for Amazon Worklink

## Dropping Support For Amazon SimpleDB

As the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/welcome.html) does not support Amazon SimpleDB, the `aws_simpledb_domain` resource has been removed.

## Dropping Support For Amazon Worklink

As the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/welcome.html) has [dropped support](https://github.com/aws/aws-sdk-go-v2/pull/2814) for Amazon Worklink, the following resources have been removed:

* `aws_worklink_fleet`
* `aws_worklink_website_certificate_authority_association`

## AWS OpsWorks Stacks End of Life

As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the following resources have been removed:

* `aws_opsworks_application`
* `aws_opsworks_custom_layer`
* `aws_opsworks_ecs_cluster_layer`
* `aws_opsworks_ganglia_layer`
* `aws_opsworks_haproxy_layer`
* `aws_opsworks_instance`
* `aws_opsworks_java_app_layer`
* `aws_opsworks_memcached_layer`
* `aws_opsworks_mysql_layer`
* `aws_opsworks_nodejs_app_layer`
* `aws_opsworks_permission`
* `aws_opsworks_php_app_layer`
* `aws_opsworks_rails_app_layer`
* `aws_opsworks_rds_db_instance`
* `aws_opsworks_stack`
* `aws_opsworks_static_web_layer`
* `aws_opsworks_user_profile`

## data-source/aws_ami

Configurations with `most_recent` set to `true` and no owner or image ID filters will now trigger an error diagnostic.
Previously, these configurations would result in only a [warning diagnostic](https://github.com/hashicorp/terraform-provider-aws/pull/40211).
To prevent this error, set the `owner` argument or include a `filter` block with an `image-id` or `owner-id` name/value pair.
To continue using unsafe filter values with `most_recent` set to `true`, set the new `allow_unsafe_filter` argument to `true`.
This is not recommended.

## data-source/aws_batch_compute_environment

`compute_environment_name` has been renamed to `name`.

## data-source/aws_globalaccelerator_accelerator

`id` is now computed only.

## data-source/aws_service_discovery_service

`tags_all` has been removed.

## data-source/aws_ecs_task_definition

Remove `inference_accelerator` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## data-source/aws_ecs_task_execution

Remove `inference_accelerator_overrides` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## data-source/aws_launch_template

Remove `elastic_inference_accelerator` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## resource/aws_api_gateway_account

`reset_on_delete` has been removed.
The destroy operation will now always reset the API Gateway account settings.
Use a [removed](https://developer.hashicorp.com/terraform/language/resources/syntax#removing-resources) block to retain the previous behavior which left the account settings unchanged upon destruction.

## resource/aws_api_gateway_deployment

The following arguments have been **removed** from the `aws_api_gateway_deployment` resource:

- `stage_name`
- `stage_description`
- `canary_settings`

Additionally, the computed attributes `invoke_url` and `execution_arn` have been removed from `aws_api_gateway_deployment`. These are now only available via the `aws_api_gateway_stage` resource.

### Migration Example

**Before (v5 and earlier, using implicit stage):**

```terraform
resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = "prod"
}
```

**After (v6+, using explicit stage):**

```terraform
resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id
}

resource "aws_api_gateway_stage" "prod" {
  stage_name    = "prod"
  rest_api_id   = aws_api_gateway_rest_api.example.id
  deployment_id = aws_api_gateway_deployment.example.id
}
```

### Importing an Implicit Stage into Terraform

If your previous configuration relied on an implicitly created stage, you can import it into a managed `aws_api_gateway_stage` resource like so:

```sh
terraform import aws_api_gateway_stage.prod <rest_api_id>/<stage_name>
```

## resource/aws_batch_compute_environment

* `compute_environment_name` has been renamed to `name`.
* `compute_environment_name_prefix` has been renamed to `name_prefix`.

## resource/aws_cloudfront_key_value_store

The `id` attribute is now set the to ID value returned by the AWS API.
For the name, use the `name` attribute.

## resource/aws_cloudfront_response_headers_policy

The `etag` argument is now computed only.

## resource/aws_ecs_task_definition

Remove `inference_accelerator` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## resource/aws_launch_template

Remove `elastic_inference_accelerator` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## resource/aws_instance

The `user_data` attribute no longer applies hashing and is now stored in clear text. **Do not include passwords or sensitive information** in `user_data`, as it will be visible in plaintext. Follow [AWS Best Practices](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) to secure your instance metadata. If you need to provide base64-encoded user data, use the `user_data_base64` attribute instead.

## resource/aws_kinesis_analytics_application

This resource is deprecated and will be removed in a future version. [Effective January 27, 2026](https://aws.amazon.com/blogs/big-data/migrate-from-amazon-kinesis-data-analytics-for-sql-to-amazon-managed-service-for-apache-flink-and-amazon-managed-service-for-apache-flink-studio/), AWS will [no longer support](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/discontinuation.html) Amazon Kinesis Data Analytics for SQL. Use the `aws_kinesisanalyticsv2_application` resource instead to manage Amazon Kinesis Data Analytics for Apache Flink applications. AWS provides guidance for migrating from [Amazon Kinesis Data Analytics for SQL Applications to Amazon Managed Service for Apache Flink Studio](https://aws.amazon.com/blogs/big-data/migrate-from-amazon-kinesis-data-analytics-for-sql-applications-to-amazon-managed-service-for-apache-flink-studio/) including [examples](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/migrating-to-kda-studio-overview.html).

## resource/aws_networkmanager_core_network

The `base_policy_region` argument has been removed. Use `base_policy_regions` instead.

## resource/aws_redshift_cluster

* The `publicly_accessible` attribute now defaults to `false`.
* Remove `snapshot_copy` from your configuration—it no longer exists. Use the `aws_redshift_snapshot_copy` resource instead.
* Remove `logging` from your configuration—it no longer exists. Use the `aws_redshift_logging` resource instead.
* Attributes `cluster_public_key`, `cluster_revision_number`, and `endpoint` are now read only and should not be set.

## resource/aws_redshift_service_account

The `aws_redshift_service_account` resource has been removed. AWS [recommends](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) that a [service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) should be used instead of an AWS account ID in any relevant IAM policy.

## resource/aws_sagemaker_notebook_instance

Remove `accelerator_types` from your configuration—it no longer exists. Instead, use `instance_type` to use [Inferentia](https://docs.aws.amazon.com/sagemaker/latest/dg/neo-supported-cloud.html).

## resource/aws_spot_instance_request

Remove `block_duration_minutes` from your configuration—it no longer exists.

## resource/aws_ssm_association

Remove `instance_id` from configuration—it no longer exists. Use `targets` instead.
