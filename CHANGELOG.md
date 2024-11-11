## 5.75.1 (Unreleased)

ENHANCEMENTS:

* data-source/aws_cloudwatch_event_bus: Add `description` attribute ([#39980](https://github.com/hashicorp/terraform-provider-aws/issues/39980))
* resource/aws_api_gateway_account: Add attribute `reset_on_delete` to properly reset CloudWatch Role ARN on deletion. ([#40004](https://github.com/hashicorp/terraform-provider-aws/issues/40004))
* resource/aws_cloudwatch_event_bus: Add `description` argument ([#39980](https://github.com/hashicorp/terraform-provider-aws/issues/39980))

BUG FIXES:

* resource/aws_api_gateway_deployment: Rolls back validation of `canary_settings` and `stage_description` when `stage_name` not set. ([#40067](https://github.com/hashicorp/terraform-provider-aws/issues/40067))
* resource/aws_dynamodb_table: Allow table TTL to be disabled by allowing `ttl[0].attribute_name` to be set when `ttl[0].enabled` is false ([#40046](https://github.com/hashicorp/terraform-provider-aws/issues/40046))
* resource/aws_sagemaker_domain: Fix issue causing a `ValidationException` on updates when RStudio is disabled on the domain ([#40049](https://github.com/hashicorp/terraform-provider-aws/issues/40049))

## 5.75.0 (November  7, 2024)

BREAKING CHANGES:

* resource/aws_api_gateway_stage: Add `canary_settings.deployment_id` attribute as `required` ([#39929](https://github.com/hashicorp/terraform-provider-aws/issues/39929))

NOTES:

* provider: validation of arguments implementing the custom `ARNType` will properly surface validation errors ([#40008](https://github.com/hashicorp/terraform-provider-aws/issues/40008))
* resource/aws_api_gateway_stage: `deployment_id` was added to `canary_settings` as a `required` attribute. This breaking change was necessary to make `canary_settings` functional. Without this change all canary traffic was routed to the main deployment ([#39929](https://github.com/hashicorp/terraform-provider-aws/issues/39929))

FEATURES:

* **New Data Source:** `aws_spot_datafeed_subscription` ([#39647](https://github.com/hashicorp/terraform-provider-aws/issues/39647))

ENHANCEMENTS:

* data-source/aws_batch_job_definition: Add `init_containers`, `share_process_namespace`, and `image_pull_secrets` attributes ([#40019](https://github.com/hashicorp/terraform-provider-aws/issues/40019))
* resource/aws_batch_job_definition: Add `init_containers` and `share_process_namespace` arguments ([#40019](https://github.com/hashicorp/terraform-provider-aws/issues/40019))
* resource/aws_batch_job_definition: Increase maximum number of `containers` arguments to 10 ([#40019](https://github.com/hashicorp/terraform-provider-aws/issues/40019))
* resource/aws_eks_addon: Add `pod_identity_association` argument ([#38357](https://github.com/hashicorp/terraform-provider-aws/issues/38357))
* resource/aws_iam_user_login_profile: Mark the `password` argument as sensitive ([#39991](https://github.com/hashicorp/terraform-provider-aws/issues/39991))

BUG FIXES:

* resource/aws_api_gateway_deployment: Fix destroy error when canary stage still exists on resource ([#39929](https://github.com/hashicorp/terraform-provider-aws/issues/39929))
* resource/aws_codedeploy_deployment_group: Remove maximum items limit on the `alarm_configuration.alarms` argument ([#39971](https://github.com/hashicorp/terraform-provider-aws/issues/39971))
* resource/aws_eks_addon: Handle `ResourceNotFound` exceptions during resource destruction ([#38357](https://github.com/hashicorp/terraform-provider-aws/issues/38357))
* resource/aws_elasticache_reserved_cache_node: Fix `Value Conversion Error` during resource creation ([#39945](https://github.com/hashicorp/terraform-provider-aws/issues/39945))
* resource/aws_lb_listener: Fix errors when updating the `tcp_idle_timeout_seconds` argument for gateway load balancers ([#40039](https://github.com/hashicorp/terraform-provider-aws/issues/40039))
* resource/aws_lb_listener: Remove the default `tcp_idle_timeout_seconds` value, preventing `ModifyListenerAttributes` API calls when a value is not explicitly configured ([#40039](https://github.com/hashicorp/terraform-provider-aws/issues/40039))
* resource/aws_vpc_ipam_pool: Fix bug when `public_ip_source = "amazon"`: `The request can only contain PubliclyAdvertisable if the AddressFamily is IPv6 and PublicIpSource is byoip.` ([#40042](https://github.com/hashicorp/terraform-provider-aws/issues/40042))

## 5.74.0 (October 31, 2024)

FEATURES:

* **New Data Source:** `aws_lb_listener_rule` ([#39865](https://github.com/hashicorp/terraform-provider-aws/issues/39865))
* **New Resource:** `aws_opensearch_authorize_vpc_endpoint_access` ([#39846](https://github.com/hashicorp/terraform-provider-aws/issues/39846))
* **New Resource:** `aws_ssmquicksetup_configuration_manager` ([#39931](https://github.com/hashicorp/terraform-provider-aws/issues/39931))

ENHANCEMENTS:

* data-source/aws_imagebuilder_distribution_configuration: Add `distribution.s3_export_configuration` attribute ([#35492](https://github.com/hashicorp/terraform-provider-aws/issues/35492))
* data-source/aws_imagebuilder_image_recipe: Fix `block_device_mapping.0.ebs.0.delete_on_termination: '' expected type 'bool', got unconvertible type 'string'` errors ([#39928](https://github.com/hashicorp/terraform-provider-aws/issues/39928))
* resource/aws_codedeploy_deployment_group: Add `termination_hook_enabled` argument ([#35482](https://github.com/hashicorp/terraform-provider-aws/issues/35482))
* resource/aws_eks_cluster: Add `zonal_shift_config` argument ([#39852](https://github.com/hashicorp/terraform-provider-aws/issues/39852))
* resource/aws_imagebuilder_distribution_configuration: Add `distribution.s3_export_configuration` argument ([#35492](https://github.com/hashicorp/terraform-provider-aws/issues/35492))
* resource/aws_imagebuilder_image_pipeline: Allow `container_recipe_arn` and `image_recipe_arn` to be updated in-place ([#39117](https://github.com/hashicorp/terraform-provider-aws/issues/39117))
* resource/aws_keyspaces_keyspace: Add `replication_specification` argument ([#36331](https://github.com/hashicorp/terraform-provider-aws/issues/36331))
* resource/aws_launch_template: Add `efa-only` as a valid value for `network_interfaces.interface_type` ([#39882](https://github.com/hashicorp/terraform-provider-aws/issues/39882))
* resource/aws_transfer_server: Add `TransferSecurityPolicy-Restricted-2024-06` as a valid value for `security_policy_name` ([#39871](https://github.com/hashicorp/terraform-provider-aws/issues/39871))

BUG FIXES:

* resource/aws_docdb_cluster: Use `master_password` on resource Create when `snapshot_identifier` is configured ([#38193](https://github.com/hashicorp/terraform-provider-aws/issues/38193))
* resource/aws_imagebuilder_container_recipe: Change `component.parameter.name`, `component.parameter.value`, `target_repository.repository_name`, and `target_repository.service` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#39117](https://github.com/hashicorp/terraform-provider-aws/issues/39117))
* resource/aws_route53_record: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panic when `geolocation_routing_policy` is empty ([#39944](https://github.com/hashicorp/terraform-provider-aws/issues/39944))
* resource/aws_ssm_patch_baseline: Update `approval_rule.approve_after_days` validation to allow a maximum value of `360` ([#39949](https://github.com/hashicorp/terraform-provider-aws/issues/39949))
* resource/aws_wafv2_web_acl: Fix `decoding JSON: unexpected end of JSON input` errors when updating from using `rule_json` to using `rule` ([#39283](https://github.com/hashicorp/terraform-provider-aws/issues/39283))
* resource/aws_wafv2_web_acl: Fix unmarshal error for incompatible types in `rule_json` ([#39878](https://github.com/hashicorp/terraform-provider-aws/issues/39878))

## 5.73.0 (October 24, 2024)

FEATURES:

* **New Data Source:** `aws_ssm_patch_baselines` ([#39779](https://github.com/hashicorp/terraform-provider-aws/issues/39779))
* **New Resource:** `aws_imagebuilder_lifecycle_policy` ([#35674](https://github.com/hashicorp/terraform-provider-aws/issues/35674))
* **New Resource:** `aws_resiliencehub_resiliency_policy` ([#38913](https://github.com/hashicorp/terraform-provider-aws/issues/38913))
* **New Resource:** `aws_sagemaker_hub` ([#39807](https://github.com/hashicorp/terraform-provider-aws/issues/39807))
* **New Resource:** `aws_sagemaker_mlflow_tracking_server` ([#39796](https://github.com/hashicorp/terraform-provider-aws/issues/39796))

ENHANCEMENTS:

* data-source/aws_elasticache_reserved_cache_node_offering: Support `valkey` as valid value for `product_description` ([#39745](https://github.com/hashicorp/terraform-provider-aws/issues/39745))
* data-source/aws_lakeformation_data_lake_settings: Add `parameters` map attribute to read `CROSS_ACCOUNT_VERSION` ([#39826](https://github.com/hashicorp/terraform-provider-aws/issues/39826))
* data-source/aws_lb: Add `enable_zonal_shift` attribute ([#39585](https://github.com/hashicorp/terraform-provider-aws/issues/39585))
* resource/aws_apprunner_auto_scaling_configuration_version: Remove the upper limit on `min_size` and `max_size` ([#39843](https://github.com/hashicorp/terraform-provider-aws/issues/39843))
* resource/aws_batch_job_definition: Ensure that new revisions are created with tags ([#39797](https://github.com/hashicorp/terraform-provider-aws/issues/39797))
* resource/aws_codedeploy_deployment_config: Add `zonal_config` argument ([#34850](https://github.com/hashicorp/terraform-provider-aws/issues/34850))
* resource/aws_dynamodb_kinesis_streaming_destination: Add `approximate_creation_date_time_precision` argument ([#38098](https://github.com/hashicorp/terraform-provider-aws/issues/38098))
* resource/aws_elasticache_cluster: Support `valkey` as valid value for `engine` ([#39745](https://github.com/hashicorp/terraform-provider-aws/issues/39745))
* resource/aws_elasticache_global_replication_group: Support Valkey versions for `engine_version` ([#39745](https://github.com/hashicorp/terraform-provider-aws/issues/39745))
* resource/aws_elasticache_replication_group: Support Valkey versions for `engine_version` ([#39745](https://github.com/hashicorp/terraform-provider-aws/issues/39745))
* resource/aws_elasticache_replication_group: Support `valkey` as valid value for `engine` ([#39745](https://github.com/hashicorp/terraform-provider-aws/issues/39745))
* resource/aws_elasticache_serverless_cache: Support `valkey` as valid value for `engine` ([#39745](https://github.com/hashicorp/terraform-provider-aws/issues/39745))
* resource/aws_kinesis_firehose_delivery_stream: Add `iceberg_configuration` argument ([#39844](https://github.com/hashicorp/terraform-provider-aws/issues/39844))
* resource/aws_lakeformation_data_lake_settings: Add `parameters` map argument enabling `CROSS_ACCOUNT_VERSION` to be set ([#39826](https://github.com/hashicorp/terraform-provider-aws/issues/39826))
* resource/aws_lb: Add `enable_zonal_shift` argument ([#39585](https://github.com/hashicorp/terraform-provider-aws/issues/39585))
* resource/aws_lb_listener: Add `tcp_idle_timeout_seconds` argument ([#39585](https://github.com/hashicorp/terraform-provider-aws/issues/39585))
* resource/aws_route53profiles_association: Add regex and string length validation for `name` argument ([#39798](https://github.com/hashicorp/terraform-provider-aws/issues/39798))
* resource/aws_s3_bucket_object: Remove the call to `kms:DescribeKey` for the S3 default AWS managed key (`alias/aws/s3`) on Read ([#39782](https://github.com/hashicorp/terraform-provider-aws/issues/39782))
* resource/aws_s3_object: Remove the call to `kms:DescribeKey` for the S3 default AWS managed key (`alias/aws/s3`) on Read ([#39782](https://github.com/hashicorp/terraform-provider-aws/issues/39782))
* resource/aws_s3_object_copy: Remove the call to `kms:DescribeKey` for the S3 default AWS managed key (`alias/aws/s3`) on Read ([#39782](https://github.com/hashicorp/terraform-provider-aws/issues/39782))
* resource/aws_sagemaker_domain: Add `default_user_settings.jupyter_lab_app_settings.app_lifecycle_management`, `default_user_settings.jupyter_lab_app_settings.built_in_lifecycle_config_arn`, `default_user_settings.jupyter_lab_app_settings.emr_settings`, `default_space_settings.jupyter_lab_app_settings.app_lifecycle_management`, `default_space_settings.jupyter_lab_app_settings.built_in_lifecycle_config_arn`, `default_space_settings.jupyter_lab_app_settings.emr_settings`, `default_user_settings.auto_mount_home_efs`, `default_user_settings.canvas_app_settings.emr_serverless_settings`, `default_user_settings.studio_web_portal_settings.hidden_instance_types`, `default_user_settings.code_editor_app_settings.app_lifecycle_management`, `default_user_settings.code_editor_app_settings.built_in_lifecycle_config_arn`, and `tag_propagation` arguments ([#39774](https://github.com/hashicorp/terraform-provider-aws/issues/39774))
* resource/aws_sagemaker_domain: Allow `app_network_access_type` and `app_security_group_management` to be updated in-place ([#39774](https://github.com/hashicorp/terraform-provider-aws/issues/39774))
* resource/aws_sagemaker_feature_group: Add `feature_definition.collection_config`, `feature_definition.collection_type`, and `throughput_config` arguments ([#39805](https://github.com/hashicorp/terraform-provider-aws/issues/39805))
* resource/aws_sagemaker_space: Add `space_settings.code_editor_app_settings.app_lifecycle_management` and `space_settings.jupyter_lab_app_settings.app_lifecycle_management` arguments ([#39800](https://github.com/hashicorp/terraform-provider-aws/issues/39800))
* resource/aws_sagemaker_user_profile: Add `user_settings.auto_mount_home_efs`, `user_settings.canvas_app_settings.emr_serverless_settings`, `user_settings.code_editor_app_settings.app_lifecycle_management`, `user_settings.code_editor_app_settings.built_in_lifecycle_config_arn`, `user_settings.jupyter_lab_app_settings.app_lifecycle_management`, `user_settings.jupyter_lab_app_settings.built_in_lifecycle_config_arn`, `user_settings.jupyter_lab_app_settings.emr_settings` and `user_settings.studio_web_portal_settings.hidden_instance_types` arguments ([#39774](https://github.com/hashicorp/terraform-provider-aws/issues/39774))

BUG FIXES:

* data-source/aws_workspaces_bundle: Return the first matching bundle when searching by `name`. This fixes a regression introduced in [v5.72.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#5720-october-15-2024) causing `multiple WorkSpaces Bundles matched; use additional constraints to reduce matches to a single WorkSpaces Bundle` errors ([#39777](https://github.com/hashicorp/terraform-provider-aws/issues/39777))
* resource/aws_dynamodb_table: Fix validation error when optional attribute in `on_demand_throughput` is excluded ([#39784](https://github.com/hashicorp/terraform-provider-aws/issues/39784))
* resource/aws_ecr_repository_policy: Fix persistent validation errors when malformed `policy` content is written to state ([#39842](https://github.com/hashicorp/terraform-provider-aws/issues/39842))
* resource/aws_elasticache_serverless_cache: Fix `InvalidParameterValue: This API supports only cross-engine upgrades to Valkey engine currently` errors on Update ([#39745](https://github.com/hashicorp/terraform-provider-aws/issues/39745))
* resource/aws_iam_policy: Fix persistent validation errors when malformed `policy` content is written to state ([#39842](https://github.com/hashicorp/terraform-provider-aws/issues/39842))
* resource/aws_iam_role_policy: Fix persistent validation errors when malformed `policy` content is written to state ([#39842](https://github.com/hashicorp/terraform-provider-aws/issues/39842))
* resource/aws_kms_key: Fix persistent validation errors when malformed `policy` content is written to state ([#39842](https://github.com/hashicorp/terraform-provider-aws/issues/39842))
* resource/aws_quicksight_data_set: Fix `InvalidParameterValueException: Invalid RowLevelPermissionDataSet. Namespace parameter should not be specified for Version 2` errors on Create and Update ([#39778](https://github.com/hashicorp/terraform-provider-aws/issues/39778))
* resource/aws_route53_record: Allow creation of records with `ttl=0` ([#39728](https://github.com/hashicorp/terraform-provider-aws/issues/39728))
* resource/aws_s3_bucket_policy: Fix persistent validation errors when malformed `policy` content is written to state ([#39842](https://github.com/hashicorp/terraform-provider-aws/issues/39842))
* resource/aws_secretsmanager_secret: Fix persistent validation errors when malformed `policy` content is written to state ([#39842](https://github.com/hashicorp/terraform-provider-aws/issues/39842))
* resource/aws_security_group_rule: Remove from state when rule not found. This fixes a regression introduced in [v5.60.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#5600-july-25-2024) ([#39834](https://github.com/hashicorp/terraform-provider-aws/issues/39834))

## 5.72.1 (October 16, 2024)

FEATURES:

* **New Resource:** `aws_iam_group_policy_attachments_exclusive` ([#39732](https://github.com/hashicorp/terraform-provider-aws/issues/39732))
* **New Resource:** `aws_iam_user_policy_attachments_exclusive` ([#39731](https://github.com/hashicorp/terraform-provider-aws/issues/39731))

ENHANCEMENTS:

* resource/aws_resourceexplorer2_view:  Add `scope` argument ([#39744](https://github.com/hashicorp/terraform-provider-aws/issues/39744))

BUG FIXES:

* data-source/aws_batch_job_definition: Properly handles ignored tags. ([#39734](https://github.com/hashicorp/terraform-provider-aws/issues/39734))
* data-source/aws_cognito_user_pool: Properly handles ignored tags. ([#39734](https://github.com/hashicorp/terraform-provider-aws/issues/39734))
* resource/aws_cognito_user_pool: Properly handles ignored tags. ([#39734](https://github.com/hashicorp/terraform-provider-aws/issues/39734))
* resource/aws_dynamodb_table: Fix crash when `billing_mode` is set to `PAY_PER_REQUEST` without `global_secondary_index` updates ([#39752](https://github.com/hashicorp/terraform-provider-aws/issues/39752))
* resource/aws_dynamodb_table_replica: Properly handles default and ignored tags. ([#39734](https://github.com/hashicorp/terraform-provider-aws/issues/39734))
* resource/aws_resourceexplorer2_index: Correctly mark incomplete `AGGREGATOR` indexes as [tainted](https://developer.hashicorp.com/terraform/cli/state/taint#the-tainted-status) on Create ([#39744](https://github.com/hashicorp/terraform-provider-aws/issues/39744))

## 5.72.0 (October 15, 2024)

NOTES:

* This version contains all the features, enhancements, and bug fixes from the [v5.71.0 release](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#5710-october-11-2024) which was removed from the Terraform Registry ([#39692](https://github.com/hashicorp/terraform-provider-aws/issues/39692))
* resource/aws_iam_role: The `managed_policy_arns` argument is deprecated. Use the `aws_iam_role_policy_attachments_exclusive` resource instead. ([#39718](https://github.com/hashicorp/terraform-provider-aws/issues/39718))

FEATURES:

* **New Resource:** `aws_iam_role_policy_attachments_exclusive` ([#39718](https://github.com/hashicorp/terraform-provider-aws/issues/39718))

ENHANCEMENTS:

* data-source/aws_workspaces_directory: Add `saml_properties` attribute ([#39060](https://github.com/hashicorp/terraform-provider-aws/issues/39060))
* resource/aws_appflow_flow: Add `source_flow_config.source_connector_properties.sapo_data.pagination_config` and `source_flow_config.source_connector_properties.sapo_data.parallelism_config` attributes ([#38932](https://github.com/hashicorp/terraform-provider-aws/issues/38932))
* resource/aws_cloudwatch_event_rule: Add tags to AWS API request on Update to support [ABAC `aws:RequestTag` conditions](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_tags.html#access_tags_control-requests) ([#39648](https://github.com/hashicorp/terraform-provider-aws/issues/39648))
* resource/aws_cloudwatch_event_target: Add `appsync_target` configuration block ([#37773](https://github.com/hashicorp/terraform-provider-aws/issues/37773))
* resource/aws_dynamodb_table: Add `on_demand_throughput` and `global_secondary_index.on_demand_throughput` arguments ([#37799](https://github.com/hashicorp/terraform-provider-aws/issues/37799))
* resource/aws_rds_cluster: Increase maximum value of `serverlessv2_scaling_configuration.max_capacity` and `serverlessv2_scaling_configuration.min_capacity` from `128` to `256` ([#39697](https://github.com/hashicorp/terraform-provider-aws/issues/39697))
* resource/aws_rds_cluster_instance: Treat `storage-optimization` status as success when creating or updating cluster DB instances ([#39691](https://github.com/hashicorp/terraform-provider-aws/issues/39691))
* resource/aws_workspaces_directory: Add `saml_properties` configuration block ([#39060](https://github.com/hashicorp/terraform-provider-aws/issues/39060))

BUG FIXES:

* data-source/aws_ssm_document: Correct `arn` for automation documents ([#39705](https://github.com/hashicorp/terraform-provider-aws/issues/39705))
* resource/aws_cognito_user_pool: Fixes error when `schema` has empty `string_attribute_constraints` or `number_attribute_constraints` ([#20386](https://github.com/hashicorp/terraform-provider-aws/issues/20386))
* resource/aws_ssm_document: Correct `arn` for automation documents ([#39705](https://github.com/hashicorp/terraform-provider-aws/issues/39705))

## 5.71.0 (October 11, 2024)

This Terraform AWS Provider version has been removed from the [Terraform Registry](https://registry.terraform.io/providers/hashicorp/aws/latest) due to `archive has incorrect checksum` errors while installing the provider on some platforms.

The next planned Terraform AWS Provider release is **v5.72.0**, scheduled for the morning (EST) of October 17, 2024.

FEATURES:

* **New Data Source:** `aws_bedrock_inference_profile` ([#39342](https://github.com/hashicorp/terraform-provider-aws/issues/39342))
* **New Data Source:** `aws_bedrock_inference_profiles` ([#39342](https://github.com/hashicorp/terraform-provider-aws/issues/39342))
* **New Data Source:** `aws_elasticache_serverless_cache` ([#39590](https://github.com/hashicorp/terraform-provider-aws/issues/39590))
* **New Data Source:** `aws_prometheus_default_scraper_configuration` ([#35280](https://github.com/hashicorp/terraform-provider-aws/issues/35280))
* **New Data Source:** `aws_route53profiles_profiles` ([#38172](https://github.com/hashicorp/terraform-provider-aws/issues/38172))
* **New Resource:** `aws_backup_restore_testing_plan` ([#37039](https://github.com/hashicorp/terraform-provider-aws/issues/37039))
* **New Resource:** `aws_backup_restore_testing_selection` ([#37039](https://github.com/hashicorp/terraform-provider-aws/issues/37039))
* **New Resource:** `aws_datazone_user_profile` ([#38810](https://github.com/hashicorp/terraform-provider-aws/issues/38810))
* **New Resource:** `aws_pinpointsmsvoicev2_configuration_set` ([#39620](https://github.com/hashicorp/terraform-provider-aws/issues/39620))
* **New Resource:** `aws_route53profiles_association` ([#38172](https://github.com/hashicorp/terraform-provider-aws/issues/38172))
* **New Resource:** `aws_route53profiles_profile` ([#38172](https://github.com/hashicorp/terraform-provider-aws/issues/38172))
* **New Resource:** `aws_route53profiles_resource_association` ([#38172](https://github.com/hashicorp/terraform-provider-aws/issues/38172))

ENHANCEMENTS:

* data-source/aws_backup_plan: Add `rule.schedule_expression_timezone` attribute ([#33653](https://github.com/hashicorp/terraform-provider-aws/issues/33653))
* data-source/aws_eip: Add `ipam_pool_id` attribute ([#39604](https://github.com/hashicorp/terraform-provider-aws/issues/39604))
* data-source/aws_vpc_endpoint_service: Add `private_dns_names` attribute ([#39659](https://github.com/hashicorp/terraform-provider-aws/issues/39659))
* resource/aws_backup_plan: Add `rule.schedule_expression_timezone` argument ([#33653](https://github.com/hashicorp/terraform-provider-aws/issues/33653))
* resource/aws_batch_compute_environment: Add plan-time validation of `update_policy.job_execution_timeout_minutes` ([#39583](https://github.com/hashicorp/terraform-provider-aws/issues/39583))
* resource/aws_batch_job_definition: Suppress unnecessary differences in `container_properties.environment` ([#21834](https://github.com/hashicorp/terraform-provider-aws/issues/21834))
* resource/aws_eip: Add `ipam_pool_id` argument in support of [public IPAM pools](https://docs.aws.amazon.com/vpc/latest/ipam/tutorials-eip-pool.html) ([#39604](https://github.com/hashicorp/terraform-provider-aws/issues/39604))
* resource/aws_route53_resolver_endpoint: Add `resolver_endpoint_type` argument
resource/aws_route53_resolver_rule: Add `ipv6` optional argument to the `target_ip` object ([#30167](https://github.com/hashicorp/terraform-provider-aws/issues/30167))
* resource/aws_vpc_ipam: Add `enable_private_gua` argument ([#39600](https://github.com/hashicorp/terraform-provider-aws/issues/39600))
* resource/aws_vpc_ipv6_cidr_block_association: Add `ip_source` and `ipv6_address_attribute` attributes ([#39600](https://github.com/hashicorp/terraform-provider-aws/issues/39600))

BUG FIXES:

* resource/aws_backup_vault: Fix `empty result` errors reading vaults in certain Regions ([#39670](https://github.com/hashicorp/terraform-provider-aws/issues/39670))
* resource/aws_elasticache_replication_group: Fix `security_group_names` causing resource replacement after import ([#39591](https://github.com/hashicorp/terraform-provider-aws/issues/39591))
* resource/aws_instance: Fixed issues with `volume_tags`, `root_block_device.*.tags`, and `ebs_block_device.*.tags` where tags overlapped with default tags. These are now handled consistently with top-level tags throughout the provider. Specifically, tags defined in both locations are no longer removed, preventing erroneous differences. ([#37441](https://github.com/hashicorp/terraform-provider-aws/issues/37441))
* resource/aws_sagemaker_workteam: Mark `workforce_name` as Optional ([#39630](https://github.com/hashicorp/terraform-provider-aws/issues/39630))
* resource/aws_securityhub_automation_rule: Increase `criteria.aws_account_id`, `criteria.generator_id`, `criteria.resource_id`, and `criteria.title` max length from `20` to `100` ([#39616](https://github.com/hashicorp/terraform-provider-aws/issues/39616))
* resource/aws_vpc_ipam_pool: Change `publicly_advertisable` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#39600](https://github.com/hashicorp/terraform-provider-aws/issues/39600))
* resource/aws_vpc_ipam_pool: Fix `InvalidParameterCombination: The request can only contain PubliclyAdvertisable if the AddressFamily is IPv6 and PublicIpSource is byoip` errors ([#39600](https://github.com/hashicorp/terraform-provider-aws/issues/39600))


## 5.70.0 (October  4, 2024)

NOTES:

* resource/aws_s3_bucket_lifecycle_configuration: Amazon S3 now applies a default minimum object size of 128 KB for S3 Lifecycle transition rules to any S3 storage class. This new default behavior will be applied to any new or modified S3 Lifecycle configuration. You can override this new default and customize the minimum object size for S3 Lifecycle transition rules to any value ([#39578](https://github.com/hashicorp/terraform-provider-aws/issues/39578))
* resource/aws_simpledb_domain: The `aws_simpledb_domain` resource has been deprecated and will be removed in a future version. Use Amazon DynamoDB instead ([#39536](https://github.com/hashicorp/terraform-provider-aws/issues/39536))
* resource/aws_worklink_fleet: The `aws_worklink_fleet` resource has been deprecated and will be removed in a future version. Use Amazon WorkSpaces Secure Browser instead ([#39538](https://github.com/hashicorp/terraform-provider-aws/issues/39538))
* resource/aws_worklink_website_certificate_authority_association: The `aws_worklink_website_certificate_authority_association` resource has been deprecated and will be removed in a future version. Use Amazon WorkSpaces Secure Browser instead ([#39538](https://github.com/hashicorp/terraform-provider-aws/issues/39538))

FEATURES:

* **New Resource:** `aws_backup_logically_air_gapped_vault` ([#39098](https://github.com/hashicorp/terraform-provider-aws/issues/39098))
* **New Resource:** `aws_ec2_transit_gateway_default_route_table_association` ([#39496](https://github.com/hashicorp/terraform-provider-aws/issues/39496))
* **New Resource:** `aws_ec2_transit_gateway_default_route_table_propagation` ([#39517](https://github.com/hashicorp/terraform-provider-aws/issues/39517))
* **New Resource:** `aws_iam_group_policies_exclusive` ([#39554](https://github.com/hashicorp/terraform-provider-aws/issues/39554))
* **New Resource:** `aws_iam_user_policies_exclusive` ([#39544](https://github.com/hashicorp/terraform-provider-aws/issues/39544))
* **New Resource:** `aws_securityhub_standards_control_association` ([#39511](https://github.com/hashicorp/terraform-provider-aws/issues/39511))

ENHANCEMENTS:

* data-source/aws_ebs_snapshot: Add `start_time` attribute ([#39557](https://github.com/hashicorp/terraform-provider-aws/issues/39557))
* resource/aws_bedrockagent_agent_action_group: Add `prepare_agent` argument ([#39486](https://github.com/hashicorp/terraform-provider-aws/issues/39486))
* resource/aws_bedrockagent_data_source: Add `vector_ingestion_configuration.custom_transformation_configuration` argument ([#39556](https://github.com/hashicorp/terraform-provider-aws/issues/39556))
* resource/aws_globalaccelerator_endpoint_group: Add `endpoint_configuration.attachment_arn` argument ([#39507](https://github.com/hashicorp/terraform-provider-aws/issues/39507))
* resource/aws_lambda_code_signing_config: Add `tags` argument and `tags_all` attribute ([#39535](https://github.com/hashicorp/terraform-provider-aws/issues/39535))
* resource/aws_lambda_event_source_mapping: Add `arn` attribute ([#39535](https://github.com/hashicorp/terraform-provider-aws/issues/39535))
* resource/aws_lambda_event_source_mapping: Add `tags` argument and `tags_all` attribute ([#39535](https://github.com/hashicorp/terraform-provider-aws/issues/39535))
* resource/aws_s3_bucket_lifecycle_configuration: Add `transition_default_minimum_object_size` argument ([#39578](https://github.com/hashicorp/terraform-provider-aws/issues/39578))

BUG FIXES:

* resource/aws_bedrockagent_agent: Fix "Provider produced inconsistent result after apply" error on update due to `customer_encryption_key_arn` not being passed during update ([#39565](https://github.com/hashicorp/terraform-provider-aws/issues/39565))
* resource/aws_bedrockagent_agent: Fix "Provider produced inconsistent result after apply" error on update due to `prompt_override_configuration` not being passed when not modified ([#39565](https://github.com/hashicorp/terraform-provider-aws/issues/39565))
* resource/aws_bedrockagent_knowledge_base: Change `knowledge_base_configuration` and `storage_configuration` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#39567](https://github.com/hashicorp/terraform-provider-aws/issues/39567))
* resource/aws_ec2_transit_gateway_vpc_attachment: Remove default value for `security_group_referencing_support` argument and mark as Computed. This suppresses the diffs shown for resources created with v5.68.0 (or earlier) ([#39519](https://github.com/hashicorp/terraform-provider-aws/issues/39519))
* resource/aws_opensearchserverless_lifecycle_policy: Fix "Provider produced inconsistent result after apply" error on update due to `policy_version` computed attribute changing ([#39528](https://github.com/hashicorp/terraform-provider-aws/issues/39528))
* resource/aws_opensearchserverless_security_policy: Fix "Provider produced inconsistent result after apply" error on update due to `policy_version` computed attribute changing ([#39528](https://github.com/hashicorp/terraform-provider-aws/issues/39528))
* resource/aws_quicksight_dashboard: Fix mapping of `sheets.filter_controls.list.cascading_control_configuration` and `sheets.parameter_controls.list.cascading_control_configuration` attributes ([#39453](https://github.com/hashicorp/terraform-provider-aws/issues/39453))

## 5.69.0 (September 26, 2024)

NOTES:

* provider: This release contains an upstream AWS SDK for Go v2 [change](https://github.com/aws/aws-sdk-go-v2/issues/2807) to DynamoDB service endpoints. The Terraform AWS Provider will now connect to a DynamoDB endpoint in the format [`(account-id).ddb.(region).amazonaws.com`](https://docs.aws.amazon.com/sdkref/latest/guide/feature-account-endpoints.html) instead of `dynamodb.(region).amazonaws.com`. If your network configuration blocks outgoing traffic to DynamoDB based on DNS names or endpoint URLs, you must adjust your configuration, because the service's DNS name will change. You may instead disable account-based endpoints for DynamoDB by setting `account_id_endpoint_mode = disabled` in a [shared config file](https://docs.aws.amazon.com/sdkref/latest/guide/settings-reference.html#ConfigFileSettings) or setting the `AWS_ACCOUNT_ID_ENDPOINT_MODE` [environment variable](https://docs.aws.amazon.com/sdkref/latest/guide/settings-reference.html#EVarSettings) to `disabled` ([#39505](https://github.com/hashicorp/terraform-provider-aws/issues/39505))
* provider: Updates to Go `1.23.1`. The issue with AWS Network Firewall dropping TLS handshake `ClientHello` messages after the **v5.65.0** upgrade to Go `1.23.0`, temporarily resolved by the **v5.67.0** downgrade to Go `1.22.7`, has been addressed by removing the `X25519Kyber768Draft00` key exchange mechanism from the HTTP client used to make AWS API calls ([#39432](https://github.com/hashicorp/terraform-provider-aws/issues/39432))
* resource/aws_alb_listener: When importing a listener that has either a default action top-level target group ARN or a default action defining a forward action defining a target group with an ARN, include both in the configuration to avoid import differences ([#39413](https://github.com/hashicorp/terraform-provider-aws/issues/39413))
* resource/aws_lb_listener: When importing a listener that has either a default action top-level target group ARN or a default action defining a forward action defining a target group with an ARN, include both in the configuration to avoid import differences ([#39413](https://github.com/hashicorp/terraform-provider-aws/issues/39413))

ENHANCEMENTS:

* data-source/aws_connect_instance: Add `tags` attribute ([#39402](https://github.com/hashicorp/terraform-provider-aws/issues/39402))
* data-source/aws_ec2_transit_gateway: Add `security_group_referencing_support` attribute ([#34542](https://github.com/hashicorp/terraform-provider-aws/issues/34542))
* data-source/aws_ec2_transit_gateway_vpc_attachment: Add `security_group_referencing_support` attribute ([#34542](https://github.com/hashicorp/terraform-provider-aws/issues/34542))
* data-source/aws_opensearchserverless_collection: Add `failure_code` and `failure_reason` attributes ([#38995](https://github.com/hashicorp/terraform-provider-aws/issues/38995))
* resource/aws_bedrockagent_agent: Add `guardrail_configuration` argument ([#39440](https://github.com/hashicorp/terraform-provider-aws/issues/39440))
* resource/aws_connect_instance: Add `tags` argument and `tags_all` attribute ([#39402](https://github.com/hashicorp/terraform-provider-aws/issues/39402))
* resource/aws_ec2_transit_gateway: Add `security_group_referencing_support` argument ([#34542](https://github.com/hashicorp/terraform-provider-aws/issues/34542))
* resource/aws_ec2_transit_gateway_vpc_attachment: Add `security_group_referencing_support` argument ([#34542](https://github.com/hashicorp/terraform-provider-aws/issues/34542))
* resource/aws_ec2_transit_gateway_vpc_attachment_accepter: Add `security_group_referencing_support` argument ([#34542](https://github.com/hashicorp/terraform-provider-aws/issues/34542))
* resource/aws_ecs_service: Add `volume_configuration.managed_ebs_volume.tag_specifications` attribute ([#38662](https://github.com/hashicorp/terraform-provider-aws/issues/38662))
* resource/aws_identitystore_group: Allow `display_name` to be updated in-place ([#39416](https://github.com/hashicorp/terraform-provider-aws/issues/39416))
* resource/aws_kinesis_stream: Tag on Create to support attribute-based access control (ABAC) ([#39504](https://github.com/hashicorp/terraform-provider-aws/issues/39504))
* resource/aws_quicksight_data_source: Add `credentials.secret_arn` argument ([#29034](https://github.com/hashicorp/terraform-provider-aws/issues/29034))

BUG FIXES:

* data-source/aws_opensearchserverless_vpc_endpoint: Correctly set `security_group_ids`. This requires a call to the EC2 `DescribeVpcEndpoints` API ([#39454](https://github.com/hashicorp/terraform-provider-aws/issues/39454))
* data-source/aws_region: Fix lookups for the `ap-southeast-5` Region ([#39389](https://github.com/hashicorp/terraform-provider-aws/issues/39389))
* resource/aws_alb_listener: Fix several of the arguments to avoiding setting zero-values in situations where they shouldn't causing warnings and import differences ([#39413](https://github.com/hashicorp/terraform-provider-aws/issues/39413))
* resource/aws_alb_listener: Remove the limitation preventing setting both default_action.0.target_group_arn and default_action.0.forward to align with the AWS API which allows you to specify both a target group list and a top-level target group ARN if the ARNs match ([#39413](https://github.com/hashicorp/terraform-provider-aws/issues/39413))
* resource/aws_db_instance: Allow replica database to be added to domain on create ([#39448](https://github.com/hashicorp/terraform-provider-aws/issues/39448))
* resource/aws_db_instance_role_association: Fix intermittent failure when instance is not in an available state ([#39457](https://github.com/hashicorp/terraform-provider-aws/issues/39457))
* resource/aws_dynamodb_tag: Fix propagation timeout when multiple tags exist ([#39491](https://github.com/hashicorp/terraform-provider-aws/issues/39491))
* resource/aws_ecs_cluster: Fix validation error with `name` attribute. ([#38993](https://github.com/hashicorp/terraform-provider-aws/issues/38993))
* resource/aws_ecs_cluster_capacity_providers: Fix validation error with `name` attribute. ([#38993](https://github.com/hashicorp/terraform-provider-aws/issues/38993))
* resource/aws_iam_role: Retry `ConcurrentModificationException`s during role creation ([#39429](https://github.com/hashicorp/terraform-provider-aws/issues/39429))
* resource/aws_inspector2_enabler: Fix `AccessDeniedException: Lambda code scanning is not supported in ...` errors ([#38254](https://github.com/hashicorp/terraform-provider-aws/issues/38254))
* resource/aws_inspector2_member_association: Improve handling of `AccessDeniedException` errors during creation ([#38254](https://github.com/hashicorp/terraform-provider-aws/issues/38254))
* resource/aws_lb_listener: Fix several of the arguments to avoiding setting zero-values in situations where they shouldn't causing warnings and import differences ([#39413](https://github.com/hashicorp/terraform-provider-aws/issues/39413))
* resource/aws_lb_listener: Remove the limitation preventing setting both default_action.0.target_group_arn and default_action.0.forward to align with the AWS API which allows you to specify both a target group list and a top-level target group ARN if the ARNs match ([#39413](https://github.com/hashicorp/terraform-provider-aws/issues/39413))
* resource/aws_lb_listener_rule: Fix several of the arguments to avoiding setting zero-values in situations where they shouldn't causing warnings and import differences ([#39413](https://github.com/hashicorp/terraform-provider-aws/issues/39413))
* resource/aws_lb_target_group: Fix several of the arguments to avoiding setting zero-values in situations where they shouldn't causing warnings and import differences ([#39413](https://github.com/hashicorp/terraform-provider-aws/issues/39413))
* resource/aws_medialive_multiplex: Fix to properly handle read failures during delete operations which were previously ignored ([#39498](https://github.com/hashicorp/terraform-provider-aws/issues/39498))
* resource/aws_opensearchserverless_vpc_endpoint: Change `name` and `vpc_id` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#39454](https://github.com/hashicorp/terraform-provider-aws/issues/39454))
* resource/aws_opensearchserverless_vpc_endpoint: Correctly set `security_group_ids`. This requires a call to the EC2 `DescribeVpcEndpoints` API ([#39454](https://github.com/hashicorp/terraform-provider-aws/issues/39454))
* resource/aws_rds_cluster_role_association: Fix intermittent failure when cluster is not in an available state ([#39457](https://github.com/hashicorp/terraform-provider-aws/issues/39457))
* resource/aws_vpc_dhcp_options: Fix a bug causing a panic crash when an option is absent ([#39427](https://github.com/hashicorp/terraform-provider-aws/issues/39427))

## 5.68.0 (September 19, 2024)

NOTES:

* resource/aws_iam_role: The `inline_policy` argument is deprecated. Use the `aws_iam_role_policy` resource instead. If Terraform should exclusively manage all inline policy associations (the current behavior of this argument), use the `aws_iam_role_policies_exclusive` resource as well. ([#39203](https://github.com/hashicorp/terraform-provider-aws/issues/39203))
* resource/aws_lexv2models_slot_type: Within the `composite_slot_type_setting` block, the `subslots` argument has been renamed `sub_slots`. See the [linked pull request](https://github.com/hashicorp/terraform-provider-aws/pull/39353) for additional justification on this change. The previous misnaming effectively made this argument unusable, therefore a breaking change in a minor version was deemed acceptable. ([#39353](https://github.com/hashicorp/terraform-provider-aws/issues/39353))

FEATURES:

* **New Data Source:** `aws_elasticache_reserved_cache_node_offering` ([#29832](https://github.com/hashicorp/terraform-provider-aws/issues/29832))
* **New Data Source:** `aws_securityhub_standards_control_associations` ([#39334](https://github.com/hashicorp/terraform-provider-aws/issues/39334))
* **New Data Source:** `aws_synthetics_runtime_version` ([#39180](https://github.com/hashicorp/terraform-provider-aws/issues/39180))
* **New Data Source:** `aws_synthetics_runtime_versions` ([#39180](https://github.com/hashicorp/terraform-provider-aws/issues/39180))
* **New Resource:** `aws_appsync_source_api_association` ([#39323](https://github.com/hashicorp/terraform-provider-aws/issues/39323))
* **New Resource:** `aws_elasticache_reserved_cache_node` ([#29832](https://github.com/hashicorp/terraform-provider-aws/issues/29832))
* **New Resource:** `aws_iam_role_policies_exclusive` ([#39203](https://github.com/hashicorp/terraform-provider-aws/issues/39203))
* **New Resource:** `aws_pinpointsmsvoicev2_opt_out_list` ([#25036](https://github.com/hashicorp/terraform-provider-aws/issues/25036))
* **New Resource:** `aws_pinpointsmsvoicev2_phone_number` ([#25036](https://github.com/hashicorp/terraform-provider-aws/issues/25036))
* **New Resource:** `aws_sesv2_account_suppression_attributes` ([#39325](https://github.com/hashicorp/terraform-provider-aws/issues/39325))

ENHANCEMENTS:

* resource/aws_s3_bucket_server_side_encryption_configuration: S3 directory buckets now support SSE-KMS ([#39366](https://github.com/hashicorp/terraform-provider-aws/issues/39366))
* resource/aws_ses_receipt_rule: Add `iam_role_arn` argument to `s3_action` configuration block ([#39364](https://github.com/hashicorp/terraform-provider-aws/issues/39364))
* resource/aws_synthetics_canary: Increase maximum `name` length to 255 characters ([#39315](https://github.com/hashicorp/terraform-provider-aws/issues/39315))

BUG FIXES:

* provider: Allows `assume_role.role_arn` to be an empty string when there is a single `assume_role` entry. ([#39328](https://github.com/hashicorp/terraform-provider-aws/issues/39328))
* resource/aws_amplify_app: Fix failure when unsetting the `environment_variables` argument ([#39397](https://github.com/hashicorp/terraform-provider-aws/issues/39397))
* resource/aws_dynamodb_table: Fix changing replicas to the default `Managed by DynamoDB` encryption setting ([#31284](https://github.com/hashicorp/terraform-provider-aws/issues/31284))
* resource/aws_dynamodb_table: Handle eventual consistency of tag creation and removal ([#39326](https://github.com/hashicorp/terraform-provider-aws/issues/39326))
* resource/aws_dynamodb_table_replica: Handle eventual consistency of tag creation and removal ([#39326](https://github.com/hashicorp/terraform-provider-aws/issues/39326))
* resource/aws_dynamodb_tag: Handle eventual consistency of tag creation and removal ([#39326](https://github.com/hashicorp/terraform-provider-aws/issues/39326))
* resource/aws_mq_broker: Fix `engine_version` mismatch with RabbitMQ 3.13 and ActiveMQ 5.18 and above ([#39024](https://github.com/hashicorp/terraform-provider-aws/issues/39024))
* resource/aws_mwaa_environment: Fix creating environments with `endpoint_management = "CUSTOMER"` ([#39394](https://github.com/hashicorp/terraform-provider-aws/issues/39394))
* resource/aws_opensearchserverless_access_policy: Fix incompatible type error when setting `policy` ([#39322](https://github.com/hashicorp/terraform-provider-aws/issues/39322))

## 5.67.0 (September 12, 2024)

BREAKING CHANGES:

* resource/aws_lexv2models_slot_type: Within the `value_selection_setting.advanced_recognition_setting` block, the `audio_recognition_setting` argument has been renamed `audio_recognition_strategy` ([#39254](https://github.com/hashicorp/terraform-provider-aws/issues/39254))

NOTES:

* provider: Downgrades to Go `1.22.6`. A small number of users have reported failed or hanging network connections using the version of the Terraform AWS provider which was first built with Go `1.23.0` (`v5.65.0`). At this point, maintainers have been unable to reproduce failures, but enough distinct users have reported issues that we are going to attempt downgrading to Go `1.22.6` for the next provider release. We will continue to coordinate with users and AWS in an attempt to identify the root cause, using this upcoming release with a reverted Go build version as a data point. ([#39256](https://github.com/hashicorp/terraform-provider-aws/issues/39256))
* resource/aws_lexv2models_slot_type: Within the `value_selection_setting.advanced_recognition_setting` block, the `audio_recognition_setting` argument has been renamed `audio_recognition_strategy`. See the [linked pull request](https://github.com/hashicorp/terraform-provider-aws/pull/39254) for additional justification on this change. The previous misnaming effectively made this argument unusable, therefore a breaking change in a minor version was deemed acceptable. ([#39254](https://github.com/hashicorp/terraform-provider-aws/issues/39254))

FEATURES:

* **New Data Source:** `aws_codebuild_fleet` ([#39237](https://github.com/hashicorp/terraform-provider-aws/issues/39237))
* **New Resource:** `aws_cloudformation_stack_instances` ([#36794](https://github.com/hashicorp/terraform-provider-aws/issues/36794))
* **New Resource:** `aws_codebuild_fleet` ([#39237](https://github.com/hashicorp/terraform-provider-aws/issues/39237))
* **New Resource:** `aws_computeoptimizer_enrollment_status` ([#35349](https://github.com/hashicorp/terraform-provider-aws/issues/35349))
* **New Resource:** `aws_computeoptimizer_recommendation_preferences` ([#35349](https://github.com/hashicorp/terraform-provider-aws/issues/35349))
* **New Resource:** `aws_costoptimizationhub_enrollment_status` ([#36440](https://github.com/hashicorp/terraform-provider-aws/issues/36440))
* **New Resource:** `aws_costoptimizationhub_preferences` ([#36526](https://github.com/hashicorp/terraform-provider-aws/issues/36526))
* **New Resource:** `aws_datazone_asset_type` ([#38812](https://github.com/hashicorp/terraform-provider-aws/issues/38812))
* **New Resource:** `aws_datazone_environment_profile` ([#38581](https://github.com/hashicorp/terraform-provider-aws/issues/38581))
* **New Resource:** `aws_lambda_function_recursion_config` ([#39153](https://github.com/hashicorp/terraform-provider-aws/issues/39153))

ENHANCEMENTS:

* data-source/aws_acm_certificate: Mark `domain` and `tags` as Optional. This enables certificates to be matched based on tags ([#31453](https://github.com/hashicorp/terraform-provider-aws/issues/31453))
* data-source/aws_kinesis_stream: Add `encryption_type` and `kms_key_id` attributes ([#39212](https://github.com/hashicorp/terraform-provider-aws/issues/39212))
* datasource/aws_cognito_user_pool: Deprecates `user_pool_tags` in favor of standard `tags`. ([#39260](https://github.com/hashicorp/terraform-provider-aws/issues/39260))
* provider: Adds support for IAM role chaining. The provider attribute `assume_role` now accepts multiple elements. ([#39255](https://github.com/hashicorp/terraform-provider-aws/issues/39255))
* resource/aws_amplify_app: Add `cache_config` argument ([#39215](https://github.com/hashicorp/terraform-provider-aws/issues/39215))
* resource/aws_cloudhsm_v2_cluster: Add `mode` argument ([#39206](https://github.com/hashicorp/terraform-provider-aws/issues/39206))
* resource/aws_cloudhsm_v2_cluster: Support `hsm2m.medium` as a valid value for `hsm_type` ([#39206](https://github.com/hashicorp/terraform-provider-aws/issues/39206))
* resource/aws_codebuild_project: Add `fleet` attribute in `environment` configuration block ([#39237](https://github.com/hashicorp/terraform-provider-aws/issues/39237))
* resource/aws_kinesis_firehose_delivery_stream: Add `snowflake_configuration.buffering_internal` and `snowflake_configuration.buffering_size` arguments ([#39214](https://github.com/hashicorp/terraform-provider-aws/issues/39214))
* resource/aws_quicksight_user: Add `READER_PRO`, `AUTHOR_PRO`, and `ADMIN_PRO` as valid values for the `user_role` argument ([#39220](https://github.com/hashicorp/terraform-provider-aws/issues/39220))
* resource/aws_sagemaker_domain: Add `default_user_settings.domain_settings.docker_settings` configuration block ([#35416](https://github.com/hashicorp/terraform-provider-aws/issues/35416))
* resource/aws_sagemaker_domain: Add `default_user_settings.studio_web_portal_settings`, `default_space_settings.jupyter_lab_app_settings`, `default_space_settings.space_storage_settings`, `default_space_settings.custom_posix_user_config`, and `default_space_settings.custom_file_system_config` configuration blocks ([#38457](https://github.com/hashicorp/terraform-provider-aws/issues/38457))
* resource/aws_sagemaker_endpoint_configuration: Add `production_variants.managed_instance_scaling` and `shadow_production_variants.managed_instance_scaling` configuration blocks ([#35479](https://github.com/hashicorp/terraform-provider-aws/issues/35479))
* resource/aws_sagemaker_model: Add `primary_container.inference_specification_name` and `container.inference_specification_name` arguments ([#35873](https://github.com/hashicorp/terraform-provider-aws/issues/35873))
* resource/aws_sagemaker_model: Add `primary_container.model_data_source.s3_data_source.model_access_config`, `primary_container.multi_model_config`, `container.model_data_source.s3_data_source.model_access_config`, and `container.multi_model_config` configuration blocks ([#35873](https://github.com/hashicorp/terraform-provider-aws/issues/35873))
* resource/aws_sagemaker_user_profile: Add `user_settings.studio_web_portal_settings` configuration block ([#38567](https://github.com/hashicorp/terraform-provider-aws/issues/38567))
* resource/aws_sfn_state_machine: Add plan-time validation of `definition` using the AWS Step Functions [Validation API](https://docs.aws.amazon.com/step-functions/latest/apireference/API_ValidateStateMachineDefinition.html) ([#39229](https://github.com/hashicorp/terraform-provider-aws/issues/39229))

BUG FIXES:

* data-source/aws_eks_cluster: Return `created_at` as an [RFC3339](https://www.rfc-editor.org/rfc/rfc3339) formatted timestamp ([#24183](https://github.com/hashicorp/terraform-provider-aws/issues/24183))
* datasource/aws_cognito_user_pool: Fixes value conversion error. ([#39260](https://github.com/hashicorp/terraform-provider-aws/issues/39260))
* provider: Fix empty tags drift on fwprovider resources ([#38636](https://github.com/hashicorp/terraform-provider-aws/issues/38636))
* resource/aws_batch_job_queue: Fixes error in schema migration function. ([#39257](https://github.com/hashicorp/terraform-provider-aws/issues/39257))
* resource/aws_cognito_user_pool: Correctly unsets tags. ([#39260](https://github.com/hashicorp/terraform-provider-aws/issues/39260))
* resource/aws_ecr_repository_policy: Fix retry logic handling eventual consistency of newly created IAM roles ([#39190](https://github.com/hashicorp/terraform-provider-aws/issues/39190))
* resource/aws_eks_cluster: Return `created_at` as an [RFC3339](https://www.rfc-editor.org/rfc/rfc3339) formatted timestamp ([#24183](https://github.com/hashicorp/terraform-provider-aws/issues/24183))
* resource/aws_iam_role: Fix to reduce Terraform reporting differences when a role's ARN temporarily appears as the role's unique ID ([#36794](https://github.com/hashicorp/terraform-provider-aws/issues/36794))
* resource/aws_networkfirewall_tls_inspection_configuration: Fix issue where `check_certificate_revovation_status` is ignored due to bad autoflex field mapping ([#39211](https://github.com/hashicorp/terraform-provider-aws/issues/39211))
* resource/aws_networkmonitor_monitor: Fixes error when optional attribute `aggregation_period` not set. ([#39279](https://github.com/hashicorp/terraform-provider-aws/issues/39279))
* resource/aws_quicksight_data_set: Change `permissions.actions` `MaxItems` from `16` to `20`. This fixes a regression introduced in [v5.66.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#5660-september--5-2024) ([#39226](https://github.com/hashicorp/terraform-provider-aws/issues/39226))
* resource/aws_quicksight_vpc_connection: Remove `vpc_connection_id` regular expression validator. This fixes a regression introduced in [v5.66.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#5660-september--5-2024) ([#39231](https://github.com/hashicorp/terraform-provider-aws/issues/39231))
* resource/aws_sagemaker_domain: Fix update for `default_user_settings.domain_settings` to include missing `security_group_ids` and `r_studio_server_pro_domain_settings` values ([#35416](https://github.com/hashicorp/terraform-provider-aws/issues/35416))
* resource/aws_sesv2_configuration_set: Allow `suppression_options.suppressed_reasons` to be an empty list (`[]`) in order to disable the suppression list ([#29671](https://github.com/hashicorp/terraform-provider-aws/issues/29671))
* resource/aws_sesv2_configuration_set_event_destination: Change `event_destination.matching_event_types` from `TypeList` to `TypeSet` as order is not significant ([#36897](https://github.com/hashicorp/terraform-provider-aws/issues/36897))
* resource/aws_verifiedaccess_endpoint: fix crash when updating `load_balancer_options.subnet_ids` ([#39196](https://github.com/hashicorp/terraform-provider-aws/issues/39196))

## 5.66.0 (September  5, 2024)

FEATURES:

* **New Data Source:** `aws_glue_registry` ([#37953](https://github.com/hashicorp/terraform-provider-aws/issues/37953))
* **New Data Source:** `aws_organizations_organizational_unit_descendant_organizational_units` ([#39120](https://github.com/hashicorp/terraform-provider-aws/issues/39120))
* **New Data Source:** `aws_quicksight_analysis` ([#31737](https://github.com/hashicorp/terraform-provider-aws/issues/31737))
* **New Resource:** `aws_datazone_environment` ([#38811](https://github.com/hashicorp/terraform-provider-aws/issues/38811))

ENHANCEMENTS:

* data-source/aws_sns_topic: Add `tags` attribute ([#38959](https://github.com/hashicorp/terraform-provider-aws/issues/38959))
* data-source/aws_transfer_server: Add `tags` attribute ([#39092](https://github.com/hashicorp/terraform-provider-aws/issues/39092))
* resource/aws_appsync_graphql_api: Add `api_type` and `merged_api_execution_role_arn` arguments ([#39159](https://github.com/hashicorp/terraform-provider-aws/issues/39159))
* resource/aws_bedrockagent_data_source: Add `vector_ingestion_configuration.chunking_configuration.semantic_chunking_configuration`, `vector_ingestion_configuration.chunking_configuration.hierarchical_chunking_configuration`, and `vector_ingestion_configuration.parsing_configuration` configuration blocks ([#39138](https://github.com/hashicorp/terraform-provider-aws/issues/39138))
* resource/aws_datazone_domain: Add `skip_deletion_protection` attribute ([#38811](https://github.com/hashicorp/terraform-provider-aws/issues/38811))
* resource/aws_docdbelastic_cluster: Add `backup_retention_period` and `preferred_backup_window` attributes ([#38452](https://github.com/hashicorp/terraform-provider-aws/issues/38452))
* resource/aws_quicksight_data_source: Add `parameters.databricks` argument ([#31737](https://github.com/hashicorp/terraform-provider-aws/issues/31737))
* resource/aws_rolesanywhere_trust_anchor: Add `notification_settings` argument ([#39108](https://github.com/hashicorp/terraform-provider-aws/issues/39108))
* resource/aws_sagemaker_endpoint: Increase Create and Update `InService` timeouts to 60 minutes ([#39090](https://github.com/hashicorp/terraform-provider-aws/issues/39090))
* resource/aws_wafv2_rule_group: Reduce `rate_based_statement.limit` minimum from `100` to `10` ([#39107](https://github.com/hashicorp/terraform-provider-aws/issues/39107))
* resource/aws_wafv2_web_acl: Reduce `rate_based_statement.limit` minimum from `100` to `10` ([#39107](https://github.com/hashicorp/terraform-provider-aws/issues/39107))

BUG FIXES:

* data-source/aws_networkmanager_core_network_policy_document: Change `segment_actions.via.with_edge_override.use_edge` to be nested set of edges, matching JSON ([#39142](https://github.com/hashicorp/terraform-provider-aws/issues/39142))
* data-source/aws_networkmanager_core_network_policy_document: Deprecate `segment_actions.via.with_edge_override.use_edge`. Use `segment_actions.via.with_edge_override.use_edge_location` instead ([#39142](https://github.com/hashicorp/terraform-provider-aws/issues/39142))
* many resources: Fixes perpetual diff when tag has a `null` value. ([#38869](https://github.com/hashicorp/terraform-provider-aws/issues/38869))
* resource/aws_appconfig_extension: Mark `role_arn` as Optional ([#38900](https://github.com/hashicorp/terraform-provider-aws/issues/38900))
* resource/aws_lexv2models_slot_type: Fix `slot_type_values` validator which limited configurations to 1 element ([#39126](https://github.com/hashicorp/terraform-provider-aws/issues/39126))
* resource/aws_quicksight_analysis: Properly send `theme_arn` argument on create and update when configured ([#31737](https://github.com/hashicorp/terraform-provider-aws/issues/31737))
* resource/aws_rolesanywhere_profile: Mark `role_arns` as Optional and send an empty list if unconfigured ([#39108](https://github.com/hashicorp/terraform-provider-aws/issues/39108))
* resource/aws_synthetics_canary: Remove `run_config.timeout_in_seconds` default value to allow creation of resources with a frequency less than 14 minutes ([#35177](https://github.com/hashicorp/terraform-provider-aws/issues/35177))

## 5.65.0 (August 29, 2024)

NOTES:

* provider: Updates to Go 1.23. We do not expect this change to impact most users. For macOS, Go 1.23 requires macOS 11 Big Sur or later; support for previous versions has been discontinued. ([#38999](https://github.com/hashicorp/terraform-provider-aws/issues/38999))

FEATURES:

* **New Data Source:** `aws_shield_protection` ([#37524](https://github.com/hashicorp/terraform-provider-aws/issues/37524))
* **New Resource:** `aws_glue_catalog_table_optimizer` ([#38052](https://github.com/hashicorp/terraform-provider-aws/issues/38052))

ENHANCEMENTS:

* data-source/aws_elb_hosted_zone_id: Add hosted zone ID for `ap-southeast-5` AWS Region ([#39052](https://github.com/hashicorp/terraform-provider-aws/issues/39052))
* data-source/aws_lb_hosted_zone_id: Add hosted zone IDs for `ap-southeast-5` AWS Region ([#39052](https://github.com/hashicorp/terraform-provider-aws/issues/39052))
* data-source/aws_s3_bucket: Add hosted zone ID for `ap-southeast-5` AWS Region ([#39052](https://github.com/hashicorp/terraform-provider-aws/issues/39052))
* provider: Support `ap-southeast-5` as a valid AWS Region ([#39049](https://github.com/hashicorp/terraform-provider-aws/issues/39049))
* resource/aws_cognito_user_pool: Add `password_policy.password_history_size` argument ([#39043](https://github.com/hashicorp/terraform-provider-aws/issues/39043))
* resource/aws_elastic_beanstalk_application_version: Add `process` argument ([#25468](https://github.com/hashicorp/terraform-provider-aws/issues/25468))
* resource/aws_elasticsearch_domain: Treat `SUCCEEDED_WITH_ISSUES` status as success when upgrading cluster ([#38086](https://github.com/hashicorp/terraform-provider-aws/issues/38086))
* resource/aws_emr_cluster: Support `io2` as a valid value for `ebs_config.type` ([#37740](https://github.com/hashicorp/terraform-provider-aws/issues/37740))
* resource/aws_emr_instance_fleet: Support `io2` as a valid value for `instance_type_configs.ebs_config.type` ([#37740](https://github.com/hashicorp/terraform-provider-aws/issues/37740))
* resource/aws_emr_instance_group: Support `io2` as a valid value for `instance_type_configs.ebs_config.type` ([#37740](https://github.com/hashicorp/terraform-provider-aws/issues/37740))
* resource/aws_glue_job: Add `job_run_queuing_enabled` argument ([#39027](https://github.com/hashicorp/terraform-provider-aws/issues/39027))
* resource/aws_lambda_event_source_mapping: Add `kms_key_arn` argument ([#39055](https://github.com/hashicorp/terraform-provider-aws/issues/39055))
* resource/aws_verifiedaccess_endpoint: Set PolicyEnabled flag to `false` on update if `policy_document` is empty ([#38675](https://github.com/hashicorp/terraform-provider-aws/issues/38675))

BUG FIXES:

* resource/aws_amplify_app: Fix crash updating `auto_branch_creation_config` ([#39041](https://github.com/hashicorp/terraform-provider-aws/issues/39041))
* resource/aws_elasticsearch_domain_policy: Change `domain_name` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#38086](https://github.com/hashicorp/terraform-provider-aws/issues/38086))
* resource/aws_elbv2_listener: Fix crash when reading forward actions not configured in state ([#39039](https://github.com/hashicorp/terraform-provider-aws/issues/39039))
* resource/aws_emr_instance_group: Properly send an `instance_count` value of `0` on create when configured ([#37740](https://github.com/hashicorp/terraform-provider-aws/issues/37740))
* resource/aws_gamelift_game_server_group: Fix crash while reading server group with a nil auto scaling group ARN ([#39022](https://github.com/hashicorp/terraform-provider-aws/issues/39022))
* resource/aws_guardduty_invite_accepter: Fix `BadRequestException: The request is rejected because an invalid or out-of-range value is specified as an input parameter` errors on resource Create ([#39084](https://github.com/hashicorp/terraform-provider-aws/issues/39084))
* resource/aws_lakeformation_permissions: Fix error when revoking `data_cells_filter` permissions ([#39026](https://github.com/hashicorp/terraform-provider-aws/issues/39026))
* resource/aws_neptune_cluster: Mark `neptune_cluster_parameter_group_name` as Computed ([#38980](https://github.com/hashicorp/terraform-provider-aws/issues/38980))
* resource/aws_neptune_cluster_instance: Mark `neptune_parameter_group_name` as Computed ([#38980](https://github.com/hashicorp/terraform-provider-aws/issues/38980))
* resource/aws_ssm_parameter: Fix `ValidationException: Parameter ARN is not supported for this operation` errors when deleting resources imported by ARN ([#39067](https://github.com/hashicorp/terraform-provider-aws/issues/39067))

## 5.64.0 (August 22, 2024)

ENHANCEMENTS:

* data-source/aws_opensearch_domain: Add `dashboard_endpoint_v2`, `domain_endpoint_v2_hosted_zone_id`, and `endpoint_v2` attributes ([#38456](https://github.com/hashicorp/terraform-provider-aws/issues/38456))
* resource/aws_appautoscaling_target: Add `suspended_state` configuration block ([#38942](https://github.com/hashicorp/terraform-provider-aws/issues/38942))
* resource/aws_dynamodb_table: Add `restore_source_table_arn` attribute ([#38953](https://github.com/hashicorp/terraform-provider-aws/issues/38953))
* resource/aws_opensearch_domain: Add `dashboard_endpoint_v2`, `domain_endpoint_v2_hosted_zone_id`, and `endpoint_v2` attributes ([#38456](https://github.com/hashicorp/terraform-provider-aws/issues/38456))

BUG FIXES:

* resource/aws_bedrockagent_agent: Fixes consistency issues where only some prompts are overridden ([#38944](https://github.com/hashicorp/terraform-provider-aws/issues/38944))
* resource/aws_cloudformation_stack_set_instance: Fix crash during construction of the `id` attribute when `deployment_targets` does not include organizational unit IDs. ([#38969](https://github.com/hashicorp/terraform-provider-aws/issues/38969))
* resource/aws_glue_trigger: Fix crash when null `action` is configured ([#38994](https://github.com/hashicorp/terraform-provider-aws/issues/38994))
* resource/aws_rds_cluster: Allow Web Service Data API (`enabled_http_endpoint`) to be enabled and disabled for `provisioned` engine mode and serverlessv2 ([#38997](https://github.com/hashicorp/terraform-provider-aws/issues/38997))

## 5.63.1 (August 20, 2024)

FEATURES:

* **New Data Source:** `aws_route53_zones` ([#17457](https://github.com/hashicorp/terraform-provider-aws/issues/17457))
* **New Data Source:** `aws_ssoadmin_permission_sets` ([#38741](https://github.com/hashicorp/terraform-provider-aws/issues/38741))

ENHANCEMENTS:

* data-source/aws_batch_job_queue: Add `job_state_time_limit_action` attribute ([#38784](https://github.com/hashicorp/terraform-provider-aws/issues/38784))
* resource/aws_batch_job_definition: Add `ecs_properties` argument ([#37871](https://github.com/hashicorp/terraform-provider-aws/issues/37871))
* resource/aws_batch_job_queue: Add `job_state_time_limit_action` argument ([#38784](https://github.com/hashicorp/terraform-provider-aws/issues/38784))

BUG FIXES:

* provider: Fix crash when flattening string pointer slices with nil items ([#38886](https://github.com/hashicorp/terraform-provider-aws/issues/38886))
* resource/aws_datazone_project: Properly surface import `id` parsing errors ([#38924](https://github.com/hashicorp/terraform-provider-aws/issues/38924))
* resource/aws_quicksight_data_set: Fix crash when setting `logical_table_map.data_transforms.project_operation.projected_columns` with null list elements ([#38886](https://github.com/hashicorp/terraform-provider-aws/issues/38886))
* resource/aws_ses_configuration_set: Fix crash when `reputation_metrics_enabled` is set to `true` ([#38921](https://github.com/hashicorp/terraform-provider-aws/issues/38921))

## 5.63.0 (August 15, 2024)

FEATURES:

* **New Data Source:** `aws_bedrockagent_agent_versions` ([#38792](https://github.com/hashicorp/terraform-provider-aws/issues/38792))
* **New Resource:** `aws_bedrock_guardrail` ([#38757](https://github.com/hashicorp/terraform-provider-aws/issues/38757))
* **New Resource:** `aws_cloudtrail_organization_delegated_admin_account` ([#38817](https://github.com/hashicorp/terraform-provider-aws/issues/38817))
* **New Resource:** `aws_datazone_environment_profile` ([#38581](https://github.com/hashicorp/terraform-provider-aws/issues/38581))
* **New Resource:** `aws_datazone_form_type` ([#38746](https://github.com/hashicorp/terraform-provider-aws/issues/38746))
* **New Resource:** `aws_datazone_glossary_term` ([#38706](https://github.com/hashicorp/terraform-provider-aws/issues/38706))
* **New Resource:** `aws_pinpoint_email_template` ([#33266](https://github.com/hashicorp/terraform-provider-aws/issues/33266))

ENHANCEMENTS:

* resource/aws_networkfirewall_logging_configuration: Change `logging_configuration.log_destination_config` `MaxItems` from `2` to `3` ([#38824](https://github.com/hashicorp/terraform-provider-aws/issues/38824))

BUG FIXES:

* data-source/aws_acm_certificate: Fix unreturned `sdkdiags.AppendErrorf` function calls ([#38854](https://github.com/hashicorp/terraform-provider-aws/issues/38854))
* resource/aws_appstream_stack: Fix unreturned `sdkdiags.AppendErrorf` function calls ([#38854](https://github.com/hashicorp/terraform-provider-aws/issues/38854))
* resource/aws_bedrockagent_agent_knowledge_base_association: Prepare agent when associating a knowledge base so it can be used ([#38799](https://github.com/hashicorp/terraform-provider-aws/issues/38799))
* resource/aws_cloudwatch_event_connection: Fix various expander type assertions to prevent crashes ([#38800](https://github.com/hashicorp/terraform-provider-aws/issues/38800))
* resource/aws_controltower_landing_zone: Fix unreturned `sdkdiags.AppendErrorf` function calls ([#38854](https://github.com/hashicorp/terraform-provider-aws/issues/38854))
* resource/aws_db_event_subscription: Fix plan-time validation of `name` and `name_prefix` ([#38194](https://github.com/hashicorp/terraform-provider-aws/issues/38194))
* resource/aws_ecs_cluster_capacity_providers: Fix unreturned `sdkdiags.AppendErrorf` function calls ([#38854](https://github.com/hashicorp/terraform-provider-aws/issues/38854))
* resource/aws_ecs_service: Fix crash from nil `service_registries` item ([#38883](https://github.com/hashicorp/terraform-provider-aws/issues/38883))
* resource/aws_ecs_task_definition: Fix perpetual `container_definitions` diffs on `healthCheck`'s default values ([#38872](https://github.com/hashicorp/terraform-provider-aws/issues/38872))
* resource/aws_ecs_task_definition: Prevent lowercasing of the first character of JSON keys in `container_definitions.dockerLabels` ([#38804](https://github.com/hashicorp/terraform-provider-aws/issues/38804))
* resource/aws_ecs_task_definition: Remove `null`s from `container_definition` array fields ([#38870](https://github.com/hashicorp/terraform-provider-aws/issues/38870))
* resource/aws_elasticache_replication_group: Fix crash when setting `replicas_per_node_group` if node groups are empty ([#38797](https://github.com/hashicorp/terraform-provider-aws/issues/38797))
* resource/aws_fms_policy: Fix unreturned `sdkdiags.AppendErrorf` function calls ([#38854](https://github.com/hashicorp/terraform-provider-aws/issues/38854))
* resource/aws_grafana_workspace: Fix crash when empty `network_access_control` block is configured ([#38775](https://github.com/hashicorp/terraform-provider-aws/issues/38775))
* resource/aws_grafana_workspace: Fix crash when empty `vpc_configuration` block is configured ([#38775](https://github.com/hashicorp/terraform-provider-aws/issues/38775))
* resource/aws_iot_thing_group: Fix crash when empty `attribute_payload` block is configured ([#38776](https://github.com/hashicorp/terraform-provider-aws/issues/38776))
* resource/aws_lexv2models_slot_type: Fix slot_type_values to have sample_value attribute ([#38856](https://github.com/hashicorp/terraform-provider-aws/issues/38856))
* resource/aws_networkmanager_connect_peer: Set all `configuration.bgp_configurations` on Read ([#38798](https://github.com/hashicorp/terraform-provider-aws/issues/38798))
* resource/aws_redshift_cluster: Set `encrypted` on snapshot restore, when enabled ([#38828](https://github.com/hashicorp/terraform-provider-aws/issues/38828))
* resource/aws_rolesanywhere_profile: Fix unreturned `sdkdiags.AppendErrorf` function calls ([#38854](https://github.com/hashicorp/terraform-provider-aws/issues/38854))
* resource/aws_rolesanywhere_trust_anchor: Fix unreturned `sdkdiags.AppendErrorf` function calls ([#38854](https://github.com/hashicorp/terraform-provider-aws/issues/38854))
* resource/aws_s3_bucket_lifecycle_configuration: Fix unreturned `sdkdiags.AppendErrorf` function calls ([#38854](https://github.com/hashicorp/terraform-provider-aws/issues/38854))

## 5.62.0 (August  8, 2024)

FEATURES:

* **New Data Source:** `aws_rds_cluster_parameter_group` ([#38416](https://github.com/hashicorp/terraform-provider-aws/issues/38416))
* **New Data Source:** `aws_secretsmanager_secret_versions` ([#35411](https://github.com/hashicorp/terraform-provider-aws/issues/35411))
* **New Resource:** `aws_ebs_snapshot_block_public_access` ([#38641](https://github.com/hashicorp/terraform-provider-aws/issues/38641))
* **New Resource:** `aws_rds_integration` ([#35199](https://github.com/hashicorp/terraform-provider-aws/issues/35199))

ENHANCEMENTS:

* data-source/aws_s3_bucket_object: Expand content types that can be read from S3 to include include `application/x-sql` ([#38737](https://github.com/hashicorp/terraform-provider-aws/issues/38737))
* data-source/aws_s3_object: Expand content types that can be read from S3 to include `application/x-sql` ([#38737](https://github.com/hashicorp/terraform-provider-aws/issues/38737))
* provider: Allow `default_tags` to be set by environment variables ([#33339](https://github.com/hashicorp/terraform-provider-aws/issues/33339))
* provider: Allow `ignore_tags.keys` and `ignore_tags.key_prefixes` to be set by environment variables ([#35264](https://github.com/hashicorp/terraform-provider-aws/issues/35264))
* resource/aws_db_option_group: Add `skip_destroy` argument ([#29663](https://github.com/hashicorp/terraform-provider-aws/issues/29663))
* resource/aws_db_parameter_group: Add `skip_destroy` argument ([#29663](https://github.com/hashicorp/terraform-provider-aws/issues/29663))
* resource/aws_dx_macsec_key_association: Add plan-time validation of `secret_arn` ([#37213](https://github.com/hashicorp/terraform-provider-aws/issues/37213))
* resource/aws_ecs_service: Add `force_delete` argument ([#38707](https://github.com/hashicorp/terraform-provider-aws/issues/38707))
* resource/aws_grafana_license_association: Add `grafana_token` argument ([#38743](https://github.com/hashicorp/terraform-provider-aws/issues/38743))
* resource/aws_lb_target_group: Add `target_health_state.unhealthy_draining_interval` argument ([#38654](https://github.com/hashicorp/terraform-provider-aws/issues/38654))
* resource/aws_lexv2models_slot: Add `sub_slot_setting` attribute ([#38698](https://github.com/hashicorp/terraform-provider-aws/issues/38698))

BUG FIXES:

* data-source/aws_ecr_repository_creation_template: Support `ROOT` as a valid value for `prefix` ([#38685](https://github.com/hashicorp/terraform-provider-aws/issues/38685))
* data-source/aws_msk_broker_nodes: Filter out nodes with no broker info ([#38042](https://github.com/hashicorp/terraform-provider-aws/issues/38042))
* resource/aws_appconfig_configuration_profile: Increase `name` max length validation to 128 ([#37539](https://github.com/hashicorp/terraform-provider-aws/issues/37539))
* resource/aws_batch_job_definition: Fix panic when checking `eks_properties` for job updates ([#38716](https://github.com/hashicorp/terraform-provider-aws/issues/38716))
* resource/aws_batch_job_definition: Fix panic when checking `retry_strategy` for job updates ([#38716](https://github.com/hashicorp/terraform-provider-aws/issues/38716))
* resource/aws_batch_job_definition: Fix panic when checking `timeout` for job updates ([#38716](https://github.com/hashicorp/terraform-provider-aws/issues/38716))
* resource/aws_ec2_capacity_block_reservation: Fix error during apply for missing `created_date` attribute ([#38689](https://github.com/hashicorp/terraform-provider-aws/issues/38689))
* resource/aws_ecr_repository_creation_template: Support `ROOT` as a valid value for `prefix` ([#38685](https://github.com/hashicorp/terraform-provider-aws/issues/38685))
* resource/aws_elbv2_trust_store_revocation: Fix to properly return errors during resource creation ([#38756](https://github.com/hashicorp/terraform-provider-aws/issues/38756))
* resource/aws_emr_cluster: Fix panic when reading an instance fleet with an empty `launch_specifications` argument ([#38773](https://github.com/hashicorp/terraform-provider-aws/issues/38773))
* resource/aws_lexv2models_bot: Handle `PreconditionFailedException` on delete for resources deleted out-of-band ([#38661](https://github.com/hashicorp/terraform-provider-aws/issues/38661))
* resource/aws_lexv2models_bot_locale: Handle `PreconditionFailedException` on delete for resources deleted out-of-band ([#38661](https://github.com/hashicorp/terraform-provider-aws/issues/38661))
* resource/aws_lexv2models_bot_version: Handle `PreconditionFailedException` on delete for resources deleted out-of-band ([#38661](https://github.com/hashicorp/terraform-provider-aws/issues/38661))
* resource/aws_networkmanager_core_network: Fix `$.network-function-groups: null found, array expected` errors when creating resource with `create_base_policy` argument ([#38642](https://github.com/hashicorp/terraform-provider-aws/issues/38642))
* resource/aws_quicksight_account_subscription: Fix panic when read returns nil account info ([#38752](https://github.com/hashicorp/terraform-provider-aws/issues/38752))
* resource/aws_sfn_state_machine: Mark `revision_id` and `state_machine_version_arn` as Computed on update if `publish` is `true` ([#38657](https://github.com/hashicorp/terraform-provider-aws/issues/38657))

## 5.61.0 (August  1, 2024)

NOTES:

* resource/aws_chatbot_teams_channel_configuration: This resource is provided on a best-effort basis, and we welcome the community's help in testing it. ([#38630](https://github.com/hashicorp/terraform-provider-aws/issues/38630))

FEATURES:

* **New Data Source:** `aws_ecr_repository_creation_template` ([#38597](https://github.com/hashicorp/terraform-provider-aws/issues/38597))
* **New Resource:** `aws_chatbot_slack_channel_configuration` ([#38124](https://github.com/hashicorp/terraform-provider-aws/issues/38124))
* **New Resource:** `aws_chatbot_teams_channel_configuration` ([#38630](https://github.com/hashicorp/terraform-provider-aws/issues/38630))
* **New Resource:** `aws_datazone_glossary` ([#38602](https://github.com/hashicorp/terraform-provider-aws/issues/38602))
* **New Resource:** `aws_ecr_repository_creation_template` ([#38597](https://github.com/hashicorp/terraform-provider-aws/issues/38597))
* **New Resource:** `aws_timestreaminfluxdb_db_instance` ([#37963](https://github.com/hashicorp/terraform-provider-aws/issues/37963))

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `upgrade_policy` attribute ([#38573](https://github.com/hashicorp/terraform-provider-aws/issues/38573))
* data-source/aws_sagemaker_prebuilt_ecr_image: Support additional `repository_name` values. See [documentation](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/sagemaker_prebuilt_ecr_image#repository_name) for details ([#38575](https://github.com/hashicorp/terraform-provider-aws/issues/38575))
* resource/aws_appsync_graphql_api: Add `enhanced_metrics_config` configuration block ([#38570](https://github.com/hashicorp/terraform-provider-aws/issues/38570))
* resource/aws_db_instance: Add `upgrade_storage_config` argument ([#36904](https://github.com/hashicorp/terraform-provider-aws/issues/36904))
* resource/aws_default_vpc: Support `ipv6_cidr_block` sizes between `/44` and `/60` in increments of /4 ([#35614](https://github.com/hashicorp/terraform-provider-aws/issues/35614))
* resource/aws_default_vpc: Support `ipv6_netmask_length` values between `44` and `60` in increments of 4 ([#35614](https://github.com/hashicorp/terraform-provider-aws/issues/35614))
* resource/aws_eks_cluster: Add `upgrade_policy` configuration block ([#38573](https://github.com/hashicorp/terraform-provider-aws/issues/38573))
* resource/aws_elasticache_user_group_association: Add configurable create and delete timeouts ([#38559](https://github.com/hashicorp/terraform-provider-aws/issues/38559))
* resource/aws_pipes_pipe: Add `log_configuration.include_execution_data` argument ([#38569](https://github.com/hashicorp/terraform-provider-aws/issues/38569))
* resource/aws_rds_cluster: Add `performance_insights_enabled`, `performance_insights_kms_key_id`, and `performance_insights_retention_period` arguments ([#29415](https://github.com/hashicorp/terraform-provider-aws/issues/29415))
* resource/aws_rds_cluster: Add `restore_to_point_in_time.source_cluster_resource_id` argument ([#38540](https://github.com/hashicorp/terraform-provider-aws/issues/38540))
* resource/aws_rds_cluster: Mark `restore_to_point_in_time.source_cluster_identifier` as Optional ([#38540](https://github.com/hashicorp/terraform-provider-aws/issues/38540))
* resource/aws_sfn_activity: Add `encryption_configuration` configuration block to support the use of Customer Managed Keys with AWS KMS to encrypt Step Functions Activity resources ([#38574](https://github.com/hashicorp/terraform-provider-aws/issues/38574))
* resource/aws_sfn_state_machine: Add `encryption_configuration` configuration block to support the use of Customer Managed Keys with AWS KMS to encrypt Step Functions State Machine resources ([#38574](https://github.com/hashicorp/terraform-provider-aws/issues/38574))
* resource/aws_ssm_patch_baseline: Remove empty fields from `json` attribute value ([#35950](https://github.com/hashicorp/terraform-provider-aws/issues/35950))
* resource/aws_storagegateway_file_system_association: Add configurable timeouts ([#38554](https://github.com/hashicorp/terraform-provider-aws/issues/38554))
* resource/aws_vpc: Support `ipv6_cidr_block` sizes between `/44` and `/60` in increments of /4 ([#35614](https://github.com/hashicorp/terraform-provider-aws/issues/35614))
* resource/aws_vpc: Support `ipv6_netmask_length` values between `44` and `60` in increments of 4 ([#35614](https://github.com/hashicorp/terraform-provider-aws/issues/35614))
* resource/aws_vpc_ipv6_cidr_block_association: Add `assign_generated_ipv6_cidr_block` and `ipv6_pool` arguments ([#27274](https://github.com/hashicorp/terraform-provider-aws/issues/27274))
* resource/aws_vpc_ipv6_cidr_block_association: Support `ipv6_cidr_block` sizes between `/44` and `/60` in increments of /4 ([#35614](https://github.com/hashicorp/terraform-provider-aws/issues/35614))
* resource/aws_vpc_ipv6_cidr_block_association: Support `ipv6_netmask_length` values between `44` and `60` in increments of 4 ([#35614](https://github.com/hashicorp/terraform-provider-aws/issues/35614))
* resource/aws_vpc_security_group_egress_rule: Add `tags` to the `AuthorizeSecurityGroupEgress` EC2 API call instead of making a separate `CreateTags` call ([#35614](https://github.com/hashicorp/terraform-provider-aws/issues/35614))
* resource/aws_vpc_security_group_ingress_rule: Add `tags` to the `AuthorizeSecurityGroupIngress` EC2 API call instead of making a separate `CreateTags` call ([#35614](https://github.com/hashicorp/terraform-provider-aws/issues/35614))
* resource/aws_wafv2_web_acl: Add `rule_json` attribute to allow raw JSON for rules. ([#38309](https://github.com/hashicorp/terraform-provider-aws/issues/38309))

BUG FIXES:

* data-source/aws_appstream_image: Fix issue where the most recent image is not returned ([#38571](https://github.com/hashicorp/terraform-provider-aws/issues/38571))
* data-source/aws_networkmanager_core_network_policy_document: Fix `CoreNetworkPolicyException` when putting policy with single wildcard in `when_sent_to` ([#38595](https://github.com/hashicorp/terraform-provider-aws/issues/38595))
* resource/aws_cloudsearch_domain: Fix `index_name` character length validation ([#38509](https://github.com/hashicorp/terraform-provider-aws/issues/38509))
* resource/aws_ecs_task_definition: Ensure that JSON keys in `container_definitions` start with a lowercase letter ([#38622](https://github.com/hashicorp/terraform-provider-aws/issues/38622))
* resource/aws_iot_provisioning_template: Properly send `type` argument on create when configured ([#38640](https://github.com/hashicorp/terraform-provider-aws/issues/38640))
* resource/aws_opensearchserverless_security_policy: Normalize `policy` content to prevent persistent differences ([#38604](https://github.com/hashicorp/terraform-provider-aws/issues/38604))
* resource/aws_pipes_pipe: Don't reset `target_parameters` if the configured value has not changed ([#38598](https://github.com/hashicorp/terraform-provider-aws/issues/38598))
* resource/aws_rds_instance: Allow `domain_dns_ips` to use single DNS server IP ([#36500](https://github.com/hashicorp/terraform-provider-aws/issues/36500))
* resource/aws_sagemaker_domain: Properly send `domain_settings.r_studio_server_pro_domain_settings.r_studio_package_manager_url` argument on create ([#38547](https://github.com/hashicorp/terraform-provider-aws/issues/38547))
* resource/aws_vpc_ipam_pool_cidr_allocation: Set `description` on Read ([#38618](https://github.com/hashicorp/terraform-provider-aws/issues/38618))
* resource/aws_vpc_ipam_pool_cidr_allocation: Set `netmask_length` on Read ([#38618](https://github.com/hashicorp/terraform-provider-aws/issues/38618))

## 5.60.0 (July 25, 2024)

NOTES:

* resource/aws_shield_subscription: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#37637](https://github.com/hashicorp/terraform-provider-aws/issues/37637))

FEATURES:

* **New Data Source:** `aws_service_principal` ([#38307](https://github.com/hashicorp/terraform-provider-aws/issues/38307))
* **New Resource:** `aws_shield_subscription` ([#37637](https://github.com/hashicorp/terraform-provider-aws/issues/37637))

ENHANCEMENTS:

* data-source/aws_cloudwatch_event_bus: Add `kms_key_identifier` attribute ([#38492](https://github.com/hashicorp/terraform-provider-aws/issues/38492))
* data-source/aws_cur_report_definition: Add `tags` attribute ([#38483](https://github.com/hashicorp/terraform-provider-aws/issues/38483))
* resource/aws_appflow_flow: Add `metadata_catalog_config` attribute ([#37566](https://github.com/hashicorp/terraform-provider-aws/issues/37566))
* resource/aws_appflow_flow: Add `prefix_hierarchy` attribute to `destination_flow_config.s3.s3_output_format_config` ([#37566](https://github.com/hashicorp/terraform-provider-aws/issues/37566))
* resource/aws_batch_job_definition: Add `eks_properties.*.pod_properties.*.image_pull_secret` argument ([#38517](https://github.com/hashicorp/terraform-provider-aws/issues/38517))
* resource/aws_cloudformation_stack_set_instance: Add `operation_preferences.concurrency_mode` argument ([#38498](https://github.com/hashicorp/terraform-provider-aws/issues/38498))
* resource/aws_cloudwatch_event_bus: Add `kms_key_identifier` argument ([#38492](https://github.com/hashicorp/terraform-provider-aws/issues/38492))
* resource/aws_cur_report_definition: Add `tags` argument and `tags_all` attribute ([#38483](https://github.com/hashicorp/terraform-provider-aws/issues/38483))
* resource/aws_db_cluster_snapshot: Add `shared_accounts` argument ([#34885](https://github.com/hashicorp/terraform-provider-aws/issues/34885))
* resource/aws_db_snapshot_copy: Add `shared_accounts` argument ([#34843](https://github.com/hashicorp/terraform-provider-aws/issues/34843))
* resource/aws_glue_connection: Add `AZURECOSMOS`, `AZURESQL`, `BIGQUERY`, `OPENSEARCH`, and `SNOWFLAKE` as valid values for the `connection_type` argument and `SparkProperties` as a valid value for the `connection_properties` argument ([#37731](https://github.com/hashicorp/terraform-provider-aws/issues/37731))
* resource/aws_iam_role: Change from partial resource creation to resource creation failed if an `inline_policy` fails to create ([#38477](https://github.com/hashicorp/terraform-provider-aws/issues/38477))
* resource/aws_rds_cluster: Add `scaling_configuration.seconds_before_timeout` argument ([#38451](https://github.com/hashicorp/terraform-provider-aws/issues/38451))
* resource/aws_sesv2_configuration_set_event_destination: Add `event_destination.event_bridge_destination` configuration block ([#38458](https://github.com/hashicorp/terraform-provider-aws/issues/38458))
* resource/aws_timestreamwrite_table: Fix `runtime error: invalid memory address or nil pointer dereference` panic when reading a non-existent table ([#38512](https://github.com/hashicorp/terraform-provider-aws/issues/38512))

BUG FIXES:

* data-source/aws_fsx_ontap_storage_virtual_machine: Correctly set `tags` on Read ([#38343](https://github.com/hashicorp/terraform-provider-aws/issues/38343))
* data-source/aws_fsx_openzfs_snapshot: Correctly set `tags` on Read ([#38343](https://github.com/hashicorp/terraform-provider-aws/issues/38343))
* resource/aws_ce_cost_category: Fix perpetual diff with the `rule` argument on update ([#38449](https://github.com/hashicorp/terraform-provider-aws/issues/38449))
* resource/aws_codebuild_webhook: Remove errant validation on `scope_configuration.domain` argument ([#38513](https://github.com/hashicorp/terraform-provider-aws/issues/38513))
* resource/aws_ecs_service: Fix `error marshaling prior state: a number is required` when upgrading from v5.58.0 to v5.59.0 ([#38490](https://github.com/hashicorp/terraform-provider-aws/issues/38490))
* resource/aws_ecs_task_definition: Fix `Provider produced inconsistent final plan` errors when `container_definitions` is [unknown](https://developer.hashicorp.com/terraform/language/expressions/references#values-not-yet-known) ([#38471](https://github.com/hashicorp/terraform-provider-aws/issues/38471))
* resource/aws_elasticache_replication_group: Fix `error marshaling prior state` when upgrading from v4.67.0 to v5.59.0 ([#38476](https://github.com/hashicorp/terraform-provider-aws/issues/38476))
* resource/aws_fsx_openzfs_volume: Correctly set `tags` on Read ([#38343](https://github.com/hashicorp/terraform-provider-aws/issues/38343))
* resource/aws_rds_cluster: Mark `ca_certificate_identifier` as Computed ([#38437](https://github.com/hashicorp/terraform-provider-aws/issues/38437))
* resource/aws_rds_cluster: Use the configured `copy_tags_to_snapshot` value when `restore_to_point_in_time` is set ([#34044](https://github.com/hashicorp/terraform-provider-aws/issues/34044))
* resource/aws_rds_cluster: Wait for no pending modified values on Update if `apply_immediately` is `true`. This fixes `InvalidParameterCombination` errors when updating `engine_version` ([#38437](https://github.com/hashicorp/terraform-provider-aws/issues/38437))

## 5.59.0 (July 19, 2024)

FEATURES:

* resource/aws_kinesis_firehose_delivery_stream: Add `secrets_manager_configuration` to `redshift_configuration`, `snowflake_configuration`, and `splunk_configuration` ([#38151](https://github.com/hashicorp/terraform-provider-aws/issues/38151))
* **New Data Source:** `aws_cloudfront_origin_access_control` ([#36301](https://github.com/hashicorp/terraform-provider-aws/issues/36301))
* **New Data Source:** `aws_timestreamwrite_database` ([#36368](https://github.com/hashicorp/terraform-provider-aws/issues/36368))
* **New Data Source:** `aws_timestreamwrite_table` ([#36599](https://github.com/hashicorp/terraform-provider-aws/issues/36599))
* **New Resource:** `aws_datazone_project` ([#38345](https://github.com/hashicorp/terraform-provider-aws/issues/38345))
* **New Resource:** `aws_grafana_workspace_service_account` ([#38101](https://github.com/hashicorp/terraform-provider-aws/issues/38101))
* **New Resource:** `aws_grafana_workspace_service_account_token` ([#38101](https://github.com/hashicorp/terraform-provider-aws/issues/38101))
* **New Resource:** `aws_rds_certificate` ([#35003](https://github.com/hashicorp/terraform-provider-aws/issues/35003))
* **New Resource:** `aws_rekognition_stream_processor` ([#37536](https://github.com/hashicorp/terraform-provider-aws/issues/37536))

ENHANCEMENTS:

* data-source/aws_elasticache_replication_group: Add `cluster_mode` attribute ([#38002](https://github.com/hashicorp/terraform-provider-aws/issues/38002))
* data-source/aws_lakeformation_data_lake_settings: Add `allow_full_table_external_data_access` attribute ([#34474](https://github.com/hashicorp/terraform-provider-aws/issues/34474))
* data-source/aws_msk_cluster: Add `broker_node_group_info` attribute ([#37705](https://github.com/hashicorp/terraform-provider-aws/issues/37705))
* resource/aws_bedrockagent_agent : Add `skip_resource_in_use_check` argument ([#37586](https://github.com/hashicorp/terraform-provider-aws/issues/37586))
* resource/aws_bedrockagent_agent_action_group: Add `action_group_executor.custom_control` argument ([#37484](https://github.com/hashicorp/terraform-provider-aws/issues/37484))
* resource/aws_bedrockagent_agent_action_group: Add `function_schema` configuration block ([#37484](https://github.com/hashicorp/terraform-provider-aws/issues/37484))
* resource/aws_bedrockagent_agent_alias : Add `routing_configuration.provisioned_throughput` argument ([#37520](https://github.com/hashicorp/terraform-provider-aws/issues/37520))
* resource/aws_codebuild_webhook: Add `scope_configuration` argument ([#38199](https://github.com/hashicorp/terraform-provider-aws/issues/38199))
* resource/aws_codepipeline: Add `timeout_in_minutes` argument to the `action` configuration block ([#36316](https://github.com/hashicorp/terraform-provider-aws/issues/36316))
* resource/aws_db_instance: Add `engine_lifecycle_support` argument ([#37708](https://github.com/hashicorp/terraform-provider-aws/issues/37708))
* resource/aws_ecs_cluster: Add `configuration.managed_storage_configuration` argument ([#37932](https://github.com/hashicorp/terraform-provider-aws/issues/37932))
* resource/aws_elasticache_replication_group: Add `cluster_mode` argument ([#38002](https://github.com/hashicorp/terraform-provider-aws/issues/38002))
* resource/aws_emrserverless_application: Add `interactive_configuration` argument ([#37889](https://github.com/hashicorp/terraform-provider-aws/issues/37889))
* resource/aws_fis_experiment_template: Add `experiment_options` configuration block ([#36900](https://github.com/hashicorp/terraform-provider-aws/issues/36900))
* resource/aws_fsx_lustre_file_system: Add `final_backup_tags` and `skip_final_backup` arguments ([#37717](https://github.com/hashicorp/terraform-provider-aws/issues/37717))
* resource/aws_fsx_ontap_volume: Add `final_backup_tags` argument ([#37717](https://github.com/hashicorp/terraform-provider-aws/issues/37717))
* resource/aws_fsx_openzfs_file_system: Add `delete_options` and `final_backup_tags` arguments ([#37717](https://github.com/hashicorp/terraform-provider-aws/issues/37717))
* resource/aws_fsx_windows_file_system: Add `final_backup_tags` argument ([#37717](https://github.com/hashicorp/terraform-provider-aws/issues/37717))
* resource/aws_imagebuilder_image_pipeline: Add `execution_role` and `workflow` arguments ([#37317](https://github.com/hashicorp/terraform-provider-aws/issues/37317))
* resource/aws_kinesis_firehose_delivery_stream: Add `secrets_manager_configuration` to `http_endpoint_configuration` ([#38245](https://github.com/hashicorp/terraform-provider-aws/issues/38245))
* resource/aws_kinesisanalyticsv2_application: Support `FLINK-1_19` as a valid value for `runtime_environment` ([#38350](https://github.com/hashicorp/terraform-provider-aws/issues/38350))
* resource/aws_lakeformation_data_lake_settings: Add `allow_full_table_external_data_access` attribute ([#34474](https://github.com/hashicorp/terraform-provider-aws/issues/34474))
* resource/aws_lb_target_group: Add `target_group_health` configuration block ([#37082](https://github.com/hashicorp/terraform-provider-aws/issues/37082))
* resource/aws_msk_replicator: Add `starting_position` argument ([#36968](https://github.com/hashicorp/terraform-provider-aws/issues/36968))
* resource/aws_rds_cluster: Add `engine_lifecycle_support` argument ([#37708](https://github.com/hashicorp/terraform-provider-aws/issues/37708))
* resource/aws_rds_global_cluster: Add `engine_lifecycle_support` argument ([#37708](https://github.com/hashicorp/terraform-provider-aws/issues/37708))
* resource/aws_redshift_cluster_snapshot: Set `arn` from `DescribeClusterSnapshots` API response ([#37996](https://github.com/hashicorp/terraform-provider-aws/issues/37996))
* resource/aws_vpclattice_listener: Support `TLS_PASSTHROUGH` as a valid value for `protocol` ([#37964](https://github.com/hashicorp/terraform-provider-aws/issues/37964))
* resource/aws_wafv2_web_acl: Add `enable_machine_learning` to `aws_managed_rules_bot_control_rule_set` configuration block ([#37006](https://github.com/hashicorp/terraform-provider-aws/issues/37006))

BUG FIXES:

* data-source/aws_efs_access_point: Set `id` the the access point ID, not the file system ID. This fixes a regression introduced in [v5.58.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#5580-july-11-2024) ([#38372](https://github.com/hashicorp/terraform-provider-aws/issues/38372))
* data-source/aws_lb_listener: Correctly set `default_action.target_group_arn` ([#37348](https://github.com/hashicorp/terraform-provider-aws/issues/37348))
* resource/aws_chime_voice_connector_group: Properly handle voice connector groups deleted out of band ([#36774](https://github.com/hashicorp/terraform-provider-aws/issues/36774))
* resource/aws_codebuild_project: Fix unsetting `concurrent_build_limit` ([#37748](https://github.com/hashicorp/terraform-provider-aws/issues/37748))
* resource/aws_codepipeline: Mark `trigger` as Computed ([#36316](https://github.com/hashicorp/terraform-provider-aws/issues/36316))
* resource/aws_ecs_service: Change `volume_configuration.managed_ebs_volume.throughput` from `TypeString` to `TypeInt` ([#38109](https://github.com/hashicorp/terraform-provider-aws/issues/38109))
* resource/aws_elasticache_replication_group: Allows setting `replicas_per_node_group` to `0` and sets the maximum to `5`. ([#38396](https://github.com/hashicorp/terraform-provider-aws/issues/38396))
* resource/aws_elasticache_replication_group: Requires `description`. ([#38396](https://github.com/hashicorp/terraform-provider-aws/issues/38396))
* resource/aws_elasticache_replication_group: When `num_cache_clusters` is set, prevents setting `replicas_per_node_group`. ([#38396](https://github.com/hashicorp/terraform-provider-aws/issues/38396))
* resource/aws_elasticache_replication_group: `num_cache_clusters` must be at least 2 when `automatic_failover_enabled` is `true`. ([#38396](https://github.com/hashicorp/terraform-provider-aws/issues/38396))
* resource/aws_elastictranscoder_pipeline: Properly handle NotFound exceptions during deletion ([#38018](https://github.com/hashicorp/terraform-provider-aws/issues/38018))
* resource/aws_elastictranscoder_preset: Properly handle NotFound exceptions during deletion ([#38018](https://github.com/hashicorp/terraform-provider-aws/issues/38018))
* resource/aws_lb_target_group: Use the configured `ip_address_type` value when `target_type` is `instance` ([#36423](https://github.com/hashicorp/terraform-provider-aws/issues/36423))
* resource/aws_lb_trust_store: Wait until trust store is `ACTIVE` on resource Create ([#38332](https://github.com/hashicorp/terraform-provider-aws/issues/38332))
* resource/aws_pinpoint_app: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panic when `campaign_hook` is empty (`{}`) ([#38323](https://github.com/hashicorp/terraform-provider-aws/issues/38323))
* resource/aws_transfer_server: Add supported values `TransferSecurityPolicy-FIPS-2024-05`, `TransferSecurityPolicy-Restricted-2018-11`, and `TransferSecurityPolicy-Restricted-2020-06` for the `security_policy_name` argument ([#38425](https://github.com/hashicorp/terraform-provider-aws/issues/38425))

## 5.58.0 (July 11, 2024)

FEATURES:

* **New Resource:** `aws_cloudwatch_log_account_policy` ([#38328](https://github.com/hashicorp/terraform-provider-aws/issues/38328))
* **New Resource:** `aws_verifiedpermissions_identity_source` ([#38181](https://github.com/hashicorp/terraform-provider-aws/issues/38181))

ENHANCEMENTS:

* data-source/aws_launch_template: Add `network_interfaces.primary_ipv6` attribute ([#37142](https://github.com/hashicorp/terraform-provider-aws/issues/37142))
* data-source/aws_mskconnect_connector: Add `tags` attribute ([#38270](https://github.com/hashicorp/terraform-provider-aws/issues/38270))
* data-source/aws_mskconnect_custom_plugin: Add `tags` attribute ([#38270](https://github.com/hashicorp/terraform-provider-aws/issues/38270))
* data-source/aws_mskconnect_worker_configuration: Add `tags` attribute ([#38270](https://github.com/hashicorp/terraform-provider-aws/issues/38270))
* data-source/aws_oam_link: Add `link_configuration` attribute ([#38277](https://github.com/hashicorp/terraform-provider-aws/issues/38277))
* resource/aws_cloudformation_stack_set_instance: Extend `deployment_targets` argument. ([#37898](https://github.com/hashicorp/terraform-provider-aws/issues/37898))
* resource/aws_cloudtrail_event_data_store: Add `billing_mode` argument ([#38273](https://github.com/hashicorp/terraform-provider-aws/issues/38273))
* resource/aws_db_instance: Fix `InvalidParameterCombination: A parameter group can't be specified during Read Replica creation for the following DB engine: postgres` errors ([#38227](https://github.com/hashicorp/terraform-provider-aws/issues/38227))
* resource/aws_ec2_capacity_reservation: Add configurable timeouts ([#36754](https://github.com/hashicorp/terraform-provider-aws/issues/36754))
* resource/aws_ec2_capacity_reservation: Retry `InsufficientInstanceCapacity` errors ([#36754](https://github.com/hashicorp/terraform-provider-aws/issues/36754))
* resource/aws_eks_cluster: Add `bootstrap_self_managed_addons` argument ([#38162](https://github.com/hashicorp/terraform-provider-aws/issues/38162))
* resource/aws_fms_policy: Add `resource_set_ids` attribute ([#38161](https://github.com/hashicorp/terraform-provider-aws/issues/38161))
* resource/aws_fsx_ontap_file_system: Add `384`, `768`, `1536`, `3072`, and `6144` as valid values for `throughput_capacity` ([#38308](https://github.com/hashicorp/terraform-provider-aws/issues/38308))
* resource/aws_fsx_ontap_file_system: Add `384`, `768`, and `1536` as valid values for `throughput_capacity_per_ha_pair` ([#38308](https://github.com/hashicorp/terraform-provider-aws/issues/38308))
* resource/aws_fsx_ontap_file_system: Add `MULTI_AZ_2` as a valid value for `deployment_type` ([#38308](https://github.com/hashicorp/terraform-provider-aws/issues/38308))
* resource/aws_globalaccelerator_cross_account_attachment: Add `cidr_block` argument to `resource` configuration block ([#38196](https://github.com/hashicorp/terraform-provider-aws/issues/38196))
* resource/aws_iam_server_certificate: Add configurable `delete` timeout ([#38212](https://github.com/hashicorp/terraform-provider-aws/issues/38212))
* resource/aws_launch_template: Add `network_interfaces.primary_ipv6` argument ([#37142](https://github.com/hashicorp/terraform-provider-aws/issues/37142))
* resource/aws_mskconnect_connector: Add `tags` argument and `tags_all` attribute ([#38270](https://github.com/hashicorp/terraform-provider-aws/issues/38270))
* resource/aws_mskconnect_custom_plugin: Add `tags` argument and `tags_all` attribute ([#38270](https://github.com/hashicorp/terraform-provider-aws/issues/38270))
* resource/aws_mskconnect_worker_configuration: Add `tags` argument and `tags_all` attribute ([#38270](https://github.com/hashicorp/terraform-provider-aws/issues/38270))
* resource/aws_mskconnect_worker_configuration: Add resource deletion logic ([#38270](https://github.com/hashicorp/terraform-provider-aws/issues/38270))
* resource/aws_oam_link: Add `link_configuration` argument ([#38277](https://github.com/hashicorp/terraform-provider-aws/issues/38277))
* resource/aws_rds_cluster: Add `ca_certificate_identifier` argument and `ca_certificate_valid_till` attribute ([#37108](https://github.com/hashicorp/terraform-provider-aws/issues/37108))
* resource/aws_ssm_association: Add `tags` argument and `tags_all` attribute ([#38271](https://github.com/hashicorp/terraform-provider-aws/issues/38271))

BUG FIXES:

* aws_dx_lag: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_dynamodb_kinesis_streaming_destination: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_ec2_capacity_block_reservation: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_opensearchserverless_access_policy: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_opensearchserverless_collection: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_opensearchserverless_security_config: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_opensearchserverless_security_policy: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_opensearchserverless_vpc_endpoint: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_ram_principal_association: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* aws_route_table: Checks for errors other than NotFound when reading. ([#38292](https://github.com/hashicorp/terraform-provider-aws/issues/38292))
* data-source/aws_ecr_repository: Fix issue where the `tags` attribute is not set ([#38272](https://github.com/hashicorp/terraform-provider-aws/issues/38272))
* data-source/aws_eks_cluster: Add `access_config.bootstrap_cluster_creator_admin_permissions` attribute ([#38295](https://github.com/hashicorp/terraform-provider-aws/issues/38295))
* resource/aws_appstream_fleet: Support `0` as a valid value for `idle_disconnect_timeout_in_seconds` ([#38274](https://github.com/hashicorp/terraform-provider-aws/issues/38274))
* resource/aws_cloudformation_stack_set_instance: Add `ForceNew` to deployment_targets attributes to ensure a new resource is recreated when the deployment_targets argument is changed, which was not the case previously. ([#37898](https://github.com/hashicorp/terraform-provider-aws/issues/37898))
* resource/aws_db_instance: Correctly mark incomplete instances as [tainted](https://developer.hashicorp.com/terraform/cli/state/taint#the-tainted-status) during creation ([#38252](https://github.com/hashicorp/terraform-provider-aws/issues/38252))
* resource/aws_eks_cluster: Set `access_config.bootstrap_cluster_creator_admin_permissions` to `true` on Read for clusters with no `access_config` configured. This allows in-place updates of existing clusters when `access_config` is configured ([#38295](https://github.com/hashicorp/terraform-provider-aws/issues/38295))
* resource/aws_elasticache_serverless_cache: Allow `cache_usage_limits.data_storage.maximum`, `cache_usage_limits.data_storage.minimum`, `cache_usage_limits.ecpu_per_second.maximum` and `cache_usage_limits.ecpu_per_second.minimum` to be updated in-place ([#38269](https://github.com/hashicorp/terraform-provider-aws/issues/38269))
* resource/aws_mskconnect_connector: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panic when `log_delivery.worker_log_delivery` is empty (`{}`) ([#38270](https://github.com/hashicorp/terraform-provider-aws/issues/38270))

## 5.57.0 (July  4, 2024)

FEATURES:

* **New Data Source:** `aws_appstream_image` ([#38225](https://github.com/hashicorp/terraform-provider-aws/issues/38225))
* **New Data Source:** `aws_cognito_user_pool` ([#37399](https://github.com/hashicorp/terraform-provider-aws/issues/37399))
* **New Data Source:** `aws_ec2_transit_gateway_peering_attachments` ([#25743](https://github.com/hashicorp/terraform-provider-aws/issues/25743))
* **New Data Source:** `aws_transfer_connector` ([#38213](https://github.com/hashicorp/terraform-provider-aws/issues/38213))

ENHANCEMENTS:

* data-source/aws_backup_plan: Add `rule` attribute ([#37890](https://github.com/hashicorp/terraform-provider-aws/issues/37890))
* resource/aws_amplify_domain_association: Add `certificate_settings` argument ([#37105](https://github.com/hashicorp/terraform-provider-aws/issues/37105))
* resource/aws_ec2_transit_gateway_peering_attachment: Add `options` argument ([#36902](https://github.com/hashicorp/terraform-provider-aws/issues/36902))
* resource/aws_iot_authorizer: Add `tags` argument ([#37152](https://github.com/hashicorp/terraform-provider-aws/issues/37152))
* resource/aws_iot_topic_rule: Add `cloudwatch_logs.batch_mode` and `error_action.cloudwatch_logs.batch_mode` arguments ([#36772](https://github.com/hashicorp/terraform-provider-aws/issues/36772))
* resource/aws_sagemaker_endpoint_configuration: Add support for `InputAndOutput` in `capture_mode` ([#37726](https://github.com/hashicorp/terraform-provider-aws/issues/37726))

BUG FIXES:

* resource/aws_iot_provisioning_template: Fix `pre_provisioning_hook` update operation ([#37152](https://github.com/hashicorp/terraform-provider-aws/issues/37152))
* resource/aws_iot_topic_rule: Retry IAM eventual consistency errors on Update ([#36286](https://github.com/hashicorp/terraform-provider-aws/issues/36286))

## 5.56.1 (June 28, 2024)

BUG FIXES:

* data-source/aws_cognito_user_pool_client: Fix `InvalidParameterException: 2 validation errors detected` errors on Read ([#38168](https://github.com/hashicorp/terraform-provider-aws/issues/38168))
* resource/aws_cognito_user: Fix a bug that caused resource recreation for resources imported with certain [import ID](https://developer.hashicorp.com/terraform/language/import#import-id) formats ([#38182](https://github.com/hashicorp/terraform-provider-aws/issues/38182))
* resource/aws_cognito_user_pool: Fix `runtime error: index out of range [0] with length 0` panic when adding `lambda_config` ([#38184](https://github.com/hashicorp/terraform-provider-aws/issues/38184))

## 5.56.0 (June 27, 2024)

FEATURES:

* **New Resource:** `aws_appfabric_app_authorization_connection` ([#38084](https://github.com/hashicorp/terraform-provider-aws/issues/38084))
* **New Resource:** `aws_appfabric_ingestion` ([#37291](https://github.com/hashicorp/terraform-provider-aws/issues/37291))
* **New Resource:** `aws_appfabric_ingestion_destination` ([#37627](https://github.com/hashicorp/terraform-provider-aws/issues/37627))
* **New Resource:** `aws_networkfirewall_tls_inspection_configuration` ([#35168](https://github.com/hashicorp/terraform-provider-aws/issues/35168))
* **New Resource:** `aws_networkmonitor_monitor` ([#35722](https://github.com/hashicorp/terraform-provider-aws/issues/35722))
* **New Resource:** `aws_networkmonitor_probe` ([#35722](https://github.com/hashicorp/terraform-provider-aws/issues/35722))

ENHANCEMENTS:

* resource/aws_controltower_control: Add `parameters` argument and `arn` attribute ([#38071](https://github.com/hashicorp/terraform-provider-aws/issues/38071))
* resource/aws_networkfirewall_logging_configuration: Add plan-time validation of `firewall_arn` ([#35168](https://github.com/hashicorp/terraform-provider-aws/issues/35168))
* resource/aws_quicksight_account_subscription: Add `iam_identity_center_instance_arn` attribute ([#36830](https://github.com/hashicorp/terraform-provider-aws/issues/36830))
* resource/aws_route53_resolver_firewall_rule: Add `firewall_domain_redirection_action` argument ([#37242](https://github.com/hashicorp/terraform-provider-aws/issues/37242))
* resource/aws_route53_resolver_firewall_rule: Add `q_type` argument ([#38074](https://github.com/hashicorp/terraform-provider-aws/issues/38074))
* resource/aws_sagemaker_domain: Add `default_user_settings.canvas_app_settings.generative_ai_settings` configuration block ([#37139](https://github.com/hashicorp/terraform-provider-aws/issues/37139))
* resource/aws_sagemaker_domain: Add `default_user_settings.code_editor_app_settings.custom_image` configuration block ([#37153](https://github.com/hashicorp/terraform-provider-aws/issues/37153))
* resource/aws_sagemaker_endpoint_configuration: Add `production_variants.inference_ami_version` and `shadow_production_variants.inference_ami_version` arguments ([#38085](https://github.com/hashicorp/terraform-provider-aws/issues/38085))
* resource/aws_sagemaker_user_profile: Add `user_settings.canvas_app_settings.generative_ai_settings` configuration block ([#37139](https://github.com/hashicorp/terraform-provider-aws/issues/37139))
* resource/aws_sagemaker_user_profile: Add `user_settings.code_editor_app_settings.custom_image` configuration block ([#37153](https://github.com/hashicorp/terraform-provider-aws/issues/37153))
* resource/aws_sagemaker_workforce: add `oidc_config.authentication_request_extra_params` and `oidc_config.scope` arguments ([#38078](https://github.com/hashicorp/terraform-provider-aws/issues/38078))
* resource/aws_sagemaker_workteam: Add `worker_access_configuration` attribute ([#38087](https://github.com/hashicorp/terraform-provider-aws/issues/38087))
* resource/aws_wafv2_web_acl: Add `sensitivity_level` argument to `sqli_match_statement` configuration block ([#38077](https://github.com/hashicorp/terraform-provider-aws/issues/38077))

BUG FIXES:

* data-source/aws_ecs_service: Correctly set `tags` ([#38067](https://github.com/hashicorp/terraform-provider-aws/issues/38067))
* resource/aws_drs_replication_configuration_template: Fix issues preventing creation and deletion ([#38143](https://github.com/hashicorp/terraform-provider-aws/issues/38143))

## 5.55.0 (June 20, 2024)

FEATURES:

* **New Resource:** `aws_drs_replication_configuration_template` ([#26399](https://github.com/hashicorp/terraform-provider-aws/issues/26399))

ENHANCEMENTS:

* data-source/aws_autoscaling_group: Add `mixed_instances_policy.launch_template.override.instance_requirements.max_spot_price_as_percentage_of_optimal_on_demand_price` attribute ([#38003](https://github.com/hashicorp/terraform-provider-aws/issues/38003))
* data-source/aws_glue_catalog_table: Add `additional_locations` argument in `storage_descriptor` ([#37891](https://github.com/hashicorp/terraform-provider-aws/issues/37891))
* data-source/aws_launch_template: Add `instance_requirements.max_spot_price_as_percentage_of_optimal_on_demand_price` attribute ([#38003](https://github.com/hashicorp/terraform-provider-aws/issues/38003))
* data-source/aws_networkmanager_core_network_policy_document: Add `attachment_policies.action.add_to_network_function_group` argument ([#38013](https://github.com/hashicorp/terraform-provider-aws/issues/38013))
* data-source/aws_networkmanager_core_network_policy_document: Add `network_function_groups` configuration block ([#38013](https://github.com/hashicorp/terraform-provider-aws/issues/38013))
* data-source/aws_networkmanager_core_network_policy_document: Add `send-via` and `send-to` as valid values for `segment_actions.action` ([#38013](https://github.com/hashicorp/terraform-provider-aws/issues/38013))
* data-source/aws_networkmanager_core_network_policy_document: Add `single-hop` and `dual-hop` as valid values for `segment_actions.mode` ([#38013](https://github.com/hashicorp/terraform-provider-aws/issues/38013))
* data-source/aws_networkmanager_core_network_policy_document: Add `when_sent_to` and `via` configuration blocks to `segment_actions` ([#38013](https://github.com/hashicorp/terraform-provider-aws/issues/38013))
* resource/aws_api_gateway_integration: Increase maximum value of `timeout_milliseconds` from `29000` (29 seconds) to `300000` (5 minutes) ([#38010](https://github.com/hashicorp/terraform-provider-aws/issues/38010))
* resource/aws_appsync_api_key: Add `api_key_id` attribute ([#36568](https://github.com/hashicorp/terraform-provider-aws/issues/36568))
* resource/aws_autoscaling_group: Add `mixed_instances_policy.launch_template.override.instance_requirements.max_spot_price_as_percentage_of_optimal_on_demand_price` argument ([#38003](https://github.com/hashicorp/terraform-provider-aws/issues/38003))
* resource/aws_autoscaling_group: Add plan-time validation of `warm_pool.max_group_prepared_capacity` and `warm_pool.min_size` ([#37174](https://github.com/hashicorp/terraform-provider-aws/issues/37174))
* resource/aws_docdb_cluster: Add `restore_to_point_in_time` argument ([#37716](https://github.com/hashicorp/terraform-provider-aws/issues/37716))
* resource/aws_dynamodb_table: Adds validation for `ttl` values. ([#37991](https://github.com/hashicorp/terraform-provider-aws/issues/37991))
* resource/aws_ec2_fleet: Add `launch_template_config.override.instance_requirements.max_spot_price_as_percentage_of_optimal_on_demand_price` argument ([#38003](https://github.com/hashicorp/terraform-provider-aws/issues/38003))
* resource/aws_glue_catalog_table: Add `additional_locations` argument in `storage_descriptor` ([#37891](https://github.com/hashicorp/terraform-provider-aws/issues/37891))
* resource/aws_glue_job: Add `maintenance_window` argument ([#37760](https://github.com/hashicorp/terraform-provider-aws/issues/37760))
* resource/aws_launch_template: Add `instance_requirements.max_spot_price_as_percentage_of_optimal_on_demand_price` argument ([#38003](https://github.com/hashicorp/terraform-provider-aws/issues/38003))

BUG FIXES:

* data-source/aws_networkmanager_core_network_policy_document: Add correct `except` values to the returned JSON document when `segment_actions.share_with_except` is configured ([#38013](https://github.com/hashicorp/terraform-provider-aws/issues/38013))
* provider: Now falls back to non-FIPS endpoint if `use_fips_endpoint` is set and no FIPS endpoint is available ([#38057](https://github.com/hashicorp/terraform-provider-aws/issues/38057))
* resource/aws_autoscaling_group: Fix bug updating `warm_pool.max_group_prepared_capacity` to `0` ([#37174](https://github.com/hashicorp/terraform-provider-aws/issues/37174))
* resource/aws_dynamodb_table: Fixes perpetual diff when `ttl.attribute_name` is set when `ttl.enabled` is not set. ([#37991](https://github.com/hashicorp/terraform-provider-aws/issues/37991))
* resource/aws_ec2_network_insights_path: Mark `destination` as Optional ([#36966](https://github.com/hashicorp/terraform-provider-aws/issues/36966))
* resource/aws_lambda_event_source_mapping: Remove the upper limit on `scaling_config.maximum_concurrency` ([#37980](https://github.com/hashicorp/terraform-provider-aws/issues/37980))
* service/transitgateway: Fix resource Read pagination regression causing `NotFound` errors ([#38011](https://github.com/hashicorp/terraform-provider-aws/issues/38011))

## 5.54.1 (June 14, 2024)

BUG FIXES:

* data-source/aws_ami: Fix `interface conversion: interface {} is types.ProductCodeValues, not string` panic ([#37977](https://github.com/hashicorp/terraform-provider-aws/issues/37977))
* resource/aws_codebuild_project: Increase maximum values of `build_batch_config.timeout_in_mins` and `build_timeout` from `480` (8 hours) to `2160` (36 hours) ([#37970](https://github.com/hashicorp/terraform-provider-aws/issues/37970))

## 5.54.0 (June 14, 2024)

NOTES:

* resource/aws_ec2_capacity_block_reservation: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#37528](https://github.com/hashicorp/terraform-provider-aws/issues/37528))

FEATURES:

* **New Data Source:** `aws_ec2_capacity_block_offering` ([#37528](https://github.com/hashicorp/terraform-provider-aws/issues/37528))
* **New Resource:** `aws_appfabric_app_authorization` ([#37468](https://github.com/hashicorp/terraform-provider-aws/issues/37468))
* **New Resource:** `aws_appfabric_app_bundle` ([#37542](https://github.com/hashicorp/terraform-provider-aws/issues/37542))
* **New Resource:** `aws_ec2_capacity_block_reservation` ([#37528](https://github.com/hashicorp/terraform-provider-aws/issues/37528))
* **New Resource:** `aws_fms_resource_set` ([#37767](https://github.com/hashicorp/terraform-provider-aws/issues/37767))
* **New Resource:** `aws_guardduty_malware_protection_plan` ([#37919](https://github.com/hashicorp/terraform-provider-aws/issues/37919))

ENHANCEMENTS:

* data-source/aws_opensearch_domain: Add `ip_address_type` argument ([#37237](https://github.com/hashicorp/terraform-provider-aws/issues/37237))
* resource/aws_ec2_traffic_mirror_session: Mark `packet_length` as Computed ([#36962](https://github.com/hashicorp/terraform-provider-aws/issues/36962))
* resource/aws_opensearch_domain: Add `ip_address_type` argument ([#37237](https://github.com/hashicorp/terraform-provider-aws/issues/37237))
* resource/aws_vpc_endpoint: Add `subnet_configuration` argument to support user defined IP addresses ([#37226](https://github.com/hashicorp/terraform-provider-aws/issues/37226))

BUG FIXES:

* data-source/aws_ami: Fix query returning no results ([#37958](https://github.com/hashicorp/terraform-provider-aws/issues/37958))
* provider: Fixes an error where some data sources were not returning `tags` ([#37966](https://github.com/hashicorp/terraform-provider-aws/issues/37966))
* resource/aws_applicationinsights_application: Change `resource_group_name` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#36962](https://github.com/hashicorp/terraform-provider-aws/issues/36962))
* resource/aws_dynamodb_table: Fix `UnknownOperationException: Tagging is not currently supported in DynamoDB Local` errors on resource Read ([#37924](https://github.com/hashicorp/terraform-provider-aws/issues/37924))
* resource/aws_ec2_capacity_reservation: Fix `InvalidCapacityReservationId.NotFound` errors during Read and Delete when resource is manually deleted ([#37127](https://github.com/hashicorp/terraform-provider-aws/issues/37127))
* resource/aws_route53_zone: Fix `InvalidInput: 1 validation error detected: Value '...' at 'resourceId' failed to satisfy constraint: Member must have length less than or equal to 32` errors for resources imported with a `/hostedzone/` prefix ([#37893](https://github.com/hashicorp/terraform-provider-aws/issues/37893))
* service/apigatewayv2: Retry on `ConflictException: Unable to complete operation due to concurrent modification` errors ([#37902](https://github.com/hashicorp/terraform-provider-aws/issues/37902))

## 5.53.0 (June  7, 2024)

FEATURES:

* **New Resource:** `aws_paymentcryptography_key` ([#37017](https://github.com/hashicorp/terraform-provider-aws/issues/37017))
* **New Resource:** `aws_paymentcryptography_key_alias` ([#37020](https://github.com/hashicorp/terraform-provider-aws/issues/37020))

ENHANCEMENTS:

* data-source/aws_customer_gateway: Add `bgp_asn_extended` argument ([#37815](https://github.com/hashicorp/terraform-provider-aws/issues/37815))
* data-source/aws_rds_engine_version: Add `supports_limitless_database` attribute ([#37271](https://github.com/hashicorp/terraform-provider-aws/issues/37271))
* provider: The `use_fips_endpoint` flag is now ignored for any service with a custom endpoint configured in `endpoints`. ([#34233](https://github.com/hashicorp/terraform-provider-aws/issues/34233))
* resource/aws_apigatewayv2_authorizer: Add configurable `delete` timeout ([#37732](https://github.com/hashicorp/terraform-provider-aws/issues/37732))
* resource/aws_customer_gateway: Add `bgp_asn_extended` argument ([#37815](https://github.com/hashicorp/terraform-provider-aws/issues/37815))
* resource/aws_fsx_lustre_file_system: Add `metadata_configuration` argument ([#37868](https://github.com/hashicorp/terraform-provider-aws/issues/37868))
* resource/aws_lb: Add support for IPv6-only Application Load Balancers ([#37700](https://github.com/hashicorp/terraform-provider-aws/issues/37700))
* resource/aws_mwaa_environment: Add `max_webservers` and `min_webservers` attributes ([#37632](https://github.com/hashicorp/terraform-provider-aws/issues/37632))
* resource/aws_pipes_pipe: Add `log_configuration` argument ([#37135](https://github.com/hashicorp/terraform-provider-aws/issues/37135))
* resource/aws_route53_record: Fix `InvalidChangeBatch` errors on resource Delete ([#37850](https://github.com/hashicorp/terraform-provider-aws/issues/37850))
* resource/aws_s3_bucket: Ignore `UnsupportedOperation` errors when reading `acceleration_status`, `server_side_encryption_configuration` and `tags` ([#37801](https://github.com/hashicorp/terraform-provider-aws/issues/37801))
* resource/aws_transfer_ssh_key: Add `ssh_key_id` attribute ([#37548](https://github.com/hashicorp/terraform-provider-aws/issues/37548))

BUG FIXES:

* resource/aws_apigatewayv2_authorizer: Fix `ConflictException` errors on resource Delete ([#37732](https://github.com/hashicorp/terraform-provider-aws/issues/37732))
* resource/aws_bedrockagent_agent: Increase `instruction` max length for validation to 4000 ([#37758](https://github.com/hashicorp/terraform-provider-aws/issues/37758))
* resource/aws_cloudwatch_log_group: Correctly handles tag updates with empty string tags ([#37668](https://github.com/hashicorp/terraform-provider-aws/issues/37668))
* resource/aws_kms_external_key: Fixes timeout error on creation when `ignore_tags` matches tag assigned to resource ([#37818](https://github.com/hashicorp/terraform-provider-aws/issues/37818))
* resource/aws_kms_key: Fixes timeout error on creation when `ignore_tags` matches tag assigned to resource ([#37818](https://github.com/hashicorp/terraform-provider-aws/issues/37818))
* resource/aws_kms_replica_external_key: Fixes timeout error on creation when `ignore_tags` matches tag assigned to resource ([#37818](https://github.com/hashicorp/terraform-provider-aws/issues/37818))
* resource/aws_kms_replica_key: Fixes timeout error on creation when `ignore_tags` matches tag assigned to resource ([#37818](https://github.com/hashicorp/terraform-provider-aws/issues/37818))
* resource/aws_mq_broker: Do not reboot on changes to `maintenance_window_start_time` or `auto_minor_version_upgrade` ([#36506](https://github.com/hashicorp/terraform-provider-aws/issues/36506))
* resource/aws_pipes_pipe: Mark `source_parameters.self_managed_kafka_parameters.credentials.basic_auth` as Optional ([#34293](https://github.com/hashicorp/terraform-provider-aws/issues/34293))
* resource/aws_secretsmanager_secret: Tags with empty values no longer remove all tags. ([#37743](https://github.com/hashicorp/terraform-provider-aws/issues/37743))
* resource/aws_ssm_parameter: Fix `Cannot import non-existent remote object` errors when importing resources with version ([#37832](https://github.com/hashicorp/terraform-provider-aws/issues/37832))
* resource/aws_vpc_endpoint: Restore pre-v5.51.0 default of `false` for `private_dns_enabled` ([#37715](https://github.com/hashicorp/terraform-provider-aws/issues/37715))
* service/chatbot: Correctly overrides region when using custom endpoint. ([#37851](https://github.com/hashicorp/terraform-provider-aws/issues/37851))
* service/costoptimizationhub: Correctly overrides region when using custom endpoint. ([#37851](https://github.com/hashicorp/terraform-provider-aws/issues/37851))
* service/cur: Correctly overrides region when using custom endpoint. ([#37851](https://github.com/hashicorp/terraform-provider-aws/issues/37851))
* service/globalaccelerator: Correctly overrides region when using custom endpoint. ([#37851](https://github.com/hashicorp/terraform-provider-aws/issues/37851))
* service/route53: Correctly overrides region when using custom endpoint. ([#37851](https://github.com/hashicorp/terraform-provider-aws/issues/37851))
* service/route53domains: Correctly overrides region when using custom endpoint. ([#37851](https://github.com/hashicorp/terraform-provider-aws/issues/37851))
* service/shield: Correctly overrides region when using custom endpoint. ([#37851](https://github.com/hashicorp/terraform-provider-aws/issues/37851))

## 5.52.0 (May 30, 2024)

ENHANCEMENTS:

* resource/aws_kinesisanalyticsv2_application: Add `application_mode` argument ([#37714](https://github.com/hashicorp/terraform-provider-aws/issues/37714))
* resource/aws_lightsail_bucket: Add support to `ListTags` function for proper key-only tag handling ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))
* resource/aws_lightsail_certificate: Add support to `ListTags` function for proper key-only tag handling ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))
* resource/aws_lightsail_container_service: Add support to `ListTags` function for proper key-only tag handling ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))
* resource/aws_lightsail_database: Add support to `ListTags` function for proper key-only tag handling ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))
* resource/aws_lightsail_distribution: Add support to `ListTags` function for proper key-only tag handling ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))
* resource/aws_lightsail_key_pair: Add support to `ListTags` function for proper key-only tag handling ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))
* resource/aws_lightsail_lb: Add support to `ListTags` function for proper key-only tag handling ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))

BUG FIXES:

* resource/aws_lightsail_database: Prevent destroy failure when resource is already deleted outside Terraform ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))
* resource/aws_lightsail_instance: Fix crash when reading a resource that has a key-only tag ([#37587](https://github.com/hashicorp/terraform-provider-aws/issues/37587))
* resource/aws_lightsail_key_pair: Prevent destroy failure when resource is already deleted outside Terraform ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))
* resource/aws_lightsail_lb: Prevent destroy failure when resource is already deleted outside Terraform ([#37711](https://github.com/hashicorp/terraform-provider-aws/issues/37711))

## 5.51.1 (May 24, 2024)

ENHANCEMENTS:

* resource/aws_ecs_service: Add `volume_configuration` argument ([#37019](https://github.com/hashicorp/terraform-provider-aws/issues/37019))
* resource/aws_ecs_task_definition: Add `configure_at_launch` parameter in `volume` argument ([#37019](https://github.com/hashicorp/terraform-provider-aws/issues/37019))

BUG FIXES:

* data-source/aws_route53_zone: Fix incorrect `name_servers` values ([#37685](https://github.com/hashicorp/terraform-provider-aws/issues/37685))
* data-source/aws_route53_zone: Permit both `name` and `zone_id` arguments when one is an empty string ([#37686](https://github.com/hashicorp/terraform-provider-aws/issues/37686))
* resource/aws_route53_zone: Fix incorrect `name_servers` values ([#37685](https://github.com/hashicorp/terraform-provider-aws/issues/37685))

## 5.51.0 (May 23, 2024)

NOTES:

* data-source/aws_lambda_function: `source_code_hash` attribute has been deprecated in favor of `code_sha256`. Will be removed in a future major version ([#37669](https://github.com/hashicorp/terraform-provider-aws/issues/37669))
* data-source/aws_lambda_layer_version: `source_code_hash` attribute has been deprecated in favor of `code_sha256`. Will be removed in a future major version ([#37646](https://github.com/hashicorp/terraform-provider-aws/issues/37646))

FEATURES:

* **New Data Source:** `aws_chatbot_slack_workspace` ([#37218](https://github.com/hashicorp/terraform-provider-aws/issues/37218))
* **New Resource:** `aws_lambda_runtime_management_config` ([#37643](https://github.com/hashicorp/terraform-provider-aws/issues/37643))
* **New Resource:** `aws_vpc_endpoint_private_dns` ([#37628](https://github.com/hashicorp/terraform-provider-aws/issues/37628))
* **New Resource:** `aws_vpc_endpoint_service_private_dns_verification` ([#37176](https://github.com/hashicorp/terraform-provider-aws/issues/37176))

ENHANCEMENTS:

* data-source/aws_lambda_function: Add `code_sha256` attribute ([#37669](https://github.com/hashicorp/terraform-provider-aws/issues/37669))
* data-source/aws_lambda_layer_version: Add `code_sha256` attribute ([#37646](https://github.com/hashicorp/terraform-provider-aws/issues/37646))
* data-source/aws_route53_traffic_policy_document: Add support for `application-load-balancer`, `elastic-beanstalk` and `network-load-balancer` `endpoint.type` values ([#37618](https://github.com/hashicorp/terraform-provider-aws/issues/37618))
* resource/aws_api_gateway_deployment: Add `canary_settings` attribute ([#37573](https://github.com/hashicorp/terraform-provider-aws/issues/37573))
* resource/aws_iam_openid_connect_provider: Allow `client_id_list` to be updated in-place ([#37612](https://github.com/hashicorp/terraform-provider-aws/issues/37612))
* resource/aws_lambda_function: Add `code_sha256` attribute ([#37669](https://github.com/hashicorp/terraform-provider-aws/issues/37669))
* resource/aws_lambda_function: Remove `replace_security_group_on_destroy` and `replacement_security_group_ids` deprecations, re-implement with alternate workflow ([#37624](https://github.com/hashicorp/terraform-provider-aws/issues/37624))
* resource/aws_lambda_layer_version: Add `code_sha256` attribute ([#37646](https://github.com/hashicorp/terraform-provider-aws/issues/37646))
* resource/aws_route53_health_check: Add plan-time validation of `cloudwatch_alarm_region` ([#37510](https://github.com/hashicorp/terraform-provider-aws/issues/37510))
* resource/aws_route53_record: Add plan-time validation of `latency_routing_policy.region` ([#37510](https://github.com/hashicorp/terraform-provider-aws/issues/37510))
* resource/aws_route53_vpc_association_authorization: Add plan-time validation of `vpc_region` ([#37510](https://github.com/hashicorp/terraform-provider-aws/issues/37510))
* resource/aws_route53_zone_association: Add plan-time validation of `vpc_region` ([#37510](https://github.com/hashicorp/terraform-provider-aws/issues/37510))
* resource/aws_wafv2_web_acl: Add `api_gateway`, `app_runner_service`, `cognito_user_pool`, and `verified_access_instance` configuration blocks to `association_config.request_body` ([#37588](https://github.com/hashicorp/terraform-provider-aws/issues/37588))

BUG FIXES:

* resource/aws_dynamodb_table_replica: Correctly set `kms_key_arn` on Read ([#37570](https://github.com/hashicorp/terraform-provider-aws/issues/37570))
* resource/aws_kms_grant: Change `grant_token` to [`Sensitive`](https://developer.hashicorp.com/terraform/plugin/best-practices/sensitive-state#using-sensitive-flag-functionality) ([#37593](https://github.com/hashicorp/terraform-provider-aws/issues/37593))
* resource/aws_lambda_function: Fix issue when `source_code_hash` causes drift even if source code has not changed ([#37669](https://github.com/hashicorp/terraform-provider-aws/issues/37669))
* resource/aws_lambda_layer_version: Fix issue when `source_code_hash` forces a replacement even if source code has not changed ([#37646](https://github.com/hashicorp/terraform-provider-aws/issues/37646))
* resource/aws_m2_deployment: Fix `state` error on `deployment_id` during start/stop update ([#37581](https://github.com/hashicorp/terraform-provider-aws/issues/37581))
* resource/aws_storagegateway_smb_file_share: Fix crash when `cache_attributes` is removed on update ([#37611](https://github.com/hashicorp/terraform-provider-aws/issues/37611))

## 5.50.0 (May 17, 2024)

ENHANCEMENTS:

* data-source/aws_budgets_budget: Add `tags` attribute ([#37361](https://github.com/hashicorp/terraform-provider-aws/issues/37361))
* data-source/aws_instance: Add `launch_time` attribute ([#37002](https://github.com/hashicorp/terraform-provider-aws/issues/37002))
* resource/aws_budgets_budget: Add `tags` argument ([#37361](https://github.com/hashicorp/terraform-provider-aws/issues/37361))
* resource/aws_budgets_budget_action: Add `tags` argument ([#37361](https://github.com/hashicorp/terraform-provider-aws/issues/37361))
* resource/aws_ecs_account_setting_default: Add support for `fargateTaskRetirementWaitPeriod` value in `Name` argument ([#37018](https://github.com/hashicorp/terraform-provider-aws/issues/37018))
* resource/aws_ssm_resource_data_sync: Add plan-time validation of `s3_destination.kms_key_arn`, `s3_destination.region` and `s3_destination.sync_format` ([#37481](https://github.com/hashicorp/terraform-provider-aws/issues/37481))

BUG FIXES:

* data-source/aws_bedrock_foundation_models: Fix validation regex for the `by_provider` argument ([#37306](https://github.com/hashicorp/terraform-provider-aws/issues/37306))
* resource/aws_dynamodb_table: Fix `UnknownOperationException: Tagging is not currently supported in DynamoDB Local` errors on resource Read ([#37472](https://github.com/hashicorp/terraform-provider-aws/issues/37472))
* resource/aws_glue_job: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panic when `notify_delay_after` is empty (`null`) ([#37347](https://github.com/hashicorp/terraform-provider-aws/issues/37347))
* resource/aws_iam_server_certificate: Now correctly reads tags after update and on read. ([#37483](https://github.com/hashicorp/terraform-provider-aws/issues/37483))
* resource/aws_lakeformation_data_cells_filter: Fix inconsistent `state` error when using `row_filter.all_rows_wildcard` ([#37433](https://github.com/hashicorp/terraform-provider-aws/issues/37433))
* resource/aws_organizations_account: Allow import of accounts with IAM access to the AWS Billing and Cost Management console ([#35662](https://github.com/hashicorp/terraform-provider-aws/issues/35662))
* resource/aws_ram_principal_association: Correct plan-time validation of `principal` to fix `panic: unexpected format for ID parts ([...]), the following id parts indexes are blank ([1])` ([#37450](https://github.com/hashicorp/terraform-provider-aws/issues/37450))
* resource/aws_route53_record: Change region default to us-east-1 ([#37565](https://github.com/hashicorp/terraform-provider-aws/issues/37565))
* resource/aws_vpc_endpoint_service: Fix destroy error when endpoint service is deleted out-of-band ([#37534](https://github.com/hashicorp/terraform-provider-aws/issues/37534))

## 5.49.0 (May 10, 2024)

FEATURES:

* **New Data Source:** `aws_datazone_environment_blueprint` ([#36600](https://github.com/hashicorp/terraform-provider-aws/issues/36600))
* **New Resource:** `aws_bedrockagent_data_source` ([#37158](https://github.com/hashicorp/terraform-provider-aws/issues/37158))
* **New Resource:** `aws_datazone_domain` ([#36600](https://github.com/hashicorp/terraform-provider-aws/issues/36600))
* **New Resource:** `aws_datazone_environment_blueprint_configuration` ([#36600](https://github.com/hashicorp/terraform-provider-aws/issues/36600))

ENHANCEMENTS:

* data-source/aws_iam_policy_document: Add `minified_json` attribute ([#35677](https://github.com/hashicorp/terraform-provider-aws/issues/35677))
* resource/aws_dynamodb_table_export: Add plan-time validation of `table_arn` ([#37288](https://github.com/hashicorp/terraform-provider-aws/issues/37288))
* resource/aws_kms_key: Add `rotation_period_in_days` argument ([#37140](https://github.com/hashicorp/terraform-provider-aws/issues/37140))
* resource/aws_securitylake_subscriber_notification: Better handles importing resource ([#37332](https://github.com/hashicorp/terraform-provider-aws/issues/37332))
* resource/aws_securitylake_subscriber_notification: Deprecates `endpoint_id` in favor of `subscriber_endpoint` ([#37332](https://github.com/hashicorp/terraform-provider-aws/issues/37332))
* resource/aws_securitylake_subscriber_notification: Handles `configuration.https_notification_configuration.authorization_api_key_value` as sensitive value ([#37332](https://github.com/hashicorp/terraform-provider-aws/issues/37332))

BUG FIXES:

* data-source/aws_fsx_ontap_storage_virtual_machine: Correctly set `tags` on Read ([#37353](https://github.com/hashicorp/terraform-provider-aws/issues/37353))
* data-source/aws_rds_orderable_db_instance: Fix `InvalidParameterValue: Invalid value 3412 for MaxRecords. Must be between 20 and 1000` errors ([#37251](https://github.com/hashicorp/terraform-provider-aws/issues/37251))
* data-source/aws_resourceexplorer2_search: Fix 401 unauthorized error due to missing `view_arn` in the AWS API request ([#36778](https://github.com/hashicorp/terraform-provider-aws/issues/36778))
* data-source/aws_resourceexplorer2_search: Fix panic caused by bad mappping between Terraform and AWS schemas ([#36778](https://github.com/hashicorp/terraform-provider-aws/issues/36778))
* data-source/aws_resourceexplorer2_search: Fix state persistence and data types ([#36778](https://github.com/hashicorp/terraform-provider-aws/issues/36778))
* resource/aws_bedrockagent_agent: Fix to use the configured `prepare_agent` value (or default value of `true` when omitted) for all create and update operations ([#37405](https://github.com/hashicorp/terraform-provider-aws/issues/37405))
* resource/aws_elasticsearch_domain: Fix handling of unset `auto_tune_options.rollback_on_disable` argument ([#37394](https://github.com/hashicorp/terraform-provider-aws/issues/37394))
* resource/aws_fsx_ontap_storage_virtual_machine: Correctly set `tags` and `tags_all` on resource Read ([#37353](https://github.com/hashicorp/terraform-provider-aws/issues/37353))
* resource/aws_fsx_openzfs_file_system: Correctly set `tags` and `tags_all` on resource Read ([#37353](https://github.com/hashicorp/terraform-provider-aws/issues/37353))
* resource/aws_kms_custom_key_store: Change `trust_anchor_certificate` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#37092](https://github.com/hashicorp/terraform-provider-aws/issues/37092))
* resource/aws_opensearch_domain: Fix handling of unset `auto_tune_options.rollback_on_disable` argument ([#37394](https://github.com/hashicorp/terraform-provider-aws/issues/37394))
* resource/aws_opensearch_domain: Wait for `auto_tune_options` to be applied during creation ([#37394](https://github.com/hashicorp/terraform-provider-aws/issues/37394))
* resource/aws_securitylake_aws_log_source: Correctly handles unspecified `source_version` ([#36268](https://github.com/hashicorp/terraform-provider-aws/issues/36268))
* resource/aws_securitylake_aws_log_source: Prevents errors when creating multiple log sources concurrently ([#36268](https://github.com/hashicorp/terraform-provider-aws/issues/36268))
* resource/aws_securitylake_custom_log_source: Prevents errors when creating multiple log sources concurrently ([#36268](https://github.com/hashicorp/terraform-provider-aws/issues/36268))
* resource/aws_securitylake_custom_log_source: Validates length of `source_name` parameter ([#36268](https://github.com/hashicorp/terraform-provider-aws/issues/36268))
* resource/aws_securitylake_subscriber: Allow more than one log source ([#36268](https://github.com/hashicorp/terraform-provider-aws/issues/36268))
* resource/aws_securitylake_subscriber: Correctly handles unspecified `access_type` ([#36268](https://github.com/hashicorp/terraform-provider-aws/issues/36268))
* resource/aws_securitylake_subscriber: Correctly handles unspecified `source_version` parameter for `aws_log_source_resource` and `custom_log_source_resource` ([#36268](https://github.com/hashicorp/terraform-provider-aws/issues/36268))
* resource/aws_securitylake_subscriber: Correctly requires `source_name` parameter for `aws_log_source_resource` and `custom_log_source_resource` ([#36268](https://github.com/hashicorp/terraform-provider-aws/issues/36268))
* resource/aws_securitylake_subscriber_notification: No longer recreates resource when not needed ([#37332](https://github.com/hashicorp/terraform-provider-aws/issues/37332))
* resource/aws_securitylake_subscriber_notification: Requires value for `configuration.https_notification_configuration.endpoint` ([#37332](https://github.com/hashicorp/terraform-provider-aws/issues/37332))
* resource/provider: Change the AWS SDK for Go v2 API client [`BackoffDelayer`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2@v1.26.1/aws/retry#BackoffDelayer) to maintain behavioral compatibility with AWS SDK for Go v1 ([#37404](https://github.com/hashicorp/terraform-provider-aws/issues/37404))

## 5.48.0 (May  2, 2024)

FEATURES:

* **New Resource:** `aws_bedrockagent_agent_knowledge_base_association` ([#37185](https://github.com/hashicorp/terraform-provider-aws/issues/37185))

ENHANCEMENTS:

* resource/aws_cloudwatch_event_target: Add `force_destroy` argument ([#37130](https://github.com/hashicorp/terraform-provider-aws/issues/37130))
* resource/aws_elasticache_replication_group: Increase default Delete timeout to 45 minutes ([#37182](https://github.com/hashicorp/terraform-provider-aws/issues/37182))
* resource/aws_elasticache_replication_group: Use the configured Delete timeout when detaching from any global replication group ([#37182](https://github.com/hashicorp/terraform-provider-aws/issues/37182))
* resource/aws_fsx_ontap_file_system: Add support for specifying 1 ha_pair with `SINGLE_AZ_1` and `MULTI_AZ_1` deployment types ([#36511](https://github.com/hashicorp/terraform-provider-aws/issues/36511))
* resource/aws_fsx_ontap_file_system: Increase `storage_capacity` maximum to 1PiB ([#36511](https://github.com/hashicorp/terraform-provider-aws/issues/36511))
* resource/aws_fsx_ontap_file_system: Support up to 12 `ha_pairs` ([#36511](https://github.com/hashicorp/terraform-provider-aws/issues/36511))
* resource/aws_fsx_ontap_file_system: Update `throughput_capacity_per_ha_pair` to support all values from `throughput_capacity` ([#36511](https://github.com/hashicorp/terraform-provider-aws/issues/36511))
* resource/aws_fsx_ontap_volume: Add `aggregate_configuration` configuration block ([#36511](https://github.com/hashicorp/terraform-provider-aws/issues/36511))
* resource/aws_fsx_ontap_volume: Add `size_in_bytes` and `volume_style` arguments ([#36511](https://github.com/hashicorp/terraform-provider-aws/issues/36511))

BUG FIXES:

* resource/aws_bcmdataexports_export: Fix `table_configurations` expand/flatten ([#37205](https://github.com/hashicorp/terraform-provider-aws/issues/37205))
* resource/aws_cloudwatch_event_connection: Add plan-time validation preventing empty `auth_parameters.oauth.oauth_http_parameters` or `auth_parameters.invocation_http_parameters`
`body`, `header` and `query_string` configuration blocks ([#26755](https://github.com/hashicorp/terraform-provider-aws/issues/26755))
* resource/aws_elasticache_replication_group: Decrease replica count after other updates ([#34819](https://github.com/hashicorp/terraform-provider-aws/issues/34819))
* resource/aws_elasticache_replication_group: Fix `unexpected state 'snapshotting'` errors when increasing or decreasing replica count ([#30493](https://github.com/hashicorp/terraform-provider-aws/issues/30493))

## 5.47.0 (April 26, 2024)

NOTES:

* provider: Updates to Go 1.22. This is the last Go release that will run on macOS 10.15 Catalina ([#36996](https://github.com/hashicorp/terraform-provider-aws/issues/36996))
* resource/aws_bedrockagent_knowledge_base: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#36783](https://github.com/hashicorp/terraform-provider-aws/issues/36783))

FEATURES:

* **New Data Source:** `aws_identitystore_groups` ([#36993](https://github.com/hashicorp/terraform-provider-aws/issues/36993))
* **New Resource:** `aws_bcmdataexports_export` ([#36847](https://github.com/hashicorp/terraform-provider-aws/issues/36847))
* **New Resource:** `aws_bedrockagent_agent` ([#36851](https://github.com/hashicorp/terraform-provider-aws/issues/36851))
* **New Resource:** `aws_bedrockagent_agent_action_group` ([#36935](https://github.com/hashicorp/terraform-provider-aws/issues/36935))
* **New Resource:** `aws_bedrockagent_agent_alias` ([#36905](https://github.com/hashicorp/terraform-provider-aws/issues/36905))
* **New Resource:** `aws_bedrockagent_knowledge_base` ([#36783](https://github.com/hashicorp/terraform-provider-aws/issues/36783))
* **New Resource:** `aws_globalaccelerator_cross_account_attachment` ([#35991](https://github.com/hashicorp/terraform-provider-aws/issues/35991))
* **New Resource:** `aws_verifiedpermissions_policy` ([#35413](https://github.com/hashicorp/terraform-provider-aws/issues/35413))

ENHANCEMENTS:

* data-source/aws_eip: Add `arn` attribute ([#35991](https://github.com/hashicorp/terraform-provider-aws/issues/35991))
* resource/aws_api_gateway_rest_api: Correctly set `root_resource_id` on resource Read ([#37040](https://github.com/hashicorp/terraform-provider-aws/issues/37040))
* resource/aws_appmesh_mesh: Add `spec.service_discovery` argument ([#37042](https://github.com/hashicorp/terraform-provider-aws/issues/37042))
* resource/aws_cloudformation_stack_set: Adds guidance on permissions when using delegated administrator account ([#37069](https://github.com/hashicorp/terraform-provider-aws/issues/37069))
* resource/aws_db_instance: Add `dedicated_log_volume` argument ([#36503](https://github.com/hashicorp/terraform-provider-aws/issues/36503))
* resource/aws_eip: Add `arn` attribute ([#35991](https://github.com/hashicorp/terraform-provider-aws/issues/35991))
* resource/aws_elasticache_replication_group: Add `transit_encryption_mode` argument ([#30403](https://github.com/hashicorp/terraform-provider-aws/issues/30403))
* resource/aws_elasticache_replication_group: Changes to the `transit_encryption_enabled` argument can now be done in-place for engine versions > `7.0.5` ([#30403](https://github.com/hashicorp/terraform-provider-aws/issues/30403))
* resource/aws_kinesis_firehose_delivery_stream: Add `snowflake_configuration` argument ([#36646](https://github.com/hashicorp/terraform-provider-aws/issues/36646))
* resource/aws_memorydb_user: Support IAM authentication mode ([#32027](https://github.com/hashicorp/terraform-provider-aws/issues/32027))
* resource/aws_sagemaker_app_image_config: Add `code_editor_app_image_config` and `jupyter_lab_image_config.jupyter_lab_image_config` arguments ([#37059](https://github.com/hashicorp/terraform-provider-aws/issues/37059))
* resource/aws_sagemaker_app_image_config: Change `kernel_gateway_image_config.kernel_spec` MaxItems to 5 ([#37059](https://github.com/hashicorp/terraform-provider-aws/issues/37059))
* resource/aws_transfer_server: Add `sftp_authentication_methods` argument ([#37015](https://github.com/hashicorp/terraform-provider-aws/issues/37015))

BUG FIXES:

* resource/aws_batch_job_definition: Fix issues where changes causing a new `revision` do not trigger changes in dependent resources and/or cause an error, "Provider produced inconsistent final plan" ([#37111](https://github.com/hashicorp/terraform-provider-aws/issues/37111))
* resource/aws_ce_cost_category: Allow up to 3 levels of `and`, `not` and `or` operand nesting for the `rule` argument ([#30862](https://github.com/hashicorp/terraform-provider-aws/issues/30862))
* resource/aws_elasticache_replication_group: Fix excessive delay on read ([#30403](https://github.com/hashicorp/terraform-provider-aws/issues/30403))
* resource/aws_servicecatalog_portfolio: Fixes error where deletion fails if resource was deleted out of band. ([#37066](https://github.com/hashicorp/terraform-provider-aws/issues/37066))
* resource/aws_servicecatalog_provisioned_product: Fixes error where tag values are not applied to products when tag values don't change. ([#37066](https://github.com/hashicorp/terraform-provider-aws/issues/37066))

## 5.46.0 (April 18, 2024)

NOTES:

* provider: When using YAML or JSON documents, such as in `template_body` of `aws_cloudformation_stack`, CRLF was previously treated as different from LF but these are now treated as equivalent in many situations ([#14270](https://github.com/hashicorp/terraform-provider-aws/issues/14270))

FEATURES:

* **New Resource:** `aws_eip_domain_name` ([#36963](https://github.com/hashicorp/terraform-provider-aws/issues/36963))

ENHANCEMENTS:

* data-source/aws_alb: Add `client_keep_alive` argument ([#36969](https://github.com/hashicorp/terraform-provider-aws/issues/36969))
* data-source/aws_eip: Add `ptr_record` attribute ([#36963](https://github.com/hashicorp/terraform-provider-aws/issues/36963))
* data-source/aws_iam_policy: Add `attachment_count` attribute ([#36759](https://github.com/hashicorp/terraform-provider-aws/issues/36759))
* data-source/aws_lb: Add `client_keep_alive` argument ([#36969](https://github.com/hashicorp/terraform-provider-aws/issues/36969))
* data-source/aws_organizations_organization: Add `master_account_name` attribute ([#36797](https://github.com/hashicorp/terraform-provider-aws/issues/36797))
* data-source/aws_vpc_dhcp_options: Add `ipv6_address_preferred_lease_time` attribute ([#36934](https://github.com/hashicorp/terraform-provider-aws/issues/36934))
* resource/aws_alb: Add `client_keep_alive` argument ([#36969](https://github.com/hashicorp/terraform-provider-aws/issues/36969))
* resource/aws_autoscaling_group: Add `alarm_specification` to the `instance_refresh.preferences` configuration block ([#36954](https://github.com/hashicorp/terraform-provider-aws/issues/36954))
* resource/aws_cloudformation_stack_set: Add retry when creating to potentially help with eventual consistency problems ([#36982](https://github.com/hashicorp/terraform-provider-aws/issues/36982))
* resource/aws_cloudfront_origin_access_control: Add `lambda` and `mediapackagev2` as valid values for `origin_access_control_origin_type` ([#34362](https://github.com/hashicorp/terraform-provider-aws/issues/34362))
* resource/aws_cloudwatch_event_rule: Add `force_destroy` attribute ([#34905](https://github.com/hashicorp/terraform-provider-aws/issues/34905))
* resource/aws_codebuild_project: Add GitLab and GitLab Self Managed support to the `report_build_status` and `build_status_config` arguments ([#36942](https://github.com/hashicorp/terraform-provider-aws/issues/36942))
* resource/aws_default_vpc_dhcp_options: Add `ipv6_address_preferred_lease_time` as Computed attribute ([#36934](https://github.com/hashicorp/terraform-provider-aws/issues/36934))
* resource/aws_dms_replication_task: Add `resource_identifier` argument ([#36901](https://github.com/hashicorp/terraform-provider-aws/issues/36901))
* resource/aws_eip: Add `ptr_record` attribute ([#36963](https://github.com/hashicorp/terraform-provider-aws/issues/36963))
* resource/aws_elasticache_serverless_cache: Add `minimum` attribute in `cache_usage_limits.data_storage` and `cache_usage_limits.ecpu_per_second` ([#36766](https://github.com/hashicorp/terraform-provider-aws/issues/36766))
* resource/aws_fsx_openzfs_file_system: Add `endpoint_ip_address` attribute ([#36767](https://github.com/hashicorp/terraform-provider-aws/issues/36767))
* resource/aws_iam_policy: Add `attachment_count` attribute ([#36759](https://github.com/hashicorp/terraform-provider-aws/issues/36759))
* resource/aws_imagebuilder_image: Add `execution_role` and `workflow` arguments ([#36953](https://github.com/hashicorp/terraform-provider-aws/issues/36953))
* resource/aws_lb: Add `client_keep_alive` argument ([#36969](https://github.com/hashicorp/terraform-provider-aws/issues/36969))
* resource/aws_mwaa_environment: Add `database_vpc_endpoint_service` and `webserver_vpc_endpoint_service` attributes ([#36903](https://github.com/hashicorp/terraform-provider-aws/issues/36903))
* resource/aws_organizations_organization: Add `master_account_name` attribute ([#36797](https://github.com/hashicorp/terraform-provider-aws/issues/36797))
* resource/aws_transfer_connector: Add `security_policy_name` argument ([#36893](https://github.com/hashicorp/terraform-provider-aws/issues/36893))
* resource/aws_vpc_dhcp_options: Add `ipv6_address_preferred_lease_time` attribute ([#36934](https://github.com/hashicorp/terraform-provider-aws/issues/36934))
* resource/aws_vpc_ipam_pool: Add `cascade` argument ([#36898](https://github.com/hashicorp/terraform-provider-aws/issues/36898))

BUG FIXES:

* data-source/aws_iam_policy_document: When using multiple principals, sort them to avoid differences based only on order ([#25967](https://github.com/hashicorp/terraform-provider-aws/issues/25967))
* resource/aws_appconfig_deployment: Fix `ConflictException` errors on resource Create ([#36980](https://github.com/hashicorp/terraform-provider-aws/issues/36980))
* resource/aws_ce_anomaly_monitor: Change `monitor_dimension` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#36773](https://github.com/hashicorp/terraform-provider-aws/issues/36773))
* resource/aws_ce_anomaly_subscription: Change `account_id` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#36773](https://github.com/hashicorp/terraform-provider-aws/issues/36773))
* resource/aws_cloudformation_stack: CRLF line endings in `template_body` no longer cause erroneous diffs ([#14270](https://github.com/hashicorp/terraform-provider-aws/issues/14270))
* resource/aws_db_proxy: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panic when `auth` is empty (`{}`) ([#36967](https://github.com/hashicorp/terraform-provider-aws/issues/36967))
* resource/aws_dms_replication_config: Adds validation to `replication_settings` to disallow `Logging.CloudWatchLogGroup` and `Logging.CloudWatchLogStream`. ([#36936](https://github.com/hashicorp/terraform-provider-aws/issues/36936))
* resource/aws_dms_replication_config: Suppresses differences in partial `replication_settings` JSON documents. ([#36936](https://github.com/hashicorp/terraform-provider-aws/issues/36936))
* resource/aws_dms_replication_task: Adds validation to `replication_task_settings` to disallow `Logging.CloudWatchLogGroup` and `Logging.CloudWatchLogStream`. ([#36936](https://github.com/hashicorp/terraform-provider-aws/issues/36936))
* resource/aws_dms_replication_task: Allows leaving `replication_task_settings` unset to use default settings. ([#36936](https://github.com/hashicorp/terraform-provider-aws/issues/36936))
* resource/aws_dms_replication_task: Suppresses differences in partial `replication_task_settings` JSON documents. ([#36936](https://github.com/hashicorp/terraform-provider-aws/issues/36936))
* resource/aws_fsx_windows_file_system: Fix error `BadRequest: AuditLogDestination must not be provided when auditing is disabled` when updating `audit_log_configuration.0.file_access_audit_log_level` and `audit_log_configuration.0.file_share_access_audit_log_level` to `"DISABLED"` ([#36928](https://github.com/hashicorp/terraform-provider-aws/issues/36928))
* resource/aws_glue_job: Mark `number_of_workers` and `worker_type` as optional/computed, preventing persistent differences when `max_capacity` is set. ([#36770](https://github.com/hashicorp/terraform-provider-aws/issues/36770))
* resource/aws_iam_user_login_profile: Fix forced re-creation when `password_reset_required` is `true` and initial password reset is completed ([#36926](https://github.com/hashicorp/terraform-provider-aws/issues/36926))
* resource/aws_lightsail_distribution: Fix to properly set `certificate_name` on create and update ([#36888](https://github.com/hashicorp/terraform-provider-aws/issues/36888))
* resource/aws_vpc_dhcp_options: Fix `NotFound` error handling on delete ([#36933](https://github.com/hashicorp/terraform-provider-aws/issues/36933))

## 5.45.0 (April 11, 2024)

NOTES:

* resource/aws_redshift_cluster: The `logging` argument is now deprecated. Use the `aws_redshift_logging` resource instead. ([#36862](https://github.com/hashicorp/terraform-provider-aws/issues/36862))
* resource/aws_redshift_cluster: The `snapshot_copy` argument is now deprecated. Use the `aws_redshift_snapshot_copy` resource instead. ([#36810](https://github.com/hashicorp/terraform-provider-aws/issues/36810))

FEATURES:

* **New Resource:** `aws_redshift_logging` ([#36862](https://github.com/hashicorp/terraform-provider-aws/issues/36862))
* **New Resource:** `aws_redshift_snapshot_copy` ([#36810](https://github.com/hashicorp/terraform-provider-aws/issues/36810))

ENHANCEMENTS:

* data-source/aws_sagemaker_prebuilt_ecr_image: Add `registry_id` for `af-south-1` AWS Region ([#36803](https://github.com/hashicorp/terraform-provider-aws/issues/36803))
* resource/aws_api_gateway_documentation_part: Add `documentation_part_id` attribute ([#36445](https://github.com/hashicorp/terraform-provider-aws/issues/36445))
* resource/aws_wafregional_web_acl_association: Add configurable timeouts ([#36445](https://github.com/hashicorp/terraform-provider-aws/issues/36445))
* resource/aws_wafregional_web_acl_association: Add plan-time validation of `resource_arn` ([#36445](https://github.com/hashicorp/terraform-provider-aws/issues/36445))

BUG FIXES:

* provider: Change the default AWS SDK for Go v2 API client [`MaxBackoff`](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/retries-timeouts/#limiting-the-max-back-off-delay) value to 300 seconds so that services migrated to AWS SDK for Go v2 maintain behavioral compatibility with AWS SDK for Go v1 ([#36855](https://github.com/hashicorp/terraform-provider-aws/issues/36855))
* resource/aws_datasync_location_object_storage: Allow update to `agent_arns` ([#36819](https://github.com/hashicorp/terraform-provider-aws/issues/36819))
* resource/aws_devopsguru_notification_channel: Fix persistent diff when `filters.message_types` or `filters.severities` contains multiple elements ([#36804](https://github.com/hashicorp/terraform-provider-aws/issues/36804))
* resource/aws_securityhub_configuration_policy: Mark `configuration_policy.enabled_standard_arns` as Optional, fixing `InvalidInputException: Invalid semantics: Enabled standards and security control configurations must be configured when Security Hub is enabled` errors ([#36740](https://github.com/hashicorp/terraform-provider-aws/issues/36740))

## 5.44.0 (April  4, 2024)

FEATURES:

* **New Data Source:** `aws_devopsguru_notification_channel` ([#36656](https://github.com/hashicorp/terraform-provider-aws/issues/36656))
* **New Data Source:** `aws_devopsguru_resource_collection` ([#36657](https://github.com/hashicorp/terraform-provider-aws/issues/36657))
* **New Data Source:** `aws_ecr_lifecycle_policy_document` ([#6133](https://github.com/hashicorp/terraform-provider-aws/issues/6133))
* **New Function:** `trim_iam_role_path` ([#36723](https://github.com/hashicorp/terraform-provider-aws/issues/36723))
* **New Resource:** `aws_devopsguru_service_integration` ([#36694](https://github.com/hashicorp/terraform-provider-aws/issues/36694))

ENHANCEMENTS:

* data-source/aws_servicecatalogappregistry_application: Add `application_tag` attribute ([#36647](https://github.com/hashicorp/terraform-provider-aws/issues/36647))
* data/aws_glue_data_catalog_encryption_settings: Add `data_catalog_encryption_settings.encryption_at_rest.catalog_encryption_service_role` attribute ([#35978](https://github.com/hashicorp/terraform-provider-aws/issues/35978))
* resource/aws_appstream_fleet: Add `desired_sessions` argument to the `compute_capacity` block. ([#34266](https://github.com/hashicorp/terraform-provider-aws/issues/34266))
* resource/aws_appstream_fleet: Add `max_sessions_per_instance` argument. ([#34266](https://github.com/hashicorp/terraform-provider-aws/issues/34266))
* resource/aws_batch_job_definition: Add update functions instead of ForceNew. Add `deregister_on_new_revision` to allow keeping prior versions ACTIVE when a new revision is published. ([#35149](https://github.com/hashicorp/terraform-provider-aws/issues/35149))
* resource/aws_db_instance: Adds warning when setting `character_set_name` when `replicate_source_db`, `restore_to_point_in_time`, or `snapshot_identifier` is set ([#36518](https://github.com/hashicorp/terraform-provider-aws/issues/36518))
* resource/aws_emr_cluster: Add `unhealthy_node_replacement` argument ([#36523](https://github.com/hashicorp/terraform-provider-aws/issues/36523))
* resource/aws_glue_data_catalog_encryption_settings: Add `data_catalog_encryption_settings.encryption_at_rest.catalog_encryption_service_role` argument ([#35978](https://github.com/hashicorp/terraform-provider-aws/issues/35978))
* resource/aws_lambda_function: Add support for `ruby3.3` `runtime` value ([#36751](https://github.com/hashicorp/terraform-provider-aws/issues/36751))
* resource/aws_lambda_layer_version: Add support for `ruby3.3` `compatible_runtimes` value ([#36751](https://github.com/hashicorp/terraform-provider-aws/issues/36751))
* resource/aws_servicecatalogappregistry_application: Add `application_tag` attribute ([#36647](https://github.com/hashicorp/terraform-provider-aws/issues/36647))
* resource/aws_transfer_server: Add `s3_storage_options` configuration block ([#36664](https://github.com/hashicorp/terraform-provider-aws/issues/36664))
* resource/aws_wafv2_web_acl: Add `address_fields` and `phone_number_fields` to `statement.managed_rule_group_statement.managed_rule_group_configs.aws_managed_rules_acfp_rule_set.request_inspection` ([#36685](https://github.com/hashicorp/terraform-provider-aws/issues/36685))

BUG FIXES:

* provider: Correctly handles user agents passed using `TF_APPEND_USER_AGENT` which contain `/`,  `(`, `)`, or space. ([#36738](https://github.com/hashicorp/terraform-provider-aws/issues/36738))
* resource/aws_batch_scheduling_policy: Fixes error where tags could not be updated ([#36517](https://github.com/hashicorp/terraform-provider-aws/issues/36517))
* resource/aws_cloudfront_key_value_store: Serialize CloudFront KeyValueStore access ([#36734](https://github.com/hashicorp/terraform-provider-aws/issues/36734))
* resource/aws_cloudfrontkeyvaluestore_key: Serialize CloudFront KeyValueStore access ([#36734](https://github.com/hashicorp/terraform-provider-aws/issues/36734))
* resource/aws_cognito_user_pool: Correct plan-time validation of `email_verification_message`, `email_verification_subject`, `admin_create_user_config.invite_message_template.email_message`, `admin_create_user_config.invite_message_template.email_subject`, `admin_create_user_config.invite_message_template.sms_message`, `sms_authentication_message`, `sms_verification_message`, `verification_message_template.email_message`, `verification_message_template.email_message_by_link`, `verification_message_template.email_subject`, `verification_message_template.email_subject_by_link`, and `verification_message_template.sms_message` to count UTF-8 characters properly ([#36661](https://github.com/hashicorp/terraform-provider-aws/issues/36661))
* resource/aws_ecr_lifecycle_policy: Add missing `tagPatternList` change detection in policy JSON ([#35231](https://github.com/hashicorp/terraform-provider-aws/issues/35231))
* resource/aws_ecs_service: Correctly set `alarms.rollback` on resource Create and Update ([#36691](https://github.com/hashicorp/terraform-provider-aws/issues/36691))
* resource/aws_iam_user: When `force_destroy` is used and there are inline or attached policies, allow resource to be destroyed ([#36640](https://github.com/hashicorp/terraform-provider-aws/issues/36640))
* resource/aws_imagebuilder_distribution_configuration: Fix validation regex for `ami_distribution_configuration.name` ([#36659](https://github.com/hashicorp/terraform-provider-aws/issues/36659))
* resource/aws_redshift_cluster: Fix error preventing modification of a configured `snapshot_copy` block ([#36655](https://github.com/hashicorp/terraform-provider-aws/issues/36655))
* resource/aws_route53_record: Fix to correctly interpret alias names with wildcards ([#36699](https://github.com/hashicorp/terraform-provider-aws/issues/36699))

## 5.43.0 (March 28, 2024)

FEATURES:

* **New Data Source:** `aws_resourceexplorer2_search` ([#36560](https://github.com/hashicorp/terraform-provider-aws/issues/36560))
* **New Data Source:** `aws_servicecatalogappregistry_application` ([#36596](https://github.com/hashicorp/terraform-provider-aws/issues/36596))
* **New Resource:** `aws_cloudfrontkeyvaluestore_key` ([#36534](https://github.com/hashicorp/terraform-provider-aws/issues/36534))
* **New Resource:** `aws_devopsguru_notification_channel` ([#36557](https://github.com/hashicorp/terraform-provider-aws/issues/36557))
* **New Resource:** `aws_dynamodb_resource_policy` ([#36595](https://github.com/hashicorp/terraform-provider-aws/issues/36595))
* **New Resource:** `aws_ec2_instance_metadata_defaults` ([#36589](https://github.com/hashicorp/terraform-provider-aws/issues/36589))
* **New Resource:** `aws_lakeformation_resource_lf_tag` ([#36537](https://github.com/hashicorp/terraform-provider-aws/issues/36537))
* **New Resource:** `aws_m2_application` ([#35399](https://github.com/hashicorp/terraform-provider-aws/issues/35399))
* **New Resource:** `aws_m2_deployment` ([#35408](https://github.com/hashicorp/terraform-provider-aws/issues/35408))
* **New Resource:** `aws_m2_environment` ([#35311](https://github.com/hashicorp/terraform-provider-aws/issues/35311))
* **New Resource:** `aws_redshiftserverless_custom_domain_association` ([#35865](https://github.com/hashicorp/terraform-provider-aws/issues/35865))
* **New Resource:** `aws_servicecatalogappregistry_application` ([#36277](https://github.com/hashicorp/terraform-provider-aws/issues/36277))

ENHANCEMENTS:

* data-source/aws_cloudfront_function: Add `key_value_store_associations` attribute ([#36585](https://github.com/hashicorp/terraform-provider-aws/issues/36585))
* data-source/aws_db_snapshot: Add `original_snapshot_create_time` attribute ([#36544](https://github.com/hashicorp/terraform-provider-aws/issues/36544))
* resource/aws_cloudfront_function: Add `key_value_store_associations` argument ([#36585](https://github.com/hashicorp/terraform-provider-aws/issues/36585))
* resource/aws_ec2_host: Add user configurable timeouts ([#36538](https://github.com/hashicorp/terraform-provider-aws/issues/36538))
* resource/aws_glacier_vault_lock: Allow `policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_iam_group_policy: Allow `policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_iam_policy: Allow `policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_iam_role: Allow `assume_role_policy` and `inline_policy.*.policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_iam_role_policy: Allow `policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_iam_user_policy: Allow `policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_kinesisanalyticsv2_application: Add support for `FLINK-1_18` `runtime_environment` value ([#36562](https://github.com/hashicorp/terraform-provider-aws/issues/36562))
* resource/aws_media_store_container_policy: Allow `policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_ssoadmin_permission_set_inline_policy: Allow `inline_policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_transfer_access: Allow `policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_transfer_user: Allow `policy` to have leading whitespace ([#36597](https://github.com/hashicorp/terraform-provider-aws/issues/36597))
* resource/aws_vpc_ipam: Add `tier` argument ([#36504](https://github.com/hashicorp/terraform-provider-aws/issues/36504))

BUG FIXES:

* data-source/aws_cur_report_definition: Direct all API calls to the `us-east-1` endpoint as this is the only Region in which AWS Cost and Usage Reports is available ([#36540](https://github.com/hashicorp/terraform-provider-aws/issues/36540))
* resource/aws_applicationinsights_application: Make `ACTIVE` a valid create target status ([#36615](https://github.com/hashicorp/terraform-provider-aws/issues/36615))
* resource/aws_autoscaling_group: Don't attempt to remove scale-in protection from instances that don't have the feature enabled ([#36586](https://github.com/hashicorp/terraform-provider-aws/issues/36586))
* resource/aws_cur_report_definition: Direct all API calls to the `us-east-1` endpoint as this is the only Region in which AWS Cost and Usage Reports is available ([#36540](https://github.com/hashicorp/terraform-provider-aws/issues/36540))
* resource/aws_elasticsearch_domain_policy: Handle delayed domain status propagation, preventing a `ValidationException`. ([#36592](https://github.com/hashicorp/terraform-provider-aws/issues/36592))
* resource/aws_iam_instance_profile: Detect when the associated `role` no longer exists ([#34099](https://github.com/hashicorp/terraform-provider-aws/issues/34099))
* resource/aws_instance: Replace an instance when an `instance_type` change also requires an architecture change, such as x86_64 to arm64 ([#36590](https://github.com/hashicorp/terraform-provider-aws/issues/36590))
* resource/aws_opensearch_domain_policy: Handle delayed domain status propagation, preventing a `ValidationException`. ([#36592](https://github.com/hashicorp/terraform-provider-aws/issues/36592))
* resource/aws_quicksight_dashboard: Fix failure when updating a dashboard takes a while ([#34227](https://github.com/hashicorp/terraform-provider-aws/issues/34227))
* resource/aws_quicksight_template: Fix "Invalid address to set" errors ([#34227](https://github.com/hashicorp/terraform-provider-aws/issues/34227))
* resource/aws_quicksight_template: Fix "a number is required" errors when state contains an empty string ([#34227](https://github.com/hashicorp/terraform-provider-aws/issues/34227))
* resource/aws_redshift_cluster: Fix `InvalidParameterCombination` errors when updating only `skip_final_snapshot` ([#36635](https://github.com/hashicorp/terraform-provider-aws/issues/36635))
* resource/aws_route53_zone: Prevent re-creation when `name` casing changes ([#36563](https://github.com/hashicorp/terraform-provider-aws/issues/36563))
* resource/aws_secretsmanager_secret_version: Fix to handle versions deleted out-of-band without raising an `InvalidRequestException` ([#36609](https://github.com/hashicorp/terraform-provider-aws/issues/36609))
* resource/aws_ssm_parameter: force create a new SSM parameter when `data_type` is updated. ([#35960](https://github.com/hashicorp/terraform-provider-aws/issues/35960))

## 5.42.0 (March 22, 2024)

FEATURES:

* **New Data Source:** `aws_redshift_producer_data_shares` ([#36481](https://github.com/hashicorp/terraform-provider-aws/issues/36481))
* **New Resource:** `aws_devopsguru_event_sources_config` ([#36485](https://github.com/hashicorp/terraform-provider-aws/issues/36485))
* **New Resource:** `aws_devopsguru_resource_collection` ([#36489](https://github.com/hashicorp/terraform-provider-aws/issues/36489))
* **New Resource:** `aws_dynamodb_table_export` ([#30399](https://github.com/hashicorp/terraform-provider-aws/issues/30399))

ENHANCEMENTS:

* data-source/aws_vpc_peering_connection: Add `ipv6_cidr_block_set` and `peer_ipv6_cidr_block_set` attributes ([#36391](https://github.com/hashicorp/terraform-provider-aws/issues/36391))
* resource/aws_datasync_location_hdfs: Add `kerberos_keytab_base64` and `kerberos_krb5_conf_base64` arguments ([#36072](https://github.com/hashicorp/terraform-provider-aws/issues/36072))
* resource/aws_finspace_kx_dataview: Add `read_write` and `segment_configuration.on_demand` arguments ([#36486](https://github.com/hashicorp/terraform-provider-aws/issues/36486))
* resource/aws_rds_cluster: Add `enable_local_write_forwarding` argument to support Aurora MySQL local write forwarding ([#34370](https://github.com/hashicorp/terraform-provider-aws/issues/34370))

BUG FIXES:

* provider: Change the default AWS SDK for Go v2 API client [`RateLimiter`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/retry#RateLimiter) to [`ratelimit.None`](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/aws/ratelimit#pkg-variables) so that services migrated to AWS SDK for Go v2 maintain behavioral compatibility with AWS SDK for Go v1 ([#36467](https://github.com/hashicorp/terraform-provider-aws/issues/36467))
* resource/aws_appautoscaling_policy: Fix errors when importing an MSK storage autoscaling policy ([#34934](https://github.com/hashicorp/terraform-provider-aws/issues/34934))
* resource/aws_appautoscaling_scheduled_action: Always send `start_time` and `end_time` values on update when configured ([#33713](https://github.com/hashicorp/terraform-provider-aws/issues/33713))
* resource/aws_appautoscaling_scheduled_action: Read correct resource by using `scalable_dimension` as an additional filter ([#34382](https://github.com/hashicorp/terraform-provider-aws/issues/34382))
* resource/aws_datasync_location_azure_blob: Fix missing `container_url` attribute value and bad `subdirectory` attribute value from state read/refresh ([#36072](https://github.com/hashicorp/terraform-provider-aws/issues/36072))
* resource/aws_datasync_location_efs: Fix missing `efs_file_system_arn` attribute value from state read/refresh ([#36072](https://github.com/hashicorp/terraform-provider-aws/issues/36072))
* resource/aws_datasync_location_hdfs: Mark `qop_configuration` as Computed ([#36072](https://github.com/hashicorp/terraform-provider-aws/issues/36072))
* resource/aws_datasync_location_nfs: Fix missing `server_hostname` attribute value from state read/refresh ([#36072](https://github.com/hashicorp/terraform-provider-aws/issues/36072))
* resource/aws_datasync_location_s3: Fix missing `s3_bucket_arn` attribute value from state read/refresh ([#36072](https://github.com/hashicorp/terraform-provider-aws/issues/36072))
* resource/aws_datasync_location_smb: Fix missing `server_hostname` attribute value from state read/refresh ([#36072](https://github.com/hashicorp/terraform-provider-aws/issues/36072))
* resource/aws_dms_replication_config: Fix persistent change in `replication_settings` ([#35670](https://github.com/hashicorp/terraform-provider-aws/issues/35670))
* resource/aws_dms_replication_task: Fix persistent change in `replication_task_settings` ([#35670](https://github.com/hashicorp/terraform-provider-aws/issues/35670))
* resource/aws_eks_access_entry: Always send `kubernetes_groups` and `user_name` values on update when configured ([#36484](https://github.com/hashicorp/terraform-provider-aws/issues/36484))
* resource/aws_glue_job: Adjust `number_of_workers` minimum value to `1` ([#36458](https://github.com/hashicorp/terraform-provider-aws/issues/36458))
* resource/aws_lexv2models_slot: Fix custom_payload typo ([#36488](https://github.com/hashicorp/terraform-provider-aws/issues/36488))
* resource/aws_route: Allow resource creation if a propagated route to the same destination exists ([#36512](https://github.com/hashicorp/terraform-provider-aws/issues/36512))
* resource/aws_vpn_connection: `local_ipv6_network_cidr`, `remote_ipv6_network_cidr`, `tunnel1_inside_ipv6_cidr`, and `tunnel2_inside_ipv6_cidr` no longer require `transit_gateway_id` to be specified ([#36405](https://github.com/hashicorp/terraform-provider-aws/issues/36405))

## 5.41.0 (March 14, 2024)

FEATURES:

* **New Data Source:** `aws_apprunner_hosted_zone_id` ([#36288](https://github.com/hashicorp/terraform-provider-aws/issues/36288))
* **New Data Source:** `aws_medialive_input` ([#36307](https://github.com/hashicorp/terraform-provider-aws/issues/36307))
* **New Resource:** `aws_lakeformation_data_cells_filter` ([#36264](https://github.com/hashicorp/terraform-provider-aws/issues/36264))
* **New Resource:** `aws_securityhub_configuration_policy` ([#35752](https://github.com/hashicorp/terraform-provider-aws/issues/35752))
* **New Resource:** `aws_securityhub_configuration_policy_association` ([#35752](https://github.com/hashicorp/terraform-provider-aws/issues/35752))
* **New Resource:** `aws_securitylake_subscriber_notification` ([#36323](https://github.com/hashicorp/terraform-provider-aws/issues/36323))

ENHANCEMENTS:

* data-source/aws_ec2_transit_gateway_peering_attachment: Add `state` attribute ([#36304](https://github.com/hashicorp/terraform-provider-aws/issues/36304))
* data-source/aws_lakeformation_permissions: Add `data_cells_filter` attribute ([#36264](https://github.com/hashicorp/terraform-provider-aws/issues/36264))
* data-source/aws_ram_resource_share: `name` is Optional ([#36062](https://github.com/hashicorp/terraform-provider-aws/issues/36062))
* resource/aws_cognito_user_pool: Add `pre_token_generation_config` configuration block ([#35236](https://github.com/hashicorp/terraform-provider-aws/issues/35236))
* resource/aws_ec2_transit_gateway_peering_attachment: Add `state` attribute ([#36304](https://github.com/hashicorp/terraform-provider-aws/issues/36304))
* resource/aws_ecs_cluster: Add default value (`DEFAULT`) for `configuration.execute_command_configuration.logging` ([#36341](https://github.com/hashicorp/terraform-provider-aws/issues/36341))
* resource/aws_lakeformation_permissions: Add `data_cells_filter` attribute ([#36264](https://github.com/hashicorp/terraform-provider-aws/issues/36264))
* resource/aws_ram_resource_association: Add plan-time validation of `resource_arn` and `resource_share_arn` ([#36062](https://github.com/hashicorp/terraform-provider-aws/issues/36062))
* resource/aws_route53domains_registered_domain: Add `billing_contact` and `billing_privacy` arguments ([#36285](https://github.com/hashicorp/terraform-provider-aws/issues/36285))
* resource/aws_securityhub_organization_configuration: Add `organization_configuration` configuration block to support [central configuration](https://docs.aws.amazon.com/securityhub/latest/userguide/start-central-configuration.html) ([#35752](https://github.com/hashicorp/terraform-provider-aws/issues/35752))
* resource/aws_securityhub_organization_configuration: Set `auto_enable` to `false`, `auto_enable_standards` to `NONE`, and `organization_configuration.configuration_type` to `LOCAL` on resource Delete ([#35752](https://github.com/hashicorp/terraform-provider-aws/issues/35752))

BUG FIXES:

* data-source/aws_iam_policy_document: Fix `Failed to marshal state to json: unsupported attribute "override_json"` and `Failed to marshal state to json: unsupported attribute "source_json"` errors when running `terraform show -json` or `terraform state rm` ([#36383](https://github.com/hashicorp/terraform-provider-aws/issues/36383))
* data-source/aws_opensearch_domain : Add `auto_tune_options.use_off_peak_window` attribute. This fixes a regression introduced in [v5.40.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#5400-march--7-2024) causing `Invalid address to set` errors ([#36298](https://github.com/hashicorp/terraform-provider-aws/issues/36298))
* resource/aws_cognito_identity_pool: Fix handling of resources deleted out of band ([#36100](https://github.com/hashicorp/terraform-provider-aws/issues/36100))
* resource/aws_cognito_identity_provider: Fix `InvalidParameterException: ActiveEncryptionCertificate is not a valid key for SAML identity provider details` errors on resource Update ([#36311](https://github.com/hashicorp/terraform-provider-aws/issues/36311))
* resource/aws_ec2_instance: Remove [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) from `ipv6_address_count` ([#36308](https://github.com/hashicorp/terraform-provider-aws/issues/36308))
* resource/aws_ecs_cluster: Fix `panic: interface conversion: interface {} is nil, not map[string]interface {}` when `configuration`, `configuration.execute_command_configuration`, or `configuration.execute_command_configuration.log_configuration` are empty ([#36341](https://github.com/hashicorp/terraform-provider-aws/issues/36341))
* resource/aws_ecs_service: Fix `panic: interface conversion: interface {} is nil, not map[string]interface {}` when `service_connect_configuration.service.timeout` is empty ([#36309](https://github.com/hashicorp/terraform-provider-aws/issues/36309))
* resource/aws_ecs_service: `service_connect_configuration.service.tls.issuer_cert_authority.aws_pca_authority_arn` is Required ([#36309](https://github.com/hashicorp/terraform-provider-aws/issues/36309))
* resource/aws_elasticache_replication_group: Fix bugs causing errors like `InvalidReplicationGroupState: Cluster not in available state to perform tagging operations.` ([#36310](https://github.com/hashicorp/terraform-provider-aws/issues/36310))
* resource/aws_finspace_kx_cluster: Prevent `command_line_arguments` and `initialization_script` updates from overwriting one another ([#36361](https://github.com/hashicorp/terraform-provider-aws/issues/36361))
* resource/aws_network_acl_rule: Fix `InvalidNetworkAclID.NotFound` errors on resource Delete ([#36326](https://github.com/hashicorp/terraform-provider-aws/issues/36326))
* resource/aws_network_acl_rule: Prevent creation of duplicate Terraform resources ([#36326](https://github.com/hashicorp/terraform-provider-aws/issues/36326))
* resource/aws_ram_principal_association: Prevent creation of duplicate Terraform resources ([#36062](https://github.com/hashicorp/terraform-provider-aws/issues/36062))
* resource/aws_ram_principal_association: Remove from state on resource Read if `principal` is disassociated outside of Terraform ([#36062](https://github.com/hashicorp/terraform-provider-aws/issues/36062))
* resource/aws_ram_resource_association: Prevent creation of duplicate Terraform resources ([#36062](https://github.com/hashicorp/terraform-provider-aws/issues/36062))
* resource/aws_route: Prevent creation of duplicate Terraform resources ([#36326](https://github.com/hashicorp/terraform-provider-aws/issues/36326))
* resource/aws_route_table: Fix `couldn't find resource` errors on resource Delete ([#36326](https://github.com/hashicorp/terraform-provider-aws/issues/36326))
* resource/aws_vpn_connection: Correct plan-time validation of `tunnel1_inside_ipv6_cidr` and `tunnel2_inside_ipv6_cidr` ([#36236](https://github.com/hashicorp/terraform-provider-aws/issues/36236))

## 5.40.0 (March  7, 2024)

FEATURES:

* **New Function:** `arn_build` ([#34952](https://github.com/hashicorp/terraform-provider-aws/issues/34952))
* **New Function:** `arn_parse` ([#34952](https://github.com/hashicorp/terraform-provider-aws/issues/34952))
* **New Resource:** `aws_account_region` ([#35739](https://github.com/hashicorp/terraform-provider-aws/issues/35739))
* **New Resource:** `aws_securitylake_subscriber` ([#35981](https://github.com/hashicorp/terraform-provider-aws/issues/35981))

ENHANCEMENTS:

* data-source/aws_rds_engine_version: Add `has_major_target` and `has_minor_target` optional arguments and `valid_major_targets` and `valid_minor_targets` attributes ([#36246](https://github.com/hashicorp/terraform-provider-aws/issues/36246))
* resource/aws_batch_job_queue: added parameter `compute_environment_order` which conflicts with `compute_environments` but aligns with AWS API. `compute_environments` has been deprecated. ([#34750](https://github.com/hashicorp/terraform-provider-aws/issues/34750))
* resource/aws_cloudfront_distribution: Remove the upper limit on `origin.custom_origin_config.origin_read_timeout` ([#36088](https://github.com/hashicorp/terraform-provider-aws/issues/36088))
* resource/aws_db_instance: Add `io2` as a valid value for `storage_type` ([#36252](https://github.com/hashicorp/terraform-provider-aws/issues/36252))
* resource/aws_elasticache_serverless_cache: Add plan-time validation of `cache_usage_limits.ecpu_per_second.maximum` ([#35927](https://github.com/hashicorp/terraform-provider-aws/issues/35927))
* resource/aws_iot_policy: Add tagging support ([#36102](https://github.com/hashicorp/terraform-provider-aws/issues/36102))
* resource/aws_iot_role_alias: Add tagging support ([#36255](https://github.com/hashicorp/terraform-provider-aws/issues/36255))
* resource/aws_opensearch_domain: Add `use_off_peak_window` argument to the `auto_tune_options` configuration block ([#36067](https://github.com/hashicorp/terraform-provider-aws/issues/36067))
* resource/aws_rds_cluster: Add `io2` as a valid value for `storage_type` ([#36252](https://github.com/hashicorp/terraform-provider-aws/issues/36252))
* resource/aws_s3_bucket_object: Adds attribute `arn`. ([#35710](https://github.com/hashicorp/terraform-provider-aws/issues/35710))
* resource/aws_s3_object: Adds attribute `arn`. ([#35710](https://github.com/hashicorp/terraform-provider-aws/issues/35710))
* resource/aws_s3_object_copy: Adds attribute `arn`. ([#35710](https://github.com/hashicorp/terraform-provider-aws/issues/35710))
* resource/aws_wafv2_rule_group: Add `evaluation_window_sec` argument to the `rate_based_statement` configuration block ([#36045](https://github.com/hashicorp/terraform-provider-aws/issues/36045))
* resource/aws_wafv2_web_acl: Add `evaluation_window_sec` argument to the `rate_based_statement` configuration block ([#36045](https://github.com/hashicorp/terraform-provider-aws/issues/36045))

BUG FIXES:

* data-source/aws_rds_engine_version: Fix bugs that could limit engine version to a default version even when not appropriate ([#36246](https://github.com/hashicorp/terraform-provider-aws/issues/36246))
* resource/aws_db_instance: Correctly sets `parameter_group_name` when `replicate_source_db` is in different region. ([#36080](https://github.com/hashicorp/terraform-provider-aws/issues/36080))
* resource/aws_elastic_beanstalk_environment: Fix `InvalidParameterValue: Environment named ... is in an invalid state for this operation. Must be Ready` errors when `tags` are updated along with other attributes ([#36074](https://github.com/hashicorp/terraform-provider-aws/issues/36074))
* resource/aws_elasticache_serverless_cache: Change `cache_usage_limits.data_storage.maximum` and `cache_usage_limits.ecpu_per_second.maximum` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#35927](https://github.com/hashicorp/terraform-provider-aws/issues/35927))
* resource/aws_medialive_channel: Fix handling of optional `encoder_settings.audio_descriptions` arguments ([#36097](https://github.com/hashicorp/terraform-provider-aws/issues/36097))
* resource/aws_rds_global_cluster: Fix bugs and delays that could occur when performing major or minor version upgrades ([#36246](https://github.com/hashicorp/terraform-provider-aws/issues/36246))
* resource/aws_s3_bucket: Tags with empty values no longer remove all tags. ([#35710](https://github.com/hashicorp/terraform-provider-aws/issues/35710))
* resource/aws_s3_bucket_object: Tags with empty values no longer remove all tags. ([#35710](https://github.com/hashicorp/terraform-provider-aws/issues/35710))
* resource/aws_s3_object: Tags with empty values no longer remove all tags. ([#35710](https://github.com/hashicorp/terraform-provider-aws/issues/35710))
* resource/aws_s3_object_copy: Tags with empty values no longer remove all tags. ([#35710](https://github.com/hashicorp/terraform-provider-aws/issues/35710))
* resource/aws_vpclattice_listener_rule: Remove `action.forward.target_groups` maximum item limit ([#36095](https://github.com/hashicorp/terraform-provider-aws/issues/36095))

## 5.39.1 (March  1, 2024)

BUG FIXES:

* data-source/aws_instance: Fix `panic: Invalid address to set` related to `root_block_device.0.tags_all` ([#36054](https://github.com/hashicorp/terraform-provider-aws/issues/36054))

## 5.39.0 (February 29, 2024)

FEATURES:

* **New Data Source:** `aws_redshift_data_shares` ([#35937](https://github.com/hashicorp/terraform-provider-aws/issues/35937))
* **New Resource:** `aws_apprunner_deployment` ([#35758](https://github.com/hashicorp/terraform-provider-aws/issues/35758))
* **New Resource:** `aws_config_retention_configuration` ([#15136](https://github.com/hashicorp/terraform-provider-aws/issues/15136))
* **New Resource:** `aws_securityhub_automation_rule` ([#34781](https://github.com/hashicorp/terraform-provider-aws/issues/34781))
* **New Resource:** `aws_shield_proactive_engagement` ([#34667](https://github.com/hashicorp/terraform-provider-aws/issues/34667))

ENHANCEMENTS:

* aws_kinesis_firehose_delivery_stream: Add `custom_time_zone` and `file_extension` arguments to the `extended_S3_configuration` configuration block ([#35969](https://github.com/hashicorp/terraform-provider-aws/issues/35969))
* resource/aws_appflow_flow: Allow `task.source_fields` to be a `null` value ([#35993](https://github.com/hashicorp/terraform-provider-aws/issues/35993))
* resource/aws_codepipeline: Add `trigger` configuration block ([#35475](https://github.com/hashicorp/terraform-provider-aws/issues/35475))
* resource/aws_config_configuration_recorder: Add plan-time validation of `aws_config_organization_custom_rule.lambda_function_arn` ([#15136](https://github.com/hashicorp/terraform-provider-aws/issues/15136))
* resource/aws_instance: Add configurable `read` timeout ([#35955](https://github.com/hashicorp/terraform-provider-aws/issues/35955))
* resource/aws_instance: Apply default tags to volumes/block devices managed through an `aws_instance`, add `ebs_block_device.*.tags_all` and `root_block_device.*.tags_all` attributes which include default tags ([#33769](https://github.com/hashicorp/terraform-provider-aws/issues/33769))
* resource/aws_mq_broker: Add `data_replication_mode` and `data_replication_primary_broker_arn` arguments, enabling support for cross-region data replication ([#35990](https://github.com/hashicorp/terraform-provider-aws/issues/35990))
* resource/aws_mwaa_environment: Add `endpoint_management` attribute ([#35961](https://github.com/hashicorp/terraform-provider-aws/issues/35961))
* resource/aws_redshiftserverless_namespace:
Add attributes `admin_password_secret_kms_key_id` and `manage_admin_password` ([#35965](https://github.com/hashicorp/terraform-provider-aws/issues/35965))
* resource/aws_shield_drt_access_log_bucket_association: Support resource import ([#34667](https://github.com/hashicorp/terraform-provider-aws/issues/34667))
* resource/aws_shield_drt_access_role_arn_association: Support resource import ([#34667](https://github.com/hashicorp/terraform-provider-aws/issues/34667))
* resource/aws_spot_instance_request: Add configurable `read` timeout ([#35955](https://github.com/hashicorp/terraform-provider-aws/issues/35955))
* resource/aws_wafv2_web_acl: Add `application_integration_url` attribute ([#35974](https://github.com/hashicorp/terraform-provider-aws/issues/35974))

BUG FIXES:

* data/aws_redshiftserverless_namespace: Properly set `iam_roles` attribute on read ([#35965](https://github.com/hashicorp/terraform-provider-aws/issues/35965))
* resource/aws_appflow_flow: Fix perpetual diff when `task.task_type` is set to `Map_all` ([#35993](https://github.com/hashicorp/terraform-provider-aws/issues/35993))
* resource/aws_config_configuration_recorder: Fix `panic: interface conversion: interface {} is nil, not map[string]interface {}` when `recording_group.exclusion_by_resource_types` is empty ([#15136](https://github.com/hashicorp/terraform-provider-aws/issues/15136))
* resource/aws_config_rule: Change `name` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#15136](https://github.com/hashicorp/terraform-provider-aws/issues/15136))
* resource/aws_config_rule: Fix `InvalidParameterValueException: PolicyText is required when Owner is CUSTOM_POLICY` errors on resource Update ([#15136](https://github.com/hashicorp/terraform-provider-aws/issues/15136))
* resource/aws_ecs_task_definition: Fix perpetual `container_definitions` diffs when `Name`s are ordered differently ([#36029](https://github.com/hashicorp/terraform-provider-aws/issues/36029))
* resource/aws_msk_replicator: Fix incorrect `detect_and_copy_new_topics` attribute value from state read/refresh ([#35966](https://github.com/hashicorp/terraform-provider-aws/issues/35966))
* resource/aws_redshiftserverless_workgroup: Fix `max_capacity` removal ([#36032](https://github.com/hashicorp/terraform-provider-aws/issues/36032))
* resource/aws_redshiftserverless_workgroup: Fix updating both `base_capacity` and `max_capacity` ([#36032](https://github.com/hashicorp/terraform-provider-aws/issues/36032))
* resource/aws_shield_drt_access_log_bucket_association: Change `log_bucket` and `role_arn_association_id` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#34667](https://github.com/hashicorp/terraform-provider-aws/issues/34667))

## 5.38.0 (February 22, 2024)

FEATURES:

* **New Data Source:** `aws_batch_job_definition` ([#34663](https://github.com/hashicorp/terraform-provider-aws/issues/34663))
* **New Data Source:** `aws_cognito_user_group` ([#34046](https://github.com/hashicorp/terraform-provider-aws/issues/34046))
* **New Data Source:** `aws_cognito_user_groups` ([#34046](https://github.com/hashicorp/terraform-provider-aws/issues/34046))

ENHANCEMENTS:

* data-source/aws_alb_target_group: Add `load_balancer_arns` attribute ([#34364](https://github.com/hashicorp/terraform-provider-aws/issues/34364))
* data-source/aws_ec2_instance_type: Add `maximum_network_cards` attribute ([#35840](https://github.com/hashicorp/terraform-provider-aws/issues/35840))
* data-source/aws_elasticache_subnet_group: Add `vpc_id` attribute ([#35887](https://github.com/hashicorp/terraform-provider-aws/issues/35887))
* data-source/aws_lb_target_group: Add `load_balancer_arns` attribute ([#34364](https://github.com/hashicorp/terraform-provider-aws/issues/34364))
* provider: Add `token_bucket_rate_limiter_capacity` parameter ([#35926](https://github.com/hashicorp/terraform-provider-aws/issues/35926))
* resource/aws_alb_target_group: Add `load_balancer_arns` attribute ([#34364](https://github.com/hashicorp/terraform-provider-aws/issues/34364))
* resource/aws_codedeploy_deployment_config: Add `arn` attribute ([#35888](https://github.com/hashicorp/terraform-provider-aws/issues/35888))
* resource/aws_codepipeline: Add `execution_mode` argument ([#35875](https://github.com/hashicorp/terraform-provider-aws/issues/35875))
* resource/aws_config_configuration_recorder: Add `recording_mode` configuration block ([#35527](https://github.com/hashicorp/terraform-provider-aws/issues/35527))
* resource/aws_db_instance: Add plan-time validation of `performance_insights_retention_period` ([#35870](https://github.com/hashicorp/terraform-provider-aws/issues/35870))
* resource/aws_elasticache_subnet_group: Add `vpc_id` attribute ([#35887](https://github.com/hashicorp/terraform-provider-aws/issues/35887))
* resource/aws_lb_target_group: Add `load_balancer_arns` attribute ([#34364](https://github.com/hashicorp/terraform-provider-aws/issues/34364))
* resource/aws_redshiftserverless_workgroup: Add `max_capacity` argument ([#35720](https://github.com/hashicorp/terraform-provider-aws/issues/35720))
* resource/aws_transfer_server: Add `TransferSecurityPolicy-2024-01` and `TransferSecurityPolicy-FIPS-2024-01` as valid values for `security_policy_name` ([#35879](https://github.com/hashicorp/terraform-provider-aws/issues/35879))

BUG FIXES:

* data-source/aws_caller_identity: Fix authentication signature error when alternate `sts_region` is specified ([#35860](https://github.com/hashicorp/terraform-provider-aws/issues/35860))
* data-source/aws_eks_access_entry: Fix `cluster_name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* data-source/aws_eks_addon: Fix `cluster_name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* data-source/aws_eks_cluster: Fix `name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* resource/aws_cloudsearch_domain: Prevent panic when reading nil `index_field` options response values ([#35900](https://github.com/hashicorp/terraform-provider-aws/issues/35900))
* resource/aws_eks_access_entry: Fix `cluster_name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* resource/aws_eks_access_policy_association: Fix `cluster_name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* resource/aws_eks_addon: Fix `cluster_name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* resource/aws_eks_cluster: Fix `name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* resource/aws_eks_fargate_profile: Fix `cluster_name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* resource/aws_eks_node_group: Fix `cluster_name` plan-time validation, allowing single-character names ([#35874](https://github.com/hashicorp/terraform-provider-aws/issues/35874))
* resource/aws_prometheus_scraper: Fixes invalid result after apply error. ([#35844](https://github.com/hashicorp/terraform-provider-aws/issues/35844))
* resource/aws_sqs_queue_policy: Retry IAM eventual consistency errors ([#35861](https://github.com/hashicorp/terraform-provider-aws/issues/35861))

## 5.37.0 (February 15, 2024)

NOTES:

* provider: Updates to Go 1.21 (used by Terraform starting with v1.6.0), which, for Windows, requires at least Windows 10 or Windows Server 2016--support for previous versions has been discontinued--and, for macOS, requires macOS 10.15 Catalina or later--support for previous versions has been discontinued. ([#35832](https://github.com/hashicorp/terraform-provider-aws/issues/35832))
* resource/aws_bedrock_provisioned_model_throughput: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#35689](https://github.com/hashicorp/terraform-provider-aws/issues/35689))

FEATURES:

* **New Data Source:** `aws_db_parameter_group` ([#35698](https://github.com/hashicorp/terraform-provider-aws/issues/35698))
* **New Resource:** `aws_bedrock_provisioned_model_throughput` ([#35689](https://github.com/hashicorp/terraform-provider-aws/issues/35689))
* **New Resource:** `aws_cloudfront_key_value_store` ([#35663](https://github.com/hashicorp/terraform-provider-aws/issues/35663))
* **New Resource:** `aws_redshift_data_share_consumer_association` ([#35771](https://github.com/hashicorp/terraform-provider-aws/issues/35771))

ENHANCEMENTS:

* data-source/aws_ecr_pull_through_cache_rule: Add `credential_arn` attribute ([#34475](https://github.com/hashicorp/terraform-provider-aws/issues/34475))
* data-source/aws_ecs_task_execution: Add `client_token` argument ([#34402](https://github.com/hashicorp/terraform-provider-aws/issues/34402))
* data-source/aws_neptune_cluster_instance: Add `skip_final_snapshot` argument ([#35698](https://github.com/hashicorp/terraform-provider-aws/issues/35698))
* data-source/aws_rds_engine_version: Improve search functionality and options by adding `latest`, `preferred_major_targets`, and `preferred_upgrade_targets`. Add `version_actual` attribute ([#35698](https://github.com/hashicorp/terraform-provider-aws/issues/35698))
* data-source/aws_rds_orderable_db_instance: Improve search functionality and options by adding `engine_latest_version` and `supports_clusters` arguments and converting `read_replica_capable`, `supported_engine_modes`, `supported_network_types`, and `supports_multi_az` to arguments for use as search criteria ([#35698](https://github.com/hashicorp/terraform-provider-aws/issues/35698))
* resource/aws_appsync_graphql_api: Add `introspection_config`, `query_depth_limit`, and `resolver_count_limit` arguments ([#35631](https://github.com/hashicorp/terraform-provider-aws/issues/35631))
* resource/aws_codeartifact_domain: Add `s3_bucket_arn` attribute ([#35760](https://github.com/hashicorp/terraform-provider-aws/issues/35760))
* resource/aws_ecr_pull_through_cache_rule: Add `credential_arn` argument ([#34475](https://github.com/hashicorp/terraform-provider-aws/issues/34475))
* resource/aws_ecs_service: Add `service_connect_configuration.service.timeout` and `service_connect_configuration.service.tls` configuration blocks ([#35684](https://github.com/hashicorp/terraform-provider-aws/issues/35684))
* resource/aws_ecs_task_definition: Add `track_latest` argument ([#30154](https://github.com/hashicorp/terraform-provider-aws/issues/30154))
* resource/aws_glue_catalog_database: Add `federated_database` argument ([#35799](https://github.com/hashicorp/terraform-provider-aws/issues/35799))
* resource/aws_glue_trigger: Add configurable `timeouts` ([#35542](https://github.com/hashicorp/terraform-provider-aws/issues/35542))
* resource/aws_rds_cluster: Add `domain` and `domain_iam_role_name` arguments to support [Kerberos authentication](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RDS_Fea_Regions_DB-eng.Feature.KerberosAuthentication.html) ([#35753](https://github.com/hashicorp/terraform-provider-aws/issues/35753))
* resource/aws_route53_record: Add `geoproximity_routing_policy` configuration block to support [geoproximity routing](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/routing-policy-geoproximity.html) ([#35565](https://github.com/hashicorp/terraform-provider-aws/issues/35565))
* resource/aws_route53_resolver_rule: Add `target_ip.protocol` argument ([#35744](https://github.com/hashicorp/terraform-provider-aws/issues/35744))
* resource/aws_sagemaker_endpoint_configuration: Add `routing_config` argument. Enables the specification of a `routing_strategy`. ([#34777](https://github.com/hashicorp/terraform-provider-aws/issues/34777))
* resource/aws_sagemaker_space: Add `ownership_settings`, `space_sharing_settings`, `space_settings.app_type`, `space_settings.code_editor_app_settings`, `space_settings.custom_file_system`, `space_settings.jupyter_lab_app_settings`, and `space_settings.space_storage_settings` arguments ([#35116](https://github.com/hashicorp/terraform-provider-aws/issues/35116))

BUG FIXES:

* provider: Fix `failed to get rate limit token, retry quota exceeded` errors ([#35817](https://github.com/hashicorp/terraform-provider-aws/issues/35817))
* resource/aws_apigateway_domain_name: Properly send changes to `ownership_verification_certificate_arn` on update ([#35777](https://github.com/hashicorp/terraform-provider-aws/issues/35777))
* resource/aws_apigatewayv2_route: Fix `BadRequestException: Unable to update route. Authorizer type is invalid or null` errors when updating `authorizer_id` ([#35821](https://github.com/hashicorp/terraform-provider-aws/issues/35821))
* resource/aws_autoscaling_group: Fix version to computed for inconsistent final plan issue ([#35774](https://github.com/hashicorp/terraform-provider-aws/issues/35774))
* resource/aws_datasync_task: Fix crash when reading empty `report_override` values ([#35778](https://github.com/hashicorp/terraform-provider-aws/issues/35778))
* resource/aws_datasync_task: Prevent ValidationErrors when empty values are sent with `report_override` arguments ([#35778](https://github.com/hashicorp/terraform-provider-aws/issues/35778))
* resource/aws_db_proxy: Change `auth` from `TypeList` to `TypeSet` as order is not significant ([#35819](https://github.com/hashicorp/terraform-provider-aws/issues/35819))
* resource/aws_ecs_account_setting_default: Remove plan-time validation of `value` ([#33393](https://github.com/hashicorp/terraform-provider-aws/issues/33393))
* resource/aws_ecs_task_definition: Fix perpetual `container_definitions` diffs when `Secrets` are ordered differently ([#35792](https://github.com/hashicorp/terraform-provider-aws/issues/35792))
* resource/aws_eks_access_policy_association: Retry IAM eventual consistency errors on create ([#35736](https://github.com/hashicorp/terraform-provider-aws/issues/35736))
* resource/aws_instance: Fix `ReservationCapacityExceeded` errors when updating `instance_type` and `capacity_reservation_specification.capacity_reservation_target.capacity_reservation_id` ([#33412](https://github.com/hashicorp/terraform-provider-aws/issues/33412))
* resource/aws_lakeformation_resource: Properly handle configured `false` values for `use_service_linked_role` ([#35799](https://github.com/hashicorp/terraform-provider-aws/issues/35799))
* resource/aws_medialive_channel: Added `client_cache` to `hls_group_settings`. ([#35738](https://github.com/hashicorp/terraform-provider-aws/issues/35738))
* resource/aws_ram_resource_share_accepter: Fix handling of out-of-band resource share deletion ([#35800](https://github.com/hashicorp/terraform-provider-aws/issues/35800))
* resource/aws_redshift_data_share_authorization: Fix read operation to properly handle shares in `ACTIVE` status ([#35771](https://github.com/hashicorp/terraform-provider-aws/issues/35771))
* resource/aws_s3_bucket_acl: Correctly updates `access_control_policy` when switching configuration to `acl`. ([#35775](https://github.com/hashicorp/terraform-provider-aws/issues/35775))
* resource/resource_share_acceptor: Wait until RAM resource share available after accepting the invitation ([#34753](https://github.com/hashicorp/terraform-provider-aws/issues/34753))

## 5.36.0 (February  8, 2024)

NOTES:

* data-source/aws_media_convert_queue: The AWS Elemental MediaConvert service has been converted to use standard [Regional endpoints](https://docs.aws.amazon.com/general/latest/gr/mediaconvert.html#mediaconvert_region) instead of deprecated per-account endpoints ([#35615](https://github.com/hashicorp/terraform-provider-aws/issues/35615))
* resource/aws_controltower_landing_zone: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#34595](https://github.com/hashicorp/terraform-provider-aws/issues/34595))
* resource/aws_media_convert_queue: The AWS Elemental MediaConvert service has been converted to use standard [Regional endpoints](https://docs.aws.amazon.com/general/latest/gr/mediaconvert.html#mediaconvert_region) instead of deprecated per-account endpoints ([#35615](https://github.com/hashicorp/terraform-provider-aws/issues/35615))

FEATURES:

* **New Resource:** `aws_controltower_landing_zone` ([#34595](https://github.com/hashicorp/terraform-provider-aws/issues/34595))
* **New Resource:** `aws_osis_pipeline` ([#35582](https://github.com/hashicorp/terraform-provider-aws/issues/35582))
* **New Resource:** `aws_redshift_data_share_authorization` ([#35703](https://github.com/hashicorp/terraform-provider-aws/issues/35703))
* **New Resource:** `aws_securitylake_custom_log_source` ([#35354](https://github.com/hashicorp/terraform-provider-aws/issues/35354))

ENHANCEMENTS:

* resource/aws_cloudwatch_metric_stream: Add plan-time validation of `output_format` ([#35569](https://github.com/hashicorp/terraform-provider-aws/issues/35569))
* resource/aws_db_instance: Add `diag.log` and `notify.log` as valid values for `enabled_cloudwatch_logs_exports` ([#35626](https://github.com/hashicorp/terraform-provider-aws/issues/35626))
* resource/aws_db_instance: Add `domain_auth_secret_arn`, `domain_dns_ips`, `domain_fqdn`, and `domain_ou` arguments to support [self-managed Active Directory](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_SQLServer_SelfManagedActiveDirectory.html) ([#35500](https://github.com/hashicorp/terraform-provider-aws/issues/35500))
* resource/aws_s3_bucket_metric: Add `filter.access_point` argument ([#35590](https://github.com/hashicorp/terraform-provider-aws/issues/35590))
* resource/aws_verifiedaccess_group: Add `sse_configuration` argument ([#34055](https://github.com/hashicorp/terraform-provider-aws/issues/34055))

BUG FIXES:

* resource/aws_db_instance: Creating resource from point-in-time recovery now handles `password` attribute correctly ([#35589](https://github.com/hashicorp/terraform-provider-aws/issues/35589))
* resource/aws_dynamodb_table: Ensure that `replica`s are always set on Read ([#35630](https://github.com/hashicorp/terraform-provider-aws/issues/35630))
* resource/aws_emr_cluster: Properly normalize `launch_specifications.on_demand_specification.allocation_strategy` and `launch_specifications.spot_specification.allocation_strategy` values to fix perpetual state differences ([#34367](https://github.com/hashicorp/terraform-provider-aws/issues/34367))
* resource/aws_kinesis_firehose_delivery_stream: Change `extended_s3_configuration.processing_configuration.processors.parameters` from `TypeList` to `TypeSet` as order is not significant ([#35672](https://github.com/hashicorp/terraform-provider-aws/issues/35672))
* resource/aws_lambda_function: Resolve consecutive diff issue in `logging_config` when values for `application_log_level` or `system_log_level` are not specified ([#35694](https://github.com/hashicorp/terraform-provider-aws/issues/35694))
* resource/aws_lb_listener: Fixes unexpected diff when using `default_action` parameters which don't match the `type`. ([#35678](https://github.com/hashicorp/terraform-provider-aws/issues/35678))
* resource/aws_lb_listener: Was incorrectly reporting conflicting `default_action[].target_group_arn` when `ignore_changes` was set. ([#35671](https://github.com/hashicorp/terraform-provider-aws/issues/35671))
* resource/aws_lb_listener: Was not storing `default_action[].forward` in state if only a single `target_group` was set. ([#35671](https://github.com/hashicorp/terraform-provider-aws/issues/35671))
* resource/aws_lb_listener_rule: Fixes unexpected diff when using `action` parameters which don't match the `type`. ([#35678](https://github.com/hashicorp/terraform-provider-aws/issues/35678))
* resource/aws_lb_listener_rule: Was incorrectly reporting conflicting `action[].target_group_arn` when `ignore_changes` was set. ([#35671](https://github.com/hashicorp/terraform-provider-aws/issues/35671))
* resource/aws_lb_listener_rule: Was not storing `action[].forward` in state if only a single `target_group` was set. ([#35671](https://github.com/hashicorp/terraform-provider-aws/issues/35671))
* resource/aws_ssm_patch_baseline: Mark `json` as Computed if there are content changes ([#35606](https://github.com/hashicorp/terraform-provider-aws/issues/35606))

## 5.35.0 (February  2, 2024)

FEATURES:

* **New Data Source:** `aws_bedrock_custom_model` ([#34310](https://github.com/hashicorp/terraform-provider-aws/issues/34310))
* **New Data Source:** `aws_bedrock_custom_models` ([#34310](https://github.com/hashicorp/terraform-provider-aws/issues/34310))
* **New Data Source:** `aws_ssmcontacts_rotation` ([#32710](https://github.com/hashicorp/terraform-provider-aws/issues/32710))
* **New Resource:** `aws_bedrock_custom_model` ([#34310](https://github.com/hashicorp/terraform-provider-aws/issues/34310))
* **New Resource:** `aws_lexv2models_slot` ([#34617](https://github.com/hashicorp/terraform-provider-aws/issues/34617))
* **New Resource:** `aws_lexv2models_slot_type` ([#35555](https://github.com/hashicorp/terraform-provider-aws/issues/35555))
* **New Resource:** `aws_rekognition_collection` ([#35407](https://github.com/hashicorp/terraform-provider-aws/issues/35407))
* **New Resource:** `aws_sesv2_email_identity_policy` ([#35486](https://github.com/hashicorp/terraform-provider-aws/issues/35486))
* **New Resource:** `aws_ssmcontacts_rotation` ([#32710](https://github.com/hashicorp/terraform-provider-aws/issues/32710))

ENHANCEMENTS:

* data-source/aws_redshift_cluster: Add `multi_az` attribute ([#35508](https://github.com/hashicorp/terraform-provider-aws/issues/35508))
* resource/aws_lakeformation_resource: Add `hybrid_access_enabled` argument ([#35571](https://github.com/hashicorp/terraform-provider-aws/issues/35571))
* resource/aws_lakeformation_resource: Add `with_federation` argument ([#35154](https://github.com/hashicorp/terraform-provider-aws/issues/35154))
* resource/aws_redshift_cluster: Add `multi_az` argument ([#35508](https://github.com/hashicorp/terraform-provider-aws/issues/35508))
* resource/aws_redshiftserverless_endpoint_access: Add `owner_account` argument ([#35509](https://github.com/hashicorp/terraform-provider-aws/issues/35509))
* resource/aws_wafv2_rule_group: Add `header_order` to `field_to_match` configuration blocks ([#35521](https://github.com/hashicorp/terraform-provider-aws/issues/35521))
* resource/aws_wafv2_web_acl: Add `header_order`to `field_to_match` configuration blocks ([#35521](https://github.com/hashicorp/terraform-provider-aws/issues/35521))

BUG FIXES:

* data-source/aws_networkmanager_core_network_policy_document: Remove `core_network_configuration.edge_locations` maximum item limit ([#35585](https://github.com/hashicorp/terraform-provider-aws/issues/35585))
* resource/aws_backup_plan: Fix `InvalidParameterValueException: Invalid lifecycle. EBS Cold Tier is not yet supported` errors on resource Create in AWS GovCloud (US) ([#35560](https://github.com/hashicorp/terraform-provider-aws/issues/35560))
* resource/aws_cognito_user_group: Allow import of user groups with names containing `/` ([#35501](https://github.com/hashicorp/terraform-provider-aws/issues/35501))
* resource/aws_dms_event_subscription: Mark `source_ids` as Optional. This fixes a regression introduced in [v5.31.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#5310-december-15-2023) ([#35541](https://github.com/hashicorp/terraform-provider-aws/issues/35541))
* resource/aws_efs_file_system: Increase `lifecycle_policy` maximum item limit to 3 ([#35522](https://github.com/hashicorp/terraform-provider-aws/issues/35522))
* resource/aws_eks_access_entry: Retry IAM eventual consistency errors on create ([#35535](https://github.com/hashicorp/terraform-provider-aws/issues/35535))
* resource/aws_finspace_kx_cluster: Increase `command_line_arguments` max length restriction from 50 to 1024. ([#35581](https://github.com/hashicorp/terraform-provider-aws/issues/35581))

## 5.34.0 (January 26, 2024)

FEATURES:

* **New Resource:** `aws_rekognition_project` ([#35429](https://github.com/hashicorp/terraform-provider-aws/issues/35429))
* **New Resource:** `aws_route53domains_delegation_signer_record` ([#33596](https://github.com/hashicorp/terraform-provider-aws/issues/33596))

ENHANCEMENTS:

* data-source/aws_codecommit_repository: Add `kms_key_id` attribute ([#35095](https://github.com/hashicorp/terraform-provider-aws/issues/35095))
* data-source/aws_imagebuilder_components: Add support for `ThirdParty` `owner` value ([#35286](https://github.com/hashicorp/terraform-provider-aws/issues/35286))
* data-source/aws_imagebuilder_container_recipes: Add support for `ThirdParty` `owner` value ([#35286](https://github.com/hashicorp/terraform-provider-aws/issues/35286))
* data-source/aws_imagebuilder_image_recipes: Add support for `ThirdParty` `owner` value ([#35286](https://github.com/hashicorp/terraform-provider-aws/issues/35286))
* data-source/aws_ssm_patch_baseline: Add `json` attribute to facilitate use with S3 buckets ([#33402](https://github.com/hashicorp/terraform-provider-aws/issues/33402))
* resource/aws_accessanalyzer_analyzer: Add `configuration` configuration block ([#35310](https://github.com/hashicorp/terraform-provider-aws/issues/35310))
* resource/aws_appflow_flow: Add `flow_status` attribute ([#34948](https://github.com/hashicorp/terraform-provider-aws/issues/34948))
* resource/aws_codecommit_repository: Add `kms_key_id` argument ([#35095](https://github.com/hashicorp/terraform-provider-aws/issues/35095))
* resource/aws_codecommit_trigger: Add plan-time validation of `trigger.destination_arn` and `trigger.events` ([#35095](https://github.com/hashicorp/terraform-provider-aws/issues/35095))
* resource/aws_ecs_capacity_provider: Add `auto_scaling_group_provider.managed_draining` argument ([#35421](https://github.com/hashicorp/terraform-provider-aws/issues/35421))
* resource/aws_fis_experiment_template: Add support for `AutoScalingGroups`, `Buckets`, `ReplicationGroups`, `Tables` and `TransitGateways` to `action.*.target` ([#35300](https://github.com/hashicorp/terraform-provider-aws/issues/35300))
* resource/aws_fsx_openzfs_file_system: Add `skip_final_backup` argument ([#35320](https://github.com/hashicorp/terraform-provider-aws/issues/35320))
* resource/aws_network_interface_sg_attachment: Increase default timeouts to 3 minutes and allow them to be configured ([#35435](https://github.com/hashicorp/terraform-provider-aws/issues/35435))
* resource/aws_prometheus_scraper: Add `role_arn` attribute ([#35453](https://github.com/hashicorp/terraform-provider-aws/issues/35453))
* resource/aws_route53domains_registered_domain: Support resource import ([#33596](https://github.com/hashicorp/terraform-provider-aws/issues/33596))
* resource/aws_ssm_patch_baseline: Add `json` attribute to facilitate use with S3 buckets ([#33402](https://github.com/hashicorp/terraform-provider-aws/issues/33402))
* resource/aws_wafv2_web_acl: Add `challenge_config` argument ([#35367](https://github.com/hashicorp/terraform-provider-aws/issues/35367))

BUG FIXES:

* resource/aws_codebuild_project: Allow `build_batch_config` to be removed on Update ([#34121](https://github.com/hashicorp/terraform-provider-aws/issues/34121))
* resource/aws_eks_access_entry: Mark `kubernetes_groups` as Computed ([#35391](https://github.com/hashicorp/terraform-provider-aws/issues/35391))
* resource/aws_eks_access_entry: Mark `type` and `user_name` as Optional, allowing values to be configured ([#35391](https://github.com/hashicorp/terraform-provider-aws/issues/35391))
* resource/aws_grafana_license_association: Fix missing `workspace_id` attribute after import ([#35290](https://github.com/hashicorp/terraform-provider-aws/issues/35290))
* resource/aws_security_group_rule: Fix `UnsupportedOperation: The functionality you requested is not available in this region` errors on Read in certain partitions ([#33484](https://github.com/hashicorp/terraform-provider-aws/issues/33484))

## 5.33.0 (January 18, 2024)

FEATURES:

* **New Data Source:** `aws_eks_access_entry` ([#35037](https://github.com/hashicorp/terraform-provider-aws/issues/35037))
* **New Resource:** `aws_eks_access_entry` ([#35037](https://github.com/hashicorp/terraform-provider-aws/issues/35037))
* **New Resource:** `aws_eks_access_policy_association` ([#35037](https://github.com/hashicorp/terraform-provider-aws/issues/35037))
* **New Resource:** `aws_lexv2models_intent` ([#34891](https://github.com/hashicorp/terraform-provider-aws/issues/34891))

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `access_config` attribute ([#35037](https://github.com/hashicorp/terraform-provider-aws/issues/35037))
* data-source/aws_secretsmanager_secret: Add `created_date` and `last_changed_date` attributes ([#35117](https://github.com/hashicorp/terraform-provider-aws/issues/35117))
* data-source/aws_secretsmanager_secret_version: Add `created_date` attribute ([#35117](https://github.com/hashicorp/terraform-provider-aws/issues/35117))
* resource/aws_backup_plan: Add `rule.lifecycle.opt_in_to_archive_for_supported_resources` and `rule.copy_action.lifecycle.opt_in_to_archive_for_supported_resources` and arguments ([#34994](https://github.com/hashicorp/terraform-provider-aws/issues/34994))
* resource/aws_eks_cluster: Add `access_config` configuration block ([#35037](https://github.com/hashicorp/terraform-provider-aws/issues/35037))
* resource/aws_lakeformation_resource: Add `use_service_linked_role` argument ([#35284](https://github.com/hashicorp/terraform-provider-aws/issues/35284))
* resource/aws_secretsmanager_secret_rotation: Add `rotate_immediately` argument ([#35105](https://github.com/hashicorp/terraform-provider-aws/issues/35105))

BUG FIXES:

* resource/aws_datasync_task: Allow `schedule` to be removed successfully ([#35282](https://github.com/hashicorp/terraform-provider-aws/issues/35282))
* resource/aws_fis_experiment_template: Fix validation error when not using `target.resource_arns` or `target.resource_tag` attributes. ([#35254](https://github.com/hashicorp/terraform-provider-aws/issues/35254))
* resource/aws_lb_listener: Fix `ValidationError: Mutual Authentication mode passthrough does not support ignoring certificate expiry` errors when `mutual_authentication.mode` is set to `passthrough` ([#35289](https://github.com/hashicorp/terraform-provider-aws/issues/35289))
* resource/aws_secretsmanager_secret_version: Fix `InvalidParameterException: The parameter RemoveFromVersionId can't be empty. Staging label AWSCURRENT is currently attached to version ..., so you must explicitly reference that version in RemoveFromVersionId` errors when a secret is updated outside Terraform ([#19943](https://github.com/hashicorp/terraform-provider-aws/issues/19943))

## 5.32.1 (January 12, 2024)

BUG FIXES:

* data-source/aws_ecr_image: Fix error when `most_recent` is not also `latest` ([#35269](https://github.com/hashicorp/terraform-provider-aws/issues/35269))
* resource/aws_iot_ca_certificate: Change `registration_config.role_arn` from `TypeBool` to `TypeString`, fixing `Inappropriate value for attribute "role_arn": a bool is required` errors ([#35234](https://github.com/hashicorp/terraform-provider-aws/issues/35234))
* resource/aws_mq_broker: Fix `interface conversion: interface {} is *schema.Set, not []string` panic ([#35265](https://github.com/hashicorp/terraform-provider-aws/issues/35265))

## 5.32.0 (January 11, 2024)

FEATURES:

* **New Data Source:** `aws_mq_broker_engine_types` ([#34232](https://github.com/hashicorp/terraform-provider-aws/issues/34232))
* **New Data Source:** `aws_msk_bootstrap_brokers` ([#32484](https://github.com/hashicorp/terraform-provider-aws/issues/32484))
* **New Data Source:** `aws_verifiedpermissions_policy_store` ([#32204](https://github.com/hashicorp/terraform-provider-aws/issues/32204))
* **New Resource:** `aws_ebs_fast_snapshot_restore` ([#35211](https://github.com/hashicorp/terraform-provider-aws/issues/35211))
* **New Resource:** `aws_elasticache_serverless_cache` ([#34951](https://github.com/hashicorp/terraform-provider-aws/issues/34951))
* **New Resource:** `aws_imagebuilder_workflow` ([#35097](https://github.com/hashicorp/terraform-provider-aws/issues/35097))
* **New Resource:** `aws_kinesis_resource_policy` ([#35167](https://github.com/hashicorp/terraform-provider-aws/issues/35167))
* **New Resource:** `aws_prometheus_scraper` ([#34749](https://github.com/hashicorp/terraform-provider-aws/issues/34749))
* **New Resource:** `aws_securitylake_aws_log_source` ([#34974](https://github.com/hashicorp/terraform-provider-aws/issues/34974))
* **New Resource:** `aws_ssoadmin_application_access_scope` ([#34811](https://github.com/hashicorp/terraform-provider-aws/issues/34811))
* **New Resource:** `aws_verifiedpermissions_policy_store` ([#32204](https://github.com/hashicorp/terraform-provider-aws/issues/32204))
* **New Resource:** `aws_verifiedpermissions_policy_template` ([#32205](https://github.com/hashicorp/terraform-provider-aws/issues/32205))
* **New Resource:** `aws_verifiedpermissions_schema` ([#32204](https://github.com/hashicorp/terraform-provider-aws/issues/32204))

ENHANCEMENTS:

* data-source/aws_batch_compute_environment: Add `update_policy` attribute ([#34353](https://github.com/hashicorp/terraform-provider-aws/issues/34353))
* data-source/aws_ecr_image: Add `image_uri` attribute ([#24526](https://github.com/hashicorp/terraform-provider-aws/issues/24526))
* data-source/aws_efs_file_system: Add `lifecycle_policy.transition_to_archive` attribute ([#35096](https://github.com/hashicorp/terraform-provider-aws/issues/35096))
* data-source/aws_efs_file_system: Add `protection` attribute ([#35029](https://github.com/hashicorp/terraform-provider-aws/issues/35029))
* data-source/aws_elastic_beanstalk_hosted_zone: Add hosted zone ID for `il-central-1` AWS Region ([#35131](https://github.com/hashicorp/terraform-provider-aws/issues/35131))
* data-source/aws_elb_hosted_zone_id: Add hosted zone ID for `ca-west-1` AWS Region ([#35131](https://github.com/hashicorp/terraform-provider-aws/issues/35131))
* data-source/aws_fsx_ontap_file_system: Add `ha_pairs` and `throughput_capacity_per_ha_pair` attributes ([#34993](https://github.com/hashicorp/terraform-provider-aws/issues/34993))
* data-source/aws_glue_catalog_table: Add `region` attribute to `target_table` block. ([#34817](https://github.com/hashicorp/terraform-provider-aws/issues/34817))
* data-source/aws_lambda_function: Add `logging_config` attribute ([#35050](https://github.com/hashicorp/terraform-provider-aws/issues/35050))
* data-source/aws_lb_hosted_zone_id: Add hosted zone IDs for `ca-west-1` AWS Region ([#35131](https://github.com/hashicorp/terraform-provider-aws/issues/35131))
* data-source/aws_lb_target_group: Add `load_balancing_anomaly_mitigation` attribute ([#35083](https://github.com/hashicorp/terraform-provider-aws/issues/35083))
* data-source/aws_msk_configuration: Remove `name` length validation ([#34399](https://github.com/hashicorp/terraform-provider-aws/issues/34399))
* data-source/aws_networkfirewall_firewall_policy: Add `firewall_policy.tls_inspection_configuration_arn` attribute ([#35094](https://github.com/hashicorp/terraform-provider-aws/issues/35094))
* data-source/aws_prometheus_workspace: Add `kms_key_arn` attribute ([#35062](https://github.com/hashicorp/terraform-provider-aws/issues/35062))
* data-source/aws_route53_resolver_endpoint: Add `protocols` attribute ([#35098](https://github.com/hashicorp/terraform-provider-aws/issues/35098))
* data-source/aws_route53_resolver_endpoint: Add `resolver_endpoint_type` attribute ([#34798](https://github.com/hashicorp/terraform-provider-aws/issues/34798))
* data-source/aws_s3_bucket: Add hosted zone ID for `ca-west-1` AWS Region ([#35131](https://github.com/hashicorp/terraform-provider-aws/issues/35131))
* provider: Support `ca-west-1` as a valid AWS Region ([#35131](https://github.com/hashicorp/terraform-provider-aws/issues/35131))
* resource/aws_appflow_flow: Add `destination_connector_properties.s3.s3_output_format_config.target_file_size` argument ([#35215](https://github.com/hashicorp/terraform-provider-aws/issues/35215))
* resource/aws_appstream_fleet: Increase `idle_disconnect_timeout_in_seconds` max value for validation to 360000 ([#35173](https://github.com/hashicorp/terraform-provider-aws/issues/35173))
* resource/aws_autoscaling_group: Add `instance_refresh.preferences.max_healthy_percentage` attribute ([#34929](https://github.com/hashicorp/terraform-provider-aws/issues/34929))
* resource/aws_autoscaling_group: Fix `ValidationError: The instance ... is not part of Auto Scaling group ...` errors on resource Delete when disabling scale-in protection for instances that are already fully terminated ([#35071](https://github.com/hashicorp/terraform-provider-aws/issues/35071))
* resource/aws_batch_compute_environment: Add `update_policy` parameter ([#34353](https://github.com/hashicorp/terraform-provider-aws/issues/34353))
* resource/aws_batch_job_definition: Add `scheduling_priority` argument and `arn_prefix` attribute ([#34997](https://github.com/hashicorp/terraform-provider-aws/issues/34997))
* resource/aws_cloud9_environment_ec2: Add `amazonlinux-2023-x86_64` and `resolve:ssm:/aws/service/cloud9/amis/amazonlinux-2023-x86_64` as valid values for `image_id` ([#35020](https://github.com/hashicorp/terraform-provider-aws/issues/35020))
* resource/aws_codepipeline: Add `pipeline_type` argument and `variable` configuration block ([#34841](https://github.com/hashicorp/terraform-provider-aws/issues/34841))
* resource/aws_dms_replication_task: Allow `cdc_start_time` to use [RFC3339](https://www.rfc-editor.org/rfc/rfc3339) formatted dates in addition to UNIX timestamps ([#31917](https://github.com/hashicorp/terraform-provider-aws/issues/31917))
* resource/aws_dms_replication_task: Remove [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) from `replication_instance_arn`, allowing in-place migration between DMS instances ([#30721](https://github.com/hashicorp/terraform-provider-aws/issues/30721))
* resource/aws_efs_file_system: Add `lifecycle_policy.transition_to_archive` argument ([#35096](https://github.com/hashicorp/terraform-provider-aws/issues/35096))
* resource/aws_efs_file_system: Add `protection` configuration block ([#35029](https://github.com/hashicorp/terraform-provider-aws/issues/35029))
* resource/aws_efs_replication_configuration: Increase Create timeout to 20 minutes ([#34955](https://github.com/hashicorp/terraform-provider-aws/issues/34955))
* resource/aws_efs_replication_configuration: Mark `destination.file_system_id` as Optional, enabling [EFS replication fallback](https://docs.aws.amazon.com/efs/latest/ug/replication-use-cases.html#replicate-existing-destination) ([#34955](https://github.com/hashicorp/terraform-provider-aws/issues/34955))
* resource/aws_finspace_kx_dataview: Increase default create, update, and delete timeouts to 4 hours ([#35207](https://github.com/hashicorp/terraform-provider-aws/issues/35207))
* resource/aws_finspace_kx_scaling_group: Increase default create, delete timeouts to 4 hours ([#35206](https://github.com/hashicorp/terraform-provider-aws/issues/35206))
* resource/aws_fsx_lustre_file_system: Allow `per_unit_storage_throughput` to be updated in-place ([#34932](https://github.com/hashicorp/terraform-provider-aws/issues/34932))
* resource/aws_fsx_ontap_file_system: Add `ha_pairs` and `throughput_capacity_per_ha_pair` arguments ([#34993](https://github.com/hashicorp/terraform-provider-aws/issues/34993))
* resource/aws_fsx_ontap_file_system: Increase maximum value of `disk_iops_configuration.iops` to `2400000` ([#34993](https://github.com/hashicorp/terraform-provider-aws/issues/34993))
* resource/aws_fsx_ontap_file_system: `throughput_capacity` is Optional ([#34993](https://github.com/hashicorp/terraform-provider-aws/issues/34993))
* resource/aws_glue_catalog_table: Add `region` attribute to `target_table` block. ([#34817](https://github.com/hashicorp/terraform-provider-aws/issues/34817))
* resource/aws_glue_classifier: Add `csv_classifier.serde` argument ([#34251](https://github.com/hashicorp/terraform-provider-aws/issues/34251))
* resource/aws_kinesis_firehose_delivery_stream: Add `opensearch_configuration.document_id_options` configuration block ([#35137](https://github.com/hashicorp/terraform-provider-aws/issues/35137))
* resource/aws_kinesis_firehose_delivery_stream: Add `splunk_configuration.buffering_interval` and `splunk_configuration.buffering_size` arguments ([#35137](https://github.com/hashicorp/terraform-provider-aws/issues/35137))
* resource/aws_kinesis_firehose_delivery_stream: Adjust `elasticsearch_configuration.buffering_interval`, `http_endpoint_configuration.buffering_interval`, `opensearch_configuration.buffering_interval`, `opensearchserverless_configuration.buffering_interval`, `redshift_configuration.s3_backup_configuration.buffering_interval`,`extended_s3_configuration.s3_backup_configuration.buffering_interval`, `elasticsearch_configuration.s3_configuration.buffering_interval`, `http_endpoint_configuration.s3_configuration.buffering_interval`, `opensearch_configuration.s3_configuration.buffering_interval`, `opensearchserverless_configuration.s3_configuration.buffering_interval`, `redshift_configuration.s3_configuration.buffering_interval` and `splunk_configuration.s3_configuration.buffering_interval` minimum values to `0` to support zero buffering ([#35137](https://github.com/hashicorp/terraform-provider-aws/issues/35137))
* resource/aws_kms_key: Add `xks_key_id` attribute ([#31216](https://github.com/hashicorp/terraform-provider-aws/issues/31216))
* resource/aws_lambda_function: Add `logging_config` configuration block in support of [advanced logging controls](https://docs.aws.amazon.com/lambda/latest/dg/monitoring-cloudwatchlogs.html#monitoring-cloudwatchlogs-advanced) ([#35050](https://github.com/hashicorp/terraform-provider-aws/issues/35050))
* resource/aws_lambda_function: Add support for `python3.12` `runtime` value ([#35049](https://github.com/hashicorp/terraform-provider-aws/issues/35049))
* resource/aws_lambda_layer_version: Add support for `python3.12` `compatible_runtimes` value ([#35049](https://github.com/hashicorp/terraform-provider-aws/issues/35049))
* resource/aws_lb_target_group: Add `load_balancing_anomaly_mitigation` argument ([#35083](https://github.com/hashicorp/terraform-provider-aws/issues/35083))
* resource/aws_lb_target_group: Add `weighted_random` as a valid value for `load_balancing_algorithm_type` ([#35083](https://github.com/hashicorp/terraform-provider-aws/issues/35083))
* resource/aws_neptune_cluster: Add `storage_type` argument ([#34985](https://github.com/hashicorp/terraform-provider-aws/issues/34985))
* resource/aws_neptune_cluster_instance: Add `storage_type` attribute ([#34985](https://github.com/hashicorp/terraform-provider-aws/issues/34985))
* resource/aws_networkfirewall_firewall: Add configurable timeouts ([#34918](https://github.com/hashicorp/terraform-provider-aws/issues/34918))
* resource/aws_networkfirewall_firewall_policy: Add `firewall_policy.tls_inspection_configuration_arn` argument ([#35094](https://github.com/hashicorp/terraform-provider-aws/issues/35094))
* resource/aws_prometheus_workspace: Add `kms_key_arn` argument, enabling encryption at-rest using AWS KMS Customer Managed Keys (CMK) ([#35062](https://github.com/hashicorp/terraform-provider-aws/issues/35062))
* resource/aws_redshiftserverless_workgroup: Add `port` argument ([#34925](https://github.com/hashicorp/terraform-provider-aws/issues/34925))
* resource/aws_route53_resolver_endpoint: Add `protocols` argument ([#35098](https://github.com/hashicorp/terraform-provider-aws/issues/35098))
* resource/aws_route53_resolver_endpoint: Add `resolver_endpoint_type` argument ([#34798](https://github.com/hashicorp/terraform-provider-aws/issues/34798))
* resource/aws_s3_bucket: Modify resource Read to support third-party S3 API implementations. Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#35035](https://github.com/hashicorp/terraform-provider-aws/issues/35035))
* resource/aws_s3_bucket: Modify server-side encryption configuration error handling, enabling support for NetApp StorageGRID ([#34890](https://github.com/hashicorp/terraform-provider-aws/issues/34890))
* resource/aws_transfer_server: Add `TransferSecurityPolicy-PQ-SSH-Experimental-2023-04` and `TransferSecurityPolicy-PQ-SSH-FIPS-Experimental-2023-04` as valid values for `security_policy_name` ([#35129](https://github.com/hashicorp/terraform-provider-aws/issues/35129))
* resource/aws_verifiedaccess_endpoint: Add `policy_document` argument ([#34264](https://github.com/hashicorp/terraform-provider-aws/issues/34264))

BUG FIXES:

* data-source/aws_lb_target_group: Change `deregistration_delay` from `TypeInt` to `TypeString` ([#31436](https://github.com/hashicorp/terraform-provider-aws/issues/31436))
* data-source/aws_s3_bucket_object: Remove any leading `./` from `key` to maintain AWS SDK for Go v1 (pre-v5.17.0) compatibility ([#35223](https://github.com/hashicorp/terraform-provider-aws/issues/35223))
* data-source/aws_s3_object: Remove any leading `./` from `key` to maintain AWS SDK for Go v1 (pre-v5.17.0) compatibility ([#35223](https://github.com/hashicorp/terraform-provider-aws/issues/35223))
* resource/aws_cloud9_environment_ec2: `image_id` is Required ([#35020](https://github.com/hashicorp/terraform-provider-aws/issues/35020))
* resource/aws_codebuild_project: Prevent erroneous diffs on `build_timeout` and `queued_timeout` for Lambda compute types ([#35043](https://github.com/hashicorp/terraform-provider-aws/issues/35043))
* resource/aws_datasync_agent: Fix import of agents created with `activation_key` by removing requirement for one of `ip_address` or `activation_key` to be set ([#35150](https://github.com/hashicorp/terraform-provider-aws/issues/35150))
* resource/aws_dms_replication_config: Prevent erroneous diffs on `replication_settings` ([#34356](https://github.com/hashicorp/terraform-provider-aws/issues/34356))
* resource/aws_dms_replication_task: Prevent erroneous diffs on `replication_task_settings` ([#34356](https://github.com/hashicorp/terraform-provider-aws/issues/34356))
* resource/aws_dynamodb_table: Fix error when waiting for snapshot to be created ([#34848](https://github.com/hashicorp/terraform-provider-aws/issues/34848))
* resource/aws_finspace_kx_dataview: Properly set `arn` attribute on read, resolving persistent differences when `tags` are configured ([#34998](https://github.com/hashicorp/terraform-provider-aws/issues/34998))
* resource/aws_glue_catalog_database: Properly handle out-of-band resource deletion ([#35195](https://github.com/hashicorp/terraform-provider-aws/issues/35195))
* resource/aws_iot_indexing_configuration: Correct plan-time validation of `thing_indexing_configuration.filter.named_shadow_names` ([#35225](https://github.com/hashicorp/terraform-provider-aws/issues/35225))
* resource/aws_kinesis_firehose_delivery_stream: Fix `InvalidArgumentException: Both BufferSizeInMBs and BufferIntervalInSeconds are required to configure buffering for lambda processor` errors on resource Update ([#26964](https://github.com/hashicorp/terraform-provider-aws/issues/26964))
* resource/aws_kinesis_firehose_delivery_stream: Fix perpetual `extended_s3_configuration.processing_configuration.processors.parameters` diffs when processor type is `Lambda` ([#35137](https://github.com/hashicorp/terraform-provider-aws/issues/35137))
* resource/aws_lambda_function: Ensure lambda does not get deployed if `source_code_hash` does not change. ([#29921](https://github.com/hashicorp/terraform-provider-aws/issues/29921))
* resource/aws_lb: Fix `ValidationError: Attributes cannot be empty` errors ([#35228](https://github.com/hashicorp/terraform-provider-aws/issues/35228))
* resource/aws_lb_target_group: Fix diff on `stickiness.cookie_name` when `stickiness.type` is `lb_cookie` ([#31436](https://github.com/hashicorp/terraform-provider-aws/issues/31436))
* resource/aws_memorydb_cluster: Treat `snapshotting` status as pending when creating cluster ([#31077](https://github.com/hashicorp/terraform-provider-aws/issues/31077))
* resource/aws_ram_principal_association: Fix `reading RAM Resource Share (...) Principal Association (...): couldn't find resource (21 retries)` errors when a high number of principals are associated with a resource share ([#34738](https://github.com/hashicorp/terraform-provider-aws/issues/34738))
* resource/aws_s3_bucket_object: Remove any leading `./` from `key` to maintain AWS SDK for Go v1 (pre-v5.17.0) compatibility ([#35223](https://github.com/hashicorp/terraform-provider-aws/issues/35223))
* resource/aws_s3_object: Remove any leading `./` from `key` to maintain AWS SDK for Go v1 (pre-v5.17.0) compatibility ([#35223](https://github.com/hashicorp/terraform-provider-aws/issues/35223))
* resource/aws_s3_object_copy: Remove any leading `./` from `key` to maintain AWS SDK for Go v1 (pre-v5.17.0) compatibility ([#35223](https://github.com/hashicorp/terraform-provider-aws/issues/35223))
* resource/aws_secretsmanager_secret_rotation: No longer ignores changes to `rotation_rules.automatically_after_days` when `rotation_rules.schedule_expression` is set. ([#35024](https://github.com/hashicorp/terraform-provider-aws/issues/35024))
* resource/aws_ses_configuration_set: Fix `tracking_options` being omitted from state and resulting in persistent diff ([#35056](https://github.com/hashicorp/terraform-provider-aws/issues/35056))
* resource/aws_ssoadmin_application: Fix `portal_options.sign_in_options.application_url` triggering `ValidationError` when unset ([#34967](https://github.com/hashicorp/terraform-provider-aws/issues/34967))

## 5.31.0 (December 15, 2023)

FEATURES:

* **New Data Source:** `aws_polly_voices` ([#34916](https://github.com/hashicorp/terraform-provider-aws/issues/34916))
* **New Data Source:** `aws_ssoadmin_application_assignments` ([#34796](https://github.com/hashicorp/terraform-provider-aws/issues/34796))
* **New Data Source:** `aws_ssoadmin_principal_application_assignments` ([#34815](https://github.com/hashicorp/terraform-provider-aws/issues/34815))
* **New Resource:** `aws_finspace_kx_dataview` ([#34828](https://github.com/hashicorp/terraform-provider-aws/issues/34828))
* **New Resource:** `aws_finspace_kx_scaling_group` ([#34832](https://github.com/hashicorp/terraform-provider-aws/issues/34832))
* **New Resource:** `aws_finspace_kx_volume` ([#34833](https://github.com/hashicorp/terraform-provider-aws/issues/34833))
* **New Resource:** `aws_ssoadmin_trusted_token_issuer` ([#34839](https://github.com/hashicorp/terraform-provider-aws/issues/34839))

ENHANCEMENTS:

* data-source/aws_cloudwatch_log_group: Add `log_group_class` attribute ([#34812](https://github.com/hashicorp/terraform-provider-aws/issues/34812))
* data-source/aws_dms_endpoint: Add `postgres_settings` attribute ([#34724](https://github.com/hashicorp/terraform-provider-aws/issues/34724))
* data-source/aws_lb: Add `connection_logs` attribute ([#34864](https://github.com/hashicorp/terraform-provider-aws/issues/34864))
* data-source/aws_lb: Add `dns_record_client_routing_policy` attribute ([#34135](https://github.com/hashicorp/terraform-provider-aws/issues/34135))
* data-source/aws_opensearchserverless_collection: Add `standby_replicas` attribute ([#34677](https://github.com/hashicorp/terraform-provider-aws/issues/34677))
* resource/aws_db_instance: Add support for IBM Db2 databases ([#34834](https://github.com/hashicorp/terraform-provider-aws/issues/34834))
* resource/aws_dms_endpoint: Add `elasticsearch_settings.use_new_mapping_type` argument ([#29470](https://github.com/hashicorp/terraform-provider-aws/issues/29470))
* resource/aws_dms_endpoint: Add `postgres_settings` configuration block ([#34724](https://github.com/hashicorp/terraform-provider-aws/issues/34724))
* resource/aws_finspace_kx_cluster: Add `database.dataview_name`, `scaling_group_configuration`, and `tickerplant_log_configuration` arguments. ([#34831](https://github.com/hashicorp/terraform-provider-aws/issues/34831))
* resource/aws_finspace_kx_cluster: The `capacity_configuration` argument is now optional. ([#34831](https://github.com/hashicorp/terraform-provider-aws/issues/34831))
* resource/aws_lb: Add `connection_logs` configuration block ([#34864](https://github.com/hashicorp/terraform-provider-aws/issues/34864))
* resource/aws_lb: Add plan-time validation that exactly one of either `subnets` or `subnet_mapping` is configured ([#33205](https://github.com/hashicorp/terraform-provider-aws/issues/33205))
* resource/aws_lb: Allow the number of `subnet_mapping`s for Application Load Balancers to be changed without recreating the resource ([#33205](https://github.com/hashicorp/terraform-provider-aws/issues/33205))
* resource/aws_lb: Allow the number of `subnet_mapping`s for Network Load Balancers to be increased without recreating the resource ([#33205](https://github.com/hashicorp/terraform-provider-aws/issues/33205))
* resource/aws_lb: Allow the number of `subnets` for Network Load Balancers to be increased without recreating the resource ([#33205](https://github.com/hashicorp/terraform-provider-aws/issues/33205))
* resource/aws_opensearchserverless_collection: Add `standby_replicas` attribute ([#34677](https://github.com/hashicorp/terraform-provider-aws/issues/34677))

BUG FIXES:

* data-source/aws_ecr_pull_through_cache_rule: Fix plan time validation for `ecr_repository_prefix` ([#34716](https://github.com/hashicorp/terraform-provider-aws/issues/34716))
* provider: Always use the S3 regional endpoint in `us-east-1` for S3 directory bucket operations. This fixes `no such host` errors ([#34893](https://github.com/hashicorp/terraform-provider-aws/issues/34893))
* resource/aws_appmesh_virtual_node: Remove limit of 50 `backend`s per virtual node ([#34774](https://github.com/hashicorp/terraform-provider-aws/issues/34774))
* resource/aws_cloudwatch_log_group: Fix `invalid new value for .skip_destroy: was cty.False, but now null` errors ([#30354](https://github.com/hashicorp/terraform-provider-aws/issues/30354))
* resource/aws_cloudwatch_log_group: Remove default value (`STANDARD`) for `log_group_class` argument and mark as Computed. This fixes `InvalidParameterException: Only Standard log class is supported` errors in AWS Regions other than AWS Commercial ([#34812](https://github.com/hashicorp/terraform-provider-aws/issues/34812))
* resource/aws_db_instance: Fix error where Terraform loses track of resource if Blue/Green Deployment is applied outside of Terraform ([#34728](https://github.com/hashicorp/terraform-provider-aws/issues/34728))
* resource/aws_dms_event_subscription: `source_ids` and `source_type` are Required ([#33731](https://github.com/hashicorp/terraform-provider-aws/issues/33731))
* resource/aws_ecr_pull_through_cache_rule: Fix plan time validation for `ecr_repository_prefix` ([#34716](https://github.com/hashicorp/terraform-provider-aws/issues/34716))
* resource/aws_lb: Correct in-place update of `security_groups` for Network Load Balancers when the new value is Computed ([#33205](https://github.com/hashicorp/terraform-provider-aws/issues/33205))
* resource/aws_lb: Fix `InvalidConfigurationRequest: Load balancer attribute key 'dns_record.client_routing_policy' is not supported on load balancers with type 'network'` errors on resource Create in AWS GovCloud (US) ([#34135](https://github.com/hashicorp/terraform-provider-aws/issues/34135))
* resource/aws_medialive_channel: Fixed errors related to setting the `failover_condition` argument ([#33410](https://github.com/hashicorp/terraform-provider-aws/issues/33410))
* resource/aws_securitylake_data_lake: Fix `reflect.Set: value of type basetypes.StringValue is not assignable to type types.ARN` panic when importing resources with `nil` ARN fields ([#34820](https://github.com/hashicorp/terraform-provider-aws/issues/34820))
* resource/aws_vpc: Increase IPAM pool allocation deletion timeout from 20 minutes to 35 minutes ([#34859](https://github.com/hashicorp/terraform-provider-aws/issues/34859))

## 5.30.0 (December  7, 2023)

FEATURES:

* **New Data Source:** `aws_codeguruprofiler_profiling_group` ([#34672](https://github.com/hashicorp/terraform-provider-aws/issues/34672))
* **New Data Source:** `aws_ecr_repositories` ([#34446](https://github.com/hashicorp/terraform-provider-aws/issues/34446))
* **New Data Source:** `aws_lb_trust_store` ([#34584](https://github.com/hashicorp/terraform-provider-aws/issues/34584))
* **New Data Source:** `aws_ssoadmin_application` ([#34773](https://github.com/hashicorp/terraform-provider-aws/issues/34773))
* **New Data Source:** `aws_ssoadmin_application_providers` ([#34670](https://github.com/hashicorp/terraform-provider-aws/issues/34670))
* **New Resource:** `aws_codeguruprofiler_profiling_group` ([#34672](https://github.com/hashicorp/terraform-provider-aws/issues/34672))
* **New Resource:** `aws_customerprofiles_domain` ([#34622](https://github.com/hashicorp/terraform-provider-aws/issues/34622))
* **New Resource:** `aws_customerprofiles_profile` ([#34622](https://github.com/hashicorp/terraform-provider-aws/issues/34622))
* **New Resource:** `aws_lb_trust_store` ([#34584](https://github.com/hashicorp/terraform-provider-aws/issues/34584))
* **New Resource:** `aws_lb_trust_store_revocation` ([#34584](https://github.com/hashicorp/terraform-provider-aws/issues/34584))
* **New Resource:** `aws_securitylake_data_lake` ([#34521](https://github.com/hashicorp/terraform-provider-aws/issues/34521))
* **New Resource:** `aws_ssoadmin_application` ([#34723](https://github.com/hashicorp/terraform-provider-aws/issues/34723))
* **New Resource:** `aws_ssoadmin_application_assignment` ([#34741](https://github.com/hashicorp/terraform-provider-aws/issues/34741))
* **New Resource:** `aws_ssoadmin_application_assignment_configuration` ([#34752](https://github.com/hashicorp/terraform-provider-aws/issues/34752))

ENHANCEMENTS:

* data-source/aws_appconfig_configuration_profile: Add `kms_key_identifier` attribute ([#34725](https://github.com/hashicorp/terraform-provider-aws/issues/34725))
* data-source/aws_lb: Add `enforce_security_group_inbound_rules_on_private_link_traffic` attribute ([#33767](https://github.com/hashicorp/terraform-provider-aws/issues/33767))
* data-source/aws_lb_listener: Add `mutual_authentication` attribute ([#34584](https://github.com/hashicorp/terraform-provider-aws/issues/34584))
* resource/aws_appconfig_configuration_profile: Add `kms_key_identifier` attribute ([#34725](https://github.com/hashicorp/terraform-provider-aws/issues/34725))
* resource/aws_appconfig_deployment: Add `kms_key_identifier` attribute ([#34739](https://github.com/hashicorp/terraform-provider-aws/issues/34739))
* resource/aws_cloudwatch_log_group: Add `log_group_class` argument ([#34679](https://github.com/hashicorp/terraform-provider-aws/issues/34679))
* resource/aws_lb: Add `enforce_security_group_inbound_rules_on_private_link_traffic` argument ([#33767](https://github.com/hashicorp/terraform-provider-aws/issues/33767))
* resource/aws_lb_listener: Add `mutual_authentication` configuration block ([#34584](https://github.com/hashicorp/terraform-provider-aws/issues/34584))
* resource/aws_s3_bucket: Fix `stack overflow` fatal errors on resource Delete when `force_destroy` is `true` and the bucket contains delete markers ([#34712](https://github.com/hashicorp/terraform-provider-aws/issues/34712))
* resource/aws_sagemaker_app: Add `resource_spec.sagemaker_image_version_alias` argument ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_sagemaker_app_image_config: Add `jupyter_lab_image_config` configuration block ([#34696](https://github.com/hashicorp/terraform-provider-aws/issues/34696))
* resource/aws_sagemaker_domain: Add `default_user_settings.code_editor_app_settings`, `default_user_settings.custom_file_system_config`, `default_user_settings.custom_posix_user_config`, `default_user_settings.default_landing_uri`, `default_user_settings.jupyter_lab_app_settings`, `default_user_settings.space_storage_settings`, `default_user_settings.studio_web_portal` arguments ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_sagemaker_domain: Add `sagemaker_image_version_alias` argument under all `default_resource_spec` blocks ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_sagemaker_domain: Add `single_sign_on_application_arn` attribute ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_sagemaker_space: Add `sagemaker_image_version_alias` argument under all `default_resource_spec` blocks ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_sagemaker_space: Add `space_display_name` argument ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_sagemaker_space: Add `url` attribute ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_sagemaker_user_profile: Add `sagemaker_image_version_alias` argument under all `default_resource_spec` blocks ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_sagemaker_user_profile: Add `user_settings.code_editor_app_settings`, `user_settings.custom_file_system_config`, `user_settings.custom_posix_user_config`, `user_settings.default_landing_uri`, `user_settings.jupyter_lab_app_settings`, `user_settings.space_storage_settings`, `user_settings.studio_web_portal` arguments ([#34729](https://github.com/hashicorp/terraform-provider-aws/issues/34729))
* resource/aws_transfer_server: Add support for `TransferSecurityPolicy-FIPS-2023-05` `security_policy_name` value ([#34709](https://github.com/hashicorp/terraform-provider-aws/issues/34709))

BUG FIXES:

* resource/aws_ami: Correctly sets `deprecation_time` on creation and update due to eventual consistency ([#34691](https://github.com/hashicorp/terraform-provider-aws/issues/34691))
* resource/aws_ami: Correctly sets `description` on update due to eventual consistency ([#34691](https://github.com/hashicorp/terraform-provider-aws/issues/34691))
* resource/aws_ami: Now allows removing `deprecation_time` ([#34691](https://github.com/hashicorp/terraform-provider-aws/issues/34691))
* resource/aws_appflow_flow: Fix perpetual diff on `destination_flow_config` ([#34770](https://github.com/hashicorp/terraform-provider-aws/issues/34770))
* resource/aws_backup_vault_policy: Fix eventual consistency error when waiting for IAM ([#34671](https://github.com/hashicorp/terraform-provider-aws/issues/34671))
* resource/aws_eks_pod_identity_association: Retry IAM eventual consistency errors on create and update ([#34717](https://github.com/hashicorp/terraform-provider-aws/issues/34717))
* resource/aws_glue_connection: Fix crash while creating resource with empty `physical_connection_requirements` configuration block ([#34737](https://github.com/hashicorp/terraform-provider-aws/issues/34737))

## 5.29.0 (November 30, 2023)

FEATURES:

* **New Resource:** `aws_docdbelastic_cluster` ([#31033](https://github.com/hashicorp/terraform-provider-aws/issues/31033))
* **New Resource:** `aws_eks_pod_identity_association` ([#34566](https://github.com/hashicorp/terraform-provider-aws/issues/34566))

ENHANCEMENTS:

* resource/aws_docdb_cluster: Add `storage_type` argument ([#34637](https://github.com/hashicorp/terraform-provider-aws/issues/34637))
* resource/aws_neptune_parameter_group: Add `name_prefix` argument ([#34500](https://github.com/hashicorp/terraform-provider-aws/issues/34500))

BUG FIXES:

* resource/aws_networkmanager_attachment_accepter: Now revokes attachment on deletion for VPC Attachments ([#34547](https://github.com/hashicorp/terraform-provider-aws/issues/34547))
* resource/aws_networkmanager_vpc_attachment: Fixes error when modifying `options` fields while waiting for acceptance ([#34547](https://github.com/hashicorp/terraform-provider-aws/issues/34547))
* resource/aws_networkmanager_vpc_attachment: Fixes error where VPC Attachments waiting for acceptance could not be deleted ([#34547](https://github.com/hashicorp/terraform-provider-aws/issues/34547))
* resource/aws_s3_directory_bucket: Fix `NotImplemented: This bucket does not support Object Versioning` errors on resource Delete when `force_destroy` is `true` ([#34647](https://github.com/hashicorp/terraform-provider-aws/issues/34647))

## 5.28.0 (November 29, 2023)

FEATURES:

* **New Data Source:** `aws_s3_directory_buckets` ([#34612](https://github.com/hashicorp/terraform-provider-aws/issues/34612))
* **New Resource:** `aws_s3_directory_bucket` ([#34612](https://github.com/hashicorp/terraform-provider-aws/issues/34612))

ENHANCEMENTS:

* resource/aws_s3control_access_grants_instance: Add `identity_center_arn` argument and `identity_center_application_arn` attribute ([#34582](https://github.com/hashicorp/terraform-provider-aws/issues/34582))

BUG FIXES:

* resource/aws_elaticache_replication_group: Fix regression caused by the introduction of the `auth_token_update_strategy` argument with a default value ([#34600](https://github.com/hashicorp/terraform-provider-aws/issues/34600))

## 5.27.0 (November 27, 2023)

NOTES:

* provider: This release includes an update to the AWS SDK for Go v2 with breaking type changes to several services: `internetmonitor`, `ivschat`, `pipes`, and `s3`. These changes primarily affect how arguments with default values are serialized for outbound requests, changing scalar types to pointers. See [this AWS SDK for Go V2 issue](https://github.com/aws/aws-sdk-go-v2/issues/2162) for additional context. The corresponding provider changes should make this breakfix transparent to users, but as with any breaking change there is the potential for missed edge cases. If errors are observed in the impacted resources, please link to this dependency update pull request in the bug report ([#34476](https://github.com/hashicorp/terraform-provider-aws/issues/34476))

FEATURES:

* **New Data Source:** `aws_emr_supported_instance_types` ([#34481](https://github.com/hashicorp/terraform-provider-aws/issues/34481))
* **New Resource:** `aws_apprunner_default_auto_scaling_configuration_version` ([#34292](https://github.com/hashicorp/terraform-provider-aws/issues/34292))
* **New Resource:** `aws_lexv2models_bot_version` ([#33858](https://github.com/hashicorp/terraform-provider-aws/issues/33858))
* **New Resource:** `aws_s3control_access_grant` ([#34564](https://github.com/hashicorp/terraform-provider-aws/issues/34564))
* **New Resource:** `aws_s3control_access_grants_instance` ([#34564](https://github.com/hashicorp/terraform-provider-aws/issues/34564))
* **New Resource:** `aws_s3control_access_grants_instance_resource_policy` ([#34564](https://github.com/hashicorp/terraform-provider-aws/issues/34564))
* **New Resource:** `aws_s3control_access_grants_location` ([#34564](https://github.com/hashicorp/terraform-provider-aws/issues/34564))

ENHANCEMENTS:

* resource/aws_apprunner_auto_scaling_configuration_version: Add `has_associated_service` and `is_default` attributes ([#34292](https://github.com/hashicorp/terraform-provider-aws/issues/34292))
* resource/aws_apprunner_service: Add `network_configuration.ip_address_type` argument ([#34292](https://github.com/hashicorp/terraform-provider-aws/issues/34292))
* resource/aws_apprunner_service: Add `source_configuration.code_repository.source_directory` argument to support monorepos ([#34292](https://github.com/hashicorp/terraform-provider-aws/issues/34292))
* resource/aws_apprunner_service: Allow `health_check_configuration` to be updated in-place ([#34292](https://github.com/hashicorp/terraform-provider-aws/issues/34292))
* resource/aws_cloudwatch_event_rule: Add `state` parameter and deprecate `is_enabled` parameter ([#34510](https://github.com/hashicorp/terraform-provider-aws/issues/34510))
* resource/aws_elaticache_replication_group: Add `auth_token_update_strategy` argument ([#34460](https://github.com/hashicorp/terraform-provider-aws/issues/34460))
* resource/aws_lambda_function: Add support for `java21` `runtime` value ([#34476](https://github.com/hashicorp/terraform-provider-aws/issues/34476))
* resource/aws_lambda_function: Add support for `python3.12` `runtime` value ([#34533](https://github.com/hashicorp/terraform-provider-aws/issues/34533))
* resource/aws_lambda_layer_version: Add support for `java21` `compatible_runtimes` value ([#34476](https://github.com/hashicorp/terraform-provider-aws/issues/34476))
* resource/aws_lambda_layer_version: Add support for `python3.12` `compatible_runtimes` value ([#34533](https://github.com/hashicorp/terraform-provider-aws/issues/34533))
* resource/aws_s3_bucket_logging: Add `target_object_key_format` configuration block to support [automatic date-based partitioning](https://docs.aws.amazon.com/AmazonS3/latest/userguide/ServerLogs.html#server-access-logging-overview) ([#34504](https://github.com/hashicorp/terraform-provider-aws/issues/34504))

BUG FIXES:

* resource/aws_appflow_flow: Fix `InvalidParameter: 2 validation error(s) found` error when `destination_flow_config` or `task` is updated ([#34456](https://github.com/hashicorp/terraform-provider-aws/issues/34456))
* resource/aws_appflow_flow: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panic ([#34456](https://github.com/hashicorp/terraform-provider-aws/issues/34456))
* resource/aws_apprunner_service: Correctly set `service_url` for private services ([#34292](https://github.com/hashicorp/terraform-provider-aws/issues/34292))
* resource/aws_glue_trigger: Fix `ConcurrentModificationException: Workflow <workflowName> was modified while adding trigger <triggerName>` errors ([#34530](https://github.com/hashicorp/terraform-provider-aws/issues/34530))
* resource/aws_lb_target_group: Adds plan- and apply-time validation for invalid parameter combinations ([#34488](https://github.com/hashicorp/terraform-provider-aws/issues/34488))
* resource/aws_lexv2_bot_locale: Fix `voice_settings.engine` validation, value conversion errors ([#34532](https://github.com/hashicorp/terraform-provider-aws/issues/34532))
* resource/aws_lexv2models_bot: Properly send `type` argument on create and update when configured ([#34524](https://github.com/hashicorp/terraform-provider-aws/issues/34524))
* resource/aws_pipes_pipe: Fix error when zero value is sent to `source_parameters` on update ([#34487](https://github.com/hashicorp/terraform-provider-aws/issues/34487))

## 5.26.0 (November 16, 2023)

FEATURES:

* **New Data Source:** `aws_iot_registration_code` ([#15098](https://github.com/hashicorp/terraform-provider-aws/issues/15098))
* **New Resource:** `aws_bedrock_model_invocation_logging_configuration` ([#34303](https://github.com/hashicorp/terraform-provider-aws/issues/34303))
* **New Resource:** `aws_iot_billing_group` ([#31237](https://github.com/hashicorp/terraform-provider-aws/issues/31237))
* **New Resource:** `aws_iot_ca_certificate` ([#15098](https://github.com/hashicorp/terraform-provider-aws/issues/15098))
* **New Resource:** `aws_iot_event_configurations` ([#31237](https://github.com/hashicorp/terraform-provider-aws/issues/31237))

ENHANCEMENTS:

* data-source/aws_autoscaling_group: Add `instance_maintenance_policy` attribute ([#34430](https://github.com/hashicorp/terraform-provider-aws/issues/34430))
* provider: Adds `https_proxy` and `no_proxy` parameters. ([#34243](https://github.com/hashicorp/terraform-provider-aws/issues/34243))
* resource/aws_autoscaling_group: Add `instance_maintenance_policy` configuration block ([#34430](https://github.com/hashicorp/terraform-provider-aws/issues/34430))
* resource/aws_finspace_kx_cluster: Increase default create and update timeouts to 4 hours to allow for increased startup times with large volumes of cached data ([#34398](https://github.com/hashicorp/terraform-provider-aws/issues/34398))
* resource/aws_finspace_kx_environment: Increase default delete timeout to 75 minutes ([#34398](https://github.com/hashicorp/terraform-provider-aws/issues/34398))
* resource/aws_iam_group_policy_attachment: Add plan-time validation of `policy_arn` ([#34378](https://github.com/hashicorp/terraform-provider-aws/issues/34378))
* resource/aws_iam_policy_attachment: Add plan-time validation of `policy_arn` ([#34378](https://github.com/hashicorp/terraform-provider-aws/issues/34378))
* resource/aws_iam_role_policy_attachment: Add plan-time validation of `policy_arn` ([#34378](https://github.com/hashicorp/terraform-provider-aws/issues/34378))
* resource/aws_iam_user_policy_attachment: Add plan-time validation of `policy_arn` ([#34378](https://github.com/hashicorp/terraform-provider-aws/issues/34378))
* resource/aws_iot_ca_certificate: Add `ca_certificate_id` attribute ([#15098](https://github.com/hashicorp/terraform-provider-aws/issues/15098))
* resource/aws_iot_policy: Add configurable timeouts ([#34329](https://github.com/hashicorp/terraform-provider-aws/issues/34329))
* resource/aws_iot_policy: When updating the resource, delete the oldest non-default version of the policy if creating a new version would exceed the maximum number of versions (5) ([#34329](https://github.com/hashicorp/terraform-provider-aws/issues/34329))
* resource/aws_lambda_function: Add support for `nodejs20.x` and `provided.al2023` `runtime` values ([#34401](https://github.com/hashicorp/terraform-provider-aws/issues/34401))
* resource/aws_lambda_layer_version: Add support for `nodejs20.x` and `provided.al2023` `compatible_runtimes` values ([#34401](https://github.com/hashicorp/terraform-provider-aws/issues/34401))
* resource/aws_quicksight_analysis: Add `definition.sheets.visuals.kpi_visual.chart_configuration.kpi_options.sparkline` attribute ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_analysis: Add `definition.sheets.visuals.kpi_visual.chart_configuration.kpi_options.visual_layout_options` attribute ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_analysis: Add `number_display_format_configuration` and `percentage_display_format_configuration` to nested `numeric_format_configuration` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Add `definition.sheets.visuals.kpi_visual.chart_configuration.kpi_options.sparkline` attribute ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Add `definition.sheets.visuals.kpi_visual.chart_configuration.kpi_options.visual_layout_options` attribute ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Add `number_display_format_configuration` and `percentage_display_format_configuration` to nested `numeric_format_configuration` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_template: Add `definition.sheets.visuals.kpi_visual.chart_configuration.kpi_options.sparkline` attribute ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_template: Add `definition.sheets.visuals.kpi_visual.chart_configuration.kpi_options.visual_layout_options` attribute ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_template: Add `number_display_format_configuration` and `percentage_display_format_configuration` to nested `numeric_format_configuration` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_rds_cluster: Add `delete_automated_backups` argument ([#34309](https://github.com/hashicorp/terraform-provider-aws/issues/34309))

BUG FIXES:

* resource/aws_chime_voice_connector: Fix `read` error when resource is not created in `us-east-1` ([#34334](https://github.com/hashicorp/terraform-provider-aws/issues/34334))
* resource/aws_chime_voice_connector_group: Fix `read` error when resource is not created in `us-east-1` ([#34334](https://github.com/hashicorp/terraform-provider-aws/issues/34334))
* resource/aws_chime_voice_connector_logging: Fix `read` error when resource is not created in `us-east-1` ([#34334](https://github.com/hashicorp/terraform-provider-aws/issues/34334))
* resource/aws_chime_voice_connector_origination: Fix `read` error when resource is not created in `us-east-1` ([#34334](https://github.com/hashicorp/terraform-provider-aws/issues/34334))
* resource/aws_chime_voice_connector_termination: Fix `read` error when resource is not created in `us-east-1` ([#34334](https://github.com/hashicorp/terraform-provider-aws/issues/34334))
* resource/aws_chime_voice_connector_termination_credentials: Fix `read` error when resource is not created in `us-east-1` ([#34334](https://github.com/hashicorp/terraform-provider-aws/issues/34334))
* resource/aws_chimesdkmediapipelines_media_insights_pipeline_configuration: Fix eventual consistency error when resource is not created in `us-east-1` ([#34334](https://github.com/hashicorp/terraform-provider-aws/issues/34334))
* resource/aws_chimesdkvoice_sip_media_application: Fix eventual consistency errors when not using `us-east-1` ([#34426](https://github.com/hashicorp/terraform-provider-aws/issues/34426))
* resource/aws_chimesdkvoice_sip_rule: Fix eventual consistency errors when not using `us-east-1` ([#34426](https://github.com/hashicorp/terraform-provider-aws/issues/34426))
* resource/aws_elasticache_user: Fix `UserNotFound: ... is not available for tagging` errors on resource Read when there is a concurrent update to the user ([#34396](https://github.com/hashicorp/terraform-provider-aws/issues/34396))
* resource/aws_grafana_workspace_api_key: Change `key` to [`Sensitive`](https://developer.hashicorp.com/terraform/plugin/best-practices/sensitive-state#using-sensitive-flag-functionality) ([#34105](https://github.com/hashicorp/terraform-provider-aws/issues/34105))
* resource/aws_iam_group_policy_attachment: Retry `ConcurrentModificationException` errors on create and delete ([#34378](https://github.com/hashicorp/terraform-provider-aws/issues/34378))
* resource/aws_iam_policy_attachment: Retry `ConcurrentModificationException` errors on create and delete ([#34378](https://github.com/hashicorp/terraform-provider-aws/issues/34378))
* resource/aws_iam_role_policy_attachment: Retry `ConcurrentModificationException` errors on create and delete ([#34378](https://github.com/hashicorp/terraform-provider-aws/issues/34378))
* resource/aws_iam_user_policy_attachment: Retry `ConcurrentModificationException` errors on create and delete ([#34378](https://github.com/hashicorp/terraform-provider-aws/issues/34378))
* resource/aws_inspector2_delegated_admin_account: Fix `errors: *target must be interface or implement error` panic ([#34424](https://github.com/hashicorp/terraform-provider-aws/issues/34424))
* resource/aws_inspector2_enabler: Fix `interface conversion: interface {} is nil, not map[string]inspector2.AccountResourceStatus` panic ([#34424](https://github.com/hashicorp/terraform-provider-aws/issues/34424))
* resource/aws_iot_ca_certificate: Change `ca_pem` and `certificate_pem` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#15098](https://github.com/hashicorp/terraform-provider-aws/issues/15098))
* resource/aws_iot_policy: Retry `DeleteConflictException` errors on delete ([#34329](https://github.com/hashicorp/terraform-provider-aws/issues/34329))
* resource/aws_quicksight_analysis: Fix handling of the nested `number_scale`, `prefix`, and `suffix` integer arguments ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_analysis: Fix handling of the nested `rolling_date` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_analysis: Fix handling of the nested `select_all_options` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_analysis: Fix handling of the nested `visual_ids` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_analysis: Fixes to various optional blocks utilizing the shared column schema definition ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_analysis: Nested `column_index` and `row_index` arguments now properly handle zero values ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Fix handling of the nested `number_scale`, `prefix`, and `suffix` integer arguments ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Fix handling of the nested `rolling_date` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Fix handling of the nested `select_all_options` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Fix handling of the nested `visual_ids` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Fixes to various optional blocks utilizing the shared column schema definition ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_dashboard: Nested `column_index` and `row_index` arguments now properly handle zero values ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_data_set: Increase `permissions.actions` maximum item limit to 20, aligning with the AWS API limits ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_data_source: Set all parameters to update aws_quicksight_data_source ([#33061](https://github.com/hashicorp/terraform-provider-aws/issues/33061))
* resource/aws_quicksight_template: Fix handling of the nested `number_scale`, `prefix`, and `suffix` integer arguments ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_template: Fix handling of the nested `rolling_date` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_template: Fix handling of the nested `select_all_options` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_template: Fix handling of the nested `visual_ids` argument ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_template: Fixes to various optional blocks utilizing the shared column schema definition ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_quicksight_template: Nested `column_index` and `row_index` arguments now properly handle zero values ([#33931](https://github.com/hashicorp/terraform-provider-aws/issues/33931))
* resource/aws_sagemaker_user_profile: Change `default_user_settings.canvas_app_settings.identity_provider_oauth_settings` from TypeSet to TypeList, preventing `interface conversion: interface {} is *schema.Set, not []interface {}` panics ([#34418](https://github.com/hashicorp/terraform-provider-aws/issues/34418))
* resource/aws_synthetics_canary: Fix to properly suppress differences when `expression` is `rate(0 minutes)` ([#34084](https://github.com/hashicorp/terraform-provider-aws/issues/34084))
* resource/aws_vpn_connection: Fix `UnsupportedOperation: The tunnel inside ip version parameter is not currently supported in this region` error when creating connections in certain partitions and Regions ([#34420](https://github.com/hashicorp/terraform-provider-aws/issues/34420))

## 5.25.0 (November 10, 2023)

NOTES:

* resource/aws_cloudtrail: The resource's [import ID](https://developer.hashicorp.com/terraform/language/import#import-id) has changed from `name` to `arn` ([#30758](https://github.com/hashicorp/terraform-provider-aws/issues/30758))

FEATURES:

* **New Data Source:** `aws_apigatewayv2_vpc_link` ([#33974](https://github.com/hashicorp/terraform-provider-aws/issues/33974))
* **New Data Source:** `aws_athena_named_query` ([#24815](https://github.com/hashicorp/terraform-provider-aws/issues/24815))
* **New Data Source:** `aws_bedrock_foundation_model` ([#34148](https://github.com/hashicorp/terraform-provider-aws/issues/34148))
* **New Data Source:** `aws_bedrock_foundation_models` ([#34148](https://github.com/hashicorp/terraform-provider-aws/issues/34148))
* **New Resource:** `aws_athena_prepared_statement` ([#33417](https://github.com/hashicorp/terraform-provider-aws/issues/33417))
* **New Resource:** `aws_lexv2models_bot_locale` ([#33949](https://github.com/hashicorp/terraform-provider-aws/issues/33949))

ENHANCEMENTS:

* provider: Adds SSO API endpoint override parameter `endpoints.sso` ([#34302](https://github.com/hashicorp/terraform-provider-aws/issues/34302))
* resource/aws_appflow_connector_profile: Add `jwt_token` and `oauth2_grant_type` arguments to the `connector_profile_config.connector_profile_credentials.salesforce` block. ([#34248](https://github.com/hashicorp/terraform-provider-aws/issues/34248))
* resource/aws_autoscaling_group: Add plan-time validation of `initial_lifecycle_hook.default_result`, `initial_lifecycle_hook.heartbeat_timeout`, `initial_lifecycle_hook.lifecycle_transition`, `initial_lifecycle_hook.name`, `initial_lifecycle_hook.notification_target_arn` and `initial_lifecycle_hook.role_arn` ([#12145](https://github.com/hashicorp/terraform-provider-aws/issues/12145))
* resource/aws_autoscaling_lifecycle_hook: Add plan-time validation of `default_result`, `heartbeat_timeout`, `lifecycle_transition`, `name`, `notification_target_arn` and `role_arn` ([#12145](https://github.com/hashicorp/terraform-provider-aws/issues/12145))
* resource/aws_datasync_task: Add `task_report_config` argument ([#33861](https://github.com/hashicorp/terraform-provider-aws/issues/33861))
* resource/aws_db_instance: Add `postgres` as a valid `engine` value for blue/green deployments ([#34216](https://github.com/hashicorp/terraform-provider-aws/issues/34216))
* resource/aws_dms_endpoint: Add `pause_replication_tasks`, which when set to `true`, pauses associated running replication tasks, regardless if they are managed by Terraform, prior to modifying the endpoint (only tasks paused by the resource will be restarted after the modification completes) ([#34316](https://github.com/hashicorp/terraform-provider-aws/issues/34316))
* resource/aws_eks_cluster: Allow `vpc_config.security_group_ids` and `vpc_config.subnet_ids` to be updated in-place ([#32409](https://github.com/hashicorp/terraform-provider-aws/issues/32409))
* resource/aws_inspector2_organization_configuration: Add `lambda_code` argument to the `auto_enable` configuration block ([#34261](https://github.com/hashicorp/terraform-provider-aws/issues/34261))
* resource/aws_route53_record: Allow import of records with an empty record name. ([#34212](https://github.com/hashicorp/terraform-provider-aws/issues/34212))
* resource/aws_sagemaker_domain: Add `default_user_settings.canvas_app_settings.direct_deploy_settings`, `default_user_settings.canvas_app_settings.identity_provider_oauth_settings` and `default_user_settings.canvas_app_settings.kendra_settings` arguments ([#34265](https://github.com/hashicorp/terraform-provider-aws/issues/34265))
* resource/aws_sagemaker_domain: Change `default_space_settings.kernel_gateway_app_settings.custom_image`, `default_user_settings.kernel_gateway_app_settings.custom_image` and `default_user_settings.r_session_app_settings.custom_image` `MaxItems` from `30` to `200` ([#34265](https://github.com/hashicorp/terraform-provider-aws/issues/34265))
* resource/aws_sagemaker_feature_group: Add `offline_store_config.s3_storage_config.resolved_output_s3_uri`, `online_store_config.storage_type` and `online_store_config.ttl_duration` arguments ([#34283](https://github.com/hashicorp/terraform-provider-aws/issues/34283))
* resource/aws_sagemaker_feature_group: Allow `online_store_config.ttl_duration` to be updated in-place ([#34283](https://github.com/hashicorp/terraform-provider-aws/issues/34283))
* resource/aws_sagemaker_model: Add `container.model_data_source` and `primary_container.model_data_source` configuration blocks ([#34158](https://github.com/hashicorp/terraform-provider-aws/issues/34158))
* resource/aws_sagemaker_space: Change `space_settings.kernel_gateway_app_settings.custom_image` `MaxItems` from `30` to `200` ([#34265](https://github.com/hashicorp/terraform-provider-aws/issues/34265))
* resource/aws_sagemaker_user_profile: Add `default_user_settings.canvas_app_settings.direct_deploy_settings`, `default_user_settings.canvas_app_settings.identity_provider_oauth_settings` and `default_user_settings.canvas_app_settings.kendra_settings` arguments ([#34265](https://github.com/hashicorp/terraform-provider-aws/issues/34265))
* resource/aws_sns_topic: Add `archive_policy` argument and `beginning_archive_time` attribute to support [message archiving](https://docs.aws.amazon.com/sns/latest/dg/fifo-message-archiving-replay.html) ([#34252](https://github.com/hashicorp/terraform-provider-aws/issues/34252))
* resource/aws_sns_topic: Add `replay_policy` argument ([#34252](https://github.com/hashicorp/terraform-provider-aws/issues/34252))

BUG FIXES:

* provider: Fix `Value Conversion Error` panic for certain resources when `null` tag values are specified ([#34319](https://github.com/hashicorp/terraform-provider-aws/issues/34319))
* provider: Fixes parsing error in AWS shared config files with extra whitespace ([#34300](https://github.com/hashicorp/terraform-provider-aws/issues/34300))
* provider: Fixes poor performance when parsing AWS shared config files ([#34300](https://github.com/hashicorp/terraform-provider-aws/issues/34300))
* resource/aws_autoscaling_group: Change all `initial_lifecycle_hook` configuration block attributes to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#34260](https://github.com/hashicorp/terraform-provider-aws/issues/34260))
* resource/aws_cloudtrail: Change the `id` attribute from the trail's name to its ARN to support [organization trails](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/creating-trail-organization.html) ([#30758](https://github.com/hashicorp/terraform-provider-aws/issues/30758))
* resource/aws_cloudwatch_event_rule: Increase `event_pattern` max length for validation to 4096 ([#34270](https://github.com/hashicorp/terraform-provider-aws/issues/34270))
* resource/aws_sagemaker_domain: Fix updating `default_space_settings.r_studio_server_pro_app_settings.access_status` from `ENABLED` to `DISABLED` ([#34265](https://github.com/hashicorp/terraform-provider-aws/issues/34265))

## 5.24.0 (November  2, 2023)

NOTES:

* resource/aws_detective_organization_admin_account: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#25237](https://github.com/hashicorp/terraform-provider-aws/issues/25237))
* resource/aws_detective_organization_configuration: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#25237](https://github.com/hashicorp/terraform-provider-aws/issues/25237))

FEATURES:

* **New Data Source:** `aws_opensearchserverless_lifecycle_policy` ([#34144](https://github.com/hashicorp/terraform-provider-aws/issues/34144))
* **New Resource:** `aws_detective_organization_admin_account` ([#25237](https://github.com/hashicorp/terraform-provider-aws/issues/25237))
* **New Resource:** `aws_detective_organization_configuration` ([#25237](https://github.com/hashicorp/terraform-provider-aws/issues/25237))
* **New Resource:** `aws_opensearchserverless_lifecycle_policy` ([#34144](https://github.com/hashicorp/terraform-provider-aws/issues/34144))
* **New Resource:** `aws_redshift_resource_policy` ([#34149](https://github.com/hashicorp/terraform-provider-aws/issues/34149))
* **New Resource:** `aws_verifiedaccess_endpoint` ([#30763](https://github.com/hashicorp/terraform-provider-aws/issues/30763))

ENHANCEMENTS:

* resource/aws_amplify_app: Add `custom_headers` argument ([#31561](https://github.com/hashicorp/terraform-provider-aws/issues/31561))
* resource/aws_batch_job_definition: Add `node_properties` argument ([#34153](https://github.com/hashicorp/terraform-provider-aws/issues/34153))
* resource/aws_finspace_kx_cluster: In-place updates are now supported for the `code`, `database`, and `initialization_script` arguments. The update timeout has been increased to 30 minutes. ([#34220](https://github.com/hashicorp/terraform-provider-aws/issues/34220))
* resource/aws_iot_topic_rule: Add `kafka.header` and `error_action.kafka.header` arguments ([#34191](https://github.com/hashicorp/terraform-provider-aws/issues/34191))
* resource/aws_networkmanager_connect_attachment: Add `NO_ENCAP` as a valid `options.protocol` value ([#34109](https://github.com/hashicorp/terraform-provider-aws/issues/34109))
* resource/aws_networkmanager_connect_peer: Add `subnet_arn` argument to support [Tunnel-less Connect attachments](https://docs.aws.amazon.com/network-manager/latest/cloudwan/cloudwan-connect-attachment.html#cloudwan-connect-tlc) ([#34109](https://github.com/hashicorp/terraform-provider-aws/issues/34109))
* resource/aws_networkmanager_connect_peer: `inside_cidr_blocks` is Optional ([#34109](https://github.com/hashicorp/terraform-provider-aws/issues/34109))
* resource/aws_rds_cluster: Remove the provider default (previously, "1") and use the AWS default for `backup_retention_period` (also, "1") to allow integration with AWS Backup ([#34187](https://github.com/hashicorp/terraform-provider-aws/issues/34187))
* resource/aws_redshift_cluster: Add `snapshot_arn` argument ([#34181](https://github.com/hashicorp/terraform-provider-aws/issues/34181))
* resource/aws_redshift_cluster: Add the `manage_master_password` and `master_password_secret_kms_key_id` arguments to support managed admin credentials ([#34182](https://github.com/hashicorp/terraform-provider-aws/issues/34182))
* resource/aws_s3_object: Add `override_provider` configuration block, allowing tags inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) to be ignored ([#33262](https://github.com/hashicorp/terraform-provider-aws/issues/33262))
* resource/aws_secretsmanager_secret_rotation: The `rotation_lambda_arn` argument is now optional to support modifying the rotation schedule of AWS-managed secrets. ([#34180](https://github.com/hashicorp/terraform-provider-aws/issues/34180))

BUG FIXES:

* data-source/aws_vpc_ipam_pools: Add `id` attribute for individual IPAM pools ([#32133](https://github.com/hashicorp/terraform-provider-aws/issues/32133))
* resource/aws_alb_listener_rule: Fixed the `action.forward.target_group` argument minimum item requirement. Previously this was set to 2, but the AWS API allows specifying a single target group. ([#33727](https://github.com/hashicorp/terraform-provider-aws/issues/33727))
* resource/aws_amplify_branch: Remove [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) from `enable_performance_mode` ([#34141](https://github.com/hashicorp/terraform-provider-aws/issues/34141))
* resource/aws_lb_listener_rule: Fixed the `action.forward.target_group` argument minimum item requirement. Previously this was set to 2, but the AWS API allows specifying a single target group. ([#33727](https://github.com/hashicorp/terraform-provider-aws/issues/33727))
* resource/aws_quicksight_analysis: Fix "expected type to be integer" errors in `window_options.bounds.*` argument validatation functions ([#34230](https://github.com/hashicorp/terraform-provider-aws/issues/34230))
* resource/aws_quicksight_dashboard: Fix "expected type to be integer" errors in `window_options.bounds.*` argument validatation functions ([#34230](https://github.com/hashicorp/terraform-provider-aws/issues/34230))
* resource/aws_quicksight_template: Fix "expected type to be integer" errors in `window_options.bounds.*` argument validatation functions ([#34230](https://github.com/hashicorp/terraform-provider-aws/issues/34230))
* resource/aws_rds_cluster: Avoid an error on delete related to `unexpected state 'scaling-compute'` ([#34187](https://github.com/hashicorp/terraform-provider-aws/issues/34187))

## 5.23.1 (October 27, 2023)

BUG FIXES:

* data-source/aws_lambda_function: Add `vpc_config.ipv6_allowed_for_dual_stack` attribute, fixing `Invalid address to set: []string{"vpc_config", "0", "ipv6_allowed_for_dual_stack"}` errors ([#34134](https://github.com/hashicorp/terraform-provider-aws/issues/34134))

## 5.23.0 (October 26, 2023)

NOTES:

* provider: This release includes an update to the AWS SDK for Go v2 with breaking type changes to several services: `finspace`, `kafka`, `medialive`, `rds`, `s3control`, `timestreamwrite`, and `xray`. These changes primarily affect how arguments with default values are serialized for outbound requests, changing scalar types to pointers. See [this AWS SDK for Go V2 issue](https://github.com/aws/aws-sdk-go-v2/issues/2162) for additional context. The corresponding provider changes should make this breakfix transparent to users, but as with any breaking change there is the potential for missed edge cases. If errors are observed in the impacted resources, please link to this dependency update pull request in the bug report. ([#34096](https://github.com/hashicorp/terraform-provider-aws/issues/34096))

FEATURES:

* **New Resource:** `aws_iot_domain_configuration` ([#24765](https://github.com/hashicorp/terraform-provider-aws/issues/24765))

ENHANCEMENTS:

* data-source/aws_imagebuilder_image: Add `image_scanning_configuration` attribute ([#34049](https://github.com/hashicorp/terraform-provider-aws/issues/34049))
* resource/aws_config_config_rule: Add `evaluation_mode` attribute ([#34033](https://github.com/hashicorp/terraform-provider-aws/issues/34033))
* resource/aws_elasticache_replication_group: Add `ip_discovery` and `network_type` arguments ([#34019](https://github.com/hashicorp/terraform-provider-aws/issues/34019))
* resource/aws_imagebuilder_image: Add `image_scanning_configuration` configuration block ([#34049](https://github.com/hashicorp/terraform-provider-aws/issues/34049))
* resource/aws_kms_key: Add configurable timeouts ([#34112](https://github.com/hashicorp/terraform-provider-aws/issues/34112))
* resource/aws_lambda_function: Add `vpc_config.ipv6_allowed_for_dual_stack` argument ([#34045](https://github.com/hashicorp/terraform-provider-aws/issues/34045))
* resource/aws_lb: Add `dns_record_client_routing_policy` attribute to configure Availability Zonal DNS affinity on Network Load Balancer (NLB) ([#33992](https://github.com/hashicorp/terraform-provider-aws/issues/33992))
* resource/aws_lb_target_group: Add `target_health_state` configuration block ([#34070](https://github.com/hashicorp/terraform-provider-aws/issues/34070))
* resource/aws_lb_target_group: Remove default value (`false`) for `connection_termination` argument and mark as Computed, to support new default behavior for UDP/TCP_UDP target groups ([#34070](https://github.com/hashicorp/terraform-provider-aws/issues/34070))
* resource/aws_neptune_cluster: Add `slowquery` as a valid `enable_cloudwatch_logs_exports` value ([#34053](https://github.com/hashicorp/terraform-provider-aws/issues/34053))

BUG FIXES:

* provider/tags: Prevent crash when `tags_all` is null ([#34073](https://github.com/hashicorp/terraform-provider-aws/issues/34073))
* resource/aws_autoscaling_group: Fix error when `launch_template` name is updated. ([#34086](https://github.com/hashicorp/terraform-provider-aws/issues/34086))
* resource/aws_dms_s3_endpoint: Don't send the default value of `false` for `add_trailing_padding_character`, maintaining compatibility with older ([pre-3.4.7](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_ReleaseNotes.html#CHAP_ReleaseNotes.DMS347)) DMS engine versions ([#34048](https://github.com/hashicorp/terraform-provider-aws/issues/34048))
* resource/aws_ecs_task_definition: Add `0` as a valid value for `volume.efs_volume_configuration.transit_encryption_port`, preventing unexpected drift ([#34020](https://github.com/hashicorp/terraform-provider-aws/issues/34020))
* resource/aws_identitystore_group: Fix updating `description` attribute when it is changed ([#34037](https://github.com/hashicorp/terraform-provider-aws/issues/34037))
* resource/aws_iot_indexing_configuration: Add `thing_indexing_configuration.filter` attribute, resolving `InvalidRequestException: NamedShadowNames Filter must not be empty for enabling NamedShadowIndexingMode` errors ([#26859](https://github.com/hashicorp/terraform-provider-aws/issues/26859))
* resource/aws_storagegateway_gateway: Support the value `0` (representing Sunday) for `maintenance_start_time.day_of_week` ([#34015](https://github.com/hashicorp/terraform-provider-aws/issues/34015))
* resource/aws_verifiedaccess_group: Fix `InvalidParameterValue: Policy Document cannot be provided when Policy Enabled is false or missing` errors when updating `policy_document` ([#34054](https://github.com/hashicorp/terraform-provider-aws/issues/34054))

## 5.22.0 (October 19, 2023)

FEATURES:

* **New Data Source:** `aws_media_convert_queue` ([#27075](https://github.com/hashicorp/terraform-provider-aws/issues/27075))
* **New Resource:** `aws_elasticsearch_vpc_endpoint` ([#33925](https://github.com/hashicorp/terraform-provider-aws/issues/33925))
* **New Resource:** `aws_msk_replicator` ([#33973](https://github.com/hashicorp/terraform-provider-aws/issues/33973))

ENHANCEMENTS:

* data-source/aws_ec2_client_vpn_endpoint: Add `self_service_portal_url` attribute ([#34007](https://github.com/hashicorp/terraform-provider-aws/issues/34007))
* resource/aws_alb: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_alb_target_group: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_cloudfront_public_key: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_db_option_group: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_docdb_cluster: Support import of `cluster_identifier_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_docdb_cluster_instance: Support import of `identifier_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_docdb_cluster_parameter_group: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_docdb_subnet_group: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_ec2_client_vpn_endpoint: Add `self_service_portal_url` attribute ([#34007](https://github.com/hashicorp/terraform-provider-aws/issues/34007))
* resource/aws_elb: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_emr_security_configuration: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_iam_group_policy: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_iam_role_policy: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_iam_user_policy: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_iot_provisioning_template: Add `type` attribute ([#33950](https://github.com/hashicorp/terraform-provider-aws/issues/33950))
* resource/aws_lb: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_lb_target_group: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_neptune_cluster: Support import of `cluster_identifier_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_neptune_cluster_instance: Support import of `identifier_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_neptune_cluster_parameter_group: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_neptune_event_subscription: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_pinpoint_app: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_rds_cluster: Support import of `cluster_identifier_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_rds_cluster_instance: Support import of `identifier_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_signer_signing_profile: Support import of `name_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_signer_signing_profile_permission: Add `signer:SignPayload` as a valid `action` value ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_signer_signing_profile_permission: Support import of `statement_id_prefix` argument ([#33852](https://github.com/hashicorp/terraform-provider-aws/issues/33852))
* resource/aws_transfer_server: Change `pre_authentication_login_banner` and `post_authentication_login_banner` length limits to 4096 ([#33937](https://github.com/hashicorp/terraform-provider-aws/issues/33937))
* resource/aws_wafv2_web_acl: Add `ja3_fingerprint` to `field_to_match` configuration blocks ([#33933](https://github.com/hashicorp/terraform-provider-aws/issues/33933))

BUG FIXES:

* data-source/aws_dms_certificate: Fix crash when certificate not found ([#34012](https://github.com/hashicorp/terraform-provider-aws/issues/34012))
* resource/aws_cloudformation_stack: Fix error when `computed` values are not set when there is no update ([#33969](https://github.com/hashicorp/terraform-provider-aws/issues/33969))
* resource/aws_codecommit_repository: Doesn't force replacement when renaming ([#32207](https://github.com/hashicorp/terraform-provider-aws/issues/32207))
* resource/aws_db_instance: Creating resource from snapshot or point-in-time recovery now handles `manage_master_user_password` and `master_user_secret_kms_key_id` attributes correctly ([#33699](https://github.com/hashicorp/terraform-provider-aws/issues/33699))
* resource/aws_elasticache_replication_group: Fix error when switching `engine_version` from `6.x` to a specific `6.<digit>` version number ([#33954](https://github.com/hashicorp/terraform-provider-aws/issues/33954))
* resource/aws_iam_role: Fix refreshing `permission_boundary` when deleted outside of Terraform ([#33963](https://github.com/hashicorp/terraform-provider-aws/issues/33963))
* resource/aws_iam_user: Fix refreshing `permission_boundary` when deleted outside of Terraform ([#33963](https://github.com/hashicorp/terraform-provider-aws/issues/33963))
* resource/aws_inspector2_enabler: Fix `Value at 'resourceTypes' failed to satisfy constraint` errors ([#33348](https://github.com/hashicorp/terraform-provider-aws/issues/33348))
* resource/aws_neptune_cluster_instance: Remove [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) from `engine_version` ([#33487](https://github.com/hashicorp/terraform-provider-aws/issues/33487))
* resource/aws_neptune_cluster_parameter_group: Fix condition where defined cluster parameters with system default values are seen as updates ([#33487](https://github.com/hashicorp/terraform-provider-aws/issues/33487))
* resource/aws_s3_bucket_object_lock_configuration: Fix `found resource` errors on Delete ([#33966](https://github.com/hashicorp/terraform-provider-aws/issues/33966))

## 5.21.0 (October 12, 2023)

FEATURES:

* **New Data Source:** `aws_servicequotas_templates` ([#33871](https://github.com/hashicorp/terraform-provider-aws/issues/33871))
* **New Resource:** `aws_ec2_image_block_public_access` ([#33810](https://github.com/hashicorp/terraform-provider-aws/issues/33810))
* **New Resource:** `aws_guardduty_organization_configuration_feature` ([#33913](https://github.com/hashicorp/terraform-provider-aws/issues/33913))
* **New Resource:** `aws_servicequotas_template_association` ([#33725](https://github.com/hashicorp/terraform-provider-aws/issues/33725))
* **New Resource:** `aws_verifiedaccess_group` ([#33297](https://github.com/hashicorp/terraform-provider-aws/issues/33297))
* **New Resource:** `aws_verifiedaccess_instance_logging_configuration` ([#33864](https://github.com/hashicorp/terraform-provider-aws/issues/33864))

ENHANCEMENTS:

* data-source/aws_dms_endpoint: Add `s3_settings.glue_catalog_generation` attribute ([#33778](https://github.com/hashicorp/terraform-provider-aws/issues/33778))
* data-source/aws_msk_cluster: Add `cluster_uuid` attribute ([#33805](https://github.com/hashicorp/terraform-provider-aws/issues/33805))
* resource/aws_codedeploy_deployment_group: Add `outdated_instances_strategy` argument ([#33844](https://github.com/hashicorp/terraform-provider-aws/issues/33844))
* resource/aws_dms_endpoint: Add `s3_settings.glue_catalog_generation` attribute ([#33778](https://github.com/hashicorp/terraform-provider-aws/issues/33778))
* resource/aws_dms_s3_endpoint: Add `glue_catalog_generation` attribute ([#33778](https://github.com/hashicorp/terraform-provider-aws/issues/33778))
* resource/aws_docdb_cluster: Add `allow_major_version_upgrade` argument ([#33790](https://github.com/hashicorp/terraform-provider-aws/issues/33790))
* resource/aws_docdb_cluster_instance: Add `copy_tags_to_snapshot` argument ([#31022](https://github.com/hashicorp/terraform-provider-aws/issues/31022))
* resource/aws_dynamodb_table: Add `import_table` configuration block ([#33802](https://github.com/hashicorp/terraform-provider-aws/issues/33802))
* resource/aws_msk_cluster: Add `cluster_uuid` attribute ([#33805](https://github.com/hashicorp/terraform-provider-aws/issues/33805))
* resource/aws_msk_serverless_cluster: Add `cluster_uuid` attribute ([#33805](https://github.com/hashicorp/terraform-provider-aws/issues/33805))
* resource/aws_networkmanager_core_network: Add `base_policy_document` argument ([#33712](https://github.com/hashicorp/terraform-provider-aws/issues/33712))
* resource/aws_redshiftserverless_workgroup: Allow `require_ssl` and `use_fips_ssl` `config_parameters` keys ([#33916](https://github.com/hashicorp/terraform-provider-aws/issues/33916))
* resource/aws_s3_bucket: Use configurable timeout for resource Delete ([#33845](https://github.com/hashicorp/terraform-provider-aws/issues/33845))
* resource/aws_verifiedaccess_instance: Add `fips_enabled` argument ([#33880](https://github.com/hashicorp/terraform-provider-aws/issues/33880))
* resource/aws_vpclattice_target_group: Add `config.lambda_event_structure_version` argument ([#33804](https://github.com/hashicorp/terraform-provider-aws/issues/33804))
* resource/aws_vpclattice_target_group: Make `config.port`, `config.protocol` and `config.vpc_identifier` optional ([#33804](https://github.com/hashicorp/terraform-provider-aws/issues/33804))
* resource/aws_wafv2_web_acl: Add `aws_managed_rules_acfp_rule_set` to `managed_rule_group_configs` configuration block ([#33915](https://github.com/hashicorp/terraform-provider-aws/issues/33915))

BUG FIXES:

* provider: Respect valid values for the `AWS_S3_US_EAST_1_REGIONAL_ENDPOINT` environment variable when configuring the S3 API client ([#33874](https://github.com/hashicorp/terraform-provider-aws/issues/33874))
* resource/aws_appflow_connector_profile: Fix various crashes ([#33856](https://github.com/hashicorp/terraform-provider-aws/issues/33856))
* resource/aws_db_parameter_group: Group names containing periods (`.`) no longer fail validation ([#33704](https://github.com/hashicorp/terraform-provider-aws/issues/33704))
* resource/aws_opensearchserverless_collection: Fix crash when error is returned ([#33918](https://github.com/hashicorp/terraform-provider-aws/issues/33918))
* resource/aws_rds_cluster_parameter_group: Group names containing periods (`.`) no longer fail validation ([#33704](https://github.com/hashicorp/terraform-provider-aws/issues/33704))

## 5.20.1 (October 10, 2023)

NOTES:

* provider: Build with [Terraform Plugin Framework v1.4.1](https://github.com/hashicorp/terraform-plugin-framework/blob/main/CHANGELOG.md#141-october-09-2023), fixing potential [initialization errors](https://github.com/hashicorp/terraform/issues/33990) when using v1.6 of the Terraform CLI.

## 5.20.0 (October  6, 2023)

FEATURES:

* **New Resource:** `aws_guardduty_detector_feature` ([#31463](https://github.com/hashicorp/terraform-provider-aws/issues/31463))
* **New Resource:** `aws_servicequotas_template` ([#33688](https://github.com/hashicorp/terraform-provider-aws/issues/33688))
* **New Resource:** `aws_sesv2_account_vdm_attributes` ([#33705](https://github.com/hashicorp/terraform-provider-aws/issues/33705))
* **New Resource:** `aws_verifiedaccess_instance_trust_provider_attachment` ([#33734](https://github.com/hashicorp/terraform-provider-aws/issues/33734))

ENHANCEMENTS:

* data-source/aws_guardduty_detector: Add `features` attribute ([#31463](https://github.com/hashicorp/terraform-provider-aws/issues/31463))
* resource/aws_finspace_kx_cluster: Increase default creation timeout to 45 minutes, default deletion timeout to 60 minutes ([#33745](https://github.com/hashicorp/terraform-provider-aws/issues/33745))
* resource/aws_finspace_kx_environment: Increase default deletion timeout to 45 minutes ([#33745](https://github.com/hashicorp/terraform-provider-aws/issues/33745))
* resource/aws_guardduty_filter: Add plan-time validation of `name` ([#21030](https://github.com/hashicorp/terraform-provider-aws/issues/21030))
* resource/aws_kinesis_firehose_delivery_stream: Add `opensearchserverless_configuration` and `msk_source_configuration` configuration blocks ([#33101](https://github.com/hashicorp/terraform-provider-aws/issues/33101))
* resource/aws_kinesis_firehose_delivery_stream: Add `opensearchserverless` as a valid `destination` value ([#33101](https://github.com/hashicorp/terraform-provider-aws/issues/33101))

BUG FIXES:

* data-source/aws_fsx_ontap_storage_virtual_machine: Fix crash when `active_directory_configuration.self_managed_active_directory_configuration.file_system_administrators_group` is not configured ([#33800](https://github.com/hashicorp/terraform-provider-aws/issues/33800))
* resource/aws_ec2_transit_gateway_route : Fix TGW route search filter to avoid routes being missed when more than 1,000 static routes are in a TGW route table ([#33765](https://github.com/hashicorp/terraform-provider-aws/issues/33765))
* resource/aws_fsx_ontap_storage_virtual_machine: Fix crash when `active_directory_configuration.self_managed_active_directory_configuration.file_system_administrators_group` is not configured ([#33800](https://github.com/hashicorp/terraform-provider-aws/issues/33800))
* resource/aws_medialive_channel: Fix VPC settings flatten/expand/docs. ([#33558](https://github.com/hashicorp/terraform-provider-aws/issues/33558))
* resource/aws_vpc_endpoint: Set `dns_options.dns_record_ip_type` to `Computed` to prevent diffs ([#33743](https://github.com/hashicorp/terraform-provider-aws/issues/33743))

## 5.19.0 (September 29, 2023)

BREAKING CHANGES:

* data-source/aws_s3_bucket_object: Following migration to [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/), the `metadata` attribute's [keys](https://developer.hashicorp.com/terraform/language/expressions/types#maps-objects) are always [returned in lowercase](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#HeadObjectOutput) ([#33660](https://github.com/hashicorp/terraform-provider-aws/issues/33660))
* data-source/aws_s3_object: Following migration to [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/), the `metadata` attribute's [keys](https://developer.hashicorp.com/terraform/language/expressions/types#maps-objects) are always [returned in lowercase](https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3#HeadObjectOutput) ([#33660](https://github.com/hashicorp/terraform-provider-aws/issues/33660))

NOTES:

* data-source/aws_s3_bucket_object: The `metadata` attribute's keys are now always returned in lowercase. Please modify configurations as necessary ([#33660](https://github.com/hashicorp/terraform-provider-aws/issues/33660))
* data-source/aws_s3_object: The `metadata` attribute's keys are now always returned in lowercase. Please modify configurations as necessary ([#33660](https://github.com/hashicorp/terraform-provider-aws/issues/33660))
* resource/aws_iam_*: This release introduces additional validation of IAM policy JSON arguments to detect duplicate keys. Previously, arguments with duplicated keys resulted in all but one of the key values being overwritten. Since this results in unexpected IAM policies being submitted to AWS, we have updated the validation logic to error in these cases. This may cause existing IAM policy arguments to fail validation, however, those policies are likely not what was originally intended. ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))

FEATURES:

* **New Resource:** `aws_cleanrooms_configured_table` ([#33602](https://github.com/hashicorp/terraform-provider-aws/issues/33602))
* **New Resource:** `aws_dms_replication_config` ([#32908](https://github.com/hashicorp/terraform-provider-aws/issues/32908))
* **New Resource:** `aws_lexv2models_bot` ([#33475](https://github.com/hashicorp/terraform-provider-aws/issues/33475))
* **New Resource:** `aws_rds_custom_db_engine_version` ([#33285](https://github.com/hashicorp/terraform-provider-aws/issues/33285))

ENHANCEMENTS:

* resource/aws_cloud9_environment_ec2: Add `ubuntu-22.04-x86_64` and `resolve:ssm:/aws/service/cloud9/amis/ubuntu-22.04-x86_64` as valid values for `image_id` ([#33662](https://github.com/hashicorp/terraform-provider-aws/issues/33662))
* resource/aws_fsx_ontap_volume: Add `bypass_snaplock_enterprise_retention` argument and `snaplock_configuration` configuration block to support [SnapLock](https://docs.aws.amazon.com/fsx/latest/ONTAPGuide/snaplock.html) ([#32530](https://github.com/hashicorp/terraform-provider-aws/issues/32530))
* resource/aws_fsx_ontap_volume: Add `copy_tags_to_backups` and `snapshot_policy` arguments ([#32530](https://github.com/hashicorp/terraform-provider-aws/issues/32530))
* resource/aws_fsx_openzfs_volume: Add `delete_volume_options` argument ([#32530](https://github.com/hashicorp/terraform-provider-aws/issues/32530))
* resource/aws_lightsail_bucket: Add `force_delete` argument ([#33586](https://github.com/hashicorp/terraform-provider-aws/issues/33586))
* resource/aws_opensearch_outbound_connection: Add `connection_properties`, `connection_mode` and `accept_connection` arguments ([#32990](https://github.com/hashicorp/terraform-provider-aws/issues/32990))
* resource/aws_wafv2_rule_group: Add `rate_based_statement.custom_key` configuration block ([#33594](https://github.com/hashicorp/terraform-provider-aws/issues/33594))
* resource/aws_wafv2_web_acl: Add `rate_based_statement.custom_key` configuration block ([#33594](https://github.com/hashicorp/terraform-provider-aws/issues/33594))

BUG FIXES:

* resource/aws_batch_job_queue: Correctly validates elements of `compute_environments` as ARNs ([#33577](https://github.com/hashicorp/terraform-provider-aws/issues/33577))
* resource/aws_cloudfront_continuous_deployment_policy: Fix `IllegalUpdate` errors when updating a staging `aws_cloudfront_distribution` that is part of continuous deployment ([#33578](https://github.com/hashicorp/terraform-provider-aws/issues/33578))
* resource/aws_cloudfront_distribution: Fix `IllegalUpdate` errors when updating a staging distribution associated with an `aws_cloudfront_continuous_deployment_policy` ([#33578](https://github.com/hashicorp/terraform-provider-aws/issues/33578))
* resource/aws_cloudfront_distribution: Fix `PreconditionFailed` errors when destroying a distribution associated with an `aws_cloudfront_continuous_deployment_policy` ([#33578](https://github.com/hashicorp/terraform-provider-aws/issues/33578))
* resource/aws_cloudfront_distribution: Fix `StagingDistributionInUse` errors when destroying a distribution associated with an `aws_cloudfront_continuous_deployment_policy` ([#33578](https://github.com/hashicorp/terraform-provider-aws/issues/33578))
* resource/aws_datasync_location_fsx_ontap_file_system: Correct handling of `protocol.smb.domain`, `protocol.smb.user` and `protocol.smb.password` ([#33641](https://github.com/hashicorp/terraform-provider-aws/issues/33641))
* resource/aws_glacier_vault_lock: Fail validation if duplicated keys are found in `policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_iam_group_policy: Fail validation if duplicated keys are found in `policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_iam_policy: Fail validation if duplicated keys are found in `policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_iam_role: Fail validation if duplicated keys are found in `assume_role_policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_iam_role_policy: Fail validation if duplicated keys are found in `policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_iam_user_policy: Fail validation if duplicated keys are found in `policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_mediastore_container_policy: Fail validation if duplicated keys are found in `policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_s3_bucket_policy: Fix intermittent `couldn't find resource` errors on resource Create ([#33537](https://github.com/hashicorp/terraform-provider-aws/issues/33537))
* resource/aws_ssoadmin_permission_set_inline_policy: Fail validation if duplicated keys are found in `inline_policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_transfer_access: Fail validation if duplicated keys are found in `policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))
* resource/aws_transfer_user: Fail validation if duplicated keys are found in `policy` ([#33570](https://github.com/hashicorp/terraform-provider-aws/issues/33570))

## 5.18.1 (September 26, 2023)

NOTES:

* documentation: Duplicate CDKTF guides with differing file extensions have been removed to resolve failures in the provider release workflow. ([#33630](https://github.com/hashicorp/terraform-provider-aws/issues/33630))

## 5.18.0 (September 21, 2023)

FEATURES:

* **New Data Source:** `aws_fsx_ontap_file_system` ([#32503](https://github.com/hashicorp/terraform-provider-aws/issues/32503))
* **New Data Source:** `aws_fsx_ontap_storage_virtual_machine` ([#32621](https://github.com/hashicorp/terraform-provider-aws/issues/32621))
* **New Data Source:** `aws_fsx_ontap_storage_virtual_machines` ([#32624](https://github.com/hashicorp/terraform-provider-aws/issues/32624))
* **New Data Source:** `aws_organizations_organizational_unit` ([#33408](https://github.com/hashicorp/terraform-provider-aws/issues/33408))
* **New Resource:** `aws_opensearch_package` ([#33227](https://github.com/hashicorp/terraform-provider-aws/issues/33227))
* **New Resource:** `aws_opensearch_package_association` ([#33227](https://github.com/hashicorp/terraform-provider-aws/issues/33227))

ENHANCEMENTS:

* resource/aws_fsx_ontap_storage_virtual_machine: Remove [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) from `active_directory_configuration.self_managed_active_directory_configuration.domain_name`, `active_directory_configuration.self_managed_active_directory_configuration.file_system_administrators_group` and `active_directory_configuration.self_managed_active_directory_configuration.organizational_unit_distinguished_name` allowing an SVM to join AD after creation ([#33466](https://github.com/hashicorp/terraform-provider-aws/issues/33466))

BUG FIXES:

* data-source/aws_sesv2_email_identity: Mark `dkim_signing_attributes.domain_signing_private_key` as sensitive ([#33477](https://github.com/hashicorp/terraform-provider-aws/issues/33477))
* resource/aws_db_instance: Fix so that `storage_throughput` can be changed when `iops` and `allocated_storage` are not changed ([#33529](https://github.com/hashicorp/terraform-provider-aws/issues/33529))
* resource/aws_db_option_group: Avoid erroneous differences being reported when an `option` `port` and/or `version` is not set ([#33511](https://github.com/hashicorp/terraform-provider-aws/issues/33511))
* resource/aws_fsx_ontap_storage_virtual_machine: Avoid recreating resource when `active_directory_configuration.self_managed_active_directory_configuration.file_system_administrators_group` is configured ([#33466](https://github.com/hashicorp/terraform-provider-aws/issues/33466))
* resource/aws_fsx_ontap_storage_virtual_machine: Change `file_system_id` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#32621](https://github.com/hashicorp/terraform-provider-aws/issues/32621))
* resource/aws_s3_bucket_accelerate_configuration: Retry resource Delete on `OperationAborted: A conflicting conditional operation is currently in progress against this resource` errors ([#33531](https://github.com/hashicorp/terraform-provider-aws/issues/33531))
* resource/aws_s3_bucket_policy: Retry resource Delete on `OperationAborted: A conflicting conditional operation is currently in progress against this resource` errors ([#33531](https://github.com/hashicorp/terraform-provider-aws/issues/33531))
* resource/aws_s3_bucket_versioning: Retry resource Delete on `OperationAborted: A conflicting conditional operation is currently in progress against this resource` errors ([#33531](https://github.com/hashicorp/terraform-provider-aws/issues/33531))
* resource/aws_sesv2_email_identity: Mark `dkim_signing_attributes.domain_signing_private_key` as sensitive ([#33477](https://github.com/hashicorp/terraform-provider-aws/issues/33477))

## 5.17.0 (September 14, 2023)

NOTES:

* data-source/aws_s3_object: Migration to [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/) means that the edge case of specifying a single `/` as the value for `key` is no longer supported ([#33358](https://github.com/hashicorp/terraform-provider-aws/issues/33358))

FEATURES:

* **New Resource:** `aws_shield_application_layer_automatic_response` ([#33432](https://github.com/hashicorp/terraform-provider-aws/issues/33432))
* **New Resource:** `aws_verifiedaccess_instance` ([#33459](https://github.com/hashicorp/terraform-provider-aws/issues/33459))

ENHANCEMENTS:

* data-source/aws_s3_object: Add `checksum_mode` argument and `checksum_crc32`, `checksum_crc32c`, `checksum_sha1` and `checksum_sha256` attributes ([#33358](https://github.com/hashicorp/terraform-provider-aws/issues/33358))
* data-source/aws_s3control_multi_region_access_point: Add `details.region.bucket_account_id` attribute ([#33416](https://github.com/hashicorp/terraform-provider-aws/issues/33416))
* resource/aws_s3_object: Add `checksum_algorithm` argument and `checksum_crc32`, `checksum_crc32c`, `checksum_sha1` and `checksum_sha256` attributes ([#33358](https://github.com/hashicorp/terraform-provider-aws/issues/33358))
* resource/aws_s3_object_copy: Add `checksum_algorithm` argument and `checksum_crc32`, `checksum_crc32c`, `checksum_sha1` and `checksum_sha256` attributes ([#33358](https://github.com/hashicorp/terraform-provider-aws/issues/33358))
* resource/aws_s3control_multi_region_access_point: Add `details.region.bucket_account_id` argument to support [cross-account Multi-Region Access Points](https://docs.aws.amazon.com/AmazonS3/latest/userguide/multi-region-access-point-buckets.html) ([#33416](https://github.com/hashicorp/terraform-provider-aws/issues/33416))
* resource/aws_s3control_multi_region_access_point: Add `details.region.region` attribute ([#33416](https://github.com/hashicorp/terraform-provider-aws/issues/33416))
* resource/aws_schemas_schema: Add `JSONSchemaDraft4` schema type support ([#33442](https://github.com/hashicorp/terraform-provider-aws/issues/33442))
* resource/aws_transfer_connector: Add `sftp_config` argument and make `as2_config` optional ([#32741](https://github.com/hashicorp/terraform-provider-aws/issues/32741))
* resource/aws_wafv2_web_acl: Retry resource Update on `WAFOptimisticLockException` errors ([#33432](https://github.com/hashicorp/terraform-provider-aws/issues/33432))

BUG FIXES:

* resource/aws_dms_replication_task: Fix error when `replication_task_settings` is `nil` ([#33456](https://github.com/hashicorp/terraform-provider-aws/issues/33456))
* resource/aws_elasticache_cluster: Fix regression for `redis` engine types caused by the new `transit_encryption_enabled` argument ([#33451](https://github.com/hashicorp/terraform-provider-aws/issues/33451))
* resource/aws_neptune_cluster: Fix ignored `kms_key_arn` on restore from DB cluster snapshot ([#33413](https://github.com/hashicorp/terraform-provider-aws/issues/33413))
* resource/aws_servicecatalog_product: Allow import on `provisioning_artifact_parameters` attribute ([#33448](https://github.com/hashicorp/terraform-provider-aws/issues/33448))
* resource/aws_subnet: Fix destroy error when there is a lingering ENI for DMS ([#33375](https://github.com/hashicorp/terraform-provider-aws/issues/33375))

## 5.16.2 (September 11, 2023)

FEATURES:

* **New Data Source:** `aws_cognito_identity_pool` ([#33053](https://github.com/hashicorp/terraform-provider-aws/issues/33053))
* **New Resource:** `aws_verifiedaccess_trust_provider` ([#33195](https://github.com/hashicorp/terraform-provider-aws/issues/33195))

ENHANCEMENTS:

* resource/aws_autoscaling_group: Change the default values of `instance_refresh.preferences.scale_in_protected_instances` and `instance_refresh.preferences.standby_instances` from `Wait` to the [Amazon EC2 Auto Scaling console recommended value](https://docs.aws.amazon.com/autoscaling/ec2/userguide/understand-instance-refresh-default-values.html) of `Ignore` ([#33382](https://github.com/hashicorp/terraform-provider-aws/issues/33382))
* resource/aws_s3control_object_lambda_access_point: Add `alias` attribute ([#33388](https://github.com/hashicorp/terraform-provider-aws/issues/33388))

BUG FIXES:

* resource/aws_autoscaling_group: Fix `ValidationError` errors when starting Auto Scaling group instance refresh ([#33382](https://github.com/hashicorp/terraform-provider-aws/issues/33382))
* resource/aws_iot_topic_rule: Fix `InvalidParameter` errors on Update with Kafka destinations ([#33360](https://github.com/hashicorp/terraform-provider-aws/issues/33360))
* resource/aws_lightsail_certificate: Fix validation of `name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_database: Fix validation of `name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_disk: Fix validation of `name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_instance: Fix validation of `name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_lb: Fix validation of `lb_name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_lb_attachment: Fix validation of `lb_name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_lb_certificate: Fix validation of `lb_name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_lb_certificate_attachment: Fix validation of `lb_name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_lb_https_redirection_policy: Fix validation of `lb_name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))
* resource/aws_lightsail_lb_stickiness_policy: Fix validation of `lb_name` ([#33405](https://github.com/hashicorp/terraform-provider-aws/issues/33405))

## 5.16.1 (September  8, 2023)

BUG FIXES:

* data-source/aws_efs_file_system: Fix `Search returned 0 results` errors when there are more than 101 file systems in the configured Region ([#33336](https://github.com/hashicorp/terraform-provider-aws/issues/33336))
* resource/aws_db_instance_automated_backups_replication: Fix `unexpected state` errors on resource Create ([#33369](https://github.com/hashicorp/terraform-provider-aws/issues/33369))
* resource/aws_glue_catalog_table: Fix removal of `metadata_location` and `table_type` `parameters` when updating Iceberg tables ([#33374](https://github.com/hashicorp/terraform-provider-aws/issues/33374))
* resource/aws_service_discovery_instance: Fix validation error "expected to match regular expression" ([#33371](https://github.com/hashicorp/terraform-provider-aws/issues/33371))

## 5.16.0 (September  8, 2023)

NOTES:

* provider: Performance regression introduced in v5.14.0 should be largely mitigated ([#33317](https://github.com/hashicorp/terraform-provider-aws/issues/33317))

FEATURES:

* **New Resource:** `aws_shield_drt_access_log_bucket_association` ([#33328](https://github.com/hashicorp/terraform-provider-aws/issues/33328))
* **New Resource:** `aws_shield_drt_access_role_arn_association` ([#33328](https://github.com/hashicorp/terraform-provider-aws/issues/33328))

ENHANCEMENTS:

* data-source/aws_api_gateway_api_key: Add `customer_id` attribute ([#33281](https://github.com/hashicorp/terraform-provider-aws/issues/33281))
* data-source/aws_fsx_windows_file_system: Add `disk_iops_configuration` attribute ([#33303](https://github.com/hashicorp/terraform-provider-aws/issues/33303))
* data-source/aws_opensearch_domain: Add `software_update_options` attribute ([#32234](https://github.com/hashicorp/terraform-provider-aws/issues/32234))
* data-source/aws_s3_objects: Add `request_payer` argument and `request_charged` attribute ([#33304](https://github.com/hashicorp/terraform-provider-aws/issues/33304))
* data-source/aws_s3_objects: Add plan-time validation of `encoding_type` ([#33304](https://github.com/hashicorp/terraform-provider-aws/issues/33304))
* resource/aws_api_gateway_account: Add `api_key_version` and `features` attributes ([#33279](https://github.com/hashicorp/terraform-provider-aws/issues/33279))
* resource/aws_api_gateway_api_key: Add `customer_id` argument ([#33281](https://github.com/hashicorp/terraform-provider-aws/issues/33281))
* resource/aws_api_gateway_api_key: Allow updating `name` ([#33281](https://github.com/hashicorp/terraform-provider-aws/issues/33281))
* resource/aws_autoscaling_group: Add `scale_in_protected_instances` and `standby_instances` attributes to `instance_refresh.preferences` configuration block ([#33310](https://github.com/hashicorp/terraform-provider-aws/issues/33310))
* resource/aws_dms_endpoint: Add `redshift-serverless` as valid value for `engine_name` ([#33316](https://github.com/hashicorp/terraform-provider-aws/issues/33316))
* resource/aws_elasticache_cluster: Add `transit_encryption_enabled` argument, enabling in-transit encryption for Memcached clusters inside a VPC ([#26987](https://github.com/hashicorp/terraform-provider-aws/issues/26987))
* resource/aws_fsx_windows_file_system: Add `disk_iops_configuration` configuration block ([#33303](https://github.com/hashicorp/terraform-provider-aws/issues/33303))
* resource/aws_glue_catalog_table: Add `open_table_format_input` configuration block to support open table formats such as [Apache Iceberg](https://iceberg.apache.org/) ([#33274](https://github.com/hashicorp/terraform-provider-aws/issues/33274))
* resource/aws_medialive_channel: Implement expand/flatten functions for `automatic_input_failover_settings` in `input_attachments` ([#33129](https://github.com/hashicorp/terraform-provider-aws/issues/33129))
* resource/aws_opensearch_domain: Add `software_update_options` attribute ([#32234](https://github.com/hashicorp/terraform-provider-aws/issues/32234))
* resource/aws_ssm_association: Add `sync_compliance` attribute ([#23515](https://github.com/hashicorp/terraform-provider-aws/issues/23515))

BUG FIXES:

* data-source/aws_identitystore_group: Restore `filter` argument to prevent `UnknownOperationException` errors in certain Regions ([#33311](https://github.com/hashicorp/terraform-provider-aws/issues/33311))
* data-source/aws_identitystore_user: Restore `filter` argument to prevent `UnknownOperationException` errors in certain Regions ([#33311](https://github.com/hashicorp/terraform-provider-aws/issues/33311))
* data-source/aws_s3_objects: Respect configured `max_keys` value if it's greater than `1000` ([#33304](https://github.com/hashicorp/terraform-provider-aws/issues/33304))
* resource/aws_api_gateway_account: Allow setting `cloudwatch_role_arn` to an empty value and set it correctly on Read, allowing its value to be determined on import ([#33279](https://github.com/hashicorp/terraform-provider-aws/issues/33279))
* resource/aws_fsx_ontap_file_system: Increase maximum value of `disk_iops_configuration.iops` to `160000` ([#33263](https://github.com/hashicorp/terraform-provider-aws/issues/33263))
* resource/aws_servicecatalog_principal_portfolio_association: Fix `ResourceNotFoundException` errors on resource Delete when configured `principal_type` is `IAM_PATTERN` ([#32243](https://github.com/hashicorp/terraform-provider-aws/issues/32243))

## 5.15.0 (August 31, 2023)

ENHANCEMENTS:

* data-source/aws_efs_file_system: Add `name` attribute ([#33243](https://github.com/hashicorp/terraform-provider-aws/issues/33243))
* data-source/aws_lakeformation_data_lake_settings: Add `read_only_admins` attribute ([#33189](https://github.com/hashicorp/terraform-provider-aws/issues/33189))
* data-source/aws_opensearch_domain: Add `cluster_config.multi_az_with_standby_enabled` attribute ([#33031](https://github.com/hashicorp/terraform-provider-aws/issues/33031))
* resource/aws_cloudformation_stack_set: Support resource import with `call_as = "DELEGATED_ADMIN"` via _StackSetName_,_CallAs_ syntax for `import` block or `terraform import` command ([#19092](https://github.com/hashicorp/terraform-provider-aws/issues/19092))
* resource/aws_cloudformation_stack_set_instance: Support resource import with `call_as = "DELEGATED_ADMIN"` via _StackSetName_,_AccountID_,_Region_,_CallAs_ syntax for `import` block or `terraform import` command ([#19092](https://github.com/hashicorp/terraform-provider-aws/issues/19092))
* resource/aws_datasync_location_fsx_openzfs_file_system: Fix `setting protocol: Invalid address to set` errors ([#33225](https://github.com/hashicorp/terraform-provider-aws/issues/33225))
* resource/aws_efs_file_system: Add `name` attribute ([#33243](https://github.com/hashicorp/terraform-provider-aws/issues/33243))
* resource/aws_fsx_openzfs_file_system: Add `endpoint_ip_address_range`, `preferred_subnet_id` and `route_table_ids` arguments to support the [Multi-AZ deployment type](https://docs.aws.amazon.com/fsx/latest/OpenZFSGuide/availability-durability.html#choosing-single-or-multi) ([#33245](https://github.com/hashicorp/terraform-provider-aws/issues/33245))
* resource/aws_lakeformation_data_lake_settings: Add `read_only_admins` argument ([#33189](https://github.com/hashicorp/terraform-provider-aws/issues/33189))
* resource/aws_opensearch_domain: Add `cluster_config.multi_az_with_standby_enabled` argument ([#33031](https://github.com/hashicorp/terraform-provider-aws/issues/33031))
* resource/aws_wafv2_rule_group: Add `name_prefix` argument ([#33206](https://github.com/hashicorp/terraform-provider-aws/issues/33206))
* resource/aws_wafv2_web_acl: Add `statement.managed_rule_group_statement.managed_rule_group_configs.aws_managed_rules_atp_rule_set.enable_regex_in_path` argument ([#33217](https://github.com/hashicorp/terraform-provider-aws/issues/33217))

BUG FIXES:

* provider: Correctly use old and new tag values when updating `tags` that are `computed` ([#33226](https://github.com/hashicorp/terraform-provider-aws/issues/33226))
* resource/aws_appflow_connector_profile: Fix validation on `oauth2` in `custom_connector_profile` ([#33192](https://github.com/hashicorp/terraform-provider-aws/issues/33192))
* resource/aws_cloudformation_stack_set: Fix `Can only set RetainStacksOnAccountRemoval if AutoDeployment is enabled` errors ([#19092](https://github.com/hashicorp/terraform-provider-aws/issues/19092))
* resource/aws_cloudwatch_event_bus_policy: Fix error during plan when the associated aws_cloudwatch_event_bus resource is manually deleted ([#33203](https://github.com/hashicorp/terraform-provider-aws/issues/33203))
* resource/aws_codeartifact_domain: Change the type of asset_size_bytes to `TypeString` instead of `TypeInt` to prevent `value out of range` panic ([#33220](https://github.com/hashicorp/terraform-provider-aws/issues/33220))
* resource/aws_efs_file_system_policy: Retry IAM eventual consistency errors ([#21734](https://github.com/hashicorp/terraform-provider-aws/issues/21734))
* resource/aws_fsx_openzfs_file_system: Wait for administrative action completion when updating root volume ([#33245](https://github.com/hashicorp/terraform-provider-aws/issues/33245))
* resource/aws_iot_thing_type: Fix error during plan when resource is manually deleted ([#33203](https://github.com/hashicorp/terraform-provider-aws/issues/33203))
* resource/aws_kms_key: Fix `tag propagation: timeout while waiting for state to become 'TRUE'` errors when any tag value is empty (`""`) ([#33226](https://github.com/hashicorp/terraform-provider-aws/issues/33226))
* resource/aws_wafv2_web_acl: Prevent deletion of the AWS-managed `ShieldMitigationRuleGroup` rule on resource Update ([#33216](https://github.com/hashicorp/terraform-provider-aws/issues/33216))

## 5.14.0 (August 24, 2023)

NOTES:

* data-source/aws_iam_policy_document: In some cases, `statement.*.condition` blocks with the same `test` and `variable` arguments were incorrectly handled by the provider. Since this results in unexpected IAM Policies being submitted to AWS, we have updated the logic to merge `values` lists in this case. This may cause existing IAM Policy documents to report a difference. However, those policies are likely not what was originally intended. ([#33093](https://github.com/hashicorp/terraform-provider-aws/issues/33093))

FEATURES:

* **New Resource:** `aws_datasync_location_azure_blob` ([#32632](https://github.com/hashicorp/terraform-provider-aws/issues/32632))
* **New Resource:** `aws_datasync_location_fsx_ontap_file_system` ([#32632](https://github.com/hashicorp/terraform-provider-aws/issues/32632))

ENHANCEMENTS:

* data-source/aws_dms_endpoint: Fix crash when specified endpoint not found ([#33158](https://github.com/hashicorp/terraform-provider-aws/issues/33158))
* data-source/aws_dms_replication_instance: Add `network_type` attribute ([#33158](https://github.com/hashicorp/terraform-provider-aws/issues/33158))
* data-source/aws_ec2_network_insights_path: Add `destination_arn` and `source_arn` attributes ([#33168](https://github.com/hashicorp/terraform-provider-aws/issues/33168))
* resource/aws_dms_replication_instance: Add `network_type` argument ([#33158](https://github.com/hashicorp/terraform-provider-aws/issues/33158))
* resource/aws_ec2_network_insights_path: Add `destination_arn` and `source_arn` attributes ([#33168](https://github.com/hashicorp/terraform-provider-aws/issues/33168))
* resource/aws_finspace_kx_environment: Add `transit_gateway_configuration.*.attachment_network_acl_configuration` argument. ([#33123](https://github.com/hashicorp/terraform-provider-aws/issues/33123))
* resource/aws_medialive_channel: Updates schemas for `selector_settings` for `audio_selector` and `selector_settings` for `caption_selector` ([#32714](https://github.com/hashicorp/terraform-provider-aws/issues/32714))
* resource/aws_ssoadmin_account_assignment: Add configurable timeouts ([#33121](https://github.com/hashicorp/terraform-provider-aws/issues/33121))
* resource/aws_ssoadmin_customer_managed_policy_attachment: Add configurable timeouts ([#33121](https://github.com/hashicorp/terraform-provider-aws/issues/33121))
* resource/aws_ssoadmin_managed_policy_attachment: Add configurable timeouts ([#33121](https://github.com/hashicorp/terraform-provider-aws/issues/33121))
* resource/aws_ssoadmin_permission_set: Add configurable timeouts ([#33121](https://github.com/hashicorp/terraform-provider-aws/issues/33121))
* resource/aws_ssoadmin_permission_set_inline_policy: Add configurable timeouts ([#33121](https://github.com/hashicorp/terraform-provider-aws/issues/33121))
* resource/aws_ssoadmin_permissions_boundary_attachment: Add configurable timeouts ([#33121](https://github.com/hashicorp/terraform-provider-aws/issues/33121))

BUG FIXES:

* data-source/aws_iam_policy_document: Fix inconsistent handling of `condition` blocks with duplicated `test` and `variable` arguments ([#33093](https://github.com/hashicorp/terraform-provider-aws/issues/33093))
* resource/aws_ec2_host: Fixed a bug that caused resource recreation when specifying an `outpost_arn` without an `asset_id` ([#33142](https://github.com/hashicorp/terraform-provider-aws/issues/33142))
* resource/aws_ec2_network_insights_analysis: Fix `setting forward_path_components: Invalid address to set` errors ([#33168](https://github.com/hashicorp/terraform-provider-aws/issues/33168))
* resource/aws_ec2_network_insights_path: Avoid recreating resource when passing an ARN as `source` or `destination` ([#33168](https://github.com/hashicorp/terraform-provider-aws/issues/33168))
* resource/aws_ec2_network_insights_path: Retry `AnalysisExistsForNetworkInsightsPath` errors on resource Delete ([#33168](https://github.com/hashicorp/terraform-provider-aws/issues/33168))
* resource/aws_kms_key: Fix `tag propagation: timeout while waiting for state to become 'TRUE'` errors when [`ignore_tags`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#ignore_tags) has been configured ([#33167](https://github.com/hashicorp/terraform-provider-aws/issues/33167))
* resource/aws_licensemanager_license_configuration: Surface `InvalidParameterValueException` errors during resource Delete ([#32845](https://github.com/hashicorp/terraform-provider-aws/issues/32845))
* resource/aws_msk_cluster_policy: Fix `Current cluster policy version needed for Update` errors ([#33118](https://github.com/hashicorp/terraform-provider-aws/issues/33118))
* resource/aws_quicksight_analysis: Change `definition.*.parameter_declarations` to a set type, preventing persistent differences ([#33120](https://github.com/hashicorp/terraform-provider-aws/issues/33120))
* resource/aws_quicksight_analysis: Fixed a bug that caused errors related to the `word_orientation` argument when using word cloud visuals. ([#33122](https://github.com/hashicorp/terraform-provider-aws/issues/33122))
* resource/aws_quicksight_analysis: Skip setting `definition.*.parameter_declarations.*.*_parameter_declaration.static_values` when empty, preventing persistent differences. ([#33161](https://github.com/hashicorp/terraform-provider-aws/issues/33161))
* resource/aws_quicksight_dashboard: Change `definition.*.parameter_declarations` to a set type, preventing persistent differences ([#33120](https://github.com/hashicorp/terraform-provider-aws/issues/33120))
* resource/aws_quicksight_dashboard: Fixed a bug that caused errors related to the `word_orientation` argument when using word cloud visuals. ([#33122](https://github.com/hashicorp/terraform-provider-aws/issues/33122))
* resource/aws_quicksight_dashboard: Skip setting `definition.*.parameter_declarations.*.*_parameter_declaration.static_values` when empty, preventing persistent differences. ([#33161](https://github.com/hashicorp/terraform-provider-aws/issues/33161))
* resource/aws_quicksight_template: Change `definition.*.parameter_declarations` to a set type, preventing persistent differences ([#33120](https://github.com/hashicorp/terraform-provider-aws/issues/33120))
* resource/aws_quicksight_template: Fixed a bug that caused errors related to the `word_orientation` argument when using word cloud visuals. ([#33122](https://github.com/hashicorp/terraform-provider-aws/issues/33122))
* resource/aws_quicksight_template: Skip setting `definition.*.parameter_declarations.*.*_parameter_declaration.static_values` when empty, preventing persistent differences. ([#33161](https://github.com/hashicorp/terraform-provider-aws/issues/33161))
* resource/aws_route53_zone: Skip disabling DNS SEC in unsupported partitions ([#33103](https://github.com/hashicorp/terraform-provider-aws/issues/33103))
* resource/aws_s3_object: Mark `acl` as Computed. This suppresses the diffs shown when migrating resources with no configured `acl` attribute value from v4.67.0 (or earlier) ([#33138](https://github.com/hashicorp/terraform-provider-aws/issues/33138))
* resource/aws_s3_object_copy: Mark `acl` as Computed. This suppresses the diffs shown when migrating resources with no configured `acl` attribute value from v4.67.0 (or earlier) ([#33138](https://github.com/hashicorp/terraform-provider-aws/issues/33138))
* resource/aws_securityhub_account: Remove default value (`SECURITY_CONTROL`) for `control_finding_generator` argument and mark as Computed ([#33095](https://github.com/hashicorp/terraform-provider-aws/issues/33095))

## 5.13.1 (August 18, 2023)

BUG FIXES:

* resource/aws_lambda_layer_version: Change `source_code_hash` back to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew). This fixes `doesn't support update` errors ([#33097](https://github.com/hashicorp/terraform-provider-aws/issues/33097))
* resource/aws_organizations_organization: Fix `current Organization ID (o-xxxxxxxxxx) does not match` errors on resource Read ([#33091](https://github.com/hashicorp/terraform-provider-aws/issues/33091))

## 5.13.0 (August 18, 2023)

FEATURES:

* **New Resource:** `aws_msk_cluster_policy` ([#32848](https://github.com/hashicorp/terraform-provider-aws/issues/32848))
* **New Resource:** `aws_opensearch_vpc_endpoint` ([#32435](https://github.com/hashicorp/terraform-provider-aws/issues/32435))
* **New Resource:** `aws_ram_sharing_with_organization` ([#25433](https://github.com/hashicorp/terraform-provider-aws/issues/25433))

ENHANCEMENTS:

* data-source/aws_imagebuilder_image_pipeline: Add `image_scanning_configuration` attribute ([#33005](https://github.com/hashicorp/terraform-provider-aws/issues/33005))
* data-source/aws_ram_resource_share: Add `resource_arns` attribute ([#22591](https://github.com/hashicorp/terraform-provider-aws/issues/22591))
* provider: Adds the `s3_us_east_1_regional_endpoint` attribute to support using the regional S3 API endpoint in `us-east-1`. ([#33024](https://github.com/hashicorp/terraform-provider-aws/issues/33024))
* resource/aws_appstream_fleet: Retry ConcurrentModificationException errors during creation ([#32958](https://github.com/hashicorp/terraform-provider-aws/issues/32958))
* resource/aws_dms_endpoint: Add `babelfish` as an `engine_name` option ([#32975](https://github.com/hashicorp/terraform-provider-aws/issues/32975))
* resource/aws_imagebuilder_image_pipeline: Add `image_scanning_configuration` configuration block ([#33005](https://github.com/hashicorp/terraform-provider-aws/issues/33005))
* resource/aws_lb: Changes to `security_groups` for Network Load Balancers force a new resource if either the old or new set of security group IDs is empty ([#32987](https://github.com/hashicorp/terraform-provider-aws/issues/32987))
* resource/aws_rds_global_cluster: Add plan-time validation of `global_cluster_identifier` ([#30996](https://github.com/hashicorp/terraform-provider-aws/issues/30996))

BUG FIXES:

* data-source/aws_ecr_repository: Correctly set `most_recent_image_tags` when only a single image is found ([#31757](https://github.com/hashicorp/terraform-provider-aws/issues/31757))
* resource/aws_budgets_budget_action: No longer times out when creating a non-triggered action ([#33015](https://github.com/hashicorp/terraform-provider-aws/issues/33015))
* resource/aws_cloudformation_stack: Marks `outputs` as Computed when there are potential changes. ([#33059](https://github.com/hashicorp/terraform-provider-aws/issues/33059))
* resource/aws_cloudwatch_event_rule: Fix ARN-based partner event bus rule ID parsing error ([#30293](https://github.com/hashicorp/terraform-provider-aws/issues/30293))
* resource/aws_ecr_registry_scanning_configuration: Correctly delete rules on resource Update ([#31449](https://github.com/hashicorp/terraform-provider-aws/issues/31449))
* resource/aws_lambda_layer_version: Fix bug causing new version to be created on every apply when `source_code_hash` is used but not changed ([#32535](https://github.com/hashicorp/terraform-provider-aws/issues/32535))
* resource/aws_lb_listener_certificate: Remove from state when listener not found ([#32412](https://github.com/hashicorp/terraform-provider-aws/issues/32412))
* resource/aws_organizations_organization: Ensure that the Organization ID specified in `terraform import` is the current Organization ([#31796](https://github.com/hashicorp/terraform-provider-aws/issues/31796))
* resource/aws_quicksight_analysis: Adjust max length of `definition.*.calculated_fields.*.expression` to 32000 characters ([#33012](https://github.com/hashicorp/terraform-provider-aws/issues/33012))
* resource/aws_quicksight_analysis: Convert `definition.*.calculated_fields` to a set type, preventing persistent differences ([#33040](https://github.com/hashicorp/terraform-provider-aws/issues/33040))
* resource/aws_quicksight_analysis: Convert `permissions` argument to TypeSet, preventing persistent differences ([#33023](https://github.com/hashicorp/terraform-provider-aws/issues/33023))
* resource/aws_quicksight_analysis: Enable `font_configuration` to be set for table header styles ([#33018](https://github.com/hashicorp/terraform-provider-aws/issues/33018))
* resource/aws_quicksight_analysis: Enable `font_configuration` to be set for table header styles ([#33018](https://github.com/hashicorp/terraform-provider-aws/issues/33018))
* resource/aws_quicksight_analysis: Enable `font_configuration` to be set for table header styles ([#33018](https://github.com/hashicorp/terraform-provider-aws/issues/33018))
* resource/aws_quicksight_analysis: Raise limit for maximum allowed `visuals` blocks per sheet to 50 ([#32856](https://github.com/hashicorp/terraform-provider-aws/issues/32856))
* resource/aws_quicksight_dashboard: Adjust max length of `definition.*.calculated_fields.*.expression` to 32000 characters ([#33012](https://github.com/hashicorp/terraform-provider-aws/issues/33012))
* resource/aws_quicksight_dashboard: Convert `definition.*.calculated_fields` to a set type, preventing persistent differences ([#33040](https://github.com/hashicorp/terraform-provider-aws/issues/33040))
* resource/aws_quicksight_dashboard: Convert `permissions` argument to TypeSet, preventing persistent differences ([#33023](https://github.com/hashicorp/terraform-provider-aws/issues/33023))
* resource/aws_quicksight_data_set: Change permission attribute type from TypeList to TypeSet ([#32984](https://github.com/hashicorp/terraform-provider-aws/issues/32984))
* resource/aws_quicksight_template: Adjust max items of `definition.*.calculated_fields` to 500 ([#33012](https://github.com/hashicorp/terraform-provider-aws/issues/33012))
* resource/aws_quicksight_template: Adjust max length of `definition.*.calculated_fields.*.expression` to 32000 characters ([#33012](https://github.com/hashicorp/terraform-provider-aws/issues/33012))
* resource/aws_quicksight_template: Convert `definition.*.calculated_fields` to a set type, preventing persistent differences ([#33040](https://github.com/hashicorp/terraform-provider-aws/issues/33040))
* resource/aws_quicksight_template: Convert `permissions` argument to TypeSet, preventing persistent differences ([#33023](https://github.com/hashicorp/terraform-provider-aws/issues/33023))
* resource/aws_s3_bucket_logging: Fix perpetual drift when `expected_bucket_owner` is configured ([#32989](https://github.com/hashicorp/terraform-provider-aws/issues/32989))
* resource/aws_sagemaker_domain: Fix validation on `s3_kms_key_id` in `sharing_settings` and `kms_key_id` ([#32661](https://github.com/hashicorp/terraform-provider-aws/issues/32661))
* resource/aws_subnet: Fix allowing IPv6 to be enabled in an update after initial creation with IPv4 only ([#32896](https://github.com/hashicorp/terraform-provider-aws/issues/32896))
* resource/aws_wafv2_web_acl: Adds `rule_group_reference_statement.rule_action_override.action_to_use.challenge` argument ([#31127](https://github.com/hashicorp/terraform-provider-aws/issues/31127))

## 5.12.0 (August 10, 2023)

NOTES:

* data-source/aws_codecatalyst_dev_environment: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#32886](https://github.com/hashicorp/terraform-provider-aws/issues/32886))
* resource/aws_codecatalyst_dev_environment: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#32366](https://github.com/hashicorp/terraform-provider-aws/issues/32366))
* resource/aws_codecatalyst_project: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#32883](https://github.com/hashicorp/terraform-provider-aws/issues/32883))
* resource/aws_codecatalyst_source_repository: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#32899](https://github.com/hashicorp/terraform-provider-aws/issues/32899))

FEATURES:

* **New Data Source:** `aws_codecatalyst_dev_environment` ([#32886](https://github.com/hashicorp/terraform-provider-aws/issues/32886))
* **New Data Source:** `aws_ec2_transit_gateway_route_table_routes` ([#30771](https://github.com/hashicorp/terraform-provider-aws/issues/30771))
* **New Data Source:** `aws_msk_vpc_connection` ([#31062](https://github.com/hashicorp/terraform-provider-aws/issues/31062))
* **New Resource:** `aws_cloudfront_continuous_deployment_policy` ([#32936](https://github.com/hashicorp/terraform-provider-aws/issues/32936))
* **New Resource:** `aws_codecatalyst_dev_environment` ([#32366](https://github.com/hashicorp/terraform-provider-aws/issues/32366))
* **New Resource:** `aws_codecatalyst_project` ([#32883](https://github.com/hashicorp/terraform-provider-aws/issues/32883))
* **New Resource:** `aws_codecatalyst_source_repository` ([#32899](https://github.com/hashicorp/terraform-provider-aws/issues/32899))
* **New Resource:** `aws_msk_vpc_connection` ([#31062](https://github.com/hashicorp/terraform-provider-aws/issues/31062))

ENHANCEMENTS:

* data-source/aws_instance: Add `metadata_options.http_protocol_ipv6` attribute ([#32759](https://github.com/hashicorp/terraform-provider-aws/issues/32759))
* data-source/aws_rds_cluster: Add `db_system_id` attribute ([#32846](https://github.com/hashicorp/terraform-provider-aws/issues/32846))
* provider: Support `il-central-1` as a valid AWS Region ([#32878](https://github.com/hashicorp/terraform-provider-aws/issues/32878))
* resource/aws_autoscaling_group: Add `ignore_failed_scaling_activities` argument ([#32914](https://github.com/hashicorp/terraform-provider-aws/issues/32914))
* resource/aws_cloudfront_distribution: Add `continuous_deployment_policy_id` and `staging` arguments to support continuous deployments ([#32936](https://github.com/hashicorp/terraform-provider-aws/issues/32936))
* resource/aws_cloudwatch_composite_alarm: Add `actions_suppressor` configuration block ([#32751](https://github.com/hashicorp/terraform-provider-aws/issues/32751))
* resource/aws_cloudwatch_events_target: Add `sagemaker_pipeline_target` argument ([#32882](https://github.com/hashicorp/terraform-provider-aws/issues/32882))
* resource/aws_fms_admin_account: Add configurable timeouts ([#32860](https://github.com/hashicorp/terraform-provider-aws/issues/32860))
* resource/aws_glue_crawler: Add `hudi_target` argument ([#32898](https://github.com/hashicorp/terraform-provider-aws/issues/32898))
* resource/aws_instance: Add `http_protocol_ipv6` attribute to `metadata_options` configuration block ([#32759](https://github.com/hashicorp/terraform-provider-aws/issues/32759))
* resource/aws_lambda_event_source_mapping: Increased the maximum number of filters to 10 ([#32890](https://github.com/hashicorp/terraform-provider-aws/issues/32890))
* resource/aws_msk_broker: Add `bootstrap_brokers_vpc_connectivity_sasl_iam`, `bootstrap_brokers_vpc_connectivity_sasl_scram` and `bootstrap_brokers_vpc_connectivity_tls` attributes ([#31062](https://github.com/hashicorp/terraform-provider-aws/issues/31062))
* resource/aws_msk_broker: Add `vpc_connectivity` attribute to the `broker_node_group_info.connectivity_info` configuration block ([#31062](https://github.com/hashicorp/terraform-provider-aws/issues/31062))
* resource/aws_rds_cluster: Add `db_system_id` argument to support RDS Custom engine types ([#32846](https://github.com/hashicorp/terraform-provider-aws/issues/32846))
* resource/aws_rds_cluster_instance: Add `custom_iam_instance_profile` argument to allow RDS Custom users to specify an IAM Instance Profile for the RDS Cluster Instance ([#32846](https://github.com/hashicorp/terraform-provider-aws/issues/32846))
* resource/aws_rds_cluster_instance: Update `engine` plan-time validation to allow for RDS Custom engine types ([#32846](https://github.com/hashicorp/terraform-provider-aws/issues/32846))

BUG FIXES:

* data-source/aws_vpclattice_service: Avoid listing tags when the service has been shared to the current account via AWS Resource Access Manager (RAM) ([#32939](https://github.com/hashicorp/terraform-provider-aws/issues/32939))
* data-source/aws_vpclattice_service_network: Avoid listing tags when the service network has been shared to the current account via AWS Resource Access Manager (RAM) ([#32939](https://github.com/hashicorp/terraform-provider-aws/issues/32939))
* resource/aws_appstream_fleet: Increased upper limit of `max_user_duration_in_seconds` to 432000 ([#32933](https://github.com/hashicorp/terraform-provider-aws/issues/32933))
* resource/aws_cloudfront_distribution: Don't call `UpdateDistribution` API if only tags are updated ([#32865](https://github.com/hashicorp/terraform-provider-aws/issues/32865))
* resource/aws_db_instance: Fix crash creating resource with empty `restore_to_point_in_time` configuration block ([#32928](https://github.com/hashicorp/terraform-provider-aws/issues/32928))
* resource/aws_emr_cluster: Fix to allow empty `args` for `bootstrap_action` ([#32956](https://github.com/hashicorp/terraform-provider-aws/issues/32956))
* resource/aws_emr_instance_fleet: Fix fleet deletion failing for terminated clusters ([#32866](https://github.com/hashicorp/terraform-provider-aws/issues/32866))
* resource/aws_fms_policy: Prevent erroneous diffs on `security_service_policy_data.managed_service_data` ([#32860](https://github.com/hashicorp/terraform-provider-aws/issues/32860))
* resource/aws_instance: Fix `InvalidParameterCombination: Network interfaces and an instance-level security groups may not be specified on the same request` errors creating Instances with `subnet_id` configured and `launch_template` referencing an `aws_launch_template` with configured `vpc_security_group_ids` ([#32854](https://github.com/hashicorp/terraform-provider-aws/issues/32854))
* resource/aws_lb: Fix to avoid creating a load balancer with same name as an existing load balancer ([#32941](https://github.com/hashicorp/terraform-provider-aws/issues/32941))

## 5.11.0 (August  3, 2023)

FEATURES:

* **New Resource:** `aws_sagemaker_pipeline` ([#32527](https://github.com/hashicorp/terraform-provider-aws/issues/32527))

ENHANCEMENTS:

* data-source/aws_cloudtrail_service_account: Add service account ID for `il-central-1` AWS Region ([#32840](https://github.com/hashicorp/terraform-provider-aws/issues/32840))
* data-source/aws_db_cluster_snapshot: Add `tags` argument ([#31602](https://github.com/hashicorp/terraform-provider-aws/issues/31602))
* data-source/aws_db_instance: Add ability to filter by `tags` ([#32740](https://github.com/hashicorp/terraform-provider-aws/issues/32740))
* data-source/aws_db_instances: Add ability to filter by `tags` ([#32740](https://github.com/hashicorp/terraform-provider-aws/issues/32740))
* data-source/aws_db_snapshot: Add `tags` argument ([#31600](https://github.com/hashicorp/terraform-provider-aws/issues/31600))
* data-source/aws_elb_hosted_zone_id: Add hosted zone ID for `il-central-1` AWS Region ([#32840](https://github.com/hashicorp/terraform-provider-aws/issues/32840))
* data-source/aws_lb_hosted_zone_id: Add hosted zone IDs for `il-central-1` AWS Region ([#32840](https://github.com/hashicorp/terraform-provider-aws/issues/32840))
* data-source/aws_s3_bucket: Add hosted zone ID for `il-central-1` AWS Region ([#32840](https://github.com/hashicorp/terraform-provider-aws/issues/32840))
* data-source/aws_vpclattice_service: Add ability to find by `name` ([#32177](https://github.com/hashicorp/terraform-provider-aws/issues/32177))
* resource/aws_finspace_kx_cluster: Adjusted `savedown_storage_configuration.size` minimum value to `10` GB. ([#32800](https://github.com/hashicorp/terraform-provider-aws/issues/32800))
* resource/aws_lambda_function: Add support for `python3.11` `runtime` value ([#32729](https://github.com/hashicorp/terraform-provider-aws/issues/32729))
* resource/aws_lambda_layer_version: Add support for `python3.11` `compatible_runtimes` value ([#32729](https://github.com/hashicorp/terraform-provider-aws/issues/32729))
* resource/aws_networkfirewall_rule_group: Add support for `REJECT` action in stateful rule actions ([#32746](https://github.com/hashicorp/terraform-provider-aws/issues/32746))
* resource/aws_route_table: Allow an existing local route to be adopted or imported and the target to be updated ([#32794](https://github.com/hashicorp/terraform-provider-aws/issues/32794))
* resource/aws_sagemaker_endpoint: Add `deployment_config.rolling_update_policy` argument ([#32418](https://github.com/hashicorp/terraform-provider-aws/issues/32418))
* resource/aws_sagemaker_endpoint: Make `deployment_config.blue_green_update_policy` optional ([#32418](https://github.com/hashicorp/terraform-provider-aws/issues/32418))

BUG FIXES:

* data-source/aws_ecs_task_execution: Fixed bug that incorrectly mapped the value of `container_overrides.memory` to `container_overrides.memory_reservation` ([#32793](https://github.com/hashicorp/terraform-provider-aws/issues/32793))
* resource/aws_db_instance_automated_backups_replication: Fix `unexpected state 'Pending'` errors on resource Create ([#31600](https://github.com/hashicorp/terraform-provider-aws/issues/31600))
* resource/aws_ec2_transit_gateway_vpc_attachment: Change `transit_gateway_default_route_table_association` and `transit_gateway_default_route_table_propagation` to Computed ([#32821](https://github.com/hashicorp/terraform-provider-aws/issues/32821))
* resource/aws_emr_studio_session_mapping: Fix `InvalidRequestException: IdentityId is invalid` errors reading resources created with `identity_name` ([#32416](https://github.com/hashicorp/terraform-provider-aws/issues/32416))
* resource/aws_quicksight_analysis: Fix an error related to setting the value for `definition.sheets.visuals.insight_visual.insight_configuration.computation` ([#32791](https://github.com/hashicorp/terraform-provider-aws/issues/32791))
* resource/aws_quicksight_analysis: Fixed a bug that incorrectly determined the valid `select_all_options` values for `custom_filter_configuration`, `custom_filter_list_configuration`, `filter_list_configuration`, `numeric_equality_filter`, and `numeric_range_filter` ([#32822](https://github.com/hashicorp/terraform-provider-aws/issues/32822))
* resource/aws_quicksight_dashboard: Fix an error related to setting the value for `definition.sheets.visuals.insight_visual.insight_configuration.computation` ([#32791](https://github.com/hashicorp/terraform-provider-aws/issues/32791))
* resource/aws_quicksight_template: Fix an error related to setting the value for `definition.sheets.visuals.insight_visual.insight_configuration.computation` ([#32791](https://github.com/hashicorp/terraform-provider-aws/issues/32791))
* resource/aws_quicksight_template: Fixed a bug that incorrectly determined the valid `select_all_options` values for `custom_filter_configuration`, `custom_filter_list_configuration`, `filter_list_configuration`, `numeric_equality_filter`, and `numeric_range_filter` ([#32822](https://github.com/hashicorp/terraform-provider-aws/issues/32822))
* resource/aws_sfn_state_machine: Fix `Provider produced inconsistent final plan` errors for `publish` ([#32844](https://github.com/hashicorp/terraform-provider-aws/issues/32844))

## 5.10.0 (July 27, 2023)

FEATURES:

* **New Resource:** `aws_iam_security_token_service_preferences` ([#32091](https://github.com/hashicorp/terraform-provider-aws/issues/32091))

ENHANCEMENTS:

* data-source/aws_nat_gateway: Add `secondary_allocation_ids`, `secondary_private_ip_addresses` and `secondary_private_ip_address_count` attributes ([#31778](https://github.com/hashicorp/terraform-provider-aws/issues/31778))
* data-source/aws_transfer_server: Add `structured_log_destinations` attribute ([#32654](https://github.com/hashicorp/terraform-provider-aws/issues/32654))
* resource/aws_batch_compute_environment: `compute_resources.allocation_strategy`, `compute_resources.bid_percentage`, `compute_resources.ec2_configuration.image_id_override`, `compute_resources.ec2_configuration.image_type`, `compute_resources.ec2_key_pair`, `compute_resources.image_id`, `compute_resources.instance_role`, `compute_resources.launch_template.launch_template_id`
, `compute_resources.launch_template.launch_template_name`, `compute_resources.tags` and `compute_resources.type` can now be updated in-place ([#30438](https://github.com/hashicorp/terraform-provider-aws/issues/30438))
* resource/aws_glue_job: Add `command.runtime` attribute ([#32528](https://github.com/hashicorp/terraform-provider-aws/issues/32528))
* resource/aws_grafana_workspace: Allow `grafana_version` to be updated in-place ([#32679](https://github.com/hashicorp/terraform-provider-aws/issues/32679))
* resource/aws_kms_grant: Allow usage of service principal as grantee and revoker ([#32595](https://github.com/hashicorp/terraform-provider-aws/issues/32595))
* resource/aws_medialive_channel: Adds schemas for `caption_descriptions`, `global_configuration`, `motion_graphics_configuration`, and `nielsen_configuration` support to `encoder settings` ([#32233](https://github.com/hashicorp/terraform-provider-aws/issues/32233))
* resource/aws_nat_gateway: Add `secondary_allocation_ids`, `secondary_private_ip_addresses` and `secondary_private_ip_address_count` arguments ([#31778](https://github.com/hashicorp/terraform-provider-aws/issues/31778))
* resource/aws_nat_gateway: Add configurable timeouts ([#31778](https://github.com/hashicorp/terraform-provider-aws/issues/31778))
* resource/aws_networkfirewall_firewall_policy: Add `firewall_policy.policy_variables` configuration block to support Suricata HOME_NET variable override ([#32400](https://github.com/hashicorp/terraform-provider-aws/issues/32400))
* resource/aws_sagemaker_domain: Add `default_user_settings.canvas_app_settings.workspace_settings` attribute ([#32526](https://github.com/hashicorp/terraform-provider-aws/issues/32526))
* resource/aws_sagemaker_user_profile: Add `user_settings.canvas_app_settings.workspace_settings` attribute ([#32526](https://github.com/hashicorp/terraform-provider-aws/issues/32526))
* resource/aws_transfer_server: Add `structured_log_destinations` argument ([#32654](https://github.com/hashicorp/terraform-provider-aws/issues/32654))

BUG FIXES:

* resource/aws_account_primary_contact: Correct plan-time validation of `phone_number` ([#32715](https://github.com/hashicorp/terraform-provider-aws/issues/32715))
* resource/aws_apigatewayv2_authorizer: Skip setting authorizer TTL when there are no identity sources ([#32629](https://github.com/hashicorp/terraform-provider-aws/issues/32629))
* resource/aws_elasticache_parameter_group: Remove from state on resource Read if deleted outside of Terraform ([#32669](https://github.com/hashicorp/terraform-provider-aws/issues/32669))
* resource/aws_elasticsearch_domain: Omit `ebs_options.throughput` and `ebs_options.iops` for unsupported volume types ([#32659](https://github.com/hashicorp/terraform-provider-aws/issues/32659))
* resource/aws_finspace_kx_cluster: `database.cache_configurations.db_paths` argument is now optional ([#32579](https://github.com/hashicorp/terraform-provider-aws/issues/32579))
* resource/aws_finspace_kx_cluster: `database.cache_configurations` argument is now optional ([#32579](https://github.com/hashicorp/terraform-provider-aws/issues/32579))
* resource/aws_lambda_invocation: Fix plan failing with deferred input values ([#32706](https://github.com/hashicorp/terraform-provider-aws/issues/32706))
* resource/aws_lightsail_domain_entry: Add support for `AAAA` `type` value ([#32664](https://github.com/hashicorp/terraform-provider-aws/issues/32664))
* resource/aws_opensearch_domain: Correctly handle `off_peak_window_options.off_peak_window.window_start_time` value of `00:00` ([#32716](https://github.com/hashicorp/terraform-provider-aws/issues/32716))
* resource/aws_quicksight_analysis: Fix exception thrown when setting the value for `definition.sheets.visuals.pie_chart_visual.chart_configuration.data_labels.measure_label_visibility` ([#32668](https://github.com/hashicorp/terraform-provider-aws/issues/32668))
* resource/aws_quicksight_analysis: Grid layout `optimized_view_port_width` argument changed to Optional ([#32644](https://github.com/hashicorp/terraform-provider-aws/issues/32644))
* resource/aws_quicksight_dashboard: Fix exception thrown when setting the value for `definition.sheets.visuals.pie_chart_visual.chart_configuration.data_labels.measure_label_visibility` ([#32668](https://github.com/hashicorp/terraform-provider-aws/issues/32668))
* resource/aws_quicksight_dashboard: Grid layout `optimized_view_port_width` argument changed to Optional ([#32644](https://github.com/hashicorp/terraform-provider-aws/issues/32644))
* resource/aws_quicksight_template: Fix exception thrown when setting the value for `definition.sheets.visuals.pie_chart_visual.chart_configuration.data_labels.measure_label_visibility` ([#32668](https://github.com/hashicorp/terraform-provider-aws/issues/32668))
* resource/aws_quicksight_template: Grid layout `optimized_view_port_width` argument changed to Optional ([#32644](https://github.com/hashicorp/terraform-provider-aws/issues/32644))
* resource/aws_vpclattice_access_log_subscription: Avoid recreating resource when passing a non-wildcard CloudWatch Logs log group ARN as `destination_arn` ([#32186](https://github.com/hashicorp/terraform-provider-aws/issues/32186))
* resource/aws_vpclattice_access_log_subscription: Avoid recreating resource when passing an ARN as `resource_identifier` ([#32186](https://github.com/hashicorp/terraform-provider-aws/issues/32186))
* resource/aws_vpclattice_service_network_service_association: Avoid recreating resource when passing an ARN as `service_identifier` or `service_network_identifier` ([#32658](https://github.com/hashicorp/terraform-provider-aws/issues/32658))
* resource/aws_vpclattice_service_network_vpc_association: Avoid recreating resource when passing an ARN as `service_network_identifier` ([#32658](https://github.com/hashicorp/terraform-provider-aws/issues/32658))

## 5.9.0 (July 20, 2023)

FEATURES:

* **New Resource:** `aws_workspaces_connection_alias` ([#32482](https://github.com/hashicorp/terraform-provider-aws/issues/32482))

ENHANCEMENTS:

* data-source/aws_appmesh_gateway_route: Add `path` to the `spec.http_route.action.rewrite` and `spec.http2_route.action.rewrite` configuration blocks ([#32449](https://github.com/hashicorp/terraform-provider-aws/issues/32449))
* data-source/aws_db_instance: Add `max_allocated_storage` attribute ([#32477](https://github.com/hashicorp/terraform-provider-aws/issues/32477))
* data-source/aws_ec2_host: Add `asset_id` attribute ([#32388](https://github.com/hashicorp/terraform-provider-aws/issues/32388))
* resource/aws_appmesh_gateway_route: Add `path` to the `spec.http_route.action.rewrite` and `spec.http2_route.action.rewrite` configuration blocks ([#32449](https://github.com/hashicorp/terraform-provider-aws/issues/32449))
* resource/aws_cloudformation_stack_set_instance: Added the `stack_instance_summaries` attribute to track all account and stack IDs for deployments to organizational units. ([#24523](https://github.com/hashicorp/terraform-provider-aws/issues/24523))
* resource/aws_cloudformation_stack_set_instance: Changes to `deployment_targets` now force a new resource. ([#24523](https://github.com/hashicorp/terraform-provider-aws/issues/24523))
* resource/aws_connect_queue: add delete function ([#32538](https://github.com/hashicorp/terraform-provider-aws/issues/32538))
* resource/aws_connect_routing_profile: add delete function ([#32540](https://github.com/hashicorp/terraform-provider-aws/issues/32540))
* resource/aws_db_instance: Add `backup_target` attribute ([#32609](https://github.com/hashicorp/terraform-provider-aws/issues/32609))
* resource/aws_ec2_host: Add `asset_id` argument ([#32388](https://github.com/hashicorp/terraform-provider-aws/issues/32388))
* resource/aws_ec2_traffic_mirror_filter_rule: Fix crash when updating `rule_number` ([#32594](https://github.com/hashicorp/terraform-provider-aws/issues/32594))
* resource/aws_lightsail_key_pair: Add `tags` attribute ([#32606](https://github.com/hashicorp/terraform-provider-aws/issues/32606))
* resource/aws_signer_signing_profile: Add `signing_material` attribute. ([#32414](https://github.com/hashicorp/terraform-provider-aws/issues/32414))
* resource/aws_signer_signing_profile: Update `platform_id` validation. ([#32414](https://github.com/hashicorp/terraform-provider-aws/issues/32414))
* resource/aws_wafv2_web_acl: Add `association_config` argument ([#31668](https://github.com/hashicorp/terraform-provider-aws/issues/31668))

BUG FIXES:

* data-source/aws_dms_replication_instance: Fixed bug that caused `replication_instance_private_ips`, `replication_instance_public_ips`, and `vpc_security_group_ids` to always return `null` ([#32551](https://github.com/hashicorp/terraform-provider-aws/issues/32551))
* data-source/aws_mq_broker: Fix `setting user: Invalid address to set` errors ([#32593](https://github.com/hashicorp/terraform-provider-aws/issues/32593))
* data-source/aws_vpc_endpoint: Add `dns_options.private_dns_only_for_inbound_resolver_endpoint` ([#32517](https://github.com/hashicorp/terraform-provider-aws/issues/32517))
* resource/aws_appflow_flow: Fix tasks not updating properly due to empty task being processed ([#26614](https://github.com/hashicorp/terraform-provider-aws/issues/26614))
* resource/aws_cloudformation_stack_set_instance: Fix error when deploying to organizational units with no accounts. ([#24523](https://github.com/hashicorp/terraform-provider-aws/issues/24523))
* resource/aws_cognito_user_pool: Suppress diff when `schema.string_attribute_constraints` is omitted for `String` attribute types ([#32445](https://github.com/hashicorp/terraform-provider-aws/issues/32445))
* resource/aws_config_config_rule: Prevent crash from unhandled read error ([#32520](https://github.com/hashicorp/terraform-provider-aws/issues/32520))
* resource/aws_datasync_agent: Prevent persistent diffs when `private_link_endpoint` is not explicitly configured. ([#32546](https://github.com/hashicorp/terraform-provider-aws/issues/32546))
* resource/aws_globalaccelerator_custom_routing_endpoint_group: Respect configured `endpoint_group_region` value on resource Create ([#32393](https://github.com/hashicorp/terraform-provider-aws/issues/32393))
* resource/aws_pipes_pipe: Fix `Error: setting target_parameters: Invalid address to set` errors when creating pipes with ecs task targets ([#32432](https://github.com/hashicorp/terraform-provider-aws/issues/32432))
* resource/aws_pipes_pipe: Fix `ValidationException` errors when updating pipe ([#32622](https://github.com/hashicorp/terraform-provider-aws/issues/32622))
* resource/aws_quicksight_analysis: Correctly expand comparison method ([#32285](https://github.com/hashicorp/terraform-provider-aws/issues/32285))
* resource/aws_quicksight_folder: Fix misidentification of parent folder at grandchild level or deeper ([#32592](https://github.com/hashicorp/terraform-provider-aws/issues/32592))
* resource/aws_quicksight_group_membership: Allow non `default` value for namespace ([#32494](https://github.com/hashicorp/terraform-provider-aws/issues/32494))
* resource/aws_route53_cidr_location: Fix `Value Conversion Error` errors ([#32596](https://github.com/hashicorp/terraform-provider-aws/issues/32596))
* resource/aws_wafv2_web_acl: Fixed error handling `response_inspection` parameters ([#31111](https://github.com/hashicorp/terraform-provider-aws/issues/31111))

## 5.8.0 (July 13, 2023)

ENHANCEMENTS:

* data-source/aws_ssm_parameter: Add `insecure_value` attribute ([#30817](https://github.com/hashicorp/terraform-provider-aws/issues/30817))
* resource/aws_fms_policy: Add `policy_option` attribute for `security_service_policy_data` block ([#25362](https://github.com/hashicorp/terraform-provider-aws/issues/25362))
* resource/aws_iam_virtual_mfa_device: Add `enable_date` and `user_name` attributes ([#32462](https://github.com/hashicorp/terraform-provider-aws/issues/32462))

BUG FIXES:

* resource/aws_config_config_rule: Prevent crash on nil describe output ([#32439](https://github.com/hashicorp/terraform-provider-aws/issues/32439))
* resource/aws_mq_broker: default `replication_user` to `false` ([#32454](https://github.com/hashicorp/terraform-provider-aws/issues/32454))
* resource/aws_quicksight_analysis: Fix exception thrown when specifying `definition.sheets.visuals.bar_chart_visual.chart_configuration.category_axis.scrollbar_options.visible_range` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))
* resource/aws_quicksight_analysis: Fix exception thrown when specifying `definition.sheets.visuals.pivot_table_visual.chart_configuration.field_options.selected_field_options.visibility` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))
* resource/aws_quicksight_analysis: Fix exception thrown when specifying `definition.sheets.visuals.pivot_table_visual.chart_configuration.field_wells.pivot_table_aggregated_field_wells.rows` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))
* resource/aws_quicksight_dashboard: Fix exception thrown when specifying `definition.sheets.visuals.bar_chart_visual.chart_configuration.category_axis.scrollbar_options.visible_range` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))
* resource/aws_quicksight_dashboard: Fix exception thrown when specifying `definition.sheets.visuals.pivot_table_visual.chart_configuration.field_options.selected_field_options.visibility` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))
* resource/aws_quicksight_dashboard: Fix exception thrown when specifying `definition.sheets.visuals.pivot_table_visual.chart_configuration.field_wells.pivot_table_aggregated_field_wells.rows` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))
* resource/aws_quicksight_template: Fix exception thrown when specifying `definition.sheets.visuals.bar_chart_visual.chart_configuration.category_axis.scrollbar_options.visible_range` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))
* resource/aws_quicksight_template: Fix exception thrown when specifying `definition.sheets.visuals.pivot_table_visual.chart_configuration.field_options.selected_field_options.visibility` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))
* resource/aws_quicksight_template: Fix exception thrown when specifying `definition.sheets.visuals.pivot_table_visual.chart_configuration.field_wells.pivot_table_aggregated_field_wells.rows` ([#32464](https://github.com/hashicorp/terraform-provider-aws/issues/32464))

## 5.7.0 (July  7, 2023)

FEATURES:

* **New Data Source:** `aws_opensearchserverless_security_config` ([#32321](https://github.com/hashicorp/terraform-provider-aws/issues/32321))
* **New Data Source:** `aws_opensearchserverless_security_policy` ([#32226](https://github.com/hashicorp/terraform-provider-aws/issues/32226))
* **New Data Source:** `aws_opensearchserverless_vpc_endpoint` ([#32276](https://github.com/hashicorp/terraform-provider-aws/issues/32276))
* **New Resource:** `aws_cleanrooms_collaboration` ([#31680](https://github.com/hashicorp/terraform-provider-aws/issues/31680))

ENHANCEMENTS:

* resource/aws_aws_keyspaces_table: Add `client_side_timestamps` configuration block ([#32339](https://github.com/hashicorp/terraform-provider-aws/issues/32339))
* resource/aws_glue_catalog_database: Add `target_database.region` argument ([#32283](https://github.com/hashicorp/terraform-provider-aws/issues/32283))
* resource/aws_glue_crawler: Add `iceberg_target` configuration block ([#32332](https://github.com/hashicorp/terraform-provider-aws/issues/32332))
* resource/aws_internetmonitor_monitor: Add `health_events_config` configuration block ([#32343](https://github.com/hashicorp/terraform-provider-aws/issues/32343))
* resource/aws_lambda_function: Support `code_signing_config_arn` in the `ap-east-1` AWS Region ([#32327](https://github.com/hashicorp/terraform-provider-aws/issues/32327))
* resource/aws_qldb_stream: Add configurable Create and Delete timeouts ([#32345](https://github.com/hashicorp/terraform-provider-aws/issues/32345))
* resource/aws_service_discovery_private_dns_namespace: Allow `description` to be updated in-place ([#32342](https://github.com/hashicorp/terraform-provider-aws/issues/32342))
* resource/aws_service_discovery_public_dns_namespace: Allow `description` to be updated in-place ([#32342](https://github.com/hashicorp/terraform-provider-aws/issues/32342))
* resource/aws_timestreamwrite_table: Add `schema` configuration block ([#32354](https://github.com/hashicorp/terraform-provider-aws/issues/32354))

BUG FIXES:

* provider: Correctly handle `forbidden_account_ids` ([#32352](https://github.com/hashicorp/terraform-provider-aws/issues/32352))
* resource/aws_kms_external_key: Correctly remove all tags ([#32371](https://github.com/hashicorp/terraform-provider-aws/issues/32371))
* resource/aws_kms_key: Correctly remove all tags ([#32371](https://github.com/hashicorp/terraform-provider-aws/issues/32371))
* resource/aws_kms_replica_external_key: Correctly remove all tags ([#32371](https://github.com/hashicorp/terraform-provider-aws/issues/32371))
* resource/aws_kms_replica_key: Correctly remove all tags ([#32371](https://github.com/hashicorp/terraform-provider-aws/issues/32371))
* resource/aws_secretsmanager_secret_rotation: Fix `InvalidParameterException: You cannot specify both rotation frequency and schedule expression together` errors on resource Update ([#31915](https://github.com/hashicorp/terraform-provider-aws/issues/31915))
* resource/aws_ssm_parameter: Skip Update if only `overwrite` parameter changes ([#32372](https://github.com/hashicorp/terraform-provider-aws/issues/32372))
* resource/aws_vpc_endpoint: Fix `InvalidParameter: PrivateDnsOnlyForInboundResolverEndpoint not supported for this service` errors creating S3 _Interface_ VPC endpoints ([#32355](https://github.com/hashicorp/terraform-provider-aws/issues/32355))

## 5.6.2 (June 30, 2023)

BUG FIXES:

* resource/aws_s3_bucket: Fix `InvalidArgument: Invalid attribute name specified` errors when listing S3 Bucket objects, caused by an [AWS SDK for Go regression](https://github.com/aws/aws-sdk-go/issues/4897) ([#32317](https://github.com/hashicorp/terraform-provider-aws/issues/32317))

## 5.6.1 (June 30, 2023)

BUG FIXES:

* provider: Prevent resource recreation if `tags` or `tags_all` are updated ([#32297](https://github.com/hashicorp/terraform-provider-aws/issues/32297))

## 5.6.0 (June 29, 2023)

FEATURES:

* **New Data Source:** `aws_opensearchserverless_access_policy` ([#32231](https://github.com/hashicorp/terraform-provider-aws/issues/32231))
* **New Data Source:** `aws_opensearchserverless_collection` ([#32247](https://github.com/hashicorp/terraform-provider-aws/issues/32247))
* **New Data Source:** `aws_sfn_alias` ([#32176](https://github.com/hashicorp/terraform-provider-aws/issues/32176))
* **New Data Source:** `aws_sfn_state_machine_versions` ([#32176](https://github.com/hashicorp/terraform-provider-aws/issues/32176))
* **New Resource:** `aws_ec2_instance_connect_endpoint` ([#31858](https://github.com/hashicorp/terraform-provider-aws/issues/31858))
* **New Resource:** `aws_sfn_alias` ([#32176](https://github.com/hashicorp/terraform-provider-aws/issues/32176))
* **New Resource:** `aws_transfer_agreement` ([#32203](https://github.com/hashicorp/terraform-provider-aws/issues/32203))
* **New Resource:** `aws_transfer_certificate` ([#32203](https://github.com/hashicorp/terraform-provider-aws/issues/32203))
* **New Resource:** `aws_transfer_connector` ([#32203](https://github.com/hashicorp/terraform-provider-aws/issues/32203))
* **New Resource:** `aws_transfer_profile` ([#32203](https://github.com/hashicorp/terraform-provider-aws/issues/32203))

ENHANCEMENTS:

* resource/aws_batch_compute_environment: Add `placement_group` attribute to the `compute_resources` configuration block ([#32200](https://github.com/hashicorp/terraform-provider-aws/issues/32200))
* resource/aws_emrserverless_application: Do not recreate the resource if `release_label` changes ([#32278](https://github.com/hashicorp/terraform-provider-aws/issues/32278))
* resource/aws_fis_experiment_template: Add `log_configuration` configuration block ([#32102](https://github.com/hashicorp/terraform-provider-aws/issues/32102))
* resource/aws_fis_experiment_template: Add `parameters` attribute to the `target` configuration block ([#32160](https://github.com/hashicorp/terraform-provider-aws/issues/32160))
* resource/aws_fis_experiment_template: Add support for `Pods` and `Tasks` to `action.*.target` ([#32152](https://github.com/hashicorp/terraform-provider-aws/issues/32152))
* resource/aws_lambda_event_source_mapping: The `queues` argument has changed from a set to a list with a maximum of one element. ([#31931](https://github.com/hashicorp/terraform-provider-aws/issues/31931))
* resource/aws_pipes_pipe: Add `activemq_broker_parameters`, `dynamodb_stream_parameters`, `kinesis_stream_parameters`, `managed_streaming_kafka_parameters`, `rabbitmq_broker_parameters`, `self_managed_kafka_parameters` and `sqs_queue_parameters` attributes to the `source_parameters` configuration block. NOTE: Because we cannot easily test all this functionality, it is best effort and we ask for community help in testing ([#31607](https://github.com/hashicorp/terraform-provider-aws/issues/31607))
* resource/aws_pipes_pipe: Add `batch_job_parameters`, `cloudwatch_logs_parameters`, `ecs_task_parameters`, `eventbridge_event_bus_parameters`, `http_parameters`, `kinesis_stream_parameters`, `lambda_function_parameters`, `redshift_data_parameters`, `sagemaker_pipeline_parameters`, `sqs_queue_parameters` and `step_function_state_machine_parameters` attributes to the `target_parameters` configuration block. NOTE: Because we cannot easily test all this functionality, it is best effort and we ask for community help in testing ([#31607](https://github.com/hashicorp/terraform-provider-aws/issues/31607))
* resource/aws_pipes_pipe: Add `enrichment_parameters` argument ([#31607](https://github.com/hashicorp/terraform-provider-aws/issues/31607))
* resource/aws_resourcegroups_group: `resource_query` no longer conflicts with `configuration` ([#30242](https://github.com/hashicorp/terraform-provider-aws/issues/30242))
* resource/aws_s3_bucket_logging: Retry on empty read of logging config ([#30916](https://github.com/hashicorp/terraform-provider-aws/issues/30916))
* resource/aws_sfn_state_machine: Add `description`, `publish`, `revision_id`, `state_machine_version_arn` and `version_description` attributes ([#32176](https://github.com/hashicorp/terraform-provider-aws/issues/32176))

BUG FIXES:

* resource/aws_db_instance: Fix resource Create returning instances not in the `available` state when `identifier_prefix` is specified ([#32287](https://github.com/hashicorp/terraform-provider-aws/issues/32287))
* resource/aws_resourcegroups_resource: Fix crash when resource Create fails ([#30242](https://github.com/hashicorp/terraform-provider-aws/issues/30242))
* resource/aws_route: Fix `reading Route in Route Table (rtb-1234abcd) with destination (1.2.3.4/5): couldn't find resource` errors when reading new resource ([#32196](https://github.com/hashicorp/terraform-provider-aws/issues/32196))
* resource/aws_vpc_security_group_egress_rule: `security_group_id` is Required ([#32148](https://github.com/hashicorp/terraform-provider-aws/issues/32148))
* resource/aws_vpc_security_group_ingress_rule: `security_group_id` is Required ([#32148](https://github.com/hashicorp/terraform-provider-aws/issues/32148))

## 5.5.0 (June 23, 2023)

NOTES:

* provider: Updates to Go 1.20, the last release that will run on any release of Windows 7, 8, Server 2008 and Server 2012. A future release will update to Go 1.21, and these platforms will no longer be supported. ([#32108](https://github.com/hashicorp/terraform-provider-aws/issues/32108))
* provider: Updates to Go 1.20, the last release that will run on macOS 10.13 High Sierra or 10.14 Mojave. A future release will update to Go 1.21, and these platforms will no longer be supported. ([#32108](https://github.com/hashicorp/terraform-provider-aws/issues/32108))
* provider: Updates to Go 1.20. The provider will now notice the `trust-ad` option in `/etc/resolv.conf` and, if set, will set the "authentic data" option in outgoing DNS requests in order to better match the behavior of the GNU libc resolver. ([#32108](https://github.com/hashicorp/terraform-provider-aws/issues/32108))

FEATURES:

* **New Data Source:** `aws_sesv2_email_identity` ([#32026](https://github.com/hashicorp/terraform-provider-aws/issues/32026))
* **New Data Source:** `aws_sesv2_email_identity_mail_from_attributes` ([#32026](https://github.com/hashicorp/terraform-provider-aws/issues/32026))
* **New Resource:** `aws_chimesdkvoice_sip_rule` ([#32070](https://github.com/hashicorp/terraform-provider-aws/issues/32070))
* **New Resource:** `aws_organizations_resource_policy` ([#32056](https://github.com/hashicorp/terraform-provider-aws/issues/32056))

ENHANCEMENTS:

* data-source/aws_organizations_organization: Return the full set of attributes when running as a [delegated administrator for AWS Organizations](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_delegate_policies.html) ([#32056](https://github.com/hashicorp/terraform-provider-aws/issues/32056))
* provider: Mask all sensitive values that appear when `TF_LOG` level is `TRACE` ([#32174](https://github.com/hashicorp/terraform-provider-aws/issues/32174))
* resource/aws_config_configuration_recorder: Add `exclusion_by_resource_types` and `recording_strategy` attributes to the `recording_group` configuration block ([#32007](https://github.com/hashicorp/terraform-provider-aws/issues/32007))
* resource/aws_datasync_task: Add `object_tags` attribute to `options` configuration block ([#27811](https://github.com/hashicorp/terraform-provider-aws/issues/27811))
* resource/aws_networkmanager_attachment_accepter: Added support for Transit Gateway route table attachments ([#32023](https://github.com/hashicorp/terraform-provider-aws/issues/32023))
* resource/aws_ses_active_receipt_rule_set: Support import ([#27604](https://github.com/hashicorp/terraform-provider-aws/issues/27604))

BUG FIXES:

* resource/aws_api_gateway_rest_api: Fix crash when `binary_media_types` is `null` ([#32169](https://github.com/hashicorp/terraform-provider-aws/issues/32169))
* resource/aws_datasync_location_object_storage: Don't ignore `server_certificate` argument ([#27811](https://github.com/hashicorp/terraform-provider-aws/issues/27811))
* resource/aws_eip: Fix `reading EC2 EIP (eipalloc-abcd1234): couldn't find resource` errors when reading new resource ([#32016](https://github.com/hashicorp/terraform-provider-aws/issues/32016))
* resource/aws_quicksight_analysis: Fix schema mapping for string set elements ([#31903](https://github.com/hashicorp/terraform-provider-aws/issues/31903))
* resource/aws_redshiftserverless_workgroup: Fix `waiting for completion: unexpected state 'AVAILABLE'` errors when deleting resource ([#32067](https://github.com/hashicorp/terraform-provider-aws/issues/32067))
* resource/aws_route_table: Fix `reading Route Table (rtb-abcd1234): couldn't find resource` errors when reading new resource ([#30999](https://github.com/hashicorp/terraform-provider-aws/issues/30999))
* resource/aws_storagegateway_smb_file_share: Fix update error when `kms_encrypted` is `true` but `kms_key_arn` is not sent in the request ([#32171](https://github.com/hashicorp/terraform-provider-aws/issues/32171))

## 5.4.0 (June 15, 2023)

FEATURES:

* **New Data Source:** `aws_organizations_policies` ([#31545](https://github.com/hashicorp/terraform-provider-aws/issues/31545))
* **New Data Source:** `aws_organizations_policies_for_target` ([#31682](https://github.com/hashicorp/terraform-provider-aws/issues/31682))
* **New Resource:** `aws_chimesdkvoice_sip_media_application` ([#31937](https://github.com/hashicorp/terraform-provider-aws/issues/31937))
* **New Resource:** `aws_opensearchserverless_collection` ([#31091](https://github.com/hashicorp/terraform-provider-aws/issues/31091))
* **New Resource:** `aws_opensearchserverless_security_config` ([#28776](https://github.com/hashicorp/terraform-provider-aws/issues/28776))
* **New Resource:** `aws_opensearchserverless_vpc_endpoint` ([#28651](https://github.com/hashicorp/terraform-provider-aws/issues/28651))

ENHANCEMENTS:

* resource/aws_elb: Add configurable Create and Update timeouts ([#31976](https://github.com/hashicorp/terraform-provider-aws/issues/31976))
* resource/aws_glue_data_quality_ruleset: Add `catalog_id` argument to `target_table` block ([#31926](https://github.com/hashicorp/terraform-provider-aws/issues/31926))

BUG FIXES:

* provider: Fix `index out of range [0] with length 0` panic ([#32004](https://github.com/hashicorp/terraform-provider-aws/issues/32004))
* resource/aws_elb: Recreate the resource if `subnets` is updated to an empty list ([#31976](https://github.com/hashicorp/terraform-provider-aws/issues/31976))
* resource/aws_lambda_provisioned_concurrency_config: The `function_name` argument now properly handles ARN values ([#31933](https://github.com/hashicorp/terraform-provider-aws/issues/31933))
* resource/aws_quicksight_data_set: Allow physical table map to be optional ([#31863](https://github.com/hashicorp/terraform-provider-aws/issues/31863))
* resource/aws_ssm_default_patch_baseline: Fix `*conns.AWSClient is not ssm.ssmClient: missing method SSMClient` panic ([#31928](https://github.com/hashicorp/terraform-provider-aws/issues/31928))

## 5.3.0 (June 13, 2023)

NOTES:

* resource/aws_instance: The `metadata_options.http_endpoint` argument now correctly defaults to `enabled`. ([#24774](https://github.com/hashicorp/terraform-provider-aws/issues/24774))
* resource/aws_lambda_function: The `replace_security_groups_on_destroy` and `replacement_security_group_ids` attributes are being deprecated as AWS no longer supports this operation. These attributes now have no effect, and will be removed in a future major version. ([#31904](https://github.com/hashicorp/terraform-provider-aws/issues/31904))

FEATURES:

* **New Data Source:** `aws_quicksight_theme` ([#31900](https://github.com/hashicorp/terraform-provider-aws/issues/31900))
* **New Resource:** `aws_opensearchserverless_access_policy` ([#28518](https://github.com/hashicorp/terraform-provider-aws/issues/28518))
* **New Resource:** `aws_opensearchserverless_security_policy` ([#28470](https://github.com/hashicorp/terraform-provider-aws/issues/28470))
* **New Resource:** `aws_quicksight_theme` ([#31900](https://github.com/hashicorp/terraform-provider-aws/issues/31900))

ENHANCEMENTS:

* data-source/aws_redshift_cluster: Add `cluster_namespace_arn` attribute ([#31884](https://github.com/hashicorp/terraform-provider-aws/issues/31884))
* resource/aws_redshift_cluster: Add `cluster_namespace_arn` attribute ([#31884](https://github.com/hashicorp/terraform-provider-aws/issues/31884))
* resource/aws_vpc_endpoint: Add `private_dns_only_for_inbound_resolver_endpoint` attribute to the `dns_options` configuration block ([#31873](https://github.com/hashicorp/terraform-provider-aws/issues/31873))

BUG FIXES:

* resource/aws_ecs_task_definition: Fix to prevent persistent diff when `efs_volume_configuration` has both `root_volume` and `authorization_config` set. ([#26880](https://github.com/hashicorp/terraform-provider-aws/issues/26880))
* resource/aws_instance: Fix default for `metadata_options.http_endpoint` argument. ([#24774](https://github.com/hashicorp/terraform-provider-aws/issues/24774))
* resource/aws_keyspaces_keyspace: Correct plan time validation for `name` ([#31352](https://github.com/hashicorp/terraform-provider-aws/issues/31352))
* resource/aws_keyspaces_table: Correct plan time validation for `keyspace_name`, `table_name` and column names ([#31352](https://github.com/hashicorp/terraform-provider-aws/issues/31352))
* resource/aws_quicksight_analysis: Fix assignment of KPI visual field well target values ([#31901](https://github.com/hashicorp/terraform-provider-aws/issues/31901))
* resource/aws_redshift_cluster: Allow `availability_zone_relocation_enabled` to be `true` when `publicly_accessible` is `true` ([#31886](https://github.com/hashicorp/terraform-provider-aws/issues/31886))
* resource/aws_vpc: Fix `reading EC2 VPC (vpc-abcd1234) Attribute (enableDnsSupport): couldn't find resource` errors when reading new resource ([#31877](https://github.com/hashicorp/terraform-provider-aws/issues/31877))

## 5.2.0 (June  9, 2023)

NOTES:

* resource/aws_mwaa_environment: Upgrading your environment to a new major version of Apache Airflow forces replacement of the resource ([#31833](https://github.com/hashicorp/terraform-provider-aws/issues/31833))

FEATURES:

* **New Data Source:** `aws_budgets_budget` ([#31691](https://github.com/hashicorp/terraform-provider-aws/issues/31691))
* **New Data Source:** `aws_ecr_pull_through_cache_rule` ([#31696](https://github.com/hashicorp/terraform-provider-aws/issues/31696))
* **New Data Source:** `aws_guardduty_finding_ids` ([#31711](https://github.com/hashicorp/terraform-provider-aws/issues/31711))
* **New Data Source:** `aws_iam_principal_policy_simulation` ([#25569](https://github.com/hashicorp/terraform-provider-aws/issues/25569))
* **New Resource:** `aws_chimesdkvoice_global_settings` ([#31365](https://github.com/hashicorp/terraform-provider-aws/issues/31365))
* **New Resource:** `aws_finspace_kx_cluster` ([#31806](https://github.com/hashicorp/terraform-provider-aws/issues/31806))
* **New Resource:** `aws_finspace_kx_database` ([#31803](https://github.com/hashicorp/terraform-provider-aws/issues/31803))
* **New Resource:** `aws_finspace_kx_environment` ([#31802](https://github.com/hashicorp/terraform-provider-aws/issues/31802))
* **New Resource:** `aws_finspace_kx_user` ([#31804](https://github.com/hashicorp/terraform-provider-aws/issues/31804))

ENHANCEMENTS:

* data/aws_ec2_transit_gateway_connect_peer: Add `bgp_peer_address` and `bgp_transit_gateway_addresses` attributes ([#31752](https://github.com/hashicorp/terraform-provider-aws/issues/31752))
* provider: Adds `retry_mode` parameter ([#31745](https://github.com/hashicorp/terraform-provider-aws/issues/31745))
* resource/aws_chime_voice_connector: Add tagging support ([#31746](https://github.com/hashicorp/terraform-provider-aws/issues/31746))
* resource/aws_ec2_transit_gateway_connect_peer: Add `bgp_peer_address` and `bgp_transit_gateway_addresses` attributes ([#31752](https://github.com/hashicorp/terraform-provider-aws/issues/31752))
* resource/aws_ec2_transit_gateway_route_table_association: Add `replace_existing_association` argument ([#31452](https://github.com/hashicorp/terraform-provider-aws/issues/31452))
* resource/aws_fis_experiment_template: Add support for `Volumes` to `actions.*.target` ([#31499](https://github.com/hashicorp/terraform-provider-aws/issues/31499))
* resource/aws_instance: Add `instance_market_options` configuration block and `instance_lifecycle` and `spot_instance_request_id` attributes ([#31495](https://github.com/hashicorp/terraform-provider-aws/issues/31495))
* resource/aws_lambda_function: Add support for `ruby3.2` `runtime` value ([#31842](https://github.com/hashicorp/terraform-provider-aws/issues/31842))
* resource/aws_lambda_layer_version: Add support for `ruby3.2` `compatible_runtimes` value ([#31842](https://github.com/hashicorp/terraform-provider-aws/issues/31842))
* resource/aws_mwaa_environment: Consider `CREATING_SNAPSHOT` a valid pending state for resource update ([#31833](https://github.com/hashicorp/terraform-provider-aws/issues/31833))
* resource/aws_networkfirewall_firewall_policy: Add `stream_exception_policy` option to `firewall_policy.stateful_engine_options` ([#31541](https://github.com/hashicorp/terraform-provider-aws/issues/31541))
* resource/aws_redshiftserverless_workgroup: Additional supported values for `config_parameter.parameter_key` ([#31747](https://github.com/hashicorp/terraform-provider-aws/issues/31747))
* resource/aws_sagemaker_model: Add `container.model_package_name` and `primary_container.model_package_name` arguments ([#31755](https://github.com/hashicorp/terraform-provider-aws/issues/31755))

BUG FIXES:

* data-source/aws_redshift_cluster: Fix crash reading clusters in `modifying` state ([#31772](https://github.com/hashicorp/terraform-provider-aws/issues/31772))
* provider/default_tags: Fix perpetual diff when identical tags are moved from `default_tags` to resource `tags`, and vice versa ([#31826](https://github.com/hashicorp/terraform-provider-aws/issues/31826))
* resource/aws_autoscaling_group: Ignore any `Failed` scaling activities due to IAM eventual consistency ([#31282](https://github.com/hashicorp/terraform-provider-aws/issues/31282))
* resource/aws_dx_connection: Convert `vlan_id` from [`TypeString`](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-types#typestring) to [`TypeInt`](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-types#typeint) in [Terraform state](https://developer.hashicorp.com/terraform/language/state) for existing resources. This fixes a regression introduced in [v5.1.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#510-june--1-2023) causing `a number is required` errors ([#31735](https://github.com/hashicorp/terraform-provider-aws/issues/31735))
* resource/aws_globalaccelerator_endpoint_group: Fix bug updating `endpoint_configuration.weight` to `0` ([#31767](https://github.com/hashicorp/terraform-provider-aws/issues/31767))
* resource/aws_medialive_channel: Fix spelling in `hls_cdn_settings` expander. ([#31844](https://github.com/hashicorp/terraform-provider-aws/issues/31844))
* resource/aws_redshiftserverless_namespace: Fix perpetual `iam_roles` diffs when the namespace contains a workgroup ([#31749](https://github.com/hashicorp/terraform-provider-aws/issues/31749))
* resource/aws_redshiftserverless_workgroup: Change `config_parameter` from `TypeList` to `TypeSet` as order is not significant ([#31747](https://github.com/hashicorp/terraform-provider-aws/issues/31747))
* resource/aws_redshiftserverless_workgroup: Fix `ValidationException: Can't update multiple configurations at the same time` errors ([#31747](https://github.com/hashicorp/terraform-provider-aws/issues/31747))
* resource/aws_vpc_endpoint: Fix tagging error preventing use in ISO partitions ([#31801](https://github.com/hashicorp/terraform-provider-aws/issues/31801))

## 5.1.0 (June  1, 2023)

BREAKING CHANGES:

* resource/aws_iam_role: The `role_last_used` attribute has been removed. Use the `aws_iam_role` data source instead. ([#31656](https://github.com/hashicorp/terraform-provider-aws/issues/31656))

NOTES:

* resource/aws_autoscaling_group: The `load_balancers` and `target_group_arns` attributes have been changed to `Computed`. This means that omitting this argument is interpreted as ignoring any existing load balancer or target group attachments. To remove all load balancer or target group attachments an empty list should be specified. ([#31527](https://github.com/hashicorp/terraform-provider-aws/issues/31527))
* resource/aws_iam_role: The `role_last_used` attribute has been removed. Use the `aws_iam_role` data source instead. See the community feedback provided in the [linked issue](https://github.com/hashicorp/terraform-provider-aws/issues/30861) for additional justification on this change. As the attribute is read-only, unlikely to be used as an input to another resource, and available in the corresponding data source, a breaking change in a minor version was deemed preferable to a long deprecation/removal cycle in this circumstance. ([#31656](https://github.com/hashicorp/terraform-provider-aws/issues/31656))
* resource/aws_redshift_cluster: Ignores the parameter `aqua_configuration_status`, since the AWS API ignores it. Now always returns `auto`. ([#31612](https://github.com/hashicorp/terraform-provider-aws/issues/31612))

FEATURES:

* **New Data Source:** `aws_vpclattice_resource_policy` ([#31372](https://github.com/hashicorp/terraform-provider-aws/issues/31372))
* **New Resource:** `aws_autoscaling_traffic_source_attachment` ([#31527](https://github.com/hashicorp/terraform-provider-aws/issues/31527))
* **New Resource:** `aws_emrcontainers_job_template` ([#31399](https://github.com/hashicorp/terraform-provider-aws/issues/31399))
* **New Resource:** `aws_glue_data_quality_ruleset` ([#31604](https://github.com/hashicorp/terraform-provider-aws/issues/31604))
* **New Resource:** `aws_quicksight_analysis` ([#31542](https://github.com/hashicorp/terraform-provider-aws/issues/31542))
* **New Resource:** `aws_quicksight_dashboard` ([#31448](https://github.com/hashicorp/terraform-provider-aws/issues/31448))
* **New Resource:** `aws_resourcegroups_resource` ([#31430](https://github.com/hashicorp/terraform-provider-aws/issues/31430))

ENHANCEMENTS:

* data-source/aws_autoscaling_group: Add `traffic_source` attribute ([#31527](https://github.com/hashicorp/terraform-provider-aws/issues/31527))
* data-source/aws_opensearch_domain: Add `off_peak_window_options` attribute ([#30965](https://github.com/hashicorp/terraform-provider-aws/issues/30965))
* provider: Increases size of HTTP request bodies in logs to 1 KB ([#31718](https://github.com/hashicorp/terraform-provider-aws/issues/31718))
* resource/aws_appsync_graphql_api: Add `visibility` argument ([#31369](https://github.com/hashicorp/terraform-provider-aws/issues/31369))
* resource/aws_appsync_graphql_api: Add plan time validation for `log_config.cloudwatch_logs_role_arn` ([#31369](https://github.com/hashicorp/terraform-provider-aws/issues/31369))
* resource/aws_autoscaling_group: Add `traffic_source` configuration block ([#31527](https://github.com/hashicorp/terraform-provider-aws/issues/31527))
* resource/aws_cloudformation_stack_set: Add `managed_execution` argument ([#25210](https://github.com/hashicorp/terraform-provider-aws/issues/25210))
* resource/aws_fsx_ontap_volume: Add `skip_final_backup` argument ([#31544](https://github.com/hashicorp/terraform-provider-aws/issues/31544))
* resource/aws_fsx_ontap_volume: Remove default value for `security_style` argument and mark as Computed ([#31544](https://github.com/hashicorp/terraform-provider-aws/issues/31544))
* resource/aws_fsx_ontap_volume: Update `ontap_volume_type` attribute to be configurable ([#31544](https://github.com/hashicorp/terraform-provider-aws/issues/31544))
* resource/aws_fsx_ontap_volume: `junction_path` is Optional ([#31544](https://github.com/hashicorp/terraform-provider-aws/issues/31544))
* resource/aws_fsx_ontap_volume: `storage_efficiency_enabled` is Optional ([#31544](https://github.com/hashicorp/terraform-provider-aws/issues/31544))
* resource/aws_grafana_workspace: Increase default Create and Update timeouts to 30 minutes ([#31422](https://github.com/hashicorp/terraform-provider-aws/issues/31422))
* resource/aws_lambda_invocation: Add lifecycle_scope CRUD to invoke on each resource state transition ([#29367](https://github.com/hashicorp/terraform-provider-aws/issues/29367))
* resource/aws_lambda_layer_version_permission: Add `skip_destroy` attribute ([#29571](https://github.com/hashicorp/terraform-provider-aws/issues/29571))
* resource/aws_lambda_provisioned_concurrency_configuration: Add `skip_destroy` argument ([#31646](https://github.com/hashicorp/terraform-provider-aws/issues/31646))
* resource/aws_opensearch_domain: Add `off_peak_window_options` configuration block ([#30965](https://github.com/hashicorp/terraform-provider-aws/issues/30965))
* resource/aws_sagemaker_endpoint_configuration: Add  and `shadow_production_variants.serverless_config.provisioned_concurrency` arguments ([#31398](https://github.com/hashicorp/terraform-provider-aws/issues/31398))
* resource/aws_transfer_server: Add support for `TransferSecurityPolicy-2023-05` `security_policy_name` value ([#31536](https://github.com/hashicorp/terraform-provider-aws/issues/31536))

BUG FIXES:

* data-source/aws_dx_connection: Fix the `vlan_id` being returned as null ([#31480](https://github.com/hashicorp/terraform-provider-aws/issues/31480))
* provider/tags: Fix crash when some `tags` are `null` and others are `computed` ([#31687](https://github.com/hashicorp/terraform-provider-aws/issues/31687))
* provider: Limits size of HTTP response bodies in logs to 4 KB ([#31718](https://github.com/hashicorp/terraform-provider-aws/issues/31718))
* resource/aws_autoscaling_group: Fix `The AutoRollback parameter cannot be set to true when the DesiredConfiguration parameter is empty` errors when refreshing instances ([#31715](https://github.com/hashicorp/terraform-provider-aws/issues/31715))
* resource/aws_autoscaling_group: Now ignores previous failed scaling activities ([#31551](https://github.com/hashicorp/terraform-provider-aws/issues/31551))
* resource/aws_cloudfront_distribution: Remove the upper limit on `origin_keepalive_timeout` ([#31608](https://github.com/hashicorp/terraform-provider-aws/issues/31608))
* resource/aws_connect_instance: Fix crash when reading instances with `CREATION_FAILED` status ([#31689](https://github.com/hashicorp/terraform-provider-aws/issues/31689))
* resource/aws_connect_security_profile: Set correct `tags` in state ([#31716](https://github.com/hashicorp/terraform-provider-aws/issues/31716))
* resource/aws_dx_connection: Fix the `vlan_id` being returned as null ([#31480](https://github.com/hashicorp/terraform-provider-aws/issues/31480))
* resource/aws_ecs_service: Fix crash when just `alarms` is updated ([#31683](https://github.com/hashicorp/terraform-provider-aws/issues/31683))
* resource/aws_fsx_ontap_volume: Change `storage_virtual_machine_id` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#31544](https://github.com/hashicorp/terraform-provider-aws/issues/31544))
* resource/aws_fsx_ontap_volume: Change `volume_type` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#31544](https://github.com/hashicorp/terraform-provider-aws/issues/31544))
* resource/aws_kendra_index: Persist `user_group_resolution_mode` value to state after creation ([#31669](https://github.com/hashicorp/terraform-provider-aws/issues/31669))
* resource/aws_medialive_channel: Fix attribute spelling in `hls_cdn_settings` expand ([#31647](https://github.com/hashicorp/terraform-provider-aws/issues/31647))
* resource/aws_quicksight_data_set: Fix join_instruction not applied when creating dataset ([#31424](https://github.com/hashicorp/terraform-provider-aws/issues/31424))
* resource/aws_quicksight_data_set: Ignore failure to read refresh properties for non-SPICE datasets ([#31488](https://github.com/hashicorp/terraform-provider-aws/issues/31488))
* resource/aws_rbin_rule: Fix crash when multiple `resource_tags` blocks are configured ([#31393](https://github.com/hashicorp/terraform-provider-aws/issues/31393))
* resource/aws_rds_cluster: Correctly update `db_cluster_instance_class` ([#31709](https://github.com/hashicorp/terraform-provider-aws/issues/31709))
* resource/aws_redshift_cluster: No longer errors on deletion when status is `Maintenance` ([#31612](https://github.com/hashicorp/terraform-provider-aws/issues/31612))
* resource/aws_route53_vpc_association_authorization: Fix `ConcurrentModification` error ([#31588](https://github.com/hashicorp/terraform-provider-aws/issues/31588))
* resource/aws_s3_bucket_replication_configuration: Replication configs sometimes need more than a second or two. This resolves a race condition and adds retry logic when reading them. ([#30995](https://github.com/hashicorp/terraform-provider-aws/issues/30995))

## 5.0.1 (May 26, 2023)

BUG FIXES:

* provider/tags: Fix crash when tags are `null` ([#31587](https://github.com/hashicorp/terraform-provider-aws/issues/31587))

## 5.0.0 (May 25, 2023)

BREAKING CHANGES:

* data-source/aws_api_gateway_rest_api: `minimum_compression_size` is now a string type to allow values set via the `body` attribute to be properly computed. ([#30969](https://github.com/hashicorp/terraform-provider-aws/issues/30969))
* data-source/aws_connect_hours_of_operation: The `hours_of_operation_arn` attribute has been removed ([#31484](https://github.com/hashicorp/terraform-provider-aws/issues/31484))
* data-source/aws_db_instance: With the retirement of EC2-Classic the `db_security_groups` attribute has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* data-source/aws_elasticache_cluster: With the retirement of EC2-Classic the `security_group_names` attribute has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* data-source/aws_elasticache_replication_group: Remove `number_cache_clusters`, `replication_group_description` arguments -- use `num_cache_clusters`, and `description`, respectively, instead ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* data-source/aws_iam_policy_document: Don't add empty `statement.sid` values to `json` attribute value ([#28539](https://github.com/hashicorp/terraform-provider-aws/issues/28539))
* data-source/aws_iam_policy_document: `source_json` and `override_json` have been removed -- use `source_policy_documents` and `override_policy_documents`, respectively, instead ([#30829](https://github.com/hashicorp/terraform-provider-aws/issues/30829))
* data-source/aws_identitystore_group: The `filter` argument has been removed ([#31312](https://github.com/hashicorp/terraform-provider-aws/issues/31312))
* data-source/aws_identitystore_user: The `filter` argument has been removed ([#31312](https://github.com/hashicorp/terraform-provider-aws/issues/31312))
* data-source/aws_launch_configuration: With the retirement of EC2-Classic the `vpc_classic_link_id` and `vpc_classic_link_security_groups` attributes have been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* data-source/aws_redshift_cluster: With the retirement of EC2-Classic the `cluster_security_groups` attribute has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* data-source/aws_secretsmanager_secret: The `rotation_enabled`, `rotation_lambda_arn` and `rotation_rules` attributes have been removed ([#31487](https://github.com/hashicorp/terraform-provider-aws/issues/31487))
* data-source/aws_vpc_peering_connection: With the retirement of EC2-Classic the `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` attributes have been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* provider: The `assume_role.duration_seconds`, `assume_role_with_web_identity.duration_seconds`, `s3_force_path_style`, `shared_credentials_file` and `skip_get_ec2_platforms` attributes have been removed ([#31155](https://github.com/hashicorp/terraform-provider-aws/issues/31155))
* provider: The `aws_subnet_ids` data source has been removed ([#31140](https://github.com/hashicorp/terraform-provider-aws/issues/31140))
* provider: With the retirement of EC2-Classic the `aws_db_security_group` resource has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* provider: With the retirement of EC2-Classic the `aws_elasticache_security_group` resource has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* provider: With the retirement of EC2-Classic the `aws_redshift_security_group` resource has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* provider: With the retirement of Macie Classic the `aws_macie_member_account_association` resource has been removed ([#31058](https://github.com/hashicorp/terraform-provider-aws/issues/31058))
* provider: With the retirement of Macie Classic the `aws_macie_s3_bucket_association` resource has been removed ([#31058](https://github.com/hashicorp/terraform-provider-aws/issues/31058))
* resource/aws_acmpca_certificate_authority: The `status` attribute has been removed ([#31084](https://github.com/hashicorp/terraform-provider-aws/issues/31084))
* resource/aws_api_gateway_rest_api: `minimum_compression_size` is now a string type to allow values set via the `body` attribute to be properly computed. ([#30969](https://github.com/hashicorp/terraform-provider-aws/issues/30969))
* resource/aws_autoscaling_attachment: `alb_target_group_arn` has been removed -- use `lb_target_group_arn` instead ([#30828](https://github.com/hashicorp/terraform-provider-aws/issues/30828))
* resource/aws_autoscaling_group: Remove deprecated `tags` attribute ([#30842](https://github.com/hashicorp/terraform-provider-aws/issues/30842))
* resource/aws_budgets_budget: The `cost_filters` attribute has been removed ([#31395](https://github.com/hashicorp/terraform-provider-aws/issues/31395))
* resource/aws_ce_anomaly_subscription: The `threshold` attribute has been removed ([#30374](https://github.com/hashicorp/terraform-provider-aws/issues/30374))
* resource/aws_cloudwatch_event_target: The `ecs_target.propagate_tags` attribute now has no default value ([#25233](https://github.com/hashicorp/terraform-provider-aws/issues/25233))
* resource/aws_codebuild_project: The `secondary_sources.auth` and `source.auth` attributes have been removed ([#31483](https://github.com/hashicorp/terraform-provider-aws/issues/31483))
* resource/aws_connect_hours_of_operation: The `hours_of_operation_arn` attribute has been removed ([#31484](https://github.com/hashicorp/terraform-provider-aws/issues/31484))
* resource/aws_connect_queue: The `quick_connect_ids_associated` attribute has been removed ([#31376](https://github.com/hashicorp/terraform-provider-aws/issues/31376))
* resource/aws_connect_routing_profile: The `queue_configs_associated` attribute has been removed ([#31376](https://github.com/hashicorp/terraform-provider-aws/issues/31376))
* resource/aws_db_instance: Remove `name` - use `db_name` instead ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_db_instance: With the retirement of EC2-Classic the `security_group_names` attribute has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_db_instance: `id` is no longer the AWS database `identifier` - `id` is now the `dbi-resource-id`. Refer to `identifier` instead of `id` to use the database's identifier ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_default_vpc: With the retirement of EC2-Classic the `enable_classiclink` and `enable_classiclink_dns_support` attributes have been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_dms_endpoint: `s3_settings.ignore_headers_row` has been removed ([#30452](https://github.com/hashicorp/terraform-provider-aws/issues/30452))
* resource/aws_docdb_cluster: `snapshot_identifier` change now properly forces replacement ([#29409](https://github.com/hashicorp/terraform-provider-aws/issues/29409))
* resource/aws_ec2_client_vpn_endpoint: The `status` attribute has been removed ([#31223](https://github.com/hashicorp/terraform-provider-aws/issues/31223))
* resource/aws_ec2_client_vpn_network_association: The `security_groups` attribute has been removed ([#31396](https://github.com/hashicorp/terraform-provider-aws/issues/31396))
* resource/aws_ec2_client_vpn_network_association: The `status` attribute has been removed ([#31223](https://github.com/hashicorp/terraform-provider-aws/issues/31223))
* resource/aws_ecs_cluster: The `capacity_providers` and `default_capacity_provider_strategy` attributes have been removed ([#31346](https://github.com/hashicorp/terraform-provider-aws/issues/31346))
* resource/aws_eip: With the retirement of EC2-Classic the `standard` domain is no longer supported ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_eip_association: With the retirement of EC2-Classic the `standard` domain is no longer supported ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_elasticache_cluster: With the retirement of EC2-Classic the `security_group_names` attribute has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_elasticache_replication_group: Remove `availability_zones`, `number_cache_clusters`, `replication_group_description` arguments -- use `preferred_cache_cluster_azs`, `num_cache_clusters`, and `description`, respectively, instead ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* resource/aws_elasticache_replication_group: Remove `cluster_mode` configuration block -- use top-level `num_node_groups` and `replicas_per_node_group` instead ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* resource/aws_kinesis_firehose_delivery_stream: Remove `s3_configuration` attribute from the root of the resource. `s3_configuration` is now a part of the following blocks: `elasticsearch_configuration`, `opensearch_configuration`, `redshift_configuration`, `splunk_configuration`, and `http_endpoint_configuration` ([#31138](https://github.com/hashicorp/terraform-provider-aws/issues/31138))
* resource/aws_kinesis_firehose_delivery_stream: Remove `s3` as an option for `destination`. Use `extended_s3` instead ([#31138](https://github.com/hashicorp/terraform-provider-aws/issues/31138))
* resource/aws_kinesis_firehose_delivery_stream: Rename `extended_s3_configuration.0.s3_backup_configuration.0.buffer_size` and `extended_s3_configuration.0.s3_backup_configuration.0.buffer_interval` to `extended_s3_configuration.0.s3_backup_configuration.0.buffering_size` and `extended_s3_configuration.0.s3_backup_configuration.0.buffering_interval`, respectively ([#31141](https://github.com/hashicorp/terraform-provider-aws/issues/31141))
* resource/aws_kinesis_firehose_delivery_stream: Rename `redshift_configuration.0.s3_backup_configuration.0.buffer_size` and `redshift_configuration.0.s3_backup_configuration.0.buffer_interval` to `redshift_configuration.0.s3_backup_configuration.0.buffering_size` and `redshift_configuration.0.s3_backup_configuration.0.buffering_interval`, respectively ([#31141](https://github.com/hashicorp/terraform-provider-aws/issues/31141))
* resource/aws_kinesis_firehose_delivery_stream: Rename `s3_configuration.0.buffer_size` and `s3_configuration.0.buffer_internval` to `s3_configuration.0.buffering_size` and `s3_configuration.0.buffering_internval`, respectively ([#31141](https://github.com/hashicorp/terraform-provider-aws/issues/31141))
* resource/aws_launch_configuration: With the retirement of EC2-Classic the `vpc_classic_link_id` and `vpc_classic_link_security_groups` attributes have been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_lightsail_instance: The `ipv6_address` attribute has been removed ([#31489](https://github.com/hashicorp/terraform-provider-aws/issues/31489))
* resource/aws_medialive_multiplex_program: The `statemux_settings` attribute has been removed. Use `statmux_settings` argument instead ([#31034](https://github.com/hashicorp/terraform-provider-aws/issues/31034))
* resource/aws_msk_cluster: The `broker_node_group_info.ebs_volume_size` attribute has been removed ([#31324](https://github.com/hashicorp/terraform-provider-aws/issues/31324))
* resource/aws_neptune_cluster: `snapshot_identifier` change now properly forces replacement ([#29409](https://github.com/hashicorp/terraform-provider-aws/issues/29409))
* resource/aws_networkmanager_core_network: Removed `policy_document` argument -- use `aws_networkmanager_core_network_policy_attachment` resource instead ([#30875](https://github.com/hashicorp/terraform-provider-aws/issues/30875))
* resource/aws_rds_cluster: The `engine` argument is now required and has no default ([#31112](https://github.com/hashicorp/terraform-provider-aws/issues/31112))
* resource/aws_rds_cluster: `snapshot_identifier` change now properly forces replacement ([#29409](https://github.com/hashicorp/terraform-provider-aws/issues/29409))
* resource/aws_rds_cluster_instance: The `engine` argument is now required and has no default ([#31112](https://github.com/hashicorp/terraform-provider-aws/issues/31112))
* resource/aws_redshift_cluster: With the retirement of EC2-Classic the `cluster_security_groups` attribute has been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_route: `instance_id` can no longer be set in configurations. Use `network_interface_id` instead, for example, setting `network_interface_id` to `aws_instance.test.primary_network_interface_id`. ([#30804](https://github.com/hashicorp/terraform-provider-aws/issues/30804))
* resource/aws_route_table: `route.*.instance_id` can no longer be set in configurations. Use `route.*.network_interface_id` instead, for example, setting `network_interface_id` to `aws_instance.test.primary_network_interface_id`. ([#30804](https://github.com/hashicorp/terraform-provider-aws/issues/30804))
* resource/aws_secretsmanager_secret: The `rotation_enabled`, `rotation_lambda_arn` and `rotation_rules` attributes have been removed ([#31487](https://github.com/hashicorp/terraform-provider-aws/issues/31487))
* resource/aws_security_group: With the retirement of EC2-Classic non-VPC security groups are no longer supported ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_security_group_rule: With the retirement of EC2-Classic non-VPC security groups are no longer supported ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_servicecatalog_product: Changes to any `provisioning_artifact_parameters` arguments now properly trigger a replacement. This fixes incorrect behavior, but may technically be breaking for configurations expecting non-functional in-place updates. ([#31061](https://github.com/hashicorp/terraform-provider-aws/issues/31061))
* resource/aws_vpc: With the retirement of EC2-Classic the `enable_classiclink` and `enable_classiclink_dns_support` attributes have been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_vpc_peering_connection: With the retirement of EC2-Classic the `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` attributes have been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_vpc_peering_connection_accepter: With the retirement of EC2-Classic the `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` attributes have been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_vpc_peering_connection_options: With the retirement of EC2-Classic the `allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` attributes have been removed ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))
* resource/aws_wafv2_web_acl: The `statement.managed_rule_group_statement.excluded_rule` and `statement.rule_group_reference_statement.excluded_rule` attributes have been removed ([#31374](https://github.com/hashicorp/terraform-provider-aws/issues/31374))
* resource/aws_wafv2_web_acl_logging_configuration: The `redacted_fields.all_query_arguments`, `redacted_fields.body` and `redacted_fields.single_query_argument` attributes have been removed ([#31486](https://github.com/hashicorp/terraform-provider-aws/issues/31486))

NOTES:

* data-source/aws_elasticache_replication_group: Update configurations to use `description` instead of the `replication_group_description` argument ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* data-source/aws_elasticache_replication_group: Update configurations to use `num_cache_clusters` instead of the `number_cache_clusters` argument ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* data-source/aws_opensearch_domain: The `kibana_endpoint` attribute has been deprecated. All configurations using `kibana_endpoint` should be updated to use the `dashboard_endpoint` attribute instead ([#31490](https://github.com/hashicorp/terraform-provider-aws/issues/31490))
* data-source/aws_quicksight_data_set: The `tags_all` attribute has been deprecated and will be removed in a future version ([#31162](https://github.com/hashicorp/terraform-provider-aws/issues/31162))
* data-source/aws_redshift_service_account: The `aws_redshift_service_account` data source has been deprecated and will be removed in a future version. AWS documentation [states that](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) a [service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) should be used instead of an AWS account ID in any relevant IAM policy ([#31006](https://github.com/hashicorp/terraform-provider-aws/issues/31006))
* data-source/aws_service_discovery_service: The `tags_all` attribute has been deprecated and will be removed in a future version ([#31162](https://github.com/hashicorp/terraform-provider-aws/issues/31162))
* resource/aws_api_gateway_rest_api: Update configurations with `minimum_compression_size` set to pass the value as a string. Valid values remain the same. ([#30969](https://github.com/hashicorp/terraform-provider-aws/issues/30969))
* resource/aws_autoscaling_attachment: Update configurations to use `lb_target_group_arn` instead of `alb_target_group_arn` which has been removed ([#30828](https://github.com/hashicorp/terraform-provider-aws/issues/30828))
* resource/aws_db_event_subscription: Configurations that define `source_ids` using the `id` attribute of `aws_db_instance` must be updated to use `identifier` instead - for example, `source_ids = [aws_db_instance.example.id]` must be updated to `source_ids = [aws_db_instance.example.identifier]` ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_db_instance: Configurations that define `replicate_source_db` using the `id` attribute of `aws_db_instance` must be updated to use `identifier` instead - for example, `replicate_source_db = aws_db_instance.example.id` must be updated to `replicate_source_db = aws_db_instance.example.identifier` ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_db_instance: The change of what `id` is, namely, a DBI Resource ID now versus DB Identifier previously, has far-reaching consequences. Configurations that refer to, for example, `aws_db_instance.example.id` will now have errors and must be changed to use `identifier` instead, for example, `aws_db_instance.example.identifier` ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_db_instance_role_association: Configurations that define `db_instance_identifier` using the `id` attribute of `aws_db_instance` must be updated to use `identifier` instead - for example, `db_instance_identifier = aws_db_instance.example.id` must be updated to `db_instance_identifier = aws_db_instance.example.identifier` ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_db_proxy_target: Configurations that define `db_instance_identifier` using the `id` attribute of `aws_db_instance` must be updated to use `identifier` instead - for example, `db_instance_identifier = aws_db_instance.example.id` must be updated to `db_instance_identifier = aws_db_instance.example.identifier` ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_db_snapshot: Configurations that define `db_instance_identifier` using the `id` attribute of `aws_db_instance` must be updated to use `identifier` instead - for example, `db_instance_identifier = aws_db_instance.example.id` must be updated to `db_instance_identifier = aws_db_instance.example.identifier` ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_docdb_cluster: Changes to the `snapshot_identifier` attribute will now trigger a replacement, rather than an in-place update. This corrects the previous behavior which resulted in a successful apply, but did not actually restore the cluster from the designated snapshot. ([#29409](https://github.com/hashicorp/terraform-provider-aws/issues/29409))
* resource/aws_dx_gateway_association: The `vpn_gateway_id` attribute has been deprecated. All configurations using `vpn_gateway_id` should be updated to use the `associated_gateway_id` attribute instead ([#31384](https://github.com/hashicorp/terraform-provider-aws/issues/31384))
* resource/aws_elasticache_replication_group: Update configurations to use `description` instead of the `replication_group_description` argument ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* resource/aws_elasticache_replication_group: Update configurations to use `num_cache_clusters` instead of the `number_cache_clusters` argument ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* resource/aws_elasticache_replication_group: Update configurations to use `preferred_cache_cluster_azs` instead of the `availability_zones` argument ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* resource/aws_elasticache_replication_group: Update configurations to use top-level `num_node_groups` and `replicas_per_node_group` instead of `cluster_mode.0.num_node_groups` and `cluster_mode.0.replicas_per_node_group`, respectively ([#31008](https://github.com/hashicorp/terraform-provider-aws/issues/31008))
* resource/aws_flow_log: The `log_group_name` attribute has been deprecated. All configurations using `log_group_name` should be updated to use the `log_destination` attribute instead ([#31382](https://github.com/hashicorp/terraform-provider-aws/issues/31382))
* resource/aws_guardduty_organization_configuration: The `auto_enable` argument has been deprecated. Use the `auto_enable_organization_members` argument instead. ([#30736](https://github.com/hashicorp/terraform-provider-aws/issues/30736))
* resource/aws_neptune_cluster: Changes to the `snapshot_identifier` attribute will now trigger a replacement, rather than an in-place update. This corrects the previous behavior which resulted in a successful apply, but did not actually restore the cluster from the designated snapshot. ([#29409](https://github.com/hashicorp/terraform-provider-aws/issues/29409))
* resource/aws_networkmanager_core_network: Update configurations to use the `aws_networkmanager_core_network_policy_attachment` resource instead of the `policy_document` argument ([#30875](https://github.com/hashicorp/terraform-provider-aws/issues/30875))
* resource/aws_opensearch_domain: The `engine_version` attribute no longer has a default value. When omitted, the underlying AWS API will use the latest OpenSearch engine version. ([#31568](https://github.com/hashicorp/terraform-provider-aws/issues/31568))
* resource/aws_opensearch_domain: The `kibana_endpoint` attribute has been deprecated. All configurations using `kibana_endpoint` should be updated to use the `dashboard_endpoint` attribute instead ([#31490](https://github.com/hashicorp/terraform-provider-aws/issues/31490))
* resource/aws_rds_cluster: Changes to the `snapshot_identifier` attribute will now trigger a replacement, rather than an in-place update. This corrects the previous behavior which resulted in a successful apply, but did not actually restore the cluster from the designated snapshot. ([#29409](https://github.com/hashicorp/terraform-provider-aws/issues/29409))
* resource/aws_rds_cluster: Configurations not including the `engine` argument must be updated to include `engine` as it is now required. Previously, not including `engine` was equivalent to `engine = "aurora"` and created a MySQL-5.6-compatible cluster ([#31112](https://github.com/hashicorp/terraform-provider-aws/issues/31112))
* resource/aws_rds_cluster_instance: Configurations not including the `engine` argument must be updated to include `engine` as it is now required. Previously, not including `engine` was equivalent to `engine = "aurora"` and created a MySQL-5.6-compatible cluster instance ([#31112](https://github.com/hashicorp/terraform-provider-aws/issues/31112))
* resource/aws_route: Since `instance_id` can no longer be set in configurations, use `network_interface_id` instead. For example, set `network_interface_id` to `aws_instance.test.primary_network_interface_id`. ([#30804](https://github.com/hashicorp/terraform-provider-aws/issues/30804))
* resource/aws_route_table: Since `route.*.instance_id` can no longer be set in configurations, use `route.*.network_interface_id` instead. For example, set `network_interface_id` to `aws_instance.test.primary_network_interface_id`. ([#30804](https://github.com/hashicorp/terraform-provider-aws/issues/30804))
* resource/aws_ssm_association: The `instance_id` attribute has been deprecated. All configurations using `instance_id` should be updated to use the `targets` attribute instead ([#31380](https://github.com/hashicorp/terraform-provider-aws/issues/31380))

ENHANCEMENTS:

* provider: Allow `computed` `tags` on resources ([#30793](https://github.com/hashicorp/terraform-provider-aws/issues/30793))
* provider: Allow `default_tags` and resource `tags` to include zero values `""` ([#30793](https://github.com/hashicorp/terraform-provider-aws/issues/30793))
* provider: Duplicate `default_tags` can now be included and will be overwritten by resource `tags` ([#30793](https://github.com/hashicorp/terraform-provider-aws/issues/30793))
* resource/aws_db_instance: Updates to `identifier` and `identifier_prefix` will no longer cause the database instance to be destroyed and recreated ([#31232](https://github.com/hashicorp/terraform-provider-aws/issues/31232))
* resource/aws_eip: Deprecate `vpc` attribute. Use `domain` instead ([#31567](https://github.com/hashicorp/terraform-provider-aws/issues/31567))
* resource/aws_guardduty_organization_configuration: Add `auto_enable_organization_members` attribute ([#30736](https://github.com/hashicorp/terraform-provider-aws/issues/30736))
* resource/aws_kinesis_firehose_delivery_stream: Add `s3_configuration` to `elasticsearch_configuration`, `opensearch_configuration`, `redshift_configuration`, `splunk_configuration`, and `http_endpoint_configuration` ([#31138](https://github.com/hashicorp/terraform-provider-aws/issues/31138))
* resource/aws_opensearch_domain: Removed `engine_version` default value ([#31568](https://github.com/hashicorp/terraform-provider-aws/issues/31568))
* resource/aws_wafv2_web_acl: Support `rule_action_override` on `rule_group_reference_statement` ([#31374](https://github.com/hashicorp/terraform-provider-aws/issues/31374))

BUG FIXES:

* resource/aws_ecs_capacity_provider: Allow an `instance_warmup_period` of `0` in the `auto_scaling_group_provider.managed_scaling` configuration block ([#24005](https://github.com/hashicorp/terraform-provider-aws/issues/24005))
* resource/aws_launch_template: Remove default values in `metadata_options` to allow default condition ([#30545](https://github.com/hashicorp/terraform-provider-aws/issues/30545))
* resource/aws_s3_bucket: Fix bucket_regional_domain_name not including region for buckets in us-east-1 ([#25724](https://github.com/hashicorp/terraform-provider-aws/issues/25724))
* resource/aws_s3_object: Remove `acl` default in order to work with S3 buckets that have ACL disabled ([#27197](https://github.com/hashicorp/terraform-provider-aws/issues/27197))
* resource/aws_s3_object_copy: Remove `acl` default in order to work with S3 buckets that have ACL disabled ([#27197](https://github.com/hashicorp/terraform-provider-aws/issues/27197))
* resource/aws_servicecatalog_product: Changes to `provisioning_artifact_parameters` arguments now properly trigger a replacement ([#31061](https://github.com/hashicorp/terraform-provider-aws/issues/31061))
* resource/aws_vpc_peering_connection: Fix crash in `vpcPeeringConnectionOptionsEqual` ([#30966](https://github.com/hashicorp/terraform-provider-aws/issues/30966))

## Previous Releases

For information on prior major releases, see their changelogs:

* [4.x](https://github.com/hashicorp/terraform-provider-aws/blob/release/4.x/CHANGELOG.md)
* [3.x](https://github.com/hashicorp/terraform-provider-aws/blob/release/3.x/CHANGELOG.md)
* [2.x and earlier](https://github.com/hashicorp/terraform-provider-aws/blob/release/2.x/CHANGELOG.md)
