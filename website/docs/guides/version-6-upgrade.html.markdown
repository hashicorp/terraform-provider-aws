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

- [Prerequisites to Upgrade to v6.0.0](#prerequisites-to-upgrade-to-v600)
- [Removed Provider Arguments](#removed-provider-arguments)
- [Enhanced Region Support](#enhanced-region-support)
- [Amazon Elastic Transcoder Deprecation](#amazon-elastic-transcoder-deprecation)
- [CloudWatch Evidently Deprecation](#cloudwatch-evidently-deprecation)
- [Nullable Boolean Validation Update](#nullable-boolean-validation-update)
- [OpsWorks Stacks Removal](#opsworks-stacks-removal)
- [S3 Global Endpoint Deprecation](#s3-global-endpoint-deprecation)
- [SimpleDB Support Removed](#simpledb-support-removed)
- [Worklink Support Removed](#worklink-support-removed)
- [Data Source `aws_ami`](#data-source-aws_ami)
- [Data Source `aws_batch_compute_environment`](#data-source-aws_batch_compute_environment)
- [Data Source `aws_ecs_task_definition`](#data-source-aws_ecs_task_definition)
- [Data Source `aws_ecs_task_execution`](#data-source-aws_ecs_task_execution)
- [Data Source `aws_elbv2_listener_rule`](#data-source-aws_elbv2_listener_rule)
- [Data Source `aws_globalaccelerator_accelerator`](#data-source-aws_globalaccelerator_accelerator)
- [Data Source `aws_identitystore_group`](#data-source-aws_identitystore_group)
- [Data Source `aws_identitystore_user`](#data-source-aws_identitystore_user)
- [Data Source `aws_kms_secret`](#data-source-aws_kms_secret)
- [Data Source `aws_launch_template`](#data-source-aws_launch_template)
- [Data Source `aws_opensearch_domain`](#data-source-aws_opensearch_domain)
- [Data Source `aws_opensearchserverless_security_config`](#data-source-aws_opensearchserverless_security_config)
- [Data Source `aws_quicksight_data_set`](#data-source-aws_quicksight_data_set)
- [Data Source `aws_region`](#data-source-aws_region)
- [Data Source `aws_s3_bucket`](#data-source-aws_s3_bucket)
- [Data Source `aws_service_discovery_service`](#data-source-aws_service_discovery_service)
- [Data Source `aws_servicequotas_templates`](#data-source-aws_servicequotas_templates)
- [Data Source `aws_ssmincidents_replication_set`](#data-source-aws_ssmincidents_replication_set)
- [Data Source `aws_vpc_endpoint_service`](#data-source-aws_vpc_endpoint_service)
- [Data Source `aws_vpc_peering_connection`](#data-source-aws_vpc_peering_connection)
- [Resource `aws_accessanalyzer_archive_rule`](#typenullablebool-validation-update)
- [Resource `aws_alb_target_group`](#typenullablebool-validation-update)
- [Resource `aws_api_gateway_account`](#resource-aws_api_gateway_account)
- [Resource `aws_api_gateway_deployment`](#resource-aws_api_gateway_deployment)
- [Resource `aws_appflow_connector_profile`](#resource-aws_appflow_connector_profile)
- [Resource `aws_appflow_flow`](#resource-aws_appflow_flow)
- [Resource `aws_batch_compute_environment`](#resource-aws_batch_compute_environment)
- [Resource `aws_batch_job_queue`](#resource-aws_batch_job_queue)
- [Resource `aws_bedrock_model_invocation_logging_configuration`](#resource-aws_bedrock_model_invocation_logging_configuration)
- [Resource `aws_cloudformation_stack_set_instance`](#resource-aws_cloudformation_stack_set_instance)
- [Resource `aws_cloudfront_key_value_store`](#resource-aws_cloudfront_key_value_store)
- [Resource `aws_cloudfront_response_headers_policy`](#resource-aws_cloudfront_response_headers_policy)
- [Resource `aws_cloudtrail_event_data_store`](#typenullablebool-validation-update)
- [Resource `aws_cognito_user_in_group`](#resource-aws_cognito_user_in_group)
- [Resource `aws_config_aggregate_authorization`](#resource-aws_config_aggregate_authorization)
- [Resource `aws_db_instance`](#resource-aws_db_instance)
- [Resource `aws_dms_endpoint`](#resource-aws_dms_endpoint)
- [Resource `aws_dx_gateway_association`](#resource-aws_dx_gateway_association)
- [Resource `aws_dx_hosted_connection`](#resource-aws_dx_hosted_connection)
- [Resource `aws_ec2_spot_instance_fleet`](#typenullablebool-validation-update)
- [Resource `aws_ecs_task_definition`](#resource-aws_ecs_task_definition)
- [Resource `aws_eip`](#resource-aws_eip)
- [Resource `aws_eks_addon`](#resource-aws_eks_addon)
- [Resource `aws_elasticache_cluster`](#typenullablebool-validation-update)
- [Resource `aws_elasticache_replication_group`](#resource-aws_elasticache_replication_group)
- [Resource `aws_elasticache_user`](#resource-aws_elasticache_user)
- [Resource `aws_elasticache_user_group`](#resource-aws_elasticache_user_group)
- [Resource `aws_evidently_feature`](#typenullablebool-validation-update)
- [Resource `aws_flow_log`](#resource-aws_flow_log)
- [Resource `aws_guardduty_detector`](#resource-aws_guardduty_detector)
- [Resource `aws_guardduty_organization_configuration`](#resource-aws_guardduty_organization_configuration)
- [Resource `aws_imagebuilder_container_recipe`](#typenullablebool-validation-update)
- [Resource `aws_imagebuilder_image_recipe`](#typenullablebool-validation-update)
- [Resource `aws_instance`](#resource-aws_instance)
- [Resource `aws_kinesis_analytics_application`](#resource-aws_kinesis_analytics_application)
- [Resource `aws_launch_template`](#resource-aws_launch_template)
- [Resource `aws_lb_listener`](#resource-aws_lb_listener)
- [Resource `aws_lb_target_group`](#typenullablebool-validation-update)
- [Resource `aws_media_store_container`](#resource-aws_media_store_container)
- [Resource `aws_media_store_container_policy`](#resource-aws_media_store_container_policy)
- [Resource `aws_mq_broker`](#typenullablebool-validation-update)
- [Resource `aws_networkmanager_core_network`](#resource-aws_networkmanager_core_network)
- [Resource `aws_opensearch_domain`](#resource-aws_opensearch_domain)
- [Resource `aws_opensearchserverless_security_config`](#resource-aws_opensearchserverless_security_config)
- [Resource `aws_paymentcryptography_key`](#resource-aws_paymentcryptography_key)
- [Resource `aws_redshift_cluster`](#resource-aws_redshift_cluster)
- [Resource `aws_redshift_service_account`](#resource-aws_redshift_service_account)
- [Resource `aws_rekognition_stream_processor`](#resource-aws_rekognition_stream_processor)
- [Resource `aws_resiliencehub_resiliency_policy`](#resource-aws_resiliencehub_resiliency_policy)
- [Resource `aws_s3_bucket`](#resource-aws_s3_bucket)
- [Resource `aws_sagemaker_image_version`](#resource-aws_sagemaker_image_version)
- [Resource `aws_sagemaker_notebook_instance`](#resource-aws_sagemaker_notebook_instance)
- [Resource `aws_servicequotas_template`](#resource-aws_servicequotas_template)
- [Resource `aws_spot_instance_request`](#resource-aws_spot_instance_request)
- [Resource `aws_ssm_association`](#resource-aws_ssm_association)
- [Resource `aws_ssmincidents_replication_set`](#resource-aws_ssmincidents_replication_set)
- [Resource `aws_verifiedpermissions_schema`](#resource-aws_verifiedpermissions_schema)
- [Resource `aws_wafv2_web_acl`](#resource-aws_wafv2_web_acl)

<!-- /TOC -->

## Prerequisites to Upgrade to v6.0.0

-> Before upgrading to version `6.0.0`, first upgrade to the latest available `5.x` version of the provider. Run [`terraform plan`](https://developer.hashicorp.com/terraform/cli/commands/plan) and confirm that:

- Your plan completes without errors or unexpected changes.
- There are no deprecation warnings related to the changes described in this guide.

If you use [version constraints](https://developer.hashicorp.com/terraform/language/providers/requirements#provider-versions) (recommended), update them to allow the `6.x` series and run [`terraform init -upgrade`](https://developer.hashicorp.com/terraform/cli/commands/init) to download the new version.

### Example

**Before:**

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

**After:**

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

## Removed Provider Arguments

Remove the following from your provider configuration—they are no longer supported:

- `endpoints.opsworks` – removed following AWS OpsWorks Stacks End of Life.
- `endpoints.simpledb` and `endpoints.sdb` – removed due to the removal of Amazon SimpleDB support.
- `endpoints.worklink` – removed due to the removal of Amazon Worklink support.

## Enhanced Region Support

Version 6.0.0 adds `region` to most resources making it significantly easier to manage infrastructure across AWS Regions without requiring multiple provider configurations. See [Enhanced Region Support](enhanced-region-support.html).

## Amazon Elastic Transcoder Deprecation

Amazon Elastic Transcoder will be [discontinued](https://aws.amazon.com/blogs/media/support-for-amazon-elastic-transcoder-ending-soon/) on **November 13, 2025**.

The following resources are deprecated and will be removed in a future major release:

- `aws_elastictranscoder_pipeline`
- `aws_elastictranscoder_preset`

Use [AWS Elemental MediaConvert](https://aws.amazon.com/blogs/media/migrating-workflows-from-amazon-elastic-transcoder-to-aws-elemental-mediaconvert/) instead.

## CloudWatch Evidently Deprecation

AWS will [end support](https://aws.amazon.com/blogs/mt/support-for-amazon-cloudwatch-evidently-ending-soon/) for CloudWatch Evidently on **October 17, 2025**.

The following resources are deprecated and will be removed in a future major release:

- `aws_evidently_feature`
- `aws_evidently_launch`
- `aws_evidently_project`
- `aws_evidently_segment`

Migrate to [AWS AppConfig Feature Flags](https://aws.amazon.com/blogs/mt/using-aws-appconfig-feature-flags/).

## Nullable Boolean Validation Update

Update your configuration to _only_ use `""`, `true`, or `false` if you use the arguments below _and_ you are using `0` or `1` to represent boolean values:

| Resource                                | Attribute(s)                                                             |
|-----------------------------------------|--------------------------------------------------------------------------|
| `aws_accessanalyzer_archive_rule`       | `filter.exists`                                                          |
| `aws_alb_target_group`                  | `preserve_client_ip`                                                     |
| `aws_cloudtrail_event_data_store`       | `suspend`                                                                |
| `aws_ec2_spot_instance_fleet`           | `terminate_instances_on_delete`                                          |
| `aws_elasticache_cluster`               | `auto_minor_version_upgrade`                                             |
| `aws_elasticache_replication_group`     | `at_rest_encryption_enabled`, `auto_minor_version_upgrade`               |
| `aws_evidently_feature`                 | `variations.value.bool_value`                                            |
| `aws_imagebuilder_container_recipe`     | `instance_configuration.block_device_mapping.ebs.delete_on_termination`, `instance_configuration.block_device_mapping.ebs.encrypted` |
| `aws_imagebuilder_image_recipe`         | `block_device_mapping.ebs.delete_on_termination`, `block_device_mapping.ebs.encrypted` |
| `aws_launch_template`                   | `block_device_mappings.ebs.delete_on_termination`, `block_device_mappings.ebs.encrypted`, `ebs_optimized`, `network_interfaces.associate_carrier_ip_address`, `network_interfaces.associate_public_ip_address`, `network_interfaces.delete_on_termination`, `network_interfaces.primary_ipv6` |
| `aws_lb_target_group`                   | `preserve_client_ip`                                                     |
| `aws_mq_broker`                         | `logs.audit`                                                             |

This is due to changes to `TypeNullableBool`.

## OpsWorks Stacks Removal

The AWS OpsWorks Stacks service has reached [End of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html). The following resources have been removed:

- `aws_opsworks_application`
- `aws_opsworks_custom_layer`
- `aws_opsworks_ecs_cluster_layer`
- `aws_opsworks_ganglia_layer`
- `aws_opsworks_haproxy_layer`
- `aws_opsworks_instance`
- `aws_opsworks_java_app_layer`
- `aws_opsworks_memcached_layer`
- `aws_opsworks_mysql_layer`
- `aws_opsworks_nodejs_app_layer`
- `aws_opsworks_permission`
- `aws_opsworks_php_app_layer`
- `aws_opsworks_rails_app_layer`
- `aws_opsworks_rds_db_instance`
- `aws_opsworks_stack`
- `aws_opsworks_static_web_layer`
- `aws_opsworks_user_profile`

## SimpleDB Support Removed

The `aws_simpledb_domain` resource has been removed, as the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/welcome.html) no longer supports Amazon SimpleDB.

## Worklink Support Removed

The following resources have been removed due to dropped support for Amazon Worklink in the [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2/pull/2814):

- `aws_worklink_fleet`
- `aws_worklink_website_certificate_authority_association`

## S3 Global Endpoint Deprecation

Support for the global S3 endpoint is deprecated. This affects S3 resources in `us-east-1` (excluding directory buckets) when `s3_us_east_1_regional_endpoint` is set to `legacy`.

`s3_us_east_1_regional_endpoint` will be removed in `v7.0.0`.

To prepare:

- Remove `s3_us_east_1_regional_endpoint` from your provider configuration, **or**
- Set its value to `regional` and verify functionality.

## Data Source `aws_ami`

When using `most_recent = true`, your configuration **must now include** an `owner` or a `filter` that identifies the image by `image-id` or `owner-id`.

- **Before (v5 and earlier):**
  Terraform allowed this setup and showed only a warning.

- **Now (v6+):**
  Terraform will stop with an **error** to prevent unsafe or ambiguous AMI lookups.

### How to fix it

Do one of the following:

- Add `owner`:

```terraform
owner = "amazon"
```

- Or add a `filter` block that includes either `image-id` or `owner-id`:

```terraform
filter {
  name   = "owner-id"
  values = ["123456789012"]
}
```

### Unsafe option (not recommended)

To override this check, you can set:

```terraform
allow_unsafe_filter = true
```

However, this may lead to unreliable results and should be avoided unless absolutely necessary.

## Data Source `aws_batch_compute_environment`

`compute_environment_name` has been renamed to `name`.

Update your configurations to replace any usage of `compute_environment_name` with `name` to use this version.

## Data Source `aws_ecs_task_definition`

Remove `inference_accelerator`—it is no longer supported. Amazon Elastic Inference reached end of life in April 2024.

## Data Source `aws_ecs_task_execution`

Remove `inference_accelerator_overrides`—it is no longer supported. Amazon Elastic Inference reached end of life in April 2024.

## Data Source `aws_elbv2_listener_rule`

Treat the following as lists of nested blocks instead of single-nested blocks:

- `action.authenticate_cognito`
- `action.authenticate_oidc`
- `action.fixed_response`
- `action.forward`
- `action.forward.stickiness`
- `action.redirect`
- `condition.host_header`
- `condition.http_header`
- `condition.http_request_method`
- `condition.path_pattern`
- `condition.query_string`
- `condition.source_ip`

The data source configuration itself does not change. However, now, include an index when referencing them. For example, update `action[0].authenticate_cognito.scope` to `action[0].authenticate_cognito[0].scope`.

## Data Source `aws_globalaccelerator_accelerator`

`id` is now **computed only** and can no longer be set manually.
If your configuration explicitly attempts to set a value for `id`, you must remove it to avoid an error.

## Data Source `aws_identitystore_group`

Remove `filter`—it is no longer supported. To locate a group, update your configuration to use `alternate_identifier` instead.

## Data Source `aws_identitystore_user`

Remove `filter`—it is no longer supported.
To locate a user, update your configuration to use `alternate_identifier` instead.

## Data Source `aws_kms_secret`

The functionality for this data source was removed in **v2.0.0** and the data source will be removed in a future version.

## Data Source `aws_launch_template`

Remove the following—they are no longer supported:

- `elastic_gpu_specifications`: Amazon Elastic Graphics reached end of life in January 2024.
- `elastic_inference_accelerator`: Amazon Elastic Inference reached end of life in April 2024.

## Data Source `aws_opensearch_domain`

Remove `kibana_endpoint`—it is no longer supported. AWS OpenSearch Service no longer uses Kibana endpoints. The service now uses **Dashboards**, accessible at the `/_dashboards/` path on the domain endpoint.
For more details, refer to the [AWS OpenSearch Dashboards documentation](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/dashboards.html).

## Data Source `aws_opensearchserverless_security_config`

Treat `saml_options` as a list of nested blocks instead of a single-nested block. The data source configuration itself does not change. However, now, include an index when referencing it. For example, update `saml_options.session_timeout` to `saml_options[0].session_timeout`.

## Data Source `aws_quicksight_data_set`

Remove `tags_all`—it is no longer supported.

## Data Source `aws_region`

`name` has been deprecated. Use `region` instead.

## Data Source `aws_s3_bucket`

`bucket_region` has been added and should be used instead of `region`, which is now used for [Enhanced Region Support](enhanced-region-support.html).

## Data Source `aws_service_discovery_service`

Remove `tags_all`—it is no longer supported.

## Data Source `aws_servicequotas_templates`

`region` has been deprecated. Use `aws_region` instead.

## Data Source `aws_ssmincidents_replication_set`

`region` has been deprecated. Use `regions` instead.

## Data Source `aws_vpc_endpoint_service`

`region` has been deprecated. Use `service_region` instead.

## Data Source `aws_vpc_peering_connection`

`region` has been deprecated. Use `requester_region` instead.

## Resource `aws_api_gateway_account`

Remove `reset_on_delete`—it is no longer supported. The destroy operation will now always reset the API Gateway account settings by default.

If you want to retain the previous behavior (where the account settings were not changed upon destruction), use a `removed` block in your configuration. For more details, see the [removing resources documentation](https://developer.hashicorp.com/terraform/language/resources/syntax#removing-resources).

## Resource `aws_api_gateway_deployment`

* Use the `aws_api_gateway_stage` resource if your configuration uses any of the following, which have been removed from the `aws_api_gateway_deployment` resource:
    - `stage_name`
    - `stage_description`
    - `canary_settings`
* Remove `invoke_url` and `execution_arn`—they are no longer supported. Use the `aws_api_gateway_stage` resource instead.

### Migration Example

**Before (v5 and earlier, using implicit stage):**

```terraform
resource "aws_api_gateway_deployment" "example" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  stage_name  = "prod"
}
```

**After (v6+, using explicit stage):**

If your previous configuration relied on an implicitly created stage, you must now define and manage that stage explicitly using the `aws_api_gateway_stage` resource. To do this, create a corresponding resource and import the existing stage into your configuration.

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

Import the existing stage, replacing `rest_api_id` and `stage_name` with your values:

```sh
terraform import aws_api_gateway_stage.prod rest_api_id/stage_name
```

## Resource `aws_appflow_connector_profile`

Importing an `aws_appflow_connector_profile` resource now uses the `name` of the Connector Profile.

## Resource `aws_appflow_flow`

Importing an `aws_appflow_flow` resource now uses the `name` of the Flow.

## Resource `aws_batch_compute_environment`

Replace any usage of `compute_environment_name` with `name` and `compute_environment_name_prefix` with `name_prefix` as they have been renamed.

## Resource `aws_batch_job_queue`

Remove `compute_environments`—it is no longer supported.
Use `compute_environment_order` configuration blocks instead. While you must update your configuration, Terraform will upgrade states with `compute_environments` to `compute_environment_order`.

**Before (v5 and earlier):**

```terraform
resource "aws_batch_job_queue" "example" {
  compute_environments = [aws_batch_compute_environment.example.arn]
  name                 = "patagonia"
  priority             = 1
  state                = "ENABLED"
}
```

**After (v6+):**

```terraform
resource "aws_batch_job_queue" "example" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.example.arn
    order               = 0
  }
  name     = "patagonia"
  priority = 1
  state    = "ENABLED"
}
```

## Resource `aws_bedrock_model_invocation_logging_configuration`

Treat the following as lists of nested blocks instead of single-nested blocks:

- `logging_config`
- `logging_config.cloudwatch_config`
- `logging_config.cloudwatch_config.large_data_delivery_s3_config`
- `logging_config.s3_config`

The resource configuration itself does not change, but you must now include an index when referencing them. For example, update `logging_config.cloudwatch_config.log_group_name` to `logging_config[0].cloudwatch_config[0].log_group_name`.

## Resource `aws_cloudformation_stack_set_instance`

`region` has been deprecated. Use `stack_set_instance_region` instead.

## Resource `aws_cloudfront_key_value_store`

Use `name` to reference the resource name. `id` represents the ID value returned by the AWS API.

## Resource `aws_cloudfront_response_headers_policy`

Do not set a value for `etag` as it is now computed only.

## Resource `aws_cognito_user_in_group`

For the `id`, use a comma-delimited string concatenating `user_pool_id`, `group_name`, and `username`. For example, in an import command, use comma-delimiting for the composite `id`.

## Resource `aws_config_aggregate_authorization`

`region` has been deprecated. Use `authorized_aws_region` instead.

## Resource `aws_db_instance`

Do not use `character_set_name` with `replicate_source_db`, `restore_to_point_in_time`, `s3_import`, or `snapshot_identifier`. The combination is no longer valid.

## Resource `aws_dms_endpoint`

`s3_settings` has been removed. Use the `aws_dms_s3_endpoint` resource rather than `s3_settings` of `aws_dms_endpoint`.

## Resource `aws_dx_gateway_association`

Remove `vpn_gateway_id`—it is no longer supported. Use `associated_gateway_id` instead.

## Resource `aws_dx_hosted_connection`

`region` has been deprecated. Use `connection_region` instead.

## Resource `aws_ecs_task_definition`

Remove `inference_accelerator`—it is no longer supported. Amazon Elastic Inference reached end of life in April 2024.

## Resource `aws_eip`

Remove `vpc`—it is no longer supported. Use `domain` instead.

## Resource `aws_eks_addon`

Remove `resolve_conflicts`—it is no longer supported. Use `resolve_conflicts_on_create` and `resolve_conflicts_on_update` instead.

## Resource `aws_elasticache_replication_group`

* `auth_token_update_strategy` no longer has a default value. If `auth_token` is set, it must also be explicitly configured.
* The ability to provide an uppercase `engine` value is deprecated. In `v7.0.0`, plan-time validation of `engine` will require an entirely lowercase value to match the returned value from the AWS API without diff suppression.
* See also [changes](#typenullablebool-validation-update) to `at_rest_encryption_enabled` and `auto_minor_version_upgrade`.

## Resource `aws_elasticache_user`

The ability to provide an uppercase `engine` value is deprecated.
In `v7.0.0`, plan-time validation of `engine` will require an entirely lowercase value to match the returned value from the AWS API without diff suppression.

## Resource `aws_elasticache_user_group`

The ability to provide an uppercase `engine` value is deprecated.
In `v7.0.0`, plan-time validation of `engine` will require an entirely lowercase value to match the returned value from the AWS API without diff suppression.

## Resource `aws_flow_log`

Remove `log_group_name`—it is no longer supported. Use `log_destination` instead.

## Resource `aws_guardduty_detector`

`datasources` is deprecated.
Use the `aws_guardduty_detector_feature` resource instead.

## Resource `aws_guardduty_organization_configuration`

* Remove `auto_enable`—it is no longer supported.
* `auto_enable_organization_members` is now required.
* `datasources` is deprecated.

## Resource `aws_instance`

* `user_data` no longer applies hashing and is now stored in clear text. **Do not include passwords or sensitive information** in `user_data`, as it will be visible in plaintext. Follow [AWS Best Practices](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) to secure your instance metadata. If you need to provide base64-encoded user data, use `user_data_base64` instead.
* Remove `cpu_core_count` and `cpu_threads_per_core`—they are no longer supported. Instead, use the `cpu_options` configuration block with `core_count` and `threads_per_core`.

## Resource `aws_kinesis_analytics_application`

This resource is deprecated and will be removed in a future version. [Effective January 27, 2026](https://aws.amazon.com/blogs/big-data/migrate-from-amazon-kinesis-data-analytics-for-sql-to-amazon-managed-service-for-apache-flink-and-amazon-managed-service-for-apache-flink-studio/), AWS will [no longer support](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/discontinuation.html) Amazon Kinesis Data Analytics for SQL. Use the `aws_kinesisanalyticsv2_application` resource instead to manage Amazon Kinesis Data Analytics for Apache Flink applications. AWS provides guidance for migrating from [Amazon Kinesis Data Analytics for SQL Applications to Amazon Managed Service for Apache Flink Studio](https://aws.amazon.com/blogs/big-data/migrate-from-amazon-kinesis-data-analytics-for-sql-applications-to-amazon-managed-service-for-apache-flink-studio/) including [examples](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/migrating-to-kda-studio-overview.html).

## Resource `aws_launch_template`

* Remove `elastic_gpu_specifications`—it is no longer supported. Amazon Elastic Graphics reached end of life in January 2024.
* Remove `elastic_inference_accelerator`—it is no longer supported. Amazon Elastic Inference reached end of life in April 2024.
* See also [changes](#typenullablebool-validation-update) to `block_device_mappings.ebs.delete_on_termination`, `block_device_mappings.ebs.encrypted`, `ebs_optimized`, `network_interfaces.associate_carrier_ip_address`, `network_interfaces.associate_public_ip_address`, `network_interfaces.delete_on_termination`, and `network_interfaces.primary_ipv6`.

## Resource `aws_lb_listener`

* For `mutual_authentication`, `advertise_trust_store_ca_names`, `ignore_client_certificate_expiry`, and `trust_store_arn` can now only be set when `mode` is `verify`.
* `trust_store_arn` is required when `mode` is `verify`.

## Resource `aws_media_store_container`

This resource is deprecated and will be removed in a future version. AWS has [announced](https://aws.amazon.com/blogs/media/support-for-aws-elemental-mediastore-ending-soon/) the discontinuation of AWS Elemental MediaStore, effective November 13, 2025. Users should begin transitioning to alternative solutions as soon as possible. For simple live streaming workflows, AWS recommends migrating to Amazon S3. For advanced use cases that require features such as packaging, DRM, or cross-region redundancy, consider using AWS Elemental MediaPackage.

## Resource `aws_media_store_container_policy`

This resource is deprecated and will be removed in a future version. AWS has [announced](https://aws.amazon.com/blogs/media/support-for-aws-elemental-mediastore-ending-soon/) the discontinuation of AWS Elemental MediaStore, effective November 13, 2025. Users should begin transitioning to alternative solutions as soon as possible. For simple live streaming workflows, AWS recommends migrating to Amazon S3. For advanced use cases that require features such as packaging, DRM, or cross-region redundancy, consider using AWS Elemental MediaPackage.

## Resource `aws_networkmanager_core_network`

Remove `base_policy_region`—it is no longer supported. Use `base_policy_regions` instead.

## Resource `aws_opensearch_domain`

Remove `kibana_endpoint`—it is no longer supported. AWS OpenSearch Service does not use Kibana endpoints (i.e., `_plugin/kibana`). Instead, OpenSearch uses Dashboards, accessible at the path `/_dashboards/` on the domain endpoint. The terminology has shifted from “Kibana” to “Dashboards.”

For more information, see the [AWS OpenSearch Dashboards documentation](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/dashboards.html).

## Resource `aws_opensearchserverless_security_config`

Treat `saml_options` as a list of nested blocks instead of a single-nested block. The resource configuration itself does not change. However, now, include an index when referencing it. For example, update `saml_options.session_timeout` to `saml_options[0].session_timeout`.

## Resource `aws_paymentcryptography_key`

Treat the `key_attributes` and `key_attributes.key_modes_of_use` as lists of nested blocks instead of single-nested blocks. The resource configuration itself does not change. However, now, include an index when referencing them. For example, update `key_attributes.key_modes_of_use.decrypt` to `key_attributes[0].key_modes_of_use[0].decrypt`.

## Resource `aws_redshift_cluster`

* `encrypted` now defaults to `true`.
* `publicly_accessible` now defaults to `false`.
* Remove `snapshot_copy`—it is no longer supported. Use the `aws_redshift_snapshot_copy` resource instead.
* Remove `logging`—it is no longer supported. Use the `aws_redshift_logging` resource instead.
* `cluster_public_key`, `cluster_revision_number`, and `endpoint` are now read only and should not be set.

## Resource `aws_redshift_service_account`

The `aws_redshift_service_account` resource has been removed. AWS [recommends](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) that a [service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) should be used instead of an AWS account ID in any relevant IAM policy.

## Resource `aws_rekognition_stream_processor`

Treat `regions_of_interest.bounding_box` as a list of nested blocks instead of a single-nested block. The resource configuration itself does not change. However, now, include an index when referencing it. For example, update `regions_of_interest[0].bounding_box.height` to `regions_of_interest[0].bounding_box[0].height`.

## Resource `aws_resiliencehub_resiliency_policy`

Treat the following as lists of nested blocks instead of single-nested blocks:

- `policy`
- `policy.az`
- `policy.hardware`
- `policy.software`
- `policy.region`

The resource configuration itself does not change. However, now, include an index when referencing them. For example, update `policy.az.rpo` to `policy[0].az[0].rpo`.

## Resource `aws_s3_bucket`

`bucket_region` has been added and should be used instead of `region`, which is now used for [Enhanced Region Support](enhanced-region-support.html).

## Resource `aws_sagemaker_image_version`

For the `id`, use a comma-delimited string concatenating `image_name` and `version`. For example, in an import command, use comma-delimiting for the composite `id`.
Use `image_name` to reference the image name.

## Resource `aws_sagemaker_notebook_instance`

Remove `accelerator_types`—it is no longer supported. Instead, use `instance_type` to use [Inferentia](https://docs.aws.amazon.com/sagemaker/latest/dg/neo-supported-cloud.html).

## Resource `aws_servicequotas_template`

`region` has been deprecated. Use `aws_region` instead.

## Resource `aws_spot_instance_request`

Remove `block_duration_minutes`—it is no longer supported.

## Resource `aws_ssm_association`

Remove `instance_id`—it is no longer supported. Use `targets` instead.

## Resource `aws_ssmincidents_replication_set`

`region` has been deprecated. Use `regions` instead.

## Resource `aws_verifiedpermissions_schema`

Treat `definition` as a list of nested blocks instead of a single-nested block. The resource configuration itself does not change. However, now, include an index when referencing it. For example, update `definition.value` to `definition[0].value`.

## Resource `aws_wafv2_web_acl`

The default value for `rule.statement.managed_rule_group_statement.managed_rule_group_configs.aws_managed_rules_bot_control_rule_set.enable_machine_learning` is now `false`.
To retain the previous behavior where the argument was omitted, explicitly set the value to `true`.
