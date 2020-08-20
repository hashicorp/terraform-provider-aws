## 3.3.0 (August 20, 2020)

ENHANCEMENTS

* data-source/aws_lambda_layer_version: Support `java8.al2` and `provided.al2` in `runtime` argument plan-time validation ([#14663](https://github.com/terraform-providers/terraform-provider-aws/issues/14663))
* provider: Support for appending information to User-Agent request headers with the `TF_APPEND_USER_AGENT` environment variable ([#14555](https://github.com/terraform-providers/terraform-provider-aws/issues/14555))
* resource/aws_apigatewayv2_api: Add `body` argument ([#12567](https://github.com/terraform-providers/terraform-provider-aws/issues/12567))
* resource/aws_customer_gateway: Support tag on create ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_dms_replication_instance: Add `allow_major_version_upgrade` argument ([#14550](https://github.com/terraform-providers/terraform-provider-aws/issues/14550))
* resource/aws_ec2_client_vpn_network_association: Allow specifying custom security groups ([#14146](https://github.com/terraform-providers/terraform-provider-aws/issues/14146))
* resource/aws_ec2_client_vpn_network_association: Support resource import ([#14146](https://github.com/terraform-providers/terraform-provider-aws/issues/14146))
* resource/aws_egress_only_intrenet_gateway:-Ssupport tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_eks_node_group: Support `AL2_ARM_64` value for `ami_type` argument plan-time validation ([#14729](https://github.com/terraform-providers/terraform-provider-aws/issues/14729))
* resource/aws_eks_node_group: Add `launch_template` configuration block ([#14639](https://github.com/terraform-providers/terraform-provider-aws/issues/14639))
* resource/aws_internet_gateway: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_lambda_function: Support `java8.al2` and `provided.al2` in `runtime` argument plan-time validation ([#14663](https://github.com/terraform-providers/terraform-provider-aws/issues/14663))
* resource/aws_lambda_layer_version: Support `java8.al2` and `provided.al2` in `compatible_runtimes` argument plan-time validation ([#14663](https://github.com/terraform-providers/terraform-provider-aws/issues/14663))
* resource/aws_launch_template: Support `elastic-gpu` and `spot-instances-request` in `tag_specifications` `resource_type` argument plan-time validation ([#14662](https://github.com/terraform-providers/terraform-provider-aws/issues/14662))
* resource/aws_network_acl: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_network_interface: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_route_table: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_security_group: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_spot_instance_request: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_storagegatway_smb_file_share: Add `audit_destination_arn` and `smb_acl_enabled` arguments ([#13572](https://github.com/terraform-providers/terraform-provider-aws/issues/13572))
* resource/aws_subnet: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_subnet: Add plan-time validation to `ipv6_cidr_block` argument ([#12303](https://github.com/terraform-providers/terraform-provider-aws/issues/12303))
* resource/aws_vpc_dhcp_options: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_vpc_peering_connection: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_vpn_connection: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_vpn_gateway: Support tag on create  ([#14501](https://github.com/terraform-providers/terraform-provider-aws/issues/14501))
* resource/aws_wafv2_rule_group: Add `forwarded_ip_config` configuration block to `geo_match_statement` ([#14685](https://github.com/terraform-providers/terraform-provider-aws/issues/14685))
* resource/aws_wafv2_web_acl: Add `forwarded_ip_config` configuration block to `rate_based_statement` and `geo_match_statement` ([#14685](https://github.com/terraform-providers/terraform-provider-aws/issues/14685))
* resource/aws_wafv2_web_acl: Support `FORWARDED_IP` value for `rate_based_statement` `aggregate_key_type` argument plan-time validation ([#14685](https://github.com/terraform-providers/terraform-provider-aws/issues/14685))

BUG FIXES

* resource/aws_api_gateway_vpc_link: Increase create, update, and delete timeouts to 20 minutes ([#10407](https://github.com/terraform-providers/terraform-provider-aws/issues/10407))
* resource/aws_apigatewayv2_stage: Set `execution_arn` attribute for HTTP APIs ([#14638](https://github.com/terraform-providers/terraform-provider-aws/issues/14638))
* resource/aws_db_parameter_group: Restore ability to update `parameter` configuration values ([#12112](https://github.com/terraform-providers/terraform-provider-aws/issues/12112))
* resource/aws_user_pool_domain: Ensure state removal when deleted outside Terraform ([#14732](https://github.com/terraform-providers/terraform-provider-aws/issues/14732))
* resource/aws_rds_cluster_parameter_group: Restore ability to update `parameter` configuration values ([#12112](https://github.com/terraform-providers/terraform-provider-aws/issues/12112))
* resource/aws_ssm_parameter: Handle retries after creation for asynchronous `data_type` validation process ([#14514](https://github.com/terraform-providers/terraform-provider-aws/issues/14514))
* resource/aws_storagegateway_nfs_file_share: Skip `UpdateSMBFileShare` API call when only `tags` change and remove extraneous `ListTagsForResource` API call during read ([#13590](https://github.com/terraform-providers/terraform-provider-aws/issues/13590))
* resource/aws_subnet: Ensure `ipv6_cidr_block` argument performs removal when removed from configuration ([#12303](https://github.com/terraform-providers/terraform-provider-aws/issues/12303))

## 3.2.0 (August 14, 2020)

ENHANCEMENTS

* data-source/aws_launch_configuration: Add `ebs_block_device` `no_device` attribute ([#14583](https://github.com/terraform-providers/terraform-provider-aws/issues/14583))
* data-source/aws_lb: Add `subnet_mapping` `private_ipv4_address` attribute ([#14545](https://github.com/terraform-providers/terraform-provider-aws/issues/14545))
* provider: Upgrade to Terraform Plugin SDK V2. There should be no breaking changes from a practitioner's perspective. Some validation errors should now feature enhanced messaging. ([#14432](https://github.com/terraform-providers/terraform-provider-aws/issues/14432))
* resource/aws_accessanalyzer_analyzer: Support `ORGANIZATION` value in `type` argument ([#14493](https://github.com/terraform-providers/terraform-provider-aws/issues/14493))
* resource/aws_codebuild_project: Support `WINDOWS_SERVER_2019_CONTAINER` value in `environment` `type` argument plan-time validation ([#14532](https://github.com/terraform-providers/terraform-provider-aws/issues/14532))
* resource/aws_organizations_organization: Support `AISERVICES_OPT_OUT_POLICY` value in `enabled_policy_types` argument plan-time validation (Support AI Opt Out policies) ([#14650](https://github.com/terraform-providers/terraform-provider-aws/issues/14650))
* resource/aws_organizations_policy: Support `AISERVICES_OPT_OUT_POLICY` value in `type` argument plan-time validation (Support AI Opt Out policies) ([#14528](https://github.com/terraform-providers/terraform-provider-aws/issues/14528))
* resource/aws_route53_health_check: Add `disabled` argument ([#14614](https://github.com/terraform-providers/terraform-provider-aws/issues/14614))

BUG FIXES

* data-source/aws_launch_template: Prevent type error with `network_interfaces` `delete_on_termination` attribute ([#14599](https://github.com/terraform-providers/terraform-provider-aws/issues/14599))
* resource/aws_acm_certificate_validation: Prevent panic with missing `DomainValidationOptions` `ResourceRecord` attribute in API response [[#14590](https://github.com/terraform-providers/terraform-provider-aws/issues/14590)] 
* resource/aws_ecr_repository: Prevent panic with missing `EncryptionConfiguration` attribute in API response ([#14584](https://github.com/terraform-providers/terraform-provider-aws/issues/14584))
* resource/aws_wafv2_rule_group: Prevent unnecessary resource recreation with `rule` updates ([#14617](https://github.com/terraform-providers/terraform-provider-aws/issues/14617))
* resource/aws_wafv2_web_acl: Prevent unnecessary resource recreation with `rule` updates ([#14616](https://github.com/terraform-providers/terraform-provider-aws/issues/14616))

## 3.1.0 (August 07, 2020)

NOTES:

* resource/aws_route53_zone_association: The addition of cross-account zone association support required the use of new `ListHostedZonesByVPC` API call and adding the VPC Region to the resource ID for new resources. Restrictive IAM permissions for Terraform and cross-region imports may require updates. ([#14215](https://github.com/terraform-providers/terraform-provider-aws/issues/14215))

FEATURES

* **New Data Source:** `aws_ec2_spot_price` ([#12504](https://github.com/terraform-providers/terraform-provider-aws/issues/12504))
* **New Resource**: `aws_route53_vpc_association_authorization` ([#14215](https://github.com/terraform-providers/terraform-provider-aws/issues/14215))

ENHANCEMENTS

* data-source/aws_ecr_repository: Allow `registry_id` as an argument ([#14368](https://github.com/terraform-providers/terraform-provider-aws/issues/14368))
* data-source/aws_ecr_repository: Add `image_scanning_configuration` and `image_tag_mutability` attributes ([#14368](https://github.com/terraform-providers/terraform-provider-aws/issues/14368))
* data-source/aws_ecr_repository: Add `encryption_configuration` attribute ([#14520](https://github.com/terraform-providers/terraform-provider-aws/issues/14520))
* resource/aws_api_gateway_method_settings: Plan-time validation added to `settings` `unauthorized_cache_control_header_strategy` and `logging_level` arguments ([#12651](https://github.com/terraform-providers/terraform-provider-aws/issues/12651))
* resource/aws_ecr_repository: Add `encryption_configuration` attribute ([#14520](https://github.com/terraform-providers/terraform-provider-aws/issues/14520))
* resource/aws_lb: Add `subnet_mapping` configuration block `private_ipv4_address` argument ([#11404](https://github.com/terraform-providers/terraform-provider-aws/issues/11404))
* resource/aws_rds_global_cluster: Add `force_destroy` and `source_db_cluster_identifier` arguments ([#14487](https://github.com/terraform-providers/terraform-provider-aws/issues/14487))
* resource/aws_rds_global_cluster: Add `global_cluster_members` attribute ([#14487](https://github.com/terraform-providers/terraform-provider-aws/issues/14487))
* resource/aws_route53_zone_association: Cross-account zone associations can now be created in conjunction with the new `aws_route53_vpc_association_authorization` resource ([#14215](https://github.com/terraform-providers/terraform-provider-aws/issues/14215))
* resource/aws_ssm_parameter: Add `data_type` argument (support `aws:ec2:image` parameters) ([#13326](https://github.com/terraform-providers/terraform-provider-aws/issues/13326))

BUG FIXES

* data-source/aws_availability_zones: Prevent unexpected plan output every apply with `group_names` attribute ([#14412](https://github.com/terraform-providers/terraform-provider-aws/issues/14412))
* data-source/aws_s3_bucket: Ensure provider `s3_force_path_style` configuration is passed through for getting S3 Bucket location with non-AWS implementations ([#14481](https://github.com/terraform-providers/terraform-provider-aws/issues/14481))
* resource/aws_api_gateway_method_settings: Allow `settings` `cache_ttl_in_seconds` argument to be set to 0 ([#12651](https://github.com/terraform-providers/terraform-provider-aws/issues/12651))
* resource/aws_elastictranscoder_preset: Prevent empty configuration block panics ([#14092](https://github.com/terraform-providers/terraform-provider-aws/issues/14092))
* resource/aws_lambda_event_source_mapping: Allow `maximum_retry_attempts` argument to be set to 0 ([#12479](https://github.com/terraform-providers/terraform-provider-aws/issues/12479))
* resource/aws_rds_cluster: Add an `InvalidDBClusterStateFault` retryable error condition for clusters part of a global cluster ([#14420](https://github.com/terraform-providers/terraform-provider-aws/issues/14420))
* resource/aws_rds_cluster: Increase retry timeout for deletion to 2 minutes ([#14420](https://github.com/terraform-providers/terraform-provider-aws/issues/14420))
* resource/aws_rds_cluster: Prevent error when both `global_cluster_identifier` and `replication_source_identifier` are configured on creation ([#14490](https://github.com/terraform-providers/terraform-provider-aws/issues/14490))
* resource/aws_s3_bucket: Ensure provider `s3_force_path_style` configuration is passed through for getting S3 Bucket location with non-AWS implementations ([#14481](https://github.com/terraform-providers/terraform-provider-aws/issues/14481))
* resource/aws_secretsmanager_secret: Allow retries for IAM eventual consistency errors ([#14459](https://github.com/terraform-providers/terraform-provider-aws/issues/14459))
* resource/aws_security_group: Ensure `name_prefix` argument with hex digits `a` through `f` is properly imported ([#14475](https://github.com/terraform-providers/terraform-provider-aws/issues/14475))
* resource/aws_spot_fleet_request: Allow `target_capacity` argument to be updated to 0 ([#12759](https://github.com/terraform-providers/terraform-provider-aws/issues/12759))
* resource/aws_spot_fleet_request: Wait for modify operation completion (default timeout of 10 minutes) ([#12759](https://github.com/terraform-providers/terraform-provider-aws/issues/12759))
* resource/aws_vpc_dhcp_options_association: Properly trigger resource recreation when VPC is deleted outside Terraform ([#14367](https://github.com/terraform-providers/terraform-provider-aws/issues/14367))

## 3.0.0 (July 31, 2020)

NOTES:
* provider: This version is built using Go 1.14.5, including security fixes to the crypto/x509 and net/http packages.

BREAKING CHANGES

* provider: New versions of the provider can only be automatically installed on Terraform 0.12 and later ([#14143](https://github.com/terraform-providers/terraform-provider-aws/issues/14143))
* provider: All "removed" attributes are cut, using them would result in a Terraform Core level error ([#14001](https://github.com/terraform-providers/terraform-provider-aws/issues/14001))
* provider: Credential ordering has changed from static, environment, shared credentials, EC2 metadata, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) to static, environment, shared credentials, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: The `AWS_METADATA_TIMEOUT` environment variable no longer has any effect as we now depend on the default AWS Go SDK EC2 Metadata client timeout of one second with two retries ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: Remove deprecated `kinesis_analytics` and `r53` custom service endpoint arguments ([#14238](https://github.com/terraform-providers/terraform-provider-aws/issues/14238))
* data-source/aws_availability_zones: Remove deprecated `blacklisted_names` and `blacklisted_zone_ids` arguments ([#14134](https://github.com/terraform-providers/terraform-provider-aws/issues/14134))
* data-source/aws_directory_service_directory: Return an error when a single result is not found ([#14006](https://github.com/terraform-providers/terraform-provider-aws/issues/14006))
* data-source/aws_ecr_repository: Return an error when a single result is not found ([#10520](https://github.com/terraform-providers/terraform-provider-aws/issues/10520))
* data-source/aws_efs_file_system: Return an error when a single result is not found ([#14005](https://github.com/terraform-providers/terraform-provider-aws/issues/14005))
* data-source/aws_launch_template: Return an error when a single result is not found ([#10521](https://github.com/terraform-providers/terraform-provider-aws/issues/10521))
* data-source/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in data-source ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* data-source/aws_route53_zone: Trailing period removed from `name` argument set in data-source ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_acm_certificate: `certificate_body`, `certificate_chain`, and `private_key` attributes are no longer stored in the Terraform state with hash values ([#9685](https://github.com/terraform-providers/terraform-provider-aws/issues/9685))
* resource/aws_acm_certificate: `domain_validation_options` attribute changed from list to set ([#14199](https://github.com/terraform-providers/terraform-provider-aws/issues/14199))
* resource/aws_acm_certificate: Plan-time validation added to `domain_name` and `subject_alternative_names` arguments to prevent usage of strings with trailing periods ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_api_gateway_method_settings: Remove `Computed` property from `throttling_burst_limit` and `throttling_rate_limit` arguments, enabling drift detection ([#14266](https://github.com/terraform-providers/terraform-provider-aws/issues/14266))
* resource/aws_api_gateway_method_settings: Update `throttling_burst_limit` and `throttling_rate_limit` argument defaults to match API default of `-1` to keep throttling disabled ([#14266](https://github.com/terraform-providers/terraform-provider-aws/issues/14266))
* resource/aws_autoscaling_group: `availability_zones` and `vpc_zone_identifier` argument conflict now reported at plan-time ([#12927](https://github.com/terraform-providers/terraform-provider-aws/issues/12927))
* resource/aws_autoscaling_group: Remove `Computed` property from `load_balancers` and `target_group_arns` arguments, enabling drift detection ([#14064](https://github.com/terraform-providers/terraform-provider-aws/issues/14064))
* resource/aws_cloudfront_distribution: `active_trusted_signers` argument renamed to `trusted_signers` to support accessing `items` in Terraform 0.12 ([#14339](https://github.com/terraform-providers/terraform-provider-aws/issues/14339))
* resource/aws_cloudwatch_log_group: Automatically trim `:*` suffix from `arn` attribute ([#14214](https://github.com/terraform-providers/terraform-provider-aws/issues/14214))
* resource/aws_codepipeline: Removes `GITHUB_TOKEN` environment variable ([#14175](https://github.com/terraform-providers/terraform-provider-aws/issues/14175))
* resource/aws_cognito_user_pool: Remove deprecated `admin_create_user_config` configuration block `unused_account_validity_days` argument ([#14294](https://github.com/terraform-providers/terraform-provider-aws/issues/14294))
* resource/aws_dx_gateway: Remove automatic `aws_dx_gateway_association` resource import ([#14124](https://github.com/terraform-providers/terraform-provider-aws/issues/14124))
* resource/aws_dx_gateway_association: Remove deprecated `vpn_gateway_id` argument ([#14144](https://github.com/terraform-providers/terraform-provider-aws/issues/14144))
* resource/aws_dx_gateway_association_proposal: Remove deprecated `vpn_gateway_id` argument ([#14144](https://github.com/terraform-providers/terraform-provider-aws/issues/14144))
* resource/aws_ebs_volume: Return an error when `iops` argument set to a value greater than 0 for volume types other than `io1` ([#14310](https://github.com/terraform-providers/terraform-provider-aws/issues/14310))
* resource/aws_elastic_transcoder_preset: Remove `video` configuration block `max_frame_rate` argument default value ([#7141](https://github.com/terraform-providers/terraform-provider-aws/issues/7141))
* resource/aws_emr_cluster: Remove deprecated `instance_group` configuration block, `core_instance_count`, `core_instance_type`, and `master_instance_type` arguments ([#14137](https://github.com/terraform-providers/terraform-provider-aws/issues/14137))
* resource/aws_glue_job: Remove deprecated `allocated_capacity` argument ([#14296](https://github.com/terraform-providers/terraform-provider-aws/issues/14296))
* resource/aws_iam_access_key: Remove deprecated `ses_smtp_password` attribute ([#14299](https://github.com/terraform-providers/terraform-provider-aws/issues/14299))
* resource/aws_iam_instance_profile: Remove deprecated `roles` argument ([#14303](https://github.com/terraform-providers/terraform-provider-aws/issues/14303))
* resource/aws_iam_server_certificate: Remove state hashing from `certificate_body`, `certificate_chain`, and `private_key` arguments for new or recreated resources ([#14187](https://github.com/terraform-providers/terraform-provider-aws/issues/14187))
* resource/aws_instance: Return an error when `ebs_block_device` `iops` or `root_block_device` `iops` argument set to a value greater than `0` for volume types other than `io1` ([#14310](https://github.com/terraform-providers/terraform-provider-aws/issues/14310))
* resource/aws_lambda_alias: Resource import no longer converts Lambda Function name to ARN ([#12876](https://github.com/terraform-providers/terraform-provider-aws/issues/12876))
* resource/aws_launch_template: `network_interfaces` `delete_on_termination` argument changed from `bool` to `string` type ([#8612](https://github.com/terraform-providers/terraform-provider-aws/issues/8612))
* resource/aws_lb_listener_rule: Remove deprecated `condition` configuration block `field` and `values` arguments ([#14309](https://github.com/terraform-providers/terraform-provider-aws/issues/14309))
* resource/aws_msk_cluster: Update `encryption_info` `encryption_in_transit` `client_broker` argument default to match API default of `TLS` ([#14132](https://github.com/terraform-providers/terraform-provider-aws/issues/14132))
* resource/aws_rds_cluster: Update `scaling_configuration` `min_capacity` argument default to match API default of `1` ([#14268](https://github.com/terraform-providers/terraform-provider-aws/issues/14268))
* resource/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in resource ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_route53_zone: Trailing period removed from `name` argument set in resource ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_s3_bucket: Remove automatic `aws_s3_bucket_policy` resource import ([#14121](https://github.com/terraform-providers/terraform-provider-aws/issues/14121))
* resource/aws_s3_bucket: Convert `region` to read-only attribute ([#14127](https://github.com/terraform-providers/terraform-provider-aws/issues/14127))
* resource/aws_s3_bucket_metric: Update `filter` argument to require at least one of the `prefix` or `tags` nested arguments ([#14230](https://github.com/terraform-providers/terraform-provider-aws/issues/14230))
* resource/aws_security_group: Remove automatic `aws_security_group_rule` resource import ([#12616](https://github.com/terraform-providers/terraform-provider-aws/issues/12616))
* resource/aws_ses_domain_identity: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_ses_domain_identity_verification: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods ([#14220](https://github.com/terraform-providers/terraform-provider-aws/issues/14220))
* resource/aws_sns_platform_application: `platform_credential` and `platform_principal` attributes are no longer stored in the Terraform state with hash values ([#3894](https://github.com/terraform-providers/terraform-provider-aws/issues/3894))
* resource/aws_spot_fleet_request: Remove 24 hour default for `valid_until` argument ([#9718](https://github.com/terraform-providers/terraform-provider-aws/issues/9718))
* resource/aws_ssm_maintenance_window_task: Remove deprecated `logging_info` and `task_parameters` configuration blocks ([#14311](https://github.com/terraform-providers/terraform-provider-aws/issues/14311))

FEATURES

* **New Data Source:** aws_workspaces_directory ([#13529](https://github.com/terraform-providers/terraform-provider-aws/issues/13529))

ENHANCEMENTS

* provider: Always enable shared configuration file support (no longer require `AWS_SDK_LOAD_CONFIG` environment variable) ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: Add `assume_role` configuration block `duration_seconds`, `policy_arns`, `tags`, and `transitive_tag_keys` arguments ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* data-source/aws_instance: Add `secondary_private_ips` attribute ([#14079](https://github.com/terraform-providers/terraform-provider-aws/issues/14079))
* data-source/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support ([#14221](https://github.com/terraform-providers/terraform-provider-aws/issues/14221))
* resource/aws_acm_certificate: Enable `domain_validation_options` usage in downstream resource `count` and `for_each` references ([#14199](https://github.com/terraform-providers/terraform-provider-aws/issues/14199))
* resource/aws_api_gateway_authorizer: Add plan-time validation to `authorizer_credentials` argument ([#12643](https://github.com/terraform-providers/terraform-provider-aws/issues/12643))
* resource/aws_api_gateway_method_settings: Add import support ([#14266](https://github.com/terraform-providers/terraform-provider-aws/issues/14266))
* resource/aws_apigatewayv2_integration: Add `request_parameters` attribute ([#14080](https://github.com/terraform-providers/terraform-provider-aws/issues/14080))
* resource/aws_apigatewayv2_integration: Add `tls_config` attribute ([#13013](https://github.com/terraform-providers/terraform-provider-aws/issues/13013))
* resource/aws_apigatewayv2_route: Support for updating route key ([#13833](https://github.com/terraform-providers/terraform-provider-aws/issues/13833))
* resource/aws_apigatewayv2_stage: Make `deployment_id` a `Computed` attribute ([#13644](https://github.com/terraform-providers/terraform-provider-aws/issues/13644))
* resource/aws_fsx_lustre_file_system: Add `deployment_type` and `per_unit_storage_throughput` attributes ([#13639](https://github.com/terraform-providers/terraform-provider-aws/issues/13639))
* resource_aws_fsx_windows_file_system - add `storage_type` argument. ([#14316](https://github.com/terraform-providers/terraform-provider-aws/issues/14316))
* resource_aws_fsx_windows_file_system: add support for multi-az ([#12676](https://github.com/terraform-providers/terraform-provider-aws/issues/12676))
* resource_aws_fsx_windows_file_system: add `SINGLE_AZ_2` deployment type ([#12676](https://github.com/terraform-providers/terraform-provider-aws/issues/12676))
* resource_aws_fsx_windows_file_system: adds `preferred_file_server_ip`, `remote_administration_endpoint` attributes ([#12676](https://github.com/terraform-providers/terraform-provider-aws/issues/12676))
* resource/aws_instance: Add `secondary_private_ips` argument (conflicts with `network_interface` configuration block) ([#14079](https://github.com/terraform-providers/terraform-provider-aws/issues/14079))

BUG FIXES

* provider: Ensure nil is not passed to RetryError helpers, may result in some bug fixes ([#14104](https://github.com/terraform-providers/terraform-provider-aws/issues/14104))
* provider: Ensure configured STS endpoint is used during `AssumeRole` API calls ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: Prefer AWS shared configuration over EC2 metadata credentials by default ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* provider: Prefer CodeBuild, ECS, EKS credentials over EC2 metadata credentials by default ([#14077](https://github.com/terraform-providers/terraform-provider-aws/issues/14077))
* data-source/aws_lb: `enable_http2` now properly set ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))
* resource/aws_acm_certificate: Prevent unexpected ordering differences with `domain_validation_options` attribute ([#14199](https://github.com/terraform-providers/terraform-provider-aws/issues/14199))
* resource/aws_api_gateway_authorizer: Allow `authorizer_result_ttl_in_seconds` to be set to 0 ([#12643](https://github.com/terraform-providers/terraform-provider-aws/issues/12643))
* resource/aws_apigatewayv2_integration: Correctly handle the `integration_method` attribute for AWS Lambda integrations([#13266](https://github.com/terraform-providers/terraform-provider-aws/issues/13266))
* resource/aws_apigatewayv2_integration: Correctly handle the `passthrough_behavior` attribute for HTTP APIs ([#13062](https://github.com/terraform-providers/terraform-provider-aws/issues/13062))
* resource/aws_apigatewayv2_stage: Correctly handle `default_route_setting` and `route_setting` `data_trace_enabled` and `logging_level` for HTTP APIs. `logging_level` is now `Computed`, meaning Terraform will only perform drift detection of its value when present in a configuration. ([#13809](https://github.com/terraform-providers/terraform-provider-aws/issues/13809))
* resource/aws_appautoscaling_target: Only retry `DeregisterScalableTarget` retries on all errors on deletion ([#14259](https://github.com/terraform-providers/terraform-provider-aws/issues/14259))
* resource/aws_dx_gateway_association: Increase default create/update/delete timeouts to 30 minutes ([#14144](https://github.com/terraform-providers/terraform-provider-aws/issues/14144))
* resource/aws_codepipeline: Only retry `CreatePipeline` errors for IAM eventual consistency errors ([#14264](https://github.com/terraform-providers/terraform-provider-aws/issues/14264))
* resource/aws_elasticsearch_domain: Update method to properly set `advanced_security_options` ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))
* resource/aws_lambda_function: Increase IAM retry timeout for creation to standard 2 minute timeout ([#14291](https://github.com/terraform-providers/terraform-provider-aws/issues/14291))
* resource/aws_lb_cookie_stickiness_policy: `lb_port` now properly set ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))
* resource/aws_network_acl_rule: Immediately return `DescribeNetworkAcls` errors on creation ([#14261](https://github.com/terraform-providers/terraform-provider-aws/issues/14261))
* resource/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support ([#14221](https://github.com/terraform-providers/terraform-provider-aws/issues/14221))
* resource/aws_sns_topic_subscription: Immediately return `ListSubscriptionsByTopic` errors ([#14262](https://github.com/terraform-providers/terraform-provider-aws/issues/14262))
* resource/aws_spot_fleet_request: Only retry `RequestSpotFleet` on IAM eventual consistency errors and use standard 2 minute timeout ([#14265](https://github.com/terraform-providers/terraform-provider-aws/issues/14265))
* resource/aws_spot_instance_request: `primary_network_interface_id` now properly set ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))
* resource/aws_ssm_activation: Only retry `CreateActivation` on IAM eventual consistency errors and use standard 2 minute timeout ([#14263](https://github.com/terraform-providers/terraform-provider-aws/issues/14263))
* resource/aws_ssm_association: `parameters` now properly set ([#14167](https://github.com/terraform-providers/terraform-provider-aws/issues/14167))

## Previous Releases

For information on prior major releases, see their changelogs:

* [2.x and earlier](https://github.com/terraform-providers/terraform-provider-aws/blob/release/2.x/CHANGELOG.md)
