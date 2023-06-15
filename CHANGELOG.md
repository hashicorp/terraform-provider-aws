## 5.4.0 (Unreleased)

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
* data-source/aws_opensearch_domain: Add `off_peak_window_options` attribute ([#35970](https://github.com/hashicorp/terraform-provider-aws/issues/35970))
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
* resource/aws_opensearch_domain: Add `off_peak_window_options` configuration block ([#35970](https://github.com/hashicorp/terraform-provider-aws/issues/35970))
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
