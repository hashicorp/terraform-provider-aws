## 3.0.0 (Unreleased)

NOTES:
* provider: This version is built using Go 1.14.5, including security fixes to the crypto/x509 and net/http packages.

BREAKING CHANGES

* provider: New versions of the provider can only be automatically installed on Terraform 0.12 and later [GH-14143]
* provider: All "removed" attributes are cut, using them would result in a Terraform Core level error [GH-14001]
* provider: Credential ordering has changed from static, environment, shared credentials, EC2 metadata, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) to static, environment, shared credentials, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) [GH-14077]
* provider: The `AWS_METADATA_TIMEOUT` environment variable no longer has any effect as we now depend on the default AWS Go SDK EC2 Metadata client timeout of one second with two retries [GH-14077]
* provider: Remove deprecated `kinesis_analytics` and `r53` custom service endpoint arguments [GH-14238]
* data-source/aws_availability_zones: Remove deprecated `blacklisted_names` and `blacklisted_zone_ids` arguments [GH-14134]
* data-source/aws_directory_service_directory: Return an error when a single result is not found [GH-14006]
* data-source/aws_ecr_repository: Return an error when a single result is not found [GH-10520]
* data-source/aws_efs_file_system: Return an error when a single result is not found [GH-14005]
* data-source/aws_launch_template: Return an error when a single result is not found [GH-10521]
* data-source/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in data-source [GH-14220]
* data-source/aws_route53_zone: Trailing period removed from `name` argument set in data-source [GH-14220]
* resource/aws_acm_certificate: `certificate_body`, `certificate_chain`, and `private_key` attributes are no longer stored in the Terraform state with hash values [GH-9685]
* resource/aws_acm_certificate: `domain_validation_options` attribute changed from list to set [GH-14199]
* resource/aws_acm_certificate: Plan-time validation added to `domain_name` and `subject_alternative_names` arguments to prevent usage of strings with trailing periods [GH-14220]
* resource/aws_api_gateway_method_settings: Remove `Computed` property from `throttling_burst_limit` and `throttling_rate_limit` arguments, enabling drift detection [GH-14266]
* resource/aws_api_gateway_method_settings: Update `throttling_burst_limit` and `throttling_rate_limit` argument defaults to match API default of `-1` to keep throttling disabled [GH-14266]
* resource/aws_autoscaling_group: `availability_zones` and `vpc_zone_identifier` argument conflict now reported at plan-time [GH-12927]
* resource/aws_autoscaling_group: Remove `Computed` property from `load_balancers` and `target_group_arns` arguments, enabling drift detection [GH-14064]
* resource/aws_cloudfront_distribution: `active_trusted_signers` argument renamed to `trusted_signers` to support accessing `items` in Terraform 0.12 [GH-14339]
* resource/aws_cloudwatch_log_group: Automatically trim `:*` suffix from `arn` attribute [GH-14214]
* resource/aws_codepipeline: Removes `GITHUB_TOKEN` environment variable [GH-14175]
* resource/aws_cognito_user_pool: Remove deprecated `admin_create_user_config` configuration block `unused_account_validity_days` argument [GH-14294]
* resource/aws_dx_gateway: Remove automatic `aws_dx_gateway_association` resource import [GH-14124]
* resource/aws_dx_gateway_association: Remove deprecated `vpn_gateway_id` argument [GH-14144]
* resource/aws_dx_gateway_association_proposal: Remove deprecated `vpn_gateway_id` argument [GH-14144]
* resource/aws_ebs_volume: Return an error when `iops` argument set to a value greater than 0 for volume types other than `io1` [GH-14310]
* resource/aws_elastic_transcoder_preset: Remove `video` configuration block `max_frame_rate` argument default value [GH-7141]
* resource/aws_emr_cluster: Remove deprecated `instance_group` configuration block, `core_instance_count`, `core_instance_type`, and `master_instance_type` arguments [GH-14137]
* resource/aws_glue_job: Remove deprecated `allocated_capacity` argument [GH-14296]
* resource/aws_iam_access_key: Remove deprecated `ses_smtp_password` attribute [GH-14299]
* resource/aws_iam_instance_profile: Remove deprecated `roles` argument [GH-14303]
* resource/aws_iam_server_certificate: Remove state hashing from `certificate_body`, `certificate_chain`, and `private_key` arguments for new or recreated resources [GH-14187]
* resource/aws_instance: Return an error when `ebs_block_device` `iops` or `root_block_device` `iops` argument set to a value greater than `0` for volume types other than `io1` [GH-14310]
* resource/aws_lambda_alias: Resource import no longer converts Lambda Function name to ARN [GH-12876]
* resource/aws_launch_template: `network_interfaces` `delete_on_termination` argument changed from `bool` to `string` type [GH-8612]
* resource/aws_lb_listener_rule: Remove deprecated `condition` configuration block `field` and `values` arguments [GH-14309]
* resource/aws_msk_cluster: Update `encryption_info` `encryption_in_transit` `client_broker` argument default to match API default of `TLS` [GH-14132]
* resource/aws_rds_cluster: Update `scaling_configuration` `min_capacity` argument default to match API default of `1` [GH-14268]
* resource/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in resource [GH-14220]
* resource/aws_route53_zone: Trailing period removed from `name` argument set in resource [GH-14220]
* resource/aws_s3_bucket: Remove automatic `aws_s3_bucket_policy` resource import [GH-14121]
* resource/aws_s3_bucket: Convert `region` to read-only attribute [GH-14127]
* resource/aws_s3_bucket_metric: Update `filter` argument to require at least one of the `prefix` or `tags` nested arguments [GH-14230]
* resource/aws_security_group: Remove automatic `aws_security_group_rule` resource import [GH-12616]
* resource/aws_ses_domain_identity: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods [GH-14220]
* resource/aws_ses_domain_identity_verification: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods [GH-14220]
* resource/aws_sns_platform_application: `platform_credential` and `platform_principal` attributes are no longer stored in the Terraform state with hash values [GH-3894]
* resource/aws_spot_fleet_request: Remove 24 hour default for `valid_until` argument [GH-9718]
* resource/aws_ssm_maintenance_window_task: Remove deprecated `logging_info` and `task_parameters` configuration blocks [GH-14311]

