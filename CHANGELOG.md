## 6.32.0 (Unreleased)

FEATURES:

* **New List Resource:** `aws_secretsmanager_secret` ([#46318](https://github.com/hashicorp/terraform-provider-aws/issues/46318))

## 6.31.0 (February 4, 2026)

NOTES:

* resource/aws_s3_bucket_abac: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_abac: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_accelerate_configuration: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_accelerate_configuration: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_acl: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_acl: Removes `expected_bucket_owner` and `acl` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_cors_configuration: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_cors_configuration: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_lifecycle_configuration: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_lifecycle_configuration: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_logging: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_logging: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_metadata_configuration: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_metadata_configuration: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_object_lock_configuration: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_object_lock_configuration: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_request_payment_configuration: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_request_payment_configuration: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_server_side_encryption_configuration: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_server_side_encryption_configuration: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_versioning: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_versioning: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))
* resource/aws_s3_bucket_website_configuration: Deprecates `expected_bucket_owner` attribute. ([#46262](https://github.com/hashicorp/terraform-provider-aws/issues/46262))
* resource/aws_s3_bucket_website_configuration: Removes `expected_bucket_owner` attribute from Resource Identity. ([#46272](https://github.com/hashicorp/terraform-provider-aws/issues/46272))

FEATURES:

* **New Data Source:** `aws_account_regions` ([#41746](https://github.com/hashicorp/terraform-provider-aws/issues/41746))
* **New Ephemeral Resource:** `aws_ecrpublic_authorization_token` ([#45841](https://github.com/hashicorp/terraform-provider-aws/issues/45841))
* **New List Resource:** `aws_cloudwatch_event_rule` ([#46304](https://github.com/hashicorp/terraform-provider-aws/issues/46304))
* **New List Resource:** `aws_cloudwatch_event_target` ([#46297](https://github.com/hashicorp/terraform-provider-aws/issues/46297))
* **New List Resource:** `aws_cloudwatch_metric_alarm` ([#46268](https://github.com/hashicorp/terraform-provider-aws/issues/46268))
* **New List Resource:** `aws_iam_role_policy` ([#46293](https://github.com/hashicorp/terraform-provider-aws/issues/46293))
* **New List Resource:** `aws_lambda_function` ([#46295](https://github.com/hashicorp/terraform-provider-aws/issues/46295))
* **New List Resource:** `aws_s3_bucket_acl` ([#46305](https://github.com/hashicorp/terraform-provider-aws/issues/46305))
* **New List Resource:** `aws_s3_bucket_policy` ([#46312](https://github.com/hashicorp/terraform-provider-aws/issues/46312))
* **New List Resource:** `aws_s3_bucket_public_access_block` ([#46309](https://github.com/hashicorp/terraform-provider-aws/issues/46309))
* **New Resource:** `aws_ssoadmin_customer_managed_policy_attachments_exclusive` ([#46191](https://github.com/hashicorp/terraform-provider-aws/issues/46191))

ENHANCEMENTS:

* resource/aws_odb_cloud_autonomous_vm_cluster: autonomous vm cluster creation using odb network ARN and exadata infrastructure ARN for resource sharing model. ([#45583](https://github.com/hashicorp/terraform-provider-aws/issues/45583))
* resource/aws_opensearch_domain: Add `serverless_vector_acceleration` to `aiml_options` ([#45882](https://github.com/hashicorp/terraform-provider-aws/issues/45882))

BUG FIXES:

* list-resource/aws_s3_bucket: Restricts listed buckets to expected region. ([#46305](https://github.com/hashicorp/terraform-provider-aws/issues/46305))
* resource/aws_elasticache_replication_group: Fixed AUTH to RBAC migration. Previously, `auth_token_update_strategy` always required `auth_token`, which caused an error when migrating from AUTH to RBAC. Now, `auth_token_update_strategy` still requires `auth_token` except when `auth_token_update_strategy` is `DELETE`. ([#45518](https://github.com/hashicorp/terraform-provider-aws/issues/45518))
* resource/aws_elasticache_replication_group: Fixed an issue with downscaling `aws_elasticache_replication_group` when `cluster_mode="enabled"` and `num_node_groups` is reduced. Previously, downscaling could fail in certain scenarios; for example, if nodes `0001`, `0002`, `0003`, `0004`, and `0005` exist, and a user manually removes `0003` and `0005`, then sets `num_node_groups = 2`, terraform would attempt to delete `0003`, `0004`, and `0005`. This is now fixed, after this fix terraform will retrieve the current node groups before resizing. ([#45893](https://github.com/hashicorp/terraform-provider-aws/issues/45893))
* resource/aws_elasticache_serverless_cache: Fix `user_group_id` removal during modification. ([#45571](https://github.com/hashicorp/terraform-provider-aws/issues/45571))
* resource/aws_elasticache_serverless_cache: Fix forced replacement when upgrading Valkey major version or switching engine between redis and valkey ([#45087](https://github.com/hashicorp/terraform-provider-aws/issues/45087))
* resource/aws_network_interface: Fix `UnauthorizedOperation` error when detaching resource that does not have an attachment ([#46211](https://github.com/hashicorp/terraform-provider-aws/issues/46211))

## 6.30.0 (January 28, 2026)

FEATURES:

* **New Resource:** `aws_ssoadmin_managed_policy_attachments_exclusive` ([#46176](https://github.com/hashicorp/terraform-provider-aws/issues/46176))

BUG FIXES:

* resource/aws_dynamodb_table: Fix panic when `global_secondary_index` or `global_secondary_index.key_schema` are `dynamic` ([#46195](https://github.com/hashicorp/terraform-provider-aws/issues/46195))

## 6.29.0 (January 28, 2026)

NOTES:

* data-source/aws_organizations_organization: Add `return_organization_only` argument to return only the results of the [`DescribeOrganization`](https://docs.aws.amazon.com/organizations/latest/APIReference/API_DescribeOrganization.html) API and avoid API limits ([#40884](https://github.com/hashicorp/terraform-provider-aws/issues/40884))
* resource/aws_cloudfront_anycast_ip_list: Because we cannot easily test all this functionality, it is best effort and we ask for community help in testing ([#43331](https://github.com/hashicorp/terraform-provider-aws/issues/43331))
* resource/aws_invoicing_invoice_unit: Deprecates `region` attribute, as the resource is global. ([#46185](https://github.com/hashicorp/terraform-provider-aws/issues/46185))
* resource/aws_organizations_organization: Add `return_organization_only` argument to return only the results of the [`DescribeOrganization`](https://docs.aws.amazon.com/organizations/latest/APIReference/API_DescribeOrganization.html) API and avoid API limits ([#40884](https://github.com/hashicorp/terraform-provider-aws/issues/40884))
* resource/aws_savingsplans_savings_plan: Because we cannot easily test this functionality, it is best effort and we ask for community help in testing ([#45834](https://github.com/hashicorp/terraform-provider-aws/issues/45834))

FEATURES:

* **New Data Source:** `aws_arcregionswitch_plan` ([#43781](https://github.com/hashicorp/terraform-provider-aws/issues/43781))
* **New Data Source:** `aws_arcregionswitch_route53_health_checks` ([#43781](https://github.com/hashicorp/terraform-provider-aws/issues/43781))
* **New Data Source:** `aws_organizations_entity_path` ([#45890](https://github.com/hashicorp/terraform-provider-aws/issues/45890))
* **New Data Source:** `aws_resourcegroupstaggingapi_required_tags` ([#45994](https://github.com/hashicorp/terraform-provider-aws/issues/45994))
* **New Data Source:** `aws_s3_bucket_object_lock_configuration` ([#45990](https://github.com/hashicorp/terraform-provider-aws/issues/45990))
* **New Data Source:** `aws_s3_bucket_replication_configuration` ([#42662](https://github.com/hashicorp/terraform-provider-aws/issues/42662))
* **New Data Source:** `aws_s3control_access_points` ([#45949](https://github.com/hashicorp/terraform-provider-aws/issues/45949))
* **New Data Source:** `aws_s3control_multi_region_access_points` ([#45974](https://github.com/hashicorp/terraform-provider-aws/issues/45974))
* **New Data Source:** `aws_savingsplans_savings_plan` ([#45834](https://github.com/hashicorp/terraform-provider-aws/issues/45834))
* **New Data Source:** `aws_wafv2_managed_rule_group` ([#45899](https://github.com/hashicorp/terraform-provider-aws/issues/45899))
* **New List Resource:** `aws_appflow_connector_profile` ([#45983](https://github.com/hashicorp/terraform-provider-aws/issues/45983))
* **New List Resource:** `aws_appflow_flow` ([#45980](https://github.com/hashicorp/terraform-provider-aws/issues/45980))
* **New List Resource:** `aws_cleanrooms_collaboration` ([#45953](https://github.com/hashicorp/terraform-provider-aws/issues/45953))
* **New List Resource:** `aws_cleanrooms_configured_table` ([#45956](https://github.com/hashicorp/terraform-provider-aws/issues/45956))
* **New List Resource:** `aws_cloudfront_key_value_store` ([#45957](https://github.com/hashicorp/terraform-provider-aws/issues/45957))
* **New List Resource:** `aws_opensearchserverless_collection` ([#46001](https://github.com/hashicorp/terraform-provider-aws/issues/46001))
* **New List Resource:** `aws_route53_record` ([#46059](https://github.com/hashicorp/terraform-provider-aws/issues/46059))
* **New List Resource:** `aws_s3_bucket` ([#46004](https://github.com/hashicorp/terraform-provider-aws/issues/46004))
* **New List Resource:** `aws_s3_object` ([#46002](https://github.com/hashicorp/terraform-provider-aws/issues/46002))
* **New List Resource:** `aws_security_group` ([#46062](https://github.com/hashicorp/terraform-provider-aws/issues/46062))
* **New Resource:** `aws_apigatewayv2_routing_rule` ([#42961](https://github.com/hashicorp/terraform-provider-aws/issues/42961))
* **New Resource:** `aws_arcregionswitch_plan` ([#43781](https://github.com/hashicorp/terraform-provider-aws/issues/43781))
* **New Resource:** `aws_cloudfront_anycast_ip_list` ([#43331](https://github.com/hashicorp/terraform-provider-aws/issues/43331))
* **New Resource:** `aws_notifications_managed_notification_account_contact_association` ([#45185](https://github.com/hashicorp/terraform-provider-aws/issues/45185))
* **New Resource:** `aws_notifications_managed_notification_additional_channel_association` ([#45186](https://github.com/hashicorp/terraform-provider-aws/issues/45186))
* **New Resource:** `aws_notifications_organizational_unit_association` ([#45197](https://github.com/hashicorp/terraform-provider-aws/issues/45197))
* **New Resource:** `aws_notifications_organizations_access` ([#45273](https://github.com/hashicorp/terraform-provider-aws/issues/45273))
* **New Resource:** `aws_opensearch_application` ([#43822](https://github.com/hashicorp/terraform-provider-aws/issues/43822))
* **New Resource:** `aws_ram_permission` ([#44114](https://github.com/hashicorp/terraform-provider-aws/issues/44114))
* **New Resource:** `aws_ram_resource_associations_exclusive` ([#45883](https://github.com/hashicorp/terraform-provider-aws/issues/45883))
* **New Resource:** `aws_sagemaker_labeling_job` ([#46041](https://github.com/hashicorp/terraform-provider-aws/issues/46041))
* **New Resource:** `aws_sagemaker_model_card` ([#45993](https://github.com/hashicorp/terraform-provider-aws/issues/45993))
* **New Resource:** `aws_sagemaker_model_card_export_job` ([#46009](https://github.com/hashicorp/terraform-provider-aws/issues/46009))
* **New Resource:** `aws_savingsplans_savings_plan` ([#45834](https://github.com/hashicorp/terraform-provider-aws/issues/45834))
* **New Resource:** `aws_sesv2_tenant_resource_association` ([#45904](https://github.com/hashicorp/terraform-provider-aws/issues/45904))
* **New Resource:** `aws_vpc_security_group_rules_exclusive` ([#45876](https://github.com/hashicorp/terraform-provider-aws/issues/45876))

ENHANCEMENTS:

* aws_api_gateway_domain_name: Add `routing_mode` argument to support dynamic routing via routing rules ([#42961](https://github.com/hashicorp/terraform-provider-aws/issues/42961))
* aws_apigatewayv2_domain_name: Add `routing_mode` argument to support dynamic routing via routing rules ([#42961](https://github.com/hashicorp/terraform-provider-aws/issues/42961))
* data-source/aws_batch_job_definition: Add `allow_privilege_escalation` attribute to `eks_properties.pod_properties.containers.security_context` ([#45896](https://github.com/hashicorp/terraform-provider-aws/issues/45896))
* data-source/aws_dynamodb_table: Add `global_secondary_index.key_schema` attribute ([#46157](https://github.com/hashicorp/terraform-provider-aws/issues/46157))
* data-source/aws_networkmanager_core_network_policy_document: Add `segment_actions.routing_policy_names` argument ([#45928](https://github.com/hashicorp/terraform-provider-aws/issues/45928))
* data-source/aws_s3_object: Add `body_base64` and `download_body` attributes. For improved performance, set `download_body = false` to ensure bodies are never downloaded ([#46163](https://github.com/hashicorp/terraform-provider-aws/issues/46163))
* data-source/aws_vpc_ipam_pool: Add `source_resource` attribute ([#44705](https://github.com/hashicorp/terraform-provider-aws/issues/44705))
* resource/aws_batch_job_definition: Add `allow_privilege_escalation` attribute to `eks_properties.pod_properties.containers.security_context` ([#45896](https://github.com/hashicorp/terraform-provider-aws/issues/45896))
* resource/aws_bedrockagent_data_source: Add `vector_ingestion_configuration.parsing_configuration.bedrock_data_automation_configuration` block ([#45966](https://github.com/hashicorp/terraform-provider-aws/issues/45966))
* resource/aws_bedrockagent_data_source: Add `vector_ingestion_configuration.parsing_configuration.bedrock_foundation_model_configuration.parsing_modality` argument ([#46056](https://github.com/hashicorp/terraform-provider-aws/issues/46056))
* resource/aws_docdb_cluster_instance: Add `certificate_rotation_restart` argument ([#45984](https://github.com/hashicorp/terraform-provider-aws/issues/45984))
* resource/aws_dynamodb_table: Add support for multi-attribute keys in global secondary indexes. Introduces hash_keys and range_keys to the gsi block and makes hash_key optional for backwards compatibility. ([#45357](https://github.com/hashicorp/terraform-provider-aws/issues/45357))
* resource/aws_dynamodb_table: Adds warning when `stream_view_type` is set and `stream_enabled` is either `false` or unset. ([#45934](https://github.com/hashicorp/terraform-provider-aws/issues/45934))
* resource/aws_ecr_account_setting: Add support for `BLOB_MOUNTING` account setting name with `ENABLED` and `DISABLED` values ([#46092](https://github.com/hashicorp/terraform-provider-aws/issues/46092))
* resource/aws_fsx_windows_file_system: Add `domain_join_service_account_secret` argument to `self_managed_active_directory` configuration block ([#45852](https://github.com/hashicorp/terraform-provider-aws/issues/45852))
* resource/aws_fsx_windows_file_system: Change `self_managed_active_directory.password` to Optional and `self_managed_active_directory.username` to Optional and Computed ([#45852](https://github.com/hashicorp/terraform-provider-aws/issues/45852))
* resource/aws_invoicing_invoice_unit: Adds resource identity support. ([#46185](https://github.com/hashicorp/terraform-provider-aws/issues/46185))
* resource/aws_invoicing_invoice_unit: Adds validation to restrict `rules` to a single element. ([#46185](https://github.com/hashicorp/terraform-provider-aws/issues/46185))
* resource/aws_lambda_function: Increase upper limit of `memory_size` from 10240 MB to 32768 MB ([#46065](https://github.com/hashicorp/terraform-provider-aws/issues/46065))
* resource/aws_launch_template: Add `network_performance_options` argument ([#46071](https://github.com/hashicorp/terraform-provider-aws/issues/46071))
* resource/aws_odb_network: Enhancements to support KMS and STS parameters in CreateOdbNetwork and UpdateOdbNetwork. ([#45636](https://github.com/hashicorp/terraform-provider-aws/issues/45636))
* resource/aws_opensearchserverless_collection: Add resource identity support ([#45981](https://github.com/hashicorp/terraform-provider-aws/issues/45981))
* resource/aws_osis_pipeline: Updates `pipeline_configuration_body` maximum length validation to 2,621,440 bytes to align with AWS API specification. ([#44881](https://github.com/hashicorp/terraform-provider-aws/issues/44881))
* resource/aws_sagemaker_endpoint: Retry IAM eventual consistency errors on Create ([#45951](https://github.com/hashicorp/terraform-provider-aws/issues/45951))
* resource/aws_sagemaker_monitoring_schedule: Add `monitoring_schedule_config.monitoring_job_definition` argument ([#45951](https://github.com/hashicorp/terraform-provider-aws/issues/45951))
* resource/aws_sagemaker_monitoring_schedule: Make `monitoring_schedule_config.monitoring_job_definition_name` argument optional ([#45951](https://github.com/hashicorp/terraform-provider-aws/issues/45951))
* resource/aws_vpc_ipam_pool: Add `source_resource` argument in support of provisioning of VPC Resource Planning Pools ([#44705](https://github.com/hashicorp/terraform-provider-aws/issues/44705))
* resource/aws_vpc_ipam_resource_discovery: Add `organizational_unit_exclusion` argument ([#45890](https://github.com/hashicorp/terraform-provider-aws/issues/45890))
* resource/aws_vpc_subnet: Add `ipv4_ipam_pool_id`, `ipv4_netmask_length`, `ipv6_ipam_pool_id`, and `ipv6_netmask_length` arguments in support of provisioning of subnets using IPAM ([#44705](https://github.com/hashicorp/terraform-provider-aws/issues/44705))
* resource/aws_vpc_subnet: Change `ipv6_cidr_block` to Optional and Computed ([#44705](https://github.com/hashicorp/terraform-provider-aws/issues/44705))

BUG FIXES:

* data-source/aws_ecr_lifecycle_policy_document: Add `rule.action.target_storage_class` and `rule.selection.storage_class` to JSON serialization ([#45909](https://github.com/hashicorp/terraform-provider-aws/issues/45909))
* data-source/aws_lakeformation_permissions: Remove incorrect validation from `catalog_id`, `data_location.catalog_id`, `database.catalog_id`, `lf_tag_policy.catalog_id`, `table.catalog_id`, and `table_with_columns.catalog_id` arguments ([#43931](https://github.com/hashicorp/terraform-provider-aws/issues/43931))
* data-source/aws_networkmanager_core_network_policy_document: Fix panic when `attachment_routing_policy_rules.action.associate_routing_policies` is empty ([#46160](https://github.com/hashicorp/terraform-provider-aws/issues/46160))
* provider: Fix crash when using custom S3 endpoints with non-standard region strings (e.g., S3-compatible storage like Ceph or MinIO) ([#46000](https://github.com/hashicorp/terraform-provider-aws/issues/46000))
* provider: When importing resources with `region` defined, in AWS European Sovereign Cloud, prevent failing due to region validation requiring region names to start with "[a-z]{2}-" ([#45895](https://github.com/hashicorp/terraform-provider-aws/issues/45895))
* resource/aws_athena_workgroup: Fix error when removing `configuration.result_configuration.encryption_configuration` argument ([#46159](https://github.com/hashicorp/terraform-provider-aws/issues/46159))
* resource/aws_bcmdataexports_export: Fix `Provider produced inconsistent result after apply` error when querying `CARBON_EMISSIONS` table without `table_configurations` ([#45972](https://github.com/hashicorp/terraform-provider-aws/issues/45972))
* resource/aws_bedrock_inference_profile: Fixed forced replacement following import when `model_source` is set ([#45713](https://github.com/hashicorp/terraform-provider-aws/issues/45713))
* resource/aws_billing_view: Fix handling of data_filter_expression ([#45293](https://github.com/hashicorp/terraform-provider-aws/issues/45293))
* resource/aws_cloudformation_stack_set: Fix perpetual diff when using `auto_deployment` with `permission_model` set to `SERVICE_MANAGED` ([#45992](https://github.com/hashicorp/terraform-provider-aws/issues/45992))
* resource/aws_cloudfront_distribution: Fix `runtime error: invalid memory address or nil pointer dereference` panic when mistakenly importing a multi-tenant distribution ([#45873](https://github.com/hashicorp/terraform-provider-aws/issues/45873))
* resource/aws_cloudfront_distribution: Prevent mistakenly importing a multi-tenant distribution ([#45873](https://github.com/hashicorp/terraform-provider-aws/issues/45873))
* resource/aws_cloudfront_multitenant_distribution: Fix "specified origin server does not exist or is not valid" errors when attempting to use Origin Access Control (OAC) ([#45977](https://github.com/hashicorp/terraform-provider-aws/issues/45977))
* resource/aws_cloudfront_multitenant_distribution: Fix `origin_group` to use correct `id` attribute name and fix field mapping to resolve `missing required field` errors ([#45921](https://github.com/hashicorp/terraform-provider-aws/issues/45921))
* resource/aws_cloudwatch_event_rule: Prevent failing on AWS European Sovereign Cloud regions due to region validation requiring region names to start with "[a-z]{2}-" ([#45895](https://github.com/hashicorp/terraform-provider-aws/issues/45895))
* resource/aws_config_configuration_recorder: Fix `InvalidRecordingGroupException: The recording group provided is not valid` errors when the `recording_group.exclusion_by_resource_type` or `recording_group.recording_strategy` argument is removed during update ([#46110](https://github.com/hashicorp/terraform-provider-aws/issues/46110))
* resource/aws_datazone_environment_profile: Prevent failing on AWS European Sovereign Cloud regions due to region validation requiring region names to start with "[a-z]{2}-" ([#45895](https://github.com/hashicorp/terraform-provider-aws/issues/45895))
* resource/aws_dynamodb_table: Fix perpetual diff for `warm_throughput` in global_secondary_index when not set in configuration. ([#46094](https://github.com/hashicorp/terraform-provider-aws/issues/46094))
* resource/aws_dynamodb_table: Fixes error when `name` is known after apply ([#45917](https://github.com/hashicorp/terraform-provider-aws/issues/45917))
* resource/aws_eks_cluster: Fix `kubernetes_network_config` argument name in EKS Auto Mode validation error message ([#45997](https://github.com/hashicorp/terraform-provider-aws/issues/45997))
* resource/aws_emrserverless_application: Prevent failing on AWS European Sovereign Cloud regions due to region validation requiring region names to start with "[a-z]{2}-" ([#45895](https://github.com/hashicorp/terraform-provider-aws/issues/45895))
* resource/aws_lakeformation_permissions: Remove incorrect validation from `catalog_id`, `data_location.catalog_id`, `database.catalog_id`, `lf_tag_policy.catalog_id`, `table.catalog_id`, and `table_with_columns.catalog_id` arguments ([#43931](https://github.com/hashicorp/terraform-provider-aws/issues/43931))
* resource/aws_lambda_event_source_mapping: Prevent failing on AWS European Sovereign Cloud regions due to region validation requiring region names to start with "[a-z]{2}-" ([#45895](https://github.com/hashicorp/terraform-provider-aws/issues/45895))
* resource/aws_lambda_invocation: Fix panic when deleting or replacing resource with empty input in CRUD lifecycle scope ([#45967](https://github.com/hashicorp/terraform-provider-aws/issues/45967))
* resource/aws_lambda_permission: Prevent failing on AWS European Sovereign Cloud regions due to region validation requiring region names to start with "[a-z]{2}-" ([#45895](https://github.com/hashicorp/terraform-provider-aws/issues/45895))
* resource/aws_lb_target_group: Fix update error when switching `health_check.protocol` from `HTTP` to `TCP` when `protocol` is `TCP` ([#46036](https://github.com/hashicorp/terraform-provider-aws/issues/46036))
* resource/aws_multitenant_cloudfront_distribution: Prevent mistakenly importing a standard distribution ([#45873](https://github.com/hashicorp/terraform-provider-aws/issues/45873))
* resource/aws_networkfirewall_firewall_policy: Support partner-managed rule groups via `firewall_policy.stateful_rule_group_reference.resource_arn` ([#46124](https://github.com/hashicorp/terraform-provider-aws/issues/46124))
* resource/aws_odb_network: Fix `delete_associated_resources` being set when value is unknown ([#45636](https://github.com/hashicorp/terraform-provider-aws/issues/45636))
* resource/aws_pipes_pipe: Prevent failing on AWS European Sovereign Cloud regions due to region validation requiring region names to start with "[a-z]{2}-" ([#45895](https://github.com/hashicorp/terraform-provider-aws/issues/45895))
* resource/aws_placement_group: Correct validation of `partition_count` ([#45042](https://github.com/hashicorp/terraform-provider-aws/issues/45042))
* resource/aws_rds_cluster: Properly set `iam_database_authentication_enabled` when restored from snapshot ([#39461](https://github.com/hashicorp/terraform-provider-aws/issues/39461))
* resource/aws_redshift_cluster: Changing `port` now works. ([#45870](https://github.com/hashicorp/terraform-provider-aws/issues/45870))
* resource/aws_redshiftserverless_workgroup: Fix `ValidationException: Base capacity cannot be updated when PerformanceTarget is Enabled` error when updating `price_performance_target` and `base_capacity` ([#46137](https://github.com/hashicorp/terraform-provider-aws/issues/46137))
* resource/aws_route53_health_check: Mark `regions` argument as `Computed` to fix an unexpected `regions` diff when it is not specified ([#45829](https://github.com/hashicorp/terraform-provider-aws/issues/45829))
* resource/aws_route53_zone: Fix `InvalidChangeBatch` errors during [ForceNew](https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew) operations when zone name changes ([#45242](https://github.com/hashicorp/terraform-provider-aws/issues/45242))
* resource/aws_route53_zone: Fixes error where Delete would fail if the remote resource had already been deleted. ([#45985](https://github.com/hashicorp/terraform-provider-aws/issues/45985))
* resource/aws_route53profiles_resource_association: Fix `Invalid JSON String Value` error on initial apply and `ConflictException` on subsequent apply when associating Route53 Resolver Query Log Configs ([#45958](https://github.com/hashicorp/terraform-provider-aws/issues/45958))
* resource/aws_route53recoverycontrolconfig_control_panel: Fix crash when create returns an error ([#45954](https://github.com/hashicorp/terraform-provider-aws/issues/45954))
* resource/aws_s3_bucket: Fix bucket creation with tags in non-commercial AWS regions by handling `UnsupportedArgument` errors during tag-on-create operations ([#46122](https://github.com/hashicorp/terraform-provider-aws/issues/46122))
* resource/aws_s3_bucket: Fix tag read and update operations in non-commercial AWS regions by handling `MethodNotAllowed` errors when S3 Control APIs are unavailable ([#46122](https://github.com/hashicorp/terraform-provider-aws/issues/46122))
* resource/aws_servicecatalog_portfolio_share: Support organization and OU IDs in addition to ARNs for GovCloud compatibility ([#39863](https://github.com/hashicorp/terraform-provider-aws/issues/39863))
* resource/aws_subnet: Mark `ipv6_cidr_block` as `ForceNew` when the existing IPv6 subnet was created with `assign_ipv6_address_on_create = true` ([#46043](https://github.com/hashicorp/terraform-provider-aws/issues/46043))
* resource/aws_vpc_endpoint: Fix persistent diffs caused by case differences in `ip_address_type` ([#45947](https://github.com/hashicorp/terraform-provider-aws/issues/45947))

## 6.28.0 (January 7, 2026)

NOTES:

* resource/aws_dynamodb_global_secondary_index: This resource type is experimental.  The schema or behavior may change without notice, and it is not subject to the backwards compatibility guarantee of the provider. ([#44999](https://github.com/hashicorp/terraform-provider-aws/issues/44999))

FEATURES:

* **New Data Source:** `aws_cloudfront_connection_group` ([#44885](https://github.com/hashicorp/terraform-provider-aws/issues/44885))
* **New Data Source:** `aws_cloudfront_distribution_tenant` ([#45088](https://github.com/hashicorp/terraform-provider-aws/issues/45088))
* **New List Resource:** `aws_kms_alias` ([#45700](https://github.com/hashicorp/terraform-provider-aws/issues/45700))
* **New List Resource:** `aws_sqs_queue` ([#45691](https://github.com/hashicorp/terraform-provider-aws/issues/45691))
* **New Resource:** `aws_cloudfront_connection_function` ([#45664](https://github.com/hashicorp/terraform-provider-aws/issues/45664))
* **New Resource:** `aws_cloudfront_connection_group` ([#44885](https://github.com/hashicorp/terraform-provider-aws/issues/44885))
* **New Resource:** `aws_cloudfront_distribution_tenant` ([#45088](https://github.com/hashicorp/terraform-provider-aws/issues/45088))
* **New Resource:** `aws_cloudfront_multitenant_distribution` ([#45535](https://github.com/hashicorp/terraform-provider-aws/issues/45535))
* **New Resource:** `aws_dynamodb_global_secondary_index` ([#44999](https://github.com/hashicorp/terraform-provider-aws/issues/44999))
* **New Resource:** `aws_ecr_pull_time_update_exclusion` ([#45765](https://github.com/hashicorp/terraform-provider-aws/issues/45765))
* **New Resource:** `aws_organizations_tag` ([#45730](https://github.com/hashicorp/terraform-provider-aws/issues/45730))
* **New Resource:** `aws_redshift_idc_application` ([#37345](https://github.com/hashicorp/terraform-provider-aws/issues/37345))
* **New Resource:** `aws_secretsmanager_tag` ([#45825](https://github.com/hashicorp/terraform-provider-aws/issues/45825))
* **New Resource:** `aws_sesv2_tenant` ([#45706](https://github.com/hashicorp/terraform-provider-aws/issues/45706))

ENHANCEMENTS:

* data-source/aws_apigateway_domain_name : Add `endpoint_access_mode` attribute ([#45741](https://github.com/hashicorp/terraform-provider-aws/issues/45741))
* data-source/aws_db_proxy: Add `endpoint_network_type` and `target_connection_network_type` attributes ([#45634](https://github.com/hashicorp/terraform-provider-aws/issues/45634))
* data-source/aws_dx_gateway: Add `tags` attribute ([#45766](https://github.com/hashicorp/terraform-provider-aws/issues/45766))
* data-source/aws_ecr_lifecycle_policy_document: Add `rule.action.target_storage_class` and `rule.selection.storage_class` arguments, and new valid values for `rule.action.type` and `rule.selection.count_type` arguments ([#45752](https://github.com/hashicorp/terraform-provider-aws/issues/45752))
* data-source/aws_iam_saml_provider: Add `saml_provider_uuid` attribute ([#45707](https://github.com/hashicorp/terraform-provider-aws/issues/45707))
* data-source/aws_lambda_function: Add `response_streaming_invoke_arn` attribute ([#45652](https://github.com/hashicorp/terraform-provider-aws/issues/45652))
* data-source/aws_lambda_function: Support `code_signing_config_arn` in AWS GovCloud (US) Regions ([#45652](https://github.com/hashicorp/terraform-provider-aws/issues/45652))
* data-source/aws_route53_resolver_firewall_rules: Add `dns_threat_protection`, `confidence_threshold`, `firewall_threat_protection_id`, `firewall_domain_redirection_action`, and `q_type` attributes ([#45711](https://github.com/hashicorp/terraform-provider-aws/issues/45711))
* data-source/aws_route53_resolver_rule: Add `target_ips` attribute ([#45492](https://github.com/hashicorp/terraform-provider-aws/issues/45492))
* data-source/aws_vpc_endpoint: Add `dns_options.private_dns_preference` and `dns_options.private_dns_specified_domains` attributes ([#45679](https://github.com/hashicorp/terraform-provider-aws/issues/45679))
* data-source/aws_vpc_endpoint: Promote `service_region` and `vpc_endpoint_type` from attributes to arguments for filtering ([#45679](https://github.com/hashicorp/terraform-provider-aws/issues/45679))
* resource/aws_alb: Enforce tag policy compliance for the `elasticloadbalancing:loadbalancer` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_alb_listener: Enforce tag policy compliance for the `elasticloadbalancing:listener` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_alb_listener_rule: Enforce tag policy compliance for the `elasticloadbalancing:listener-rule` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_alb_target_group: Enforce tag policy compliance for the `elasticloadbalancing:targetgroup` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_apigateway_domain_name: Add `endpoint_access_mode` argument and configurable timeout for create and update ([#45741](https://github.com/hashicorp/terraform-provider-aws/issues/45741))
* resource/aws_athena_workgroup: Add `customer_content_encryption_configuration` argument ([#45744](https://github.com/hashicorp/terraform-provider-aws/issues/45744))
* resource/aws_athena_workgroup: Add `enable_minimum_encryption_configuration` argument ([#45744](https://github.com/hashicorp/terraform-provider-aws/issues/45744))
* resource/aws_athena_workgroup: Add `monitoring_configuration` argument ([#45744](https://github.com/hashicorp/terraform-provider-aws/issues/45744))
* resource/aws_cleanrooms_collaboration: Add resource identity support ([#45548](https://github.com/hashicorp/terraform-provider-aws/issues/45548))
* resource/aws_cloudfront_distribution: Add `connection_function_association` and `viewer_mtls_config` arguments ([#45847](https://github.com/hashicorp/terraform-provider-aws/issues/45847))
* resource/aws_cloudfront_distribution: Add `owner_account_id` argument to `vpc_origin_config` for cross-account VPC origin support ([#45011](https://github.com/hashicorp/terraform-provider-aws/issues/45011))
* resource/aws_cloudwatch_log_subscription_filter: Add `apply_on_transformed_logs` argument ([#45826](https://github.com/hashicorp/terraform-provider-aws/issues/45826))
* resource/aws_cloudwatch_log_subscription_filter: Add `emit_system_fields` argument ([#45760](https://github.com/hashicorp/terraform-provider-aws/issues/45760))
* resource/aws_db_proxy: Add `endpoint_network_type` and `target_connection_network_type` arguments ([#45634](https://github.com/hashicorp/terraform-provider-aws/issues/45634))
* resource/aws_docdb_cluster_instance: Enforce tag policy compliance for the `rds:db` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_docdb_global_cluster: Enforce tag policy compliance for the `rds:global-cluster` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_dx_gateway: Add `tags` argument and `tags_all` attribute. This functionality requires the `directconnect:TagResource` and `directconnect:UntagResource` IAM permissions ([#45766](https://github.com/hashicorp/terraform-provider-aws/issues/45766))
* resource/aws_ecr_repository_creation_template: Support `CREATE_ON_PUSH` as a valid value for `applied_for` ([#45720](https://github.com/hashicorp/terraform-provider-aws/issues/45720))
* resource/aws_ecs_capacity_provider: Add `managed_instances_provider.instance_launch_template.capacity_option_type` argument ([#45667](https://github.com/hashicorp/terraform-provider-aws/issues/45667))
* resource/aws_fsx_lustre_file_system: Enforce tag policy compliance for the `fsx:file-system` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_fsx_ontap_file_system: Enforce tag policy compliance for the `fsx:file-system` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_fsx_openzfs_file_system: Enforce tag policy compliance for the `fsx:file-system` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_fsx_openzfs_snapshot: Enforce tag policy compliance for the `fsx:snapshot` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_fsx_openzfs_volume: Enforce tag policy compliance for the `fsx:volume` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_fsx_windows_file_system: Enforce tag policy compliance for the `fsx:file-system` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_guardduty_filter: Add `finding_criteria.criterion.matches` and `finding_criteria.criterion.not_matches` arguments ([#45758](https://github.com/hashicorp/terraform-provider-aws/issues/45758))
* resource/aws_iam_policy: Add `delay_after_policy_creation_in_ms` argument. This functionality requires the `iam:SetDefaultPolicyVersion` IAM permission ([#42054](https://github.com/hashicorp/terraform-provider-aws/issues/42054))
* resource/aws_iam_saml_provider: Add `saml_provider_uuid` attribute ([#45707](https://github.com/hashicorp/terraform-provider-aws/issues/45707))
* resource/aws_iam_virtual_mfa_device: Add `serial_number` attribute ([#45751](https://github.com/hashicorp/terraform-provider-aws/issues/45751))
* resource/aws_imagebuilder_image: Add `logging_configuration` argument ([#45749](https://github.com/hashicorp/terraform-provider-aws/issues/45749))
* resource/aws_imagebuilder_image_pipeline: Add `logging_configuration` argument ([#45749](https://github.com/hashicorp/terraform-provider-aws/issues/45749))
* resource/aws_inspector_assessment_target: Add plan-time validation of `resource_group_arn` ([#45688](https://github.com/hashicorp/terraform-provider-aws/issues/45688))
* resource/aws_inspector_assessment_template: Add plan-time validation of `rules_package_arns` and `target_arn` ([#45688](https://github.com/hashicorp/terraform-provider-aws/issues/45688))
* resource/aws_lambda_event_source_mapping: Add `provisioned_poller_config.poller_group_name` argument ([#45313](https://github.com/hashicorp/terraform-provider-aws/issues/45313))
* resource/aws_lambda_event_source_mapping: Support Amazon MSK and self-managed Apache Kafka destinations (`kafka://topic-name`) for `destination_config.on_failure.destination_arn` argument ([#45802](https://github.com/hashicorp/terraform-provider-aws/issues/45802))
* resource/aws_lambda_function: Add `response_streaming_invoke_arn` attribute ([#45652](https://github.com/hashicorp/terraform-provider-aws/issues/45652))
* resource/aws_lambda_function: Support `code_signing_config_arn` in AWS GovCloud (US) Regions ([#45652](https://github.com/hashicorp/terraform-provider-aws/issues/45652))
* resource/aws_lambda_function_url: Automatically add the `lambda:InvokeFunction` permission, with the `InvokedViaFunctionUrl` flag set to `true`, to the function on creation when `authorization_type` is `NONE` ([#44858](https://github.com/hashicorp/terraform-provider-aws/issues/44858))
* resource/aws_lambda_permission: Add `invoked_via_function_url` argument ([#44858](https://github.com/hashicorp/terraform-provider-aws/issues/44858))
* resource/aws_lb_target_group_attachment: Add `quic_server_id` argument ([#45666](https://github.com/hashicorp/terraform-provider-aws/issues/45666))
* resource/aws_lb_target_group_attachment: Add plan-time validation of `target_group_arn` ([#45666](https://github.com/hashicorp/terraform-provider-aws/issues/45666))
* resource/aws_neptune_cluster: Enforce tag policy compliance for the `rds:cluster` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_neptune_cluster_instance: Enforce tag policy compliance for the `rds:db` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_neptune_global_cluster: Enforce tag policy compliance for the `rds:global-cluster` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_networkmanager_vpc_attachment: Enable in-place updates of `routing_policy_label` argument. This functionality requires the `networkmanager: PutAttachmentRoutingPolicyLabel` and `networkmanager: RemoveAttachmentRoutingPolicyLabel` IAM permissions ([#45728](https://github.com/hashicorp/terraform-provider-aws/issues/45728))
* resource/aws_osis_pipeline: Add `pipeline_role_arn` argument to support specifying a IAM role at the pipeline level ([#45806](https://github.com/hashicorp/terraform-provider-aws/issues/45806))
* resource/aws_rds_cluster: Enforce tag policy compliance for the `rds:cluster` tag type ([#45671](https://github.com/hashicorp/terraform-provider-aws/issues/45671))
* resource/aws_redshift_data_share_consumer_association: Add plan-time validation of `consumer_region` ([#45688](https://github.com/hashicorp/terraform-provider-aws/issues/45688))
* resource/aws_route53_resolver_firewall_rule: Add `dns_threat_protection`, `confidence_threshold`, and `firewall_threat_protection_id` arguments to support DNS Firewall Advanced rules ([#45711](https://github.com/hashicorp/terraform-provider-aws/issues/45711))
* resource/aws_transfer_web_app: Add `endpoint_details.vpc` configuration block to support VPC hosted Transfer Family web app ([#45745](https://github.com/hashicorp/terraform-provider-aws/issues/45745))
* resource/aws_vpc_endpoint: Add `dns_options.private_dns_preference` and `dns_options.private_dns_specified_domains` arguments ([#45679](https://github.com/hashicorp/terraform-provider-aws/issues/45679))
* resource/aws_vpclattice_service_network_resource_association: Add `private_dns_enabled` argument ([#45673](https://github.com/hashicorp/terraform-provider-aws/issues/45673))
* resource/aws_vpn_connection: Support in-place updates for `tunnel*_inside_cidr` and `tunnel*_inside_ipv6_cidr` arguments ([#45781](https://github.com/hashicorp/terraform-provider-aws/issues/45781))

BUG FIXES:

* data-source/aws_ecr_authorization_token: Fix value of `proxy_endpoint` when `registry_id` is specified ([#45754](https://github.com/hashicorp/terraform-provider-aws/issues/45754))
* data-source/aws_networkmanager_core_network_policy_document: Support `account-id`, not `account`, as a valid value for `attachment_policies.conditions.type`. This fixes a regression introduced in [v6.27.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6270-december-17-2025) ([#45788](https://github.com/hashicorp/terraform-provider-aws/issues/45788))
* data-source/aws_vpc_endpoint: Add missing implementation for `service_region` attribute ([#45679](https://github.com/hashicorp/terraform-provider-aws/issues/45679))
* provider: Fix handling of `user_agent` values where the product name contains a forward slash ([#45715](https://github.com/hashicorp/terraform-provider-aws/issues/45715))
* resource/aws_batch_job_definition: Fix crash during update when `node_properties` has `NodeRangeProperties.ecsProperties` set ([#45676](https://github.com/hashicorp/terraform-provider-aws/issues/45676))
* resource/aws_batch_job_definition: Fix handling of logically deleted results in List ([#45694](https://github.com/hashicorp/terraform-provider-aws/issues/45694))
* resource/aws_cloudwatch_log_subscription_filter: CloudWatch Logs: `PutSubscriptionFilter`: Retry `ValidationException: Make sure you have given CloudWatch Logs permission to assume the provided role` ([#43762](https://github.com/hashicorp/terraform-provider-aws/issues/43762))
* resource/aws_ec2_subnet_cidr_reservation: Fix 255 subnet CIDR reservation limit ([#45778](https://github.com/hashicorp/terraform-provider-aws/issues/45778))
* resource/aws_nat_gateway: Handle eventual consistency with attached appliances on delete ([#45842](https://github.com/hashicorp/terraform-provider-aws/issues/45842))
* resource/aws_vpc: Fix `reading EC2 VPC (...) default Security Group: empty result` and `reading EC2 VPC (...) main Route Table: empty result` errors when importing RAM-shared VPCs. This fixes a regression introduced in [v6.17.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6170-october-16-2025) ([#45780](https://github.com/hashicorp/terraform-provider-aws/issues/45780))
* resource/aws_vpc_endpoint: Fix "InvalidParameter: DnsOptions PrivateDnsOnlyForInboundResolverEndpoint is applicable only to Interface VPC Endpoints" error when creating S3 gateway VPC endpoint with IPv6 enabled ([#45849](https://github.com/hashicorp/terraform-provider-aws/issues/45849))
* resource/aws_vpc_endpoint: `private_dns_enabled` argument is now marked as `ForceNew` ([#45679](https://github.com/hashicorp/terraform-provider-aws/issues/45679))

## 6.27.0 (December 17, 2025)

FEATURES:

* **New Data Source:** `aws_organizations_account` ([#45543](https://github.com/hashicorp/terraform-provider-aws/issues/45543))
* **New Function:** `user_agent` ([#45464](https://github.com/hashicorp/terraform-provider-aws/issues/45464))
* **New List Resource:** `aws_kms_key` ([#45514](https://github.com/hashicorp/terraform-provider-aws/issues/45514))
* **New Resource:** `aws_cloudfront_trust_store` ([#45534](https://github.com/hashicorp/terraform-provider-aws/issues/45534))

ENHANCEMENTS:

* data-source/aws_datazone_domain: Add `root_domain_unit_id` attribute ([#44964](https://github.com/hashicorp/terraform-provider-aws/issues/44964))
* data-source/aws_networkmanager_core_network_policy_document: Add `routing_policies` and `attachment_routing_policy_rules` arguments ([#45246](https://github.com/hashicorp/terraform-provider-aws/issues/45246))
* data-source/aws_route53_resolver_endpoint: Add `rni_enhanced_metrics_enabled` attribute ([#45630](https://github.com/hashicorp/terraform-provider-aws/issues/45630))
* data-source/aws_route53_resolver_endpoint: Add `target_name_server_metrics_enabled` attribute ([#45630](https://github.com/hashicorp/terraform-provider-aws/issues/45630))
* provider: Add `user_agent` argument ([#45464](https://github.com/hashicorp/terraform-provider-aws/issues/45464))
* provider: The [`provider_meta` block](https://developer.hashicorp.com/terraform/internals/provider-meta) is now supported. The `user_agent` argument enables module authors to include additional product information in the `User-Agent` header sent during all AWS API requests made during Create, Read, Update, and Delete operations. ([#45464](https://github.com/hashicorp/terraform-provider-aws/issues/45464))
* resource/aws_bedrockagent_knowledge_base: Add `knowledge_base_configuration.kendra_knowledge_base_configuration` argument ([#44388](https://github.com/hashicorp/terraform-provider-aws/issues/44388))
* resource/aws_bedrockagent_knowledge_base: Add `knowledge_base_configuration.sql_knowledge_base_configuration` and `storage_configuration.neptune_analytics_configuration` arguments ([#45465](https://github.com/hashicorp/terraform-provider-aws/issues/45465))
* resource/aws_bedrockagent_knowledge_base: Add `storage_configuration.mongo_db_atlas_configuration` argument ([#37220](https://github.com/hashicorp/terraform-provider-aws/issues/37220))
* resource/aws_bedrockagent_knowledge_base: Add `storage_configuration.opensearch_managed_cluster_configuration` argument ([#44060](https://github.com/hashicorp/terraform-provider-aws/issues/44060))
* resource/aws_bedrockagent_knowledge_base: Add `storage_configuration.s3_vectors_configuration` block ([#45468](https://github.com/hashicorp/terraform-provider-aws/issues/45468))
* resource/aws_bedrockagent_knowledge_base: Make `knowledge_base_configuration.vector_knowledge_base_configuration` and ``storage_configuration` optional ([#44388](https://github.com/hashicorp/terraform-provider-aws/issues/44388))
* resource/aws_codebuild_project: Add `cache.cache_namespace` argument ([#45584](https://github.com/hashicorp/terraform-provider-aws/issues/45584))
* resource/aws_datazone_domain: Add `root_domain_unit_id` argument ([#44964](https://github.com/hashicorp/terraform-provider-aws/issues/44964))
* resource/aws_lambda_function: `code_sha256` is now optional and computed ([#45618](https://github.com/hashicorp/terraform-provider-aws/issues/45618))
* resource/aws_networkmanager_connect_attachment: Add `routing_policy_label` argument ([#45246](https://github.com/hashicorp/terraform-provider-aws/issues/45246))
* resource/aws_networkmanager_connect_peer: Support 4 byte ASNs in `bgp_options.peer_asn` ([#45246](https://github.com/hashicorp/terraform-provider-aws/issues/45246))
* resource/aws_networkmanager_connect_peer: Support 4 byte ASNs in `configuration.bgp_configurations.peer_asn` ([#45639](https://github.com/hashicorp/terraform-provider-aws/issues/45639))
* resource/aws_networkmanager_dx_gateway_attachment: Add `routing_policy_label` argument ([#45246](https://github.com/hashicorp/terraform-provider-aws/issues/45246))
* resource/aws_networkmanager_site_to_site_vpn_attachment: Add `routing_policy_label` argument ([#45246](https://github.com/hashicorp/terraform-provider-aws/issues/45246))
* resource/aws_networkmanager_transit_gateway_route_table_attachment: Add `routing_policy_label` argument ([#45246](https://github.com/hashicorp/terraform-provider-aws/issues/45246))
* resource/aws_networkmanager_vpc_attachment: Add `routing_policy_label` argument ([#45246](https://github.com/hashicorp/terraform-provider-aws/issues/45246))
* resource/aws_route53_resolver_endpoint: Add `rni_enhanced_metrics_enabled` argument ([#45630](https://github.com/hashicorp/terraform-provider-aws/issues/45630))
* resource/aws_route53_resolver_endpoint: Add `target_name_server_metrics_enabled` argument ([#45630](https://github.com/hashicorp/terraform-provider-aws/issues/45630))
* resource/aws_vpclattice_service_network_vpc_association: Add `private_dns_enabled` and `dns_options` arguments ([#45619](https://github.com/hashicorp/terraform-provider-aws/issues/45619))

BUG FIXES:

* data-source/aws_networkmanager_core_network_policy_document: Correct plan-time validation of `attachment_policies.conditions.type` to allow `account` instead of `account-id` ([#45246](https://github.com/hashicorp/terraform-provider-aws/issues/45246))
* resource/aws_bedrockagent_knowledge_base: Mark `knowledge_base_configuration.vector_knowledge_base_configuration.embedding_model_configuration` and `knowledge_base_configuration.vector_knowledge_base_configuration.supplemental_data_storage_configuration` as `ForceNew` ([#45465](https://github.com/hashicorp/terraform-provider-aws/issues/45465))
* resource/aws_dynamodb_table: Fix perpetual diff on `global_secondary_index` when using `ignore_changes` lifecycle meta-argument ([#41113](https://github.com/hashicorp/terraform-provider-aws/issues/41113))
* resource/aws_iam_user: Fix `NoSuchEntity` errors when `name` and `tags` arguments are both updated ([#45608](https://github.com/hashicorp/terraform-provider-aws/issues/45608))
* resource/aws_lakeformation_data_cells_filter: Fix `excluded_column_names` ordering causing "Provider produced inconsistent result after apply" errors ([#45453](https://github.com/hashicorp/terraform-provider-aws/issues/45453))
* resource/aws_neptune_global_cluster: Fix a regression in the minor version upgrade workflow triggered by upstream changes to the API error response text ([#45605](https://github.com/hashicorp/terraform-provider-aws/issues/45605))
* resource/aws_networkmanager_connect_peer: Change `bgp_options` and `bgp_options.peer_asn` to Optional, Computed and ForceNew ([#45639](https://github.com/hashicorp/terraform-provider-aws/issues/45639))
* resource/aws_odb_cloud_vm_cluster: Enable deletion of vm cluster in resource shared account. ([#45552](https://github.com/hashicorp/terraform-provider-aws/issues/45552))
* resource/aws_rds_global_cluster: Fix a regression in the minor version upgrade workflow triggered by upstream changes to the API error response text ([#45605](https://github.com/hashicorp/terraform-provider-aws/issues/45605))
* resource/aws_s3_bucket: Fix ``endpoint rule error, AccountId must only contain a-z, A-Z, 0-9 and `-`​`` errors when the provider is configured with [`skip_requesting_account_id = true`](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#skip_requesting_account_id-1). This fixes a regression introduced in [v6.23.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6230-november-26-2025) ([#45576](https://github.com/hashicorp/terraform-provider-aws/issues/45576))
* resource/aws_verifiedpermissions_identity_source: Fixes error when updating resource ([#45540](https://github.com/hashicorp/terraform-provider-aws/issues/45540))
* resource/aws_verifiedpermissions_identity_source: Prevents eventual consistency error with associated Policy Store ([#45540](https://github.com/hashicorp/terraform-provider-aws/issues/45540))
* resource/aws_verifiedpermissions_identity_source: Removes AutoFlex error log messages ([#45540](https://github.com/hashicorp/terraform-provider-aws/issues/45540))

## 6.26.0 (December 10, 2025)

FEATURES:

* **New List Resource:** `aws_batch_job_definition` ([#45401](https://github.com/hashicorp/terraform-provider-aws/issues/45401))
* **New List Resource:** `aws_codebuild_project` ([#45400](https://github.com/hashicorp/terraform-provider-aws/issues/45400))
* **New List Resource:** `aws_lambda_capacity_provider` ([#45467](https://github.com/hashicorp/terraform-provider-aws/issues/45467))
* **New List Resource:** `aws_ssm_parameter` ([#45512](https://github.com/hashicorp/terraform-provider-aws/issues/45512))
* **New Resource:** `aws_iam_outbound_web_identity_federation` ([#45217](https://github.com/hashicorp/terraform-provider-aws/issues/45217))

ENHANCEMENTS:

* data-source/aws_db_instance: Add `upgrade_rollout_order` attribute ([#45527](https://github.com/hashicorp/terraform-provider-aws/issues/45527))
* data-source/aws_eks_node_group : Add `update_config` block including `update_strategy` attribute ([#41487](https://github.com/hashicorp/terraform-provider-aws/issues/41487))
* data-source/aws_rds_cluster: Add `upgrade_rollout_order` attribute ([#45527](https://github.com/hashicorp/terraform-provider-aws/issues/45527))
* resource/aws_bedrockagent_agent: Add `session_summary_configuration.max_recent_sessions` argument ([#45449](https://github.com/hashicorp/terraform-provider-aws/issues/45449))
* resource/aws_db_instance: Add `upgrade_rollout_order` attribute ([#45527](https://github.com/hashicorp/terraform-provider-aws/issues/45527))
* resource/aws_eks_node_group : Add `update_config.update_strategy` attribute ([#41487](https://github.com/hashicorp/terraform-provider-aws/issues/41487))
* resource/aws_kinesisanalyticsv2_application: Add `application_configuration.application_encryption_configuration` argument ([#45356](https://github.com/hashicorp/terraform-provider-aws/issues/45356))
* resource/aws_kinesisanalyticsv2_application: Support `FLINK-1_20` as a valid value for `runtime_environment` ([#45356](https://github.com/hashicorp/terraform-provider-aws/issues/45356))
* resource/aws_lambda_capacity_provider: Add resource identity support ([#45456](https://github.com/hashicorp/terraform-provider-aws/issues/45456))
* resource/aws_odb_network_peering_connection: Add network peering creation using `odb_network_arn` for resource sharing model. ([#45509](https://github.com/hashicorp/terraform-provider-aws/issues/45509))
* resource/aws_rds_cluster: Add `upgrade_rollout_order` attribute ([#45527](https://github.com/hashicorp/terraform-provider-aws/issues/45527))
* resource/aws_s3vectors_index: Add `encryption_configuration` block ([#45470](https://github.com/hashicorp/terraform-provider-aws/issues/45470))
* resource/aws_s3vectors_index: Add `metadata_configuration` block ([#45470](https://github.com/hashicorp/terraform-provider-aws/issues/45470))

BUG FIXES:

* data-source/aws_ec2_transit_gateway: Fix potential crash when reading `encryption_support`. This addresses a regression introduced in [v6.25.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6250-december-4-2025). ([#45462](https://github.com/hashicorp/terraform-provider-aws/issues/45462))
* resource/aws_api_gateway_integration: Fix `timeout_milliseconds` validation to allow up to 900,000 ms when `response_transfer_mode` is `STREAM` ([#45482](https://github.com/hashicorp/terraform-provider-aws/issues/45482))
* resource/aws_bedrock_model_invocation_logging_configuration: Mark `logging_config.s3_config.bucket_name`, `logging_config.cloudwatch_config.log_group_name`, `logging_config.cloudwatch_config.role_arn`, and `logging_config.cloudwatch_config.large_data_delivery_s3_config.bucket_name` as Required ([#45469](https://github.com/hashicorp/terraform-provider-aws/issues/45469))
* resource/aws_ec2_transit_gateway: Fix potential crash when setting `encryption_support`. This addresses a regression introduced in [v6.25.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6250-december-4-2025). ([#45462](https://github.com/hashicorp/terraform-provider-aws/issues/45462))
* resource/aws_lambda_function: Fix persistent diff when `image_config` has `null` values set in config ([#45511](https://github.com/hashicorp/terraform-provider-aws/issues/45511))
* resource/aws_notifications_event_rule: Fix persistent diff when `event_pattern` argument is not specified in config ([#45524](https://github.com/hashicorp/terraform-provider-aws/issues/45524))
* resource/aws_route53_zone: Operations to enable accelerated recovery are enforced to run serially when multiple hosted zones are configured ([#45457](https://github.com/hashicorp/terraform-provider-aws/issues/45457))
* resource/aws_sagemaker_model: Mark `vpc_config.security_group_ids` and `vpc_config.subnets` as `ForceNew` ([#45491](https://github.com/hashicorp/terraform-provider-aws/issues/45491))
* resource/aws_secretsmanager_secret_version: Avoid sending GetSecretValue calls when the secret is write-only ([#44876](https://github.com/hashicorp/terraform-provider-aws/issues/44876))

## 6.25.0 (December 4, 2025)

FEATURES:

* **New Resource:** `aws_cloudwatch_log_transformer` ([#44300](https://github.com/hashicorp/terraform-provider-aws/issues/44300))
* **New Resource:** `aws_eks_capability` ([#45326](https://github.com/hashicorp/terraform-provider-aws/issues/45326))

ENHANCEMENTS:

* data-source/aws_backup_plan: Add `rule.scan_action` and `scan_setting` attributes ([#45392](https://github.com/hashicorp/terraform-provider-aws/issues/45392))
* data-source/aws_cloudwatch_log_group: Add `deletion_protection_enabled` attribute ([#45298](https://github.com/hashicorp/terraform-provider-aws/issues/45298))
* data-source/aws_ec2_transit_gateway: Add `encryption_support` attribute ([#45317](https://github.com/hashicorp/terraform-provider-aws/issues/45317))
* data-source/aws_lambda_function: Add `durable_config` attribute ([#45359](https://github.com/hashicorp/terraform-provider-aws/issues/45359))
* data-source/aws_lb: Add `health_check_logs` attribute ([#45269](https://github.com/hashicorp/terraform-provider-aws/issues/45269))
* data-source/aws_lb_target_group: Add `target_control_port` attribute ([#45270](https://github.com/hashicorp/terraform-provider-aws/issues/45270))
* data-source/aws_route53_zone: Add `enable_accelerated_recovery` attribute ([#45302](https://github.com/hashicorp/terraform-provider-aws/issues/45302))
* data-source/aws_transfer_connector: Add `egress_config` attribute to expose VPC Lattice connectivity configuration ([#45314](https://github.com/hashicorp/terraform-provider-aws/issues/45314))
* data-source/aws_workspaces_directory: Add `tenancy` attribute ([#43134](https://github.com/hashicorp/terraform-provider-aws/issues/43134))
* resource/aws_api_gateway_integration: Add `integration_target` argument ([#45311](https://github.com/hashicorp/terraform-provider-aws/issues/45311))
* resource/aws_api_gateway_integration: Add `response_transfer_mode` argument ([#45329](https://github.com/hashicorp/terraform-provider-aws/issues/45329))
* resource/aws_athena_workgroup: Add `configuration.managed_query_results_configuration` block ([#44273](https://github.com/hashicorp/terraform-provider-aws/issues/44273))
* resource/aws_backup_plan: Support malware scanning by adding `rule.scan_action` and `scan_setting` configuration blocks ([#45392](https://github.com/hashicorp/terraform-provider-aws/issues/45392))
* resource/aws_bedrockagentcore_gateway: Add `interceptor_configuration` argument ([#45344](https://github.com/hashicorp/terraform-provider-aws/issues/45344))
* resource/aws_cloudwatch_log_group: Add `deletion_protection_enabled` argument ([#45298](https://github.com/hashicorp/terraform-provider-aws/issues/45298))
* resource/aws_ec2_transit_gateway: Add `encryption_support` argument ([#45317](https://github.com/hashicorp/terraform-provider-aws/issues/45317))
* resource/aws_flow_log: Add `regional_nat_gateway_id` argument ([#45380](https://github.com/hashicorp/terraform-provider-aws/issues/45380))
* resource/aws_kms_ciphertext: Add `plaintext_wo` and `plaintext_wo_version` arguments to support write-only input ([#43592](https://github.com/hashicorp/terraform-provider-aws/issues/43592))
* resource/aws_lambda_function: Add `durable_config` argument ([#45359](https://github.com/hashicorp/terraform-provider-aws/issues/45359))
* resource/aws_lb: Add `health_check_logs` configuration block ([#45269](https://github.com/hashicorp/terraform-provider-aws/issues/45269))
* resource/aws_lb_target_group: Add `target_control_port` argument to support the ALB Target Optimizer ([#45270](https://github.com/hashicorp/terraform-provider-aws/issues/45270))
* resource/aws_rolesanywhere_profile: Add `accept_role_session_name` argument ([#45391](https://github.com/hashicorp/terraform-provider-aws/issues/45391))
* resource/aws_rolesanywhere_profile: Add plan-time validation of `managed_policy_arns` and `role_arns` ([#45391](https://github.com/hashicorp/terraform-provider-aws/issues/45391))
* resource/aws_route53_zone: Add `enable_accelerated_recovery` argument ([#45302](https://github.com/hashicorp/terraform-provider-aws/issues/45302))
* resource/aws_ssm_association: Add `calendar_names` argument ([#45363](https://github.com/hashicorp/terraform-provider-aws/issues/45363))
* resource/aws_transfer_connector: Add `egress_config` argument to support VPC Lattice connectivity for SFTP connectors ([#45314](https://github.com/hashicorp/terraform-provider-aws/issues/45314))
* resource/aws_transfer_connector: Make `url` argument optional to support VPC Lattice connectors ([#45314](https://github.com/hashicorp/terraform-provider-aws/issues/45314))
* resource/aws_workspaces_directory: Add `tenancy` argument ([#43134](https://github.com/hashicorp/terraform-provider-aws/issues/43134))

## 6.24.0 (December 2, 2025)

FEATURES:

* **New Resource:** `aws_lambda_capacity_provider` ([#45342](https://github.com/hashicorp/terraform-provider-aws/issues/45342))
* **New Resource:** `aws_s3tables_table_bucket_replication` ([#45360](https://github.com/hashicorp/terraform-provider-aws/issues/45360))
* **New Resource:** `aws_s3tables_table_replication` ([#45360](https://github.com/hashicorp/terraform-provider-aws/issues/45360))
* **New Resource:** `aws_s3vectors_index` ([#43393](https://github.com/hashicorp/terraform-provider-aws/issues/43393))
* **New Resource:** `aws_s3vectors_vector_bucket` ([#43393](https://github.com/hashicorp/terraform-provider-aws/issues/43393))
* **New Resource:** `aws_s3vectors_vector_bucket_policy` ([#43393](https://github.com/hashicorp/terraform-provider-aws/issues/43393))

ENHANCEMENTS:

* data-source/aws_lambda_function: Add `capacity_provider_config` attribute ([#45342](https://github.com/hashicorp/terraform-provider-aws/issues/45342))
* data-source/aws_vpc_nat_gateway: Support regional NAT Gateways by adding `auto_provision_zones`, `auto_scaling_ips`, `availability_mode`, `availability_zone_address`, `regional_nat_gateway_address`, and `route_table_id` attributes ([#45240](https://github.com/hashicorp/terraform-provider-aws/issues/45240))
* resource/aws_backup_plan: Add `target_logically_air_gapped_backup_vault_arn` argument to `rule` block ([#45321](https://github.com/hashicorp/terraform-provider-aws/issues/45321))
* resource/aws_lambda_function: Add `capacity_provider_config` and `publish_to` arguments ([#45342](https://github.com/hashicorp/terraform-provider-aws/issues/45342))
* resource/aws_resourceexplorer2_index: Deprecates `id`. Use `arn` instead. ([#45345](https://github.com/hashicorp/terraform-provider-aws/issues/45345))
* resource/aws_resourceexplorer2_view: Deprecates `id`. Use `arn` instead. ([#45345](https://github.com/hashicorp/terraform-provider-aws/issues/45345))
* resource/aws_vpc_nat_gateway: Make `subnet_id` argument optional to support regional NAT Gateways ([#45420](https://github.com/hashicorp/terraform-provider-aws/issues/45420))
* resource/aws_vpc_nat_gateway: Support regional NAT Gateways by adding `availability_mode`, `availability_zone_address`, and `vpc_id` arguments, and `auto_provision_zones`, `auto_scaling_ips`, `regional_nat_gateway_address`, and `route_table_id` attributes. This functionality requires the `ec2:DescribeAvailabilityZones` IAM permission ([#45240](https://github.com/hashicorp/terraform-provider-aws/issues/45240))
* resource/aws_vpn_connection: Add `bgp_log_enabled`, `bgp_log_group_arn`, and `bgp_log_stream_arn` arguments to `tunnel1_log_options.cloudwatch_log_options` and `tunnel2_log_options.cloudwatch_log_options` blocks ([#45271](https://github.com/hashicorp/terraform-provider-aws/issues/45271))

## 6.23.0 (November 26, 2025)

NOTES:

* resource/aws_s3_bucket: To support ABAC (Attribute Based Access Control) in general purpose buckets, this resource will now attempt to send tags in the create request and use the S3 Control tagging APIs [`TagResource`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_control_TagResource.html), [`UntagResource`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_control_UntagResource.html), and [`ListTagsForResource`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_control_ListTagsForResource.html) for read and update operations. The calling principal must have the corresponding `s3:TagResource`, `s3:UntagResource`, and `s3:ListTagsForResource` [IAM permissions](https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazons3.html#amazons3-actions-as-permissions). If the principal lacks the appropriate permissions, the provider will fall back to tagging after creation and using the S3 tagging APIs [`PutBucketTagging`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketTagging.html), [`DeleteBucketTagging`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketTagging.html), and [`GetBucketTagging`](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketTagging.html) instead. With ABAC enabled, tag modifications may fail with the fall back behavior. See the [AWS documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/buckets-tagging-enable-abac.html) for additional details on enabling ABAC in general purpose buckets. ([#45251](https://github.com/hashicorp/terraform-provider-aws/issues/45251))

FEATURES:

* **New Resource:** `aws_ecs_express_gateway_service` ([#45235](https://github.com/hashicorp/terraform-provider-aws/issues/45235))
* **New Resource:** `aws_s3_bucket_abac` ([#45251](https://github.com/hashicorp/terraform-provider-aws/issues/45251))
* **New Resource:** `aws_vpc_encryption_control` ([#45263](https://github.com/hashicorp/terraform-provider-aws/issues/45263))
* **New Resource:** `aws_vpn_concentrator` ([#45175](https://github.com/hashicorp/terraform-provider-aws/issues/45175))

ENHANCEMENTS:

* action/aws_lambda_invoke: Add `tenant_id` argument ([#45170](https://github.com/hashicorp/terraform-provider-aws/issues/45170))
* data-source/aws_eks_cluster: Add `control_plane_scaling_config` attribute ([#45258](https://github.com/hashicorp/terraform-provider-aws/issues/45258))
* data-source/aws_lambda_function: Add `tenancy_config` attribute ([#45170](https://github.com/hashicorp/terraform-provider-aws/issues/45170))
* data-source/aws_lambda_invocation: Add `tenant_id` argument ([#45170](https://github.com/hashicorp/terraform-provider-aws/issues/45170))
* data-source/aws_vpn_connection: Add `vpn_concentrator_id` attribute ([#45175](https://github.com/hashicorp/terraform-provider-aws/issues/45175))
* resoource/aws_ecs_capacity_provider: Add `managed_instances_provider.infrastructure_optimization` argument ([#45142](https://github.com/hashicorp/terraform-provider-aws/issues/45142))
* resource/aws_docdb_cluster: Add `network_type` argument ([#45140](https://github.com/hashicorp/terraform-provider-aws/issues/45140))
* resource/aws_docdb_subnet_group: Add `supported_network_types` attribute ([#45140](https://github.com/hashicorp/terraform-provider-aws/issues/45140))
* resource/aws_eks_cluster: Add `control_plane_scaling_config` configuration block to support EKS Provisioned Control Plane ([#45258](https://github.com/hashicorp/terraform-provider-aws/issues/45258))
* resource/aws_lambda_function: Add `tenancy_config` argument ([#45170](https://github.com/hashicorp/terraform-provider-aws/issues/45170))
* resource/aws_lambda_invocation: Add `tenant_id` argument ([#45170](https://github.com/hashicorp/terraform-provider-aws/issues/45170))
* resource/aws_s3_bucket: Tag on creation when the `s3:TagResource` permission is present ([#45251](https://github.com/hashicorp/terraform-provider-aws/issues/45251))
* resource/aws_s3_bucket: Use the S3 Control tagging APIs when the `s3:TagResource`, `s3:UntagResource`, and `s3:ListTagsForResource` permissions are present ([#45251](https://github.com/hashicorp/terraform-provider-aws/issues/45251))
* resource/aws_vpn_connection: Add `vpn_concentrator_id` argument to support Site-to-Site VPN Concentrator ([#45175](https://github.com/hashicorp/terraform-provider-aws/issues/45175))

## 6.22.1 (November 21, 2025)

ENHANCEMENTS:

* resource/aws_fsx_openzfs_file_system: Support `INTELLIGENT_TIERING` storage type and add `read_cache_configuration` argument ([#45159](https://github.com/hashicorp/terraform-provider-aws/issues/45159))
* resource/aws_msk_cluster: Add `rebalancing` configuration block to support intelligent rebalancing for Express broker clusters ([#45073](https://github.com/hashicorp/terraform-provider-aws/issues/45073))

BUG FIXES:

* provider: Fix crash in required tag validation interceptor when tag values are unknown. This addresses a regression introduced in [v6.22.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6220-november-20-2025). ([#45201](https://github.com/hashicorp/terraform-provider-aws/issues/45201))
* provider: Fix early return logic in the required tag validation interceptor. This addresses a performance regression introduced in [v6.22.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6220-november-20-2025). ([#45201](https://github.com/hashicorp/terraform-provider-aws/issues/45201))
* resource/aws_accessanalyzer_analyzer: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panics when `configuration.unused_access.analysis_rule.exclusion.resource_tags` contains `null` values ([#45202](https://github.com/hashicorp/terraform-provider-aws/issues/45202))
* resource/aws_odb_cloud_vm_cluster: Fix incorrect validation error when arguments are configured using variables. This addresses a regression introduced in [v6.22.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6220-november-20-2025) ([#45205](https://github.com/hashicorp/terraform-provider-aws/issues/45205))

## 6.22.0 (November 20, 2025)

NOTES:

* resource/aws_s3_bucket_server_side_encryption_configuration: Starting in March 2026, Amazon S3 will introduce a new default bucket security setting by automatically disabling server-side encryption with customer-provided keys (SSE-C) for all new buckets. Use the `blocked_encryption_types` argument to manage this behavior for specific buckets. ([#45105](https://github.com/hashicorp/terraform-provider-aws/issues/45105))

FEATURES:

* **New Ephemeral Resource:** `aws_ecr_authorization_token` ([#44949](https://github.com/hashicorp/terraform-provider-aws/issues/44949))
* **New Guide:** `Tag Policy Compliance` ([#45143](https://github.com/hashicorp/terraform-provider-aws/issues/45143))
* **New Resource:** `aws_billing_view` ([#45097](https://github.com/hashicorp/terraform-provider-aws/issues/45097))
* **New Resource:** `aws_vpclattice_domain_verification` ([#45085](https://github.com/hashicorp/terraform-provider-aws/issues/45085))

ENHANCEMENTS:

* data-source/aws_lb_listener: Add `default_action.jwt_validation` attribute ([#45089](https://github.com/hashicorp/terraform-provider-aws/issues/45089))
* data-source/aws_lb_listener_rule: Add `action.jwt_validation` attribute ([#45089](https://github.com/hashicorp/terraform-provider-aws/issues/45089))
* data-source/aws_route53_zone: Support filtering by `tags` only or by `vpc_id` only ([#39671](https://github.com/hashicorp/terraform-provider-aws/issues/39671))
* provider: Add support for enforcing tag policy compliance. This opt-in feature can be enabled via the new `tag_policy_compliance` provider argument, or the `TF_AWS_TAG_POLICY_COMPLIANCE` environment variable. When enabled, the principal executing Terraform must have the `tags:ListRequiredTags` IAM permission. ([#45143](https://github.com/hashicorp/terraform-provider-aws/issues/45143))
* resource/aws_backup_logically_air_gapped_vault: Add `encryption_key_arn` argument ([#45020](https://github.com/hashicorp/terraform-provider-aws/issues/45020))
* resource/aws_bedrock_guardrail: Add `input_action`, `input_enabled`, `input_modalities`, `output_action`, `output_enabled`, and `output_modalities` arguments to the `content_policy_config.filters_config` block ([#45104](https://github.com/hashicorp/terraform-provider-aws/issues/45104))
* resource/aws_bedrockagent_knowledge_base: Add `storage_configuration.rds_configuration.field_mapping.custom_metadata_field` argument ([#45075](https://github.com/hashicorp/terraform-provider-aws/issues/45075))
* resource/aws_bedrockagentcore_agent_runtime: Add `agent_runtime_artifact.code_configuration` block ([#45091](https://github.com/hashicorp/terraform-provider-aws/issues/45091))
* resource/aws_bedrockagentcore_agent_runtime: Make `agent_runtime_artifact.container_configuration` block optional ([#45091](https://github.com/hashicorp/terraform-provider-aws/issues/45091))
* resource/aws_dynamodb_table: Add `global_table_witness` argument ([#43908](https://github.com/hashicorp/terraform-provider-aws/issues/43908))
* resource/aws_emr_managed_scaling_policy: Add `scaling_strategy` and `utilization_performance_index` arguments ([#45132](https://github.com/hashicorp/terraform-provider-aws/issues/45132))
* resource/aws_fis_experiment_template: Add plan-time validation of `log_configuration.cloudwatch_logs_configuration.log_group_arn` ([#35941](https://github.com/hashicorp/terraform-provider-aws/issues/35941))
* resource/aws_fis_experiment_template: Add support for `Functions` to `action.*.target` ([#41209](https://github.com/hashicorp/terraform-provider-aws/issues/41209))
* resource/aws_lambda_invocation: Add import support ([#41240](https://github.com/hashicorp/terraform-provider-aws/issues/41240))
* resource/aws_lb_listener: Support `jwt-validation` as a valid `default_action.type` and add `default_action.jwt_validation` configuration block ([#45089](https://github.com/hashicorp/terraform-provider-aws/issues/45089))
* resource/aws_lb_listener_rule: Support `jwt-validation` as a valid `action.type` and add `action.jwt_validation` configuration block ([#45089](https://github.com/hashicorp/terraform-provider-aws/issues/45089))
* resource/aws_odb_cloud_vm_cluster: vm cluster creation using odb network ARN and exadata infrastructure ARN for resource sharing model. ([#45003](https://github.com/hashicorp/terraform-provider-aws/issues/45003))
* resource/aws_organizations_organization: Add `SECURITYHUB_POLICY` as a valid value for `enabled_policy_types` argument ([#45135](https://github.com/hashicorp/terraform-provider-aws/issues/45135))
* resource/aws_prometheus_query_logging_configuration: Add plan-time validation of `destination.cloudwatch_logs.log_group_arn` ([#35941](https://github.com/hashicorp/terraform-provider-aws/issues/35941))
* resource/aws_prometheus_workspace: Add plan-time validation of `logging_configuration.log_group_arn` ([#35941](https://github.com/hashicorp/terraform-provider-aws/issues/35941))
* resource/aws_s3_bucket_server_side_encryption_configuration: Add `rule.blocked_encryption_types` argument ([#45105](https://github.com/hashicorp/terraform-provider-aws/issues/45105))
* resource/aws_sagemaker_model: Add `container.additional_model_data_source` and `primary_container.additional_model_data_source` arguments ([#44407](https://github.com/hashicorp/terraform-provider-aws/issues/44407))
* resource/aws_sfn_state_machine: Add plan-time validation of `logging_configuration.log_destination` ([#35941](https://github.com/hashicorp/terraform-provider-aws/issues/35941))
* resource/aws_timestreaminfluxdb_db_cluster: Add `engine_type` attribute ([#44899](https://github.com/hashicorp/terraform-provider-aws/issues/44899))
* resource/aws_timestreaminfluxdb_db_cluster: Add validation to ensure InfluxDB V2 clusters have required fields and InfluxDB V3 clusters (when using V3 parameter groups) do not have forbidden V2 fields. This functionality requires the `timestream-influxdb:GetDbParameterGroup` IAM permission ([#44899](https://github.com/hashicorp/terraform-provider-aws/issues/44899))
* resource/aws_vpclattice_resource_configuration: Add `custom_domain_name` and `domain_verification_id` arguments and `domain_verification_arn` and `domain_verification_status` attributes to support custom domain names for resource configurations ([#45085](https://github.com/hashicorp/terraform-provider-aws/issues/45085))
* resource/aws_vpn_connection: Add `tunnel_bandwidth` argument to support higher bandwidth tunnels ([#45070](https://github.com/hashicorp/terraform-provider-aws/issues/45070))

BUG FIXES:

* resource/aws_db_instance: Fix blue/green deployments failing with "not in available state" by improving stability and handling `storage-config-upgrade` and `storage-initialization` statuses ([#41275](https://github.com/hashicorp/terraform-provider-aws/issues/41275))
* resource/aws_elastic_beanstalk_configuration_template: Fix updates not applying by including `ResourceName` for option settings and preventing duplicate add/remove operations ([#45077](https://github.com/hashicorp/terraform-provider-aws/issues/45077))
* resource/aws_odb_cloud_vm_cluster: support for hyphen in odb cloud vm cluster hostname prefix. ([#45003](https://github.com/hashicorp/terraform-provider-aws/issues/45003))
* resource/aws_quicksight_account_settings: Add `region` argument ([#45083](https://github.com/hashicorp/terraform-provider-aws/issues/45083))
* resource/aws_s3_directory_bucket: Fix plan-time `AWS resource not found during refresh` warnings causing resource replacement when `ReadOnly` `s3express:SessionMode` is enforced ([#45086](https://github.com/hashicorp/terraform-provider-aws/issues/45086))
* resource/aws_ssoadmin_account_assignment: Correct `target_type` argument to required ([#45092](https://github.com/hashicorp/terraform-provider-aws/issues/45092))
* resource/aws_timestreaminfluxdb_db_cluster: Make `allocated_storage`, `bucket`, `organization`, `username`, and `password` optional to support InfluxDB V3 clusters ([#44899](https://github.com/hashicorp/terraform-provider-aws/issues/44899))

## 6.21.0 (November 13, 2025)

BREAKING CHANGES:

* resource/aws_bedrockagentcore_browser: Rename `network_configuration.network_mode_config` to `network_configuration.vpc_config` ([#44828](https://github.com/hashicorp/terraform-provider-aws/issues/44828))

FEATURES:

* **New Action:** `aws_dynamodb_create_backup` ([#45001](https://github.com/hashicorp/terraform-provider-aws/issues/45001))
* **New Resource:** `aws_networkflowmonitor_monitor` ([#44782](https://github.com/hashicorp/terraform-provider-aws/issues/44782))
* **New Resource:** `aws_networkflowmonitor_scope` ([#44782](https://github.com/hashicorp/terraform-provider-aws/issues/44782))
* **New Resource:** `aws_observabilityadmin_centralization_rule_for_organization` ([#44806](https://github.com/hashicorp/terraform-provider-aws/issues/44806))

ENHANCEMENTS:

* data-source/aws_ecs_service: Add `capacity_provider_strategy`, `created_at`, `created_by`, `deployment_configuration`, `deployment_controller`, `deployments`, `enable_ecs_managed_tags`, `enable_execute_command`, `events`, `health_check_grace_period_seconds`, `iam_role`, `network_configuration`, `ordered_placement_strategy`, `pending_count`, `placement_constraints`, `platform_family`, `platform_version`, `propagate_tags`, `running_count`, `service_connect_configuration`, `service_registries`, `status`, and `task_sets` attributes ([#44842](https://github.com/hashicorp/terraform-provider-aws/issues/44842))
* resource/aws_bedrockagentcore_gateway_target: Add `target_configuration.mcp.mcp_server` block ([#44991](https://github.com/hashicorp/terraform-provider-aws/issues/44991))
* resource/aws_bedrockagentcore_gateway_target: Make `credential_provider_configuration` block optional ([#44991](https://github.com/hashicorp/terraform-provider-aws/issues/44991))
* resource/aws_cloudwatch_log_delivery_destination: Make `delivery_destination_type` and `delivery_destination_configuration` optional to support AWS X-Ray as a destination ([#44995](https://github.com/hashicorp/terraform-provider-aws/issues/44995))
* resource/aws_ecs_service: Add support for `LINEAR` and `CANARY` deployment strategies with `deployment_configuration.linear_configuration` and `deployment_configuration.canary_configuration` blocks ([#44842](https://github.com/hashicorp/terraform-provider-aws/issues/44842))
* resource/aws_lambda_function: Add support for `java25` `runtime` value ([#45024](https://github.com/hashicorp/terraform-provider-aws/issues/45024))
* resource/aws_lambda_function: Add support for `nodejs24.x` `runtime` value ([#45024](https://github.com/hashicorp/terraform-provider-aws/issues/45024))
* resource/aws_lambda_function: Add support for `python3.14` `runtime` value ([#45024](https://github.com/hashicorp/terraform-provider-aws/issues/45024))
* resource/aws_lambda_layer_version: Add support for `java25` `compatible_runtimes` value ([#45024](https://github.com/hashicorp/terraform-provider-aws/issues/45024))
* resource/aws_lambda_layer_version: Add support for `nodejs24.x` `compatible_runtimes` value ([#45024](https://github.com/hashicorp/terraform-provider-aws/issues/45024))
* resource/aws_lambda_layer_version: Add support for `python3.14` `compatible_runtimes` value ([#45024](https://github.com/hashicorp/terraform-provider-aws/issues/45024))
* resource/aws_s3tables_table: Add tagging support ([#44996](https://github.com/hashicorp/terraform-provider-aws/issues/44996))
* resource/aws_s3tables_table_bucket: Add tagging support ([#44996](https://github.com/hashicorp/terraform-provider-aws/issues/44996))
* resource/aws_sagemaker_endpoint_configuration: Add `execution_role_arn` argument and make `model_name` optional in `production_variants` and `shadow_production_variants` blocks to support Inference Components ([#44977](https://github.com/hashicorp/terraform-provider-aws/issues/44977))
* resource/aws_sns_topic: Fix `AuthorizationError ... is not authorized to perform: iam:PassRole on resource ...` IAM eventual consistency errors on Create and Update ([#45018](https://github.com/hashicorp/terraform-provider-aws/issues/45018))

BUG FIXES:

* provider: Fix situation where refreshes of removed infrastructure appear as errors rather than warnings ([#45022](https://github.com/hashicorp/terraform-provider-aws/issues/45022))
* resource/aws_acmpca_certificate_authority: Prevents error when upgrading from provider pre-v6.0 without refreshing ([#45050](https://github.com/hashicorp/terraform-provider-aws/issues/45050))
* resource/aws_apprunner_service: Prevents error when upgrading from provider pre-v6.0 without refreshing ([#45051](https://github.com/hashicorp/terraform-provider-aws/issues/45051))
* resource/aws_ec2_image_block_public_access: Add `region` argument ([#45023](https://github.com/hashicorp/terraform-provider-aws/issues/45023))
* resource/aws_ec2_serial_console_access: Add `region` argument ([#45064](https://github.com/hashicorp/terraform-provider-aws/issues/45064))
* resource/aws_emrcontainers_job_template: Fix `ValidationException: Value null at 'jobTemplateData.configurationOverrides.monitoringConfiguration.cloudWatchMonitoringConfiguration.logGroupName' failed to satisfy constraint: Member must not be null` error ([#45029](https://github.com/hashicorp/terraform-provider-aws/issues/45029))
* resource/aws_emrcontainers_job_template: Fix `setting job_template_data: job_template_data.0.configuration_overrides.0.application_configuration.0: '' expected a map, got 'slice'` error ([#45029](https://github.com/hashicorp/terraform-provider-aws/issues/45029))
* resource/aws_emrcontainers_job_template: Mark `job_template_data.job_driver.configuration_overrides.monitoring_configuration.persistent_app_ui` argument as computed ([#45029](https://github.com/hashicorp/terraform-provider-aws/issues/45029))
* resource/aws_invoicing_invoice_unit: Fix `Provider returned invalid result object after apply` error occurred when updating the resource ([#45030](https://github.com/hashicorp/terraform-provider-aws/issues/45030))
* resource/aws_opensearch_authorize_vpc_endpoint_access: Fix reading the resource when more than one principal is authorized. The [import ID](https://developer.hashicorp.com/terraform/language/block/import#id) has changed from `domain_name` to `domain_name` and `account` separated by a comma ([#44982](https://github.com/hashicorp/terraform-provider-aws/issues/44982))
* resource/aws_redshift_cluster: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_cluster_snapshot: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_event_subscription: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_hsm_client_certificate: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_hsm_configuration: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_integration: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_parameter_group: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_snapshot_copy_grant: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_snapshot_schedule: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_subnet_group: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_usage_limit: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_sagemaker_endpoint: Fix bug where `endpoint_config_name` was not correctly updated, causing the endpoint to retain the old configuration ([#42843](https://github.com/hashicorp/terraform-provider-aws/issues/42843))
* resource/aws_wafv2_web_acl_logging_configuration: Fix the validation for `redacted_fields.single_header.name` ([#44987](https://github.com/hashicorp/terraform-provider-aws/issues/44987))

## 6.20.0 (November 6, 2025)

FEATURES:

* **New Resource:** `aws_ec2_allowed_images_settings` ([#44800](https://github.com/hashicorp/terraform-provider-aws/issues/44800))
* **New Resource:** `aws_fis_target_account_configuration` ([#44875](https://github.com/hashicorp/terraform-provider-aws/issues/44875))
* **New Resource:** `aws_invoicing_invoice_unit` ([#44892](https://github.com/hashicorp/terraform-provider-aws/issues/44892))

ENHANCEMENTS:

* data-source/aws_connect_routing_profile: Add `media_concurrencies.cross_channel_behavior` attribute ([#44934](https://github.com/hashicorp/terraform-provider-aws/issues/44934))
* data-source/aws_elasticache_replication_group: Add `node_group_configuration` attribute to expose node group details including availability zones, replica counts, and slot ranges ([#44879](https://github.com/hashicorp/terraform-provider-aws/issues/44879))
* data-source/aws_kinesis_stream: Add `max_record_size_in_kib` attribute ([#44915](https://github.com/hashicorp/terraform-provider-aws/issues/44915))
* data-source/aws_opensearch_domain: Add `identity_center_options` attribute ([#44626](https://github.com/hashicorp/terraform-provider-aws/issues/44626))
* provider: Support `us-isob-west-1` as a valid AWS Region ([#44944](https://github.com/hashicorp/terraform-provider-aws/issues/44944))
* resource/aws_cloudfront_distribution: Add `logging_v1_enabled` attribute ([#44838](https://github.com/hashicorp/terraform-provider-aws/issues/44838))
* resource/aws_connect_routing_profile: Add `media_concurrencies.cross_channel_behavior` argument ([#44934](https://github.com/hashicorp/terraform-provider-aws/issues/44934))
* resource/aws_ec2_client_vpn_route: Allow IPv6 address ranges for `destination_cidr_block` ([#44926](https://github.com/hashicorp/terraform-provider-aws/issues/44926))
* resource/aws_ec2_instance_connect_endpoint: Add `ip_address_type` argument ([#44616](https://github.com/hashicorp/terraform-provider-aws/issues/44616))
* resource/aws_eks_node_group: Add `max_parallel_nodes_repaired_count`, `max_parallel_nodes_repaired_percentage`, `max_unhealthy_node_threshold_count`, `max_unhealthy_node_threshold_percentage`, and `node_repair_config_overrides` to the `node_repair_config` schema ([#44894](https://github.com/hashicorp/terraform-provider-aws/issues/44894))
* resource/aws_elasticache_replication_group: Add `node_group_configuration` block to support availability zone specification and snapshot restoration for cluster mode enabled replication groups ([#44879](https://github.com/hashicorp/terraform-provider-aws/issues/44879))
* resource/aws_glue_job: Ensure that `timeout` is unconfigured for Ray jobs ([#35012](https://github.com/hashicorp/terraform-provider-aws/issues/35012))
* resource/aws_kinesis_stream: Add `max_record_size_in_kib` argument to support for Kinesis 10MiB payloads. This functionality requires the `kinesis:UpdateMaxRecordSize` IAM permission ([#44915](https://github.com/hashicorp/terraform-provider-aws/issues/44915))
* resource/aws_opensearch_domain: Add `identity_center_options` configuration block ([#44626](https://github.com/hashicorp/terraform-provider-aws/issues/44626))
* resource/aws_transfer_server: Add support for `TransferSecurityPolicy-AS2Restricted-2025-07` `security_policy_name` value ([#44865](https://github.com/hashicorp/terraform-provider-aws/issues/44865))
* resource/aws_transfer_server: Support `TransferSecurityPolicy-AS2Restricted-2025-07` as a valid value for `security_policy_name` ([#44652](https://github.com/hashicorp/terraform-provider-aws/issues/44652))

BUG FIXES:

* resource/aws_cloudfront_continuous_deployment_policy: Fix `Source type "...cloudfront.stagingDistributionDNSNamesModel" does not implement attr.Value` error. This fixes a regression introduced in [v6.17.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6170-october-16-2025) ([#44972](https://github.com/hashicorp/terraform-provider-aws/issues/44972))
* resource/aws_cloudfront_distribution: Change `logging_config.bucket` argument from `Required` to `Optional` ([#44838](https://github.com/hashicorp/terraform-provider-aws/issues/44838))
* resource/aws_cloudfront_distribution: Fix inability to configure `logging_config.include_cookies` argument while keeping V1 logging disabled ([#44838](https://github.com/hashicorp/terraform-provider-aws/issues/44838))
* resource/aws_cloudfront_vpc_origin: Fix `Source type "...cloudfront.originSSLProtocolsModel" does not implement attr.Value` and `missing required field, CreateVpcOriginInput.VpcOriginEndpointConfig` errors. This fixes a regression introduced in [v6.17.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6170-october-16-2025) ([#44861](https://github.com/hashicorp/terraform-provider-aws/issues/44861))
* resource/aws_glue_job: Allow Ray jobs to be updated ([#35012](https://github.com/hashicorp/terraform-provider-aws/issues/35012))
* resource/aws_glue_job: Allow a zero (`0`) value for `timeout` for Apache Spark streaming ETL jobs. This allows the job to be configured with no timeout ([#44920](https://github.com/hashicorp/terraform-provider-aws/issues/44920))
* resource/aws_lakeformation_lf_tags: Remove incorrect validation from `catalog_id`, `database.catalog_id`, `table.catalog_id`, and `table_with_columns.catalog_id` arguments ([#44890](https://github.com/hashicorp/terraform-provider-aws/issues/44890))
* resource/aws_launch_template: Allow an empty (`""`) value for `block_device_mappings.ebs.kms_key_id`. This fixes a regression introduced in [v6.16.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#6160-october-9-2025) ([#44708](https://github.com/hashicorp/terraform-provider-aws/issues/44708))
* resource/aws_redshift_cluster: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_cluster_snapshot: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_event_subscription: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_hsm_client_certificate: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_hsm_configuration: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_integration: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_parameter_group: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_snapshot_copy_grant: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_snapshot_schedule: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_subnet_group: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))
* resource/aws_redshift_usage_limit: Prevents errors with empty tag values. ([#44952](https://github.com/hashicorp/terraform-provider-aws/issues/44952))

## 6.19.0 (October 30, 2025)

FEATURES:

* **New Data Source:** `aws_ecrpublic_images` ([#44795](https://github.com/hashicorp/terraform-provider-aws/issues/44795))
* **New Resource:** `aws_lakeformation_identity_center_configuration` ([#44867](https://github.com/hashicorp/terraform-provider-aws/issues/44867))

ENHANCEMENTS:

* action/aws_lambda_invoke: Output logs in a progress message when `log_type` is `Tail` ([#44843](https://github.com/hashicorp/terraform-provider-aws/issues/44843))
* data-source/aws_imagebuilder_image_recipe: Add `ami_tags` attribute ([#44731](https://github.com/hashicorp/terraform-provider-aws/issues/44731))
* data-source/aws_lb_listener_rule: Add `regex_values` attribute to `condition.host_header`, `condition.http_header` and `condition.path_pattern` blocks ([#44741](https://github.com/hashicorp/terraform-provider-aws/issues/44741))
* data-source/aws_lb_listener_rule: Add `transform` attribute ([#44702](https://github.com/hashicorp/terraform-provider-aws/issues/44702))
* resource/aws_bedrockagentcore_gateway: Add validator to ensure correct `authorizer_configuration` and `authorizer_type` config ([#44826](https://github.com/hashicorp/terraform-provider-aws/issues/44826))
* resource/aws_emrserverless_application: Add `monitoring_configuration` argument ([#43317](https://github.com/hashicorp/terraform-provider-aws/issues/43317))
* resource/aws_emrserverless_application: Add `runtime_configuration` argument ([#43302](https://github.com/hashicorp/terraform-provider-aws/issues/43302))
* resource/aws_identitystore_group: Adds `arn` attribute. ([#44867](https://github.com/hashicorp/terraform-provider-aws/issues/44867))
* resource/aws_imagebuilder_image_recipe: Add `ami_tags` argument ([#44731](https://github.com/hashicorp/terraform-provider-aws/issues/44731))
* resource/aws_lb_listener_rule: Add `regex_values` argument to `condition.host_header`, `condition.http_header` and `condition.path_pattern` blocks ([#44741](https://github.com/hashicorp/terraform-provider-aws/issues/44741))
* resource/aws_lb_listener_rule: Add `transform` configuration block ([#44702](https://github.com/hashicorp/terraform-provider-aws/issues/44702))
* resource/aws_lb_listener_rule: The `values` argument in `condition.host_header`, `condition.http_header` and `condition.path_pattern` is now optional ([#44741](https://github.com/hashicorp/terraform-provider-aws/issues/44741))
* resource/aws_quicksight_data_set: Increase upper limit of `physical_table_map.relational_table.name` from 64 to 256 characters ([#44807](https://github.com/hashicorp/terraform-provider-aws/issues/44807))
* resource/aws_sagemaker_notebook_instance: Add `notebook-al2023-v1` to valid `platform_identifier` values ([#44570](https://github.com/hashicorp/terraform-provider-aws/issues/44570))
* resource/aws_sqs_queue: Remove `account_id` and `region` from Resource Identity schema ([#44846](https://github.com/hashicorp/terraform-provider-aws/issues/44846))
* resource/aws_sqs_queue_policy: Remove `account_id` and `region` from Resource Identity schema ([#44846](https://github.com/hashicorp/terraform-provider-aws/issues/44846))
* resource/aws_sqs_queue_redrive_allow_policy: Remove `account_id` and `region` from Resource Identity schema ([#44846](https://github.com/hashicorp/terraform-provider-aws/issues/44846))
* resource/aws_sqs_queue_redrive_policy: Remove `account_id` and `region` from Resource Identity schema ([#44846](https://github.com/hashicorp/terraform-provider-aws/issues/44846))

BUG FIXES:

* data-source/aws_lakeformation_permissions: Allows IAM Identity Center Groups as `principal`. ([#44867](https://github.com/hashicorp/terraform-provider-aws/issues/44867))
* provider: Fix crash when setting override region during provider initialization ([#44860](https://github.com/hashicorp/terraform-provider-aws/issues/44860))
* resource/aws_bedrockagentcore_gateway: Change `authorizer_configuration` block from `Required` to `Optional` ([#44812](https://github.com/hashicorp/terraform-provider-aws/issues/44812))
* resource/aws_bedrockagentcore_gateway: Mark `authorizer_type` argument as `ForceNew` ([#44812](https://github.com/hashicorp/terraform-provider-aws/issues/44812))
* resource/aws_lakeformation_permissions: Allows IAM Identity Center Groups as `principal`. ([#44867](https://github.com/hashicorp/terraform-provider-aws/issues/44867))

## 6.18.0 (October 23, 2025)

NOTES:

* data-source/aws_organizations_organization: The `accounts.status` and `non_master_accounts.status` attributes are deprecated. Use the `accounts.state` and `non_master_accounts.state` attributes instead. ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))
* data-source/aws_organizations_organizational_unit_child_accounts: The `accounts.status` attribute is deprecated. Use `accounts.state` instead. ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))
* data-source/aws_organizations_organizational_unit_descendant_accounts: The `accounts.status` attribute is deprecated. Use `accounts.state` instead. ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))
* resource/aws_organizations_account: The `status` attribute is deprecated. Use `state` instead. ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))
* resource/aws_organizations_organization: The `accounts.status` and `non_master_accounts.status` attributes are deprecated. Use the `accounts.state` and `non_master_accounts.state` attributes instead. ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))

FEATURES:

* **New List Resource:** `aws_iam_policy` ([#44703](https://github.com/hashicorp/terraform-provider-aws/issues/44703))
* **New List Resource:** `aws_iam_role_policy_attachment` ([#44739](https://github.com/hashicorp/terraform-provider-aws/issues/44739))
* **New Resource:** `aws_bedrockagentcore_memory` ([#44306](https://github.com/hashicorp/terraform-provider-aws/issues/44306))
* **New Resource:** `aws_bedrockagentcore_memory_strategy` ([#44306](https://github.com/hashicorp/terraform-provider-aws/issues/44306))
* **New Resource:** `aws_bedrockagentcore_oauth2_credential_provider` ([#44307](https://github.com/hashicorp/terraform-provider-aws/issues/44307))
* **New Resource:** `aws_bedrockagentcore_token_vault_cmk` ([#44606](https://github.com/hashicorp/terraform-provider-aws/issues/44606))
* **New Resource:** `aws_bedrockagentcore_workload_identity` ([#44308](https://github.com/hashicorp/terraform-provider-aws/issues/44308))

ENHANCEMENTS:

* data-source/aws_iam_policy: Adds validation for `path_prefix` attribute ([#44703](https://github.com/hashicorp/terraform-provider-aws/issues/44703))
* data-source/aws_organizations_organization: Add `state`, `joined_method`, and `joined_timestamp` attributes to the `accounts` and `non_master_accounts` blocks ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))
* data-source/aws_organizations_organizational_unit_child_accounts: Add `state`, `joined_method`, and `joined_timestamp` attributes to the `accounts` block ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))
* data-source/aws_organizations_organizational_unit_descendant_accounts: Add `state`, `joined_method`, and `joined_timestamp` attributes to the `accounts` block ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))
* resource/aws_appstream_directory_config: Add `certificate_based_auth_properties` argument ([#44679](https://github.com/hashicorp/terraform-provider-aws/issues/44679))
* resource/aws_iam_policy: Adds validation for `path` attribute ([#44703](https://github.com/hashicorp/terraform-provider-aws/issues/44703))
* resource/aws_odb_network: Add `delete_associated_resources` attribute to enable practitioner to delete associated oci resource. ([#44754](https://github.com/hashicorp/terraform-provider-aws/issues/44754))
* resource/aws_organizations_account: Add `state` attribute ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))
* resource/aws_organizations_organization: Add `state`, `joined_method`, and `joined_timestamp` attributes to the `accounts` and `non_master_accounts` blocks ([#44327](https://github.com/hashicorp/terraform-provider-aws/issues/44327))

BUG FIXES:

* data-source/aws_vpn_connection: Properly set `tags` attribute ([#44761](https://github.com/hashicorp/terraform-provider-aws/issues/44761))
* resource/aws_rds_cluster: Fix "When modifying Provisioned IOPS storage, specify a value for both allocated storage and iops" error when updating RDS clusters with Provisioned IOPS storage ([#44706](https://github.com/hashicorp/terraform-provider-aws/issues/44706))
* resource/guardduty_detector_feature: Fix `additional_configuration` block to ignore ordering ([#44627](https://github.com/hashicorp/terraform-provider-aws/issues/44627))

## 6.17.0 (October 16, 2025)

NOTES:

* resource/aws_quicksight_account_subscription: Because we cannot easily test all this functionality, it is best effort and we ask for community help in testing ([#44638](https://github.com/hashicorp/terraform-provider-aws/issues/44638))

FEATURES:

* **New Data Source:** `aws_rds_global_cluster` ([#37286](https://github.com/hashicorp/terraform-provider-aws/issues/37286))
* **New Data Source:** `aws_vpn_connection` ([#44622](https://github.com/hashicorp/terraform-provider-aws/issues/44622))
* **New List Resource:** `aws_subnet` ([#44671](https://github.com/hashicorp/terraform-provider-aws/issues/44671))
* **New List Resource:** `aws_vpc` ([#44609](https://github.com/hashicorp/terraform-provider-aws/issues/44609))
* **New Resource:** `aws_bedrockagentcore_agent_runtime` ([#44301](https://github.com/hashicorp/terraform-provider-aws/issues/44301))
* **New Resource:** `aws_bedrockagentcore_agent_runtime_endpoint` ([#44301](https://github.com/hashicorp/terraform-provider-aws/issues/44301))
* **New Resource:** `aws_bedrockagentcore_api_key_credential_provider` ([#44302](https://github.com/hashicorp/terraform-provider-aws/issues/44302))
* **New Resource:** `aws_bedrockagentcore_browser` ([#44303](https://github.com/hashicorp/terraform-provider-aws/issues/44303))
* **New Resource:** `aws_bedrockagentcore_code_interpreter` ([#44304](https://github.com/hashicorp/terraform-provider-aws/issues/44304))
* **New Resource:** `aws_bedrockagentcore_gateway` ([#44305](https://github.com/hashicorp/terraform-provider-aws/issues/44305))
* **New Resource:** `aws_bedrockagentcore_gateway_target` ([#44305](https://github.com/hashicorp/terraform-provider-aws/issues/44305))

ENHANCEMENTS:

* resource/aws_imagebuilder_container_recipe: Update EBS `throughput` maximum validation from 1000 to 2000 MiB/s for gp3 volumes ([#44604](https://github.com/hashicorp/terraform-provider-aws/issues/44604))
* resource/aws_imagebuilder_image_recipe: Update EBS `throughput` maximum validation from 1000 to 2000 MiB/s for gp3 volumes ([#44604](https://github.com/hashicorp/terraform-provider-aws/issues/44604))
* resource/aws_launch_template: Update EBS `throughput` maximum validation from 1000 to 2000 MiB/s for gp3 volumes ([#44604](https://github.com/hashicorp/terraform-provider-aws/issues/44604))
* resource/aws_quicksight_account_subscription: Add `admin_pro_group`, `author_pro_group`, and `reader_pro_group` arguments ([#44638](https://github.com/hashicorp/terraform-provider-aws/issues/44638))

BUG FIXES:

* resource/aws_ec2_transit_gateway_route_table_propagation.test: Fix bug causing `inconsistent final plan` errors ([#44542](https://github.com/hashicorp/terraform-provider-aws/issues/44542))
* resource/aws_lambda_function: Reset non-API attributes (`source_code_hash`, `s3_bucket`, `s3_key`, `s3_object_version` and `filename`) to their previous values when an update operation fails ([#42829](https://github.com/hashicorp/terraform-provider-aws/issues/42829))

## 6.16.0 (October 9, 2025)

FEATURES:

* **New Action:** `aws_transcribe_start_transcription_job` ([#44445](https://github.com/hashicorp/terraform-provider-aws/issues/44445))
* **New Data Source:** `aws_odb_cloud_autonomous_vm_clusters` ([#44336](https://github.com/hashicorp/terraform-provider-aws/issues/44336))
* **New Data Source:** `aws_odb_cloud_exadata_infrastructures` ([#44336](https://github.com/hashicorp/terraform-provider-aws/issues/44336))
* **New Data Source:** `aws_odb_cloud_vm_clusters` ([#44336](https://github.com/hashicorp/terraform-provider-aws/issues/44336))
* **New Data Source:** `aws_odb_network_peering_connections` ([#44336](https://github.com/hashicorp/terraform-provider-aws/issues/44336))
* **New Data Source:** `aws_odb_networks` ([#44336](https://github.com/hashicorp/terraform-provider-aws/issues/44336))
* **New Resource:** `aws_prometheus_resource_policy` ([#44256](https://github.com/hashicorp/terraform-provider-aws/issues/44256))
* **New Resource:** `aws_transfer_host_key` ([#44559](https://github.com/hashicorp/terraform-provider-aws/issues/44559))
* **New Resource:** `aws_transfer_web_app` ([#42708](https://github.com/hashicorp/terraform-provider-aws/issues/42708))
* **New Resource:** `aws_transfer_web_app_customization` ([#42708](https://github.com/hashicorp/terraform-provider-aws/issues/42708))

ENHANCEMENTS:

* resource/aws_codebuild_project: Add `auto_retry_limit` argument ([#40035](https://github.com/hashicorp/terraform-provider-aws/issues/40035))
* resource/aws_emrserverless_application: Add `scheduler_configuration` block ([#44589](https://github.com/hashicorp/terraform-provider-aws/issues/44589))
* resource/aws_lambda_event_source_mapping: Add `schema_registry_config` configuration blocks to `amazon_managed_kafka_event_source_config` and `self_managed_kafka_event_source_config` blocks ([#44540](https://github.com/hashicorp/terraform-provider-aws/issues/44540))
* resource/aws_ssmcontacts_contact: Add resource identity support ([#44548](https://github.com/hashicorp/terraform-provider-aws/issues/44548))
* resource/aws_vpclattice_resource_gateway: Add `ipv4_addresses_per_eni` argument ([#44560](https://github.com/hashicorp/terraform-provider-aws/issues/44560))

BUG FIXES:

* provider: Correctly validate AWS European Sovereign Cloud Regions in ARNs ([#44573](https://github.com/hashicorp/terraform-provider-aws/issues/44573))
* provider: Fix `Missing Resource Identity After Update` errors for non-refreshed and failed updates of Plugin Framework based resources ([#44518](https://github.com/hashicorp/terraform-provider-aws/issues/44518))
* provider: Fix `Unexpected Identity Change` errors when fully-null identity values in state are updated to valid values for Plugin Framework based resources ([#44518](https://github.com/hashicorp/terraform-provider-aws/issues/44518))
* resource/aws_datazone_environment: Correctly updates `glossary_terms`. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_environment: Prevents `unknown value` error when optional `account_identifier` is not specified. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_environment: Prevents `unknown value` error when optional `account_region` is not specified. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_environment: Prevents error when updating. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_environment: Prevents occasional `unexpected state` error when deleting. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_environment: Properly passes `blueprint_identifier` on creation. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_environment: Sets values for `user_parameters` when importing. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_environment: Values in `user_parameters` should not be updateable. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_project: No longer ignores errors when deleting. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_datazone_project: No longer returns error when already deleting. ([#44491](https://github.com/hashicorp/terraform-provider-aws/issues/44491))
* resource/aws_dynamodb_table: Do not retry on `LimitExceededException` ([#44576](https://github.com/hashicorp/terraform-provider-aws/issues/44576))
* resource/aws_ivschat_room: Set `maximum_message_rate_per_second` validation maximum to `100` ([#44572](https://github.com/hashicorp/terraform-provider-aws/issues/44572))
* resource/aws_launch_template: `kms_key_id` validation now accepts key ID, alias, and alias ARN in addition to key ARN ([#44505](https://github.com/hashicorp/terraform-provider-aws/issues/44505))
* resource/aws_servicecatalog_portfolio_share: Add global mutex lock around create and delete operations to prevent `ThrottlingException` errors ([#24730](https://github.com/hashicorp/terraform-provider-aws/issues/24730))

## 6.15.0 (October 2, 2025)

BREAKING CHANGES:

* resource/aws_ecs_service: Fix behavior when updating `capacity_provider_strategy` to avoid ECS service recreation after recent AWS changes ([#43533](https://github.com/hashicorp/terraform-provider-aws/issues/43533))

FEATURES:

* **New Action:** `aws_codebuild_start_build` ([#44444](https://github.com/hashicorp/terraform-provider-aws/issues/44444))
* **New Action:** `aws_events_put_events` ([#44487](https://github.com/hashicorp/terraform-provider-aws/issues/44487))
* **New Action:** `aws_sfn_start_execution` ([#44464](https://github.com/hashicorp/terraform-provider-aws/issues/44464))
* **New Data Source:** `aws_appconfig_application` ([#44168](https://github.com/hashicorp/terraform-provider-aws/issues/44168))
* **New Data Source:** `aws_odb_db_node` ([#43792](https://github.com/hashicorp/terraform-provider-aws/issues/43792))
* **New Data Source:** `aws_odb_db_nodes` ([#43792](https://github.com/hashicorp/terraform-provider-aws/issues/43792))
* **New Data Source:** `aws_odb_db_server` ([#43792](https://github.com/hashicorp/terraform-provider-aws/issues/43792))
* **New Data Source:** `aws_odb_db_servers` ([#43792](https://github.com/hashicorp/terraform-provider-aws/issues/43792))
* **New Data Source:** `aws_odb_db_system_shapes` ([#43825](https://github.com/hashicorp/terraform-provider-aws/issues/43825))
* **New Data Source:** `aws_odb_gi_versions` ([#43825](https://github.com/hashicorp/terraform-provider-aws/issues/43825))
* **New Resource:** `aws_lakeformation_lf_tag_expression` ([#43883](https://github.com/hashicorp/terraform-provider-aws/issues/43883))

ENHANCEMENTS:

* data-source/aws_dms_endpoint: Add `mysql_settings` attribute ([#44516](https://github.com/hashicorp/terraform-provider-aws/issues/44516))
* data-source/aws_ec2_instance_type_offering: Add `location` attribute ([#44328](https://github.com/hashicorp/terraform-provider-aws/issues/44328))
* data-source/aws_rds_proxy: Add `default_auth_scheme` attribute ([#44309](https://github.com/hashicorp/terraform-provider-aws/issues/44309))
* resource/aws_cleanrooms_configured_table: Add resource identity support ([#44435](https://github.com/hashicorp/terraform-provider-aws/issues/44435))
* resource/aws_cloudfront_distribution: Add `ip_address_type` argument to `origin.custom_origin_config` block ([#44463](https://github.com/hashicorp/terraform-provider-aws/issues/44463))
* resource/aws_connect_instance: Add resource identity support ([#44346](https://github.com/hashicorp/terraform-provider-aws/issues/44346))
* resource/aws_connect_phone_number: Add resource identity support ([#44365](https://github.com/hashicorp/terraform-provider-aws/issues/44365))
* resource/aws_dms_endpoint: Add `mysql_settings` configuration block ([#44516](https://github.com/hashicorp/terraform-provider-aws/issues/44516))
* resource/aws_dsql_cluster: Adds attribute `force_destroy`. ([#44406](https://github.com/hashicorp/terraform-provider-aws/issues/44406))
* resource/aws_ebs_volume: Update `throughput` maximum validation from 1000 to 2000 MiB/s for gp3 volumes ([#44514](https://github.com/hashicorp/terraform-provider-aws/issues/44514))
* resource/aws_ecs_capacity_provider: Add `cluster` and `managed_instances_provider` arguments ([#44509](https://github.com/hashicorp/terraform-provider-aws/issues/44509))
* resource/aws_ecs_capacity_provider: Make `auto_scaling_group_provider` optional ([#44509](https://github.com/hashicorp/terraform-provider-aws/issues/44509))
* resource/aws_iam_service_specific_credential: Add support for Bedrock API keys with `credential_age_days`, `service_credential_alias`, `service_credential_secret`, `create_date`, and `expiration_date` attributes ([#44299](https://github.com/hashicorp/terraform-provider-aws/issues/44299))
* resource/aws_networkfirewall_logging_configuration: Add `enable_monitoring_dashboard` argument ([#44515](https://github.com/hashicorp/terraform-provider-aws/issues/44515))
* resource/aws_opensearch_domain: Add `aiml_options` argument ([#44417](https://github.com/hashicorp/terraform-provider-aws/issues/44417))
* resource/aws_pinpointsmsvoicev2_phone_number: Update `two_way_channel_arn` argument to accept `connect.[region].amazonaws.com` in addition to ARNs ([#44372](https://github.com/hashicorp/terraform-provider-aws/issues/44372))
* resource/aws_rds_proxy: Add `default_auth_scheme` argument ([#44309](https://github.com/hashicorp/terraform-provider-aws/issues/44309))
* resource/aws_rds_proxy: Make `auth` configuration block optional ([#44309](https://github.com/hashicorp/terraform-provider-aws/issues/44309))
* resource/aws_route53recoverycontrolconfig_cluster: Add `network_type` argument ([#44377](https://github.com/hashicorp/terraform-provider-aws/issues/44377))
* resource/aws_route53recoverycontrolconfig_cluster: Add tagging support ([#44473](https://github.com/hashicorp/terraform-provider-aws/issues/44473))
* resource/aws_route53recoverycontrolconfig_control_panel: Add tagging support ([#44473](https://github.com/hashicorp/terraform-provider-aws/issues/44473))
* resource/aws_route53recoverycontrolconfig_safety_rule: Add tagging support ([#44473](https://github.com/hashicorp/terraform-provider-aws/issues/44473))
* resource/aws_s3control_bucket: Add resource identity support ([#44379](https://github.com/hashicorp/terraform-provider-aws/issues/44379))
* resource/aws_sfn_activity: Add `arn` argument ([#44408](https://github.com/hashicorp/terraform-provider-aws/issues/44408))
* resource/aws_sfn_activity: Add resource identity support ([#44408](https://github.com/hashicorp/terraform-provider-aws/issues/44408))
* resource/aws_sfn_alias: Add resource identity support ([#44408](https://github.com/hashicorp/terraform-provider-aws/issues/44408))
* resource/aws_ssmcontacts_contact_channel: Add resource identity support ([#44369](https://github.com/hashicorp/terraform-provider-aws/issues/44369))

BUG FIXES:

* data-source/aws_lb: Fix `Invalid address to set: []string{"secondary_ips_auto_assigned_per_subnet"}` errors ([#44485](https://github.com/hashicorp/terraform-provider-aws/issues/44485))
* data-source/aws_networkfirewall_firewall_policy: Fix failure to retrieve multiple `firewall_policy.stateful_rule_group_reference` attributes ([#44482](https://github.com/hashicorp/terraform-provider-aws/issues/44482))
* data-source/aws_servicequotas_service_quota: Fixed a panic that occurred when a non-existing `quota_name` was provided ([#44449](https://github.com/hashicorp/terraform-provider-aws/issues/44449))
* resource/aws_bedrock_provisioned_model_throughput: Fix `AttributeName("arn") still remains in the path: could not find attribute or block "arn" in schema` errors when upgrading from a pre-v6.0.0 provider version ([#44434](https://github.com/hashicorp/terraform-provider-aws/issues/44434))
* resource/aws_chatbot_slack_channel_configuration: Force resource replacement when `configuration_name` is modified ([#43996](https://github.com/hashicorp/terraform-provider-aws/issues/43996))
* resource/aws_cloudwatch_event_rule: Do not retry on `LimitExceededException` ([#44489](https://github.com/hashicorp/terraform-provider-aws/issues/44489))
* resource/aws_cloudwatch_log_resource_policy: Do not retry on `LimitExceededException` ([#44522](https://github.com/hashicorp/terraform-provider-aws/issues/44522))
* resource/aws_default_vpc: Correctly set `ipv6_cidr_block` when the VPC has multiple associated IPv6 CIDRs ([#44362](https://github.com/hashicorp/terraform-provider-aws/issues/44362))
* resource/aws_dms_endpoint: Ensure that `postgres_settings` are updated ([#44389](https://github.com/hashicorp/terraform-provider-aws/issues/44389))
* resource/aws_dsql_cluster: Prevents error when optional attribute `deletion_protection_enabled` not set. ([#44406](https://github.com/hashicorp/terraform-provider-aws/issues/44406))
* resource/aws_eks_cluster: Change `compute_config`, `kubernetes_network_config.elastic_load_balancing`, and `storage_config.` to Optional and Computed, allowing EKS Auto Mode settings to be enabled, disabled, and removed from configuration ([#44334](https://github.com/hashicorp/terraform-provider-aws/issues/44334))
* resource/aws_elastic_beanstalk_configuration_template: Fix `inconsistent final plan` error in some cases with `setting` elements. ([#44461](https://github.com/hashicorp/terraform-provider-aws/issues/44461))
* resource/aws_elastic_beanstalk_environment: Fix `inconsistent final plan` error in some cases with `setting` elements. ([#44461](https://github.com/hashicorp/terraform-provider-aws/issues/44461))
* resource/aws_elasticache_cluster: Fix `provider produced unexpected value` for `cache_usage_limits` argument. ([#43841](https://github.com/hashicorp/terraform-provider-aws/issues/43841))
* resource/aws_fsx_lustre_file_system: Fixed to update `metadata_configuration` first to allow simultaneous increase of `metadata_configuration.iops` and `storage_capacity` ([#44456](https://github.com/hashicorp/terraform-provider-aws/issues/44456))
* resource/aws_instance: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panics when `capacity_reservation_target` is empty ([#44459](https://github.com/hashicorp/terraform-provider-aws/issues/44459))
* resource/aws_kinesisanalyticsv2_application: Ensure that configured `application_configuration.run_configuration` values are respected during update ([#43490](https://github.com/hashicorp/terraform-provider-aws/issues/43490))
* resource/aws_odb_cloud_autonomous_vm_cluster : Fixed planmodifier for computed attribute. ([#44401](https://github.com/hashicorp/terraform-provider-aws/issues/44401))
* resource/aws_odb_cloud_vm_cluster : Fixed planmodifier for computed attribute. Fixed planmodifier from display_name attribute. ([#44401](https://github.com/hashicorp/terraform-provider-aws/issues/44401))
* resource/aws_odb_cloud_vm_cluster : Fixed planmodifier for data_storage_size_in_tbs. Marked it mandatory. Fixed gi-version issue during creation ([#44498](https://github.com/hashicorp/terraform-provider-aws/issues/44498))
* resource/aws_odb_network_peering_connection : Fixed planmodifier for computed attribute. ([#44401](https://github.com/hashicorp/terraform-provider-aws/issues/44401))
* resource/aws_rds_cluster: Fixes error when setting `database_insights_mode` with `global_cluster_identifier`. ([#44404](https://github.com/hashicorp/terraform-provider-aws/issues/44404))
* resource/aws_route53_health_check: Fix `child_health_threshold` to properly accept explicitly specified zero value ([#44006](https://github.com/hashicorp/terraform-provider-aws/issues/44006))
* resource/aws_s3_bucket_lifecycle_configuration: Allows unsetting `noncurrent_version_expiration.newer_noncurrent_versions` and `noncurrent_version_transition.newer_noncurrent_versions`. ([#44442](https://github.com/hashicorp/terraform-provider-aws/issues/44442))
* resource/aws_s3_bucket_lifecycle_configuration: Do not warn if no filter element is set ([#43590](https://github.com/hashicorp/terraform-provider-aws/issues/43590))
* resource/aws_vpc: Correctly set `ipv6_cidr_block` when the VPC has multiple associated IPv6 CIDRs ([#44362](https://github.com/hashicorp/terraform-provider-aws/issues/44362))

## 6.14.1 (September 22, 2025)

NOTES:

* provider: This release contains both internal provider fixes and a Terraform Plugin SDK V2 update related to a [regression](https://github.com/hashicorp/terraform-provider-aws/issues/44366) which may impact resources that support resource identity ([#44375](https://github.com/hashicorp/terraform-provider-aws/issues/44375))

BUG FIXES:

* provider: Fix `Missing Resource Identity After Update` errors for non-refreshed and failed updates ([#44375](https://github.com/hashicorp/terraform-provider-aws/issues/44375))
* provider: Fix `Unexpected Identity Change` errors when fully-null identity values in state are updated to valid values ([#44375](https://github.com/hashicorp/terraform-provider-aws/issues/44375))

## 6.14.0 (September 18, 2025)

FEATURES:

* **New Action:** `aws_cloudfront_create_invalidation` ([#43955](https://github.com/hashicorp/terraform-provider-aws/issues/43955))
* **New Action:** `aws_ec2_stop_instance` ([#43700](https://github.com/hashicorp/terraform-provider-aws/issues/43700))
* **New Action:** `aws_lambda_invoke` ([#43972](https://github.com/hashicorp/terraform-provider-aws/issues/43972))
* **New Action:** `aws_ses_send_email` ([#44214](https://github.com/hashicorp/terraform-provider-aws/issues/44214))
* **New Action:** `aws_sns_publish` ([#44232](https://github.com/hashicorp/terraform-provider-aws/issues/44232))
* **New Data Source:** `aws_billing_views` ([#44272](https://github.com/hashicorp/terraform-provider-aws/issues/44272))
* **New Data Source:** `aws_odb_cloud_autonomous_vm_cluster` ([#43809](https://github.com/hashicorp/terraform-provider-aws/issues/43809))
* **New Data Source:** `aws_odb_cloud_exadata_infrastructure` ([#43650](https://github.com/hashicorp/terraform-provider-aws/issues/43650))
* **New Data Source:** `aws_odb_cloud_vm_cluster` ([#43790](https://github.com/hashicorp/terraform-provider-aws/issues/43790))
* **New Data Source:** `aws_odb_network` ([#43715](https://github.com/hashicorp/terraform-provider-aws/issues/43715))
* **New Data Source:** `aws_odb_network_peering_connection` ([#43757](https://github.com/hashicorp/terraform-provider-aws/issues/43757))
* **New List Resource:** `aws_batch_job_queue` ([#43960](https://github.com/hashicorp/terraform-provider-aws/issues/43960))
* **New List Resource:** `aws_cloudwatch_log_group` ([#44129](https://github.com/hashicorp/terraform-provider-aws/issues/44129))
* **New List Resource:** `aws_iam_role` ([#44129](https://github.com/hashicorp/terraform-provider-aws/issues/44129))
* **New List Resource:** `aws_instance` ([#44129](https://github.com/hashicorp/terraform-provider-aws/issues/44129))
* **New Resource:** `aws_controltower_baseline` ([#42397](https://github.com/hashicorp/terraform-provider-aws/issues/42397))
* **New Resource:** `aws_odb_cloud_autonomous_vm_cluster` ([#43809](https://github.com/hashicorp/terraform-provider-aws/issues/43809))
* **New Resource:** `aws_odb_cloud_exadata_infrastructure` ([#43650](https://github.com/hashicorp/terraform-provider-aws/issues/43650))
* **New Resource:** `aws_odb_cloud_vm_cluster` ([#43790](https://github.com/hashicorp/terraform-provider-aws/issues/43790))
* **New Resource:** `aws_odb_network` ([#43715](https://github.com/hashicorp/terraform-provider-aws/issues/43715))
* **New Resource:** `aws_odb_network_peering_connection` ([#43757](https://github.com/hashicorp/terraform-provider-aws/issues/43757))

ENHANCEMENTS:

* resource/aws_ecs_service: Add `deployment_configuration.lifecycle_hook.hook_details` argument ([#44289](https://github.com/hashicorp/terraform-provider-aws/issues/44289))
* resource/aws_rds_global_cluster: Remove provider-side conflict between `source_db_cluster_identifier` and `engine` arguments ([#44252](https://github.com/hashicorp/terraform-provider-aws/issues/44252))
* resource/aws_scheduler_schedule: Add `action_after_completion` argument ([#44264](https://github.com/hashicorp/terraform-provider-aws/issues/44264))
* resource/aws_sfn_state_machine: Add resource identity support ([#44286](https://github.com/hashicorp/terraform-provider-aws/issues/44286))

BUG FIXES:

* resource/aws_elasticache_user_group: Ignore `InvalidParameterValue: User xxx is not a member of user group xxx` errors during group modification ([#43520](https://github.com/hashicorp/terraform-provider-aws/issues/43520))
* resource/aws_sagemaker_endpoint_configuration: Fix panic when empty `async_inference_config.output_config.notification_config` block is specified ([#44310](https://github.com/hashicorp/terraform-provider-aws/issues/44310))

## 6.13.0 (September 11, 2025)

ENHANCEMENTS:

* data-source/aws_budgets_budget: Add `billing_view_arn` attribute ([#44241](https://github.com/hashicorp/terraform-provider-aws/issues/44241))
* data-source/aws_dynamodb_table: Add `warm_throughput` and `global_secondary_index.warm_throughput` attributes ([#41308](https://github.com/hashicorp/terraform-provider-aws/issues/41308))
* data-source/aws_elastic_beanstalk_hosted_zone: Add hosted zone IDs for `ap-southeast-5`, `ap-southeast-7`, `eu-south-2`, and `me-central-1` AWS Regions ([#44132](https://github.com/hashicorp/terraform-provider-aws/issues/44132))
* data-source/aws_elb_hosted_zone_id: Add hosted zone ID for `ap-southeast-6` AWS Region ([#44132](https://github.com/hashicorp/terraform-provider-aws/issues/44132))
* data-source/aws_lb_hosted_zone_id: Add hosted zone IDs for `ap-southeast-6` AWS Region ([#44132](https://github.com/hashicorp/terraform-provider-aws/issues/44132))
* data-source/aws_s3_bucket: Add hosted zone ID for `ap-southeast-6` AWS Region ([#44132](https://github.com/hashicorp/terraform-provider-aws/issues/44132))
* resource/aws_appautoscaling_policy: Add `predictive_scaling_policy_configuration` argument ([#44211](https://github.com/hashicorp/terraform-provider-aws/issues/44211))
* resource/aws_appautoscaling_policy: Add plan-time validation of `policy_type` ([#44211](https://github.com/hashicorp/terraform-provider-aws/issues/44211))
* resource/aws_appautoscaling_policy: Add plan-time validation of `step_scaling_policy_configuration.adjustment_type` and `step_scaling_policy_configuration.metric_aggregation_type` ([#44211](https://github.com/hashicorp/terraform-provider-aws/issues/44211))
* resource/aws_bedrock_guardrail: Add `input_action`, `output_action`, `input_enabled`, and `output_enabled` arguments to `word_policy_config.managed_word_lists_config` and `word_policy_config.words_config` configuration blocks ([#44224](https://github.com/hashicorp/terraform-provider-aws/issues/44224))
* resource/aws_budgets_budget: Add `billing_view_arn` argument ([#44241](https://github.com/hashicorp/terraform-provider-aws/issues/44241))
* resource/aws_cloudfront_distribution: Add `origin.response_completion_timeout` argument ([#44163](https://github.com/hashicorp/terraform-provider-aws/issues/44163))
* resource/aws_codebuild_webhook: Add `pull_request_build_policy` configuration block ([#44201](https://github.com/hashicorp/terraform-provider-aws/issues/44201))
* resource/aws_dynamodb_table: Add `warm_throughput` and `global_secondary_index.warm_throughput` arguments ([#41308](https://github.com/hashicorp/terraform-provider-aws/issues/41308))
* resource/aws_ecs_account_setting_default: Support `dualStackIPv6` as a valid value for `name` ([#44165](https://github.com/hashicorp/terraform-provider-aws/issues/44165))
* resource/aws_glue_catalog_table_optimizer: Add `iceberg_configuration.run_rate_in_hours` argument to `retention_configuration` and `orphan_file_deletion_configuration` blocks ([#44207](https://github.com/hashicorp/terraform-provider-aws/issues/44207))
* resource/aws_networkfirewall_rule_group: Add IPv6 CIDR block support to `address_definition` arguments in `source` and `destination` blocks within `rule_group.rules_source.stateless_rules_and_custom_actions.stateless_rule.rule_definition.match_attributes` ([#44215](https://github.com/hashicorp/terraform-provider-aws/issues/44215))
* resource/aws_networkmanager_vpc_attachment: Add `options.dns_support` and `options.security_group_referencing_support` arguments ([#43742](https://github.com/hashicorp/terraform-provider-aws/issues/43742))
* resource/aws_networkmanager_vpc_attachment: Change `options` to Optional and Computed ([#43742](https://github.com/hashicorp/terraform-provider-aws/issues/43742))
* resource/aws_opensearch_package: Add `engine_version` argument ([#44155](https://github.com/hashicorp/terraform-provider-aws/issues/44155))
* resource/aws_opensearch_package: Add waiter to ensure package validation completes ([#44155](https://github.com/hashicorp/terraform-provider-aws/issues/44155))
* resource/aws_synthetics_canary: Add `schedule.retry_config` configuration block ([#44244](https://github.com/hashicorp/terraform-provider-aws/issues/44244))
* resource/aws_vpc_endpoint: Add resource identity support ([#44194](https://github.com/hashicorp/terraform-provider-aws/issues/44194))
* resource/aws_vpc_security_group_egress_rule: Add resource identity support ([#44198](https://github.com/hashicorp/terraform-provider-aws/issues/44198))
* resource/aws_vpc_security_group_ingress_rule: Add resource identity support ([#44198](https://github.com/hashicorp/terraform-provider-aws/issues/44198))

BUG FIXES:

* resource/aws_appautoscaling_policy: Fix `interface conversion: interface {} is nil, not map[string]interface {}` panics when `step_scaling_policy_configuration` is empty ([#44211](https://github.com/hashicorp/terraform-provider-aws/issues/44211))
* resource/aws_cognito_managed_login_branding: Fix `reading Cognito Managed Login Branding by client ... couldn't find resource` errors when a user pool contains multiple client apps ([#44204](https://github.com/hashicorp/terraform-provider-aws/issues/44204))
* resource/aws_eks_cluster: Supports null `compute_config.node_role_arn` when disabling auto mode or built-in node pools ([#42483](https://github.com/hashicorp/terraform-provider-aws/issues/42483))
* resource/aws_flow_log: Fix `Error decoding ... from prior state: unsupported attribute "log_group_name"` errors when upgrading from a pre-v6.0.0 provider version ([#44191](https://github.com/hashicorp/terraform-provider-aws/issues/44191))
* resource/aws_launch_template: Fix `Error decoding ... from prior state: unsupported attribute "elastic_gpu_specifications"` errors when upgrading from a pre-v6.0.0 provider version ([#44195](https://github.com/hashicorp/terraform-provider-aws/issues/44195))
* resource/aws_rds_cluster_role_association: Make `feature_name` optional ([#44143](https://github.com/hashicorp/terraform-provider-aws/issues/44143))
* resource/aws_s3_bucket_lifecycle_configuration: Ignore `MethodNotAllowed` errors when deleting non-existent lifecycle configurations ([#44189](https://github.com/hashicorp/terraform-provider-aws/issues/44189))
* resource/aws_secretsmanager_secret: Return diagnostic `warning` when remote policy is invalid ([#44228](https://github.com/hashicorp/terraform-provider-aws/issues/44228))
* resource/aws_servicecatalog_provisioned_product: Restore `timeouts.read` arguments removed in v6.12.0 ([#44238](https://github.com/hashicorp/terraform-provider-aws/issues/44238))

## 6.12.0 (September 4, 2025)

NOTES:

* resource/aws_s3_bucket_acl: The `access_control_policy.grant.grantee.display_name` attribute is deprecated. AWS has [ended support for this attribute](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Grantee.html). API responses began inconsistently returning it on July 15, 2025, and will stop returning it entirely on November 21, 2025. This attribute will be removed in a future major version. ([#44090](https://github.com/hashicorp/terraform-provider-aws/issues/44090))
* resource/aws_s3_bucket_acl: The `access_control_policy.owner.display_name` attribute is deprecated. AWS has [ended support for this attribute](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Owner.html). API responses began inconsistently returning it on July 15, 2025, and will stop returning it entirely on November 21, 2025. This attribute will be removed in a future major version. ([#44090](https://github.com/hashicorp/terraform-provider-aws/issues/44090))
* resource/aws_s3_bucket_logging: The `target_grant.grantee.display_name` attribute is deprecated. AWS has [ended support for this attribute](https://docs.aws.amazon.com/AmazonS3/latest/API/API_Grantee.html). API responses began inconsistently returning it on July 15, 2025, and will stop returning it entirely on November 21, 2025. This attribute will be removed in a future major version. ([#44090](https://github.com/hashicorp/terraform-provider-aws/issues/44090))

FEATURES:

* **New Resource:** `aws_cognito_managed_login_branding` ([#43817](https://github.com/hashicorp/terraform-provider-aws/issues/43817))

ENHANCEMENTS:

* data-source/aws_efs_mount_target: Add `ip_address_type` and `ipv6_address` attributes ([#44079](https://github.com/hashicorp/terraform-provider-aws/issues/44079))
* data-source/aws_instance: Add `placement_group_id` attribute ([#38527](https://github.com/hashicorp/terraform-provider-aws/issues/38527))
* data-source/aws_lambda_function: Add `source_kms_key_arn` attribute ([#44080](https://github.com/hashicorp/terraform-provider-aws/issues/44080))
* data-source/aws_launch_template: Add `placement.group_id` attribute ([#44097](https://github.com/hashicorp/terraform-provider-aws/issues/44097))
* provider: Support `ap-southeast-6` as a valid AWS Region ([#44127](https://github.com/hashicorp/terraform-provider-aws/issues/44127))
* resource/aws_ecs_service: Remove Terraform default for `availability_zone_rebalancing` and change the attribute to Optional and Computed. This allow ECS to default to `ENABLED` for new resources compatible with *AvailabilityZoneRebalancing* and maintain an existing service's `availability_zone_rebalancing` value during update when not configured. If an existing service never had an `availability_zone_rebalancing` value configured and is updated, ECS will treat this as `DISABLED` ([#43241](https://github.com/hashicorp/terraform-provider-aws/issues/43241))
* resource/aws_efs_mount_target: Add `ip_address_type` and `ipv6_address` arguments to support IPv6 connectivity ([#44079](https://github.com/hashicorp/terraform-provider-aws/issues/44079))
* resource/aws_fsx_openzfs_file_system: Remove maximum items limit on the `user_and_group_quotas` argument ([#44120](https://github.com/hashicorp/terraform-provider-aws/issues/44120))
* resource/aws_fsx_openzfs_volume: Remove maximum items limit on the `user_and_group_quotas` argument ([#44118](https://github.com/hashicorp/terraform-provider-aws/issues/44118))
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
* resource/aws_synthetics_canary: Add `run_config.ephemeral_storage` argument. ([#44105](https://github.com/hashicorp/terraform-provider-aws/issues/44105))

BUG FIXES:

* resource/aws_s3tables_table_policy: Remove plan-time validation of `name` and `namespace` ([#44072](https://github.com/hashicorp/terraform-provider-aws/issues/44072))
* resource/aws_servicecatalog_provisioned_product: Set `provisioning_parameters` and `provisioning_artifact_id` to the values from the last successful deployment when update fails ([#43956](https://github.com/hashicorp/terraform-provider-aws/issues/43956))
* resource/aws_wafv2_web_acl: Fix performance of update when the WebACL has a large number of rules ([#42740](https://github.com/hashicorp/terraform-provider-aws/issues/42740))

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
