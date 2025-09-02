## 6.12.0 (Unreleased)

NOTES:

* resource/aws_s3_bucket_acl: The `access_control_policy.grant.grantee.display_name` attribute is deprecated. AWS has [ended support for this attribute](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Grantee.html). API responses began inconsistently returning it on July 15, 2025, and will stop returning it entirely on November 21, 2025. This attribute will be removed in a future major version. ([#44090](https://github.com/hashicorp/terraform-provider-aws/issues/44090))
* resource/aws_s3_bucket_acl: The `access_control_policy.owner.display_name` attribute is deprecated. AWS has [ended support for this attribute](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Owner.html). API responses began inconsistently returning it on July 15, 2025, and will stop returning it entirely on November 21, 2025. This attribute will be removed in a future major version. ([#44090](https://github.com/hashicorp/terraform-provider-aws/issues/44090))
* resource/aws_s3_bucket_logging: The `target_grant.grantee.display_name` attribute is deprecated. AWS has [ended support for this attribute](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Grantee.html). API responses began inconsistently returning it on July 15, 2025, and will stop returning it entirely on November 21, 2025. This attribute will be removed in a future major version. ([#44090](https://github.com/hashicorp/terraform-provider-aws/issues/44090))

ENHANCEMENTS:

* data-source/aws_efs_mount_target: Add `ip_address_type` and `ipv6_address` attributes ([#44079](https://github.com/hashicorp/terraform-provider-aws/issues/44079))
* data-source/aws_instance: Add `placement_group_id` attribute ([#38527](https://github.com/hashicorp/terraform-provider-aws/issues/38527))
* data-source/aws_lambda_function: Add `source_kms_key_arn` attribute ([#44080](https://github.com/hashicorp/terraform-provider-aws/issues/44080))
* data-source/aws_launch_template: Add `placement.group_id` attribute ([#44097](https://github.com/hashicorp/terraform-provider-aws/issues/44097))
* resource/aws_efs_mount_target: Add `ip_address_type` and `ipv6_address` arguments to support IPv6 connectivity ([#44079](https://github.com/hashicorp/terraform-provider-aws/issues/44079))
* resource/aws_instance: Add `placement_group_id` argument ([#38527](https://github.com/hashicorp/terraform-provider-aws/issues/38527))
* resource/aws_instance: Add resource identity support ([#44068](https://github.com/hashicorp/terraform-provider-aws/issues/44068))
* resource/aws_lambda_function: Add `source_kms_key_arn` argument ([#44080](https://github.com/hashicorp/terraform-provider-aws/issues/44080))
* resource/aws_launch_template: Add `placement.group_id` argument ([#44097](https://github.com/hashicorp/terraform-provider-aws/issues/44097))
* resource/aws_ssm_association: Add resource identity support ([#44075](https://github.com/hashicorp/terraform-provider-aws/issues/44075))
* resource/aws_ssm_document: Add resource identity support ([#44075](https://github.com/hashicorp/terraform-provider-aws/issues/44075))
* resource/aws_ssm_maintenance_window: Add resource identity support ([#44075](https://github.com/hashicorp/terraform-provider-aws/issues/44075))
* resource/aws_ssm_maintenance_window_target: Add resource identity support ([#44075](https://github.com/hashicorp/terraform-provider-aws/issues/44075))
* resource/aws_ssm_maintenance_window_task: Add resource identity support ([#44075](https://github.com/hashicorp/terraform-provider-aws/issues/44075))
* resource/aws_ssm_patch_baseline: Add resource identity support ([#44075](https://github.com/hashicorp/terraform-provider-aws/issues/44075))

BUG FIXES:

* resource/aws_s3tables_table_policy: Remove plan-time validation of `name` and `namespace` ([#44072](https://github.com/hashicorp/terraform-provider-aws/issues/44072))

## 6.11.0 (August 28, 2025)

FEATURES:

* **New Resource:** `aws_timestreaminfluxdb_db_cluster` ([#42382](https://github.com/hashicorp/terraform-provider-aws/issues/42382))
* **New Resource:** `aws_workspacesweb_browser_settings_association` ([#43735](https://github.com/hashicorp/terraform-provider-aws/issues/43735))
* **New Resource:** `aws_workspacesweb_data_protection_settings_association` ([#43773](https://github.com/hashicorp/terraform-provider-aws/issues/43773))
* **New Resource:** `aws_workspacesweb_identity_provider` ([#43729](https://github.com/hashicorp/terraform-provider-aws/issues/43729))
* **New Resource:** `aws_workspacesweb_ip_access_settings_association` ([#43774](https://github.com/hashicorp/terraform-provider-aws/issues/43774))
* **New Resource:** `aws_workspacesweb_network_settings_association` ([#43775](https://github.com/hashicorp/terraform-provider-aws/issues/43775))
* **New Resource:** `aws_workspacesweb_portal` ([#43444](https://github.com/hashicorp/terraform-provider-aws/issues/43444))
* **New Resource:** `aws_workspacesweb_session_logger` ([#43863](https://github.com/hashicorp/terraform-provider-aws/issues/43863))
* **New Resource:** `aws_workspacesweb_session_logger_association` ([#43866](https://github.com/hashicorp/terraform-provider-aws/issues/43866))
* **New Resource:** `aws_workspacesweb_trust_store` ([#43408](https://github.com/hashicorp/terraform-provider-aws/issues/43408))
* **New Resource:** `aws_workspacesweb_trust_store_association` ([#43778](https://github.com/hashicorp/terraform-provider-aws/issues/43778))
* **New Resource:** `aws_workspacesweb_user_access_logging_settings_association` ([#43776](https://github.com/hashicorp/terraform-provider-aws/issues/43776))
* **New Resource:** `aws_workspacesweb_user_settings_association` ([#43777](https://github.com/hashicorp/terraform-provider-aws/issues/43777))

ENHANCEMENTS:

* data-source/aws_ec2_client_vpn_endpoint: Add `endpoint_ip_address_type` and `traffic_ip_address_type` attributes ([#44059](https://github.com/hashicorp/terraform-provider-aws/issues/44059))
* data-source/aws_network_interface: Add `attachment.network_card_index` attribute ([#42188](https://github.com/hashicorp/terraform-provider-aws/issues/42188))
* data-source/aws_sesv2_email_identity: Add `verification_status` attribute ([#44045](https://github.com/hashicorp/terraform-provider-aws/issues/44045))
* data-source/aws_signer_signing_profile: Add `signing_material` and `signing_parameters` attributes ([#43921](https://github.com/hashicorp/terraform-provider-aws/issues/43921))
* data-source/aws_vpc_ipam: Add `metered_account` attribute ([#43967](https://github.com/hashicorp/terraform-provider-aws/issues/43967))
* resource/aws_datazone_domain: Add `domain_version` and `service_role` arguments to support V2 domains ([#44042](https://github.com/hashicorp/terraform-provider-aws/issues/44042))
* resource/aws_dlm_lifecycle_policy: Add `copy_tags`, `create_interval`, `exclusions`, `extend_deletion`, `policy_language`, `resource_type` and `retain_interval` attributes to `policy_details` configuration block ([#41055](https://github.com/hashicorp/terraform-provider-aws/issues/41055))
* resource/aws_dlm_lifecycle_policy: Add `default_policy` argument ([#41055](https://github.com/hashicorp/terraform-provider-aws/issues/41055))
* resource/aws_dlm_lifecycle_policy: Add `policy_details.create_rule.scripts` argument ([#41055](https://github.com/hashicorp/terraform-provider-aws/issues/41055))
* resource/aws_dlm_lifecycle_policy: Add `policy_details.schedule.cross_region_copy_rule.target_region` argument ([#33796](https://github.com/hashicorp/terraform-provider-aws/issues/33796))
* resource/aws_dlm_lifecycle_policy: Make `policy_details.schedule.cross_region_copy_rule.target` optional ([#33796](https://github.com/hashicorp/terraform-provider-aws/issues/33796))
* resource/aws_dlm_lifecycle_policy:Add `policy_details.schedule.archive_rule` argument ([#41055](https://github.com/hashicorp/terraform-provider-aws/issues/41055))
* resource/aws_dynamodb_contributor_insights: Add `mode` argument in support of [CloudWatch contributor insights modes](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/contributorinsights_HowItWorks.html#contributorinsights_HowItWorks.Modes) ([#43914](https://github.com/hashicorp/terraform-provider-aws/issues/43914))
* resource/aws_ec2_client_vpn_endpoint: Add `endpoint_ip_address_type` and `traffic_ip_address_type` arguments to support IPv6 connectivity in Client VPN ([#44059](https://github.com/hashicorp/terraform-provider-aws/issues/44059))
* resource/aws_ec2_client_vpn_endpoint: Make `client_cidr_block` optional ([#44059](https://github.com/hashicorp/terraform-provider-aws/issues/44059))
* resource/aws_ecr_lifecycle_policy: Add resource identity support ([#44041](https://github.com/hashicorp/terraform-provider-aws/issues/44041))
* resource/aws_ecr_repository: Add resource identity support ([#44041](https://github.com/hashicorp/terraform-provider-aws/issues/44041))
* resource/aws_ecr_repository_policy: Add resource identity support ([#44041](https://github.com/hashicorp/terraform-provider-aws/issues/44041))
* resource/aws_ecs_service: Add `sigint_rollback` argument ([#43986](https://github.com/hashicorp/terraform-provider-aws/issues/43986))
* resource/aws_ecs_service: Change `deployment_configuration` to Optional and Computed ([#43986](https://github.com/hashicorp/terraform-provider-aws/issues/43986))
* resource/aws_eks_cluster: Allow `remote_network_config` to be updated in-place, enabling support for EKS hybrid nodes on existing clusters ([#42928](https://github.com/hashicorp/terraform-provider-aws/issues/42928))
* resource/aws_elasticache_global_replication_group: Change `engine` to Optional and Computed ([#42636](https://github.com/hashicorp/terraform-provider-aws/issues/42636))
* resource/aws_inspector2_filter: Support `code_repository_project_name`, `code_repository_provider_type`, `ecr_image_in_use_count`, and `ecr_image_last_in_use_at` in `filter_criteria` ([#43950](https://github.com/hashicorp/terraform-provider-aws/issues/43950))
* resource/aws_iot_thing_principal_attachment: Add `thing_principal_type` argument ([#43916](https://github.com/hashicorp/terraform-provider-aws/issues/43916))
* resource/aws_kms_alias: Add resource identity support ([#44025](https://github.com/hashicorp/terraform-provider-aws/issues/44025))
* resource/aws_kms_external_key: Add `key_spec` argument ([#44011](https://github.com/hashicorp/terraform-provider-aws/issues/44011))
* resource/aws_kms_external_key: Change `key_usage` to Optional and Computed ([#44011](https://github.com/hashicorp/terraform-provider-aws/issues/44011))
* resource/aws_kms_key: Add resource identity support ([#44025](https://github.com/hashicorp/terraform-provider-aws/issues/44025))
* resource/aws_lb: Add `secondary_ips_auto_assigned_per_subnet` argument for Network Load Balancers ([#43699](https://github.com/hashicorp/terraform-provider-aws/issues/43699))
* resource/aws_mwaa_environment: Add `worker_replacement_strategy` argument ([#43946](https://github.com/hashicorp/terraform-provider-aws/issues/43946))
* resource/aws_network_interface: Add `attachment.network_card_index` argument ([#42188](https://github.com/hashicorp/terraform-provider-aws/issues/42188))
* resource/aws_network_interface_attachment: Add `network_card_index` argument ([#42188](https://github.com/hashicorp/terraform-provider-aws/issues/42188))
* resource/aws_route53_resolver_rule: Add resource identity support ([#44048](https://github.com/hashicorp/terraform-provider-aws/issues/44048))
* resource/aws_route53_resolver_rule_association: Add resource identity support ([#44048](https://github.com/hashicorp/terraform-provider-aws/issues/44048))
* resource/aws_route: Add resource identity support ([#43910](https://github.com/hashicorp/terraform-provider-aws/issues/43910))
* resource/aws_route_table: Add resource identity support ([#43990](https://github.com/hashicorp/terraform-provider-aws/issues/43990))
* resource/aws_s3_bucket_acl: Add resource identity support ([#44043](https://github.com/hashicorp/terraform-provider-aws/issues/44043))
* resource/aws_s3_bucket_cors_configuration: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_logging: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_notification: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_ownership_controls: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_policy: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_public_access_block: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_server_side_encryption_configuration: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_versioning: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_website_configuration: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3tables_table_bucket: Add `force_destroy` argument ([#43922](https://github.com/hashicorp/terraform-provider-aws/issues/43922))
* resource/aws_secretsmanager_secret_version: Add resource identity support ([#44031](https://github.com/hashicorp/terraform-provider-aws/issues/44031))
* resource/aws_sesv2_email_identity: Add `verification_status` attribute ([#44045](https://github.com/hashicorp/terraform-provider-aws/issues/44045))
* resource/aws_signer_signing_profile: Add `signing_parameters` argument ([#43921](https://github.com/hashicorp/terraform-provider-aws/issues/43921))
* resource/aws_synthetics_canary: Add `vpc_config.ipv6_allowed_for_dual_stack` argument ([#43989](https://github.com/hashicorp/terraform-provider-aws/issues/43989))
* resource/aws_vpc_ipam: Add `metered_account` argument ([#43967](https://github.com/hashicorp/terraform-provider-aws/issues/43967))

BUG FIXES:

* data-source/aws_glue_catalog_table: Add `partition_keys.parameters` attribute ([#26702](https://github.com/hashicorp/terraform-provider-aws/issues/26702))
* resource/aws_cognito_user_pool: Fixed to accept an empty `email_mfa_configuration` block ([#43926](https://github.com/hashicorp/terraform-provider-aws/issues/43926))
* resource/aws_db_instance: Fixes the behavior when modifying `database_insights_mode` when using custom KMS key ([#44050](https://github.com/hashicorp/terraform-provider-aws/issues/44050))
* resource/aws_dx_hosted_connection: Fix `DescribeHostedConnections failed for connection dxcon-xxxx doesn't exist` by pointing to the correct connection ID when doing the describe. ([#43499](https://github.com/hashicorp/terraform-provider-aws/issues/43499))
* resource/aws_glue_catalog_table: Add `partition_keys.parameters` argument, fixing `Invalid address to set: []string{"partition_keys", "0", "parameters"}` errors ([#26702](https://github.com/hashicorp/terraform-provider-aws/issues/26702))
* resource/aws_imagebuilder_image_recipe: Increase upper limit of `block_device_mapping.ebs.iops` from `10000` to `100000` ([#43981](https://github.com/hashicorp/terraform-provider-aws/issues/43981))
* resource/aws_nat_gateway: Fix inconsistent final plan for `secondary_private_ip_addresses` ([#43708](https://github.com/hashicorp/terraform-provider-aws/issues/43708))
* resource/aws_spot_instance_request: Change `network_interface.network_card_index` to Computed ([#38336](https://github.com/hashicorp/terraform-provider-aws/issues/38336))
* resource/aws_timestreaminfluxdb_db_instance: Fix tag-only update errors ([#42382](https://github.com/hashicorp/terraform-provider-aws/issues/42382))
* resource/aws_wafv2_web_acl: Add missing flattening of `name` in `response_inspection.header` blocks for `AWSManagedRulesATPRuleSet` and `AWSManagedRulesACFPRuleSet` to avoid persistent plan diffs ([#44032](https://github.com/hashicorp/terraform-provider-aws/issues/44032))

## 6.10.0 (August 21, 2025)

NOTES:

* resource/aws_instance: The `network_interface` block has been deprecated. Use `primary_network_interface` for the primary network interface and `aws_network_interface_attachment` resources for other network interfaces. ([#43953](https://github.com/hashicorp/terraform-provider-aws/issues/43953))
* resource/aws_spot_instance_request: The `network_interface` block has been deprecated. Use `primary_network_interface` for the primary network interface and `aws_network_interface_attachment` resources for other network interfaces. ([#43953](https://github.com/hashicorp/terraform-provider-aws/issues/43953))

ENHANCEMENTS:

* data-source/aws_ecr_repository: Add `image_tag_mutability_exclusion_filter` attribute ([#43886](https://github.com/hashicorp/terraform-provider-aws/issues/43886))
* data-source/aws_ecr_repository_creation_template: Add `image_tag_mutability_exclusion_filter` attribute ([#43886](https://github.com/hashicorp/terraform-provider-aws/issues/43886))
* resource/aws_cloudwatch_event_target: Add resource identity support ([#43984](https://github.com/hashicorp/terraform-provider-aws/issues/43984))
* resource/aws_ecr_repository_creation_template: Add `image_tag_mutability_exclusion_filter` configuration block ([#43886](https://github.com/hashicorp/terraform-provider-aws/issues/43886))
* resource/aws_glue_job: Support `G.12X`, `G.16X`, `R.1X`, `R.2X`, `R.4X`, and `R.8X` as valid values for `worker_type` ([#43988](https://github.com/hashicorp/terraform-provider-aws/issues/43988))
* resource/aws_lambda_permission: Add resource identity support ([#43954](https://github.com/hashicorp/terraform-provider-aws/issues/43954))
* resource/aws_lightsail_static_ip_attachment: Support resource import ([#43874](https://github.com/hashicorp/terraform-provider-aws/issues/43874))
* resource/aws_s3_bucket_cors_configuration: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_logging: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_notification: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_ownership_controls: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_policy: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_public_access_block: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_server_side_encryption_configuration: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_versioning: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_s3_bucket_website_configuration: Add resource identity support ([#43976](https://github.com/hashicorp/terraform-provider-aws/issues/43976))
* resource/aws_secretsmanager_secret: Add resource identity support ([#43872](https://github.com/hashicorp/terraform-provider-aws/issues/43872))
* resource/aws_secretsmanager_secret_policy: Add resource identity support ([#43872](https://github.com/hashicorp/terraform-provider-aws/issues/43872))
* resource/aws_secretsmanager_secret_rotation: Add resource identity support ([#43872](https://github.com/hashicorp/terraform-provider-aws/issues/43872))
* resource/aws_sqs_queue: Add resource identity support ([#43918](https://github.com/hashicorp/terraform-provider-aws/issues/43918))
* resource/aws_sqs_queue_policy: Add resource identity support ([#43918](https://github.com/hashicorp/terraform-provider-aws/issues/43918))
* resource/aws_sqs_queue_redrive_allow_policy: Add resource identity support ([#43918](https://github.com/hashicorp/terraform-provider-aws/issues/43918))
* resource/aws_sqs_queue_redrive_policy: Add resource identity support ([#43918](https://github.com/hashicorp/terraform-provider-aws/issues/43918))

BUG FIXES:

* resource/aws_batch_compute_environment: Allow in-place updates of compute environments that have the `SPOT_PRICE_CAPACITY_OPTIMIZED` strategy ([#40148](https://github.com/hashicorp/terraform-provider-aws/issues/40148))
* resource/aws_imagebuilder_lifecycle_policy: Fix `Provider produced inconsistent result after apply` error when `policy_detail.exclusion_rules.amis.is_public` is omitted ([#43925](https://github.com/hashicorp/terraform-provider-aws/issues/43925))
* resource/aws_instance: Adds `primary_network_interface` to allow importing resources with custom primary network interface. ([#43953](https://github.com/hashicorp/terraform-provider-aws/issues/43953))
* resource/aws_rds_cluster: Fixes the behavior when enabling database_insights_mode="advanced" without changing performance insights retention window ([#43919](https://github.com/hashicorp/terraform-provider-aws/issues/43919))
* resource/aws_rds_cluster: Fixes the behavior when modifying `database_insights_mode` when using custom KMS key ([#43942](https://github.com/hashicorp/terraform-provider-aws/issues/43942))
* resource/aws_spot_instance_request: Adds `primary_network_interface` to allow importing resources with custom primary network interface. ([#43953](https://github.com/hashicorp/terraform-provider-aws/issues/43953))

## 6.9.0 (August 14, 2025)

FEATURES:

* **New Resource:** `aws_appsync_api` ([#43787](https://github.com/hashicorp/terraform-provider-aws/issues/43787))
* **New Resource:** `aws_appsync_channel_namespace` ([#43787](https://github.com/hashicorp/terraform-provider-aws/issues/43787))

ENHANCEMENTS:

* data-source/aws_eks_cluster: Add `deletion_protection` attribute ([#43779](https://github.com/hashicorp/terraform-provider-aws/issues/43779))
* resource/aws_cloudwatch_event_rule: Add resource identity support ([#43758](https://github.com/hashicorp/terraform-provider-aws/issues/43758))
* resource/aws_cloudwatch_metric_alarm: Add resource identity support ([#43759](https://github.com/hashicorp/terraform-provider-aws/issues/43759))
* resource/aws_dynamodb_table: Add `replica.deletion_protection_enabled` argument ([#43240](https://github.com/hashicorp/terraform-provider-aws/issues/43240))
* resource/aws_eks_cluster: Add `deletion_protection` argument ([#43779](https://github.com/hashicorp/terraform-provider-aws/issues/43779))
* resource/aws_lambda_function: Add resource identity support ([#43821](https://github.com/hashicorp/terraform-provider-aws/issues/43821))
* resource/aws_sns_topic_data_protection_policy: Add resource identity support ([#43830](https://github.com/hashicorp/terraform-provider-aws/issues/43830))
* resource/aws_sns_topic_policy: Add resource identity support ([#43830](https://github.com/hashicorp/terraform-provider-aws/issues/43830))
* resource/aws_sns_topic_subscription: Add resource identity support ([#43830](https://github.com/hashicorp/terraform-provider-aws/issues/43830))
* resource/aws_subnet: Add resource identity support ([#43833](https://github.com/hashicorp/terraform-provider-aws/issues/43833))

BUG FIXES:

* data-source/aws_lambda_function: Fix missing value for `reserved_concurrent_executions` attribute when a published version exists. This functionality requires the `lambda:GetFunctionConcurrency` IAM permission ([#43753](https://github.com/hashicorp/terraform-provider-aws/issues/43753))
* data-source/aws_networkfirewall_firewall_policy: Add missing schema definition for `firewall_policy.stateful_engine_options.flow_timeouts` ([#43852](https://github.com/hashicorp/terraform-provider-aws/issues/43852))
* resource/aws_cognito_risk_configuration: Make `account_takeover_risk_configuration.notify_configuration` optional ([#33624](https://github.com/hashicorp/terraform-provider-aws/issues/33624))
* resource/aws_ecs_service: Fix tagging failure after upgrading to v6 provider ([#43816](https://github.com/hashicorp/terraform-provider-aws/issues/43816))
* resource/aws_ecs_service: Fix refreshing `service_connect_configuration` when deleted outside of Terraform ([#43871](https://github.com/hashicorp/terraform-provider-aws/issues/43871))
* resource/aws_lambda_function: Fix missing value for `reserved_concurrent_executions` attribute when a published version exists. This functionality requires the `lambda:GetFunctionConcurrency` IAM permission ([#43753](https://github.com/hashicorp/terraform-provider-aws/issues/43753))
* resource/aws_s3tables_table: Fix `runtime error: invalid memory address or nil pointer dereference` panics when `GetTableMaintenanceConfiguration` returns an error ([#43764](https://github.com/hashicorp/terraform-provider-aws/issues/43764))
* resource/aws_sagemaker_user_profile: Fix incomplete regex for `user_profile_name` ([#43807](https://github.com/hashicorp/terraform-provider-aws/issues/43807))
* resource/aws_servicequotas_service_quota: Add validation, during `create`, to check if new value is less than current value of quota ([#43545](https://github.com/hashicorp/terraform-provider-aws/issues/43545))
* resource/aws_storagegateway_gateway: Handle `InvalidGatewayRequestException: The specified gateway is not connected` errors during Read by using the [`ListGateways` API](https://docs.aws.amazon.com/storagegateway/latest/APIReference/API_ListGateways.html) to return minimal information about a disconnected gateway. This functionality requires the `storagegateway:ListGateways` IAM permission ([#43819](https://github.com/hashicorp/terraform-provider-aws/issues/43819))
* resource/aws_vpc_ipam_pool_cidr: Fix `netmask_length` not being saved and diffed correctly ([#43262](https://github.com/hashicorp/terraform-provider-aws/issues/43262))

## 6.8.0 (August 7, 2025)

FEATURES:

* **New Resource:** `aws_networkfirewall_vpc_endpoint_association` ([#43675](https://github.com/hashicorp/terraform-provider-aws/issues/43675))
* **New Resource:** `aws_quicksight_custom_permissions` ([#43613](https://github.com/hashicorp/terraform-provider-aws/issues/43613))
* **New Resource:** `aws_quicksight_role_custom_permission` ([#43613](https://github.com/hashicorp/terraform-provider-aws/issues/43613))
* **New Resource:** `aws_quicksight_user_custom_permission` ([#43613](https://github.com/hashicorp/terraform-provider-aws/issues/43613))
* **New Resource:** `aws_wafv2_web_acl_rule_group_association` ([#43561](https://github.com/hashicorp/terraform-provider-aws/issues/43561))

ENHANCEMENTS:

* data-source/aws_quicksight_user: Add `custom_permissions_name` attribute ([#43613](https://github.com/hashicorp/terraform-provider-aws/issues/43613))
* data-source/aws_wafv2_web_acl: Add `resource_arn` argument to enable finding web ACLs by resource ARN ([#43597](https://github.com/hashicorp/terraform-provider-aws/issues/43597))
* data-source/aws_wafv2_web_acl: Add support for `CLOUDFRONT` `scope` web ACLs using `resource_arn` ([#43597](https://github.com/hashicorp/terraform-provider-aws/issues/43597))
* resource/aws_bedrock_guardrail: Add `input_action`, `output_action`, `input_enabled`, and `output_enabled` attributes to `sensitive_information_policy_config.pii_entities_config` and `sensitive_information_policy_config.regexes_config` configuration blocks ([#43702](https://github.com/hashicorp/terraform-provider-aws/issues/43702))
* resource/aws_cloudwatch_log_group: Add resource identity support ([#43719](https://github.com/hashicorp/terraform-provider-aws/issues/43719))
* resource/aws_computeoptimizer_recommendation_preferences: Add `AuroraDBClusterStorage` as a valid `resource_type` ([#43677](https://github.com/hashicorp/terraform-provider-aws/issues/43677))
* resource/aws_docdb_cluster: Add `serverless_v2_scaling_configuration` argument in support of [Amazon DocumentDB serverless](https://docs.aws.amazon.com/documentdb/latest/developerguide/docdb-serverless.html) ([#43667](https://github.com/hashicorp/terraform-provider-aws/issues/43667))
* resource/aws_ecr_repository: Add `image_tag_mutability_exclusion_filter` argument ([#43642](https://github.com/hashicorp/terraform-provider-aws/issues/43642))
* resource/aws_ecr_repository: Support `IMMUTABLE_WITH_EXCLUSION` and `MUTABLE_WITH_EXCLUSION` as valid values for `image_tag_mutability` ([#43642](https://github.com/hashicorp/terraform-provider-aws/issues/43642))
* resource/aws_inspector2_enabler: Support resource import ([#43673](https://github.com/hashicorp/terraform-provider-aws/issues/43673))
* resource/aws_instance: Adds `force_destroy` argument that allows destruction even when `disable_api_termination` and `disable_api_stop` are `true` ([#43722](https://github.com/hashicorp/terraform-provider-aws/issues/43722))
* resource/aws_ivs_channel: Add resource identity support ([#43704](https://github.com/hashicorp/terraform-provider-aws/issues/43704))
* resource/aws_ivs_playback_key_pair: Add resource identity support ([#43704](https://github.com/hashicorp/terraform-provider-aws/issues/43704))
* resource/aws_ivs_recording_configuration: Add resource identity support ([#43704](https://github.com/hashicorp/terraform-provider-aws/issues/43704))
* resource/aws_ivschat_logging_configuration: Add resource identity support ([#43697](https://github.com/hashicorp/terraform-provider-aws/issues/43697))
* resource/aws_ivschat_room: Add resource identity support ([#43697](https://github.com/hashicorp/terraform-provider-aws/issues/43697))
* resource/aws_kinesis_firehose_delivery_stream: Add `iceberg_configuration.append_only` argument ([#43647](https://github.com/hashicorp/terraform-provider-aws/issues/43647))
* resource/aws_lightsail_static_ip: Support resource import ([#43672](https://github.com/hashicorp/terraform-provider-aws/issues/43672))
* resource/aws_opensearch_domain_policy: Support resource import ([#43674](https://github.com/hashicorp/terraform-provider-aws/issues/43674))
* resource/aws_quicksight_user: Add plan-time validation of `iam_arn` ([#43613](https://github.com/hashicorp/terraform-provider-aws/issues/43613))
* resource/aws_quicksight_user: Change `user_name` to Optional and Computed ([#43613](https://github.com/hashicorp/terraform-provider-aws/issues/43613))
* resource/aws_quicksight_user: Support `IAM_IDENTITY_CENTER` as a valid value for `identity_type` ([#43613](https://github.com/hashicorp/terraform-provider-aws/issues/43613))
* resource/aws_quicksight_user: Support `RESTRICTED_AUTHOR` and `RESTRICTED_READER` as valid values for `user_role` ([#43613](https://github.com/hashicorp/terraform-provider-aws/issues/43613))
* resource/aws_security_group: Add parameterized resource identity support ([#43744](https://github.com/hashicorp/terraform-provider-aws/issues/43744))
* resource/aws_sqs_queue: Increase upper limit of `max_message_size` from 256 KiB to 1024 KiB ([#43710](https://github.com/hashicorp/terraform-provider-aws/issues/43710))
* resource/aws_ssm_parameter: Add resource identity support ([#43736](https://github.com/hashicorp/terraform-provider-aws/issues/43736))

BUG FIXES:

* ephemeral-resource/aws_lambda_invocation: Fix plan inconsistency issue due to improperly assigned payload values ([#43676](https://github.com/hashicorp/terraform-provider-aws/issues/43676))
* provider: Fix failure to detect resources deleted outside of Terraform as missing for numerous resource types ([#43659](https://github.com/hashicorp/terraform-provider-aws/issues/43659))
* resource/aws_batch_compute_environment: Fix `inconsistent final plan` error when `compute_resource.launch_template.version` is unknown during an update ([#43337](https://github.com/hashicorp/terraform-provider-aws/issues/43337))
* resource/aws_bedrockagent_flow: Prevent `created_at` becoming `null` on Update ([#43654](https://github.com/hashicorp/terraform-provider-aws/issues/43654))
* resource/aws_ec2_managed_prefix_list: Fix `PrefixListVersionMismatch: The prefix list has the incorrect version number` errors when updating entry description ([#43661](https://github.com/hashicorp/terraform-provider-aws/issues/43661))
* resource/aws_fsx_lustre_file_system: Fix validation of SSD read cache size for file systems using the Intelligent-Tiering storage class ([#43605](https://github.com/hashicorp/terraform-provider-aws/issues/43605))
* resource/aws_instance: Prevent destruction of resource when `disable_api_termination` is `true` ([#43722](https://github.com/hashicorp/terraform-provider-aws/issues/43722))
* resource/aws_kms_key: Restore pre-v6.3.0 retry delay behavior when waiting for continuous target state occurrences. This fixes certain tag update timeouts ([#43716](https://github.com/hashicorp/terraform-provider-aws/issues/43716))
* resource/aws_s3tables_table_bucket: Fix crash on `maintenance_configuration` read failure ([#43707](https://github.com/hashicorp/terraform-provider-aws/issues/43707))
* resource/aws_sagemaker_image: Fix `image_name` regular expression validation ([#43751](https://github.com/hashicorp/terraform-provider-aws/issues/43751))
* resource/aws_timestreaminfluxdb_db_instance: Don't mark `network_type` as [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) if the value is not configured. This fixes a problem with `terraform apply -refresh=false` after upgrade from `v5.90.0` and below ([#43534](https://github.com/hashicorp/terraform-provider-aws/issues/43534))
* resource/aws_wafv2_regex_pattern_set: Remove maximum items limit on the `regular_expression` argument ([#43693](https://github.com/hashicorp/terraform-provider-aws/issues/43693))

## 6.7.0 (July 31, 2025)

FEATURES:

* **New Resource:** `aws_quicksight_ip_restriction` ([#43596](https://github.com/hashicorp/terraform-provider-aws/issues/43596))
* **New Resource:** `aws_quicksight_key_registration` ([#43587](https://github.com/hashicorp/terraform-provider-aws/issues/43587))

ENHANCEMENTS:

* data-source/aws_codebuild_fleet: Add `instance_type` attribute in `compute_configuration` block ([#43449](https://github.com/hashicorp/terraform-provider-aws/issues/43449))
* data-source/aws_ebs_volume: Add `volume_initialization_rate` attribute ([#43565](https://github.com/hashicorp/terraform-provider-aws/issues/43565))
* data-source/aws_ecs_service: Support `load_balancer` attribute ([#43582](https://github.com/hashicorp/terraform-provider-aws/issues/43582))
* data-source/aws_s3_access_point: Add `tags` attribute. This functionality requires the `s3:ListTagsForResource` IAM permission with S3 Access Points for general purpose buckets and the `s3express:ListTagsForResource` IAM permission with S3 Access Points for directory buckets ([#43630](https://github.com/hashicorp/terraform-provider-aws/issues/43630))
* data-source/aws_verifiedpermissions_policy_store: Add `deletion_protection` attribute ([#43452](https://github.com/hashicorp/terraform-provider-aws/issues/43452))
* resource/aws_athena_workgroup: Add `configuration.identity_center_configuration` argument ([#38717](https://github.com/hashicorp/terraform-provider-aws/issues/38717))
* resource/aws_cleanrooms_collaboration: Add `analytics_engine` argument ([#43614](https://github.com/hashicorp/terraform-provider-aws/issues/43614))
* resource/aws_codebuild_fleet: Add `instance_type` argument in `compute_configuration` block to support custom instance types ([#43449](https://github.com/hashicorp/terraform-provider-aws/issues/43449))
* resource/aws_ebs_volume: Add `volume_initialization_rate` argument ([#43565](https://github.com/hashicorp/terraform-provider-aws/issues/43565))
* resource/aws_s3_access_point: Add `tags` argument and `tags_all` attribute. This functionality requires the `s3:ListTagsForResource`, `s3:TagResource`, and `s3:UntagResource` IAM permissions with S3 Access Points for general purpose buckets and the `s3express:ListTagsForResource`, `s3express:TagResource`, and `s3express:UntagResource` IAM permissions with S3 Access Points for directory buckets ([#43630](https://github.com/hashicorp/terraform-provider-aws/issues/43630))
* resource/aws_verifiedpermissions_policy_store: Add `deletion_protection` argument ([#43452](https://github.com/hashicorp/terraform-provider-aws/issues/43452))

BUG FIXES:

* resource/aws_bedrockagent_flow: Fix `missing required field, CreateFlowInput.Definition.Nodes[0].Configuration[prompt].SourceConfiguration[resource].PromptArn` errors on Create ([#43595](https://github.com/hashicorp/terraform-provider-aws/issues/43595))
* resource/aws_s3_bucket: Accept `NoSuchTagSetError` responses from S3-compatible services ([#43589](https://github.com/hashicorp/terraform-provider-aws/issues/43589))
* resource/aws_s3_object: Accept `NoSuchTagSetError` responses from S3-compatible services ([#43589](https://github.com/hashicorp/terraform-provider-aws/issues/43589))
* resource/aws_servicequotas_service_quota: Fix error when updating a pending service quota request ([#43606](https://github.com/hashicorp/terraform-provider-aws/issues/43606))
* resource/aws_ssm_parameter: Fix `Provider produced inconsistent final plan` errors when changing from using `value` to using `value_wo` ([#42877](https://github.com/hashicorp/terraform-provider-aws/issues/42877))
* resource/aws_ssm_parameter: Fix `version` not being updated when `description` changes ([#42595](https://github.com/hashicorp/terraform-provider-aws/issues/42595))

## 6.6.0 (July 28, 2025)

FEATURES:

* **New Resource:** `aws_connect_phone_number_contact_flow_association` ([#43557](https://github.com/hashicorp/terraform-provider-aws/issues/43557))
* **New Resource:** `aws_nat_gateway_eip_association` ([#42591](https://github.com/hashicorp/terraform-provider-aws/issues/42591))

ENHANCEMENTS:

* data-source/aws_cloudwatch_event_bus: Add `log_config` attribute ([#43453](https://github.com/hashicorp/terraform-provider-aws/issues/43453))
* data-source/aws_ssm_patch_baseline: Add `available_security_updates_compliance_status` argument ([#43560](https://github.com/hashicorp/terraform-provider-aws/issues/43560))
* feature/aws_bedrock_guardrail: Add `cross_region_config`, `content_policy_config.tier_config`, and `topic_policy_config.tier_config` arguments ([#43517](https://github.com/hashicorp/terraform-provider-aws/issues/43517))
* resource/aws_athena_database: Add `workgroup` argument ([#36628](https://github.com/hashicorp/terraform-provider-aws/issues/36628))
* resource/aws_batch_compute_environment: Add `compute_resources.ec2_configuration.image_kubernetes_version` argument ([#43454](https://github.com/hashicorp/terraform-provider-aws/issues/43454))
* resource/aws_cloudwatch_event_bus: Add `log_config` argument ([#43453](https://github.com/hashicorp/terraform-provider-aws/issues/43453))
* resource/aws_cognito_resource_server: Allow `name` to be updated in-place ([#41702](https://github.com/hashicorp/terraform-provider-aws/issues/41702))
* resource/aws_cognito_user_pool: Allow `name` to be updated in-place ([#42639](https://github.com/hashicorp/terraform-provider-aws/issues/42639))
* resource/aws_globalaccelerator_custom_routing_endpoint_group: Add resource identity support ([#43539](https://github.com/hashicorp/terraform-provider-aws/issues/43539))
* resource/aws_globalaccelerator_custom_routing_listener: Add resource identity support ([#43539](https://github.com/hashicorp/terraform-provider-aws/issues/43539))
* resource/aws_globalaccelerator_endpoint_group: Add resource identity support ([#43539](https://github.com/hashicorp/terraform-provider-aws/issues/43539))
* resource/aws_globalaccelerator_listener: Add resource identity support ([#43539](https://github.com/hashicorp/terraform-provider-aws/issues/43539))
* resource/aws_imagebuilder_container_recipe: Add resource identity support ([#43540](https://github.com/hashicorp/terraform-provider-aws/issues/43540))
* resource/aws_imagebuilder_distribution_configuration: Add resource identity support ([#43540](https://github.com/hashicorp/terraform-provider-aws/issues/43540))
* resource/aws_imagebuilder_image: Add resource identity support ([#43540](https://github.com/hashicorp/terraform-provider-aws/issues/43540))
* resource/aws_imagebuilder_image_pipeline: Add resource identity support ([#43540](https://github.com/hashicorp/terraform-provider-aws/issues/43540))
* resource/aws_imagebuilder_image_recipe: Add resource identity support ([#43540](https://github.com/hashicorp/terraform-provider-aws/issues/43540))
* resource/aws_imagebuilder_infrastructure_configuration: Add resource identity support ([#43540](https://github.com/hashicorp/terraform-provider-aws/issues/43540))
* resource/aws_imagebuilder_workflow: Add resource identity support ([#43540](https://github.com/hashicorp/terraform-provider-aws/issues/43540))
* resource/aws_inspector_assessment_target: Add resource identity support ([#43542](https://github.com/hashicorp/terraform-provider-aws/issues/43542))
* resource/aws_inspector_assessment_template: Add resource identity support ([#43542](https://github.com/hashicorp/terraform-provider-aws/issues/43542))
* resource/aws_inspector_resource_group: Add resource identity support ([#43542](https://github.com/hashicorp/terraform-provider-aws/issues/43542))
* resource/aws_nat_gateway: Change `secondary_allocation_ids` to Optional and Computed ([#42591](https://github.com/hashicorp/terraform-provider-aws/issues/42591))
* resource/aws_ssm_patch_baseline: Add `available_security_updates_compliance_status` argument ([#43560](https://github.com/hashicorp/terraform-provider-aws/issues/43560))
* resource/aws_ssm_service_setting: Support short format (with `/ssm/` prefix) for `setting_id` ([#43562](https://github.com/hashicorp/terraform-provider-aws/issues/43562))

BUG FIXES:

* resource/aws_appsync_api_cache: Fix "missing required field" error during update ([#43523](https://github.com/hashicorp/terraform-provider-aws/issues/43523))
* resource/aws_cloudwatch_log_delivery_destination: Fix update failure when tags are set ([#43576](https://github.com/hashicorp/terraform-provider-aws/issues/43576))
* resource/aws_ecs_service: Fix unspecified `test_listener_rule` incorrectly being set as empty string in `load_balancer.advanced_configuration` block ([#43558](https://github.com/hashicorp/terraform-provider-aws/issues/43558))

## 6.5.0 (July 24, 2025)

NOTES:

* resource/aws_cognito_log_delivery_configuration: Because we cannot easily test all this functionality, it is best effort and we ask for community help in testing ([#43396](https://github.com/hashicorp/terraform-provider-aws/issues/43396))
* resource/aws_ecs_service: Acceptance tests cannot fully reproduce scenarios with deployments older than 3 months. Community feedback on this fix is appreciated, particularly for long-running ECS services with in-place updates ([#43502](https://github.com/hashicorp/terraform-provider-aws/issues/43502))

FEATURES:

* **New Data Source:** `aws_ecr_images` ([#42577](https://github.com/hashicorp/terraform-provider-aws/issues/42577))
* **New Resource:** `aws_cognito_log_delivery_configuration` ([#43396](https://github.com/hashicorp/terraform-provider-aws/issues/43396))
* **New Resource:** `aws_networkfirewall_firewall_transit_gateway_attachment_accepter` ([#43430](https://github.com/hashicorp/terraform-provider-aws/issues/43430))
* **New Resource:** `aws_s3_bucket_metadata_configuration` ([#41364](https://github.com/hashicorp/terraform-provider-aws/issues/41364))

ENHANCEMENTS:

* data-source/aws_dms_endpoint: Add `postgres_settings.authentication_method` and `postgres_settings.service_access_role_arn` attributes ([#43440](https://github.com/hashicorp/terraform-provider-aws/issues/43440))
* data-source/aws_networkfirewall_firewall: Add `availability_zone_change_protection`, `availability_zone_mapping`, `firewall_status.sync_states.attachment.status_message`, `firewall_status.transit_gateway_attachment_sync_states`, `transit_gateway_id`, and `transit_gateway_owner_account_id` attributes ([#43430](https://github.com/hashicorp/terraform-provider-aws/issues/43430))
* resource/aws_alb_listener: Add resource identity support ([#43161](https://github.com/hashicorp/terraform-provider-aws/issues/43161))
* resource/aws_alb_listener_rule: Add resource identity support ([#43155](https://github.com/hashicorp/terraform-provider-aws/issues/43155))
* resource/aws_alb_target_group: Add resource identity support ([#43171](https://github.com/hashicorp/terraform-provider-aws/issues/43171))
* resource/aws_dms_endpoint: Add `oracle_settings` configuration block for authentication method ([#43125](https://github.com/hashicorp/terraform-provider-aws/issues/43125))
* resource/aws_dms_endpoint: Add `postgres_settings.authentication_method` and `postgres_settings.service_access_role_arn` arguments ([#43440](https://github.com/hashicorp/terraform-provider-aws/issues/43440))
* resource/aws_dms_endpoint: Add plan-time validation of `postgres_settings.database_mode`, `postgres_settings.map_long_varchar_as`, and `postgres_settings.plugin_name` arguments ([#43440](https://github.com/hashicorp/terraform-provider-aws/issues/43440))
* resource/aws_dms_replication_instance: Add `dns_name_servers` attribute and `kerberos_authentication_settings` configuration block for Kerberos authentication settings ([#43125](https://github.com/hashicorp/terraform-provider-aws/issues/43125))
* resource/aws_dx_gateway_association: Add `transit_gateway_attachment_id` attribute. This functionality requires the `ec2:DescribeTransitGatewayAttachments` IAM permission ([#43436](https://github.com/hashicorp/terraform-provider-aws/issues/43436))
* resource/aws_globalaccelerator_accelerator: Add resource identity support ([#43200](https://github.com/hashicorp/terraform-provider-aws/issues/43200))
* resource/aws_globalaccelerator_custom_routing_accelerator: Add resource identity support ([#43423](https://github.com/hashicorp/terraform-provider-aws/issues/43423))
* resource/aws_glue_registry: Add resource identity support ([#43450](https://github.com/hashicorp/terraform-provider-aws/issues/43450))
* resource/aws_glue_schema: Add resource identity support ([#43450](https://github.com/hashicorp/terraform-provider-aws/issues/43450))
* resource/aws_iam_openid_connect_provider: Add resource identity support ([#43503](https://github.com/hashicorp/terraform-provider-aws/issues/43503))
* resource/aws_iam_policy: Add resource identity support ([#43503](https://github.com/hashicorp/terraform-provider-aws/issues/43503))
* resource/aws_iam_saml_provider: Add resource identity support ([#43503](https://github.com/hashicorp/terraform-provider-aws/issues/43503))
* resource/aws_iam_service_linked_role: Add resource identity support ([#43503](https://github.com/hashicorp/terraform-provider-aws/issues/43503))
* resource/aws_inspector2_enabler: Support `CODE_REPOSITORY` as a valid value for `resource_types` ([#43525](https://github.com/hashicorp/terraform-provider-aws/issues/43525))
* resource/aws_inspector2_organization_configuration: Add `auto_enable.code_repository` argument ([#43525](https://github.com/hashicorp/terraform-provider-aws/issues/43525))
* resource/aws_lb_listener: Add resource identity support ([#43161](https://github.com/hashicorp/terraform-provider-aws/issues/43161))
* resource/aws_lb_listener_rule: Add resource identity support ([#43155](https://github.com/hashicorp/terraform-provider-aws/issues/43155))
* resource/aws_lb_target_group: Add resource identity support ([#43171](https://github.com/hashicorp/terraform-provider-aws/issues/43171))
* resource/aws_lb_trust_store: Add resource identity support ([#43186](https://github.com/hashicorp/terraform-provider-aws/issues/43186))
* resource/aws_networkfirewall_firewall: Add `availability_zone_change_protection`, `availability_zone_mapping`, and `transit_gateway_id` arguments and `firewall_status.transit_gateway_attachment_sync_states` and `transit_gateway_owner_account_id` attributes ([#43430](https://github.com/hashicorp/terraform-provider-aws/issues/43430))
* resource/aws_networkfirewall_firewall: Mark `subnet_mapping` and `vpc_id` as Optional ([#43430](https://github.com/hashicorp/terraform-provider-aws/issues/43430))
* resource/aws_quicksight_account_subscription: Add import support. This resource can now be imported via the `aws_account_id` argument. ([#43501](https://github.com/hashicorp/terraform-provider-aws/issues/43501))
* resource/aws_sns_topic: Add resource identity support ([#43202](https://github.com/hashicorp/terraform-provider-aws/issues/43202))
* resource/aws_wafv2_rule_group: Add `rules_json` argument ([#43397](https://github.com/hashicorp/terraform-provider-aws/issues/43397))
* resource/aws_wafv2_web_acl: Add `statement.rate_based_statement.custom_key.asn` argument ([#43506](https://github.com/hashicorp/terraform-provider-aws/issues/43506))

BUG FIXES:

* provider: Prevent planned `forces replacement` on `region` for numerous resource types when upgrading from a pre-v6.0.0 provider version and `-refresh=false` is in effect ([#43516](https://github.com/hashicorp/terraform-provider-aws/issues/43516))
* resource/aws_api_gateway_resource: Recompute `path` when `path_part` is updated ([#43215](https://github.com/hashicorp/terraform-provider-aws/issues/43215))
* resource/aws_bedrockagent_flow: Remove `definition.connection` and `definition.node` list length limits ([#43471](https://github.com/hashicorp/terraform-provider-aws/issues/43471))
* resource/aws_ecs_service: Improve stabilization logic to handle both new deployments and in-place updates correctly. This fixes a regression introduced in [v6.4.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#640-july-17-2025) ([#43502](https://github.com/hashicorp/terraform-provider-aws/issues/43502))
* resource/aws_instance: Recompute `ipv6_addresses` when `ipv6_address_count` is updated ([#43158](https://github.com/hashicorp/terraform-provider-aws/issues/43158))

## 6.4.0 (July 17, 2025)

FEATURES:

* **New Data Source:** `aws_s3_access_point` ([#43391](https://github.com/hashicorp/terraform-provider-aws/issues/43391))
* **New Resource:** `aws_bedrockagent_flow` ([#42201](https://github.com/hashicorp/terraform-provider-aws/issues/42201))
* **New Resource:** `aws_fsx_s3_access_point_attachment` ([#43391](https://github.com/hashicorp/terraform-provider-aws/issues/43391))

ENHANCEMENTS:

* data-source/aws_bedrock_inference_profiles: Add `type` argument ([#43150](https://github.com/hashicorp/terraform-provider-aws/issues/43150))
* data-source/aws_lakeformation_resource: Support `hybrid_access_enabled`, `with_federation` and `with_privileged_access` attributes ([#43377](https://github.com/hashicorp/terraform-provider-aws/issues/43377))
* resource/aws_acm_certificate: Support `options.export` argument to issue an exportable certificate ([#43207](https://github.com/hashicorp/terraform-provider-aws/issues/43207))
* resource/aws_cloudwatch_log_metric_filter: Add `apply_on_transformed_logs` argument ([#43381](https://github.com/hashicorp/terraform-provider-aws/issues/43381))
* resource/aws_datasync_location_object_storage: Make `agent_arns` optional ([#43400](https://github.com/hashicorp/terraform-provider-aws/issues/43400))
* resource/aws_ecs_service: Add `deployment_configuration` argument ([#43434](https://github.com/hashicorp/terraform-provider-aws/issues/43434))
* resource/aws_ecs_service: Add `load_balancer.advanced_configuration` argument ([#43434](https://github.com/hashicorp/terraform-provider-aws/issues/43434))
* resource/aws_ecs_service: Add `service.client_alias.test_traffic_rules` argument ([#43434](https://github.com/hashicorp/terraform-provider-aws/issues/43434))
* resource/aws_ecs_service: `deployment_controller.type` changes no longer force a replacement ([#43434](https://github.com/hashicorp/terraform-provider-aws/issues/43434))
* resource/aws_lakeformation_resource: Support `with_privileged_access` argument ([#43377](https://github.com/hashicorp/terraform-provider-aws/issues/43377))
* resource/aws_s3_bucket_public_access_block: Add `skip_destroy` argument ([#43415](https://github.com/hashicorp/terraform-provider-aws/issues/43415))

BUG FIXES:

* resource/aws_bedrockagent_agent_action_group: Correctly set `parent_action_group_signature` on Read ([#43355](https://github.com/hashicorp/terraform-provider-aws/issues/43355))
* resource/aws_datazone_environment_blueprint_configuration: Fix `Inappropriate value for attribute "regional_parameters"` errors during planning. This fixes a regression introduced in [v6.0.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#600-june-18-2025) ([#43382](https://github.com/hashicorp/terraform-provider-aws/issues/43382))
* resource/aws_ec2_transit_gateway_route_table_propagation: Don't mark `transit_gateway_attachment_id` as [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) if the value is known not to change ([#43405](https://github.com/hashicorp/terraform-provider-aws/issues/43405))
* resource/aws_lambda_function: Fix `waiting for Lambda Function (...) version publish: unexpected state '', wanted target 'Successful'` errors on Update. This fixes a regression introduced in [v6.2.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#620-july--2-2025) ([#43416](https://github.com/hashicorp/terraform-provider-aws/issues/43416))
* resource/aws_lexv2models_slot: Fix error when `sub_slot_setting.slot_specification.value_elicitation_setting.prompt_specification.prompt_attempts_specification` and `value_elicitation_setting.prompt_specification.prompt_attempts_specification` have default values ([#43358](https://github.com/hashicorp/terraform-provider-aws/issues/43358))
* resource/aws_securitylake_data_lake: Allow `meta_store_role_arn` to be updated in-place ([#36874](https://github.com/hashicorp/terraform-provider-aws/issues/36874))

## 6.3.0 (July 10, 2025)

FEATURES:

* **New Resource:** `aws_prometheus_query_logging_configuration` ([#43222](https://github.com/hashicorp/terraform-provider-aws/issues/43222))

ENHANCEMENTS:

* data-source/aws_cloudfront_distribution: Add `anycast_ip_list_id` attribute ([#43196](https://github.com/hashicorp/terraform-provider-aws/issues/43196))
* data-source/aws_networkmanager_core_network_policy_document: Add `core_network_configuration.dns_support` and `core_network_configuration.security_group_referencing_support` arguments ([#43277](https://github.com/hashicorp/terraform-provider-aws/issues/43277))
* resource/aws_cloudfront_distribution: Add `anycast_ip_list_id` argument ([#43196](https://github.com/hashicorp/terraform-provider-aws/issues/43196))
* resource/aws_dynamodb_table: Add `replica.consistency_mode` argument in support of [multi-Region strong consistency](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/V2globaltables_HowItWorks.html#V2globaltables_HowItWorks.choosing-consistency-mode) for Amazon DynamoDB global tables ([#43236](https://github.com/hashicorp/terraform-provider-aws/issues/43236))

BUG FIXES:

* provider: Fix `runtime error: invalid memory address or nil pointer dereference` panics for numerous resource types when modifying `tags` ([#43324](https://github.com/hashicorp/terraform-provider-aws/issues/43324))
* resource/aws_bedrockagent_agent_action_group: Add missing prepare agent call when deleting an action group ([#43232](https://github.com/hashicorp/terraform-provider-aws/issues/43232))
* resource/aws_bedrockagent_agent_action_group: Retry `operation can't be performed on Agent when it is in Preparing state.` errors during agent action group base creation, update, and deletion. ([#43232](https://github.com/hashicorp/terraform-provider-aws/issues/43232))
* resource/aws_bedrockagent_agent_knowledge_base_association: Add missing prepare agent call when deleting a knowledge base association ([#43232](https://github.com/hashicorp/terraform-provider-aws/issues/43232))
* resource/aws_bedrockagent_agent_knowledge_base_association: Retry `operation can't be performed on Agent when it is in Preparing state.` errors during agent knowledge base creation and disassociation ([#43232](https://github.com/hashicorp/terraform-provider-aws/issues/43232))
* resource/aws_cloudfrontkeyvaluestore_keys_exclusive: Fix errant deletion of key value pairs when a value is changed ([#43208](https://github.com/hashicorp/terraform-provider-aws/issues/43208))
* resource/aws_cognito_user_pool_domain: Correctly update `managed_login_version` for custom Cognito domains ([#43252](https://github.com/hashicorp/terraform-provider-aws/issues/43252))
* resource/aws_db_instance_role_association: Retry `InvalidDBInstanceState` errors on delete ([#43303](https://github.com/hashicorp/terraform-provider-aws/issues/43303))
* resource/aws_medialive_channel: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panics when configuration blocks are empty ([#43308](https://github.com/hashicorp/terraform-provider-aws/issues/43308))
* resource/aws_rds_cluster_role_association: Retry `InvalidDBClusterStateFault` errors on delete ([#43303](https://github.com/hashicorp/terraform-provider-aws/issues/43303))
* resource/aws_redshift_cluster: Correctly set `availability_zone_relocation_enabled` ([#43270](https://github.com/hashicorp/terraform-provider-aws/issues/43270))
* resource/aws_route53profiles_resource_association: Change `resource_properties` to Computed to enable `vpc_endpoint` associations ([#42562](https://github.com/hashicorp/terraform-provider-aws/issues/42562))
* resource/aws_ssoadmin_application: Updates value of `arn` when refreshing state. ([#43273](https://github.com/hashicorp/terraform-provider-aws/issues/43273))

## 6.2.0 (July  2, 2025)

NOTES:

* resource/aws_s3_bucket_object: The format of the `id` attribute has changed from `key` to `bucket`**/**`key`. All configurations using `id` should be updated to use the `key` attribute instead ([#43119](https://github.com/hashicorp/terraform-provider-aws/issues/43119))
* resource/aws_s3_object: The format of the `id` attribute has changed from `key` to `bucket`**/**`key`. All configurations using `id` should be updated to use the `key` attribute instead ([#43119](https://github.com/hashicorp/terraform-provider-aws/issues/43119))

ENHANCEMENTS:

* data-source/aws_kinesis_stream_consumer: Add `tags` attribute. This functionality requires the `kinesis:ListTagsForResource` IAM permission ([#43173](https://github.com/hashicorp/terraform-provider-aws/issues/43173))
* data-source/aws_networkfirewall_firewall_policy: Add `firewall_policy.stateful_rule_group_reference.deep_threat_inspection` attribute ([#43137](https://github.com/hashicorp/terraform-provider-aws/issues/43137))
* resource/aws_accessanalyzer_analyzer: Add `configuration.internal_access` argument ([#43138](https://github.com/hashicorp/terraform-provider-aws/issues/43138))
* resource/aws_amplify_app: Add `job_config` argument ([#43136](https://github.com/hashicorp/terraform-provider-aws/issues/43136))
* resource/aws_amplify_branch: Add `enable_skew_protection` argument ([#43218](https://github.com/hashicorp/terraform-provider-aws/issues/43218))
* resource/aws_cloudtrail: Support `errorCode`, `eventType`, `sessionCredentialFromConsole`, and `vpcEndpointId` as valid values for `advanced_event_selector.field_selector.field` ([#43091](https://github.com/hashicorp/terraform-provider-aws/issues/43091))
* resource/aws_cloudtrail_event_data_store: Support `errorCode`, `eventType`, `sessionCredentialFromConsole`, and `vpcEndpointId` as valid values for `advanced_event_selector.field_selector.field` ([#43091](https://github.com/hashicorp/terraform-provider-aws/issues/43091))
* resource/aws_cloudwatch_event_archive: Add `kms_key_identifier` argument ([#43139](https://github.com/hashicorp/terraform-provider-aws/issues/43139))
* resource/aws_cloudwatch_log_group: Support `DELIVERY` as a valid value for `log_group_class` ([#42658](https://github.com/hashicorp/terraform-provider-aws/issues/42658))
* resource/aws_codebuild_project: Add `environment.docker_server` configuration block ([#42982](https://github.com/hashicorp/terraform-provider-aws/issues/42982))
* resource/aws_eks_pod_identity_association: Add `disable_session_tags` and `target_role_arn` arguments and `external_id` attribute ([#42979](https://github.com/hashicorp/terraform-provider-aws/issues/42979))
* resource/aws_emr_cluster: Add `os_release_label` argument ([#43018](https://github.com/hashicorp/terraform-provider-aws/issues/43018))
* resource/aws_fms_policy: Add `resource_tag_logical_operator` argument ([#43031](https://github.com/hashicorp/terraform-provider-aws/issues/43031))
* resource/aws_glue_job: Support `job_mode` argument ([#42607](https://github.com/hashicorp/terraform-provider-aws/issues/42607))
* resource/aws_kinesis_stream_consumer: Add `tags` argument and `tags_all` attribute. This functionality requires the `kinesis:ListTagsForResource`, `kinesis:TagResource`, and `kinesis:UntagResource` IAM permissions ([#43173](https://github.com/hashicorp/terraform-provider-aws/issues/43173))
* resource/aws_kms_key: Support `HMAC_224`, `HMAC_384`, `HMAC_512`, `ML_DSA_44`, `ML_DSA_65`, and `ML_DSA_87` as valid values for `customer_master_key_spec` ([#43128](https://github.com/hashicorp/terraform-provider-aws/issues/43128))
* resource/aws_lightsail_instance_public_ports: `-1` is now a valid value for `port_info.from_port` and `port_info.to_port` ([#37703](https://github.com/hashicorp/terraform-provider-aws/issues/37703))
* resource/aws_networkfirewall_firewall_policy: Add `firewall_policy.stateful_rule_group_reference.deep_threat_inspection` argument ([#43137](https://github.com/hashicorp/terraform-provider-aws/issues/43137))
* resource/aws_rbin_rule: Add `exclude_resource_tags` argument ([#43189](https://github.com/hashicorp/terraform-provider-aws/issues/43189))
* resource/aws_s3_directory_bucket: Add `tags` argument and `tags_all` attribute. This functionality requires the `s3express:ListTagsForResource`, `s3express:TagResource`, and `s3express:UntagResource` IAM permissions ([#43256](https://github.com/hashicorp/terraform-provider-aws/issues/43256))
* resource/aws_s3tables_table: Add `metadata` argument ([#43112](https://github.com/hashicorp/terraform-provider-aws/issues/43112))
* resource/aws_wafv2_web_acl: Add `aws_managed_rules_anti_ddos_rule_set` to `managed_rule_group_configs` configuration block in support of L7 DDoS protection ([#43149](https://github.com/hashicorp/terraform-provider-aws/issues/43149))

BUG FIXES:

* provider: Fix `Unexpected Identity Change` errors for numerous resource types when refreshing resources created or refreshed by Terraform AWS Provider v6.0.0 ([#43221](https://github.com/hashicorp/terraform-provider-aws/issues/43221))
* resource/aws_appflow_connector_profile: Fixes error refreshing resource state ([#43221](https://github.com/hashicorp/terraform-provider-aws/issues/43221))
* resource/aws_bcmdataexports_export: Fixes error when refreshing state with resources created before v6.0.0 ([#43090](https://github.com/hashicorp/terraform-provider-aws/issues/43090))
* resource/aws_bedrockagent_agent: Retry `Exceeded the number of retries on OptLock failure. Too many concurrent requests.` errors during update ([#43179](https://github.com/hashicorp/terraform-provider-aws/issues/43179))
* resource/aws_bedrockagent_agent: Retry `Prepare operation can't be performed on Agent when it is in Preparing state.` errors during prepare ([#43179](https://github.com/hashicorp/terraform-provider-aws/issues/43179))
* resource/aws_bedrockagent_agent: Retry `Update operation can't be performed on Agent when it is in Preparing state.` errors during update ([#43179](https://github.com/hashicorp/terraform-provider-aws/issues/43179))
* resource/aws_bedrockagent_agent_collaborator: Retry `operation can't be performed on Agent when it is in Preparing state.` errors during agent collaborator update and disassociation ([#43179](https://github.com/hashicorp/terraform-provider-aws/issues/43179))
* resource/aws_cloudwatch_query_definition: Support ARNs as valid values for `log_group_names` ([#43183](https://github.com/hashicorp/terraform-provider-aws/issues/43183))
* resource/aws_cur_report_definition: Allow an empty (`""`) value for `s3_prefix`. This fixes a regression introduced in [v6.0.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#600-june-18-2025) ([#43159](https://github.com/hashicorp/terraform-provider-aws/issues/43159))
* resource/aws_elasticsearch_domain: Disable publishing for `log_publishing_options` removed on Update. This prevents a perpetual diff ([#43033](https://github.com/hashicorp/terraform-provider-aws/issues/43033))
* resource/aws_elasticsearch_domain: Fix `ValidationException: The Resource Access Policy specified for the CloudWatch Logs log group ... does not grant sufficient permissions for Amazon Elasticsearch Service to create a log stream` IAM eventual consistency errors on Create ([#43033](https://github.com/hashicorp/terraform-provider-aws/issues/43033))
* resource/aws_lambda_function: Fix perpetual `logging_config` diffs when `log_format` is set to `JSON` and `publish = true` ([#42660](https://github.com/hashicorp/terraform-provider-aws/issues/42660))
* resource/aws_lexv2models_intent: Add semantic equality check for `confirmation_setting.prompt_specification.prompt_attempts_specification` defaults ([#43147](https://github.com/hashicorp/terraform-provider-aws/issues/43147))
* resource/aws_opensearch_domain: Disable publishing for `log_publishing_options` removed on Update. This prevents a perpetual diff ([#43033](https://github.com/hashicorp/terraform-provider-aws/issues/43033))
* resource/aws_opensearch_domain: Fix `ValidationException: The Resource Access Policy specified for the CloudWatch Logs log group ... does not grant sufficient permissions for Amazon Elasticsearch Service to create a log stream` IAM eventual consistency errors on Create ([#43033](https://github.com/hashicorp/terraform-provider-aws/issues/43033))
* resource/aws_quicksight_analysis: `WHOLE` is now a valid value for `definition.sheets.visuals.pie_chart_visual.chart_configuration.donut_options.arc_options.arc_thickness` ([#37116](https://github.com/hashicorp/terraform-provider-aws/issues/37116))
* resource/aws_quicksight_dashboard: `WHOLE` is now a valid value for `definition.sheets.visuals.pie_chart_visual.chart_configuration.donut_options.arc_options.arc_thickness` ([#37116](https://github.com/hashicorp/terraform-provider-aws/issues/37116))
* resource/aws_quicksight_template: `WHOLE` is now a valid value for `definition.sheets.visuals.pie_chart_visual.chart_configuration.donut_options.arc_options.arc_thickness` ([#37116](https://github.com/hashicorp/terraform-provider-aws/issues/37116))
* resource/aws_quicksight_user: Remove [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) from `email` ([#43014](https://github.com/hashicorp/terraform-provider-aws/issues/43014))
* resource/aws_verifiedpermissions_schema: Fix `Value Conversion Error` errors when upgrading existing resources to Terraform AWS Provider v6.0.0 ([#43116](https://github.com/hashicorp/terraform-provider-aws/issues/43116))

## 6.1.0 (June 26, 2025)

> [!IMPORTANT]
> Terraform AWS Provider version `v6.1.0` was [removed](https://github.com/hashicorp/terraform-provider-aws/issues/43213) from the Terraform Registry shortly after release due to a [significant bug](https://github.com/hashicorp/terraform-provider-aws/issues/43199) that could not be remediated quickly.
>
> All changes originally included in the removed release are included in version [`v6.2.0`](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#620-july--2-2025).

## 6.0.0 (June 18, 2025)

BREAKING CHANGES:

* data-source/aws_ami: The severity of the diagnostic returned when `most_recent` is `true` and owner and image ID filter criteria has been increased to an error. Existing configurations which were previously receiving a warning diagnostic will now fail to apply. To prevent this error, set the `owner` argument or include a `filter` block with an `image-id` or `owner-id` name/value pair. To continue using unsafe filter values with `most_recent` set to `true`, set the new `allow_unsafe_filter` argument to `true`. This is not recommended. ([#42114](https://github.com/hashicorp/terraform-provider-aws/issues/42114))
* data-source/aws_ecs_task_definition: Remove `inference_accelerator` attribute. Amazon Elastic Inference reached end of life on April, 2024. ([#42137](https://github.com/hashicorp/terraform-provider-aws/issues/42137))
* data-source/aws_ecs_task_execution: Remove `inference_accelerator_overrides` attribute. Amazon Elastic Inference reached end of life on April, 2024. ([#42137](https://github.com/hashicorp/terraform-provider-aws/issues/42137))
* data-source/aws_elbv2_listener_rule: The `action.authenticate_cognito`, `action.authenticate_oidc`, `action.fixed_response`, `action.forward`, `action.forward.stickiness`, `action.redirect`, `condition.host_header`, `condition.http_header`, `condition.http_request_method`, `condition.path_pattern`, `condition.query_string`, and `condition.source_ip` attributes are now list nested blocks instead of single nested blocks ([#42283](https://github.com/hashicorp/terraform-provider-aws/issues/42283))
* data-source/aws_identitystore_user: `filter` has been removed ([#42325](https://github.com/hashicorp/terraform-provider-aws/issues/42325))
* data-source/aws_launch_template: Remove `elastic_inference_accelerator` attribute. Amazon Elastic Inference reached end of life on April, 2024. ([#42137](https://github.com/hashicorp/terraform-provider-aws/issues/42137))
* data-source/aws_launch_template: `elastic_gpu_specifications` has been removed ([#42312](https://github.com/hashicorp/terraform-provider-aws/issues/42312))
* data-source/aws_opensearch_domain: `kibana_endpoint` has been removed ([#42268](https://github.com/hashicorp/terraform-provider-aws/issues/42268))
* data-source/aws_opensearchserverless_security_config: `saml_options` is now a list nested block instead of a single nested block ([#42270](https://github.com/hashicorp/terraform-provider-aws/issues/42270))
* data-source/aws_service_discovery_service: Remove `tags_all` attribute ([#42136](https://github.com/hashicorp/terraform-provider-aws/issues/42136))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_application` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_custom_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_ecs_cluster_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_ganglia_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_haproxy_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_instance` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_java_app_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_memcached_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_mysql_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_nodejs_app_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_permission` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_php_app_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_rails_app_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_rds_db_instance` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_stack` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_static_web_layer` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the AWS OpsWorks Stacks service has reached [End Of Life](https://docs.aws.amazon.com/opsworks/latest/userguide/stacks-eol-faqs.html), the `aws_opsworks_user_profile` resource has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: As the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/welcome.html) does not support Amazon SimpleDB the `aws_simpledb_domain` resource has been removed. Add a [constraint](https://developer.hashicorp.com/terraform/language/providers/requirements#version-constraints) to v5 of the Terraform AWS Provider for continued use of this resource ([#41775](https://github.com/hashicorp/terraform-provider-aws/issues/41775))
* provider: As the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/welcome.html) does not support Amazon Worklink, the `aws_worklink_fleet` resource has been removed ([#42059](https://github.com/hashicorp/terraform-provider-aws/issues/42059))
* provider: As the [AWS SDK for Go v2](https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/welcome.html) does not support Amazon Worklink, the `aws_worklink_website_certificate_authority_association` resource has been removed ([#42059](https://github.com/hashicorp/terraform-provider-aws/issues/42059))
* provider: The `aws_redshift_service_account` resource has been removed. AWS [recommends](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) that a [service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) should be used instead of an AWS account ID in any relevant IAM policy ([#41941](https://github.com/hashicorp/terraform-provider-aws/issues/41941))
* provider: The `endpoints.iotanalytics` and `endpoints.iotevents` configuration arguments have been removed ([#42703](https://github.com/hashicorp/terraform-provider-aws/issues/42703))
* provider: The `endpoints.opsworks` configuration argument has been removed ([#41948](https://github.com/hashicorp/terraform-provider-aws/issues/41948))
* provider: The `endpoints.simpledb` and `endpoints.sdb` configuration arguments have been removed ([#41775](https://github.com/hashicorp/terraform-provider-aws/issues/41775))
* provider: The `endpoints.worklink` configuration argument has been removed ([#42059](https://github.com/hashicorp/terraform-provider-aws/issues/42059))
* resource/aws_accessanalyzer_archive_rule: `filter.exists` now only accepts one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_alb_target_group: `preserve_client_ip` now only accepts one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_api_gateway_account: The `reset_on_delete` argument has been removed ([#42226](https://github.com/hashicorp/terraform-provider-aws/issues/42226))
* resource/aws_api_gateway_deployment: Remove `canary_settings`, `execution_arn`, `invoke_url`, `stage_description`, and `stage_name` arguments. Instead, use the `aws_api_gateway_stage` resource to manage stages. ([#42249](https://github.com/hashicorp/terraform-provider-aws/issues/42249))
* resource/aws_batch_compute_environment: Rename `compute_environment_name` to `name`
resource/aws_batch_compute_environment: Rename `compute_environment_name_prefix` to `name_prefix` ([#38050](https://github.com/hashicorp/terraform-provider-aws/issues/38050))
* resource/aws_batch_compute_environment_data_source: Rename `compute_environment_name` to `name` ([#38050](https://github.com/hashicorp/terraform-provider-aws/issues/38050))
* resource/aws_batch_job_queue: Remove deprecated parameter `compute_environments` in place of `compute_environment_order` ([#40751](https://github.com/hashicorp/terraform-provider-aws/issues/40751))
* resource/aws_bedrock_model_invocation_logging_configuration: `logging_config`, `logging_config.cloudwatch_config`, `logging_config.cloudwatch_config.large_data_delivery_s3_config`, and `logging_config.s3_config` are now list nested blocks instead of single nested blocks ([#42307](https://github.com/hashicorp/terraform-provider-aws/issues/42307))
* resource/aws_cloudfront_key_value_store: Attribute `id` is now set to remote object's `Id` instead of `name` ([#42230](https://github.com/hashicorp/terraform-provider-aws/issues/42230))
* resource/aws_cloudfront_response_headers_policy: The `etag` argument is now computed only ([#38448](https://github.com/hashicorp/terraform-provider-aws/issues/38448))
* resource/aws_cloudtrail_event_data_store: `suspend` now only accepts one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_cognito_user_in_group: The `id` attribute is now a comma-delimited string concatenating the `user_pool_id`, `group_name`, and `username` arguments ([#34082](https://github.com/hashicorp/terraform-provider-aws/issues/34082))
* resource/aws_cur_report_definition: The `s3_prefix` argument is now required ([#38446](https://github.com/hashicorp/terraform-provider-aws/issues/38446))
* resource/aws_db_instance: `character_set_name` now cannot be set with `replicate_source_db`, `restore_to_point_in_time`, `s3_import`, or `snapshot_identifier`. ([#42348](https://github.com/hashicorp/terraform-provider-aws/issues/42348))
* resource/aws_dms_endpoint: Remove `s3_settings` attribute. Use `aws_dms_s3_endpoint` instead ([#42379](https://github.com/hashicorp/terraform-provider-aws/issues/42379))
* resource/aws_dx_gateway_association: `vpn_gateway_id` has been removed ([#42323](https://github.com/hashicorp/terraform-provider-aws/issues/42323))
* resource/aws_ec2_spot_instance_fleet: `terminate_instances_on_delete` now only accepts one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_ec2_spot_instance_request: Remove `block_duration_minutes` attribute ([#42060](https://github.com/hashicorp/terraform-provider-aws/issues/42060))
* resource/aws_ecs_task_definition: Remove `inference_accelerator` attribute. Amazon Elastic Inference reached end of life on April, 2024. ([#42137](https://github.com/hashicorp/terraform-provider-aws/issues/42137))
* resource/aws_eip: `vpc` has been removed. Use `domain` instead. ([#42340](https://github.com/hashicorp/terraform-provider-aws/issues/42340))
* resource/aws_eks_addon: `resolve_conflicts` has been removed. Use `resolve_conflicts_on_create` and `resolve_conflicts_on_update` instead. ([#42318](https://github.com/hashicorp/terraform-provider-aws/issues/42318))
* resource/aws_elasticache_cluster: `auto_minor_version_upgrade` now only accepts one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_elasticache_replication_group: `at_rest_encryption_enabled` and `auto_minor_version_upgrade` now only accept one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_elasticache_replication_group: `auth_token_update_strategy` no longer has a default value. If `auth_token` is set, `auth_token_update_strategy` must also be explicitly configured. ([#42336](https://github.com/hashicorp/terraform-provider-aws/issues/42336))
* resource/aws_evidently_feature: `variations.value.bool_value` now only accepts one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_flow_log: `log_group_name` has been removed. Use `log_destination` instead. ([#42333](https://github.com/hashicorp/terraform-provider-aws/issues/42333))
* resource/aws_globalaccelerator_accelerator: The `id` attribute is now computed only ([#42097](https://github.com/hashicorp/terraform-provider-aws/issues/42097))
* resource/aws_guardduty_detector: Deprecates `datasources`. Use `aws_guardduty_detector_feature` resources instead. ([#42436](https://github.com/hashicorp/terraform-provider-aws/issues/42436))
* resource/aws_guardduty_organization_configuration: The `auto_enable` attribute has been removed ([#42251](https://github.com/hashicorp/terraform-provider-aws/issues/42251))
* resource/aws_identitystore_group: `filter` has been removed ([#42325](https://github.com/hashicorp/terraform-provider-aws/issues/42325))
* resource/aws_imagebuilder_container_recipe: `instance_configuration.block_device_mapping.ebs.delete_on_termination` and `instance_configuration.block_device_mapping.ebs.encrypted` now only accept one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_imagebuilder_image_recipe: `block_device_mapping.ebs.delete_on_termination` and `block_device_mapping.ebs.encrypted` now only accept one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_instance: Remove `cpu_core_count` and `cpu_threads_per_core`. Instead, use `cpu_options`. ([#42280](https://github.com/hashicorp/terraform-provider-aws/issues/42280))
* resource/aws_instance: `user_data` now displays cleartext instead of a hash. Base64 encoded content should use `user_data_base64` instead. ([#42078](https://github.com/hashicorp/terraform-provider-aws/issues/42078))
* resource/aws_launch_template:  `block_device_mappings.ebs.delete_on_termination`, `block_device_mappings.ebs.encrypted`, `ebs_optimized`, `network_interfaces.associate_carrier_ip_address`, `network_interfaces.associate_public_ip_address`, `network_interfaces.delete_on_termination`, and `network_interfaces.primary_ipv6` now only accept one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_launch_template: Remove `elastic_inference_accelerator` attribute. Amazon Elastic Inference reached end of life on April, 2024. ([#42137](https://github.com/hashicorp/terraform-provider-aws/issues/42137))
* resource/aws_launch_template: `elastic_gpu_specifications` has been removed ([#42312](https://github.com/hashicorp/terraform-provider-aws/issues/42312))
* resource/aws_lb_listener: `mutual_authentication` attributes `advertise_trust_store_ca_names`, `ignore_client_certificate_expiry`, and `trust_store_arn` are only valid if `mode` is `verify` ([#42326](https://github.com/hashicorp/terraform-provider-aws/issues/42326))
* resource/aws_lb_target_group: `preserve_client_ip` now only accepts one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_mq_broker: `logs.audit` now only accepts one of `""` (empty string), `true`, or `false` ([#42434](https://github.com/hashicorp/terraform-provider-aws/issues/42434))
* resource/aws_networkmanager_core_network: The `base_policy_region` argument has been removed. Use `base_policy_regions` instead. ([#38398](https://github.com/hashicorp/terraform-provider-aws/issues/38398))
* resource/aws_opensearch_domain: `kibana_endpoint` has been removed ([#42268](https://github.com/hashicorp/terraform-provider-aws/issues/42268))
* resource/aws_opensearchserverless_security_config: `saml_options` is now a list nested block instead of a single nested block ([#42270](https://github.com/hashicorp/terraform-provider-aws/issues/42270))
* resource/aws_paymentcryptography_key: `key_attributes` and `key_attributes.key_modes_of_use` are now list nested blocks instead of single nested blocks. ([#42264](https://github.com/hashicorp/terraform-provider-aws/issues/42264))
* resource/aws_quicksight_data_set: `tags_all` has been removed ([#42260](https://github.com/hashicorp/terraform-provider-aws/issues/42260))
* resource/aws_redshift_cluster: Attributes `cluster_public_key`, `cluster_revision_number`, and `endpoint` are now read only and should not be set ([#42119](https://github.com/hashicorp/terraform-provider-aws/issues/42119))
* resource/aws_redshift_cluster: The `logging` attribute has been removed ([#42013](https://github.com/hashicorp/terraform-provider-aws/issues/42013))
* resource/aws_redshift_cluster: The `publicly_accessible` attribute now defaults to `false` ([#41978](https://github.com/hashicorp/terraform-provider-aws/issues/41978))
* resource/aws_redshift_cluster: The `snapshot_copy` attribute has been removed ([#41995](https://github.com/hashicorp/terraform-provider-aws/issues/41995))
* resource/aws_rekognition_stream_processor: `regions_of_interest.bounding_box` is now a list nested block instead of a single nested block ([#41380](https://github.com/hashicorp/terraform-provider-aws/issues/41380))
* resource/aws_resiliencehub_resiliency_policy: `policy`, `policy.az`, `policy.hardware`, `policy.software`, and `policy.region` are now list nested blocks instead of single nested blocks ([#42297](https://github.com/hashicorp/terraform-provider-aws/issues/42297))
* resource/aws_sagemaker_app_image_config: Exactly one `code_editor_app_image_config`, `jupyter_lab_image_config`, or `kernel_gateway_image_config` block must be configured ([#42753](https://github.com/hashicorp/terraform-provider-aws/issues/42753))
* resource/aws_sagemaker_image_version: `id` is now a comma-delimited string concatenating `image_name` and `version` ([#42536](https://github.com/hashicorp/terraform-provider-aws/issues/42536))
* resource/aws_sagemaker_notebook_instance: Remove `accelerator_types` from your configuration—it no longer exists. Instead, use `instance_type` to use [Inferentia](https://docs.aws.amazon.com/sagemaker/latest/dg/neo-supported-cloud.html). ([#42099](https://github.com/hashicorp/terraform-provider-aws/issues/42099))
* resource/aws_ssm_association: Remove `instance_id` argument ([#42224](https://github.com/hashicorp/terraform-provider-aws/issues/42224))
* resource/aws_verifiedpermissions_schema: `definition` is now a list nested block instead of a single nested block ([#42305](https://github.com/hashicorp/terraform-provider-aws/issues/42305))
* resource/aws_wafv2_web_acl: `rule.statement.managed_rule_group_statement.managed_rule_group_configs.aws_managed_rules_bot_control_rule_set.enable_machine_learning` now defaults to `false` ([#39858](https://github.com/hashicorp/terraform-provider-aws/issues/39858))

NOTES:

* data-source/aws_cloudtrail_service_account: This data source is deprecated. AWS recommends using a service principal name instead of an AWS account ID in any relevant IAM policy. ([#42320](https://github.com/hashicorp/terraform-provider-aws/issues/42320))
* data-source/aws_kms_secret: This data source will be removed in a future version ([#42524](https://github.com/hashicorp/terraform-provider-aws/issues/42524))
* data-source/aws_region: The `name` attribute has been deprecated. All configurations using `name` should be updated to use the `region` attribute instead ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))
* data-source/aws_s3_bucket: Add `bucket_region` attribute. Use of the `bucket_region` attribute instead of the `region` attribute is encouraged ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* data-source/aws_servicequotas_templates: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `aws_region` attribute instead ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))
* data-source/aws_ssmincidents_replication_set: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `regions` attribute instead ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* data-source/aws_vpc_endpoint_service: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `service_region` attribute instead ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* data-source/aws_vpc_peering_connection: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `requester_region` attribute instead ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* provider: Support for the global S3 endpoint is deprecated, along with the `s3_us_east_1_regional_endpoint` argument. The ability to use the global S3 endpoint will be removed in `v7.0.0`. ([#42375](https://github.com/hashicorp/terraform-provider-aws/issues/42375))
* resource/aws_cloudformation_stack_set_instance: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `stack_set_instance_region` attribute instead ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* resource/aws_codeconnections_host: Deprecates `id` in favor of `arn` ([#42232](https://github.com/hashicorp/terraform-provider-aws/issues/42232))
* resource/aws_config_aggregate_authorization: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `authorized_aws_region` attribute instead ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* resource/aws_dx_hosted_connection: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `connection_region` attribute instead ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* resource/aws_elasticache_replication_group: The ability to provide an uppercase `engine` value is deprecated ([#42419](https://github.com/hashicorp/terraform-provider-aws/issues/42419))
* resource/aws_elasticache_user: The ability to provide an uppercase `engine` value is deprecated ([#42419](https://github.com/hashicorp/terraform-provider-aws/issues/42419))
* resource/aws_elasticache_user_group: The ability to provide an uppercase `engine` value is deprecated ([#42419](https://github.com/hashicorp/terraform-provider-aws/issues/42419))
* resource/aws_elastictranscoder_pipeline: This resource is deprecated. Use [AWS Elemental MediaConvert](https://aws.amazon.com/blogs/media/migrating-workflows-from-amazon-elastic-transcoder-to-aws-elemental-mediaconvert/) instead. ([#42313](https://github.com/hashicorp/terraform-provider-aws/issues/42313))
* resource/aws_elastictranscoder_preset: This resource is deprecated. Use [AWS Elemental MediaConvert](https://aws.amazon.com/blogs/media/migrating-workflows-from-amazon-elastic-transcoder-to-aws-elemental-mediaconvert/) instead. ([#42313](https://github.com/hashicorp/terraform-provider-aws/issues/42313))
* resource/aws_evidently_feature: This resource is deprecated. Use [AWS AppConfig feature flags](https://aws.amazon.com/blogs/mt/using-aws-appconfig-feature-flags/) instead. ([#42227](https://github.com/hashicorp/terraform-provider-aws/issues/42227))
* resource/aws_evidently_launch: This resource is deprecated. Use [AWS AppConfig feature flags](https://aws.amazon.com/blogs/mt/using-aws-appconfig-feature-flags/) instead. ([#42227](https://github.com/hashicorp/terraform-provider-aws/issues/42227))
* resource/aws_evidently_project: This resource is deprecated. Use [AWS AppConfig feature flags](https://aws.amazon.com/blogs/mt/using-aws-appconfig-feature-flags/) instead. ([#42227](https://github.com/hashicorp/terraform-provider-aws/issues/42227))
* resource/aws_evidently_segment: This resource is deprecated. Use [AWS AppConfig feature flags](https://aws.amazon.com/blogs/mt/using-aws-appconfig-feature-flags/) instead. ([#42227](https://github.com/hashicorp/terraform-provider-aws/issues/42227))
* resource/aws_guardduty_organization_configuration: `datasources` now returns a deprecation warning ([#42251](https://github.com/hashicorp/terraform-provider-aws/issues/42251))
* resource/aws_kinesis_analytics_application: Effective January 27, 2026, AWS will no longer support Kinesis Data Analytics for SQL. This resource is deprecated and will be removed in a future version. Use the `aws_kinesisanalyticsv2_application` resource instead ([#42102](https://github.com/hashicorp/terraform-provider-aws/issues/42102))
* resource/aws_media_store_container: This resource is deprecated. It will be removed in a future version. Use S3, AWS MediaPackage, or other storage solution instead. ([#42265](https://github.com/hashicorp/terraform-provider-aws/issues/42265))
* resource/aws_media_store_container_policy: This resource is deprecated. It will be removed in a future version. Use S3, AWS MediaPackage, or other storage solution instead. ([#42265](https://github.com/hashicorp/terraform-provider-aws/issues/42265))
* resource/aws_redshift_cluster: The default value of `encrypted` is now `true` to match the AWS API. ([#42631](https://github.com/hashicorp/terraform-provider-aws/issues/42631))
* resource/aws_s3_bucket: Add `bucket_region` attribute. Use of the `bucket_region` attribute instead of the `region` attribute is encouraged ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* resource/aws_service_discovery_service: `health_check_custom_config.failure_threshold` is deprecated. The argument is no longer supported by AWS and is always set to 1 ([#40777](https://github.com/hashicorp/terraform-provider-aws/issues/40777))
* resource/aws_servicequotas_template: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `aws_region` attribute instead ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))
* resource/aws_ssmincidents_replication_set: The `region` attribute has been deprecated. All configurations using `region` should be updated to use the `regions` attribute instead ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))

ENHANCEMENTS:

* data-source/aws_ami: Add `allow_unsafe_filter` argument ([#42114](https://github.com/hashicorp/terraform-provider-aws/issues/42114))
* data-source/aws_availability_zone: Add `group_long_name` attribute ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* data-source/aws_availability_zone: Mark `region` as Optional, allowing a value to be configured ([#42014](https://github.com/hashicorp/terraform-provider-aws/issues/42014))
* resource/aws_auditmanager_assessment: Add plan-time validation of `roles.role_arn` and `roles.role_type` ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))
* provider: Add enhanced `region` support to most resources, data sources, and ephemeral resources, allowing per-resource Region targeting without requiring multiple provider configurations. See the [Enhanced Region Support guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/enhanced-region-support) for more information. ([#43075](https://github.com/hashicorp/terraform-provider-aws/issues/43075))
* resource/aws_auditmanager_control: Add plan-time validation of `control_mapping_sources.source_frequency`, `control_mapping_sources.source_set_up_option`, and `control_mapping_sources.source_type` ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))
* resource/aws_auditmanager_framework_share: Add plan-time validation of `destination_account` ([#42741](https://github.com/hashicorp/terraform-provider-aws/issues/42741))
* resource/aws_auditmanager_organization_admin_account_registration: Add plan-time validation of `admin_account_id` ([#42741](https://github.com/hashicorp/terraform-provider-aws/issues/42741))
* resource/aws_cognito_user_in_group: Add import support ([#34082](https://github.com/hashicorp/terraform-provider-aws/issues/34082))
* resource/aws_ecs_service: Add `arn` attribute ([#42733](https://github.com/hashicorp/terraform-provider-aws/issues/42733))
* resource/aws_guardduty_detector: Adds validation to `finding_publishing_frequency`. ([#42436](https://github.com/hashicorp/terraform-provider-aws/issues/42436))
* resource/aws_lb_listener: `mutual_authentication` attribute `trust_store_arn` is required if `mode` is `verify` ([#42326](https://github.com/hashicorp/terraform-provider-aws/issues/42326))
* resource/aws_quicksight_iam_policy_assignment: Add plan-time validation of `policy_arn` ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))
* resource/aws_sagemaker_image_version: Add `aliases` argument ([#42610](https://github.com/hashicorp/terraform-provider-aws/issues/42610))
* resource/aws_securitylake_subscriber: Add plan-time validation of `access_type` `source.aws_log_source_resource.source_name`, and `subscriber_identity.external_id` ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))

BUG FIXES:

* resource/aws_auditmanager_control: Fix `Provider produced inconsistent result after apply` errors ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))
* resource/aws_redshift_cluster: Fixes permanent diff when `encrypted` is not explicitly set to `true`. ([#42631](https://github.com/hashicorp/terraform-provider-aws/issues/42631))
* resource/aws_rekognition_stream_processor: Fix `regions_of_interest.bounding_box` and `regions_of_interest.polygon` argument validation ([#41380](https://github.com/hashicorp/terraform-provider-aws/issues/41380))
* resource/aws_sagemaker_image_version: Read the correct image version after creation rather than always fetching the latest ([#42536](https://github.com/hashicorp/terraform-provider-aws/issues/42536))
* resource/aws_securitylake_subscriber: Change `access_type` to [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) ([#42131](https://github.com/hashicorp/terraform-provider-aws/issues/42131))

## Previous Releases

For information on prior major releases, see their changelogs:

* [5.x](https://github.com/hashicorp/terraform-provider-aws/blob/release/5.x/CHANGELOG.md)
* [4.x](https://github.com/hashicorp/terraform-provider-aws/blob/release/4.x/CHANGELOG.md)
* [3.x](https://github.com/hashicorp/terraform-provider-aws/blob/release/3.x/CHANGELOG.md)
* [2.x and earlier](https://github.com/hashicorp/terraform-provider-aws/blob/release/2.x/CHANGELOG.md)
