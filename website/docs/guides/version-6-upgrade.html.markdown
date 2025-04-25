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
- [AWS CloudWatch Evidently Deprecation](#aws-cloudwatch-evidently-deprecation)
- [Amazon Elastic Transcoder Deprecation](#amazon-elastic-transcoder-deprecation)
- [data-source/aws_ami](#data-sourceaws_ami)
- [data-source/aws_batch_compute_environment](#data-sourceaws_batch_compute_environment)
- [data-source/aws_ecs_task_definition](#data-sourceaws_ecs_task_definition)
- [data-source/aws_ecs_task_execution](#data-sourceaws_ecs_task_execution)
- [data-source/aws_elbv2_listener_rule](#data-sourceaws_elbv2_listener_rule)
- [data-source/aws_globalaccelerator_accelerator](#data-sourceaws_globalaccelerator_accelerator)
- [data-source/aws_identitystore_user](#data-sourceaws_identitystore_group)
- [data-source/aws_identitystore_user](#data-sourceaws_identitystore_user)
- [data-source/aws_launch_template](#data-sourceaws_launch_template)
- [data-source/aws_opensearch_domain](#data-sourceaws_opensearch_domain)
- [data-source/aws_opensearchserverless_security_config](#data-sourceaws_opensearchserverless_security_config)
- [data-source/aws_quicksight_data_set](#data-sourceaws_quicksight_data_set)
- [data-source/aws_service_discovery_service](#data-sourceaws_service_discovery_service)
- [resource/aws_api_gateway_account](#resourceaws_api_gateway_account)
- [resource/aws_api_gateway_deployment](#resourceaws_api_gateway_deployment)
- [resource/aws_batch_compute_environment](#resourceaws_batch_compute_environment)
- [resource/aws_bedrock_model_invocation_logging_configuration](#resourceaws_bedrock_model_invocation_logging_configuration)
- [resource/aws_cloudfront_key_value_store](#resourceaws_cloudfront_key_value_store)
- [resource/aws_cloudfront_response_headers_policy](#resourceaws_cloudfront_response_headers_policy)
- [resource/aws_db_instance](#resourceaws_db_instance)
- [resource/aws_dms_endpoint](#resourceaws_dms_endpoint)
- [resource/aws_dx_gateway_association](#resourceaws_dx_gateway_association)
- [resource/aws_ecs_task_definition](#resourceaws_ecs_task_definition)
- [resource/aws_eip](#resourceaws_eip)
- [resource/aws_elasticache_replication_group](#resourceaws_elasticache_replication_group)
- [resource/aws_eks_addon](#resourceaws_eks_addon)
- [resource/aws_flow_log](#resourceaws_flow_log)
- [resource/aws_guardduty_organization_configuration](#resourceaws_guardduty_organization_configuration)
- [resource/aws_instance](#resourceaws_instance)
- [resource/aws_kinesis_analytics_application](#resourceaws_kinesis_analytics_application)
- [resource/aws_launch_template](#resourceaws_launch_template)
- [resource/aws_media_store_container](#resourceaws_media_store_container)
- [resource/aws_media_store_container_policy](#resourceaws_media_store_container_policy)
- [resource/aws_networkmanager_core_network](#resourceaws_networkmanager_core_network)
- [resource/aws_opensearch_domain](#resourceaws_opensearch_domain)
- [resource/aws_opensearchserverless_security_config](#resourceaws_opensearchserverless_security_config)
- [resource/aws_paymentcryptography_key](#resourceaws_paymentcryptography_key)
- [resource/aws_redshift_cluster](#resourceaws_redshift_cluster)
- [resource/aws_redshift_service_account](#resourceaws_redshift_service_account)
- [resource/aws_rekognition_stream_processor](#resourceaws_rekognition_stream_processor)
- [resource/aws_resiliencehub_resiliency_policy](#resourceaws_resiliencehub_resiliency_policy)
- [resource/aws_sagemaker_notebook_instance](#resourceaws_sagemaker_notebook_instance)
- [resource/aws_spot_instance_request](#resourceaws_spot_instance_request)
- [resource/aws_ssm_association](#resourceaws_ssm_association)
- [resource/aws_verifiedpermissions_schema](#resourceaws_verifiedpermissions_schema)

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

## AWS CloudWatch Evidently Deprecation

Effective October 17, 2025, AWS will [no longer support Cloudwatch Evidently](https://aws.amazon.com/blogs/mt/support-for-amazon-cloudwatch-evidently-ending-soon/).
The following resources have been deprecated and will be removed in a future major version.

* `aws_evidently_feature`
* `aws_evidently_launch`
* `aws_evidently_project`
* `aws_evidently_segment`

Use [AWS AppConfig Feature Flags](https://aws.amazon.com/blogs/mt/using-aws-appconfig-feature-flags/) instead.

## Amazon Elastic Transcoder Deprecation

AWS has made the decision to [discontinue Amazon Elastic Transcoder](https://aws.amazon.com/blogs/media/support-for-amazon-elastic-transcoder-ending-soon/), effective November 13, 2025.
The following resources have been deprecated and will be removed in a future major version.

* `aws_elastictranscoder_pipeline`
* `aws_elastictranscoder_preset`

Use [AWS Elemental MediaConvert](https://aws.amazon.com/blogs/media/migrating-workflows-from-amazon-elastic-transcoder-to-aws-elemental-mediaconvert/) instead.

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

## data-source/aws_identitystore_group

`filter` has been removed.
Use `alternate_identifier` instead.

## data-source/aws_identitystore_user

`filter` has been removed.
Use `alternate_identifier` instead.

## data-source/aws_opensearchserverless_security_config

The `saml_options` attribute is now a list nested block instead of a single nested block.
When referencing this attribute, the index must now be included in the attribute address.
For example, `saml_options.session_timeout` would now be referenced as `saml_options[0].session_timeout`.

## data-source/aws_quicksight_data_set

`tags_all` has been removed.

## data-source/aws_service_discovery_service

`tags_all` has been removed.

## data-source/aws_ecs_task_definition

Remove `inference_accelerator` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## data-source/aws_ecs_task_execution

Remove `inference_accelerator_overrides` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## data-source/aws_elbv2_listener_rule

The following attributes are now list nested blocks instead of single nested blocks.

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

When referencing these attributes, the indices must now be included in the attribute address.
For example, `action[0].authenticate_cognito.scope` would now be referenced as `action[0].authenticate_cognito[0].scope`.

## data-source/aws_launch_template

Remove `elastic_gpu_specifications` from your configuration—it no longer exists. Amazon Elastic Graphics reached end of life in January 2024.

Remove `elastic_inference_accelerator` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## data-source/aws_opensearch_domain

Remove `kibana_endpoint` from your configuration—it no longer exists. AWS OpenSearch Service does **not** use Kibana endpoints (i.e., `_plugin/kibana`). Instead, OpenSearch uses **Dashboards**, accessible at the path `/_dashboards/` on the domain endpoint. The terminology has shifted from “Kibana” to “Dashboards.”

For more information, see the [AWS OpenSearch Dashboards documentation](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/dashboards.html).

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

## resource/aws_bedrock_model_invocation_logging_configuration

The following arguments are now list nested blocks instead of single nested blocks.

- `logging_config`
- `logging_config.cloudwatch_config`
- `logging_config.cloudwatch_config.large_data_delivery_s3_config`
- `logging_config.s3_config`

When referencing these arguments, the indices must now be included in the attribute address.
For example, `logging_config.cloudwatch_config.log_group_name` would now be referenced as `logging_config[0].cloudwatch_config[0].log_group_name`.

## resource/aws_cloudfront_key_value_store

The `id` attribute is now set the to ID value returned by the AWS API.
For the name, use the `name` attribute.

## resource/aws_cloudfront_response_headers_policy

The `etag` argument is now computed only.

## resource/aws_db_instance

The `character_set_name` now cannot be set with `replicate_source_db`, `restore_to_point_in_time`, `s3_import`, or `snapshot_identifier`.

## resource/aws_dms_endpoint

`s3_settings` has been removed. Use `aws_dms_s3_endpoint` instead.

## resource/aws_dx_gateway_association

The `vpn_gateway_id` attribute has been removed.
Use the `associated_gateway_id` attribute instead.

## resource/aws_ecs_task_definition

Remove `inference_accelerator` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## resource/aws_eip

The `vpc` argument has been removed.
Use `domain` instead.

## resource/aws_elasticache_replication_group

The `auth_token_update_strategy` argument no longer has a default value.
If `auth_token` is set, this argument must also be explicitly configured.

## resource/aws_eks_addon

The `resolve_conflicts` argument has been removed. Use `resolve_conflicts_on_create` and `resolve_conflicts_on_update` instead.

## resource/aws_guardduty_organization_configuration

The `auto_enable` attribute has been removed and the `auto_enable_organization_members` attribute is now required.

`datasources` now returns a deprecation warning.

## resource/aws_launch_template

Remove `elastic_gpu_specifications` from your configuration—it no longer exists. Amazon Elastic Graphics reached end of life in January 2024.

Remove `elastic_inference_accelerator` from your configuration—it no longer exists. Amazon Elastic Inference reached end of life in April 2024.

## resource/aws_media_store_container

This resource is deprecated and will be removed in a future version. AWS has [announced](https://aws.amazon.com/blogs/media/support-for-aws-elemental-mediastore-ending-soon/) the discontinuation of AWS Elemental MediaStore, effective November 13, 2025. Users should begin transitioning to alternative solutions as soon as possible. For simple live streaming workflows, AWS recommends migrating to Amazon S3. For advanced use cases that require features such as packaging, DRM, or cross-region redundancy, consider using AWS Elemental MediaPackage.

## resource/aws_media_store_container_policy

This resource is deprecated and will be removed in a future version. AWS has [announced](https://aws.amazon.com/blogs/media/support-for-aws-elemental-mediastore-ending-soon/) the discontinuation of AWS Elemental MediaStore, effective November 13, 2025. Users should begin transitioning to alternative solutions as soon as possible. For simple live streaming workflows, AWS recommends migrating to Amazon S3. For advanced use cases that require features such as packaging, DRM, or cross-region redundancy, consider using AWS Elemental MediaPackage.

## resource/aws_flow_log

Remove `log_group_name` from your configuration—it no longer exists.
Use `log_destination` instead.

## resource/aws_instance

* The `user_data` attribute no longer applies hashing and is now stored in clear text. **Do not include passwords or sensitive information** in `user_data`, as it will be visible in plaintext. Follow [AWS Best Practices](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html) to secure your instance metadata. If you need to provide base64-encoded user data, use the `user_data_base64` attribute instead.
* Remove `cpu_core_count` and `cpu_threads_per_core` from your configuration—they no longer exist. Instead, use the `cpu_options` configuration block with `core_count` and `threads_per_core`.

## resource/aws_kinesis_analytics_application

This resource is deprecated and will be removed in a future version. [Effective January 27, 2026](https://aws.amazon.com/blogs/big-data/migrate-from-amazon-kinesis-data-analytics-for-sql-to-amazon-managed-service-for-apache-flink-and-amazon-managed-service-for-apache-flink-studio/), AWS will [no longer support](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/discontinuation.html) Amazon Kinesis Data Analytics for SQL. Use the `aws_kinesisanalyticsv2_application` resource instead to manage Amazon Kinesis Data Analytics for Apache Flink applications. AWS provides guidance for migrating from [Amazon Kinesis Data Analytics for SQL Applications to Amazon Managed Service for Apache Flink Studio](https://aws.amazon.com/blogs/big-data/migrate-from-amazon-kinesis-data-analytics-for-sql-applications-to-amazon-managed-service-for-apache-flink-studio/) including [examples](https://docs.aws.amazon.com/kinesisanalytics/latest/dev/migrating-to-kda-studio-overview.html).

## resource/aws_networkmanager_core_network

The `base_policy_region` argument has been removed. Use `base_policy_regions` instead.

## resource/aws_opensearch_domain

Remove `kibana_endpoint` from your configuration—it no longer exists. AWS OpenSearch Service does **not** use Kibana endpoints (i.e., `_plugin/kibana`). Instead, OpenSearch uses **Dashboards**, accessible at the path `/_dashboards/` on the domain endpoint. The terminology has shifted from “Kibana” to “Dashboards.”

For more information, see the [AWS OpenSearch Dashboards documentation](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/dashboards.html).

## resource/aws_opensearchserverless_security_config

The `saml_options` argument is now a list nested block instead of a single nested block.
When referencing this argument, the index must now be included in the attribute address.
For example, `saml_options.session_timeout` would now be referenced as `saml_options[0].session_timeout`.

## resource/aws_paymentcryptography_key

The `key_attributes` and `key_attributes.key_modes_of_use` arguments are now list nested blocks instead of single nested blocks.
When referencing these arguments, the indices must now be included in the attribute address.
For example, `key_attributes.key_modes_of_use.decrypt` would now be referenced as `key_attributes[0].key_modes_of_use[0].decrypt`.

## resource/aws_redshift_cluster

* The `publicly_accessible` attribute now defaults to `false`.
* Remove `snapshot_copy` from your configuration—it no longer exists. Use the `aws_redshift_snapshot_copy` resource instead.
* Remove `logging` from your configuration—it no longer exists. Use the `aws_redshift_logging` resource instead.
* Attributes `cluster_public_key`, `cluster_revision_number`, and `endpoint` are now read only and should not be set.

## resource/aws_redshift_service_account

The `aws_redshift_service_account` resource has been removed. AWS [recommends](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) that a [service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) should be used instead of an AWS account ID in any relevant IAM policy.

## resource/aws_rekognition_stream_processor

The `regions_of_interest.bounding_box` argument is now a list nested block instead of a single nested block.
When referencing this argument, the index must now be included in the attribute address.
For example, `regions_of_interest[0].bounding_box.height` would now be referenced as `regions_of_interest[0].bounding_box[0].height`.

## resource/aws_resiliencehub_resiliency_policy

The following arguments are now list nested blocks instead of single nested blocks.

- `policy`
- `policy.az`
- `policy.hardware`
- `policy.software`
- `policy.region`

When referencing these arguments, the indices must now be included in the attribute address.
For example, `policy.az.rpo` would now be referenced as `policy[0].az[0].rpo`.

## resource/aws_sagemaker_notebook_instance

Remove `accelerator_types` from your configuration—it no longer exists. Instead, use `instance_type` to use [Inferentia](https://docs.aws.amazon.com/sagemaker/latest/dg/neo-supported-cloud.html).

## resource/aws_spot_instance_request

Remove `block_duration_minutes` from your configuration—it no longer exists.

## resource/aws_ssm_association

Remove `instance_id` from configuration—it no longer exists. Use `targets` instead.

## resource/aws_verifiedpermissions_schema

The `definition` argument is now a list nested block instead of a single nested block.
When referencing this argument, the index must now be included in the attribute address.
For example, `definition.value` would now be referenced as `definition[0].value`.