FEATURES

* **New Data Source:** aws_workspaces_directory [GH-13529]

ENHANCEMENTS

* provider: Always enable shared configuration file support (no longer require `AWS_SDK_LOAD_CONFIG` environment variable) [GH-14077]
* provider: Add `assume_role` configuration block `duration_seconds`, `policy_arns`, `tags`, and `transitive_tag_keys` arguments [GH-14077]
* data-source/aws_instance: Add `secondary_private_ips` attribute [GH-14079]
* data-source/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support [GH-14221]
* resource/aws_acm_certificate: Enable `domain_validation_options` usage in downstream resource `count` and `for_each` references [GH-14199]
* resource/aws_api_gateway_authorizer: Add plan-time validation to `authorizer_credentials` argument [GH-12643]
* resource/aws_api_gateway_method_settings: Add import support [GH-14266]
* resource/aws_apigatewayv2_integration: Add `request_parameters` attribute [GH-14080]
* resource/aws_apigatewayv2_integration: Add `tls_config` attribute [GH-13013]
* resource/aws_apigatewayv2_route: Support for updating route key [GH-13833]
* resource/aws_apigatewayv2_stage: Make `deployment_id` a `Computed` attribute [GH-13644]
* resource/aws_fsx_lustre_file_system: Add `deployment_type` and `per_unit_storage_throughput` attributes [GH-13639]
* resource_aws_fsx_windows_file_system - add `storage_type` argument. [GH-14316]
* resource_aws_fsx_windows_file_system: add support for multi-az [GH-12676]
* resource_aws_fsx_windows_file_system: add `SINGLE_AZ_2` deployment type [GH-12676]
* resource_aws_fsx_windows_file_system: adds `preferred_file_server_ip`, `remote_administration_endpoint` attributes [GH-12676]
* resource/aws_instance: Add `secondary_private_ips` argument (conflicts with `network_interface` configuration block) [GH-14079]

BUG FIXES

* provider: Ensure nil is not passed to RetryError helpers, may result in some bug fixes [GH-14104]
* provider: Ensure configured STS endpoint is used during `AssumeRole` API calls [GH-14077]
* provider: Prefer AWS shared configuration over EC2 metadata credentials by default [GH-14077]
* provider: Prefer CodeBuild, ECS, EKS credentials over EC2 metadata credentials by default [GH-14077]
* data-source/aws_lb: `enable_http2` now properly set [GH-14167]
* resource/aws_acm_certificate: Prevent unexpected ordering differences with `domain_validation_options` attribute [GH-14199]
* resource/aws_api_gateway_authorizer: Allow `authorizer_result_ttl_in_seconds` to be set to 0 [GH-12643]
* resource/aws_apigatewayv2_integration: Correctly handle the `integration_method` attribute for AWS Lambda integrations[GH-13266]
* resource/aws_apigatewayv2_integration: Correctly handle the `passthrough_behavior` attribute for HTTP APIs [GH-13062]
* resource/aws_apigatewayv2_stage: Correctly handle `default_route_setting` and `route_setting` `data_trace_enabled` and `logging_level` for HTTP APIs. `logging_level` is now `Computed`, meaning Terraform will only perform drift detection of its value when present in a configuration. [GH-13809]
* resource/aws_appautoscaling_target: Only retry `DeregisterScalableTarget` retries on all errors on deletion [GH-14259]
* resource/aws_dx_gateway_association: Increase default create/update/delete timeouts to 30 minutes [GH-14144]
* resource/aws_codepipeline: Only retry `CreatePipeline` errors for IAM eventual consistency errors [GH-14264]
* resource/aws_elasticsearch_domain: Update method to properly set `advanced_security_options` [GH-14167]
* resource/aws_lambda_function: Increase IAM retry timeout for creation to standard 2 minute timeout [GH-14291]
* resource/aws_lb_cookie_stickiness_policy: `lb_port` now properly set [GH-14167]
* resource/aws_network_acl_rule: Immediately return `DescribeNetworkAcls` errors on creation [GH-14261]
* resource/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support [GH-14221]
* resource/aws_sns_topic_subscription: Immediately return `ListSubscriptionsByTopic` errors [GH-14262]
* resource/aws_spot_fleet_request: Only retry `RequestSpotFleet` on IAM eventual consistency errors and use standard 2 minute timeout [GH-14265]
* resource/aws_spot_instance_request: `primary_network_interface_id` now properly set [GH-14167]
* resource/aws_ssm_activation: Only retry `CreateActivation` on IAM eventual consistency errors and use standard 2 minute timeout [GH-14263]
* resource/aws_ssm_association: `parameters` now properly set [GH-14167]

## Previous Releases

For information on prior major releases, see their changelogs:

* [2.x and earlier](https://github.com/terraform-providers/terraform-provider-aws/blob/release/2.x/CHANGELOG.md)
