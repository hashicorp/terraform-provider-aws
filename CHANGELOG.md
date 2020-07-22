## 3.0.0 (Unreleased)

BREAKING CHANGES

* provider: New versions of the provider can only be automatically installed on Terraform 0.12 and later [GH-14143]
* provider: All "removed" attributes are cut, using them would result in a Terraform Core level error [GH-14001]
* provider: Credential ordering has changed from static, environment, shared credentials, EC2 metadata, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) to static, environment, shared credentials, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) [GH-14077]
* provider: The `AWS_METADATA_TIMEOUT` environment variable no longer has any effect as we now depend on the default AWS Go SDK EC2 Metadata client timeout of one second with two retries [GH-14077]
* data-source/aws_availability_zones: Remove deprecated `blacklisted_names` and `blacklisted_zone_ids` arguments [GH-14134]
* data-source/aws_directory_service_directory: Return an error when a single result is not found [GH-14006]
* data-source/aws_ecr_repository: Return an error when a single result is not found [GH-10520]
* data-source/aws_efs_file_system: Return an error when a single result is not found [GH-14005]
* data-source/aws_launch_template: Return an error when a single result is not found [GH-10521]
* resource/aws_acm_certificate: `certificate_body`, `certificate_chain`, and `private_key` attributes are no longer stored in the Terraform state with hash values [GH-9685]
* resource/aws_api_gateway_method_settings: Remove `Computed` property from `throttling_burst_limit` and `throttling_rate_limit` arguments, enabling drift detection [GH-14266]
* resource/aws_api_gateway_method_settings: Update `throttling_burst_limit` and `throttling_rate_limit` argument defaults to match API default of `-1` to keep throttling disabled [GH-14266]
* resource/aws_apigatewayv2_integration: Correctly handle the `passthrough_behavior` attribute for HTTP APIs [GH-13062]
* resource/aws_autoscaling_group: `availability_zones` and `vpc_zone_identifier` argument conflict now reported at plan-time [GH-12927]
* resource/aws_autoscaling_group: Remove `Computed` property from `load_balancers` and `target_group_arns` arguments, enabling drift detection [GH-14064]
* resource/aws_dx_gateway: Remove automatic `aws_dx_gateway_association` resource import [GH-14124]
* resource/aws_elastic_transcoder_preset: Remove `video` configuration block `max_frame_rate` argument default value [GH-7141]
* resource/aws_emr_cluster: Remove deprecated `instance_group` configuration block, `core_instance_count`, `core_instance_type`, and `master_instance_type` arguments [GH-14137]
* resource_aws_fsx_windows_file_system: add support for multi-az [GH-12676]
* resource_aws_fsx_windows_file_system: add `SINGLE_AZ_2` deployment type [GH-12676]
* resource_aws_fsx_windows_file_system: adds `preferred_file_server_ip`, `remote_administration_endpoint` attributes [GH-12676]
* resource/aws_lambda_alias: Resource import no longer converts Lambda Function name to ARN [GH-12876]
* resource/aws_launch_template: `network_interfaces` `delete_on_termination` argument changed from `bool` to `string` type [GH-8612]
* resource/aws_msk_cluster: Update `encryption_info` `encryption_in_transit` `client_broker` argument default to match API default of `TLS` [GH-14132]
* resource/aws_rds_cluster: Update `scaling_configuration` `min_capacity` argument default to match API default of `1` [GH-14268]
* resource/aws_s3_bucket: Remove automatic `aws_s3_bucket_policy` resource import [GH-14121]
* resource/aws_s3_bucket: Convert `region` to read-only attribute [GH-14127]
* resource/aws_security_group: Remove automatic `aws_security_group_rule` resource import [GH-12616]
* resource/aws_sns_platform_application: `platform_credential` and `platform_principal` attributes are no longer stored in the Terraform state with hash values [GH-3894]
* resource/aws_spot_fleet_request: Remove 24 hour default for `valid_until` argument [GH-9718]

FEATURES

* **New Data Source:** aws_workspaces_directory [GH-13529]

ENHANCEMENTS

* provider: Always enable shared configuration file support (no longer require `AWS_SDK_LOAD_CONFIG` environment variable) [GH-14077]
* provider: Add `assume_role` configuration block `duration_seconds`, `policy_arns`, `tags`, and `transitive_tag_keys` arguments [GH-14077]
* data-source/aws_instance: Add `secondary_private_ips` attribute [GH-14079]
* resource/aws_api_gateway_method_settings: Add import support [GH-14266]
* resource/aws_instance: Add `secondary_private_ips` argument (conflicts with `network_interface` configuration block) [GH-14079]

BUG FIXES

* provider: Ensure nil is not passed to RetryError helpers, may result in some bug fixes [GH-14104]
* provider: Ensure configured STS endpoint is used during `AssumeRole` API calls [GH-14077]
* provider: Prefer AWS shared configuration over EC2 metadata credentials by default [GH-14077]
* provider: Prefer CodeBuild, ECS, EKS credentials over EC2 metadata credentials by default [GH-14077]
* resource/aws_codepipeline: Only retry `CreatePipeline` errors for IAM eventual consistency errors [GH-14264]
* resource/aws_network_acl_rule: Immediately return `DescribeNetworkAcls` errors on creation [GH-14261]
* resource/aws_spot_fleet_request: Only retry `RequestSpotFleet` on IAM eventual consistency errors and use standard 2 minute timeout [GH-14265]
* resource/aws_ssm_activation: Only retry `CreateActivation` on IAM eventual consistency errors and use standard 2 minute timeout [GH-14263]

## Previous Releases

For information on prior major releases, see their changelogs:

* [2.x and earlier](https://github.com/terraform-providers/terraform-provider-aws/blob/release/2.x/CHANGELOG.md)
