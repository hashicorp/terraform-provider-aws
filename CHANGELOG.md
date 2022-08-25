## 4.28.0 (Unreleased)

NOTES:

* resource/aws_db_instance: With the retirement of EC2-Classic the`security_group_names` attribute has been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_db_security_group: With the retirement of EC2-Classic the`aws_db_security_group` resource has been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_elasticache_cluster: With the retirement of EC2-Classic the`security_group_names` attribute has been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_elasticache_security_group: With the retirement of EC2-Classic the`aws_elasticache_security_group` resource has been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_launch_configuration: With the retirement of EC2-Classic the`vpc_classic_link_id` and `vpc_classic_link_security_groups` attributes have been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_redshift_cluster: With the retirement of EC2-Classic the`cluster_security_groups` attribute has been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_redshift_security_group: With the retirement of EC2-Classic the`aws_redshift_security_group` resource has been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_vpc: With the retirement of EC2-Classic the`enable_classiclink` and `enable_classiclink_dns_support` attributes have been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_vpc_peering_connection: With the retirement of EC2-Classic the`allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` attributes have been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_vpc_peering_connection_accepter: With the retirement of EC2-Classic the`allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` attributes have been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))
* resource/aws_vpc_peering_connection_options: With the retirement of EC2-Classic the`allow_classic_link_to_remote_vpc` and `allow_vpc_to_remote_classic_link` attributes have been deprecated and will be removed in a future version ([#26427](https://github.com/hashicorp/terraform-provider-aws/issues/26427))

FEATURES:

* **New Data Source:** `aws_ec2_network_insights_analysis` ([#23532](https://github.com/hashicorp/terraform-provider-aws/issues/23532))
* **New Data Source:** `aws_ec2_network_insights_path` ([#23532](https://github.com/hashicorp/terraform-provider-aws/issues/23532))
* **New Data Source:** `aws_ec2_transit_gateway_attachment` ([#26264](https://github.com/hashicorp/terraform-provider-aws/issues/26264))
* **New Data Source:** `aws_location_tracker_association` ([#26404](https://github.com/hashicorp/terraform-provider-aws/issues/26404))
* **New Resource:** `aws_ec2_network_insights_analysis` ([#23532](https://github.com/hashicorp/terraform-provider-aws/issues/23532))
* **New Resource:** `aws_ec2_transit_gateway_policy_table` ([#26264](https://github.com/hashicorp/terraform-provider-aws/issues/26264))
* **New Resource:** `aws_ec2_transit_gateway_policy_table_association` ([#26264](https://github.com/hashicorp/terraform-provider-aws/issues/26264))
* **New Resource:** `aws_grafana_workspace_api_key` ([#25286](https://github.com/hashicorp/terraform-provider-aws/issues/25286))
* **New Resource:** `aws_networkmanager_transit_gateway_peering` ([#26264](https://github.com/hashicorp/terraform-provider-aws/issues/26264))
* **New Resource:** `aws_networkmanager_transit_gateway_route_table_attachment` ([#26264](https://github.com/hashicorp/terraform-provider-aws/issues/26264))
* **New Resource:** `aws_redshiftserverless_workgroup` ([#26467](https://github.com/hashicorp/terraform-provider-aws/issues/26467))

ENHANCEMENTS:

* data-source/aws_db_instance: Add `network_type` attribute ([#26185](https://github.com/hashicorp/terraform-provider-aws/issues/26185))
* data-source/aws_db_subnet_group: Add `supported_network_types` attribute ([#26185](https://github.com/hashicorp/terraform-provider-aws/issues/26185))
* data-source/aws_rds_orderable_db_instance: Add `supported_network_types` attribute ([#26185](https://github.com/hashicorp/terraform-provider-aws/issues/26185))
* resource/aws_db_instance: Add `network_type` argument ([#26185](https://github.com/hashicorp/terraform-provider-aws/issues/26185))
* resource/aws_db_subnet_group: Add `supported_network_types` argument ([#26185](https://github.com/hashicorp/terraform-provider-aws/issues/26185))
* resource/aws_glue_job: Add support for `3.9` as valid `python_version` value ([#26407](https://github.com/hashicorp/terraform-provider-aws/issues/26407))
* resource/aws_kendra_index: The `document_metadata_configuration_updates` argument can now be updated. Refer to the documentation for more details. ([#20294](https://github.com/hashicorp/terraform-provider-aws/issues/20294))

BUG FIXES:

* resource/aws_appstream_fleet: Fix crash when providing empty `domain_join_info` (_e.g._, `directory_name = ""`) ([#26454](https://github.com/hashicorp/terraform-provider-aws/issues/26454))
* resource/aws_eip: Include any provider-level configured `default_tags` on resource Create ([#26308](https://github.com/hashicorp/terraform-provider-aws/issues/26308))
* resource/aws_kinesis_firehose_delivery_stream: Updating `tags` no longer causes an unnecessary update ([#26451](https://github.com/hashicorp/terraform-provider-aws/issues/26451))
* resource/aws_organizations_policy: Prevent `InvalidParameter` errors by handling `content` as generic JSON, not an IAM policy ([#26279](https://github.com/hashicorp/terraform-provider-aws/issues/26279))

## 4.27.0 (August 19, 2022)

FEATURES:

* **New Resource:** `aws_msk_serverless_cluster` ([#25684](https://github.com/hashicorp/terraform-provider-aws/issues/25684))
* **New Resource:** `aws_networkmanager_attachment_accepter` ([#26227](https://github.com/hashicorp/terraform-provider-aws/issues/26227))
* **New Resource:** `aws_networkmanager_vpc_attachment` ([#26227](https://github.com/hashicorp/terraform-provider-aws/issues/26227))

ENHANCEMENTS:

* data-source/aws_networkfirewall_firewall: Add `capacity_usage_summary`, `configuration_sync_state_summary`, and `status` attributes to the `firewall_status` block ([#26284](https://github.com/hashicorp/terraform-provider-aws/issues/26284))
* resource/aws_acm_certificate: Add `not_after` argument ([#26281](https://github.com/hashicorp/terraform-provider-aws/issues/26281))
* resource/aws_acm_certificate: Add `not_before` argument ([#26281](https://github.com/hashicorp/terraform-provider-aws/issues/26281))
* resource/aws_chime_voice_connector_logging: Add `enable_media_metric_logs` argument ([#26283](https://github.com/hashicorp/terraform-provider-aws/issues/26283))
* resource/aws_cloudfront_distribution: Support `http3` and `http2and3` as valid values for the `http_version` argument ([#26313](https://github.com/hashicorp/terraform-provider-aws/issues/26313))
* resource/aws_inspector_assessment_template: Add `event_subscription` configuration block ([#26334](https://github.com/hashicorp/terraform-provider-aws/issues/26334))
* resource/aws_lb_target_group: Add `ip_address_type` argument ([#26320](https://github.com/hashicorp/terraform-provider-aws/issues/26320))
* resource/aws_opsworks_stack: Add plan-time validation for `custom_cookbooks_source.type` ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))

BUG FIXES:

* resource/aws_appflow_flow: Correctly specify `trigger_config.trigger_properties.scheduled.schedule_start_time` during create and update ([#26289](https://github.com/hashicorp/terraform-provider-aws/issues/26289))
* resource/aws_db_instance: Prevent `InvalidParameterCombination: No modifications were requested` errors when only `delete_automated_backups`, `final_snapshot_identifier` and/or `skip_final_snapshot` change ([#26286](https://github.com/hashicorp/terraform-provider-aws/issues/26286))
* resource/aws_opsworks_custom_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_ecs_cluster_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_ganglia_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_haproxy_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_java_app_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_memcached_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_mysql_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_nodejs_app_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_php_app_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_rails_app_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_stack: Correctly apply `tags` during create if `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))
* resource/aws_opsworks_static_web_layer: Correctly apply `tags` during create if the stack's `region` is not equal to the configured AWS Region ([#26278](https://github.com/hashicorp/terraform-provider-aws/issues/26278))

## 4.26.0 (August 12, 2022)

FEATURES:

* **New Data Source:** `aws_fsx_openzfs_snapshot` ([#26184](https://github.com/hashicorp/terraform-provider-aws/issues/26184))
* **New Data Source:** `aws_networkfirewall_firewall` ([#25495](https://github.com/hashicorp/terraform-provider-aws/issues/25495))
* **New Data Source:** `aws_prometheus_workspace` ([#26120](https://github.com/hashicorp/terraform-provider-aws/issues/26120))
* **New Resource:** `aws_comprehend_entity_recognizer` ([#26244](https://github.com/hashicorp/terraform-provider-aws/issues/26244))
* **New Resource:** `aws_connect_instance_storage_config` ([#26152](https://github.com/hashicorp/terraform-provider-aws/issues/26152))
* **New Resource:** `aws_directory_service_radius_settings` ([#14045](https://github.com/hashicorp/terraform-provider-aws/issues/14045))
* **New Resource:** `aws_directory_service_region` ([#25755](https://github.com/hashicorp/terraform-provider-aws/issues/25755))
* **New Resource:** `aws_dynamodb_table_replica` ([#26250](https://github.com/hashicorp/terraform-provider-aws/issues/26250))
* **New Resource:** `aws_location_tracker_association` ([#26061](https://github.com/hashicorp/terraform-provider-aws/issues/26061))

ENHANCEMENTS:

* data-source/aws_directory_service_directory: Add `radius_settings` attribute ([#14045](https://github.com/hashicorp/terraform-provider-aws/issues/14045))
* data-source/aws_directory_service_directory: Set `dns_ip_addresses` to the owner directory's DNS IP addresses for SharedMicrosoftAD directories ([#20819](https://github.com/hashicorp/terraform-provider-aws/issues/20819))
* data-source/aws_elasticsearch_domain: Add `throughput` attribute to the `ebs_options` configuration block ([#26045](https://github.com/hashicorp/terraform-provider-aws/issues/26045))
* data-source/aws_opensearch_domain: Add `throughput` attribute to the `ebs_options` configuration block ([#26045](https://github.com/hashicorp/terraform-provider-aws/issues/26045))
* resource/aws_autoscaling_group: Better error handling when attempting to create Auto Scaling groups with incompatible options ([#25987](https://github.com/hashicorp/terraform-provider-aws/issues/25987))
* resource/aws_backup_vault: Add `force_destroy` argument ([#26199](https://github.com/hashicorp/terraform-provider-aws/issues/26199))
* resource/aws_directory_service_directory: Add `desired_number_of_domain_controllers` argument ([#25755](https://github.com/hashicorp/terraform-provider-aws/issues/25755))
* resource/aws_directory_service_directory: Add configurable timeouts for Create, Update and Delete ([#25755](https://github.com/hashicorp/terraform-provider-aws/issues/25755))
* resource/aws_directory_service_shared_directory: Add configurable timeouts for Delete ([#25755](https://github.com/hashicorp/terraform-provider-aws/issues/25755))
* resource/aws_directory_service_shared_directory_accepter: Add configurable timeouts for Create and Delete ([#25755](https://github.com/hashicorp/terraform-provider-aws/issues/25755))
* resource/aws_elasticsearch_domain: Add `throughput` attribute to the `ebs_options` configuration block ([#26045](https://github.com/hashicorp/terraform-provider-aws/issues/26045))
* resource/aws_glue_job: Add `execution_class` argument ([#26188](https://github.com/hashicorp/terraform-provider-aws/issues/26188))
* resource/aws_macie2_classification_job: Add `bucket_criteria` attribute to the `s3_job_definition` configuration block ([#19837](https://github.com/hashicorp/terraform-provider-aws/issues/19837))
* resource/aws_opensearch_domain: Add `throughput` attribute to the `ebs_options` configuration block ([#26045](https://github.com/hashicorp/terraform-provider-aws/issues/26045))

BUG FIXES:

* resource/aws_appflow_flow: Fix `trigger_properties.scheduled` being set during resource read ([#26240](https://github.com/hashicorp/terraform-provider-aws/issues/26240))
* resource/aws_db_instance: Add retries (for handling IAM eventual consistency) when creating database replicas that use enhanced monitoring ([#20926](https://github.com/hashicorp/terraform-provider-aws/issues/20926))
* resource/aws_db_instance: Apply `monitoring_interval` and `monitoring_role_arn` when creating via `restore_to_point_in_time` ([#20926](https://github.com/hashicorp/terraform-provider-aws/issues/20926))
* resource/aws_dynamodb_table: Fix `replica.*.propagate_tags` not propagating tags to newly added replicas ([#26257](https://github.com/hashicorp/terraform-provider-aws/issues/26257))
* resource/aws_emr_instance_group: Handle deleted instance groups during resource read ([#26154](https://github.com/hashicorp/terraform-provider-aws/issues/26154))
* resource/aws_emr_instance_group: Mark `instance_count` as Computed to prevent diff when autoscaling is active ([#26154](https://github.com/hashicorp/terraform-provider-aws/issues/26154))
* resource/aws_lb_listener: Fix `ValidationError` when tags are added on `create` ([#26194](https://github.com/hashicorp/terraform-provider-aws/issues/26194))
* resource/aws_lb_target_group: Fix `ValidationError` when tags are added on `create` ([#26194](https://github.com/hashicorp/terraform-provider-aws/issues/26194))
* resource/aws_macie2_classification_job: Fix incorrect plan diff for `TagScopeTerm()` when updating resources ([#19837](https://github.com/hashicorp/terraform-provider-aws/issues/19837))
* resource/aws_security_group_rule: Disallow empty strings in `prefix_list_ids` ([#26220](https://github.com/hashicorp/terraform-provider-aws/issues/26220))

## 4.25.0 (August  4, 2022)

FEATURES:

* **New Data Source:** `aws_waf_subscribed_rule_group` ([#10563](https://github.com/hashicorp/terraform-provider-aws/issues/10563))
* **New Data Source:** `aws_wafregional_subscribed_rule_group` ([#10563](https://github.com/hashicorp/terraform-provider-aws/issues/10563))
* **New Resource:** `aws_kendra_data_source` ([#25686](https://github.com/hashicorp/terraform-provider-aws/issues/25686))
* **New Resource:** `aws_macie2_classification_export_configuration` ([#19856](https://github.com/hashicorp/terraform-provider-aws/issues/19856))
* **New Resource:** `aws_transcribe_language_model` ([#25698](https://github.com/hashicorp/terraform-provider-aws/issues/25698))

ENHANCEMENTS:

* data-source/aws_alb: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ami: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ami_ids: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_availability_zone: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_availability_zones: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_customer_gateway: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_dx_location: Add `available_macsec_port_speeds` attribute ([#26110](https://github.com/hashicorp/terraform-provider-aws/issues/26110))
* data-source/aws_ebs_default_kms_key: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ebs_encryption_by_default: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ebs_snapshot: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ebs_snapshot_ids: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ebs_volume: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ebs_volumes: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_client_vpn_endpoint: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_coip_pool: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_coip_pools: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_host: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_instance_type: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_instance_type_offering: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_instance_type_offerings: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_instance_types: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_local_gateway: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_local_gateway_route_table: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_local_gateway_route_tables: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_local_gateway_virtual_interface: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_local_gateway_virtual_interface_group: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_local_gateway_virtual_interface_groups: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_local_gateways: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_managed_prefix_list: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_serial_console_access: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_spot_price: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_connect: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_connect_peer: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_dx_gateway_attachment: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_multicast_domain: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_peering_attachment: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_route_table: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_route_tables: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_vpc_attachment: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_vpc_attachments: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_ec2_transit_gateway_vpn_attachment: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_eip: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_eips: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_instance: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_instances: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_internet_gateway: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_key_pair: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_launch_template: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_lb: Add `preserve_host_header` attribute ([#26056](https://github.com/hashicorp/terraform-provider-aws/issues/26056))
* data-source/aws_lb: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_lb_listener: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_lb_target_group: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_nat_gateway: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_nat_gateways: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_network_acls: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_network_interface: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_network_interfaces: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_prefix_list: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_route: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_route_table: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_route_tables: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_security_group: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_security_groups: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_subnet: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_subnet_ids: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_subnets: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpc: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpc_dhcp_options: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpc_endpoint: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpc_endpoint_service: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpc_ipam_pool: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpc_ipam_preview_next_cidr: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpc_peering_connection: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpc_peering_connections: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpcs: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* data-source/aws_vpn_gateway: Allow customizable read timeout ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* resource/aws_ecrpublic_repository: Add `tags` argument and `tags_all` attribute to support resource tagging ([#26057](https://github.com/hashicorp/terraform-provider-aws/issues/26057))
* resource/aws_fsx_openzfs_file_system: Add `root_volume_configuration.record_size_kib` argument ([#26049](https://github.com/hashicorp/terraform-provider-aws/issues/26049))
* resource/aws_fsx_openzfs_volume: Add `record_size_kib` argument ([#26049](https://github.com/hashicorp/terraform-provider-aws/issues/26049))
* resource/aws_globalaccelerator_accelerator: Support `DUAL_STACK` value for `ip_address_type` ([#26055](https://github.com/hashicorp/terraform-provider-aws/issues/26055))
* resource/aws_iam_role_policy: Add plan time validation to `role` argument ([#26082](https://github.com/hashicorp/terraform-provider-aws/issues/26082))
* resource/aws_internet_gateway: Allow customizable timeouts ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* resource/aws_internet_gateway_attachment: Allow customizable timeouts ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))
* resource/aws_lb: Add `preserve_host_header` argument ([#26056](https://github.com/hashicorp/terraform-provider-aws/issues/26056))
* resource/aws_s3_bucket: Allow customizable timeouts ([#26121](https://github.com/hashicorp/terraform-provider-aws/issues/26121))

BUG FIXES:

* resource/aws_api_gateway_rest_api: Add `put_rest_api_mode` argument to address race conditions when importing OpenAPI Specifications ([#26051](https://github.com/hashicorp/terraform-provider-aws/issues/26051))
* resource/aws_appstream_fleet: Fix IAM `InvalidRoleException` error on creation ([#26060](https://github.com/hashicorp/terraform-provider-aws/issues/26060))

## 4.24.0 (July 29, 2022)

FEATURES:

* **New Resource:** `aws_acmpca_permission` ([#12485](https://github.com/hashicorp/terraform-provider-aws/issues/12485))
* **New Resource:** `aws_ssm_service_setting` ([#13018](https://github.com/hashicorp/terraform-provider-aws/issues/13018))

ENHANCEMENTS:

* data-source/aws_ecs_service: Add `tags` attribute ([#25961](https://github.com/hashicorp/terraform-provider-aws/issues/25961))
* resource/aws_datasync_task: Add `includes` argument ([#25929](https://github.com/hashicorp/terraform-provider-aws/issues/25929))
* resource/aws_guardduty_detector: Add `malware_protection` attribute to the `datasources` configuration block ([#25994](https://github.com/hashicorp/terraform-provider-aws/issues/25994))
* resource/aws_guardduty_organization_configuration: Add `malware_protection` attribute to the `datasources` configuration block ([#25992](https://github.com/hashicorp/terraform-provider-aws/issues/25992))
* resource/aws_security_group: Additional plan-time validation for `name` and `name_prefix` ([#15011](https://github.com/hashicorp/terraform-provider-aws/issues/15011))
* resource/aws_security_group_rule: Add configurable Create timeout ([#24340](https://github.com/hashicorp/terraform-provider-aws/issues/24340))
* resource/aws_ses_configuration_set: Add `tracking_options.0.custom_redirect_domain` argument (NOTE: This enhancement is provided as best effort due to testing limitations, i.e., the requirement of a verified domain) ([#26032](https://github.com/hashicorp/terraform-provider-aws/issues/26032))

BUG FIXES:

* data-source/aws_networkmanager_core_network_policy_document: Fix bug where bool values for `attachment-policy.action.require-acceptance` can only be `true` or omitted ([#26010](https://github.com/hashicorp/terraform-provider-aws/issues/26010))
* resource/aws_appmesh_gateway_route: Fix crash when only one of hostname rewrite or path rewrite is configured ([#26012](https://github.com/hashicorp/terraform-provider-aws/issues/26012))
* resource/aws_ce_anomaly_subscription:Fix crash upon adding or removing monitor ARNs to `monitor_arn_list`. ([#25941](https://github.com/hashicorp/terraform-provider-aws/issues/25941))
* resource/aws_cognito_identity_pool_provider_principal_tag: Fix read operation when using an OIDC provider ([#25964](https://github.com/hashicorp/terraform-provider-aws/issues/25964))
* resource/aws_route53_record: Don't ignore `dualstack` prefix in Route 53 Record alias names ([#10672](https://github.com/hashicorp/terraform-provider-aws/issues/10672))
* resource/aws_s3_bucket: Prevents unexpected import of existing bucket in `us-east-1`. ([#26011](https://github.com/hashicorp/terraform-provider-aws/issues/26011))
* resource/aws_s3_bucket: Refactored `object_lock_enabled` parameter's default assignment behavior to protect partitions without Object Lock available. ([#25098](https://github.com/hashicorp/terraform-provider-aws/issues/25098))

## 4.23.0 (July 22, 2022)

FEATURES:

* **New Data Source:** `aws_connect_user_hierarchy_group` ([#24777](https://github.com/hashicorp/terraform-provider-aws/issues/24777))
* **New Data Source:** `aws_location_geofence_collection` ([#25844](https://github.com/hashicorp/terraform-provider-aws/issues/25844))
* **New Data Source:** `aws_networkfirewall_firewall_policy` ([#24748](https://github.com/hashicorp/terraform-provider-aws/issues/24748))
* **New Data Source:** `aws_s3_account_public_access_block` ([#25781](https://github.com/hashicorp/terraform-provider-aws/issues/25781))
* **New Resource:** `aws_connect_user` ([#24832](https://github.com/hashicorp/terraform-provider-aws/issues/24832))
* **New Resource:** `aws_connect_vocabulary` ([#24849](https://github.com/hashicorp/terraform-provider-aws/issues/24849))
* **New Resource:** `aws_location_geofence_collection` ([#25762](https://github.com/hashicorp/terraform-provider-aws/issues/25762))
* **New Resource:** `aws_redshiftserverless_namespace` ([#25889](https://github.com/hashicorp/terraform-provider-aws/issues/25889))
* **New Resource:** `aws_rolesanywhere_profile` ([#25850](https://github.com/hashicorp/terraform-provider-aws/issues/25850))
* **New Resource:** `aws_rolesanywhere_trust_anchor` ([#25779](https://github.com/hashicorp/terraform-provider-aws/issues/25779))
* **New Resource:** `aws_transcribe_vocabulary` ([#25863](https://github.com/hashicorp/terraform-provider-aws/issues/25863))
* **New Resource:** `aws_transcribe_vocabulary_filter` ([#25918](https://github.com/hashicorp/terraform-provider-aws/issues/25918))

ENHANCEMENTS:

* data-source/aws_imagebuilder_container_recipe: Add `throughput` attribute to the `block_device_mapping` configuration block ([#25790](https://github.com/hashicorp/terraform-provider-aws/issues/25790))
* data-source/aws_imagebuilder_image_recipe: Add `throughput` attribute to the `block_device_mapping` configuration block ([#25790](https://github.com/hashicorp/terraform-provider-aws/issues/25790))
* data/aws_outposts_asset: Add `rack_elevation` attribute ([#25822](https://github.com/hashicorp/terraform-provider-aws/issues/25822))
* resource/aws_appmesh_gateway_route: Add `http2_route.action.rewrite`, `http2_route.match.hostname`, `http_route.action.rewrite` and `http_route.match.hostname` arguments ([#25819](https://github.com/hashicorp/terraform-provider-aws/issues/25819))
* resource/aws_ce_cost_category: Add `tags` argument and `tags_all` attribute to support resource tagging ([#25432](https://github.com/hashicorp/terraform-provider-aws/issues/25432))
* resource/aws_db_instance_automated_backups_replication: Add support for custom timeouts (create and delete) ([#25796](https://github.com/hashicorp/terraform-provider-aws/issues/25796))
* resource/aws_dynamodb_table: Add `replica.*.propagate_tags` argument to allow propagating tags to replicas ([#25866](https://github.com/hashicorp/terraform-provider-aws/issues/25866))
* resource/aws_flow_log: Add `transit_gateway_id` and `transit_gateway_attachment_id` arguments ([#25913](https://github.com/hashicorp/terraform-provider-aws/issues/25913))
* resource/aws_fsx_openzfs_file_system: Allow in-place update of `storage_capacity`, `throughput_capacity`, and `disk_iops_configuration`. ([#25841](https://github.com/hashicorp/terraform-provider-aws/issues/25841))
* resource/aws_guardduty_organization_configuration: Add `kubernetes` attribute to the `datasources` configuration block ([#25131](https://github.com/hashicorp/terraform-provider-aws/issues/25131))
* resource/aws_imagebuilder_container_recipe: Add `throughput` argument to the `block_device_mapping` configuration block ([#25790](https://github.com/hashicorp/terraform-provider-aws/issues/25790))
* resource/aws_imagebuilder_image_recipe: Add `throughput` argument to the `block_device_mapping` configuration block ([#25790](https://github.com/hashicorp/terraform-provider-aws/issues/25790))
* resource/aws_rds_cluster_instance: Allow `performance_insights_retention_period` values that are multiples of `31` ([#25729](https://github.com/hashicorp/terraform-provider-aws/issues/25729))

BUG FIXES:

* data-source/aws_networkmanager_core_network_policy_document: Fix bug where bool values in `segments` blocks weren't being included in json payloads ([#25789](https://github.com/hashicorp/terraform-provider-aws/issues/25789))
* resource/aws_connect_hours_of_operation: Fix tags not being updated ([#24864](https://github.com/hashicorp/terraform-provider-aws/issues/24864))
* resource/aws_connect_queue: Fix tags not being updated ([#24864](https://github.com/hashicorp/terraform-provider-aws/issues/24864))
* resource/aws_connect_quick_connect: Fix tags not being updated ([#24864](https://github.com/hashicorp/terraform-provider-aws/issues/24864))
* resource/aws_connect_routing_profile: Fix tags not being updated ([#24864](https://github.com/hashicorp/terraform-provider-aws/issues/24864))
* resource/aws_connect_security_profile: Fix tags not being updated ([#24864](https://github.com/hashicorp/terraform-provider-aws/issues/24864))
* resource/aws_connect_user_hierarchy_group: Fix tags not being updated ([#24864](https://github.com/hashicorp/terraform-provider-aws/issues/24864))
* resource/aws_iam_role: Fix diffs in `assume_role_policy` when there are no semantic changes ([#23060](https://github.com/hashicorp/terraform-provider-aws/issues/23060))
* resource/aws_iam_role: Fix problem with exclusive management of inline and managed policies when empty (i.e., remove out-of-band policies) ([#23060](https://github.com/hashicorp/terraform-provider-aws/issues/23060))
* resource/aws_rds_cluster: Prevent failure of AWS RDS Cluster creation when it is in `rebooting` state. ([#25718](https://github.com/hashicorp/terraform-provider-aws/issues/25718))
* resource/aws_route_table: Retry resource Create for EC2 eventual consistency ([#25793](https://github.com/hashicorp/terraform-provider-aws/issues/25793))
* resource/aws_storagegateway_gateway: Only manage `average_download_rate_limit_in_bits_per_sec` and `average_upload_rate_limit_in_bits_per_sec` when gateway type supports rate limits ([#25922](https://github.com/hashicorp/terraform-provider-aws/issues/25922))

## 4.22.0 (July  8, 2022)

FEATURES:

* **New Data Source:** `aws_location_route_calculator` ([#25689](https://github.com/hashicorp/terraform-provider-aws/issues/25689))
* **New Data Source:** `aws_location_tracker` ([#25639](https://github.com/hashicorp/terraform-provider-aws/issues/25639))
* **New Data Source:** `aws_secretsmanager_random_password` ([#25704](https://github.com/hashicorp/terraform-provider-aws/issues/25704))
* **New Resource:** `aws_directory_service_shared_directory` ([#24766](https://github.com/hashicorp/terraform-provider-aws/issues/24766))
* **New Resource:** `aws_directory_service_shared_directory_accepter` ([#24766](https://github.com/hashicorp/terraform-provider-aws/issues/24766))
* **New Resource:** `aws_lightsail_database` ([#18663](https://github.com/hashicorp/terraform-provider-aws/issues/18663))
* **New Resource:** `aws_location_route_calculator` ([#25656](https://github.com/hashicorp/terraform-provider-aws/issues/25656))
* **New Resource:** `aws_transcribe_medical_vocabulary` ([#25723](https://github.com/hashicorp/terraform-provider-aws/issues/25723))

ENHANCEMENTS:

* data-source/aws_imagebuilder_distribution_configuration: Add `fast_launch_configuration` attribute to the `distribution` configuration block ([#25671](https://github.com/hashicorp/terraform-provider-aws/issues/25671))
* resource/aws_acmpca_certificate_authority: Add `revocation_configuration.ocsp_configuration` argument ([#25720](https://github.com/hashicorp/terraform-provider-aws/issues/25720))
* resource/aws_apprunner_service: Add `observability_configuration` argument configuration block ([#25697](https://github.com/hashicorp/terraform-provider-aws/issues/25697))
* resource/aws_autoscaling_group: Add `default_instance_warmup` attribute ([#25722](https://github.com/hashicorp/terraform-provider-aws/issues/25722))
* resource/aws_config_remediation_configuration: Add `parameter.*.static_values` attribute for a list of values ([#25738](https://github.com/hashicorp/terraform-provider-aws/issues/25738))
* resource/aws_dynamodb_table: Add `replica.*.point_in_time_recovery` argument ([#25659](https://github.com/hashicorp/terraform-provider-aws/issues/25659))
* resource/aws_ecr_repository: Add `force_delete` parameter. ([#9913](https://github.com/hashicorp/terraform-provider-aws/issues/9913))
* resource/aws_ecs_service: Add configurable timeouts for Create and Delete. ([#25641](https://github.com/hashicorp/terraform-provider-aws/issues/25641))
* resource/aws_emr_cluster: Add `core_instance_group.ebs_config.throughput` and `master_instance_group.ebs_config.throughput` arguments ([#25668](https://github.com/hashicorp/terraform-provider-aws/issues/25668))
* resource/aws_emr_cluster: Add `gp3` EBS volume support ([#25668](https://github.com/hashicorp/terraform-provider-aws/issues/25668))
* resource/aws_emr_cluster: Add `sc1` EBS volume support ([#25255](https://github.com/hashicorp/terraform-provider-aws/issues/25255))
* resource/aws_gamelift_game_session_queue: Add `notification_target` argument ([#25544](https://github.com/hashicorp/terraform-provider-aws/issues/25544))
* resource/aws_imagebuilder_distribution_configuration: Add `fast_launch_configuration` argument to the `distribution` configuration block ([#25671](https://github.com/hashicorp/terraform-provider-aws/issues/25671))
* resource/aws_placement_group: Add `spread_level` argument ([#25615](https://github.com/hashicorp/terraform-provider-aws/issues/25615))
* resource/aws_sagemaker_notebook_instance: Add `accelerator_types` argument ([#10210](https://github.com/hashicorp/terraform-provider-aws/issues/10210))
* resource/aws_sagemaker_project: Increase SageMaker Project create and delete timeout to 15 minutes ([#25638](https://github.com/hashicorp/terraform-provider-aws/issues/25638))
* resource/aws_ssm_parameter: Add `insecure_value` argument to enable dynamic use of SSM parameter values ([#25721](https://github.com/hashicorp/terraform-provider-aws/issues/25721))
* resource/aws_vpc_ipam_pool_cidr: Better error reporting ([#25287](https://github.com/hashicorp/terraform-provider-aws/issues/25287))

BUG FIXES:

* provider: Ensure that the configured `assume_role_with_web_identity` value is used ([#25681](https://github.com/hashicorp/terraform-provider-aws/issues/25681))
* resource/aws_acmpca_certificate_authority: Fix crash when `revocation_configuration` block is empty ([#25695](https://github.com/hashicorp/terraform-provider-aws/issues/25695))
* resource/aws_cognito_risk_configuration: Increase maximum allowed length of `account_takeover_risk_configuration.notify_configuration.block_email.html_body`, `account_takeover_risk_configuration.notify_configuration.block_email.text_body`, `account_takeover_risk_configuration.notify_configuration.mfa_email.html_body`, `account_takeover_risk_configuration.notify_configuration.mfa_email.text_body`, `account_takeover_risk_configuration.notify_configuration.no_action_email.html_body` and `account_takeover_risk_configuration.notify_configuration.no_action_email.text_body` arguments from `2000` to `20000` ([#25645](https://github.com/hashicorp/terraform-provider-aws/issues/25645))
* resource/aws_dynamodb_table: Prevent `restore_source_name` from forcing replacement when removed to enable restoring from a PITR backup ([#25659](https://github.com/hashicorp/terraform-provider-aws/issues/25659))
* resource/aws_dynamodb_table: Respect custom timeouts including when working with replicas ([#25659](https://github.com/hashicorp/terraform-provider-aws/issues/25659))
* resource/aws_ec2_transit_gateway: Fix MaxItems and subnet size validation in `transit_gateway_cidr_blocks` ([#25673](https://github.com/hashicorp/terraform-provider-aws/issues/25673))
* resource/aws_ecs_service: Fix "unexpected new value" errors on creation. ([#25641](https://github.com/hashicorp/terraform-provider-aws/issues/25641))
* resource/aws_ecs_service: Fix error where tags are sometimes not retrieved. ([#25641](https://github.com/hashicorp/terraform-provider-aws/issues/25641))
* resource/aws_emr_managed_scaling_policy: Support `maximum_ondemand_capacity_units` value of `0` ([#17134](https://github.com/hashicorp/terraform-provider-aws/issues/17134))

## 4.21.0 (June 30, 2022)

FEATURES:

* **New Data Source:** `aws_kendra_experience` ([#25601](https://github.com/hashicorp/terraform-provider-aws/issues/25601))
* **New Data Source:** `aws_kendra_query_suggestions_block_list` ([#25592](https://github.com/hashicorp/terraform-provider-aws/issues/25592))
* **New Data Source:** `aws_kendra_thesaurus` ([#25555](https://github.com/hashicorp/terraform-provider-aws/issues/25555))
* **New Data Source:** `aws_service_discovery_http_namespace` ([#25162](https://github.com/hashicorp/terraform-provider-aws/issues/25162))
* **New Data Source:** `aws_service_discovery_service` ([#25162](https://github.com/hashicorp/terraform-provider-aws/issues/25162))
* **New Resource:** `aws_accessanalyzer_archive_rule` ([#25514](https://github.com/hashicorp/terraform-provider-aws/issues/25514))
* **New Resource:** `aws_apprunner_observability_configuration` ([#25591](https://github.com/hashicorp/terraform-provider-aws/issues/25591))
* **New Resource:** `aws_lakeformation_resource_lf_tags` ([#25565](https://github.com/hashicorp/terraform-provider-aws/issues/25565))

ENHANCEMENTS:

* data-source/aws_ami: Add `include_deprecated` argument ([#25566](https://github.com/hashicorp/terraform-provider-aws/issues/25566))
* data-source/aws_ami: Make `owners` optional ([#25566](https://github.com/hashicorp/terraform-provider-aws/issues/25566))
* data-source/aws_service_discovery_dns_namespace: Add `tags` attribute ([#25162](https://github.com/hashicorp/terraform-provider-aws/issues/25162))
* data/aws_key_pair: New attribute `public_key` populated by setting the new `include_public_key` argument ([#25371](https://github.com/hashicorp/terraform-provider-aws/issues/25371))
* resource/aws_connect_instance: Configurable Create and Delete timeouts ([#24861](https://github.com/hashicorp/terraform-provider-aws/issues/24861))
* resource/aws_key_pair: Added 2 new attributes - `key_type` and `create_time` ([#25371](https://github.com/hashicorp/terraform-provider-aws/issues/25371))
* resource/aws_sagemaker_model: Add `repository_auth_config` arguments in support of [Private Docker Registry](https://docs.aws.amazon.com/sagemaker/latest/dg/your-algorithms-containers-inference-private.html) ([#25557](https://github.com/hashicorp/terraform-provider-aws/issues/25557))
* resource/aws_service_discovery_http_namespace: Add `http_name` attribute ([#25162](https://github.com/hashicorp/terraform-provider-aws/issues/25162))
* resource/aws_wafv2_web_acl: Add `rule.action.captcha` argument ([#21766](https://github.com/hashicorp/terraform-provider-aws/issues/21766))

BUG FIXES:

* resource/aws_api_gateway_model: Remove length validation from `schema` argument ([#25623](https://github.com/hashicorp/terraform-provider-aws/issues/25623))
* resource/aws_appstream_fleet_stack_association: Fix association not being found after creation ([#25370](https://github.com/hashicorp/terraform-provider-aws/issues/25370))
* resource/aws_appstream_stack: Fix crash when setting `embed_host_domains` ([#25372](https://github.com/hashicorp/terraform-provider-aws/issues/25372))
* resource/aws_route53_record: Successfully allow renaming of `set_identifier` (specified with multiple routing policies) ([#25620](https://github.com/hashicorp/terraform-provider-aws/issues/25620))

## 4.20.1 (June 24, 2022)

BUG FIXES:

* resource/aws_default_vpc_dhcp_options: Fix `missing expected [` error introduced in [v4.20.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#4200-june-23-2022) ([#25562](https://github.com/hashicorp/terraform-provider-aws/issues/25562))

## 4.20.0 (June 23, 2022)

FEATURES:

* **New Data Source:** `aws_kendra_faq` ([#25523](https://github.com/hashicorp/terraform-provider-aws/issues/25523))
* **New Data Source:** `aws_kendra_index` ([#25473](https://github.com/hashicorp/terraform-provider-aws/issues/25473))
* **New Data Source:** `aws_outposts_asset` ([#25476](https://github.com/hashicorp/terraform-provider-aws/issues/25476))
* **New Data Source:** `aws_outposts_assets` ([#25476](https://github.com/hashicorp/terraform-provider-aws/issues/25476))
* **New Resource:** `aws_applicationinsights_application` ([#25195](https://github.com/hashicorp/terraform-provider-aws/issues/25195))
* **New Resource:** `aws_ce_anomaly_monitor` ([#25177](https://github.com/hashicorp/terraform-provider-aws/issues/25177))
* **New Resource:** `aws_ce_anomaly_subscription` ([#25224](https://github.com/hashicorp/terraform-provider-aws/issues/25224))
* **New Resource:** `aws_ce_cost_allocation_tag` ([#25272](https://github.com/hashicorp/terraform-provider-aws/issues/25272))
* **New Resource:** `aws_cloudwatchrum_app_monitor` ([#25180](https://github.com/hashicorp/terraform-provider-aws/issues/25180))
* **New Resource:** `aws_cognito_risk_configuration` ([#25282](https://github.com/hashicorp/terraform-provider-aws/issues/25282))
* **New Resource:** `aws_kendra_experience` ([#25315](https://github.com/hashicorp/terraform-provider-aws/issues/25315))
* **New Resource:** `aws_kendra_faq` ([#25515](https://github.com/hashicorp/terraform-provider-aws/issues/25515))
* **New Resource:** `aws_kendra_query_suggestions_block_list` ([#25198](https://github.com/hashicorp/terraform-provider-aws/issues/25198))
* **New Resource:** `aws_kendra_thesaurus` ([#25199](https://github.com/hashicorp/terraform-provider-aws/issues/25199))
* **New Resource:** `aws_lakeformation_lf_tag` ([#19523](https://github.com/hashicorp/terraform-provider-aws/issues/19523))
* **New Resource:** `aws_location_tracker` ([#25466](https://github.com/hashicorp/terraform-provider-aws/issues/25466))

ENHANCEMENTS:

* data-source/aws_instance: Add `disable_api_stop` attribute ([#25185](https://github.com/hashicorp/terraform-provider-aws/issues/25185))
* data-source/aws_instance: Add `private_dns_name_options` attribute ([#25161](https://github.com/hashicorp/terraform-provider-aws/issues/25161))
* data-source/aws_instance: Correctly set `credit_specification` for T4g instances ([#25161](https://github.com/hashicorp/terraform-provider-aws/issues/25161))
* data-source/aws_launch_template: Add `disable_api_stop` attribute ([#25185](https://github.com/hashicorp/terraform-provider-aws/issues/25185))
* data-source/aws_launch_template: Correctly set `credit_specification` for T4g instances ([#25161](https://github.com/hashicorp/terraform-provider-aws/issues/25161))
* data-source/aws_vpc_endpoint: Add `dns_options` and `ip_address_type` attributes ([#25190](https://github.com/hashicorp/terraform-provider-aws/issues/25190))
* data-source/aws_vpc_endpoint_service: Add `supported_ip_address_types` attribute ([#25189](https://github.com/hashicorp/terraform-provider-aws/issues/25189))
* resource/aws_cloudwatch_event_api_destination: Remove validation of a maximum value for the `invocation_rate_limit_per_second` argument ([#25277](https://github.com/hashicorp/terraform-provider-aws/issues/25277))
* resource/aws_datasync_location_efs: Add `access_point_arn`, `file_system_access_role_arn`, and `in_transit_encryption` arguments ([#25182](https://github.com/hashicorp/terraform-provider-aws/issues/25182))
* resource/aws_datasync_location_efs: Add plan time validations for `ec2_config.security_group_arns` ([#25182](https://github.com/hashicorp/terraform-provider-aws/issues/25182))
* resource/aws_ec2_host: Add `outpost_arn` argument ([#25464](https://github.com/hashicorp/terraform-provider-aws/issues/25464))
* resource/aws_instance: Add `disable_api_stop` argument ([#25185](https://github.com/hashicorp/terraform-provider-aws/issues/25185))
* resource/aws_instance: Add `private_dns_name_options` argument ([#25161](https://github.com/hashicorp/terraform-provider-aws/issues/25161))
* resource/aws_instance: Correctly handle `credit_specification` for T4g instances ([#25161](https://github.com/hashicorp/terraform-provider-aws/issues/25161))
* resource/aws_launch_template: Add `disable_api_stop` argument ([#25185](https://github.com/hashicorp/terraform-provider-aws/issues/25185))
* resource/aws_launch_template: Correctly handle `credit_specification` for T4g instances ([#25161](https://github.com/hashicorp/terraform-provider-aws/issues/25161))
* resource/aws_s3_bucket_metric: Add validation to ensure name is <= 64 characters. ([#25260](https://github.com/hashicorp/terraform-provider-aws/issues/25260))
* resource/aws_sagemaker_endpoint_configuration: Add `serverless_config` argument ([#25218](https://github.com/hashicorp/terraform-provider-aws/issues/25218))
* resource/aws_sagemaker_endpoint_configuration: Make `production_variants.initial_instance_count` and `production_variants.instance_type` arguments optional ([#25218](https://github.com/hashicorp/terraform-provider-aws/issues/25218))
* resource/aws_sagemaker_notebook_instance: Add `instance_metadata_service_configuration` argument ([#25236](https://github.com/hashicorp/terraform-provider-aws/issues/25236))
* resource/aws_sagemaker_notebook_instance: Support `notebook-al2-v2` value for `platform_identifier` ([#25236](https://github.com/hashicorp/terraform-provider-aws/issues/25236))
* resource/aws_synthetics_canary: Add `delete_lambda` argument ([#25284](https://github.com/hashicorp/terraform-provider-aws/issues/25284))
* resource/aws_vpc_endpoint: Add `dns_options` and `ip_address_type` arguments ([#25190](https://github.com/hashicorp/terraform-provider-aws/issues/25190))
* resource/aws_vpc_endpoint_service: Add `supported_ip_address_types` argument ([#25189](https://github.com/hashicorp/terraform-provider-aws/issues/25189))
* resource/aws_vpn_connection: Add `outside_ip_address_type` and `transport_transit_gateway_attachment_id` arguments in support of [Private IP VPNs](https://docs.aws.amazon.com/vpn/latest/s2svpn/private-ip-dx.html) ([#25529](https://github.com/hashicorp/terraform-provider-aws/issues/25529))

BUG FIXES:

* data-source/aws_ecr_repository: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* data-source/aws_elasticache_cluster: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* data-source/aws_iam_policy: Add validation to prevent setting incompatible parameters. ([#25538](https://github.com/hashicorp/terraform-provider-aws/issues/25538))
* data-source/aws_iam_policy: Now loads tags. ([#25538](https://github.com/hashicorp/terraform-provider-aws/issues/25538))
* data-source/aws_lb: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* data-source/aws_lb_listener: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* data-source/aws_lb_target_group: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* data-source/aws_sqs_queue: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_api_gateway_model: Suppress whitespace differences between model schemas ([#25245](https://github.com/hashicorp/terraform-provider-aws/issues/25245))
* resource/aws_ce_cost_category: Allow duplicate values in `split_charge_rule.parameter.values` argument ([#25488](https://github.com/hashicorp/terraform-provider-aws/issues/25488))
* resource/aws_ce_cost_category: Fix error passing `split_charge_rule.parameter` to the AWS API ([#25488](https://github.com/hashicorp/terraform-provider-aws/issues/25488))
* resource/aws_cloudwatch_composite_alarm: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_cloudwatch_event_bus: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_cloudwatch_event_rule: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_cloudwatch_metric_alarm: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_cloudwatch_metric_stream: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_cognito_user_pool: Correctly handle missing or empty `account_recovery_setting` attribute ([#25184](https://github.com/hashicorp/terraform-provider-aws/issues/25184))
* resource/aws_ecr_repository: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_ecs_capacity_provider: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_ecs_cluster: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_ecs_service: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_ecs_task_definition: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_ecs_task_set: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_elasticache_cluster: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_elasticache_parameter_group: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_elasticache_replication_group: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_elasticache_subnet_group: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_elasticache_user: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_elasticache_user_group: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_instance_profile: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_openid_connect_provider: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_policy: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_role: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_saml_provider: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_server_certificate: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_service_linked_role: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_user: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_iam_virtual_mfa_device: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_keyspaces_table: Relax validation of the `schema_definition.column.type` argument to allow collection types ([#25230](https://github.com/hashicorp/terraform-provider-aws/issues/25230))
* resource/aws_launch_configuration: Remove default value for `associate_public_ip_address` argument and mark as Computed. This fixes a regression introduced in [v4.17.0](https://github.com/hashicorp/terraform-provider-aws/blob/main/CHANGELOG.md#4170-june--3-2022) via [#17695](https://github.com/hashicorp/terraform-provider-aws/issues/17695) when no value is configured, whilst honoring any configured value ([#25450](https://github.com/hashicorp/terraform-provider-aws/issues/25450))
* resource/aws_lb: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_lb_listener: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_lb_listener_rule: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_lb_target_group: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_sns_topic: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))
* resource/aws_sqs_queue: Prevent ISO-partition tagging precautions from eating legit errors ([#25549](https://github.com/hashicorp/terraform-provider-aws/issues/25549))

## 4.19.0 (June 17, 2022)

FEATURES:

* **New Resource:** `aws_kendra_index` ([#24920](https://github.com/hashicorp/terraform-provider-aws/issues/24920))
* **New Resource:** `aws_lightsail_container_service` ([#20625](https://github.com/hashicorp/terraform-provider-aws/issues/20625))
* **New Resource:** `aws_lightsail_container_service_deployment_version` ([#20625](https://github.com/hashicorp/terraform-provider-aws/issues/20625))

BUG FIXES:

* resource/aws_dynamodb_table_item: Fix to remove attribute from table item on update ([#25326](https://github.com/hashicorp/terraform-provider-aws/issues/25326))
* resource/aws_ec2_managed_prefix_list_entry: Fix error when attempting to create or delete multiple list entries ([#25046](https://github.com/hashicorp/terraform-provider-aws/issues/25046))

## 4.18.0 (June 10, 2022)

FEATURES:

* **New Resource:** `aws_ce_anomaly_monitor` ([#25177](https://github.com/hashicorp/terraform-provider-aws/issues/25177))
* **New Resource:** `aws_emrserverless_application` ([#25144](https://github.com/hashicorp/terraform-provider-aws/issues/25144))

ENHANCEMENTS:

* data-source/aws_cloudwatch_logs_groups: Make `log_group_name_prefix` optional ([#25187](https://github.com/hashicorp/terraform-provider-aws/issues/25187))
* data-source/aws_cognito_user_pool_client: Add `enable_propagate_additional_user_context_data` argument ([#25181](https://github.com/hashicorp/terraform-provider-aws/issues/25181))
* data-source/aws_ram_resource_share: Add `resource_share_status` argument. ([#25159](https://github.com/hashicorp/terraform-provider-aws/issues/25159))
* resource/aws_cognito_user_pool_client: Add `enable_propagate_additional_user_context_data` argument ([#25181](https://github.com/hashicorp/terraform-provider-aws/issues/25181))
* resource/aws_ebs_snapshot_copy: Add support for `timeouts` configuration block. ([#20912](https://github.com/hashicorp/terraform-provider-aws/issues/20912))
* resource/aws_ebs_volume: Add `final_snapshot` argument ([#21916](https://github.com/hashicorp/terraform-provider-aws/issues/21916))
* resource/aws_s3_bucket: Add error handling for `ErrCodeNotImplemented` and `ErrCodeXNotImplemented` errors when ready bucket information. ([#24764](https://github.com/hashicorp/terraform-provider-aws/issues/24764))
* resource/aws_vpc_ipam_pool_cidr_allocation: improve internal search mechanism ([#25257](https://github.com/hashicorp/terraform-provider-aws/issues/25257))

BUG FIXES:

* resource/aws_snapshot_create_volume_permission: Error if `account_id` is the snapshot's owner ([#12103](https://github.com/hashicorp/terraform-provider-aws/issues/12103))
* resource/aws_ssm_parameter: Allow `Intelligent-Tiering` to upgrade to `Advanced` tier as needed. ([#25174](https://github.com/hashicorp/terraform-provider-aws/issues/25174))

## 4.17.1 (June  3, 2022)

BUG FIXES:

* resource/aws_ram_resource_share: Fix regression in v4.17.0 where `permission_arns` would get clobbered if already set ([#25158](https://github.com/hashicorp/terraform-provider-aws/issues/25158))

## 4.17.0 (June  3, 2022)

FEATURES:

* **New Data Source:** `aws_redshift_cluster_credentials` ([#25092](https://github.com/hashicorp/terraform-provider-aws/issues/25092))
* **New Resource:** `aws_acmpca_policy` ([#25109](https://github.com/hashicorp/terraform-provider-aws/issues/25109))
* **New Resource:** `aws_redshift_cluster_iam_roles` ([#25096](https://github.com/hashicorp/terraform-provider-aws/issues/25096))
* **New Resource:** `aws_redshift_hsm_configuration` ([#25093](https://github.com/hashicorp/terraform-provider-aws/issues/25093))
* **New Resource:** `aws_redshiftdata_statement` ([#25104](https://github.com/hashicorp/terraform-provider-aws/issues/25104))

ENHANCEMENTS:

* resource/aws_dms_endpoint: Add `redshift_settings` configuration block ([#21846](https://github.com/hashicorp/terraform-provider-aws/issues/21846))
* resource/aws_dms_endpoint: Add ability to use AWS Secrets Manager with the `aurora-postgresql` and `mongodb` engines ([#23691](https://github.com/hashicorp/terraform-provider-aws/issues/23691))
* resource/aws_dms_endpoint: Add ability to use AWS Secrets Manager with the `aurora`, `mariadb` and `mysql` engines ([#24846](https://github.com/hashicorp/terraform-provider-aws/issues/24846))
* resource/aws_dms_endpoint: Add ability to use AWS Secrets Manager with the `redshift` engine ([#25080](https://github.com/hashicorp/terraform-provider-aws/issues/25080))
* resource/aws_dms_endpoint: Add ability to use AWS Secrets Manager with the `sqlserver` engine ([#22646](https://github.com/hashicorp/terraform-provider-aws/issues/22646))
* resource/aws_guardduty_detector: Add `kubernetes` attribute to the `datasources` configuration block ([#22859](https://github.com/hashicorp/terraform-provider-aws/issues/22859))
* resource/aws_ram_resource_share: Add `permission_arns` argument. ([#25113](https://github.com/hashicorp/terraform-provider-aws/issues/25113))
* resource/aws_redshift_cluster: The `default_iam_role_arn` argument is now Computed ([#25096](https://github.com/hashicorp/terraform-provider-aws/issues/25096))

BUG FIXES:

* data-source/aws_launch_configuration: Correct data type for `ebs_block_device.throughput` and `root_block_device.throughput` attributes ([#25097](https://github.com/hashicorp/terraform-provider-aws/issues/25097))
* resource/aws_db_instance_role_association: Extend timeout to 10 minutes ([#25145](https://github.com/hashicorp/terraform-provider-aws/issues/25145))
* resource/aws_ebs_volume: Fix to preserve `iops` when changing EBS volume type (`io1`, `io2`, `gp3`) ([#23280](https://github.com/hashicorp/terraform-provider-aws/issues/23280))
* resource/aws_launch_configuration: Honor associate_public_ip_address = false ([#17695](https://github.com/hashicorp/terraform-provider-aws/issues/17695))
* resource/aws_rds_cluster_role_association: Extend timeout to 10 minutes ([#25145](https://github.com/hashicorp/terraform-provider-aws/issues/25145))
* resource/aws_servicecatalog_provisioned_product: Correctly handle resources in a `TAINTED` state ([#25130](https://github.com/hashicorp/terraform-provider-aws/issues/25130))

## 4.16.0 (May 27, 2022)

FEATURES:

* **New Data Source:** `aws_location_place_index` ([#24980](https://github.com/hashicorp/terraform-provider-aws/issues/24980))
* **New Data Source:** `aws_redshift_subnet_group` ([#25053](https://github.com/hashicorp/terraform-provider-aws/issues/25053))
* **New Resource:** `aws_efs_replication_configuration` ([#22844](https://github.com/hashicorp/terraform-provider-aws/issues/22844))
* **New Resource:** `aws_location_place_index` ([#24821](https://github.com/hashicorp/terraform-provider-aws/issues/24821))
* **New Resource:** `aws_redshift_authentication_profile` ([#24907](https://github.com/hashicorp/terraform-provider-aws/issues/24907))
* **New Resource:** `aws_redshift_endpoint_access` ([#25073](https://github.com/hashicorp/terraform-provider-aws/issues/25073))
* **New Resource:** `aws_redshift_hsm_client_certificate` ([#24906](https://github.com/hashicorp/terraform-provider-aws/issues/24906))
* **New Resource:** `aws_redshift_usage_limit` ([#24916](https://github.com/hashicorp/terraform-provider-aws/issues/24916))

ENHANCEMENTS:

* data-source/aws_ami: Add `tpm_support` attribute ([#25045](https://github.com/hashicorp/terraform-provider-aws/issues/25045))
* data-source/aws_redshift_cluster: Add `aqua_configuration_status` attribute. ([#24856](https://github.com/hashicorp/terraform-provider-aws/issues/24856))
* data-source/aws_redshift_cluster: Add `arn`, `cluster_nodes`, `cluster_nodes`, `maintenance_track_name`, `manual_snapshot_retention_period`, `log_destination_type`, and `log_exports` attributes. ([#24982](https://github.com/hashicorp/terraform-provider-aws/issues/24982))
* data-source/aws_cloudfront_response_headers_policy: Add `server_timing_headers_config` attribute ([#24913](https://github.com/hashicorp/terraform-provider-aws/issues/24913))
* resource/aws_ami: Add `tpm_support` argument ([#25045](https://github.com/hashicorp/terraform-provider-aws/issues/25045))
* resource/aws_ami_copy: Add `tpm_support` argument ([#25045](https://github.com/hashicorp/terraform-provider-aws/issues/25045))
* resource/aws_ami_from_instance: Add `tpm_support` argument ([#25045](https://github.com/hashicorp/terraform-provider-aws/issues/25045))
* resource/aws_autoscaling_group: Add `context` argument ([#24951](https://github.com/hashicorp/terraform-provider-aws/issues/24951))
* resource/aws_autoscaling_group: Add `mixed_instances_policy.launch_template.override.instance_requirements` argument ([#24795](https://github.com/hashicorp/terraform-provider-aws/issues/24795))
* resource/aws_cloudfront_response_headers_policy: Add `server_timing_headers_config` argument ([#24913](https://github.com/hashicorp/terraform-provider-aws/issues/24913))
* resource/aws_cloudsearch_domain: Add `index_field.source_fields` argument ([#24915](https://github.com/hashicorp/terraform-provider-aws/issues/24915))
* resource/aws_cloudwatch_metric_stream: Add `statistics_configuration` argument ([#24882](https://github.com/hashicorp/terraform-provider-aws/issues/24882))
* resource/aws_elasticache_global_replication_group: Add support for upgrading `engine_version`. ([#25077](https://github.com/hashicorp/terraform-provider-aws/issues/25077))
* resource/aws_msk_cluster: Support multiple attribute updates by refreshing `current_version` after each update ([#25062](https://github.com/hashicorp/terraform-provider-aws/issues/25062))
* resource/aws_redshift_cluster: Add `aqua_configuration_status` and `apply_immediately` arguments. ([#24856](https://github.com/hashicorp/terraform-provider-aws/issues/24856))
* resource/aws_redshift_cluster: Add `default_iam_role_arn`, `maintenance_track_name`, and `manual_snapshot_retention_period` arguments. ([#24982](https://github.com/hashicorp/terraform-provider-aws/issues/24982))
* resource/aws_redshift_cluster: Add `logging.log_destination_type` and `logging.log_exports` arguments. ([#24886](https://github.com/hashicorp/terraform-provider-aws/issues/24886))
* resource/aws_redshift_cluster: Add plan-time validation for `iam_roles`, `owner_account`, and `port`. ([#24856](https://github.com/hashicorp/terraform-provider-aws/issues/24856))
* resource/aws_redshift_event_subscription: Add plan time validations for `event_categories`, `source_type`, and `severity`. ([#24909](https://github.com/hashicorp/terraform-provider-aws/issues/24909))
* resource/aws_transfer_server: Add support for `TransferSecurityPolicy-2022-03` `security_policy_name` value ([#25060](https://github.com/hashicorp/terraform-provider-aws/issues/25060))

BUG FIXES:

* resource/aws_appflow_flow: Amend `task_properties` validation to avoid conflicting type assumption ([#24889](https://github.com/hashicorp/terraform-provider-aws/issues/24889))
* resource/aws_db_proxy_target: Fix `InvalidDBInstanceState: DB Instance is in an unsupported state - CREATING, needs to be in [AVAILABLE, MODIFYING, BACKING_UP]` error on resource Create ([#24875](https://github.com/hashicorp/terraform-provider-aws/issues/24875))
* resource/aws_instance: Correctly delete instance on destroy when `disable_api_termination` is `true` ([#19277](https://github.com/hashicorp/terraform-provider-aws/issues/19277))
* resource/aws_instance: Prevent error `InvalidParameterCombination: The parameter GroupName within placement information cannot be specified when instanceInterruptionBehavior is set to 'STOP'` when using a launch template that sets `instance_interruption_behavior` to `stop` ([#24695](https://github.com/hashicorp/terraform-provider-aws/issues/24695))
* resource/aws_msk_cluster: Prevent crash on apply when `client_authentication.tls` is empty ([#25072](https://github.com/hashicorp/terraform-provider-aws/issues/25072))
* resource/aws_servicecatalog_provisioned_product: Add possible `TAINTED` target state for resource update and remove one of the internal waiters during read ([#24804](https://github.com/hashicorp/terraform-provider-aws/issues/24804))

## 4.15.1 (May 20, 2022)

BUG FIXES:

* resource/aws_organizations_account: Fix reading account state for existing accounts ([#24899](https://github.com/hashicorp/terraform-provider-aws/issues/24899))

## 4.15.0 (May 20, 2022)

BREAKING CHANGES:

* resource/aws_msk_cluster: The `ebs_volume_size` argument is deprecated in favor of the `storage_info` block. The `storage_info` block can set `volume_size` and `provisioned_throughput` ([#24767](https://github.com/hashicorp/terraform-provider-aws/issues/24767))

FEATURES:

* **New Data Source:** `aws_lb_hosted_zone_id` ([#24749](https://github.com/hashicorp/terraform-provider-aws/issues/24749))
* **New Data Source:** `aws_networkmanager_core_network_policy_document` ([#24368](https://github.com/hashicorp/terraform-provider-aws/issues/24368))
* **New Resource:** `aws_db_snapshot_copy` ([#9886](https://github.com/hashicorp/terraform-provider-aws/issues/9886))
* **New Resource:** `aws_keyspaces_table` ([#24351](https://github.com/hashicorp/terraform-provider-aws/issues/24351))

ENHANCEMENTS:

* data-source/aws_route53_resolver_rules: add `name_regex` argument ([#24582](https://github.com/hashicorp/terraform-provider-aws/issues/24582))
* resource/aws_autoscaling_group: Add `instance_refresh.preferences.skip_matching` argument ([#23059](https://github.com/hashicorp/terraform-provider-aws/issues/23059))
* resource/aws_autoscaling_policy: Add `enabled` argument ([#12625](https://github.com/hashicorp/terraform-provider-aws/issues/12625))
* resource/aws_ec2_fleet: Add `arn` attribute ([#24732](https://github.com/hashicorp/terraform-provider-aws/issues/24732))
* resource/aws_ec2_fleet: Add `launch_template_config.override.instance_requirements` argument ([#24732](https://github.com/hashicorp/terraform-provider-aws/issues/24732))
* resource/aws_ec2_fleet: Add support for `capacity-optimized` and `capacity-optimized-prioritized` values for `spot_options.allocation_strategy` ([#24732](https://github.com/hashicorp/terraform-provider-aws/issues/24732))
* resource/aws_lambda_function: Add support for `nodejs16.x` `runtime` value ([#24768](https://github.com/hashicorp/terraform-provider-aws/issues/24768))
* resource/aws_lambda_layer_version: Add support for `nodejs16.x` `compatible_runtimes` value ([#24768](https://github.com/hashicorp/terraform-provider-aws/issues/24768))
* resource/aws_organizations_account: Add `create_govcloud` argument and `govcloud_id` attribute ([#24447](https://github.com/hashicorp/terraform-provider-aws/issues/24447))
* resource/aws_s3_bucket_website_configuration: Add `routing_rules` parameter to be used instead of `routing_rule` to support configurations with empty String values ([#24198](https://github.com/hashicorp/terraform-provider-aws/issues/24198))

BUG FIXES:

* resource/aws_autoscaling_group: Wait for correct number of ELBs when `wait_for_elb_capacity` is configured ([#20806](https://github.com/hashicorp/terraform-provider-aws/issues/20806))
* resource/aws_elasticache_replication_group: Fix perpetual diff on `auto_minor_version_upgrade` ([#24688](https://github.com/hashicorp/terraform-provider-aws/issues/24688))

## 4.14.0 (May 13, 2022)

FEATURES:

* **New Data Source:** `aws_connect_routing_profile` ([#23525](https://github.com/hashicorp/terraform-provider-aws/issues/23525))
* **New Data Source:** `aws_connect_security_profile` ([#23524](https://github.com/hashicorp/terraform-provider-aws/issues/23524))
* **New Data Source:** `aws_connect_user_hierarchy_structure` ([#23527](https://github.com/hashicorp/terraform-provider-aws/issues/23527))
* **New Data Source:** `aws_location_map` ([#24693](https://github.com/hashicorp/terraform-provider-aws/issues/24693))
* **New Resource:** `aws_appflow_connector_profile` ([#23892](https://github.com/hashicorp/terraform-provider-aws/issues/23892))
* **New Resource:** `aws_appflow_flow` ([#24017](https://github.com/hashicorp/terraform-provider-aws/issues/24017))
* **New Resource:** `aws_appintegrations_event_integration` ([#23904](https://github.com/hashicorp/terraform-provider-aws/issues/23904))
* **New Resource:** `aws_connect_user_hierarchy_group` ([#23531](https://github.com/hashicorp/terraform-provider-aws/issues/23531))
* **New Resource:** `aws_location_map` ([#24682](https://github.com/hashicorp/terraform-provider-aws/issues/24682))

ENHANCEMENTS:

* data-source/aws_acm_certificate: Add `certificate` and `certificate_chain` attributes ([#24593](https://github.com/hashicorp/terraform-provider-aws/issues/24593))
* data-source/aws_autoscaling_group: Add `enabled_metrics` attribute ([#24691](https://github.com/hashicorp/terraform-provider-aws/issues/24691))
* data-source/aws_codestarconnections_connection: Support lookup by `name` ([#19262](https://github.com/hashicorp/terraform-provider-aws/issues/19262))
* data-source/aws_launch_template: Add `instance_requirements` attribute ([#24543](https://github.com/hashicorp/terraform-provider-aws/issues/24543))
* resource/aws_ebs_volume: Add support for `multi_attach_enabled` with `io2` volumes ([#19060](https://github.com/hashicorp/terraform-provider-aws/issues/19060))
* resource/aws_launch_template: Add `instance_requirements` argument ([#24543](https://github.com/hashicorp/terraform-provider-aws/issues/24543))
* resource/aws_servicecatalog_provisioned_product: Wait for provisioning to finish ([#24758](https://github.com/hashicorp/terraform-provider-aws/issues/24758))
* resource/aws_servicecatalog_provisioned_product: Wait for update to finish ([#24758](https://github.com/hashicorp/terraform-provider-aws/issues/24758))
* resource/aws_spot_fleet_request: Add `overrides.instance_requirements` argument ([#24448](https://github.com/hashicorp/terraform-provider-aws/issues/24448))

BUG FIXES:

* resource/aws_alb_listener_rule: Don't force recreate listener rule on priority change. ([#23768](https://github.com/hashicorp/terraform-provider-aws/issues/23768))
* resource/aws_default_subnet: Fix `InvalidSubnet.Conflict` errors when associating IPv6 CIDR blocks ([#24685](https://github.com/hashicorp/terraform-provider-aws/issues/24685))
* resource/aws_ebs_volume: Add configurable timeouts ([#24745](https://github.com/hashicorp/terraform-provider-aws/issues/24745))
* resource/aws_imagebuilder_image_recipe: Fix `ResourceDependencyException` errors when a dependency is modified ([#24708](https://github.com/hashicorp/terraform-provider-aws/issues/24708))
* resource/aws_kms_key: Retry on `MalformedPolicyDocumentException` errors when updating key policy ([#24697](https://github.com/hashicorp/terraform-provider-aws/issues/24697))
* resource/aws_servicecatalog_provisioned_product: Prevent error when retrieving a provisioned product in a non-available state ([#24758](https://github.com/hashicorp/terraform-provider-aws/issues/24758))
* resource/aws_subnet: Fix `InvalidSubnet.Conflict` errors when associating IPv6 CIDR blocks ([#24685](https://github.com/hashicorp/terraform-provider-aws/issues/24685))

## 4.13.0 (May  5, 2022)

FEATURES:

* **New Data Source:** `aws_emrcontainers_virtual_cluster` ([#20003](https://github.com/hashicorp/terraform-provider-aws/issues/20003))
* **New Data Source:** `aws_iam_instance_profiles` ([#24423](https://github.com/hashicorp/terraform-provider-aws/issues/24423))
* **New Data Source:** `aws_secretsmanager_secrets` ([#24514](https://github.com/hashicorp/terraform-provider-aws/issues/24514))
* **New Resource:** `aws_emrcontainers_virtual_cluster` ([#20003](https://github.com/hashicorp/terraform-provider-aws/issues/20003))
* **New Resource:** `aws_iot_topic_rule_destination` ([#24395](https://github.com/hashicorp/terraform-provider-aws/issues/24395))

ENHANCEMENTS:

* data-source/aws_ami: Add `deprecation_time` attribute ([#24489](https://github.com/hashicorp/terraform-provider-aws/issues/24489))
* data-source/aws_msk_cluster: Add `bootstrap_brokers_public_sasl_iam`, `bootstrap_brokers_public_sasl_scram` and `bootstrap_brokers_public_tls` attributes ([#21005](https://github.com/hashicorp/terraform-provider-aws/issues/21005))
* data-source/aws_ssm_patch_baseline: Add the following attributes: `approved_patches`, `approved_patches_compliance_level`, `approval_rule`, `global_filter`, `rejected_patches`, `rejected_patches_action`, `source` ([#24401](https://github.com/hashicorp/terraform-provider-aws/issues/24401))
* resource/aws_ami: Add `deprecation_time` argument ([#24489](https://github.com/hashicorp/terraform-provider-aws/issues/24489))
* resource/aws_ami_copy: Add `deprecation_time` argument ([#24489](https://github.com/hashicorp/terraform-provider-aws/issues/24489))
* resource/aws_ami_from_instance: Add `deprecation_time` argument ([#24489](https://github.com/hashicorp/terraform-provider-aws/issues/24489))
* resource/aws_iot_topic_rule: Add `http` and `error_action.http` arguments ([#16087](https://github.com/hashicorp/terraform-provider-aws/issues/16087))
* resource/aws_iot_topic_rule: Add `kafka` and `error_action.kafka` arguments ([#24395](https://github.com/hashicorp/terraform-provider-aws/issues/24395))
* resource/aws_iot_topic_rule: Add `s3.canned_acl` and `error_action.s3.canned_acl` arguments ([#19175](https://github.com/hashicorp/terraform-provider-aws/issues/19175))
* resource/aws_iot_topic_rule: Add `timestream` and `error_action.timestream` arguments ([#22337](https://github.com/hashicorp/terraform-provider-aws/issues/22337))
* resource/aws_lambda_permission: Add `function_url_auth_type` argument ([#24510](https://github.com/hashicorp/terraform-provider-aws/issues/24510))
* resource/aws_msk_cluster: Add `bootstrap_brokers_public_sasl_iam`, `bootstrap_brokers_public_sasl_scram` and `bootstrap_brokers_public_tls` attributes ([#21005](https://github.com/hashicorp/terraform-provider-aws/issues/21005))
* resource/aws_msk_cluster: Add `broker_node_group_info.connectivity_info` argument to support [public access](https://docs.aws.amazon.com/msk/latest/developerguide/public-access.html) ([#21005](https://github.com/hashicorp/terraform-provider-aws/issues/21005))
* resource/aws_msk_cluster: Add `client_authentication.unauthenticated` argument ([#21005](https://github.com/hashicorp/terraform-provider-aws/issues/21005))
* resource/aws_msk_cluster: Allow in-place update of `client_authentication` and `encryption_info.encryption_in_transit.client_broker` ([#21005](https://github.com/hashicorp/terraform-provider-aws/issues/21005))

BUG FIXES:

* resource/aws_cloudfront_distribution: Fix PreconditionFailed errors when other CloudFront resources are changed before the distribution ([#24537](https://github.com/hashicorp/terraform-provider-aws/issues/24537))
* resource/aws_ecs_service: Fix retry when using the `wait_for_steady_state` parameter ([#24541](https://github.com/hashicorp/terraform-provider-aws/issues/24541))
* resource/aws_launch_template: Fix crash when reading `license_specification` ([#24579](https://github.com/hashicorp/terraform-provider-aws/issues/24579))
* resource/aws_ssm_document: Always include `attachment_sources` when updating SSM documents ([#24530](https://github.com/hashicorp/terraform-provider-aws/issues/24530))

## 4.12.1 (April 29, 2022)

ENHANCEMENTS:

* resource/aws_kms_key: Add support for HMAC_256 customer master key spec ([#24450](https://github.com/hashicorp/terraform-provider-aws/issues/24450))

BUG FIXES:

* resource/aws_acm_certificate_validation: Restore certificate issuance timestamp as the resource `id` value, fixing error on existing resource Read ([#24453](https://github.com/hashicorp/terraform-provider-aws/issues/24453))
* resource/aws_kms_alias: Fix reserved prefix used in `name` and `name_prefix` plan time validation ([#24469](https://github.com/hashicorp/terraform-provider-aws/issues/24469))

## 4.12.0 (April 28, 2022)

FEATURES:

* **New Data Source:** `aws_ce_cost_category` ([#24402](https://github.com/hashicorp/terraform-provider-aws/issues/24402))
* **New Data Source:** `aws_ce_tags` ([#24402](https://github.com/hashicorp/terraform-provider-aws/issues/24402))
* **New Data Source:** `aws_cloudfront_origin_access_identities` ([#24382](https://github.com/hashicorp/terraform-provider-aws/issues/24382))
* **New Data Source:** `aws_mq_broker_instance_type_offerings` ([#24394](https://github.com/hashicorp/terraform-provider-aws/issues/24394))
* **New Resource:** `aws_athena_data_catalog` ([#22968](https://github.com/hashicorp/terraform-provider-aws/issues/22968))
* **New Resource:** `aws_ce_cost_category` ([#24402](https://github.com/hashicorp/terraform-provider-aws/issues/24402))
* **New Resource:** `aws_docdb_event_subscription` ([#24379](https://github.com/hashicorp/terraform-provider-aws/issues/24379))

ENHANCEMENTS:

* data-source/aws_grafana_workspace: Add `tags` attribute ([#24358](https://github.com/hashicorp/terraform-provider-aws/issues/24358))
* data-source/aws_instance: Add `maintenance_options` attribute ([#24377](https://github.com/hashicorp/terraform-provider-aws/issues/24377))
* data-source/aws_launch_template: Add `maintenance_options` attribute ([#24377](https://github.com/hashicorp/terraform-provider-aws/issues/24377))
* provider: Add support for Assume Role with Web Identity. ([#24441](https://github.com/hashicorp/terraform-provider-aws/issues/24441))
* resource/aws_acm_certificate: Add `validation_option` argument ([#3853](https://github.com/hashicorp/terraform-provider-aws/issues/3853))
* resource/aws_acm_certificate_validation: Increase default resource Create (certificate issuance) timeout to 75 minutes ([#20073](https://github.com/hashicorp/terraform-provider-aws/issues/20073))
* resource/aws_emr_cluster: Add `list_steps_states` argument ([#20871](https://github.com/hashicorp/terraform-provider-aws/issues/20871))
* resource/aws_grafana_workspace: Add `tags` argument ([#24358](https://github.com/hashicorp/terraform-provider-aws/issues/24358))
* resource/aws_instance: Add `maintenance_options` argument ([#24377](https://github.com/hashicorp/terraform-provider-aws/issues/24377))
* resource/aws_launch_template: Add `maintenance_options` argument ([#24377](https://github.com/hashicorp/terraform-provider-aws/issues/24377))
* resource/aws_mq_broker: Make `maintenance_window_start_time` updateable without recreation. ([#24385](https://github.com/hashicorp/terraform-provider-aws/issues/24385))
* resource/aws_rds_cluster: Add `serverlessv2_scaling_configuration` argument to support [Aurora Serverless v2](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v2.html) ([#24363](https://github.com/hashicorp/terraform-provider-aws/issues/24363))
* resource/aws_spot_fleet_request: Add `terminate_instances_on_delete` argument ([#17268](https://github.com/hashicorp/terraform-provider-aws/issues/17268))

BUG FIXES:

* data-source/aws_kms_alias: Fix `name` plan time validation ([#13000](https://github.com/hashicorp/terraform-provider-aws/issues/13000))
* provider: Setting `skip_metadata_api_check = false` now overrides `AWS_EC2_METADATA_DISABLED` environment variable. ([#24441](https://github.com/hashicorp/terraform-provider-aws/issues/24441))
* resource/aws_acm_certificate: Correctly handle SAN entries that match `domain_name` ([#20073](https://github.com/hashicorp/terraform-provider-aws/issues/20073))
* resource/aws_dms_replication_task: Fix to stop the task before updating, if required ([#24047](https://github.com/hashicorp/terraform-provider-aws/issues/24047))
* resource/aws_ec2_availability_zone_group: Don't crash if `group_name` is not found ([#24422](https://github.com/hashicorp/terraform-provider-aws/issues/24422))
* resource/aws_elasticache_cluster: Update regex pattern to target specific Redis V6 versions through the `engine_version` attribute ([#23734](https://github.com/hashicorp/terraform-provider-aws/issues/23734))
* resource/aws_elasticache_replication_group: Update regex pattern to target specific Redis V6 versions through the `engine_version` attribute ([#23734](https://github.com/hashicorp/terraform-provider-aws/issues/23734))
* resource/aws_kms_alias: Fix `name` and `name_prefix` plan time validation ([#13000](https://github.com/hashicorp/terraform-provider-aws/issues/13000))
* resource/aws_lb: Fix bug causing an error on update if tags unsupported in ISO region ([#24334](https://github.com/hashicorp/terraform-provider-aws/issues/24334))
* resource/aws_s3_bucket_policy: Let resource be removed from tfstate if bucket deleted outside Terraform ([#23510](https://github.com/hashicorp/terraform-provider-aws/issues/23510))
* resource/aws_s3_bucket_versioning: Let resource be removed from tfstate if bucket deleted outside Terraform ([#23510](https://github.com/hashicorp/terraform-provider-aws/issues/23510))
* resource/aws_ses_receipt_filter: Allow period character (`.`) in `name` argument ([#24383](https://github.com/hashicorp/terraform-provider-aws/issues/24383))

## 4.11.0 (April 22, 2022)

FEATURES:

* **New Data Source:** `aws_s3_bucket_policy` ([#17738](https://github.com/hashicorp/terraform-provider-aws/issues/17738))
* **New Resource:** `aws_transfer_workflow` ([#24248](https://github.com/hashicorp/terraform-provider-aws/issues/24248))

ENHANCEMENTS:

* data-source/aws_imagebuilder_infrastructure_configuration: Add `instance_metadata_options` attribute ([#24285](https://github.com/hashicorp/terraform-provider-aws/issues/24285))
* data-source/aws_opensearch_domain: Add `cold_storage_options` attribute to the `cluster_config` configuration block ([#24284](https://github.com/hashicorp/terraform-provider-aws/issues/24284))
* resource/aws_db_proxy: Add `auth.username` argument ([#24264](https://github.com/hashicorp/terraform-provider-aws/issues/24264))
* resource/aws_elasticache_user: Add plan-time validation of password argumnet length ([#24274](https://github.com/hashicorp/terraform-provider-aws/issues/24274))
* resource/aws_elasticsearch_domain: For Elasticsearch versions 6.7+, allow in-place update of `node_to_node_encryption.0.enabled` and `encrypt_at_rest.0.enabled`. ([#24222](https://github.com/hashicorp/terraform-provider-aws/issues/24222))
* resource/aws_fsx_ontap_file_system: Add support for `SINGLE_AZ_1` `deployment_type`. ([#24280](https://github.com/hashicorp/terraform-provider-aws/issues/24280))
* resource/aws_imagebuilder_infrastructure_configuration: Add `instance_metadata_options` argument ([#24285](https://github.com/hashicorp/terraform-provider-aws/issues/24285))
* resource/aws_instance: Add `capacity_reservation_specification.capacity_reservation_target.capacity_reservation_resource_group_arn` argument ([#24283](https://github.com/hashicorp/terraform-provider-aws/issues/24283))
* resource/aws_instance: Add `network_interface.network_card_index` argument ([#24283](https://github.com/hashicorp/terraform-provider-aws/issues/24283))
* resource/aws_opensearch_domain: Add `cold_storage_options` argument to the `cluster_config` configuration block ([#24284](https://github.com/hashicorp/terraform-provider-aws/issues/24284))
* resource/aws_opensearch_domain: For Elasticsearch versions 6.7+, allow in-place update of `node_to_node_encryption.0.enabled` and `encrypt_at_rest.0.enabled`. ([#24222](https://github.com/hashicorp/terraform-provider-aws/issues/24222))
* resource/aws_transfer_server: Add `workflow_details` argument ([#24248](https://github.com/hashicorp/terraform-provider-aws/issues/24248))
* resource/aws_waf_byte_match_set: Additional supported values for `byte_match_tuples.field_to_match.type` argument ([#24286](https://github.com/hashicorp/terraform-provider-aws/issues/24286))
* resource/aws_wafregional_web_acl: Additional supported values for `logging_configuration.redacted_fields.field_to_match.type` argument ([#24286](https://github.com/hashicorp/terraform-provider-aws/issues/24286))
* resource/aws_workspaces_workspace: Additional supported values for `workspace_properties.compute_type_name` argument ([#24286](https://github.com/hashicorp/terraform-provider-aws/issues/24286))

BUG FIXES:

* data-source/aws_db_instance: Prevent panic when setting instance connection endpoint values ([#24299](https://github.com/hashicorp/terraform-provider-aws/issues/24299))
* data-source/aws_efs_file_system: Prevent panic when searching by tag returns 0 or multiple results ([#24298](https://github.com/hashicorp/terraform-provider-aws/issues/24298))
* data-source/aws_elasticache_cluster: Gracefully handle additional tagging error type in non-standard AWS partitions (i.e., ISO) ([#24275](https://github.com/hashicorp/terraform-provider-aws/issues/24275))
* resource/aws_appstream_user_stack_association: Prevent panic during resource read ([#24303](https://github.com/hashicorp/terraform-provider-aws/issues/24303))
* resource/aws_cloudformation_stack_set: Prevent `Validation` errors when `operation_preferences.failure_tolerance_count` is zero ([#24250](https://github.com/hashicorp/terraform-provider-aws/issues/24250))
* resource/aws_elastic_beanstalk_environment: Correctly set `cname_prefix` attribute ([#24278](https://github.com/hashicorp/terraform-provider-aws/issues/24278))
* resource/aws_elasticache_cluster: Gracefully handle additional tagging error type in non-standard AWS partitions (i.e., ISO) ([#24275](https://github.com/hashicorp/terraform-provider-aws/issues/24275))
* resource/aws_elasticache_parameter_group: Gracefully handle additional tagging error type in non-standard AWS partitions (i.e., ISO) ([#24275](https://github.com/hashicorp/terraform-provider-aws/issues/24275))
* resource/aws_elasticache_replication_group: Gracefully handle additional tagging error type in non-standard AWS partitions (i.e., ISO) ([#24275](https://github.com/hashicorp/terraform-provider-aws/issues/24275))
* resource/aws_elasticache_subnet_group: Gracefully handle additional tagging error type in non-standard AWS partitions (i.e., ISO) ([#24275](https://github.com/hashicorp/terraform-provider-aws/issues/24275))
* resource/aws_elasticache_user: Gracefully handle additional tagging error type in non-standard AWS partitions (i.e., ISO) ([#24275](https://github.com/hashicorp/terraform-provider-aws/issues/24275))
* resource/aws_elasticache_user_group: Gracefully handle additional tagging error type in non-standard AWS partitions (i.e., ISO) ([#24275](https://github.com/hashicorp/terraform-provider-aws/issues/24275))
* resource/aws_instance: Fix issue with assuming Placement and disableApiTermination instance attributes exist when managing a Snowball Edge device ([#19256](https://github.com/hashicorp/terraform-provider-aws/issues/19256))
* resource/aws_kinesis_firehose_delivery_stream: Increase the maximum length of the `processing_configuration.processors.parameters.parameter_value` argument's value to `5120` ([#24312](https://github.com/hashicorp/terraform-provider-aws/issues/24312))
* resource/aws_macie2_member: Correct type for `invitation_disable_email_notification` parameter ([#24304](https://github.com/hashicorp/terraform-provider-aws/issues/24304))
* resource/aws_s3_bucket_server_side_encryption_configuration: Retry on `ServerSideEncryptionConfigurationNotFoundError` errors due to eventual consistency ([#24266](https://github.com/hashicorp/terraform-provider-aws/issues/24266))
* resource/aws_sfn_state_machine: Prevent panic during resource update ([#24302](https://github.com/hashicorp/terraform-provider-aws/issues/24302))
* resource/aws_shield_protection_group: When updating resource tags, use the `protection_group_arn` parameter instead of `arn`. ([#24296](https://github.com/hashicorp/terraform-provider-aws/issues/24296))
* resource/aws_ssm_association: Prevent panic when `wait_for_success_timeout_seconds` is configured ([#24300](https://github.com/hashicorp/terraform-provider-aws/issues/24300))

## 4.10.0 (April 14, 2022)

FEATURES:

* **New Data Source:** `aws_iam_saml_provider` ([#10498](https://github.com/hashicorp/terraform-provider-aws/issues/10498))
* **New Data Source:** `aws_nat_gateways` ([#24190](https://github.com/hashicorp/terraform-provider-aws/issues/24190))
* **New Resource:** `aws_datasync_location_fsx_openzfs_file_system` ([#24200](https://github.com/hashicorp/terraform-provider-aws/issues/24200))
* **New Resource:** `aws_elasticache_user_group_association` ([#24204](https://github.com/hashicorp/terraform-provider-aws/issues/24204))
* **New Resource:** `aws_qldb_stream` ([#19297](https://github.com/hashicorp/terraform-provider-aws/issues/19297))

ENHANCEMENTS:

* data-source/aws_qldb_ledger: Add `kms_key` and `tags` attributes ([#19297](https://github.com/hashicorp/terraform-provider-aws/issues/19297))
* resource/aws_ami_launch_permission: Add `group` argument ([#20677](https://github.com/hashicorp/terraform-provider-aws/issues/20677))
* resource/aws_ami_launch_permission: Add `organization_arn` and `organizational_unit_arn` arguments ([#21694](https://github.com/hashicorp/terraform-provider-aws/issues/21694))
* resource/aws_athena_database: Add `properties` argument. ([#24172](https://github.com/hashicorp/terraform-provider-aws/issues/24172))
* resource/aws_athena_database: Add import support. ([#24172](https://github.com/hashicorp/terraform-provider-aws/issues/24172))
* resource/aws_config_config_rule: Add `source.custom_policy_details` argument. ([#24057](https://github.com/hashicorp/terraform-provider-aws/issues/24057))
* resource/aws_config_config_rule: Add plan time validation for `source.source_detail.event_source` and `source.source_detail.message_type`. ([#24057](https://github.com/hashicorp/terraform-provider-aws/issues/24057))
* resource/aws_config_config_rule: Make `source.source_identifier` optional. ([#24057](https://github.com/hashicorp/terraform-provider-aws/issues/24057))
* resource/aws_eks_addon: Add `preserve` argument ([#24218](https://github.com/hashicorp/terraform-provider-aws/issues/24218))
* resource/aws_grafana_workspace: Add plan time validations for `authentication_providers`, `authentication_providers`, `authentication_providers`. ([#24170](https://github.com/hashicorp/terraform-provider-aws/issues/24170))
* resource/aws_qldb_ledger: Add `kms_key` argument ([#19297](https://github.com/hashicorp/terraform-provider-aws/issues/19297))
* resource/aws_vpc_ipam_scope: Add pagination when describing IPAM Scopes ([#24188](https://github.com/hashicorp/terraform-provider-aws/issues/24188))

BUG FIXES:

* resource/aws_athena_database: Add drift detection for `comment`. ([#24172](https://github.com/hashicorp/terraform-provider-aws/issues/24172))
* resource/aws_cloudformation_stack_set: Prevent `InvalidParameter` errors when updating `operation_preferences` ([#24202](https://github.com/hashicorp/terraform-provider-aws/issues/24202))
* resource/aws_cloudwatch_event_connection: Add validation to `auth_parameters.api_key.key`, `auth_parameters.api_key.value`, `auth_parameters.basic.username`, `auth_parameters.basic.password`, `auth_parameters.oauth.authorization_endpoint`, `auth_parameters.oauth.client_parameters.client_id` and `auth_parameters.oauth.client_parameters.client_secret` arguments ([#24154](https://github.com/hashicorp/terraform-provider-aws/issues/24154))
* resource/aws_cloudwatch_log_subscription_filter: Retry resource create and update when a conflicting operation error is returned ([#24148](https://github.com/hashicorp/terraform-provider-aws/issues/24148))
* resource/aws_ecs_service: Retry when using the `wait_for_steady_state` parameter and `ResourceNotReady` errors are returned from the AWS API ([#24223](https://github.com/hashicorp/terraform-provider-aws/issues/24223))
* resource/aws_ecs_service: Wait for service to reach an active state after create and update operations ([#24223](https://github.com/hashicorp/terraform-provider-aws/issues/24223))
* resource/aws_emr_cluster: Ignore `UnknownOperationException` errors when reading a cluster's auto-termination policy ([#24237](https://github.com/hashicorp/terraform-provider-aws/issues/24237))
* resource/aws_lambda_function_url: Ignore `ResourceConflictException` errors caused by existing `FunctionURLAllowPublicAccess` permission statements ([#24220](https://github.com/hashicorp/terraform-provider-aws/issues/24220))
* resource/aws_vpc_ipam_scope: Prevent panic when describing IPAM Scopes by ID ([#24188](https://github.com/hashicorp/terraform-provider-aws/issues/24188))

## 4.9.0 (April 07, 2022)

NOTES:

* resource/aws_s3_bucket: The `acceleration_status`, `acl`, `cors_rule`, `grant`, `lifecycle_rule`, `logging`, `object_lock_configuration.rule`, `policy`, `replication_configuration`, `request_payer`, `server_side_encryption_configuration`, `versioning`, and `website` parameters are now Optional. Please refer to the documentation for details on drift detection and potential conflicts when configuring these parameters with the standalone `aws_s3_bucket_*` resources. ([#23985](https://github.com/hashicorp/terraform-provider-aws/issues/23985))

FEATURES:

* **New Data Source:** `aws_eks_addon_version` ([#23157](https://github.com/hashicorp/terraform-provider-aws/issues/23157))
* **New Data Source:** `aws_lambda_function_url` ([#24053](https://github.com/hashicorp/terraform-provider-aws/issues/24053))
* **New Data Source:** `aws_memorydb_acl` ([#23891](https://github.com/hashicorp/terraform-provider-aws/issues/23891))
* **New Data Source:** `aws_memorydb_cluster` ([#23991](https://github.com/hashicorp/terraform-provider-aws/issues/23991))
* **New Data Source:** `aws_memorydb_snapshot` ([#23990](https://github.com/hashicorp/terraform-provider-aws/issues/23990))
* **New Data Source:** `aws_memorydb_user` ([#23890](https://github.com/hashicorp/terraform-provider-aws/issues/23890))
* **New Data Source:** `aws_opensearch_domain` ([#23902](https://github.com/hashicorp/terraform-provider-aws/issues/23902))
* **New Data Source:** `aws_ssm_maintenance_windows` ([#24011](https://github.com/hashicorp/terraform-provider-aws/issues/24011))
* **New Resource:** `aws_db_instance_automated_backups_replication` ([#23759](https://github.com/hashicorp/terraform-provider-aws/issues/23759))
* **New Resource:** `aws_dynamodb_contributor_insights` ([#23947](https://github.com/hashicorp/terraform-provider-aws/issues/23947))
* **New Resource:** `aws_iot_indexing_configuration` ([#9929](https://github.com/hashicorp/terraform-provider-aws/issues/9929))
* **New Resource:** `aws_iot_logging_options` ([#13392](https://github.com/hashicorp/terraform-provider-aws/issues/13392))
* **New Resource:** `aws_iot_provisioning_template` ([#12108](https://github.com/hashicorp/terraform-provider-aws/issues/12108))
* **New Resource:** `aws_lambda_function_url` ([#24053](https://github.com/hashicorp/terraform-provider-aws/issues/24053))
* **New Resource:** `aws_opensearch_domain` ([#23902](https://github.com/hashicorp/terraform-provider-aws/issues/23902))
* **New Resource:** `aws_opensearch_domain_policy` ([#23902](https://github.com/hashicorp/terraform-provider-aws/issues/23902))
* **New Resource:** `aws_opensearch_domain_saml_options` ([#23902](https://github.com/hashicorp/terraform-provider-aws/issues/23902))
* **New Resource:** `aws_rds_cluster_activity_stream` ([#22097](https://github.com/hashicorp/terraform-provider-aws/issues/22097))

ENHANCEMENTS:

* data-source/aws_imagebuilder_distribution_configuration: Add `account_id` attribute to the `launch_template_configuration` attribute of the `distribution` configuration block ([#23924](https://github.com/hashicorp/terraform-provider-aws/issues/23924))
* data-source/aws_route: Add `core_network_arn` argument ([#24024](https://github.com/hashicorp/terraform-provider-aws/issues/24024))
* data-source/aws_route_table: Add 'routes.core_network_arn' attribute' ([#24024](https://github.com/hashicorp/terraform-provider-aws/issues/24024))
* provider: Add support for reading custom CA bundle setting from shared config files ([#24064](https://github.com/hashicorp/terraform-provider-aws/issues/24064))
* resource/aws_cloudformation_stack_set: Add `operation_preferences` argument ([#23908](https://github.com/hashicorp/terraform-provider-aws/issues/23908))
* resource/aws_default_route_table: Add `core_network_arn` argument to the `route` configuration block ([#24024](https://github.com/hashicorp/terraform-provider-aws/issues/24024))
* resource/aws_dlm_lifecycle_policy: Add `policy_details.schedule.create_rule.cron_expression`, `policy_details.schedule.retain_rule.interval`, `policy_details.schedule.retain_rule.interval_unit`, `policy_details.policy_type`, `policy_details.schedule.deprecate_rule`, `policy_details.parameters`, `policy_details.schedule.variable_tags`, `policy_details.schedule.fast_restore_rule`, `policy_details.schedule.share_rule`, `policy_details.resource_locations`, `policy_details.schedule.create_rule.location`, `policy_details.action` and `policy_details.event_source` arguments ([#23880](https://github.com/hashicorp/terraform-provider-aws/issues/23880))
* resource/aws_dlm_lifecycle_policy: Add plan time validations for `policy_details.resource_types` and `description` arguments ([#23880](https://github.com/hashicorp/terraform-provider-aws/issues/23880))
* resource/aws_dlm_lifecycle_policy: Make `policy_details.resource_types`, `policy_details.schedule`, `policy_details.target_tags`, `policy_details.schedule.retain_rule` and `policy_details.schedule.create_rule.interval` arguments optional ([#23880](https://github.com/hashicorp/terraform-provider-aws/issues/23880))
* resource/aws_elasticache_cluster: Add `auto_minor_version_upgrade` argument ([#23996](https://github.com/hashicorp/terraform-provider-aws/issues/23996))
* resource/aws_fms_policy: Retry when `InternalErrorException` errors are returned from the AWS API ([#23952](https://github.com/hashicorp/terraform-provider-aws/issues/23952))
* resource/aws_fsx_ontap_file_system: Support updating `storage_capacity`, `throughput_capacity`, and `disk_iops_configuration`. ([#24002](https://github.com/hashicorp/terraform-provider-aws/issues/24002))
* resource/aws_imagebuilder_distribution_configuration: Add `account_id` argument to the `launch_template_configuration` attribute of the `distribution` configuration block ([#23924](https://github.com/hashicorp/terraform-provider-aws/issues/23924))
* resource/aws_iot_authorizer: Add `enable_caching_for_http` argument ([#23993](https://github.com/hashicorp/terraform-provider-aws/issues/23993))
* resource/aws_lambda_permission: Add `principal_org_id` argument. ([#24001](https://github.com/hashicorp/terraform-provider-aws/issues/24001))
* resource/aws_mq_broker: Add validation to `broker_name` and `security_groups` arguments ([#18088](https://github.com/hashicorp/terraform-provider-aws/issues/18088))
* resource/aws_organizations_account: Add `close_on_deletion` argument to close account on deletion ([#23930](https://github.com/hashicorp/terraform-provider-aws/issues/23930))
* resource/aws_route: Add `core_network_arn` argument ([#24024](https://github.com/hashicorp/terraform-provider-aws/issues/24024))
* resource/aws_route_table: Add `core_network_arn` argument to the `route` configuration block ([#24024](https://github.com/hashicorp/terraform-provider-aws/issues/24024))
* resource/aws_s3_bucket: Speed up resource deletion, especially when the S3 buckets contains a large number of objects and `force_destroy` is `true` ([#24020](https://github.com/hashicorp/terraform-provider-aws/issues/24020))
* resource/aws_s3_bucket: Update `acceleration_status` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_accelerate_configuration` resource. ([#23816](https://github.com/hashicorp/terraform-provider-aws/issues/23816))
* resource/aws_s3_bucket: Update `acl` and `grant` parameters to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring these parameters with the standalone `aws_s3_bucket_acl` resource. ([#23798](https://github.com/hashicorp/terraform-provider-aws/issues/23798))
* resource/aws_s3_bucket: Update `cors_rule` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_cors_configuration` resource. ([#23817](https://github.com/hashicorp/terraform-provider-aws/issues/23817))
* resource/aws_s3_bucket: Update `lifecycle_rule` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_lifecycle_configuration` resource. ([#23818](https://github.com/hashicorp/terraform-provider-aws/issues/23818))
* resource/aws_s3_bucket: Update `logging` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_logging` resource. ([#23819](https://github.com/hashicorp/terraform-provider-aws/issues/23819))
* resource/aws_s3_bucket: Update `object_lock_configuration.rule` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_object_lock_configuration` resource. ([#23984](https://github.com/hashicorp/terraform-provider-aws/issues/23984))
* resource/aws_s3_bucket: Update `policy` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_policy` resource. ([#23843](https://github.com/hashicorp/terraform-provider-aws/issues/23843))
* resource/aws_s3_bucket: Update `replication_configuration` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_replication_configuration` resource. ([#23842](https://github.com/hashicorp/terraform-provider-aws/issues/23842))
* resource/aws_s3_bucket: Update `request_payer` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_request_payment_configuration` resource. ([#23844](https://github.com/hashicorp/terraform-provider-aws/issues/23844))
* resource/aws_s3_bucket: Update `server_side_encryption_configuration` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_server_side_encryption_configuration` resource. ([#23822](https://github.com/hashicorp/terraform-provider-aws/issues/23822))
* resource/aws_s3_bucket: Update `versioning` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_versioning` resource. ([#23820](https://github.com/hashicorp/terraform-provider-aws/issues/23820))
* resource/aws_s3_bucket: Update `website` parameter to be configurable. Please refer to the documentation for details on drift detection and potential conflicts when configuring this parameter with the standalone `aws_s3_bucket_website_configuration` resource. ([#23821](https://github.com/hashicorp/terraform-provider-aws/issues/23821))
* resource/aws_storagegateway_gateway: Add `maintenance_start_time` argument ([#15355](https://github.com/hashicorp/terraform-provider-aws/issues/15355))
* resource/aws_storagegateway_nfs_file_share: Add `bucket_region` and `vpc_endpoint_dns_name` arguments to support PrivateLink endpoints ([#24038](https://github.com/hashicorp/terraform-provider-aws/issues/24038))
* resource/aws_vpc_ipam: add `cascade` argument ([#23973](https://github.com/hashicorp/terraform-provider-aws/issues/23973))
* resource/aws_vpn_connection: Add `core_network_arn` and `core_network_attachment_arn` attributes ([#24024](https://github.com/hashicorp/terraform-provider-aws/issues/24024))
* resource/aws_xray_group: Add `insights_configuration` argument ([#24028](https://github.com/hashicorp/terraform-provider-aws/issues/24028))

BUG FIXES:

* data-source/aws_elasticache_cluster: Allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#23979](https://github.com/hashicorp/terraform-provider-aws/issues/23979))
* resource/aws_backup_report_plan: Wait for asynchronous lifecycle operations to complete ([#23967](https://github.com/hashicorp/terraform-provider-aws/issues/23967))
* resource/aws_cloudformation_stack_set: Consider `QUEUED` a valid pending state for resource creation ([#22160](https://github.com/hashicorp/terraform-provider-aws/issues/22160))
* resource/aws_dynamodb_table_item: Allow `item` names to still succeed if they include non-letters ([#14075](https://github.com/hashicorp/terraform-provider-aws/issues/14075))
* resource/aws_elasticache_cluster: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#23979](https://github.com/hashicorp/terraform-provider-aws/issues/23979))
* resource/aws_elasticache_parameter_group: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#23979](https://github.com/hashicorp/terraform-provider-aws/issues/23979))
* resource/aws_elasticache_replication_group: Allow disabling `auto_minor_version_upgrade` ([#23996](https://github.com/hashicorp/terraform-provider-aws/issues/23996))
* resource/aws_elasticache_replication_group: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#23979](https://github.com/hashicorp/terraform-provider-aws/issues/23979))
* resource/aws_elasticache_replication_group: Waits for available state before updating tags ([#24021](https://github.com/hashicorp/terraform-provider-aws/issues/24021))
* resource/aws_elasticache_subnet_group: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#23979](https://github.com/hashicorp/terraform-provider-aws/issues/23979))
* resource/aws_elasticache_user: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#23979](https://github.com/hashicorp/terraform-provider-aws/issues/23979))
* resource/aws_elasticache_user_group: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#23979](https://github.com/hashicorp/terraform-provider-aws/issues/23979))
* resource/aws_elasticsearch_domain_saml_option: Fix difference caused by `subject_key` default not matching AWS default; old and new defaults are equivalent ([#20892](https://github.com/hashicorp/terraform-provider-aws/issues/20892))
* resource/aws_lb: Fix attribute key not recognized issue preventing creation in ISO-B regions ([#23972](https://github.com/hashicorp/terraform-provider-aws/issues/23972))
* resource/aws_redshift_cluster: Correctly use `number_of_nodes` argument value when restoring from snapshot ([#13203](https://github.com/hashicorp/terraform-provider-aws/issues/13203))
* resource/aws_route: Ensure that resource ID is set in case of wait-for-creation time out ([#24024](https://github.com/hashicorp/terraform-provider-aws/issues/24024))
* resource/aws_s3_bucket_lifecycle_configuration: Prevent `MalformedXML` errors when handling diffs in `rule.filter` ([#23893](https://github.com/hashicorp/terraform-provider-aws/issues/23893))

## 4.8.0 (March 25, 2022)

FEATURES:

* **New Data Source:** `aws_mskconnect_connector` ([#23792](https://github.com/hashicorp/terraform-provider-aws/issues/23544))
* **New Resource:** `aws_mskconnect_connector` ([#23765](https://github.com/hashicorp/terraform-provider-aws/issues/23544))

ENHANCEMENTS:

* data-source/aws_eips: Set `public_ips` for VPC as well as EC2 Classic ([#23859](https://github.com/hashicorp/terraform-provider-aws/issues/23859))
* data-source/aws_elasticache_cluster: Add `log_delivery_configuration` attribute ([#20068](https://github.com/hashicorp/terraform-provider-aws/issues/20068))
* data-source/aws_elasticache_replication_group: Add `log_delivery_configuration` attribute ([#20068](https://github.com/hashicorp/terraform-provider-aws/issues/20068))
* data-source/aws_elasticsearch_domain: Add `cold_storage_options` attribute to the `cluster_config` configuration block ([#19713](https://github.com/hashicorp/terraform-provider-aws/issues/19713))
* data-source/aws_lambda_function: Add `ephemeral_storage` attribute ([#23873](https://github.com/hashicorp/terraform-provider-aws/issues/23873))
* resource/aws_elasticache_cluster: Add `log_delivery_configuration` argument ([#20068](https://github.com/hashicorp/terraform-provider-aws/issues/20068))
* resource/aws_elasticache_replication_group: Add `log_delivery_configuration` argument ([#20068](https://github.com/hashicorp/terraform-provider-aws/issues/20068))
* resource/aws_elasticsearch_domain: Add `cold_storage_options` argument to the `cluster_config` configuration block ([#19713](https://github.com/hashicorp/terraform-provider-aws/issues/19713))
* resource/aws_elasticsearch_domain: Add configurable Create and Delete timeouts ([#19713](https://github.com/hashicorp/terraform-provider-aws/issues/19713))
* resource/aws_lambda_function: Add `ephemeral_storage` argument ([#23873](https://github.com/hashicorp/terraform-provider-aws/issues/23873))
* resource/aws_lambda_function: Add error handling for `ResourceConflictException` errors on create and update ([#23879](https://github.com/hashicorp/terraform-provider-aws/issues/23879))
* resource/aws_mskconnect_custom_plugin: Implement resource Delete ([#23544](https://github.com/hashicorp/terraform-provider-aws/issues/23544))
* resource/aws_mwaa_environment: Add `schedulers` argument ([#21941](https://github.com/hashicorp/terraform-provider-aws/issues/21941))
* resource/aws_network_firewall_policy: Allow use of managed rule group arns for network firewall managed rule groups. ([#22355](https://github.com/hashicorp/terraform-provider-aws/issues/22355))

BUG FIXES:

* resource/aws_autoscaling_group: Fix issue where group was not recreated if `initial_lifecycle_hook` changed ([#20708](https://github.com/hashicorp/terraform-provider-aws/issues/20708))
* resource/aws_cloudfront_distribution: Fix default value of `origin_path` in `origin` block ([#20709](https://github.com/hashicorp/terraform-provider-aws/issues/20709))
* resource/aws_cloudwatch_event_target: Fix setting `path_parameter_values`. ([#23862](https://github.com/hashicorp/terraform-provider-aws/issues/23862))

## 4.7.0 (March 24, 2022)

FEATURES:

* **New Data Source:** `aws_cloudwatch_event_bus` ([#23792](https://github.com/hashicorp/terraform-provider-aws/issues/23792))
* **New Data Source:** `aws_imagebuilder_image_pipelines` ([#23741](https://github.com/hashicorp/terraform-provider-aws/issues/23741))
* **New Data Source:** `aws_memorydb_parameter_group` ([#23814](https://github.com/hashicorp/terraform-provider-aws/issues/23814))
* **New Data Source:** `aws_route53_traffic_policy_document` ([#23602](https://github.com/hashicorp/terraform-provider-aws/issues/23602))
* **New Resource:** `aws_cognito_user_in_group` ([#23765](https://github.com/hashicorp/terraform-provider-aws/issues/23765))
* **New Resource:** `aws_keyspaces_keyspace` ([#23770](https://github.com/hashicorp/terraform-provider-aws/issues/23770))
* **New Resource:** `aws_route53_traffic_policy` ([#23602](https://github.com/hashicorp/terraform-provider-aws/issues/23602))
* **New Resource:** `aws_route53_traffic_policy_instance` ([#23602](https://github.com/hashicorp/terraform-provider-aws/issues/23602))

ENHANCEMENTS:

* data-source/aws_imagebuilder_distribution_configuration: Add `organization_arns` and `organizational_unit_arns` attributes to the `distribution.ami_distribution_configuration.launch_permission` configuration block ([#22104](https://github.com/hashicorp/terraform-provider-aws/issues/22104))
* data-source/aws_msk_cluster: Add `zookeeper_connect_string_tls` attribute ([#23804](https://github.com/hashicorp/terraform-provider-aws/issues/23804))
* data-source/aws_ssm_document: Support `TEXT` `document_format` ([#23757](https://github.com/hashicorp/terraform-provider-aws/issues/23757))
* resource/aws_api_gateway_stage: Add `canary_settings` argument. ([#23754](https://github.com/hashicorp/terraform-provider-aws/issues/23754))
* resource/aws_athena_workgroup: Add `acl_configuration` and `expected_bucket_owner` arguments to the `configuration.result_configuration` block ([#23748](https://github.com/hashicorp/terraform-provider-aws/issues/23748))
* resource/aws_autoscaling_group: Add `instance_reuse_policy` argument to support [Warm Pool scale-in](https://aws.amazon.com/about-aws/whats-new/2022/02/amazon-ec2-auto-scaling-warm-pools-supports-hibernating-returning-instances-warm-pools-scale-in/) ([#23769](https://github.com/hashicorp/terraform-provider-aws/issues/23769))
* resource/aws_autoscaling_group: Update documentation to include [Warm Pool hibernation](https://aws.amazon.com/about-aws/whats-new/2022/02/amazon-ec2-auto-scaling-warm-pools-supports-hibernating-returning-instances-warm-pools-scale-in/) ([#23772](https://github.com/hashicorp/terraform-provider-aws/issues/23772))
* resource/aws_cloudformation_stack_set_instance: Add `operation_preferences` argument ([#23666](https://github.com/hashicorp/terraform-provider-aws/issues/23666))
* resource/aws_cloudwatch_log_subscription_filter: Add plan time validations for `name`, `destination_arn`, `filter_pattern`, `role_arn`, `distribution`. ([#23760](https://github.com/hashicorp/terraform-provider-aws/issues/23760))
* resource/aws_glue_schema: Update documentation to include [Protobuf data format support](https://aws.amazon.com/about-aws/whats-new/2022/02/aws-glue-schema-registry-protocol-buffers) ([#23815](https://github.com/hashicorp/terraform-provider-aws/issues/23815))
* resource/aws_imagebuilder_distribution_configuration: Add `organization_arns` and `organizational_unit_arns` arguments to the `distribution.ami_distribution_configuration.launch_permission` configuration block ([#22104](https://github.com/hashicorp/terraform-provider-aws/issues/22104))
* resource/aws_instance: Add `user_data_replace_on_change` attribute ([#23604](https://github.com/hashicorp/terraform-provider-aws/issues/23604))
* resource/aws_ssm_maintenance_window_task: Add `arn` and `window_task_id` attributes. ([#23756](https://github.com/hashicorp/terraform-provider-aws/issues/23756))
* resource/aws_ssm_maintenance_window_task: Add `cutoff_behavior` argument. ([#23756](https://github.com/hashicorp/terraform-provider-aws/issues/23756))

BUG FIXES:

* data-source/aws_ssm_document: Dont generate `arn` for AWS managed docs. ([#23757](https://github.com/hashicorp/terraform-provider-aws/issues/23757))
* resource/aws_ecs_service: Ensure that `load_balancer` and `service_registries` can be updated in-place ([#23786](https://github.com/hashicorp/terraform-provider-aws/issues/23786))
* resource/aws_launch_template: Fix `network_interfaces.device_index` and `network_interfaces.network_card_index` of `0` not being set ([#23767](https://github.com/hashicorp/terraform-provider-aws/issues/23767))
* resource/aws_ssm_maintenance_window_task: Allow creating a window taks without targets. ([#23756](https://github.com/hashicorp/terraform-provider-aws/issues/23756))

## 4.6.0 (March 18, 2022)

FEATURES:

* **New Data Source:** `aws_networkmanager_connection` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_connections` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_device` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_devices` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_global_network` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_global_networks` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_link` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_links` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_site` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Data Source:** `aws_networkmanager_sites` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_gamelift_game_server_group` ([#23606](https://github.com/hashicorp/terraform-provider-aws/issues/23606))
* **New Resource:** `aws_networkmanager_connection` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_networkmanager_customer_gateway_association` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_networkmanager_device` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_networkmanager_global_network` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_networkmanager_link` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_networkmanager_link_association` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_networkmanager_site` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_networkmanager_transit_gateway_connect_peer_association` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_networkmanager_transit_gateway_registration` ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* **New Resource:** `aws_vpc_endpoint_security_group_association` ([#13737](https://github.com/hashicorp/terraform-provider-aws/issues/13737))

ENHANCEMENTS:

* data-source/aws_ec2_transit_gateway_connect_peer: Add `arn` attribute ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* data-source/aws_imagebuilder_image: Add `container_recipe_arn` attribute ([#23647](https://github.com/hashicorp/terraform-provider-aws/issues/23647))
* data-source/aws_launch_template: Add `capacity_reservation_resource_group_arn` attribute to the `capacity_reservation_specification.capacity_reservation_target` configuration block ([#23365](https://github.com/hashicorp/terraform-provider-aws/issues/23365))
* data-source/aws_launch_template: Add `capacity_reservation_specification`, `cpu_options`, `elastic_inference_accelerator` and `license_specification` attributes ([#23365](https://github.com/hashicorp/terraform-provider-aws/issues/23365))
* data-source/aws_launch_template: Add `ipv4_prefixes`, `ipv4_prefix_count`, `ipv6_prefixes` and `ipv6_prefix_count` attributes to the `network_interfaces` configuration block ([#23365](https://github.com/hashicorp/terraform-provider-aws/issues/23365))
* data-source/aws_launch_template: Add `private_dns_name_options` attribute ([#23365](https://github.com/hashicorp/terraform-provider-aws/issues/23365))
* data-source/aws_redshift_cluster: Add `availability_zone_relocation_enabled` attribute. ([#20812](https://github.com/hashicorp/terraform-provider-aws/issues/20812))
* resource/aws_appconfig_configuration_profile: Add `type` argument to support [AWS AppConfig Feature Flags](https://aws.amazon.com/blogs/mt/using-aws-appconfig-feature-flags/) ([#23719](https://github.com/hashicorp/terraform-provider-aws/issues/23719))
* resource/aws_athena_database: Add `acl_configuration` and `expected_bucket_owner` arguments ([#23745](https://github.com/hashicorp/terraform-provider-aws/issues/23745))
* resource/aws_athena_database: Add `comment` argument to support database descriptions ([#23745](https://github.com/hashicorp/terraform-provider-aws/issues/23745))
* resource/aws_athena_database: Do not recreate the resource if `bucket` changes ([#23745](https://github.com/hashicorp/terraform-provider-aws/issues/23745))
* resource/aws_cloud9_environment_ec2: Add `connection_type` and `image_id` arguments ([#19195](https://github.com/hashicorp/terraform-provider-aws/issues/19195))
* resource/aws_cloudformation_stack_set:_instance: Add `call_as` argument ([#23339](https://github.com/hashicorp/terraform-provider-aws/issues/23339))
* resource/aws_dms_replication_task: Add optional `start_replication_task` and `status` argument ([#23692](https://github.com/hashicorp/terraform-provider-aws/issues/23692))
* resource/aws_ec2_transit_gateway_connect_peer: Add `arn` attribute ([#13251](https://github.com/hashicorp/terraform-provider-aws/issues/13251))
* resource/aws_ecs_service: `enable_ecs_managed_tags`, `load_balancer`, `propagate_tags` and `service_registries` can now be updated in-place ([#23600](https://github.com/hashicorp/terraform-provider-aws/issues/23600))
* resource/aws_imagebuilder_image: Add `container_recipe_arn` argument ([#23647](https://github.com/hashicorp/terraform-provider-aws/issues/23647))
* resource/aws_iot_certificate: Add `ca_pem` argument, enabling the use of existing IoT certificates ([#23126](https://github.com/hashicorp/terraform-provider-aws/issues/23126))
* resource/aws_iot_topic_rule: Add `cloudwatch_logs` and `error_action.cloudwatch_logs` arguments ([#23440](https://github.com/hashicorp/terraform-provider-aws/issues/23440))
* resource/aws_launch_configuration: Add `ephemeral_block_device.no_device` argument ([#23152](https://github.com/hashicorp/terraform-provider-aws/issues/23152))
* resource/aws_launch_template: Add `capacity_reservation_resource_group_arn` argument to the `capacity_reservation_specification.capacity_reservation_target` configuration block ([#23365](https://github.com/hashicorp/terraform-provider-aws/issues/23365))
* resource/aws_launch_template: Add `ipv4_prefixes`, `ipv4_prefix_count`, `ipv6_prefixes` and `ipv6_prefix_count` arguments to the `network_interfaces` configuration block ([#23365](https://github.com/hashicorp/terraform-provider-aws/issues/23365))
* resource/aws_launch_template: Add `private_dns_name_options` argument ([#23365](https://github.com/hashicorp/terraform-provider-aws/issues/23365))
* resource/aws_msk_configuration: Correctly set `latest_revision` as Computed when `server_properties` changes ([#23662](https://github.com/hashicorp/terraform-provider-aws/issues/23662))
* resource/aws_quicksight_user: Allow custom values for `namespace` ([#23607](https://github.com/hashicorp/terraform-provider-aws/issues/23607))
* resource/aws_rds_cluster: Add `db_cluster_instance_class`, `allocated_storage`, `storage_type`, and `iops` arguments to support [Multi-AZ deployments for MySQL & PostgreSQL](https://aws.amazon.com/blogs/aws/amazon-rds-multi-az-db-cluster/) ([#23684](https://github.com/hashicorp/terraform-provider-aws/issues/23684))
* resource/aws_rds_global_cluster: Add configurable timeouts ([#23560](https://github.com/hashicorp/terraform-provider-aws/issues/23560))
* resource/aws_rds_instance: Add `source_db_instance_automated_backup_arn` option within `restore_to_point_in_time` attribute ([#23086](https://github.com/hashicorp/terraform-provider-aws/issues/23086))
* resource/aws_redshift_cluster: Add `availability_zone_relocation_enabled` attribute and allow `availability_zone` to be changed in-place. ([#20812](https://github.com/hashicorp/terraform-provider-aws/issues/20812))
* resource/aws_transfer_server: Add `pre_authentication_login_banner` and `post_authentication_login_banner` arguments ([#23631](https://github.com/hashicorp/terraform-provider-aws/issues/23631))
* resource/aws_vpc_endpoint: The `security_group_ids` attribute can now be empty when the resource is created. In this case the VPC's default security is associated with the VPC endpoint ([#13737](https://github.com/hashicorp/terraform-provider-aws/issues/13737))

BUG FIXES:

* resource/aws_amplify_app: Allow `repository` to be updated in-place ([#23517](https://github.com/hashicorp/terraform-provider-aws/issues/23517))
* resource/aws_api_gateway_stage: Fixed issue with providing `cache_cluster_size` without `cache_cluster_enabled` resulted in waiter error ([#23091](https://github.com/hashicorp/terraform-provider-aws/issues/23091))
* resource/aws_athena_database: Remove from state on resource Read if deleted outside of Terraform ([#23745](https://github.com/hashicorp/terraform-provider-aws/issues/23745))
* resource/aws_cloudformation_stack_set: Use `call_as` attribute when reading stack sets, fixing an error raised when using a delegated admistrator account ([#23339](https://github.com/hashicorp/terraform-provider-aws/issues/23339))
* resource/aws_cloudsearch_domain: Set correct defaults for `index_field.facet`, `index_field.highlight`, `index_field.return`, `index_field.search` and `index_field.sort`, preventing spurious resource diffs ([#23687](https://github.com/hashicorp/terraform-provider-aws/issues/23687))
* resource/aws_db_instance: Fix issues where configured update timeout was not respected, and update would fail if instance were in the process of being configured. ([#23560](https://github.com/hashicorp/terraform-provider-aws/issues/23560))
* resource/aws_rds_event_subscription: Fix issue where `enabled` was sometimes not updated ([#23560](https://github.com/hashicorp/terraform-provider-aws/issues/23560))
* resource/aws_rds_global_cluster: Fix ability to perform cluster version upgrades, including of clusters in distinct regions, such as previously got error: "Invalid database cluster identifier" ([#23560](https://github.com/hashicorp/terraform-provider-aws/issues/23560))
* resource/aws_route53domains_registered_domain: Redirect all Route 53 Domains AWS API calls to the `us-east-1` Region ([#23672](https://github.com/hashicorp/terraform-provider-aws/issues/23672))
* resource/aws_s3_bucket_acl: Fix resource import for S3 bucket names consisting of uppercase letters, underscores, and a maximum of 255 characters ([#23678](https://github.com/hashicorp/terraform-provider-aws/issues/23678))
* resource/aws_s3_bucket_lifecycle_configuration: Support empty string filtering (default behavior of the `aws_s3_bucket.lifecycle_rule` parameter in provider versions prior to v4.0) ([#23746](https://github.com/hashicorp/terraform-provider-aws/issues/23746))
* resource/aws_s3_bucket_replication_configuration: Change `rule` configuration block to list instead of set ([#23703](https://github.com/hashicorp/terraform-provider-aws/issues/23703))
* resource/aws_s3_bucket_replication_configuration: Set `rule.id` as Computed to prevent drift when the value is not configured ([#23703](https://github.com/hashicorp/terraform-provider-aws/issues/23703))
* resource/aws_s3_bucket_versioning: Add missing support for `Disabled` bucket versioning ([#23723](https://github.com/hashicorp/terraform-provider-aws/issues/23723))

## 4.5.0 (March 11, 2022)

ENHANCEMENTS:

* resource/aws_account_alternate_contact: Add configurable timeouts ([#23516](https://github.com/hashicorp/terraform-provider-aws/issues/23516))
* resource/aws_s3_bucket: Add error handling for `NotImplemented` errors when reading `object_lock_enabled` and `object_lock_configuration` into terraform state. ([#13366](https://github.com/hashicorp/terraform-provider-aws/issues/13366))
* resource/aws_s3_bucket: Add top-level `object_lock_enabled` parameter ([#23556](https://github.com/hashicorp/terraform-provider-aws/issues/23556))
* resource/aws_s3_bucket_replication_configuration: Add `token` field to specify
x-amz-bucket-object-lock-token for enabling replication on object lock enabled
buckets or enabling object lock on an existing bucket. ([#23624](https://github.com/hashicorp/terraform-provider-aws/issues/23624))
* resource/aws_servicecatalog_budget_resource_association: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_constraint: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_organizations_access: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_portfolio: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_portfolio_share: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_principal_portfolio_association: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_product: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_product_portfolio_association: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_provisioned_product: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_provisioning_artifact: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_service_action: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_tag_option: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_servicecatalog_tag_option_resource_association: Add configurable timeouts ([#23518](https://github.com/hashicorp/terraform-provider-aws/issues/23518))
* resource/aws_synthetics_canary: Add optional `environment_variables` to `run_config`. ([#23574](https://github.com/hashicorp/terraform-provider-aws/issues/23574))

BUG FIXES:

* resource/aws_account_alternate_contact: Improve eventual consistency handling to avoid "no resource found" on updates ([#23516](https://github.com/hashicorp/terraform-provider-aws/issues/23516))
* resource/aws_image_builder_image_recipe: Fix regression in 4.3.0 whereby Windows-based images wouldn't build because of the newly introduced `systems_manager_agent.uninstall_after_build` argument. ([#23580](https://github.com/hashicorp/terraform-provider-aws/issues/23580))
* resource/aws_kms_external_key: Increase `tags` eventual consistency timeout from 5 minutes to 10 minutes ([#23593](https://github.com/hashicorp/terraform-provider-aws/issues/23593))
* resource/aws_kms_key: Increase `description` and `tags` eventual consistency timeouts from 5 minutes to 10 minutes ([#23593](https://github.com/hashicorp/terraform-provider-aws/issues/23593))
* resource/aws_kms_replica_external_key: Increase `tags` eventual consistency timeout from 5 minutes to 10 minutes ([#23593](https://github.com/hashicorp/terraform-provider-aws/issues/23593))
* resource/aws_kms_replica_key: Increase `tags` eventual consistency timeout from 5 minutes to 10 minutes ([#23593](https://github.com/hashicorp/terraform-provider-aws/issues/23593))
* resource/aws_s3_bucket_lifecycle_configuration: Correctly configure `rule.filter.object_size_greater_than` and `rule.filter.object_size_less_than` in API requests and terraform state ([#23441](https://github.com/hashicorp/terraform-provider-aws/issues/23441))
* resource/aws_s3_bucket_lifecycle_configuration: Prevent drift when `rule.noncurrent_version_expiration.newer_noncurrent_versions` or `rule.noncurrent_version_transition.newer_noncurrent_versions` is not specified ([#23441](https://github.com/hashicorp/terraform-provider-aws/issues/23441))
* resource/aws_s3_bucket_replication_configuration: Correctly configure empty `rule.filter` configuration block in API requests ([#23586](https://github.com/hashicorp/terraform-provider-aws/issues/23586))
* resource/aws_s3_bucket_replication_configuration: Ensure both `key` and `value` arguments of the `rule.filter.tag` configuration block are correctly populated in the outgoing API request and terraform state. ([#23579](https://github.com/hashicorp/terraform-provider-aws/issues/23579))
* resource/aws_s3_bucket_replication_configuration: Prevent inconsistent final plan when `rule.filter.prefix` is an empty string ([#23586](https://github.com/hashicorp/terraform-provider-aws/issues/23586))

## 4.4.0 (March 04, 2022)

FEATURES:

* **New Data Source:** `aws_connect_queue` ([#22768](https://github.com/hashicorp/terraform-provider-aws/issues/22768))
* **New Data Source:** `aws_ec2_serial_console_access` ([#23443](https://github.com/hashicorp/terraform-provider-aws/issues/23443))
* **New Data Source:** `aws_ec2_transit_gateway_connect` ([#22181](https://github.com/hashicorp/terraform-provider-aws/issues/22181))
* **New Data Source:** `aws_ec2_transit_gateway_connect_peer` ([#22181](https://github.com/hashicorp/terraform-provider-aws/issues/22181))
* **New Resource:** `aws_apprunner_vpc_connector` ([#23173](https://github.com/hashicorp/terraform-provider-aws/issues/23173))
* **New Resource:** `aws_connect_routing_profile` ([#22813](https://github.com/hashicorp/terraform-provider-aws/issues/22813))
* **New Resource:** `aws_connect_user_hierarchy_structure` ([#22836](https://github.com/hashicorp/terraform-provider-aws/issues/22836))
* **New Resource:** `aws_ec2_network_insights_path` ([#23330](https://github.com/hashicorp/terraform-provider-aws/issues/23330))
* **New Resource:** `aws_ec2_serial_console_access` ([#23443](https://github.com/hashicorp/terraform-provider-aws/issues/23443))
* **New Resource:** `aws_ec2_transit_gateway_connect` ([#22181](https://github.com/hashicorp/terraform-provider-aws/issues/22181))
* **New Resource:** `aws_ec2_transit_gateway_connect_peer` ([#22181](https://github.com/hashicorp/terraform-provider-aws/issues/22181))
* **New Resource:** `aws_grafana_license_association` ([#23401](https://github.com/hashicorp/terraform-provider-aws/issues/23401))
* **New Resource:** `aws_route53domains_registered_domain` ([#12711](https://github.com/hashicorp/terraform-provider-aws/issues/12711))

ENHANCEMENTS:

* data-source/aws_ec2_transit_gateway: Add `transit_gateway_cidr_blocks` attribute ([#22181](https://github.com/hashicorp/terraform-provider-aws/issues/22181))
* data-source/aws_eks_node_group: Add `taints` attribute ([#23452](https://github.com/hashicorp/terraform-provider-aws/issues/23452))
* resource/aws_apprunner_service: Add `network_configuration` argument ([#23173](https://github.com/hashicorp/terraform-provider-aws/issues/23173))
* resource/aws_cloudwatch_metric_alarm: Additional allowed values for `extended_statistic` and `metric_query.metric.stat` arguments ([#22942](https://github.com/hashicorp/terraform-provider-aws/issues/22942))
* resource/aws_ec2_transit_gateway: Add [custom `timeouts`](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts) block ([#22181](https://github.com/hashicorp/terraform-provider-aws/issues/22181))
* resource/aws_ec2_transit_gateway: Add `transit_gateway_cidr_blocks` argument ([#22181](https://github.com/hashicorp/terraform-provider-aws/issues/22181))
* resource/aws_eks_cluster: Retry when `ResourceInUseException` errors are returned from the AWS API during resource deletion ([#23366](https://github.com/hashicorp/terraform-provider-aws/issues/23366))
* resource/aws_glue_job: Add support for [streaming jobs](https://docs.aws.amazon.com/glue/latest/dg/add-job-streaming.html) by removing the default value for the `timeout` argument and marking it as Computed ([#23275](https://github.com/hashicorp/terraform-provider-aws/issues/23275))
* resource/aws_lambda_function: Add support for `dotnet6` `runtime` value ([#23426](https://github.com/hashicorp/terraform-provider-aws/issues/23426))
* resource/aws_lambda_layer_version: Add support for `dotnet6` `compatible_runtimes` value ([#23426](https://github.com/hashicorp/terraform-provider-aws/issues/23426))
* resource/aws_route: `nat_gateway_id` target no longer conflicts with `destination_ipv6_cidr_block` ([#23427](https://github.com/hashicorp/terraform-provider-aws/issues/23427))

BUG FIXES:

* resource/aws_dms_endpoint: Fix bug where KMS key was ignored for DynamoDB, OpenSearch, Kafka, Kinesis, Oracle, PostgreSQL, and S3 engines. ([#23444](https://github.com/hashicorp/terraform-provider-aws/issues/23444))
* resource/aws_networkfirewall_rule_group: Allow any character in `source` and `destination` `rule_group.rules_source.stateful_rule.header` arguments as per the AWS API docs ([#22727](https://github.com/hashicorp/terraform-provider-aws/issues/22727))
* resource/aws_opsworks_application: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_custom_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_ecs_cluster_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_ganglia_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_haproxy_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_instance: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_java_app_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_memcached_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_mysql_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_nodejs_app_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_php_app_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_rails_app_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_rds_db_instance: Correctly remove from state in certain deletion situations ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_stack: Fix error reported on successful deletion, lack of eventual consistency wait ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_static_web_layer: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_opsworks_user_profile: Fix error reported on successful deletion ([#23397](https://github.com/hashicorp/terraform-provider-aws/issues/23397))
* resource/aws_route53_resolver_firewall_domain_list: Remove limit for number of `domains`. ([#23485](https://github.com/hashicorp/terraform-provider-aws/issues/23485))
* resource/aws_synthetics_canary: Retry canary creation if it fails because of IAM propagation. ([#23394](https://github.com/hashicorp/terraform-provider-aws/issues/23394))

## 4.3.0 (February 28, 2022)

NOTES:

* resource/aws_internet_gateway: Set `vpc_id` as Computed to prevent drift when the `aws_internet_gateway_attachment` resource is used ([#16386](https://github.com/hashicorp/terraform-provider-aws/issues/16386))
* resource/aws_s3_bucket_lifecycle_configuration: The `prefix` argument of the `rule` configuration block has been deprecated. Use the `filter` configuration block instead. ([#23325](https://github.com/hashicorp/terraform-provider-aws/issues/23325))

FEATURES:

* **New Data Source:** `aws_ec2_transit_gateway_multicast_domain` ([#22756](https://github.com/hashicorp/terraform-provider-aws/issues/22756))
* **New Data Source:** `aws_ec2_transit_gateway_vpc_attachments` ([#12409](https://github.com/hashicorp/terraform-provider-aws/issues/12409))
* **New Resource:** `aws_ec2_transit_gateway_multicast_domain` ([#22756](https://github.com/hashicorp/terraform-provider-aws/issues/22756))
* **New Resource:** `aws_ec2_transit_gateway_multicast_domain_association` ([#22756](https://github.com/hashicorp/terraform-provider-aws/issues/22756))
* **New Resource:** `aws_ec2_transit_gateway_multicast_group_member` ([#22756](https://github.com/hashicorp/terraform-provider-aws/issues/22756))
* **New Resource:** `aws_ec2_transit_gateway_multicast_group_source` ([#22756](https://github.com/hashicorp/terraform-provider-aws/issues/22756))
* **New Resource:** `aws_internet_gateway_attachment` ([#16386](https://github.com/hashicorp/terraform-provider-aws/issues/16386))
* **New Resource:** `aws_opsworks_ecs_cluster_layer` ([#12495](https://github.com/hashicorp/terraform-provider-aws/issues/12495))
* **New Resource:** `aws_vpc_endpoint_policy` ([#17039](https://github.com/hashicorp/terraform-provider-aws/issues/17039))

ENHANCEMENTS:

* data-source/aws_ec2_transit_gateway: Add `multicast_support` attribute ([#22756](https://github.com/hashicorp/terraform-provider-aws/issues/22756))
* provider: Improves error message when `Profile` and static credential environment variables are set. ([#23388](https://github.com/hashicorp/terraform-provider-aws/issues/23388))
* provider: Makes `region` an optional parameter to allow sourcing from shared config files and IMDS ([#23384](https://github.com/hashicorp/terraform-provider-aws/issues/23384))
* provider: Retrieves region from IMDS when credentials retrieved from IMDS. ([#23388](https://github.com/hashicorp/terraform-provider-aws/issues/23388))
* resource/aws_connect_queue: The `quick_connect_ids` argument can now be updated in-place ([#22821](https://github.com/hashicorp/terraform-provider-aws/issues/22821))
* resource/aws_connect_security_profile: add `permissions` attribute to read ([#22761](https://github.com/hashicorp/terraform-provider-aws/issues/22761))
* resource/aws_ec2_fleet: Add `context` argument ([#23304](https://github.com/hashicorp/terraform-provider-aws/issues/23304))
* resource/aws_ec2_transit_gateway: Add `multicast_support` argument ([#22756](https://github.com/hashicorp/terraform-provider-aws/issues/22756))
* resource/aws_imagebuilder_image_pipeline: Add `schedule.timezone` argument ([#23322](https://github.com/hashicorp/terraform-provider-aws/issues/23322))
* resource/aws_imagebuilder_image_recipe: Add `systems_manager_agent.uninstall_after_build` argument ([#23293](https://github.com/hashicorp/terraform-provider-aws/issues/23293))
* resource/aws_instance: Prevent double base64 encoding of `user_data` and `user_data_base64` on update ([#23362](https://github.com/hashicorp/terraform-provider-aws/issues/23362))
* resource/aws_s3_bucket: Add error handling for `NotImplemented` error when reading `logging` into terraform state ([#23398](https://github.com/hashicorp/terraform-provider-aws/issues/23398))
* resource/aws_s3_bucket_object_lock_configuration: Mark `token` argument as sensitive ([#23368](https://github.com/hashicorp/terraform-provider-aws/issues/23368))
* resource/aws_servicecatalog_provisioned_product: Add `outputs` attribute ([#23270](https://github.com/hashicorp/terraform-provider-aws/issues/23270))

BUG FIXES:

* provider: Validates names of named profiles before use. ([#23388](https://github.com/hashicorp/terraform-provider-aws/issues/23388))
* resource/aws_dms_replication_task: Allow `cdc_start_position` to be computed ([#23328](https://github.com/hashicorp/terraform-provider-aws/issues/23328))
* resource/aws_ecs_cluster: Fix bug preventing describing clusters in ISO regions ([#23341](https://github.com/hashicorp/terraform-provider-aws/issues/23341))

## 4.2.0 (February 18, 2022)

FEATURES:

* **New Data Source:** `aws_grafana_workspace` ([#22874](https://github.com/hashicorp/terraform-provider-aws/issues/22874))
* **New Data Source:** `aws_iam_openid_connect_provider` ([#23240](https://github.com/hashicorp/terraform-provider-aws/issues/23240))
* **New Data Source:** `aws_ssm_instances` ([#23162](https://github.com/hashicorp/terraform-provider-aws/issues/23162))
* **New Resource:** `aws_cloudtrail_event_data_store` ([#22490](https://github.com/hashicorp/terraform-provider-aws/issues/22490))
* **New Resource:** `aws_grafana_workspace` ([#22874](https://github.com/hashicorp/terraform-provider-aws/issues/22874))

ENHANCEMENTS:

* provider: Add `custom_ca_bundle` argument ([#23279](https://github.com/hashicorp/terraform-provider-aws/issues/23279))
* provider: Add `sts_region` argument ([#23212](https://github.com/hashicorp/terraform-provider-aws/issues/23212))
* provider: Expands environment variables in file paths in provider configuration. ([#23282](https://github.com/hashicorp/terraform-provider-aws/issues/23282))
* provider: Updates list of valid AWS regions ([#23282](https://github.com/hashicorp/terraform-provider-aws/issues/23282))
* resource/aws_dms_endpoint: Add `s3_settings.add_column_name`, `s3_settings.canned_acl_for_objects`, `s3_settings.cdc_inserts_and_updates`, `s3_settings.cdc_inserts_only`, `s3_settings.cdc_max_batch_interval`, `s3_settings.cdc_min_file_size`, `s3_settings.cdc_path`, `s3_settings.csv_no_sup_value`, `s3_settings.csv_null_value`, `s3_settings.data_page_size`, `s3_settings.date_partition_delimiter`, `s3_settings.date_partition_sequence`, `s3_settings.dict_page_size_limit`, `s3_settings.enable_statistics`, `s3_settings.encoding_type`, `s3_settings.ignore_headers_row`, `s3_settings.include_op_for_full_load`, `s3_settings.max_file_size`, `s3_settings.preserve_transactions`, `s3_settings.rfc_4180`, `s3_settings.row_group_length`, `s3_settings.timestamp_column_name`, `s3_settings.use_csv_no_sup_value` arguments ([#20913](https://github.com/hashicorp/terraform-provider-aws/issues/20913))
* resource/aws_elasticache_replication_group: Add plan-time validation to `description` and `replication_group_description` to ensure non-empty strings ([#23254](https://github.com/hashicorp/terraform-provider-aws/issues/23254))
* resource/aws_fms_policy: Add `delete_unused_fm_managed_resources` argument ([#21295](https://github.com/hashicorp/terraform-provider-aws/issues/21295))
* resource/aws_fms_policy: Add `tags` argument and `tags_all` attribute to support resource tagging ([#21299](https://github.com/hashicorp/terraform-provider-aws/issues/21299))
* resource/aws_imagebuilder_image_recipe: Update plan time validation of `block_device_mapping.ebs.kms_key_id`, `block_device_mapping.ebs.snapshot_id`, `block_device_mapping.ebs.volume_type`, `name`, `parent_image`. ([#23235](https://github.com/hashicorp/terraform-provider-aws/issues/23235))
* resource/aws_instance: Allow updates to `user_data` and `user_data_base64` without forcing resource replacement ([#18043](https://github.com/hashicorp/terraform-provider-aws/issues/18043))
* resource/aws_s3_bucket: Add error handling for `MethodNotAllowed` and `XNotImplemented` errors when reading `website` into terraform state. ([#23278](https://github.com/hashicorp/terraform-provider-aws/issues/23278))
* resource/aws_s3_bucket: Add error handling for `NotImplemented` errors when reading `acceleration_status`, `policy`, or `request_payer` into terraform state. ([#23278](https://github.com/hashicorp/terraform-provider-aws/issues/23278))

BUG FIXES:

* provider: Credentials with expiry, such as assuming a role, would not renew. ([#23282](https://github.com/hashicorp/terraform-provider-aws/issues/23282))
* provider: Setting a custom CA bundle caused the provider to fail. ([#23282](https://github.com/hashicorp/terraform-provider-aws/issues/23282))
* resource/aws_iam_instance_profile: Improve tag handling in ISO regions ([#23283](https://github.com/hashicorp/terraform-provider-aws/issues/23283))
* resource/aws_iam_openid_connect_provider: Improve tag handling in ISO regions ([#23283](https://github.com/hashicorp/terraform-provider-aws/issues/23283))
* resource/aws_iam_policy: Improve tag handling in ISO regions ([#23283](https://github.com/hashicorp/terraform-provider-aws/issues/23283))
* resource/aws_iam_saml_provider: Improve tag handling in ISO regions ([#23283](https://github.com/hashicorp/terraform-provider-aws/issues/23283))
* resource/aws_iam_server_certificate: Improve tag handling in ISO regions ([#23283](https://github.com/hashicorp/terraform-provider-aws/issues/23283))
* resource/aws_iam_service_linked_role: Improve tag handling in ISO regions ([#23283](https://github.com/hashicorp/terraform-provider-aws/issues/23283))
* resource/aws_iam_virtual_mfa_device: Improve tag handling in ISO regions ([#23283](https://github.com/hashicorp/terraform-provider-aws/issues/23283))
* resource/aws_s3_bucket_lifecycle_configuration: Ensure both `key` and `value` arguments of the `filter` `tag` configuration block are correctly populated in the outgoing API request and terraform state. ([#23252](https://github.com/hashicorp/terraform-provider-aws/issues/23252))
* resource/aws_s3_bucket_lifecycle_configuration: Prevent non-empty plans when `filter` is an empty configuration block ([#23232](https://github.com/hashicorp/terraform-provider-aws/issues/23232))

## 4.1.0 (February 15, 2022)

FEATURES:

* **New Data Source:** `aws_backup_framework` ([#23193](https://github.com/hashicorp/terraform-provider-aws/issues/23193))
* **New Data Source:** `aws_backup_report_plan` ([#23146](https://github.com/hashicorp/terraform-provider-aws/issues/23146))
* **New Data Source:** `aws_imagebuilder_container_recipe` ([#23040](https://github.com/hashicorp/terraform-provider-aws/issues/23040))
* **New Data Source:** `aws_imagebuilder_container_recipes` ([#23134](https://github.com/hashicorp/terraform-provider-aws/issues/23134))
* **New Data Source:** `aws_service` ([#16640](https://github.com/hashicorp/terraform-provider-aws/issues/16640))
* **New Resource:** `aws_backup_framework` ([#23175](https://github.com/hashicorp/terraform-provider-aws/issues/23175))
* **New Resource:** `aws_backup_report_plan` ([#23098](https://github.com/hashicorp/terraform-provider-aws/issues/23098))
* **New Resource:** `aws_gamelift_script` ([#11560](https://github.com/hashicorp/terraform-provider-aws/issues/11560))
* **New Resource:** `aws_iam_service_specific_credential` ([#16185](https://github.com/hashicorp/terraform-provider-aws/issues/16185))
* **New Resource:** `aws_iam_signing_certificate` ([#23161](https://github.com/hashicorp/terraform-provider-aws/issues/23161))
* **New Resource:** `aws_iam_virtual_mfa_device` ([#23113](https://github.com/hashicorp/terraform-provider-aws/issues/23113))
* **New Resource:** `aws_imagebuilder_container_recipe` ([#22965](https://github.com/hashicorp/terraform-provider-aws/issues/22965))

ENHANCEMENTS:

* data-source/aws_imagebuilder_image_pipeline: Add `container_recipe_arn` attribute ([#23111](https://github.com/hashicorp/terraform-provider-aws/issues/23111))
* data-source/aws_kms_public_key: Add `public_key_pem` attribute ([#23130](https://github.com/hashicorp/terraform-provider-aws/issues/23130))
* resource/aws_api_gateway_authorizer: Add `arn` attribute. ([#23151](https://github.com/hashicorp/terraform-provider-aws/issues/23151))
* resource/aws_autoscaling_group: Disable scale-in protection before draining instances ([#23187](https://github.com/hashicorp/terraform-provider-aws/issues/23187))
* resource/aws_cloudformation_stack_set: Add `call_as` argument ([#22440](https://github.com/hashicorp/terraform-provider-aws/issues/22440))
* resource/aws_elastic_transcoder_preset: Add plan time validations to `audio.audio_packing_mode`,  `audio.channels`,
`audio.codec`,`audio.sample_rate`, `audio_codec_options.bit_depth`, `audio_codec_options.bit_order`,
`audio_codec_options.profile`, `audio_codec_options.signed`, `audio_codec_options.signed`,
`container`, `thumbnails.aspect_ratio`, `thumbnails.format`, `thumbnails.padding_policy`, `thumbnails.sizing_policy`,
`type`, `video.aspect_ratio`, `video.codec`, `video.display_aspect_ratio`, `video.fixed_gop`, `video.frame_rate`,   `video.max_frame_rate`,  `video.padding_policy`, `video.sizing_policy`, `video_watermarks.horizontal_align`,
`video_watermarks.id`, `video_watermarks.sizing_policy`, `video_watermarks.target`, `video_watermarks.vertical_align` ([#13974](https://github.com/hashicorp/terraform-provider-aws/issues/13974))
* resource/aws_elastic_transcoder_preset: Allow `audio.bit_rate` to be computed. ([#13974](https://github.com/hashicorp/terraform-provider-aws/issues/13974))
* resource/aws_gamelift_build: Add `object_version` argument to `storage_location` block. ([#22966](https://github.com/hashicorp/terraform-provider-aws/issues/22966))
* resource/aws_gamelift_build: Add import support ([#22966](https://github.com/hashicorp/terraform-provider-aws/issues/22966))
* resource/aws_gamelift_fleet: Add `certificate_configuration` argument ([#22967](https://github.com/hashicorp/terraform-provider-aws/issues/22967))
* resource/aws_gamelift_fleet: Add import support ([#22967](https://github.com/hashicorp/terraform-provider-aws/issues/22967))
* resource/aws_gamelift_fleet: Add plan time validation to `ec2_instance_type` ([#22967](https://github.com/hashicorp/terraform-provider-aws/issues/22967))
* resource/aws_gamelift_fleet: Adds `script_arn` attribute. ([#11560](https://github.com/hashicorp/terraform-provider-aws/issues/11560))
* resource/aws_gamelift_fleet: Adds `script_id` argument. ([#11560](https://github.com/hashicorp/terraform-provider-aws/issues/11560))
* resource/aws_glue_catalog_database: Add support `create_table_default_permission` argument ([#22964](https://github.com/hashicorp/terraform-provider-aws/issues/22964))
* resource/aws_glue_trigger: Add `event_batching_condition` argument. ([#22963](https://github.com/hashicorp/terraform-provider-aws/issues/22963))
* resource/aws_iam_user_login_profile: Make `pgp_key` optional ([#12384](https://github.com/hashicorp/terraform-provider-aws/issues/12384))
* resource/aws_imagebuilder_image_pipeline: Add `container_recipe_arn` argument ([#23111](https://github.com/hashicorp/terraform-provider-aws/issues/23111))
* resource/aws_prometheus_workspace: Add `tags` argument and `tags_all` attribute to support resource tagging ([#23202](https://github.com/hashicorp/terraform-provider-aws/issues/23202))
* resource/aws_ssm_association: Add `arn` attribute ([#17732](https://github.com/hashicorp/terraform-provider-aws/issues/17732))
* resource/aws_ssm_association: Add `wait_for_success_timeout_seconds` argument ([#17732](https://github.com/hashicorp/terraform-provider-aws/issues/17732))
* resource/aws_ssm_association: Add plan time validation to `association_name`, `document_version`, `schedule_expression`, `output_location.s3_bucket_name`, `output_location.s3_key_prefix`, `targets.key`, `targets.values`, `automation_target_parameter_name` ([#17732](https://github.com/hashicorp/terraform-provider-aws/issues/17732))

BUG FIXES:

* data-source/aws_vpc_ipam_pool: error if no pool found ([#23195](https://github.com/hashicorp/terraform-provider-aws/issues/23195))
* provider: Support `ap-northeast-3`, `ap-southeast-3` and `us-iso-west-1` as valid AWS Regions ([#23191](https://github.com/hashicorp/terraform-provider-aws/issues/23191))
* provider: Use AWS HTTP client which allows IMDS authentication in container environments and custom RootCAs in ISO regions ([#23191](https://github.com/hashicorp/terraform-provider-aws/issues/23191))
* resource/aws_appmesh_route: Handle zero `max_retries` ([#23035](https://github.com/hashicorp/terraform-provider-aws/issues/23035))
* resource/aws_elastic_transcoder_preset: Allow `video_codec_options` to be empty. ([#13974](https://github.com/hashicorp/terraform-provider-aws/issues/13974))
* resource/aws_rds_cluster: Fix crash when configured `engine_version` string is shorter than the `EngineVersion` string returned from the AWS API ([#23039](https://github.com/hashicorp/terraform-provider-aws/issues/23039))
* resource/aws_s3_bucket_lifecycle_configuration: Correctly handle the `days` value of the `rule` `transition` configuration block when set to `0` ([#23120](https://github.com/hashicorp/terraform-provider-aws/issues/23120))
* resource/aws_s3_bucket_lifecycle_configuration: Fix extraneous diffs especially after import ([#23144](https://github.com/hashicorp/terraform-provider-aws/issues/23144))
* resource/aws_sagemaker_endpoint_configuration: Emptiness check for arguments, Allow not passing `async_inference_config.kms_key_id`. ([#22960](https://github.com/hashicorp/terraform-provider-aws/issues/22960))
* resource/aws_vpn_connection: Add support for `ipsec.1-aes256` connection type ([#23127](https://github.com/hashicorp/terraform-provider-aws/issues/23127))

## 4.0.0 (February 10, 2022)

BREAKING CHANGES:

* data-source/aws_connect_hours_of_operation: The hours_of_operation_arn attribute is renamed to arn ([#22375](https://github.com/hashicorp/terraform-provider-aws/issues/22375))
* resource/aws_batch_compute_environment: No `compute_resources` configuration block can be specified when `type` is `UNMANAGED` ([#22805](https://github.com/hashicorp/terraform-provider-aws/issues/22805))
* resource/aws_cloudwatch_event_target: The `ecs_target` `launch_type` argument no longer has a default value (previously was `EC2`) ([#22803](https://github.com/hashicorp/terraform-provider-aws/issues/22803))
* resource/aws_cloudwatch_event_target: `ecs_target.0.launch_type` can no longer be set to `""`; instead, remove or set to `null` ([#22954](https://github.com/hashicorp/terraform-provider-aws/issues/22954))
* resource/aws_connect_hours_of_operation: The hours_of_operation_arn attribute is renamed to arn ([#22375](https://github.com/hashicorp/terraform-provider-aws/issues/22375))
* resource/aws_default_network_acl: These arguments can no longer be set to `""`: `egress.*.cidr_block`, `egress.*.ipv6_cidr_block`, `ingress.*.cidr_block`, or `ingress.*.ipv6_cidr_block` ([#22928](https://github.com/hashicorp/terraform-provider-aws/issues/22928))
* resource/aws_default_route_table: These arguments can no longer be set to `""`: `route.*.cidr_block`, `route.*.ipv6_cidr_block` ([#22931](https://github.com/hashicorp/terraform-provider-aws/issues/22931))
* resource/aws_default_vpc: `ipv6_cidr_block` can no longer be set to `""`; remove or set to `null` ([#22948](https://github.com/hashicorp/terraform-provider-aws/issues/22948))
* resource/aws_efs_mount_target: `ip_address` can no longer be set to `""`; instead, remove or set to `null` ([#22954](https://github.com/hashicorp/terraform-provider-aws/issues/22954))
* resource/aws_elasticache_cluster: Either `engine` or `replication_group_id` must be specified ([#20482](https://github.com/hashicorp/terraform-provider-aws/issues/20482))
* resource/aws_elasticsearch_domain: `ebs_options.0.volume_type` can no longer be set to `""`; instead, remove or set to `null` ([#22954](https://github.com/hashicorp/terraform-provider-aws/issues/22954))
* resource/aws_fsx_ontap_storage_virtual_machine: Remove deprecated `active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name`, migrating value to `active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name` ([#22915](https://github.com/hashicorp/terraform-provider-aws/issues/22915))
* resource/aws_instance: `private_ip` can no longer be set to `""`; remove or set to `null` ([#22948](https://github.com/hashicorp/terraform-provider-aws/issues/22948))
* resource/aws_lb_target_group: For `protocol = "TCP"`, `stickiness` can no longer be type set to `lb_cookie` even when `enabled = false`; instead use type `source_ip` ([#22996](https://github.com/hashicorp/terraform-provider-aws/issues/22996))
* resource/aws_network_acl: These arguments can no longer be set to `""`: `egress.*.cidr_block`, `egress.*.ipv6_cidr_block`, `ingress.*.cidr_block`, or `ingress.*.ipv6_cidr_block` ([#22928](https://github.com/hashicorp/terraform-provider-aws/issues/22928))
* resource/aws_route: Exactly one of these can be set: `destination_cidr_block`, `destination_ipv6_cidr_block`, `destination_prefix_list_id`. These arguments can no longer be set to `""`: `destination_cidr_block`, `destination_ipv6_cidr_block`. ([#22931](https://github.com/hashicorp/terraform-provider-aws/issues/22931))
* resource/aws_route_table: These arguments can no longer be set to `""`: `route.*.cidr_block`, `route.*.ipv6_cidr_block` ([#22931](https://github.com/hashicorp/terraform-provider-aws/issues/22931))
* resource/aws_s3_bucket: The `acceleration_status` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_accelerate_configuration` resource instead. ([#22610](https://github.com/hashicorp/terraform-provider-aws/issues/22610))
* resource/aws_s3_bucket: The `acl` and `grant` arguments have been deprecated and are now read-only. Use the `aws_s3_bucket_acl` resource instead. ([#22537](https://github.com/hashicorp/terraform-provider-aws/issues/22537))
* resource/aws_s3_bucket: The `cors_rule` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_cors_configuration` resource instead. ([#22611](https://github.com/hashicorp/terraform-provider-aws/issues/22611))
* resource/aws_s3_bucket: The `lifecycle_rule` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_lifecycle_configuration` resource instead. ([#22581](https://github.com/hashicorp/terraform-provider-aws/issues/22581))
* resource/aws_s3_bucket: The `logging` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_logging` resource instead. ([#22599](https://github.com/hashicorp/terraform-provider-aws/issues/22599))
* resource/aws_s3_bucket: The `object_lock_configuration` `rule` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_object_lock_configuration` resource instead. ([#22612](https://github.com/hashicorp/terraform-provider-aws/issues/22612))
* resource/aws_s3_bucket: The `policy` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_policy` resource instead. ([#22538](https://github.com/hashicorp/terraform-provider-aws/issues/22538))
* resource/aws_s3_bucket: The `replication_configuration` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_replication_configuration` resource instead. ([#22604](https://github.com/hashicorp/terraform-provider-aws/issues/22604))
* resource/aws_s3_bucket: The `request_payer` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_request_payment_configuration` resource instead. ([#22613](https://github.com/hashicorp/terraform-provider-aws/issues/22613))
* resource/aws_s3_bucket: The `server_side_encryption_configuration` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_server_side_encryption_configuration` resource instead. ([#22605](https://github.com/hashicorp/terraform-provider-aws/issues/22605))
* resource/aws_s3_bucket: The `versioning` argument has been deprecated and is now read-only. Use the `aws_s3_bucket_versioning` resource instead. ([#22606](https://github.com/hashicorp/terraform-provider-aws/issues/22606))
* resource/aws_s3_bucket: The `website`, `website_domain`, and `website_endpoint` arguments have been deprecated and are now read-only. Use the `aws_s3_bucket_website_configuration` resource instead. ([#22614](https://github.com/hashicorp/terraform-provider-aws/issues/22614))
* resource/aws_vpc: `ipv6_cidr_block` can no longer be set to `""`; remove or set to `null` ([#22948](https://github.com/hashicorp/terraform-provider-aws/issues/22948))
* resource/aws_vpc_ipv6_cidr_block_association: `ipv6_cidr_block` can no longer be set to `""`; remove or set to `null` ([#22948](https://github.com/hashicorp/terraform-provider-aws/issues/22948))

NOTES:

* data-source/aws_cognito_user_pools: The type of the `ids` and `arns` attributes has changed from Set to List. If no volumes match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_db_event_categories: The type of the `ids` attribute has changed from Set to List. If no event categories match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_ebs_volumes: The type of the `ids` attribute has changed from Set to List. If no volumes match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_ec2_coip_pools: The type of the `pool_ids` attribute has changed from Set to List. If no COIP pools match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_ec2_local_gateway_route_tables: The type of the `ids` attribute has changed from Set to List. If no local gateway route tables match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_ec2_local_gateway_virtual_interface_groups: The type of the `ids` and `local_gateway_virtual_interface_ids` attributes has changed from Set to List. If no local gateway virtual interface groups match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_ec2_local_gateways: The type of the `ids` attribute has changed from Set to List. If no local gateways match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_ec2_transit_gateway_route_tables: The type of the `ids` attribute has changed from Set to List. If no transit gateway route tables match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_efs_access_points: The type of the `ids` and `arns` attributes has changed from Set to List. If no access points match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_elasticache_replication_group: The `number_cache_clusters` attribute has been deprecated. All configurations using `number_cache_clusters` should be updated to use the `num_cache_clusters` attribute instead ([#22667](https://github.com/hashicorp/terraform-provider-aws/issues/22667))
* data-source/aws_elasticache_replication_group: The `replication_group_description` attribute has been deprecated. All configurations using `replication_group_description` should be updated to use the `description` attribute instead ([#22667](https://github.com/hashicorp/terraform-provider-aws/issues/22667))
* data-source/aws_emr_release_labels: The type of the `ids` attribute has changed from Set to List. If no release labels match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_iam_policy_document: The `source_json` and `override_json` attributes have been deprecated. Use the `source_policy_documents` and `override_policy_documents` attributes respectively instead. ([#22890](https://github.com/hashicorp/terraform-provider-aws/issues/22890))
* data-source/aws_inspector_rules_packages: If no rules packages match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_instances: If no instances match the specified criteria an empty list is returned (previously an error was raised) ([#5055](https://github.com/hashicorp/terraform-provider-aws/issues/5055))
* data-source/aws_ip_ranges: If no ranges match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_network_acls: The type of the `ids` attribute has changed from Set to List. If no NACLs match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_network_interfaces: The type of the `ids` attribute has changed from Set to List. If no network interfaces match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_route_tables: The type of the `ids` attribute has changed from Set to List. If no route tables match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_s3_bucket_object: The data source is deprecated; use `aws_s3_object` instead ([#22877](https://github.com/hashicorp/terraform-provider-aws/issues/22877))
* data-source/aws_s3_bucket_objects: The data source is deprecated; use `aws_s3_objects` instead ([#22877](https://github.com/hashicorp/terraform-provider-aws/issues/22877))
* data-source/aws_security_groups: If no security groups match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_ssoadmin_instances: The type of the `identity_store_ids` and `arns` attributes has changed from Set to List. If no instances match the specified criteria an empty list is returned (previously an error was raised) ([#21219](https://github.com/hashicorp/terraform-provider-aws/issues/21219))
* data-source/aws_subnet_ids: The `aws_subnet_ids` data source has been deprecated and will be removed in a future version. Use the `aws_subnets` data source instead ([#22743](https://github.com/hashicorp/terraform-provider-aws/issues/22743))
* data-source/aws_vpcs: The type of the `ids` attributes has changed from Set to List. If no VPCs match the specified criteria an empty list is returned (previously an error was raised) ([#22253](https://github.com/hashicorp/terraform-provider-aws/issues/22253))
* provider: The `assume_role.duration_seconds` argument has been deprecated. All configurations using `assume_role.duration_seconds` should be updated to use the new `assume_role.duration` argument instead. ([#23077](https://github.com/hashicorp/terraform-provider-aws/issues/23077))
* resource/aws_acmpca_certificate_authority: The `status` attribute has been deprecated. Use the `enabled` attribute instead. ([#22878](https://github.com/hashicorp/terraform-provider-aws/issues/22878))
* resource/aws_autoscaling_attachment: The `alb_target_group_arn` argument has been deprecated. All configurations using `alb_target_group_arn` should be updated to use the new `lb_target_group_arn` argument instead ([#22662](https://github.com/hashicorp/terraform-provider-aws/issues/22662))
* resource/aws_autoscaling_group: The `tags` argument has been deprecated. All configurations using `tags` should be updated to use the `tag` argument instead ([#22663](https://github.com/hashicorp/terraform-provider-aws/issues/22663))
* resource/aws_budgets_budget: The `cost_filters` attribute has been deprecated. Use the `cost_filter` attribute instead. ([#22888](https://github.com/hashicorp/terraform-provider-aws/issues/22888))
* resource/aws_connect_hours_of_operation: Timeout support has been removed as it is not needed for this resource ([#22375](https://github.com/hashicorp/terraform-provider-aws/issues/22375))
* resource/aws_customer_gateway: `ip_address` can no longer be set to `""` ([#22926](https://github.com/hashicorp/terraform-provider-aws/issues/22926))
* resource/aws_db_instance: The `name` argument has been deprecated. All configurations using `name` should be updated to use the `db_name` argument instead ([#22668](https://github.com/hashicorp/terraform-provider-aws/issues/22668))
* resource/aws_default_subnet: If no default subnet exists in the specified Availability Zone one is now created. The `force_destroy` destroy argument has been added (defaults to `false`). Setting this argument to `true` deletes the default subnet on `terraform destroy` ([#22253](https://github.com/hashicorp/terraform-provider-aws/issues/22253))
* resource/aws_default_vpc: If no default VPC exists in the current AWS Region one is now created. The `force_destroy` destroy argument has been added (defaults to `false`). Setting this argument to `true` deletes the default VPC on `terraform destroy` ([#22253](https://github.com/hashicorp/terraform-provider-aws/issues/22253))
* resource/aws_ec2_client_vpn_endpoint: The `status` attribute has been deprecated ([#22887](https://github.com/hashicorp/terraform-provider-aws/issues/22887))
* resource/aws_ec2_client_vpn_endpoint: The type of the `dns_servers` argument has changed from Set to List ([#22889](https://github.com/hashicorp/terraform-provider-aws/issues/22889))
* resource/aws_ec2_client_vpn_network_association: The `security_groups` argument has been deprecated. Use the `security_group_ids` argument of the `aws_ec2_client_vpn_endpoint` resource instead ([#22911](https://github.com/hashicorp/terraform-provider-aws/issues/22911))
* resource/aws_ec2_client_vpn_network_association: The `status` attribute has been deprecated ([#22887](https://github.com/hashicorp/terraform-provider-aws/issues/22887))
* resource/aws_ec2_client_vpn_route: Add [custom `timeouts`](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts) block ([#22911](https://github.com/hashicorp/terraform-provider-aws/issues/22911))
* resource/aws_ecs_cluster: The `capacity_providers` and `default_capacity_provider_strategy` arguments have been deprecated. Use the `aws_ecs_cluster_capacity_providers` resource instead. ([#22783](https://github.com/hashicorp/terraform-provider-aws/issues/22783))
* resource/aws_elasticache_replication_group: The `cluster_mode` argument has been deprecated. All configurations using `cluster_mode` should be updated to use the root-level `num_node_groups` and `replicas_per_node_group` arguments instead ([#22666](https://github.com/hashicorp/terraform-provider-aws/issues/22666))
* resource/aws_elasticache_replication_group: The `number_cache_clusters` argument has been deprecated. All configurations using `number_cache_clusters` should be updated to use the `num_cache_clusters` argument instead ([#22666](https://github.com/hashicorp/terraform-provider-aws/issues/22666))
* resource/aws_elasticache_replication_group: The `replication_group_description` argument has been deprecated. All configurations using `replication_group_description` should be updated to use the `description` argument instead ([#22666](https://github.com/hashicorp/terraform-provider-aws/issues/22666))
* resource/aws_route: The `instance_id` argument has been deprecated. All configurations using `instance_id` should be updated to use the `network_interface_id` argument instead ([#22664](https://github.com/hashicorp/terraform-provider-aws/issues/22664))
* resource/aws_route_table: The `instance_id` argument of the `route` configuration block has been deprecated. All configurations using `route` `instance_id` should be updated to use the `route` `network_interface_id` argument instead ([#22664](https://github.com/hashicorp/terraform-provider-aws/issues/22664))
* resource/aws_s3_bucket_object: The resource is deprecated; use `aws_s3_object` instead ([#22877](https://github.com/hashicorp/terraform-provider-aws/issues/22877))

FEATURES:

* **New Data Source:** `aws_cloudfront_realtime_log_config` ([#22620](https://github.com/hashicorp/terraform-provider-aws/issues/22620))
* **New Data Source:** `aws_ec2_client_vpn_endpoint` ([#14218](https://github.com/hashicorp/terraform-provider-aws/issues/14218))
* **New Data Source:** `aws_eips` ([#7537](https://github.com/hashicorp/terraform-provider-aws/issues/7537))
* **New Data Source:** `aws_s3_object` ([#22850](https://github.com/hashicorp/terraform-provider-aws/issues/22850))
* **New Data Source:** `aws_s3_objects` ([#22850](https://github.com/hashicorp/terraform-provider-aws/issues/22850))
* **New Resource:** `aws_cognito_user` ([#19919](https://github.com/hashicorp/terraform-provider-aws/issues/19919))
* **New Resource:** `aws_dataexchange_revision` ([#22933](https://github.com/hashicorp/terraform-provider-aws/issues/22933))
* **New Resource:** `aws_network_acl_association` ([#18807](https://github.com/hashicorp/terraform-provider-aws/issues/18807))
* **New Resource:** `aws_s3_bucket_accelerate_configuration` ([#22617](https://github.com/hashicorp/terraform-provider-aws/issues/22617))
* **New Resource:** `aws_s3_bucket_acl` ([#22853](https://github.com/hashicorp/terraform-provider-aws/issues/22853))
* **New Resource:** `aws_s3_bucket_cors_configuration` ([#12141](https://github.com/hashicorp/terraform-provider-aws/issues/12141))
* **New Resource:** `aws_s3_bucket_lifecycle_configuration` ([#22579](https://github.com/hashicorp/terraform-provider-aws/issues/22579))
* **New Resource:** `aws_s3_bucket_logging` ([#22608](https://github.com/hashicorp/terraform-provider-aws/issues/22608))
* **New Resource:** `aws_s3_bucket_object_lock_configuration` ([#22644](https://github.com/hashicorp/terraform-provider-aws/issues/22644))
* **New Resource:** `aws_s3_bucket_request_payment_configuration` ([#22649](https://github.com/hashicorp/terraform-provider-aws/issues/22649))
* **New Resource:** `aws_s3_bucket_server_side_encryption_configuration` ([#22609](https://github.com/hashicorp/terraform-provider-aws/issues/22609))
* **New Resource:** `aws_s3_bucket_versioning` ([#5132](https://github.com/hashicorp/terraform-provider-aws/issues/5132))
* **New Resource:** `aws_s3_bucket_website_configuration` ([#22648](https://github.com/hashicorp/terraform-provider-aws/issues/22648))
* **New Resource:** `aws_s3_object` ([#22850](https://github.com/hashicorp/terraform-provider-aws/issues/22850))

ENHANCEMENTS:

* data-source/aws_ami: Add `boot_mode` attribute. ([#22939](https://github.com/hashicorp/terraform-provider-aws/issues/22939))
* data-source/aws_cloudwatch_log_group: Automatically trim `:*` suffix from `arn` attribute ([#22043](https://github.com/hashicorp/terraform-provider-aws/issues/22043))
* data-source/aws_ec2_client_vpn_endpoint: Add `security_group_ids` and `vpc_id` attributes ([#22911](https://github.com/hashicorp/terraform-provider-aws/issues/22911))
* data-source/aws_elasticache_replication_group: Add `description`, `num_cache_clusters`, `num_node_groups`, and `replicas_per_node_group` attributes ([#22667](https://github.com/hashicorp/terraform-provider-aws/issues/22667))
* data-source/aws_imagebuilder_distribution_configuration: Add `container_distribution_configuration` attribute to the `distribution` configuration block ([#22838](https://github.com/hashicorp/terraform-provider-aws/issues/22838))
* data-source/aws_imagebuilder_distribution_configuration: Add `launch_template_configuration` attribute to the `distribution` configuration block ([#22884](https://github.com/hashicorp/terraform-provider-aws/issues/22884))
* data-source/aws_imagebuilder_image_recipe: Add `parameter` attribute to the `component` configuration block ([#22856](https://github.com/hashicorp/terraform-provider-aws/issues/22856))
* provider: Add `duration` argument to the `assume_role` configuration block ([#23077](https://github.com/hashicorp/terraform-provider-aws/issues/23077))
* provider: Add `ec2_metadata_service_endpoint`, `ec2_metadata_service_endpoint_mode`, `use_dualstack_endpoint`, `use_fips_endpoint` arguments ([#22804](https://github.com/hashicorp/terraform-provider-aws/issues/22804))
* provider: Add environment variables `TF_AWS_DYNAMODB_ENDPOINT`, `TF_AWS_IAM_ENDPOINT`, `TF_AWS_S3_ENDPOINT`, and `TF_AWS_STS_ENDPOINT`. ([#23052](https://github.com/hashicorp/terraform-provider-aws/issues/23052))
* provider: Add support for `shared_config_file` parameter ([#20587](https://github.com/hashicorp/terraform-provider-aws/issues/20587))
* provider: Add support for `shared_credentials_files` parameter and deprecates `shared_credentials_file` ([#23080](https://github.com/hashicorp/terraform-provider-aws/issues/23080))
* provider: Adds `s3_use_path_style` parameter and deprecates `s3_force_path_style`. ([#23055](https://github.com/hashicorp/terraform-provider-aws/issues/23055))
* provider: Changes `shared_config_file` parameter to `shared_config_files` ([#23080](https://github.com/hashicorp/terraform-provider-aws/issues/23080))
* provider: Updates AWS authentication to use AWS SDK for Go v2 <https://aws.github.io/aws-sdk-go-v2/docs/> ([#20587](https://github.com/hashicorp/terraform-provider-aws/issues/20587))
* resource/aws_ami: Add `boot_mode` and `ebs_block_device.outpost_arn` arguments. ([#22939](https://github.com/hashicorp/terraform-provider-aws/issues/22939))
* resource/aws_ami_copy: Add `boot_mode` and `ebs_block_device.outpost_arn` attributes ([#22972](https://github.com/hashicorp/terraform-provider-aws/issues/22972))
* resource/aws_ami_from_instance: Add `boot_mode` and `ebs_block_device.outpost_arn` attributes ([#22972](https://github.com/hashicorp/terraform-provider-aws/issues/22972))
* resource/aws_api_gateway_domain_name: Add `ownership_verification_certificate_arn` argument. ([#21076](https://github.com/hashicorp/terraform-provider-aws/issues/21076))
* resource/aws_apigatewayv2_domain_name: Add `domain_name_configuration.ownership_verification_certificate_arn` argument. ([#21076](https://github.com/hashicorp/terraform-provider-aws/issues/21076))
* resource/aws_autoscaling_attachment: Add `lb_target_group_arn` argument ([#22662](https://github.com/hashicorp/terraform-provider-aws/issues/22662))
* resource/aws_cloudwatch_event_target: Add plan time validation for `input`, `input_path`, `run_command_targets.values`, `http_target.header_parameters`, `http_target.query_string_parameters`, `redshift_target.database`, `redshift_target.db_user`, `redshift_target.secrets_manager_arn`, `redshift_target.sql`, `redshift_target.statement_name`, `retry_policy.maximum_event_age_in_seconds`, `retry_policy.maximum_retry_attempts`. ([#22946](https://github.com/hashicorp/terraform-provider-aws/issues/22946))
* resource/aws_db_instance: Add `db_name` argument ([#22668](https://github.com/hashicorp/terraform-provider-aws/issues/22668))
* resource/aws_ec2_client_vpn_authorization_rule: Configurable Create and Delete timeouts ([#20688](https://github.com/hashicorp/terraform-provider-aws/issues/20688))
* resource/aws_ec2_client_vpn_endpoint: Add `client_connect_options` argument ([#22793](https://github.com/hashicorp/terraform-provider-aws/issues/22793))
* resource/aws_ec2_client_vpn_endpoint: Add `client_login_banner_options` argument ([#22793](https://github.com/hashicorp/terraform-provider-aws/issues/22793))
* resource/aws_ec2_client_vpn_endpoint: Add `security_group_ids` and `vpc_id` arguments ([#22911](https://github.com/hashicorp/terraform-provider-aws/issues/22911))
* resource/aws_ec2_client_vpn_endpoint: Add `session_timeout_hours` argument ([#22793](https://github.com/hashicorp/terraform-provider-aws/issues/22793))
* resource/aws_ec2_client_vpn_endpoint: Add `vpn_port` argument ([#22793](https://github.com/hashicorp/terraform-provider-aws/issues/22793))
* resource/aws_ec2_client_vpn_network_association: Configurable Create and Delete timeouts ([#20689](https://github.com/hashicorp/terraform-provider-aws/issues/20689))
* resource/aws_elasticache_replication_group: Add `description` argument ([#22666](https://github.com/hashicorp/terraform-provider-aws/issues/22666))
* resource/aws_elasticache_replication_group: Add `num_cache_clusters` argument ([#22666](https://github.com/hashicorp/terraform-provider-aws/issues/22666))
* resource/aws_elasticache_replication_group: Add `num_node_groups` and `replicas_per_node_group` arguments ([#22666](https://github.com/hashicorp/terraform-provider-aws/issues/22666))
* resource/aws_fsx_lustre_file_system: Add `log_configuration` argument. ([#22935](https://github.com/hashicorp/terraform-provider-aws/issues/22935))
* resource/aws_fsx_ontap_file_system: Reduce the minimum valid value of the `throughput_capacity` argument to `128` (128 MB/s) ([#22898](https://github.com/hashicorp/terraform-provider-aws/issues/22898))
* resource/aws_glue_partition_index: Add support for custom timeouts. ([#22941](https://github.com/hashicorp/terraform-provider-aws/issues/22941))
* resource/aws_imagebuilder_distribution_configuration: Add `launch_template_configuration` argument to the `distribution` configuration block ([#22842](https://github.com/hashicorp/terraform-provider-aws/issues/22842))
* resource/aws_imagebuilder_image_recipe: Add `parameter` argument to the `component` configuration block ([#22837](https://github.com/hashicorp/terraform-provider-aws/issues/22837))
* resource/aws_mq_broker: `auto_minor_version_upgrade` and `host_instance_type` can be changed without recreating broker ([#20661](https://github.com/hashicorp/terraform-provider-aws/issues/20661))
* resource/aws_s3_bucket_cors_configuration: Retry when `NoSuchCORSConfiguration` errors are returned from the AWS API ([#22977](https://github.com/hashicorp/terraform-provider-aws/issues/22977))
* resource/aws_s3_bucket_versioning: Add eventual consistency handling to help ensure bucket versioning is stabilized. ([#21076](https://github.com/hashicorp/terraform-provider-aws/issues/21076))
* resource/aws_vpn_connection: Add the ability to revert changes to unconfigured tunnel options made outside of Terraform to their [documented default values](https://docs.aws.amazon.com/vpn/latest/s2svpn/VPNTunnels.html) ([#17031](https://github.com/hashicorp/terraform-provider-aws/issues/17031))
* resource/aws_vpn_connection: Mark `customer_gateway_configuration` as [`Sensitive`](https://www.terraform.io/plugin/sdkv2/best-practices/sensitive-state#using-the-sensitive-flag) ([#15806](https://github.com/hashicorp/terraform-provider-aws/issues/15806))
* resource/aws_wafv2_web_acl: Support `version` on `managed_rule_group_statement` ([#21732](https://github.com/hashicorp/terraform-provider-aws/issues/21732))

BUG FIXES:

* data-source/aws_vpc_peering_connections: Return empty array instead of error when no connections found. ([#17382](https://github.com/hashicorp/terraform-provider-aws/issues/17382))
* resource/aws_cloudformation_stack: Retry resource Create and Update for IAM eventual consistency ([#22840](https://github.com/hashicorp/terraform-provider-aws/issues/22840))
* resource/aws_cloudwatch_event_target: Preserve order of `http_target.path_parameter_values`. ([#22946](https://github.com/hashicorp/terraform-provider-aws/issues/22946))
* resource/aws_db_instance: Fix error with reboot of replica ([#22178](https://github.com/hashicorp/terraform-provider-aws/issues/22178))
* resource/aws_ec2_client_vpn_authorization_rule: Don't raise an error when `InvalidClientVpnEndpointId.NotFound` is returned during refresh ([#20688](https://github.com/hashicorp/terraform-provider-aws/issues/20688))
* resource/aws_ec2_client_vpn_endpoint: `connection_log_options.cloudwatch_log_stream` argument is Computed, preventing spurious resource diffs ([#22891](https://github.com/hashicorp/terraform-provider-aws/issues/22891))
* resource/aws_ecs_capacity_provider: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_ecs_cluster: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_ecs_service: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_ecs_task_definition: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_ecs_task_set: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_route_table_association: Handle nil 'AssociationState' in ISO regions ([#22806](https://github.com/hashicorp/terraform-provider-aws/issues/22806))
* resource/aws_route_table_association: Retry resource Read for EC2 eventual consistency ([#22927](https://github.com/hashicorp/terraform-provider-aws/issues/22927))
* resource/aws_vpc_ipam: Correct update of `description` ([#22863](https://github.com/hashicorp/terraform-provider-aws/issues/22863))
* resource/aws_waf_rule_group: Prevent panic when expanding the rule group's set of `activated_rule` ([#22978](https://github.com/hashicorp/terraform-provider-aws/issues/22978))
* resource/aws_wafregional_rule_group: Prevent panic when expanding the rule group's set of `activated_rule` ([#22978](https://github.com/hashicorp/terraform-provider-aws/issues/22978))

## 3.75.2 (May 20, 2022)

ENHANCEMENTS:

* resource/aws_lambda_function: Add support for `nodejs16.x` `runtime` value ([#24874](https://github.com/hashicorp/terraform-provider-aws/issues/24874))
* resource/aws_lambda_layer_version: Add support for `nodejs16.x` `compatible_runtimes` value ([#24874](https://github.com/hashicorp/terraform-provider-aws/issues/24874))
* resource/aws_s3_bucket_website_configuration: Add `routing_rules` parameter to be used instead of `routing_rule` to support configurations with empty String values ([#24199](https://github.com/hashicorp/terraform-provider-aws/issues/24199))

## 3.75.1 (March 24, 2022)

BUG FIXES:

* resource/aws_route_table_association: Retry resource Read for EC2 eventual consistency ([#23806](https://github.com/hashicorp/terraform-provider-aws/issues/23806))

## 3.75.0 (March 18, 2022)

NOTES:

* resource/aws_s3_bucket: The `acceleration_status` argument has been deprecated. Use the `aws_s3_bucket_accelerate_configuration` resource instead. ([#23471](https://github.com/hashicorp/terraform-provider-aws/issues/23471))
* resource/aws_s3_bucket: The `acl` and `grant` arguments have been deprecated. Use the `aws_s3_bucket_acl` resource instead. ([#23419](https://github.com/hashicorp/terraform-provider-aws/issues/23419))
* resource/aws_s3_bucket: The `cors_rule` argument has been deprecated. Use the `aws_s3_bucket_cors_configuration` resource instead. ([#23434](https://github.com/hashicorp/terraform-provider-aws/issues/23434))
* resource/aws_s3_bucket: The `lifecycle_rule` argument has been deprecated. Use the `aws_s3_bucket_lifecycle_configuration` resource instead. ([#23445](https://github.com/hashicorp/terraform-provider-aws/issues/23445))
* resource/aws_s3_bucket: The `logging` argument has been deprecated. Use the `aws_s3_bucket_logging` resource instead. ([#23430](https://github.com/hashicorp/terraform-provider-aws/issues/23430))
* resource/aws_s3_bucket: The `object_lock_configuration.object_lock_enabled` argument has been deprecated. Use the top-level argument `object_lock_enabled` instead. ([#23449](https://github.com/hashicorp/terraform-provider-aws/issues/23449))
* resource/aws_s3_bucket: The `object_lock_configuration.rule` argument has been deprecated. Use the `aws_s3_bucket_object_lock_configuration` resource instead. ([#23449](https://github.com/hashicorp/terraform-provider-aws/issues/23449))
* resource/aws_s3_bucket: The `replication_configuration` argument has been deprecated. Use the `aws_s3_bucket_replication_configuration` resource instead. ([#23716](https://github.com/hashicorp/terraform-provider-aws/issues/23716))
* resource/aws_s3_bucket: The `request_payer` argument has been deprecated. Use the `aws_s3_bucket_request_payment_configuration` resource instead. ([#23473](https://github.com/hashicorp/terraform-provider-aws/issues/23473))
* resource/aws_s3_bucket: The `server_side_encryption_configuration` argument has been deprecated. Use the `aws_s3_bucket_server_side_encryption_configuration` resource instead. ([#23476](https://github.com/hashicorp/terraform-provider-aws/issues/23476))
* resource/aws_s3_bucket: The `versioning` argument has been deprecated. Use the `aws_s3_bucket_versioning` resource instead. ([#23432](https://github.com/hashicorp/terraform-provider-aws/issues/23432))
* resource/aws_s3_bucket: The `website`, `website_domain`, and `website_endpoint` arguments have been deprecated. Use the `aws_s3_bucket_website_configuration` resource instead. ([#23435](https://github.com/hashicorp/terraform-provider-aws/issues/23435))

FEATURES:

* **New Resource:** `aws_s3_bucket_accelerate_configuration` ([#23471](https://github.com/hashicorp/terraform-provider-aws/issues/23471))
* **New Resource:** `aws_s3_bucket_acl` ([#23419](https://github.com/hashicorp/terraform-provider-aws/issues/23419))
* **New Resource:** `aws_s3_bucket_cors_configuration` ([#23434](https://github.com/hashicorp/terraform-provider-aws/issues/23434))
* **New Resource:** `aws_s3_bucket_lifecycle_configuration` ([#23445](https://github.com/hashicorp/terraform-provider-aws/issues/23445))
* **New Resource:** `aws_s3_bucket_logging` ([#23430](https://github.com/hashicorp/terraform-provider-aws/issues/23430))
* **New Resource:** `aws_s3_bucket_object_lock_configuration` ([#23449](https://github.com/hashicorp/terraform-provider-aws/issues/23449))
* **New Resource:** `aws_s3_bucket_request_payment_configuration` ([#23473](https://github.com/hashicorp/terraform-provider-aws/issues/23473))
* **New Resource:** `aws_s3_bucket_server_side_encryption_configuration` ([#23476](https://github.com/hashicorp/terraform-provider-aws/issues/23476))
* **New Resource:** `aws_s3_bucket_versioning` ([#23432](https://github.com/hashicorp/terraform-provider-aws/issues/23432))
* **New Resource:** `aws_s3_bucket_website_configuration` ([#23435](https://github.com/hashicorp/terraform-provider-aws/issues/23435))

ENHANCEMENTS:

* resource/aws_lambda_function: Add support for `dotnet6` `runtime` value ([#23670](https://github.com/hashicorp/terraform-provider-aws/issues/23670))
* resource/aws_lambda_layer_version: Add support for `dotnet6` `compatible_runtimes` value ([#23670](https://github.com/hashicorp/terraform-provider-aws/issues/23670))
* resource/aws_s3_bucket: Add top-level `object_lock_enabled` parameter ([#23449](https://github.com/hashicorp/terraform-provider-aws/issues/23449))
* resource/aws_s3_bucket_acl: Support resource import for S3 bucket names consisting of uppercase letters, underscores, and a maximum of 255 characters ([#23679](https://github.com/hashicorp/terraform-provider-aws/issues/23679))
* resource/aws_s3_bucket_lifecycle_configuration: Support empty string filtering (default behavior of the `aws_s3_bucket.lifecycle_rule` parameter in provider versions prior to v4.0) ([#23750](https://github.com/hashicorp/terraform-provider-aws/issues/23750))
* resource/aws_s3_bucket_replication_configuration: Add `token` field to specify
x-amz-bucket-object-lock-token for enabling replication on object lock enabled
buckets or enabling object lock on an existing bucket. ([#23716](https://github.com/hashicorp/terraform-provider-aws/issues/23716))
* resource/aws_s3_bucket_versioning: Add missing support for `Disabled` bucket versioning ([#23731](https://github.com/hashicorp/terraform-provider-aws/issues/23731))

BUG FIXES:

* resource/aws_s3_bucket: Prevent panic when expanding the bucket's list of `cors_rule` ([#7547](https://github.com/hashicorp/terraform-provider-aws/issues/7547))
* resource/aws_s3_bucket_replication_configuration: Change `rule` configuration block to list instead of set ([#23737](https://github.com/hashicorp/terraform-provider-aws/issues/23737))
* resource/aws_s3_bucket_replication_configuration: Correctly configure empty `rule.filter` configuration block in API requests ([#23716](https://github.com/hashicorp/terraform-provider-aws/issues/23716))
* resource/aws_s3_bucket_replication_configuration: Ensure both `key` and `value` arguments of the `rule.filter.tag` configuration block are correctly populated in the outgoing API request and terraform state. ([#23716](https://github.com/hashicorp/terraform-provider-aws/issues/23716))
* resource/aws_s3_bucket_replication_configuration: Prevent inconsistent final plan when `rule.filter.prefix` is an empty string ([#23716](https://github.com/hashicorp/terraform-provider-aws/issues/23716))
* resource/aws_s3_bucket_replication_configuration: Set `rule.id` as Computed to prevent drift when the value is not configured ([#23737](https://github.com/hashicorp/terraform-provider-aws/issues/23737))

## 3.74.3 (February 17, 2022)

BUG FIXES:

* resource/aws_ecs_capacity_provider: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_ecs_cluster: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_ecs_service: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_ecs_task_definition: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_ecs_task_set: Fix tagging error preventing use in ISO partitions ([#23030](https://github.com/hashicorp/terraform-provider-aws/issues/23030))
* resource/aws_waf_rule_group: Prevent panic when expanding the rule group's set of `activated_rule` ([#22978](https://github.com/hashicorp/terraform-provider-aws/issues/22978))
* resource/aws_wafregional_rule_group: Prevent panic when expanding the rule group's set of `activated_rule` ([#22978](https://github.com/hashicorp/terraform-provider-aws/issues/22978))

## 3.74.2 (February 11, 2022)

BUG FIXES:

* resource/aws_rds_cluster: Fix crash when configured `engine_version` string is shorter than the `EngineVersion` string returned from the AWS API ([#23039](https://github.com/hashicorp/terraform-provider-aws/issues/23039))
* resource/aws_vpn_connection: Add support for `ipsec.1-aes256` connection type ([#23127](https://github.com/hashicorp/terraform-provider-aws/issues/23127))

## 3.74.1 (February 7, 2022)

BUG FIXES:

* resource/aws_backup_selection: Fix permanent diffs for `condition` and `not_resources` arguments causing resource recreation ([#22882](https://github.com/hashicorp/terraform-provider-aws/issues/22882))

## 3.74.0 (January 28, 2022)

FEATURES:

* **New Data Source:** `aws_api_gateway_export` ([#22731](https://github.com/hashicorp/terraform-provider-aws/issues/22731))
* **New Data Source:** `aws_api_gateway_sdk` ([#22731](https://github.com/hashicorp/terraform-provider-aws/issues/22731))
* **New Data Source:** `aws_apigatewayv2_export` ([#22732](https://github.com/hashicorp/terraform-provider-aws/issues/22732))
* **New Data Source:** `aws_connect_contact_flow_module` ([#22518](https://github.com/hashicorp/terraform-provider-aws/issues/22518))
* **New Data Source:** `aws_connect_prompt` ([#22636](https://github.com/hashicorp/terraform-provider-aws/issues/22636))
* **New Data Source:** `aws_connect_quick_connect` ([#22527](https://github.com/hashicorp/terraform-provider-aws/issues/22527))
* **New Data Source:** `aws_datapipeline_pipeline` ([#22597](https://github.com/hashicorp/terraform-provider-aws/issues/22597))
* **New Data Source:** `aws_datapipeline_pipeline_definition` ([#22597](https://github.com/hashicorp/terraform-provider-aws/issues/22597))
* **New Data Source:** `aws_imagebuilder_components` ([#21881](https://github.com/hashicorp/terraform-provider-aws/issues/21881))
* **New Data Source:** `aws_imagebuilder_distribution_configurations` ([#22733](https://github.com/hashicorp/terraform-provider-aws/issues/22733))
* **New Data Source:** `aws_imagebuilder_infrastructure_configurations` ([#22723](https://github.com/hashicorp/terraform-provider-aws/issues/22723))
* **New Resource:** `aws_connect_queue` ([#22566](https://github.com/hashicorp/terraform-provider-aws/issues/22566))
* **New Resource:** `aws_connect_security_profile` ([#22369](https://github.com/hashicorp/terraform-provider-aws/issues/22369))
* **New Resource:** `aws_dataexchange_data_set` ([#22697](https://github.com/hashicorp/terraform-provider-aws/issues/22697))
* **New Resource:** `aws_datapipeline_pipeline_definition` ([#22597](https://github.com/hashicorp/terraform-provider-aws/issues/22597))
* **New Resource:** `aws_devicefarm_test_grid_project` ([#22688](https://github.com/hashicorp/terraform-provider-aws/issues/22688))
* **New Resource:** `aws_ecs_cluster_capacity_providers` ([#22672](https://github.com/hashicorp/terraform-provider-aws/issues/22672))
* **New Resource:** `aws_sagemaker_project` ([#21534](https://github.com/hashicorp/terraform-provider-aws/issues/21534))

ENHANCEMENTS:

* resource/aws_api_gateway_stage: Add `web_acl_arn` attribute ([#18561](https://github.com/hashicorp/terraform-provider-aws/issues/18561))
* resource/aws_elasticache_replication_group: Add `user_group_ids` to associate `aws_elasticache_user_group` with `aws_elasticache_replication_group` ([#20406](https://github.com/hashicorp/terraform-provider-aws/issues/20406))
* resource/aws_imagebuilder_distribution_configuration: Add `container_distribution_configuration` argument ([#22758](https://github.com/hashicorp/terraform-provider-aws/issues/22758))
* resource/aws_iot_role_alias: Increase the maximum allowed value of the `credential_duration` argument to `43200` (12 hours) ([#22757](https://github.com/hashicorp/terraform-provider-aws/issues/22757))
* resource/aws_network_interface: Add `private_ip_list`, `private_ip_list_enabled`, `ipv6_address_list`, and `ipv6_address_list_enabled` attributes ([#17846](https://github.com/hashicorp/terraform-provider-aws/issues/17846))
* resource/aws_s3_bucket_notification: Add `eventbridge` argument ([#22045](https://github.com/hashicorp/terraform-provider-aws/issues/22045))
* resource/aws_vpc_endpoint_subnet_association: Fix resource importing ([#22796](https://github.com/hashicorp/terraform-provider-aws/issues/22796))

BUG FIXES:

* data-source/aws_ecr_repository: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* data-source/aws_lb: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* data-source/aws_lb: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* data-source/aws_lb_listener: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* data-source/aws_lb_target_group: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* data-source/aws_sqs_queue: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* data-source/aws_vpc: Suppress errors if main route table cannot be found ([#22724](https://github.com/hashicorp/terraform-provider-aws/issues/22724))
* resource/aws_cloudfront_distribution: Increase the maximum valid `origin_keepalive_timeout` value to `180` ([#22632](https://github.com/hashicorp/terraform-provider-aws/issues/22632))
* resource/aws_cloudwatch_composite_alarm: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_cloudwatch_event_bus: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_cloudwatch_event_rule: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_cloudwatch_metric_alarm: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_cloudwatch_metric_stream: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_ecr_repository: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* resource/aws_ecs_capacity_provider: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* resource/aws_ecs_cluster: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* resource/aws_ecs_cluster: Provide new resource `aws_ecs_cluster_capacity_providers` to avoid bugs using `capacity_providers` and `default_capacity_provider_strategy`, which arguments will be deprecated in a future version ([#22672](https://github.com/hashicorp/terraform-provider-aws/issues/22672))
* resource/aws_ecs_service: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* resource/aws_ecs_task_definition: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* resource/aws_ecs_task_set: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* resource/aws_instance: Prevent panic when reading the instance's block device mappings ([#22719](https://github.com/hashicorp/terraform-provider-aws/issues/22719))
* resource/aws_internet_gateway: No longer give up before the attachment timeout (4m) is exceeded (previously it was giving up after 20 not found checks). ([#22713](https://github.com/hashicorp/terraform-provider-aws/issues/22713))
* resource/aws_lambda_function: Prevent errors when attempting to configure code signing in the `ap-southeast-3` AWS Region ([#22693](https://github.com/hashicorp/terraform-provider-aws/issues/22693))
* resource/aws_lb: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_lb_listener: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_lb_listener_rule: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_lb_target_group: Further refine tag error handling for ISO regions ([#22717](https://github.com/hashicorp/terraform-provider-aws/issues/22717))
* resource/aws_sns_topic: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* resource/aws_sqs_queue: Further refine tag error handling in ISO partitions ([#22780](https://github.com/hashicorp/terraform-provider-aws/issues/22780))
* resource/aws_vpc: Suppress errors if main route table, default NACL or default security group cannot be found ([#22724](https://github.com/hashicorp/terraform-provider-aws/issues/22724))
* resource/aws_vpc_dhcp_options_association: Support `default` DHCP Options ID ([#22722](https://github.com/hashicorp/terraform-provider-aws/issues/22722))

## 3.73.0 (January 21, 2022)

FEATURES:

* **New Data Source:** `aws_cloudfront_origin_access_identity` ([#22572](https://github.com/hashicorp/terraform-provider-aws/issues/22572))
* **New Data Source:** `aws_vpc_ipam_preview_next_cidr` ([#22643](https://github.com/hashicorp/terraform-provider-aws/issues/22643))
* **New Resource:** `aws_appsync_api_cache` ([#22578](https://github.com/hashicorp/terraform-provider-aws/issues/22578))
* **New Resource:** `aws_appsync_domain_name` ([#22487](https://github.com/hashicorp/terraform-provider-aws/issues/22487))
* **New Resource:** `aws_appsync_domain_name_api_association` ([#22487](https://github.com/hashicorp/terraform-provider-aws/issues/22487))
* **New Resource:** `aws_cloudsearch_domain` ([#17723](https://github.com/hashicorp/terraform-provider-aws/issues/17723))
* **New Resource:** `aws_cloudsearch_domain_service_access_policy` ([#17723](https://github.com/hashicorp/terraform-provider-aws/issues/17723))
* **New Resource:** `aws_detective_invitation_accepter` ([#22163](https://github.com/hashicorp/terraform-provider-aws/issues/22163))
* **New Resource:** `aws_detective_member` ([#22163](https://github.com/hashicorp/terraform-provider-aws/issues/22163))
* **New Resource:** `aws_fsx_data_repository_association` ([#22291](https://github.com/hashicorp/terraform-provider-aws/issues/22291))
* **New Resource:** `aws_lambda_invocation` ([#19488](https://github.com/hashicorp/terraform-provider-aws/issues/19488))


ENHANCEMENTS:

* data-source/aws_cognito_user_pool_clients: Add `client_names` attribute ([#22615](https://github.com/hashicorp/terraform-provider-aws/issues/22615))
* data-source/aws_imagebuilder_image_recipe: Add `user_data_base64` attribute ([#21763](https://github.com/hashicorp/terraform-provider-aws/issues/21763))
* resource/aws_dynamodb_table: Add special case handling when switching `billing_mode` from `PAY_PER_REQUEST` to `PROVISIONED` and provisioned throughput is ignored. ([#22630](https://github.com/hashicorp/terraform-provider-aws/issues/22630))
* resource/aws_fsx_lustre_file_system: Add `file_system_type_version` argument ([#22291](https://github.com/hashicorp/terraform-provider-aws/issues/22291))
* resource/aws_imagebuilder_image_recipe: Add `user_data_base64` argument ([#21763](https://github.com/hashicorp/terraform-provider-aws/issues/21763))
* resource/aws_opsworks_custom_layer: Add plan time validation for `ebs_volume.type` and `custom_json`. ([#12433](https://github.com/hashicorp/terraform-provider-aws/issues/12433))
* resource/aws_opsworks_custom_layer: Add support for `cloudwatch_configuration` ([#12433](https://github.com/hashicorp/terraform-provider-aws/issues/12433))
* resource/aws_security_group: Ensure that the Security Group is found 3 times in a row before declaring that it has been created ([#22420](https://github.com/hashicorp/terraform-provider-aws/issues/22420))

BUG FIXES:

* resource/aws_apprunner_custom_domain_association: Add the status `binding_certificate` as a valid target when waiting for creation. ([#20222](https://github.com/hashicorp/terraform-provider-aws/issues/20222))
* resource/aws_cloudfront_distribution: Increase the maximum valid `origin_keepalive_timeout` value to `180` ([#22632](https://github.com/hashicorp/terraform-provider-aws/issues/22632))
* resource/aws_ecr_lifecycle_policy: Fix diffs in `policy` when no changes are detected ([#22665](https://github.com/hashicorp/terraform-provider-aws/issues/22665))
* resource/aws_load_balancer_policy: Suppress `policy_attribute` differences ([#21776](https://github.com/hashicorp/terraform-provider-aws/issues/21776))

## 3.72.0 (January 13, 2022)

FEATURES:

* **New Data Source:** `aws_cognito_user_pool_client` ([#22477](https://github.com/hashicorp/terraform-provider-aws/issues/22477))
* **New Resource:** `aws_cognito_identity_pool_provider_principal_tag` ([#22514](https://github.com/hashicorp/terraform-provider-aws/issues/22514))
* **New Resource:** `aws_connect_contact_flow_module` ([#22349](https://github.com/hashicorp/terraform-provider-aws/issues/22349))
* **New Resource:** `aws_connect_quick_connect` ([#22250](https://github.com/hashicorp/terraform-provider-aws/issues/22250))
* **New Resource:** `aws_devicefarm_instance_profile` ([#22458](https://github.com/hashicorp/terraform-provider-aws/issues/22458))
* **New Resource:** `aws_memorydb_snapshot` ([#22486](https://github.com/hashicorp/terraform-provider-aws/issues/22486))
* **New Resource:** `aws_shield_protection_health_check_association` ([#21993](https://github.com/hashicorp/terraform-provider-aws/issues/21993))

ENHANCEMENTS:

* data-source/aws_cloudfront_distribution: Add `aliases` attribute ([#22552](https://github.com/hashicorp/terraform-provider-aws/issues/22552))
* data-source/aws_customer_gateway: Add `certificate_arn` attribute ([#22435](https://github.com/hashicorp/terraform-provider-aws/issues/22435))
* data-source/aws_ebs_snapshot: Add `storage_tier` and `outpost_arn` attributes. ([#22342](https://github.com/hashicorp/terraform-provider-aws/issues/22342))
* data-source/aws_ecr_repository: Allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22535](https://github.com/hashicorp/terraform-provider-aws/issues/22535))
* data-source/aws_eks_cluster: Add `ip_family` to the `kubernetes_network_config` configuration block ([#22485](https://github.com/hashicorp/terraform-provider-aws/issues/22485))
* data-source/aws_elb_service_account: Add account ID for `ap-southeast-3` AWS Region ([#22453](https://github.com/hashicorp/terraform-provider-aws/issues/22453))
* data-source/aws_iam_role: Allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22544](https://github.com/hashicorp/terraform-provider-aws/issues/22544))
* data-source/aws_iam_user: Allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22544](https://github.com/hashicorp/terraform-provider-aws/issues/22544))
* data-source/aws_instance: Add the `instance_metadata_tags` attribute to the `metadata_options` configuration block ([#22463](https://github.com/hashicorp/terraform-provider-aws/issues/22463))
* data-source/aws_launch_template: Add the `instance_metadata_tags` attribute to the `metadata_options` configuration block ([#22463](https://github.com/hashicorp/terraform-provider-aws/issues/22463))
* data-source/aws_lb: Allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22551](https://github.com/hashicorp/terraform-provider-aws/issues/22551))
* data-source/aws_lb_listener: Allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22551](https://github.com/hashicorp/terraform-provider-aws/issues/22551))
* data-source/aws_lb_target_group: Allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22551](https://github.com/hashicorp/terraform-provider-aws/issues/22551))
* data-source/aws_sagemaker_prebuilt_ecr_image: Add account IDs for the BlazingText image in `af-south-1` and `eu-south-1` AWS Regions ([#22455](https://github.com/hashicorp/terraform-provider-aws/issues/22455))
* data-source/aws_sagemaker_prebuilt_ecr_image: Add account IDs for the DeepAR Forecasting image in `af-south-1` and `eu-south-1` AWS Regions ([#22455](https://github.com/hashicorp/terraform-provider-aws/issues/22455))
* data-source/aws_sagemaker_prebuilt_ecr_image: Add account IDs for the Factorization Machines image in `af-south-1`, `ap-northeast-3` and `eu-south-1` AWS Regions ([#22455](https://github.com/hashicorp/terraform-provider-aws/issues/22455))
* data-source/aws_sagemaker_prebuilt_ecr_image: Add account IDs for the Spark ML Serving image in `af-south-1`, `ap-east-1`, `cn-north-1`, `cn-northwest-1`, `eu-north-1`, `eu-south-1`, `eu-west-3`, `me-south-1` and `sa-east-1` AWS Regions ([#22455](https://github.com/hashicorp/terraform-provider-aws/issues/22455))
* data-source/aws_sagemaker_prebuilt_ecr_image: Add account IDs for the XGBoost image in `af-south-1`, `ap-northeast-3` and `eu-south-1` AWS Regions ([#22455](https://github.com/hashicorp/terraform-provider-aws/issues/22455))
* data-source/aws_sqs_queue: Allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22516](https://github.com/hashicorp/terraform-provider-aws/issues/22516))
* resource/aws_appsync_datasource: Add `authorization_config` attribute to the `http_config` configuration block ([#22411](https://github.com/hashicorp/terraform-provider-aws/issues/22411))
* resource/aws_appsync_datasource: Add `delta_sync_config` and `versioned` to the `dynamodb_config` configuration block ([#22411](https://github.com/hashicorp/terraform-provider-aws/issues/22411))
* resource/aws_appsync_datasource: Add `relational_database_config` argument ([#22411](https://github.com/hashicorp/terraform-provider-aws/issues/22411))
* resource/aws_appsync_datasource: Add plan time validation for `service_role_arn` and `lambda_config.function_arn` ([#22411](https://github.com/hashicorp/terraform-provider-aws/issues/22411))
* resource/aws_appsync_function: Add `max_batch_size` and `sync_config` arguments. ([#22484](https://github.com/hashicorp/terraform-provider-aws/issues/22484))
* resource/aws_appsync_resolver: Add `max_batch_size` and `sync_config` arguments. ([#22510](https://github.com/hashicorp/terraform-provider-aws/issues/22510))
* resource/aws_backup_selection: Add `condition` configuration block and `not_resources` argument in support of fine-grained backup plan [resource assignment](https://docs.aws.amazon.com/aws-backup/latest/devguide/assigning-resources.html) ([#22074](https://github.com/hashicorp/terraform-provider-aws/issues/22074))
* resource/aws_cloudwatch_composite_alarm: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22556](https://github.com/hashicorp/terraform-provider-aws/issues/22556))
* resource/aws_cloudwatch_event_bus: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22550](https://github.com/hashicorp/terraform-provider-aws/issues/22550))
* resource/aws_cloudwatch_event_rule: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22550](https://github.com/hashicorp/terraform-provider-aws/issues/22550))
* resource/aws_cloudwatch_log_destination_policy: Add `force_update` argument. ([#22460](https://github.com/hashicorp/terraform-provider-aws/issues/22460))
* resource/aws_cloudwatch_log_destination_policy: Add plan time validation for `access_policy`. ([#22460](https://github.com/hashicorp/terraform-provider-aws/issues/22460))
* resource/aws_cloudwatch_metric_alarm: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22556](https://github.com/hashicorp/terraform-provider-aws/issues/22556))
* resource/aws_cloudwatch_metric_stream: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22556](https://github.com/hashicorp/terraform-provider-aws/issues/22556))
* resource/aws_connect_contact_flow: add delete function ([#22303](https://github.com/hashicorp/terraform-provider-aws/issues/22303))
* resource/aws_customer_gateway: Add `certificate_arn` argument ([#22435](https://github.com/hashicorp/terraform-provider-aws/issues/22435))
* resource/aws_ebs_snapshot: Add `outpost_arn`, `storage_tier`, `permanent_restore`, `temporary_restore_days` arguments ([#22342](https://github.com/hashicorp/terraform-provider-aws/issues/22342))
* resource/aws_ebs_snapshot_copy: Add `storage_tier`, `permanent_restore`, `temporary_restore_days` arguments ([#22342](https://github.com/hashicorp/terraform-provider-aws/issues/22342))
* resource/aws_ebs_snapshot_import: Add `storage_tier`, `permanent_restore`, `temporary_restore_days` arguments ([#22342](https://github.com/hashicorp/terraform-provider-aws/issues/22342))
* resource/aws_ecr_repository: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22535](https://github.com/hashicorp/terraform-provider-aws/issues/22535))
* resource/aws_ecs_capacity_provider: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22529](https://github.com/hashicorp/terraform-provider-aws/issues/22529))
* resource/aws_ecs_cluster: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22529](https://github.com/hashicorp/terraform-provider-aws/issues/22529))
* resource/aws_ecs_service: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22529](https://github.com/hashicorp/terraform-provider-aws/issues/22529))
* resource/aws_ecs_task_definition: Add `skip_destroy` argument to optionally prevent overwriting previous revision ([#22269](https://github.com/hashicorp/terraform-provider-aws/issues/22269))
* resource/aws_ecs_task_definition: Add plan time validation for `family` ([#18610](https://github.com/hashicorp/terraform-provider-aws/issues/18610))
* resource/aws_ecs_task_definition: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22529](https://github.com/hashicorp/terraform-provider-aws/issues/22529))
* resource/aws_ecs_task_set: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22529](https://github.com/hashicorp/terraform-provider-aws/issues/22529))
* resource/aws_eks_cluster: Add `ip_family` to the `kubernetes_network_config` configuration block ([#22485](https://github.com/hashicorp/terraform-provider-aws/issues/22485))
* resource/aws_glue_crawler: add `delta_target` argument. ([#22472](https://github.com/hashicorp/terraform-provider-aws/issues/22472))
* resource/aws_iam_role: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22544](https://github.com/hashicorp/terraform-provider-aws/issues/22544))
* resource/aws_iam_user: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22544](https://github.com/hashicorp/terraform-provider-aws/issues/22544))
* resource/aws_instance: Add the `instance_metadata_tags` argument to the `metadata_options` configuration block ([#22463](https://github.com/hashicorp/terraform-provider-aws/issues/22463))
* resource/aws_launch_template: Add the `instance_metadata_tags` argument to the `metadata_options` configuration block ([#22463](https://github.com/hashicorp/terraform-provider-aws/issues/22463))
* resource/aws_lb: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22551](https://github.com/hashicorp/terraform-provider-aws/issues/22551))
* resource/aws_lb_listener: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22551](https://github.com/hashicorp/terraform-provider-aws/issues/22551))
* resource/aws_lb_listener_rule: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22551](https://github.com/hashicorp/terraform-provider-aws/issues/22551))
* resource/aws_lb_target_group: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22551](https://github.com/hashicorp/terraform-provider-aws/issues/22551))
* resource/aws_s3_bucket: Add additional protection against `object_lock_configuration` causing errors in partitions (e.g., ISO) where not supported ([#22575](https://github.com/hashicorp/terraform-provider-aws/issues/22575))
* resource/aws_sns_topic: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22511](https://github.com/hashicorp/terraform-provider-aws/issues/22511))
* resource/aws_sqs_queue: Attempt `tags`-on-create, fallback to tag after create, and allow some `tags` errors to be non-fatal to support non-standard AWS partitions (i.e., ISO) ([#22516](https://github.com/hashicorp/terraform-provider-aws/issues/22516))
* resource/aws_vpc: Add `ipv6_cidr_block_network_border_group` argument ([#22211](https://github.com/hashicorp/terraform-provider-aws/issues/22211))
* resource/aws_vpc_ipam_pool_cidr_allocation: Add `disallowed_cidrs` argument ([#22470](https://github.com/hashicorp/terraform-provider-aws/issues/22470))
* resource/aws_vpc_ipam_preview_next_cidr: Add `disallowed_cidrs` argument ([#22501](https://github.com/hashicorp/terraform-provider-aws/issues/22501))
* resource/aws_vpn_connection: Add `vgw_telemetry.certificate_arn` attribute ([#19311](https://github.com/hashicorp/terraform-provider-aws/issues/19311))
* resource/aws_vpn_connection: `customer_gateway_id`, `transit_gateway_id` and `vpn_gateway_id` can be updated without recreating the resource ([#19311](https://github.com/hashicorp/terraform-provider-aws/issues/19311))
* resource/aws_vpn_connection: `tunnel1_preshared_key` and `tunnel2_preshared_key` can be updated without recreating the resource ([#19311](https://github.com/hashicorp/terraform-provider-aws/issues/19311))

BUG FIXES:

* data-source/aws_vpc_ipam_pool: Return an error if more than 1 IPAM Pool matches ([#22438](https://github.com/hashicorp/terraform-provider-aws/issues/22438))
* data-source/aws_vpc_ipam_pool: Set `address_family`, `allocation_default_netmask_length`, `allocation_max_netmask_length`, `allocation_min_netmask_length` and `tags` attributes ([#22438](https://github.com/hashicorp/terraform-provider-aws/issues/22438))
* resource/aws_cloudfront_distribution: Increase the maximum valid `origin_read_timeout` value to `180` ([#22461](https://github.com/hashicorp/terraform-provider-aws/issues/22461))
* resource/aws_fsx_lustre_file_system: Add missing values to `per_unit_storage_throughput` validation ([#22462](https://github.com/hashicorp/terraform-provider-aws/issues/22462))
* resource/aws_fsx_openzfs_file_system: Change `root_volume_configuration.copy_tags_to_snapshots` to ForceNew ([#22480](https://github.com/hashicorp/terraform-provider-aws/issues/22480))
* resource/aws_fsx_openzfs_file_system: Fix crash with nil `root_volume_configuration.nfs_exports` value ([#22480](https://github.com/hashicorp/terraform-provider-aws/issues/22480))
* resource/aws_memorydb_cluster: Correctly propagate configurable timeouts to waiters. ([#22489](https://github.com/hashicorp/terraform-provider-aws/issues/22489))
* resource/aws_route53_record: Fix import with underscores in names ([#21556](https://github.com/hashicorp/terraform-provider-aws/issues/21556))
* resource/aws_sqs_queue: Don't timeout when a queue policy `Condition` value contains an empty array ([#22547](https://github.com/hashicorp/terraform-provider-aws/issues/22547))
* resource/aws_ssm_parameter: Mark `version` as Computed when `value` changes ([#22522](https://github.com/hashicorp/terraform-provider-aws/issues/22522))
* resource/aws_subnet: Protect against errors when `availability_zone_id` is not supported in a partition (e.g., ISO) ([#22580](https://github.com/hashicorp/terraform-provider-aws/issues/22580))
* resource/aws_subnet: Resource-based naming is not available in the `ap-southeast-3` region ([#22531](https://github.com/hashicorp/terraform-provider-aws/issues/22531))

## 3.71.0 (January 06, 2022)

FEATURES:

* **New Data Source:** `aws_batch_scheduling_policy` ([#22335](https://github.com/hashicorp/terraform-provider-aws/issues/22335))
* **New Data Source:** `aws_cognito_user_pool_clients` ([#22289](https://github.com/hashicorp/terraform-provider-aws/issues/22289))
* **New Data Source:** `aws_cognito_user_pool_signing_certificate` ([#22285](https://github.com/hashicorp/terraform-provider-aws/issues/22285))
* **New Data Source:** `aws_mskconnect_custom_plugin` ([#22333](https://github.com/hashicorp/terraform-provider-aws/issues/22333))
* **New Data Source:** `aws_mskconnect_worker_configuration` ([#22414](https://github.com/hashicorp/terraform-provider-aws/issues/22414))
* **New Data Source:** `aws_organizations_resource_tags` ([#22371](https://github.com/hashicorp/terraform-provider-aws/issues/22371))
* **New Data Source:** `aws_ses_active_receipt_rule_set` ([#22310](https://github.com/hashicorp/terraform-provider-aws/issues/22310))
* **New Data Source:** `aws_ses_domain_identity` ([#22321](https://github.com/hashicorp/terraform-provider-aws/issues/22321))
* **New Data Source:** `aws_ses_email_identity` ([#22321](https://github.com/hashicorp/terraform-provider-aws/issues/22321))
* **New Resource:** `aws_batch_scheduling_policy` ([#22262](https://github.com/hashicorp/terraform-provider-aws/issues/22262))
* **New Resource:** `aws_cloud9_environment_membership` ([#11857](https://github.com/hashicorp/terraform-provider-aws/issues/11857))
* **New Resource:** `aws_codebuild_resource_policy` ([#22196](https://github.com/hashicorp/terraform-provider-aws/issues/22196))
* **New Resource:** `aws_datasync_location_fsx_lustre_file_system` ([#22346](https://github.com/hashicorp/terraform-provider-aws/issues/22346))
* **New Resource:** `aws_datasync_location_hdfs` ([#22347](https://github.com/hashicorp/terraform-provider-aws/issues/22347))
* **New Resource:** `aws_devicefarm_device_pool` ([#21025](https://github.com/hashicorp/terraform-provider-aws/issues/21025))
* **New Resource:** `aws_devicefarm_network_profile` ([#22448](https://github.com/hashicorp/terraform-provider-aws/issues/22448))
* **New Resource:** `aws_devicefarm_upload` ([#22443](https://github.com/hashicorp/terraform-provider-aws/issues/22443))
* **New Resource:** `aws_fsx_openzfs_file_system` ([#22234](https://github.com/hashicorp/terraform-provider-aws/issues/22234))
* **New Resource:** `aws_fsx_openzfs_snapshot` ([#22234](https://github.com/hashicorp/terraform-provider-aws/issues/22234))
* **New Resource:** `aws_fsx_openzfs_volume` ([#22234](https://github.com/hashicorp/terraform-provider-aws/issues/22234))
* **New Resource:** `aws_memorydb_cluster` ([#22388](https://github.com/hashicorp/terraform-provider-aws/issues/22388))
* **New Resource:** `aws_memorydb_parameter_group` ([#22304](https://github.com/hashicorp/terraform-provider-aws/issues/22304))
* **New Resource:** `aws_memorydb_subnet_group` ([#22256](https://github.com/hashicorp/terraform-provider-aws/issues/22256))
* **New Resource:** `aws_memorydb_user` ([#22261](https://github.com/hashicorp/terraform-provider-aws/issues/22261))
* **New Resource:** `aws_mskconnect_custom_plugin` ([#22333](https://github.com/hashicorp/terraform-provider-aws/issues/22333))
* **New Resource:** `aws_mskconnect_worker_configuration` ([#22414](https://github.com/hashicorp/terraform-provider-aws/issues/22414))
* **New Resource:** `aws_sagemaker_device` ([#22427](https://github.com/hashicorp/terraform-provider-aws/issues/22427))
* **New Resource:** `aws_vpc_endpoint_connection_accepter` ([#19083](https://github.com/hashicorp/terraform-provider-aws/issues/19083))
* **New Resource:** `aws_vpc_ipam_organization_admin_account` ([#22394](https://github.com/hashicorp/terraform-provider-aws/issues/22394))

ENHANCEMENTS:

* data-source/aws_batch_job_queue: Add `scheduling_policy_arn` attribute ([#22348](https://github.com/hashicorp/terraform-provider-aws/issues/22348))
* data-source/aws_cloudtrail_service_account: Add service account ID for `ap-southeast-3` AWS Region ([#22295](https://github.com/hashicorp/terraform-provider-aws/issues/22295))
* data-source/aws_ecs_task_definition: Add `arn` attribute. ([#21856](https://github.com/hashicorp/terraform-provider-aws/issues/21856))
* data-source/aws_elb_hosted_zone_id: Add hosted zone ID for `ap-southeast-3` AWS Region ([#22295](https://github.com/hashicorp/terraform-provider-aws/issues/22295))
* data-source/aws_s3_bucket: Add hosted zone ID for `ap-southeast-3` AWS Region ([#22295](https://github.com/hashicorp/terraform-provider-aws/issues/22295))
* data-source/aws_ssm_parameters_by_path: Add `recursive` argument ([#22222](https://github.com/hashicorp/terraform-provider-aws/issues/22222))
* data-source/aws_subnet: Add `enable_dns64`, `ipv6_native`, `enable_resource_name_dns_aaaa_record_on_launch`, `enable_resource_name_dns_a_record_on_launch` and `private_dns_hostname_type_on_launch` attributes ([#22339](https://github.com/hashicorp/terraform-provider-aws/issues/22339))
* provider: Add validation for the `duration`, `external_id` and `session_name` arguments in the `assume_role` configuration block ([#18085](https://github.com/hashicorp/terraform-provider-aws/issues/18085))
* resource/aws_batch_job_queue: Add `scheduling_policy_arn` attribute ([#22298](https://github.com/hashicorp/terraform-provider-aws/issues/22298))
* resource/aws_cloud9_environment_ec2: Add plan time validations for `name`, `automatic_stop_time_minutes`, `description`. ([#18560](https://github.com/hashicorp/terraform-provider-aws/issues/18560))
* resource/aws_cloudfront_distribution: Add plan time validation to `ordered_cache_behavior.forwarded_values.cookies`, `ordered_cache_behavior.lambda_function_association.event_type`, `ordered_cache_behavior.lambda_function_association.lambda_arn`, `ordered_cache_behavior.function_association.lambda_arn`, `ordered_cache_behavior.function_association.event_type`, `ordered_cache_behavior.viewer_protocol_policy`, `comment`, `default_cache_behavior.forwarded_values.cookies`, `default_cache_behavior.lambda_function_association.event_type`, `ordered_cache_behavior.lambda_function_association.lambda_arn`, `default_cache_behavior.function_association.lambda_arn`, `default_cache_behavior.function_association.event_type`, `default_cache_behavior.viewer_protocol_policy`, `origin.custom_origin_config.origin_keepalive_timeout`, `origin.custom_origin_config.origin_read_timeout`, `origin.custom_origin_config.origin_protocol_policy`, `origin.custom_origin_config.origin_ssl_protocols`, `price_class`, `viewer_certificate.acm_certificate_arn`, `viewer_certificate.minimum_protocol_version`, `viewer_certificate.ssl_support_method`. ([#21034](https://github.com/hashicorp/terraform-provider-aws/issues/21034))
* resource/aws_codebuild_project: Add `artifacts.bucket_owner_access`, `secondary_artifacts.bucket_owner_access`, `logs_config.s3_logs.bucket_owner_access`, `project_visibility`, `resource_access_role` arguments. ([#22189](https://github.com/hashicorp/terraform-provider-aws/issues/22189))
* resource/aws_codebuild_project: Add `public_project_alias` attribute. ([#22189](https://github.com/hashicorp/terraform-provider-aws/issues/22189))
* resource/aws_codebuild_project: Add `secondary_source_version` argument ([#22345](https://github.com/hashicorp/terraform-provider-aws/issues/22345))
* resource/aws_codebuild_project: Add plan time validation for `cache.modes` and `service_role`. ([#22189](https://github.com/hashicorp/terraform-provider-aws/issues/22189))
* resource/aws_codepipeline: Add plan time validation to `name`, `role_arn`, `stage.name`, `stage.action.name`, `stage.action.name`, `stage.action.run_order`, `stage.action.namespace`, `action.configuration`, and `action.version` ([#18451](https://github.com/hashicorp/terraform-provider-aws/issues/18451))
* resource/aws_codepipeline_webhook: Add `arn` attribute. ([#22406](https://github.com/hashicorp/terraform-provider-aws/issues/22406))
* resource/aws_codepipeline_webhook: Add plan time validation for `authentication_configuration.secret_token`, `filter.json_path`, `filter.match_equals`, `name`. ([#22406](https://github.com/hashicorp/terraform-provider-aws/issues/22406))
* resource/aws_codepipeline_webhook: Allow updating `filter` in place. ([#22406](https://github.com/hashicorp/terraform-provider-aws/issues/22406))
* resource/aws_dax_cluster: Add `cluster_endpoint_encryption_type` argument ([#22396](https://github.com/hashicorp/terraform-provider-aws/issues/22396))
* resource/aws_dx_private_virtual_interface: Add `sitelink_enabled` argument ([#22350](https://github.com/hashicorp/terraform-provider-aws/issues/22350))
* resource/aws_dx_transit_virtual_interface: Add `sitelink_enabled` argument ([#22350](https://github.com/hashicorp/terraform-provider-aws/issues/22350))
* resource/aws_ecr_replication_configuration: Add `repository_filter` to `replication_configuration` block ([#21231](https://github.com/hashicorp/terraform-provider-aws/issues/21231))
* resource/aws_ecr_replication_configuration: Increase `MaxItems` for `rule` to `10` and for `destination` to `25` ([#22281](https://github.com/hashicorp/terraform-provider-aws/issues/22281))
* resource/aws_elasticsearch_domain: Tag on create ([#18082](https://github.com/hashicorp/terraform-provider-aws/issues/18082))
* resource/aws_glue_trigger: Add `start_on_creation` argument ([#22439](https://github.com/hashicorp/terraform-provider-aws/issues/22439))
* resource/aws_kinesis_firehose_delivery_stream: Add `error_output_prefix` argument to `extended_s3_configuration` `s3_backup_configuration` configuration block ([#11229](https://github.com/hashicorp/terraform-provider-aws/issues/11229))
* resource/aws_kinesis_firehose_delivery_stream: Add `error_output_prefix` argument to `redshift_configuration` `s3_backup_configuration` configuration block ([#11229](https://github.com/hashicorp/terraform-provider-aws/issues/11229))
* resource/aws_kinesis_firehose_delivery_stream: Add `error_output_prefix` argument to `s3_configuration` configuration block ([#11229](https://github.com/hashicorp/terraform-provider-aws/issues/11229))
* resource/aws_networkfirewall_resource_policy: Handle delete-after-create eventual consistency ([#22402](https://github.com/hashicorp/terraform-provider-aws/issues/22402))
* resource/aws_kinesis_stream: Improve reading kinesis stream state. ([#15489](https://github.com/hashicorp/terraform-provider-aws/issues/15489))
* resource/aws_kinesis_stream_consumer: Improve reading kinesis stream state ([#15489](https://github.com/hashicorp/terraform-provider-aws/issues/15489))
* resource/aws_s3_bucket: Add hosted zone ID for `ap-southeast-3` AWS Region ([#22295](https://github.com/hashicorp/terraform-provider-aws/issues/22295))
* resource/aws_s3_bucket_object: Support objects greater than 5GB in size by using the [Amazon S3 upload manager](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/sdk-utilities.html#upload-manager) ([#21727](https://github.com/hashicorp/terraform-provider-aws/issues/21727))
* resource/aws_sagemaker_app: Add `lifecycle_config_arn` and `sagemaker_image_version_arn` arguments to `resource_spec` configuration block ([#21508](https://github.com/hashicorp/terraform-provider-aws/issues/21508))
* resource/aws_sagemaker_domain: Add `lifecycle_config_arn` and `sagemaker_image_version_arn` arguments to `default_resource_spec` configuration block ([#21508](https://github.com/hashicorp/terraform-provider-aws/issues/21508))
* resource/aws_sagemaker_user_profile: Add `lifecycle_config_arn` and `sagemaker_image_version_arn` arguments to `default_resource_spec` configuration block ([#21508](https://github.com/hashicorp/terraform-provider-aws/issues/21508))
* resource/aws_subnet: Add `enable_dns64`, `ipv6_native`, `enable_resource_name_dns_aaaa_record_on_launch`, `enable_resource_name_dns_a_record_on_launch` and `private_dns_hostname_type_on_launch` arguments ([#22339](https://github.com/hashicorp/terraform-provider-aws/issues/22339))
* resource/aws_timestreamwrite_table: Add `magnetic_store_write_properties` argument. ([#22363](https://github.com/hashicorp/terraform-provider-aws/issues/22363))

BUG FIXES:

* resource/aws_appstream_fleet: Correctly create resource with `stream_view` argument ([#22395](https://github.com/hashicorp/terraform-provider-aws/issues/22395))
* resource/aws_codebuild_project: Fix plan validation to take into account computed values for `cache.location` ([#21458](https://github.com/hashicorp/terraform-provider-aws/issues/21458))
* resource/aws_dynamodb_table: Remove extraneous `kms_key_arn` attribute from the `ttl` configuration block ([#21334](https://github.com/hashicorp/terraform-provider-aws/issues/21334))
* resource/aws_ec2_traffic_mirror_filter_rule: Prevent crash during resource read ([#22315](https://github.com/hashicorp/terraform-provider-aws/issues/22315))
* resource/aws_launch_template: Correctly set `default_version` and `latest_version` as Computed when `name`, `name_prefix` or `description` change ([#22277](https://github.com/hashicorp/terraform-provider-aws/issues/22277))
* resource/aws_networkfirewall_rule_group: Allow any character in `ip_set` `definition` as per the AWS API docs ([#22284](https://github.com/hashicorp/terraform-provider-aws/issues/22284))
* resource/aws_ses_event_destination: Allow `.` and `@` characters in `cloudwatch_destination.default_value` argument ([#22359](https://github.com/hashicorp/terraform-provider-aws/issues/22359))
* resource/aws_ssoadmin_managed_policy_attachment: Fix missing call to `ProvisionPermissionSet` after detaching the managed policy ([#21773](https://github.com/hashicorp/terraform-provider-aws/issues/21773))
* resource/aws_vpc_ipam_pool_cidr_allocation: update `cidr` and `netmask_length` attributes netmask to a minimum of 0 and maximum of 32 ([#22418](https://github.com/hashicorp/terraform-provider-aws/issues/22418))

## 3.70.0 (December 16, 2021)

NOTES:

* resource/aws_fsx_ontap_storage_virtual_machine: The `active_directory_configuration.self_managed_active_directory_configuration.organizational_unit_distinguidshed_name` attribute has been deprecated. All configurations using `active_directory_configuration.self_managed_active_directory_configuration.organizational_unit_distinguidshed_name` should be updated to use the new `active_directory_configuration.self_managed_active_directory_configuration.organizational_unit_distinguished_name` attribute instead ([#22246](https://github.com/hashicorp/terraform-provider-aws/issues/22246))

FEATURES:

* **New Data Source:** `aws_connect_bot_association` ([#21097](https://github.com/hashicorp/terraform-provider-aws/issues/21097))
* **New Data Source:** `aws_connect_hours_of_operation` ([#22207](https://github.com/hashicorp/terraform-provider-aws/issues/22207))
* **New Data Source:** `aws_connect_lambda_function_association` ([#21276](https://github.com/hashicorp/terraform-provider-aws/issues/21276))
* **New Resource:** `aws_connect_bot_association` ([#21097](https://github.com/hashicorp/terraform-provider-aws/issues/21097))
* **New Resource:** `aws_connect_hours_of_operation` ([#21934](https://github.com/hashicorp/terraform-provider-aws/issues/21934))
* **New Resource:** `aws_connect_lambda_function_association` ([#21276](https://github.com/hashicorp/terraform-provider-aws/issues/21276))
* **New Resource:** `aws_ecr_pull_through_cache_rule` ([#22172](https://github.com/hashicorp/terraform-provider-aws/issues/22172))
* **New Resource:** `aws_ecr_registry_scanning_configuration` ([#22179](https://github.com/hashicorp/terraform-provider-aws/issues/22179))
* **New Resource:** `aws_ecrpublic_repository_policy` ([#16901](https://github.com/hashicorp/terraform-provider-aws/issues/16901))

ENHANCEMENTS:

* data-source/aws_sagemaker_prebuilt_ecr_image: Add Hugging Face DLCs ([#21983](https://github.com/hashicorp/terraform-provider-aws/issues/21983))
* resource/aws_appsync_graphql_api: Add `lambda_authorizer_config` argument ([#20857](https://github.com/hashicorp/terraform-provider-aws/issues/20857))
* resource/aws_dynamodb_table: Allows restoring to point-in-time ([#19292](https://github.com/hashicorp/terraform-provider-aws/issues/19292))
* resource/aws_fsx_backup: Add `volume_id` argument to support Amazon FSx for NetApp ONTAP backup ([#21960](https://github.com/hashicorp/terraform-provider-aws/issues/21960))
* resource/aws_networkfirewall_firewall_policy: Add `stateful_default_actions` and `stateful_engine_options` configuration blocks. Add `priority` attribute to `stateful_rule_group_reference` block ([#21955](https://github.com/hashicorp/terraform-provider-aws/issues/21955))
* resource/aws_networkfirewall_firewall_rule_group: Add `stateful_rule_options` configuration block ([#21955](https://github.com/hashicorp/terraform-provider-aws/issues/21955))
* resource/aws_route: Extend creation timeout to 5 minutes ([#21531](https://github.com/hashicorp/terraform-provider-aws/issues/21531))
* resource/aws_route_table: Extend creation timeout to 5 minutes ([#21531](https://github.com/hashicorp/terraform-provider-aws/issues/21531))
* resource/iam_service_linked_role: Add `tags` argument ([#22185](https://github.com/hashicorp/terraform-provider-aws/issues/22185))

BUG FIXES:

* data-source/aws_s3_bucket: Correct Route 53 hosted zone ID for S3 websites in the `eu-south-1`, `af-south-1` and `us-gov-east-1` AWS Regions ([#22227](https://github.com/hashicorp/terraform-provider-aws/issues/22227))
* resource/aws_cloudwatch_event_bus_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22165](https://github.com/hashicorp/terraform-provider-aws/issues/22165))
* resource/aws_ecr_lifecycle_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22142](https://github.com/hashicorp/terraform-provider-aws/issues/22142))
* resource/aws_elasticsearch_domain: Fix erroneous diffs in `access_policies` when no changes made or policies are equivalent ([#22157](https://github.com/hashicorp/terraform-provider-aws/issues/22157))
* resource/aws_elasticsearch_domain: Fix erroneous diffs in `advanced_options` due to AWS defaults being returned ([#22157](https://github.com/hashicorp/terraform-provider-aws/issues/22157))
* resource/aws_elasticsearch_domain_policy: Fix erroneous diffs in `access_policies` when no changes made or policies are equivalent ([#22157](https://github.com/hashicorp/terraform-provider-aws/issues/22157))
* resource/aws_emr_cluster: Wait for the cluster to reach a terminated state on deletion ([#12578](https://github.com/hashicorp/terraform-provider-aws/issues/12578))
* resource/aws_glacier_vault: Fix erroneous diffs in `access_policy` when no changes made or policies are equivalent ([#22166](https://github.com/hashicorp/terraform-provider-aws/issues/22166))
* resource/aws_glacier_vault_lock: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22166](https://github.com/hashicorp/terraform-provider-aws/issues/22166))
* resource/aws_glue_resource_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22167](https://github.com/hashicorp/terraform-provider-aws/issues/22167))
* resource/aws_iam_role: Fix eventual consistency problem with `arn` sometimes being a unique ID instead of the role ARN ([#22217](https://github.com/hashicorp/terraform-provider-aws/issues/22217))
* resource/aws_iot_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22169](https://github.com/hashicorp/terraform-provider-aws/issues/22169))
* resource/aws_media_store_container_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22170](https://github.com/hashicorp/terraform-provider-aws/issues/22170))
* resource/aws_networkfirewall_resource_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22171](https://github.com/hashicorp/terraform-provider-aws/issues/22171))
* resource/aws_s3_access_point: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22255](https://github.com/hashicorp/terraform-provider-aws/issues/22255))
* resource/aws_s3_bucket: Correct Route 53 hosted zone ID for S3 websites in the `eu-south-1`, `af-south-1` and `us-gov-east-1` AWS Regions ([#22227](https://github.com/hashicorp/terraform-provider-aws/issues/22227))
* resource/aws_s3_bucket: Ensure `versioning` is set correctly when nested values are explicitly set to `false`. ([#22221](https://github.com/hashicorp/terraform-provider-aws/issues/22221))
* resource/aws_s3control_access_point_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22255](https://github.com/hashicorp/terraform-provider-aws/issues/22255))
* resource/aws_s3control_bucket_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22255](https://github.com/hashicorp/terraform-provider-aws/issues/22255))
* resource/aws_s3control_multi_region_access_point_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22255](https://github.com/hashicorp/terraform-provider-aws/issues/22255))
* resource/aws_s3control_object_lambda_access_point_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22255](https://github.com/hashicorp/terraform-provider-aws/issues/22255))
* resource/aws_sagemaker_model_package_group_policy: Fix erroneous diffs in `resource_policy` when no changes made or policies are equivalent ([#22259](https://github.com/hashicorp/terraform-provider-aws/issues/22259))
* resource/aws_secretsmanager_secret: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22217](https://github.com/hashicorp/terraform-provider-aws/issues/22217))
* resource/aws_secretsmanager_secret_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22217](https://github.com/hashicorp/terraform-provider-aws/issues/22217))
* resource/aws_ses_identity_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22214](https://github.com/hashicorp/terraform-provider-aws/issues/22214))
* resource/aws_sns_topic: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22213](https://github.com/hashicorp/terraform-provider-aws/issues/22213))
* resource/aws_sns_topic_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22213](https://github.com/hashicorp/terraform-provider-aws/issues/22213))
* resource/aws_sqs_queue: Fix "error reading, empty result" and various eventual consistency errors ([#22194](https://github.com/hashicorp/terraform-provider-aws/issues/22194))
* resource/aws_sqs_queue: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22194](https://github.com/hashicorp/terraform-provider-aws/issues/22194))
* resource/aws_sqs_queue_policy: Fix "error reading, empty result" and various eventual consistency errors ([#22194](https://github.com/hashicorp/terraform-provider-aws/issues/22194))
* resource/aws_sqs_queue_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22194](https://github.com/hashicorp/terraform-provider-aws/issues/22194))
* resource/aws_ssoadmin_permission_set_inline_policy: Fix erroneous diffs in `inline_policy` when no changes made or policies are equivalent ([#22192](https://github.com/hashicorp/terraform-provider-aws/issues/22192))
* resource/aws_transfer_access: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22193](https://github.com/hashicorp/terraform-provider-aws/issues/22193))
* resource/aws_transfer_user: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22193](https://github.com/hashicorp/terraform-provider-aws/issues/22193))

## 3.69.0 (December 09, 2021)

FEATURES:

* **New Resource:** `aws_codecommit_approval_rule_template_association` ([#13467](https://github.com/hashicorp/terraform-provider-aws/issues/13467))
* **New Resource:** `aws_detective_graph` ([#22042](https://github.com/hashicorp/terraform-provider-aws/issues/22042))
* **New Resource:** `aws_ec2_subnet_cidr_reservation` ([#22051](https://github.com/hashicorp/terraform-provider-aws/issues/22051))
* **New Resource:** `aws_ecs_task_set` ([#22096](https://github.com/hashicorp/terraform-provider-aws/issues/22096))
* **New Resource:** `aws_emr_studio` ([#21855](https://github.com/hashicorp/terraform-provider-aws/issues/21855))
* **New Resource:** `aws_emr_studio_session_mapping` ([#22140](https://github.com/hashicorp/terraform-provider-aws/issues/22140))

ENHANCEMENTS:

* data-source/aws_dynamodb_table: Add `table_class` attribute ([#22110](https://github.com/hashicorp/terraform-provider-aws/issues/22110))
* resource/aws_backup_region_settings: Add `resource_type_management_preference` argument ([#22021](https://github.com/hashicorp/terraform-provider-aws/issues/22021))
* resource/aws_cloudtrail: Add plan time validations for `cloud_watch_logs_group_arn`, `cloud_watch_logs_role_arn`, `name`, `s3_key_prefix`. ([#21882](https://github.com/hashicorp/terraform-provider-aws/issues/21882))
* resource/aws_dynamodb_table: Add `table_class` argument ([#22110](https://github.com/hashicorp/terraform-provider-aws/issues/22110))
* resource/aws_ecs_task_definition: Add `runtime_platform` argument in support of Fargate for ECS Windows containers ([#22016](https://github.com/hashicorp/terraform-provider-aws/issues/22016))
* resource/aws_elasticache_replication_group: Add `data_tiering_enabled` argument ([#22066](https://github.com/hashicorp/terraform-provider-aws/issues/22066))
* resource/aws_elasticsearch_domain: Add `auto_tune_options` configuration block ([#21652](https://github.com/hashicorp/terraform-provider-aws/issues/21652))
* resource/aws_kinesis_stream: Add `stream_mode_details` argument in support of Kinesis Data Streams On-Demand ([#22002](https://github.com/hashicorp/terraform-provider-aws/issues/22002))
* resource/aws_lambda_event_source_mapping: Add `filter_criteria` argument ([#21937](https://github.com/hashicorp/terraform-provider-aws/issues/21937))
* resource/aws_sqs_queue: Add `sqs_managed_sse_enabled` argument ([#21954](https://github.com/hashicorp/terraform-provider-aws/issues/21954))
* resource/aws_transfer_server: Add `function` argument in support of custom identity providers ([#22039](https://github.com/hashicorp/terraform-provider-aws/issues/22039))

BUG FIXES:

* data-source/aws_ecs_cluster: Ensure that `setting` attribute is set consistently ([#22119](https://github.com/hashicorp/terraform-provider-aws/issues/22119))
* resource/aws_api_gateway_rest_api: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22115](https://github.com/hashicorp/terraform-provider-aws/issues/22115))
* resource/aws_api_gateway_rest_api_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22115](https://github.com/hashicorp/terraform-provider-aws/issues/22115))
* resource/aws_appstream_image_builder: Correctly create resource with `image_arn` argument ([#22077](https://github.com/hashicorp/terraform-provider-aws/issues/22077))
* resource/aws_backup_vault_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22130](https://github.com/hashicorp/terraform-provider-aws/issues/22130))
* resource/aws_cloudwatch_log_resource_policy: Fix erroneous diffs in `policy_document` when no changes made or policies are equivalent ([#22135](https://github.com/hashicorp/terraform-provider-aws/issues/22135))
* resource/aws_codeartifact_domain_permissions_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22136](https://github.com/hashicorp/terraform-provider-aws/issues/22136))
* resource/aws_codeartifact_repository_permissions_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22136](https://github.com/hashicorp/terraform-provider-aws/issues/22136))
* resource/aws_ecr_registry_policy: Fix order-related diffs in `policy` ([#22004](https://github.com/hashicorp/terraform-provider-aws/issues/22004))
* resource/aws_ecr_repository_policy: Fix order-related diffs in `policy` ([#22004](https://github.com/hashicorp/terraform-provider-aws/issues/22004))
* resource/aws_efs_file_system_policy: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22100](https://github.com/hashicorp/terraform-provider-aws/issues/22100))
* resource/aws_iam_group_policy: Fix order-related diffs in `policy` ([#22067](https://github.com/hashicorp/terraform-provider-aws/issues/22067))
* resource/aws_iam_policy: Fix order-related diffs in `policy` ([#22067](https://github.com/hashicorp/terraform-provider-aws/issues/22067))
* resource/aws_iam_role: Fix order-related diffs in `policy` ([#22099](https://github.com/hashicorp/terraform-provider-aws/issues/22099))
* resource/aws_iam_role: Prevent `arn` attribute from ever containing a unique ID immediately after role creation ([#22004](https://github.com/hashicorp/terraform-provider-aws/issues/22004))
* resource/aws_iam_role_policy: Fix order-related diffs in `policy` ([#22067](https://github.com/hashicorp/terraform-provider-aws/issues/22067))
* resource/aws_iam_user_policy: Fix order-related diffs in `policy` ([#22067](https://github.com/hashicorp/terraform-provider-aws/issues/22067))
* resource/aws_lb: Correctly configure `enable_waf_fail_open` during resource creation ([#22072](https://github.com/hashicorp/terraform-provider-aws/issues/22072))
* resource/aws_redshift_cluster: Adds retries to enabling and disabling the redshift cluster's logging ([#22080](https://github.com/hashicorp/terraform-provider-aws/issues/22080))
* resource/aws_s3_bucket_replication_configuration: Fix `MalformedXML` errors for replication rules using XML schema V1 ([#22026](https://github.com/hashicorp/terraform-provider-aws/issues/22026))
* resource/aws_vpc_endpoint: Fix erroneous diffs in `policy` when no changes made or policies are equivalent ([#22137](https://github.com/hashicorp/terraform-provider-aws/issues/22137))

## 3.68.0 (December 02, 2021)

FEATURES:

* **New Data Source:** `aws_codecommit_approval_rule_template` ([#11487](https://github.com/hashicorp/terraform-provider-aws/issues/11487))
* **New Data Source:** `aws_vpc_pool_data_source` ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))
* **New Resource:** `aws_codecommit_approval_rule_template` ([#11487](https://github.com/hashicorp/terraform-provider-aws/issues/11487))
* **New Resource:** `aws_vpc_ipam` ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))
* **New Resource:** `aws_vpc_ipam_pool` ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))
* **New Resource:** `aws_vpc_ipam_scope` ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))
* **New Resource:** `aws_vpc_ipam_pool_cidr` ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))
* **New Resource:** `aws_vpc_ipam_pool_cidr_allocation` ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))
* **New Resource:** `aws_vpc_ipv6_cidr_block_association` ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))

ENHANCEMENTS:

* data-source/aws_autoscaling_groups: Add support for tag filters ([#21966](https://github.com/hashicorp/terraform-provider-aws/issues/21966))
* resource/aws_account_alternate_contact: Add `account_id` argument ([#21888](https://github.com/hashicorp/terraform-provider-aws/issues/21888))
* resource/aws_lb_target_group: Add support for `connection_termination` argument for NLBs ([#21130](https://github.com/hashicorp/terraform-provider-aws/issues/21130))
* resource/aws_synthetics_canary: Add `artifact_config` argument. ([#21963](https://github.com/hashicorp/terraform-provider-aws/issues/21963))
* resource/aws_synthetics_canary: Make `artifact_s3_location` updateable. ([#21963](https://github.com/hashicorp/terraform-provider-aws/issues/21963))
* resource/aws_vpc: `cidr_block` value can now either be set explicitly or computed computed via AWS IPAM ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))
* resource/aws_vpc_ipv4_cidr_block_association: `cidr_block` value can now either be set explicitly or computed via AWS IPAM ([#21998](https://github.com/hashicorp/terraform-provider-aws/issues/21998))

BUG FIXES:

* data-source/aws_cloudfront_distribution: Correct `hosted_zone_id` for AWS China regions ([#21943](https://github.com/hashicorp/terraform-provider-aws/issues/21943))
* resource/aws_cloudfront_distribution: Correct `hosted_zone_id` for AWS China regions ([#21943](https://github.com/hashicorp/terraform-provider-aws/issues/21943))
* resource/aws_kms_external_key: Fix order-related diffs in `policy` ([#21990](https://github.com/hashicorp/terraform-provider-aws/issues/21990))
* resource/aws_kms_key: Fix order-related diffs in `policy` ([#21969](https://github.com/hashicorp/terraform-provider-aws/issues/21969))
* resource/aws_kms_replica_external_key: Fix order-related diffs in `policy` ([#21990](https://github.com/hashicorp/terraform-provider-aws/issues/21990))
* resource/aws_kms_replica_key: Fix order-related diffs in `policy` ([#21990](https://github.com/hashicorp/terraform-provider-aws/issues/21990))
* resource/aws_s3_bucket: Fix order-related diffs in `policy` ([#21997](https://github.com/hashicorp/terraform-provider-aws/issues/21997))
* resource/aws_s3_bucket_policy: Fix order-related diffs in `policy` ([#21997](https://github.com/hashicorp/terraform-provider-aws/issues/21997))
* resource/aws_s3_bucket_replication_configuration: Mark `event_threshold` in `destination` `metrics` configuration block as `Optional` ([#21901](https://github.com/hashicorp/terraform-provider-aws/issues/21901))

## 3.67.0 (November 25, 2021)

FEATURES:

* **New Data Source:** `aws_ec2_instance_types` ([#21850](https://github.com/hashicorp/terraform-provider-aws/issues/21850))
* **New Data Source:** `aws_imagebuilder_image_recipes` ([#21814](https://github.com/hashicorp/terraform-provider-aws/issues/21814))
* **New Resource:** `aws_account_alternate_contact` ([#21789](https://github.com/hashicorp/terraform-provider-aws/issues/21789))
* **New Resource:** `aws_appstream_stack_fleet_association` ([#21484](https://github.com/hashicorp/terraform-provider-aws/issues/21484))
* **New Resource:** `aws_appstream_stack_user_association` ([#21485](https://github.com/hashicorp/terraform-provider-aws/issues/21485))
* **New Resource:** `aws_appstream_user` ([#21485](https://github.com/hashicorp/terraform-provider-aws/issues/21485))
* **New Resource:** `aws_fsx_ontap_storage_virtual_machine` ([#21780](https://github.com/hashicorp/terraform-provider-aws/issues/21780))
* **New Resource:** `aws_fsx_ontap_volume` ([#21889](https://github.com/hashicorp/terraform-provider-aws/issues/21889))

ENHANCEMENTS:

* data source/aws_lambda_function: Add `image_uri` attribute ([#21015](https://github.com/hashicorp/terraform-provider-aws/issues/21015))
* data-source/aws_elb: Add `desync_mitigation_mode` attribute ([#14764](https://github.com/hashicorp/terraform-provider-aws/issues/14764))
* data-source/aws_lb: Add `desync_mitigation_mode` attribute ([#14764](https://github.com/hashicorp/terraform-provider-aws/issues/14764))
* data-source/aws_lb: Add `enable_waf_fail_open` attribute ([#16393](https://github.com/hashicorp/terraform-provider-aws/issues/16393))
* resource/aws_athena_workgroup: Add `engine_version` argument ([#17733](https://github.com/hashicorp/terraform-provider-aws/issues/17733))
* resource/aws_cloudtrail: Add `exclude_management_event_sources` argument ([#17203](https://github.com/hashicorp/terraform-provider-aws/issues/17203))
* resource/aws_dlm_lifecycle_policy: Add `cross_region_copy_rule` argument in the `schedule` configuration block ([#12868](https://github.com/hashicorp/terraform-provider-aws/issues/12868))
* resource/aws_ec2_fleet: Support in-place update of Launch Template config ([#15387](https://github.com/hashicorp/terraform-provider-aws/issues/15387))
* resource/aws_ecs_service: Allow `capacity_provider_strategy` changes to be updated in place, when possible ([#20707](https://github.com/hashicorp/terraform-provider-aws/issues/20707))
* resource/aws_elasticache_replication_group: Allow `auth_token` argument to be rotated without destroy and create ([#16203](https://github.com/hashicorp/terraform-provider-aws/issues/16203))
* resource/aws_elb: Add `desync_mitigation_mode` argument ([#14764](https://github.com/hashicorp/terraform-provider-aws/issues/14764))
* resource/aws_lb: Add `desync_mitigation_mode` argument ([#14764](https://github.com/hashicorp/terraform-provider-aws/issues/14764))
* resource/aws_lb: Add `enable_waf_fail_open` argument ([#16393](https://github.com/hashicorp/terraform-provider-aws/issues/16393))
* resource/aws_lb: Update `name` and `name_prefix` plan-time validation to exclude `"internal-"` ([#10693](https://github.com/hashicorp/terraform-provider-aws/issues/10693))
* resource/aws_ssm_association: Add `s3_region` argument to `output_location` configuration block ([#21803](https://github.com/hashicorp/terraform-provider-aws/issues/21803))

BUG FIXES:

* data-source/aws_iam_policy_document: No longer show changes when there's a single condition ([#19533](https://github.com/hashicorp/terraform-provider-aws/issues/19533))
* resource/aws_apprunner_service: Make instance_role_arn optional ([#20149](https://github.com/hashicorp/terraform-provider-aws/issues/20149))
* resource/aws_autoscaling_group: Prevent infinite wait for capacity when increasing `min_size` and not specifying `desired_capacity` ([#12018](https://github.com/hashicorp/terraform-provider-aws/issues/12018))
* resource/aws_ecs_service: Mark `enable_ecs_managed_tags` as `ForceNew` ([#7983](https://github.com/hashicorp/terraform-provider-aws/issues/7983))
* resource/aws_imagebuilder_image_recipe: Enabled updates without failures due to `aws_imagebuilder_image_pipeline` dependencies. ([#21884](https://github.com/hashicorp/terraform-provider-aws/issues/21884))
* resource/aws_rds_cluster_instance: Fix error from unexpected state `storage-optimization` ([#21900](https://github.com/hashicorp/terraform-provider-aws/issues/21900))
* resource/aws_s3_bucket: Prevent `OperationAborted` conflict errors when simultaneously applying `aws_s3_bucket_policy`, `aws_s3_bucket_public_access_block` changes ([#12949](https://github.com/hashicorp/terraform-provider-aws/issues/12949))
* resource/aws_s3_bucket_policy: Prevent `OperationAborted` conflict errors when simultaneously applying `aws_s3_bucket`, `aws_s3_bucket_public_access_block` changes ([#12949](https://github.com/hashicorp/terraform-provider-aws/issues/12949))
* resource/aws_s3_bucket_public_access_block: Prevent `OperationAborted` conflict errors when simultaneously applying `aws_s3_bucket_policy`, `aws_s3_bucket` changes ([#12949](https://github.com/hashicorp/terraform-provider-aws/issues/12949))

## 3.66.0 (November 18, 2021)

FEATURES:

* **New Data Source:** `aws_emr_release_labels` ([#21767](https://github.com/hashicorp/terraform-provider-aws/issues/21767))
* **New Resource:** `aws_appstream_directory_config` ([#21505](https://github.com/hashicorp/terraform-provider-aws/issues/21505))
* **New Resource:** `aws_iot_thing_group` ([#21799](https://github.com/hashicorp/terraform-provider-aws/issues/21799))
* **New Resource:** `aws_iot_thing_group_membership` ([#21799](https://github.com/hashicorp/terraform-provider-aws/issues/21799))
* **New Resource:** `aws_lambda_layer_version_permission` ([#11941](https://github.com/hashicorp/terraform-provider-aws/issues/11941))
* **New Resource:** `aws_s3_bucket_replication_configuration` ([#20777](https://github.com/hashicorp/terraform-provider-aws/issues/20777))
* **New Resource:** `aws_s3control_access_point_policy` ([#19294](https://github.com/hashicorp/terraform-provider-aws/issues/19294))
* **New Resource:** `aws_s3control_multi_region_access_point` ([#21060](https://github.com/hashicorp/terraform-provider-aws/issues/21060))
* **New Resource:** `aws_s3control_multi_region_access_point_policy` ([#21060](https://github.com/hashicorp/terraform-provider-aws/issues/21060))
* **New Resource:** `aws_s3control_object_lambda_access_point` ([#19294](https://github.com/hashicorp/terraform-provider-aws/issues/19294))
* **New Resource:** `aws_s3control_object_lambda_access_point_policy` ([#19294](https://github.com/hashicorp/terraform-provider-aws/issues/19294))
* **New Resource:** `aws_securityhub_finding_aggregator` ([#21560](https://github.com/hashicorp/terraform-provider-aws/issues/21560))

ENHANCEMENTS:

* aws_s3_access_point: Add `alias` attribute ([#19294](https://github.com/hashicorp/terraform-provider-aws/issues/19294))
* aws_s3_access_point: Add `endpoints` attribute ([#19294](https://github.com/hashicorp/terraform-provider-aws/issues/19294))
* data-source/aws_ec2_instance_type: Add `encryption_in_transit_supported` attribute ([#21837](https://github.com/hashicorp/terraform-provider-aws/issues/21837))
* resource/aws_emr_cluster: Add `auto_termination_policy` argument ([#21702](https://github.com/hashicorp/terraform-provider-aws/issues/21702))
* resource/aws_iot_thing_type: Add `tags` argument and `tags_all` attribute to support resource tagging ([#21769](https://github.com/hashicorp/terraform-provider-aws/issues/21769))
* resource/aws_kinesis_firehose_delivery_stream: Add `dynamic_partitioning_configuration` configuration block ([#20769](https://github.com/hashicorp/terraform-provider-aws/issues/20769))
* resource/aws_lambda_layer_version: Add `skip_destroy` argument ([#11997](https://github.com/hashicorp/terraform-provider-aws/issues/11997))
* resource/aws_neptune_cluster: Support in-place update of `engine_version` ([#21760](https://github.com/hashicorp/terraform-provider-aws/issues/21760))
* resource/aws_route53_resolver_dnssec_config: Increase resource creation and deletion timeouts to 10 minutes ([#21797](https://github.com/hashicorp/terraform-provider-aws/issues/21797))
* resource/aws_sagemaker_endpoint: Add `deployment_config` argument ([#21765](https://github.com/hashicorp/terraform-provider-aws/issues/21765))

BUG FIXES:

* aws_s3_access_point: `vpc_configuration.vpc_id` is _ForceNew_ ([#19294](https://github.com/hashicorp/terraform-provider-aws/issues/19294))
* data-source/aws_cloudfront_response_headers_policy: Correctly set `custom_headers_config` attribute ([#21838](https://github.com/hashicorp/terraform-provider-aws/issues/21838))
* resource/aws_autoscaling_group: Fix pending state in instance refresh ([#21777](https://github.com/hashicorp/terraform-provider-aws/issues/21777))
* resource/aws_cloudfront_cache_policy: Fix 0 values for `default_ttl`, `max_ttl` and `min_ttl` arguments ([#21793](https://github.com/hashicorp/terraform-provider-aws/issues/21793))
* resource/aws_internet_gateway: Allow `available` as a *pending* state during gateway detach ([#21794](https://github.com/hashicorp/terraform-provider-aws/issues/21794))
* resource/aws_lambda_layer_version: Increase MaxItems for compatible_runtimes field to 15. ([#21825](https://github.com/hashicorp/terraform-provider-aws/issues/21825))
* resource/aws_route: On route creation with high custom creation timeout configured, the aws_route resource does no longer give up before the create timeout is exceeded (previously it was giving up after 20 not found checks). ([#21831](https://github.com/hashicorp/terraform-provider-aws/issues/21831))
* resource/aws_security_group: Fix lack of pagination when describing security groups ([#21743](https://github.com/hashicorp/terraform-provider-aws/issues/21743))

## 3.65.0 (November 11, 2021)

FEATURES:

* **New Data Source:** `aws_key_pair` ([#15829](https://github.com/hashicorp/terraform-provider-aws/issues/15829))
* **New Resource:** `aws_cloudfront_field_level_encryption_config` ([#15033](https://github.com/hashicorp/terraform-provider-aws/issues/15033))
* **New Resource:** `aws_cloudfront_field_level_encryption_profile` ([#12509](https://github.com/hashicorp/terraform-provider-aws/issues/12509))
* **New Resource:** `aws_docdb_global_cluster` ([#20978](https://github.com/hashicorp/terraform-provider-aws/issues/20978))
* **New Resource:** `aws_s3_bucket_intelligent_tiering_configuration` ([#20329](https://github.com/hashicorp/terraform-provider-aws/issues/20329))

ENHANCEMENTS:

* resource/aws_batch_job_queue: Remove limit of 3 items from the `compute_environments` argument ([#21737](https://github.com/hashicorp/terraform-provider-aws/issues/21737))
* resource/aws_cloudfront_function: Add `live_etag_version` attribute ([#19697](https://github.com/hashicorp/terraform-provider-aws/issues/19697))
* resource/aws_datasync_s3_location: Add validation to `agent_arns`, `s3_bucket_arn` and `s3_config.bucket_access_role_arn` arguments ([#21661](https://github.com/hashicorp/terraform-provider-aws/issues/21661))
* resource/aws_docdb_cluster: Add `global_cluster_identifier` argument ([#20978](https://github.com/hashicorp/terraform-provider-aws/issues/20978))
* resource/aws_ebs_encryption_by_default: Add import support ([#21717](https://github.com/hashicorp/terraform-provider-aws/issues/21717))

BUG FIXES:

* data-source/aws_network_interface: Correctly set `attachment` attribute ([#21542](https://github.com/hashicorp/terraform-provider-aws/issues/21542))
* data-source/aws_route: Fix lack of pagination when describing route tables ([#21710](https://github.com/hashicorp/terraform-provider-aws/issues/21710))
* resource/aws_cloudfront_cache_policy: Fix assorted crashes ([#12509](https://github.com/hashicorp/terraform-provider-aws/issues/12509))
* resource/aws_cloudfront_cache_policy: The `parameters_in_cache_key_and_forwarded_to_origin` argument is required ([#12509](https://github.com/hashicorp/terraform-provider-aws/issues/12509))
* resource/aws_cloudfront_function: The `etag` attribute always the `DEVELOPMENT` version's value ([#19697](https://github.com/hashicorp/terraform-provider-aws/issues/19697))
* resource/aws_cloudfront_origin_request_policy: Fix assorted crashes ([#12509](https://github.com/hashicorp/terraform-provider-aws/issues/12509))
* resource/aws_default_route_table: Fix lack of pagination when describing route tables ([#21710](https://github.com/hashicorp/terraform-provider-aws/issues/21710))
* resource/aws_eks_node_group: Respect order of configured `instance_types` ([#21404](https://github.com/hashicorp/terraform-provider-aws/issues/21404))
* resource/aws_elasticsearch_domain: Fix tagging on creation ([#21738](https://github.com/hashicorp/terraform-provider-aws/issues/21738))
* resource/aws_internet_gateway: Retry resource read after creation to deal with EC2 API eventual consistency ([#21542](https://github.com/hashicorp/terraform-provider-aws/issues/21542))
* resource/aws_route_table: Fix lack of pagination when describing route tables ([#21710](https://github.com/hashicorp/terraform-provider-aws/issues/21710))
* resource/aws_route_table_association: Fix lack of pagination when describing route tables ([#21710](https://github.com/hashicorp/terraform-provider-aws/issues/21710))
* resource/aws_security_group_rule: Fix resource import for rules with `icmp` or `icmpv6` protocol ([#21163](https://github.com/hashicorp/terraform-provider-aws/issues/21163))
* resource/aws_servicecatalog_provisioned_product: Allow empty values in provisioning parameters ([#21669](https://github.com/hashicorp/terraform-provider-aws/issues/21669))

## 3.64.2 (November 05, 2021)

BUG FIXES:

* provider: Additional fixes to allow setting endpoints with non-standard, legacy keys. ([#21657](https://github.com/hashicorp/terraform-provider-aws/issues/21657))

## 3.64.1 (November 05, 2021)

BUG FIXES:

* provider: Fix bug preventing custom endpoints from being set ([#21639](https://github.com/hashicorp/terraform-provider-aws/issues/21639))
* provider: Fix bug preventing proper assignment of custom endpoints ([#21641](https://github.com/hashicorp/terraform-provider-aws/issues/21641))

## 3.64.0 (November 04, 2021)

FEATURES:

* **New Data Source:** `aws_cloudfront_response_headers_policy` ([#21620](https://github.com/hashicorp/terraform-provider-aws/issues/21620))
* **New Data Source:** `aws_iam_user_ssh_key` ([#21335](https://github.com/hashicorp/terraform-provider-aws/issues/21335))
* **New Resource:** `aws_backup_vault_lock_configuration` ([#21315](https://github.com/hashicorp/terraform-provider-aws/issues/21315))
* **New Resource:** `aws_cloudfront_response_headers_policy` ([#21620](https://github.com/hashicorp/terraform-provider-aws/issues/21620))
* **New Resource:** `aws_kms_replica_external_key` ([#20533](https://github.com/hashicorp/terraform-provider-aws/issues/20533))
* **New Resource:** `aws_kms_replica_key` ([#20533](https://github.com/hashicorp/terraform-provider-aws/issues/20533))
* **New Resource:** `aws_prometheus_alert_manager_definition` ([#21431](https://github.com/hashicorp/terraform-provider-aws/issues/21431))
* **New Resource:** `aws_prometheus_rule_group_namespace` ([#21470](https://github.com/hashicorp/terraform-provider-aws/issues/21470))

ENHANCEMENTS:

* data-source/aws_kms_key: Add `multi_region` and `multi_region_configuration` attributes ([#20533](https://github.com/hashicorp/terraform-provider-aws/issues/20533))
* data-source/aws_launch_template: Add `network_card_index` attribute to `network_interfaces` configuration block ([#21555](https://github.com/hashicorp/terraform-provider-aws/issues/21555))
* data-source/aws_network_interface: Add `arn` attribute ([#21265](https://github.com/hashicorp/terraform-provider-aws/issues/21265))
* data-source/aws_s3_bucket: Return `hosted_zone_id` attribute for `cn-northwest-1` (Ningxia) region ([#21337](https://github.com/hashicorp/terraform-provider-aws/issues/21337))
* resource/aws_apigateway_usage_plan: Add `throttle` argument for `api_stages` block. ([#21461](https://github.com/hashicorp/terraform-provider-aws/issues/21461))
* resource/aws_batch_compute_environment: Add `ec2_configuration` argument to `compute_resources` configuration block ([#21565](https://github.com/hashicorp/terraform-provider-aws/issues/21565))
* resource/aws_cloudfront_distribution: Add `response_headers_policy_id` argument to `default_cache_behavior` configuration block ([#21620](https://github.com/hashicorp/terraform-provider-aws/issues/21620))
* resource/aws_cloudfront_distribution: Add `response_headers_policy_id` argument to `ordered_cache_behavior` configuration block ([#21620](https://github.com/hashicorp/terraform-provider-aws/issues/21620))
* resource/aws_dms_endpoint: Add `include_transaction_details`, `include_partition_value`, `partition_include_schema_table`, `include_table_alter_operations`, `include_control_details` and `include_null_and_empty` arguments to `kinesis_settings` configuration block ([#20084](https://github.com/hashicorp/terraform-provider-aws/issues/20084))
* resource/aws_eks_node_group: Support for `BOTTLEROCKET_ARM_64` and `BOTTLEROCKET_x86_64` `ami_type` argument values ([#21616](https://github.com/hashicorp/terraform-provider-aws/issues/21616))
* resource/aws_glue_crawler: Add `dlq_event_queue_arn` and `event_queue_arn` arguments to the `s3_target` configuration block ([#21467](https://github.com/hashicorp/terraform-provider-aws/issues/21467))
* resource/aws_glue_data_catalog_encryption_settings: Disable encryption on resource deletion ([#21452](https://github.com/hashicorp/terraform-provider-aws/issues/21452))
* resource/aws_kinesisanalyticsv2_application: `runtime_environment` now supports `FLINK-1_13` ([#21341](https://github.com/hashicorp/terraform-provider-aws/issues/21341))
* resource/aws_kms_external_key: Add `multi_region` argument ([#20533](https://github.com/hashicorp/terraform-provider-aws/issues/20533))
* resource/aws_kms_key: Add `multi_region` argument ([#20533](https://github.com/hashicorp/terraform-provider-aws/issues/20533))
* resource/aws_launch_template: Add `network_card_index` argument to `network_interfaces` configuration block ([#21555](https://github.com/hashicorp/terraform-provider-aws/issues/21555))
* resource/aws_network_interface: Add `arn` and `owner_id` attributes ([#21265](https://github.com/hashicorp/terraform-provider-aws/issues/21265))
* resource/aws_network_interface: Add `ipv4_prefix`, `ipv4_prefix_count`, `ipv6_prefix` and `ipv6_prefix_count` arguments ([#21265](https://github.com/hashicorp/terraform-provider-aws/issues/21265))
* resource/aws_route53_key_signing_key: Deactivate key-signing key with `ACTION_NEEDED` status before deletion ([#21369](https://github.com/hashicorp/terraform-provider-aws/issues/21369))
* resource/aws_s3_bucket: Add `metrics` and `replication_time` arguments to `replication_configuration.rules` configuration block to support Amazon S3 Replication Time Control ([#21176](https://github.com/hashicorp/terraform-provider-aws/issues/21176))
* resource/aws_s3_bucket: Return `hosted_zone_id` attribute for `cn-northwest-1` (Ningxia) region ([#21337](https://github.com/hashicorp/terraform-provider-aws/issues/21337))
* resource/aws_storage_gateway_nfs_file_share: Add `audit_destination_arn` argument. ([#21482](https://github.com/hashicorp/terraform-provider-aws/issues/21482))

BUG FIXES:

* aws/resource_aws_lex_slot_type: Correctly determine `version` attribute ([#21509](https://github.com/hashicorp/terraform-provider-aws/issues/21509))
* resource/aws_cloudwatch_metric_alarm: Fix imported 'treat_missing_data' diff ([#21363](https://github.com/hashicorp/terraform-provider-aws/issues/21363))
* resource/aws_codedeploy_deployment_group: Correctly update `deployment_group_name` argument ([#21362](https://github.com/hashicorp/terraform-provider-aws/issues/21362))
* resource/aws_db_event_subscription: Fix adding new `event_categories` to existing resource ([#21338](https://github.com/hashicorp/terraform-provider-aws/issues/21338))
* resource/aws_flow_log: parameters of destination_options block now properly force resource rebuild ([#21434](https://github.com/hashicorp/terraform-provider-aws/issues/21434))
* resource/aws_kinesisanalyticsv2_application: Correctly update `run_configuration` argument ([#21303](https://github.com/hashicorp/terraform-provider-aws/issues/21303))
* resource/aws_placement_group: `partition_count` argument is Computed, preventing spurious resource diffs ([#21555](https://github.com/hashicorp/terraform-provider-aws/issues/21555))

## 3.63.0 (October 14, 2021)

FEATURES:

* **New Resource:** `aws_chime_voice_connector_termination_credentials` ([#21162](https://github.com/hashicorp/terraform-provider-aws/issues/21162))
* **New Resource:** `aws_glue_partition_index` ([#21234](https://github.com/hashicorp/terraform-provider-aws/issues/21234))
* **New Resource:** `aws_sagemaker_model_package_group_policy` ([#21250](https://github.com/hashicorp/terraform-provider-aws/issues/21250))

ENHANCEMENTS:

* data-source/aws_instance: Add `placement_partition_number` attribute ([#7777](https://github.com/hashicorp/terraform-provider-aws/issues/7777))
* data-source/glue_connection: Add tagging support. ([#21226](https://github.com/hashicorp/terraform-provider-aws/issues/21226))
* resource/aws_flow_log: Add `destination_options` argument to support Apache Parquet, Hive-compatible prefixes and hourly partitioned files ([#21285](https://github.com/hashicorp/terraform-provider-aws/issues/21285))
* resource/aws_glue_resource_policy: Add `enable_hybrid` argument. ([#21239](https://github.com/hashicorp/terraform-provider-aws/issues/21239))
* resource/aws_instance: Add `placement_partition_number` argument ([#7777](https://github.com/hashicorp/terraform-provider-aws/issues/7777))
* resource/aws_placement_group: Add `partition_count` argument ([#15360](https://github.com/hashicorp/terraform-provider-aws/issues/15360))
* resource/aws_rds_cluster: Add `db_instance_parameter_group_name` attribute to allow major version upgrade using custom parameter groups ([#17111](https://github.com/hashicorp/terraform-provider-aws/issues/17111))
* resource/aws_rds_cluster: Add `enable_global_write_forwarding` attribute ([#17111](https://github.com/hashicorp/terraform-provider-aws/issues/17111))
* resource/glue_connection: Add tagging support. ([#21226](https://github.com/hashicorp/terraform-provider-aws/issues/21226))
* resource/rds_cluster_instance: Add `performance_insights_retention_period` attribute ([#17111](https://github.com/hashicorp/terraform-provider-aws/issues/17111))

BUG FIXES:

* resource/aws_glue_catalog_table: change `partition_index.keys` to list instead of set ([#21234](https://github.com/hashicorp/terraform-provider-aws/issues/21234))
* resource/aws_imagebuilder_distribution_configuration: remove hard limit on distribution target accounts ([#21254](https://github.com/hashicorp/terraform-provider-aws/issues/21254))
* resource/aws_rds_cluster: Add possible pending states for cluster update ([#17111](https://github.com/hashicorp/terraform-provider-aws/issues/17111))
* resource/aws_rds_cluster_instance: Remove force new resource on the `engine_version` parameter to allow upgrade without remove instances ([#17111](https://github.com/hashicorp/terraform-provider-aws/issues/17111))
* resource/glue_catalog_table: Ignore not exists errors on delete ([#21227](https://github.com/hashicorp/terraform-provider-aws/issues/21227))

## 3.62.0 (October 08, 2021)

FEATURES:

* **New Resource:** `aws_dx_connection_confirmation` ([#16489](https://github.com/hashicorp/terraform-provider-aws/issues/16489))
* **New Resource:** `aws_dx_hosted_connection` ([#16489](https://github.com/hashicorp/terraform-provider-aws/issues/16489))

ENHANCEMENTS:

* resource/aws_cloudformation_stack_set_instance: Add `deployment_targets` `organizational_unit_ids` argument ([#21193](https://github.com/hashicorp/terraform-provider-aws/issues/21193))
* resource/aws_db_instance: Add `replica_mode` argument ([#17991](https://github.com/hashicorp/terraform-provider-aws/issues/17991))
* resource/aws_default_route_table: Add [custom `timeouts`](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts) block ([#21161](https://github.com/hashicorp/terraform-provider-aws/issues/21161))
* resource/aws_dms_endpoint: Add `message_format`, `include_transaction_details`, `include_partition_value`, `partition_include_schema_table`, `include_table_alter_operations`, `include_control_details`, `message_max_bytes`, `include_null_and_empty`, `security_protocol`, `ssl_client_certificate_arn`, `ssl_client_key_arn`, `ssl_client_key_password`, `ssl_ca_certificate_arn`, `sasl_username`, `sasl_password` and `no_hex_prefix` arguments to `kafka_settings` configuration block ([#20904](https://github.com/hashicorp/terraform-provider-aws/issues/20904))
* resource/aws_dms_endpoint: Add plan time validation for `mongodb_settings.auth_type`, `mongodb_settings.auth_mechanism`, `mongodb_settings.nesting_level` and `s3_settings.compression_type` arguments ([#21174](https://github.com/hashicorp/terraform-provider-aws/issues/21174))
* resource/aws_dms_endpoint: Added missing `engine_name` values for sources and/or targets ([#21174](https://github.com/hashicorp/terraform-provider-aws/issues/21174))
* resource/aws_dms_replication_task: Add `cdc_start_position` argument ([#21201](https://github.com/hashicorp/terraform-provider-aws/issues/21201))
* resource/aws_dx_lag: Add `connection_id` argument ([#16489](https://github.com/hashicorp/terraform-provider-aws/issues/16489))
* resource/aws_emr_cluster: Add `log_encryption_kms_key_id` argument ([#17706](https://github.com/hashicorp/terraform-provider-aws/issues/17706))
* resource/aws_lex_bot: Added waiter support to account for `BUILDING` status ([#21122](https://github.com/hashicorp/terraform-provider-aws/issues/21122))
* resource/aws_route_table: Add [custom `timeouts`](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts) block ([#21161](https://github.com/hashicorp/terraform-provider-aws/issues/21161))
* resource/aws_volume_attachment: Add `stop_instance_before_detaching` argument ([#21144](https://github.com/hashicorp/terraform-provider-aws/issues/21144))
* resource/aws_vpn_gateway_route_propagation: Add [custom `timeouts`](https://www.terraform.io/docs/language/resources/syntax.html#operation-timeouts) block ([#21161](https://github.com/hashicorp/terraform-provider-aws/issues/21161))

BUG FIXES:

* aws/resource_aws_lex_bot: Correctly determine `version` attribute ([#20383](https://github.com/hashicorp/terraform-provider-aws/issues/20383))
* aws/resource_aws_lex_intent: Correctly determine `version` attribute ([#20383](https://github.com/hashicorp/terraform-provider-aws/issues/20383))
* resource/aws_appstream_fleet: More error validation in waiter ([#21125](https://github.com/hashicorp/terraform-provider-aws/issues/21125))
* resource/aws_appstream_stack:  More error validation in waiter ([#21125](https://github.com/hashicorp/terraform-provider-aws/issues/21125))
* resource/aws_autoscalingplans_scaling_plan: Fix updates to `scaling_instruction` argument ([#17987](https://github.com/hashicorp/terraform-provider-aws/issues/17987))
* resource/aws_elasticache_replication_group: Properly updates tags on Replication Group member clusters when scaling up ([#21185](https://github.com/hashicorp/terraform-provider-aws/issues/21185))
* resource/aws_elasticache_replication_group: Properly updates tags on the Replication Group in addition to the member clusters ([#21185](https://github.com/hashicorp/terraform-provider-aws/issues/21185))
* resource/aws_lb_target_group: Handle attributes at creation: `deregistration_delay`, `load_balancing_algorithm_type`, `preserve_client_ip`, `proxy_protocol_v2`, `slow_start`, `stickiness`, and `lambda_multi_value_headers_enabled` ([#21187](https://github.com/hashicorp/terraform-provider-aws/issues/21187))
* resource/aws_route: Use custom `timeouts` values ([#21161](https://github.com/hashicorp/terraform-provider-aws/issues/21161))
* resource/aws_ses_configuration_set: Fix ARN ([#21188](https://github.com/hashicorp/terraform-provider-aws/issues/21188))

## 3.61.0 (October 01, 2021)

FEATURES:

* **New Data Source:** `aws_cloudcontrolapi_resource` ([#21110](https://github.com/hashicorp/terraform-provider-aws/issues/21110))
* **New Data Source:** `aws_db_proxy` ([#21053](https://github.com/hashicorp/terraform-provider-aws/issues/21053))
* **New Data Source:** `aws_ec2_host` ([#10817](https://github.com/hashicorp/terraform-provider-aws/issues/10817))
* **New Data Source:** `aws_kinesis_firehose_delivery_stream` ([#18445](https://github.com/hashicorp/terraform-provider-aws/issues/18445))
* **New Data Source:** `aws_ssm_parameters_by_path` ([#9615](https://github.com/hashicorp/terraform-provider-aws/issues/9615))
* **New Resource:** `aws_appstream_image_builder` ([#21036](https://github.com/hashicorp/terraform-provider-aws/issues/21036))
* **New Resource:** `aws_cloudcontrolapi_resource` ([#21110](https://github.com/hashicorp/terraform-provider-aws/issues/21110))
* **New Resource:** `aws_ec2_host` ([#10817](https://github.com/hashicorp/terraform-provider-aws/issues/10817))
* **New Resource:** `aws_iot_authorizer` ([#14671](https://github.com/hashicorp/terraform-provider-aws/issues/14671))
* **New Resource:** `aws_quicksight_data_source` ([#20710](https://github.com/hashicorp/terraform-provider-aws/issues/20710))
* **New Resource:** `aws_redshift_scheduled_action` ([#13474](https://github.com/hashicorp/terraform-provider-aws/issues/13474))
* **New Resource:** `aws_sagemaker_studio_lifecycle_config` ([#21041](https://github.com/hashicorp/terraform-provider-aws/issues/21041))

ENHANCEMENTS:

* data-source/aws_lambda_function: Add support for Graviton2 with `architectures` attribute ([#21091](https://github.com/hashicorp/terraform-provider-aws/issues/21091))
* data-source/aws_lambda_layer_version: Add support for Graviton2 with `compatible_architectures` attribute ([#21091](https://github.com/hashicorp/terraform-provider-aws/issues/21091))
* provider: Add parameter `http_proxy` to provider configuration ([#21077](https://github.com/hashicorp/terraform-provider-aws/issues/21077))
* resource/aws_lb_target_group: Support `alb` value for `target_type` argument ([#21069](https://github.com/hashicorp/terraform-provider-aws/issues/21069))
* resource/aws_lambda_function: Add support for Graviton2 with `architectures` argument ([#21091](https://github.com/hashicorp/terraform-provider-aws/issues/21091))
* resource/aws_lambda_layer_version: Add support for Graviton2 with `compatible_architectures` argument ([#21091](https://github.com/hashicorp/terraform-provider-aws/issues/21091))
* resource/aws_sagemaker_app_image_config: Add tagging support. ([#21037](https://github.com/hashicorp/terraform-provider-aws/issues/21037))
* resource/aws_sagemaker_domain: Add `default_user_settings.jupyter_server_app_settings.lifecycle_config_arns` and `default_user_settings.kernel_gateway_app_settings.lifecycle_config_arns` arguments ([#21041](https://github.com/hashicorp/terraform-provider-aws/issues/21041))
* resource/aws_user_profile: Add `user_settings.jupyter_server_app_settings.lifecycle_config_arns` and `user_settings.kernel_gateway_app_settings.lifecycle_config_arns` arguments ([#21041](https://github.com/hashicorp/terraform-provider-aws/issues/21041))

BUG FIXES:

* resource/aws_dx_connection: Mark `provider_name` as Computed to avoid resource recreation with pre-v3.56.0 configurations ([#21085](https://github.com/hashicorp/terraform-provider-aws/issues/21085))
* resource/aws_dx_lag: Mark `provider_name` as Computed to avoid resource recreation with pre-v3.56.0 configurations ([#21085](https://github.com/hashicorp/terraform-provider-aws/issues/21085))
* resource/aws_route_table_association: Wait for up to 40 not found checks when creating a new route table association ([#21062](https://github.com/hashicorp/terraform-provider-aws/issues/21062))

## 3.60.0 (September 23, 2021)

FEATURES:

* **New Data Source:** `aws_cloudfront_log_delivery_canonical_user_id` ([#15167](https://github.com/hashicorp/terraform-provider-aws/issues/15167))
* **New Data Source:** `aws_cloudwatch_log_groups` ([#17151](https://github.com/hashicorp/terraform-provider-aws/issues/17151))
* **New Data Source:** `aws_connect_contact_flow` ([#16854](https://github.com/hashicorp/terraform-provider-aws/issues/16854))
* **New Data Source:** `aws_connect_instance` ([#16709](https://github.com/hashicorp/terraform-provider-aws/issues/16709))
* **New Data Source:** `aws_iam_users` ([#20877](https://github.com/hashicorp/terraform-provider-aws/issues/20877))
* **New Data Source:** `aws_msk_broker_nodes` ([#20615](https://github.com/hashicorp/terraform-provider-aws/issues/20615))
* **New Data Source:** `aws_msk_kafka_version` ([#20638](https://github.com/hashicorp/terraform-provider-aws/issues/20638))
* **New Resource:** `aws_appstream_fleet` ([#20543](https://github.com/hashicorp/terraform-provider-aws/issues/20543))
* **New Resource:** `aws_chime_voice_connector_streaming` ([#20933](https://github.com/hashicorp/terraform-provider-aws/issues/20933))
* **New Resource:** `aws_connect_contact_flow` ([#16854](https://github.com/hashicorp/terraform-provider-aws/issues/16854))
* **New Resource:** `aws_connect_instance` ([#16709](https://github.com/hashicorp/terraform-provider-aws/issues/16709))
* **New Resource:** `aws_ec2_managed_prefix_list_entry` ([#19394](https://github.com/hashicorp/terraform-provider-aws/issues/19394))
* **New Resource:** `aws_fsx_ontap_filesystem` ([#20951](https://github.com/hashicorp/terraform-provider-aws/issues/20951))
* **New Resource:** `aws_sagemaker_flow_definition` ([#20825](https://github.com/hashicorp/terraform-provider-aws/issues/20825))

ENHANCEMENTS:

* data-source/efs_file_system: Add `transition_to_primary_storage_class` to `lifecycle_policy`. ([#20971](https://github.com/hashicorp/terraform-provider-aws/issues/20971))
* resource/aws_msk_cluster: Add `zookeeper_connect_string_tls` attribute ([#15661](https://github.com/hashicorp/terraform-provider-aws/issues/15661))
* resource/aws_msk_cluster: Configurable Create, Update and Delete timeouts ([#17726](https://github.com/hashicorp/terraform-provider-aws/issues/17726))

BUG FIXES:

* data-source/aws_launch_template: Fix `error setting metadata_options` ([#21008](https://github.com/hashicorp/terraform-provider-aws/issues/21008))
* resource/aws_cognito_user_pool: Fix removal of `lambda_config` ([#20952](https://github.com/hashicorp/terraform-provider-aws/issues/20952))
* resource/aws_msk_cluster: Don't recreate cluster if order of `broker_node_group_info.client_subnets` or `broker_node_group_info.security_groups` entries change ([#14627](https://github.com/hashicorp/terraform-provider-aws/issues/14627))
* resource/efs_file_system: Allow multiple lifecycle policies. ([#20971](https://github.com/hashicorp/terraform-provider-aws/issues/20971))

## 3.59.0 (September 16, 2021)

FEATURES:

* **New Data Source:** `aws_eks_clusters` ([#20315](https://github.com/hashicorp/terraform-provider-aws/issues/20315))
* **New Data Source:** `aws_eks_node_group` ([#13564](https://github.com/hashicorp/terraform-provider-aws/issues/13564))
* **New Data Source:** `aws_eks_node_groups` ([#13564](https://github.com/hashicorp/terraform-provider-aws/issues/13564))
* **New Resource:** `aws_chime_voice_connector_logging` ([#20863](https://github.com/hashicorp/terraform-provider-aws/issues/20863))
* **New Resource:** `aws_transfer_access` ([#20342](https://github.com/hashicorp/terraform-provider-aws/issues/20342))

ENHANCEMENTS:

* resource/aws_cloudtrail: Add `advanced_event_selector` argument ([#19368](https://github.com/hashicorp/terraform-provider-aws/issues/19368))
* resource/aws_config_delivery_channel: Add `s3_kms_key_arn` argument ([#20600](https://github.com/hashicorp/terraform-provider-aws/issues/20600))
* resource/aws_ec2_client_vpn_endpoint: Add `self_service_portal` and `authentication_options.self_service_saml_provider_arn` arguments to support self-service portal ([#17897](https://github.com/hashicorp/terraform-provider-aws/issues/17897))
* resource/aws_ec2_managed_prefix_list: allow updating `max_entries`. ([#20797](https://github.com/hashicorp/terraform-provider-aws/issues/20797))
* resource/aws_efs_file_system: Add `lifecycle_policy.transition_to_primary_storage_class` argument to support Intelligent-Tiering ([#20874](https://github.com/hashicorp/terraform-provider-aws/issues/20874))
* resource/aws_efs_file_system_policy: Add `bypass_policy_lockout_safety_check` argument ([#20838](https://github.com/hashicorp/terraform-provider-aws/issues/20838))
* resource/aws_iam_role: Add plan time validation for `path`, `permissions_boundary`, `managed_policy_arns`. ([#19532](https://github.com/hashicorp/terraform-provider-aws/issues/19532))
* resource/aws_iam_role: Retry `assume_role_policy` updates for IAM eventual consistency ([#12436](https://github.com/hashicorp/terraform-provider-aws/issues/12436))
* resource/aws_iam_role: `name_prefix` is now Computed ([#20785](https://github.com/hashicorp/terraform-provider-aws/issues/20785))
* resource/aws_launch_template: add plan time validation to `spot_options.block_duration_minutes` ([#20796](https://github.com/hashicorp/terraform-provider-aws/issues/20796))
* resource/aws_launch_template: add support for `http_protocol_ipv6` to `metadata_options`. ([#20796](https://github.com/hashicorp/terraform-provider-aws/issues/20796))
* resource/aws_mwaa_environment: Increase resource creation timeout to 2 hours ([#20861](https://github.com/hashicorp/terraform-provider-aws/issues/20861))
* resource/aws_route53_health_check: Add plan time validation for `regions` ([#20795](https://github.com/hashicorp/terraform-provider-aws/issues/20795))
* resource/aws_sagemaker_endpoint_configuration: Add `async_inference_config` argument ([#20809](https://github.com/hashicorp/terraform-provider-aws/issues/20809))
* resource/aws_transfer_server: Add `directory_id` argument to support Microsoft Active Directory (AD) authentication ([#20342](https://github.com/hashicorp/terraform-provider-aws/issues/20342))

BUG FIXES:

* resource/aws_cognito_user_pool: Fix continual diff on `email_configuration.configuration_set` ([#20791](https://github.com/hashicorp/terraform-provider-aws/issues/20791))
* resource/aws_db_instance: Fix updating `license_model`. ([#20779](https://github.com/hashicorp/terraform-provider-aws/issues/20779))
* resource/aws_iam_role: Change `name_prefix` validation to a range of 1 to 38 characters ([#20785](https://github.com/hashicorp/terraform-provider-aws/issues/20785))
* resource/aws_imagebuilder_distribution_configuration: Improve validation error message of `name` argument ([#20842](https://github.com/hashicorp/terraform-provider-aws/issues/20842))
* resource/aws_kms_key: Extends timeouts for policy and tag propagation to 5 minutes each ([#20914](https://github.com/hashicorp/terraform-provider-aws/issues/20914))
* resource/aws_route53_health_check: Fix update for `ip_address` ([#20795](https://github.com/hashicorp/terraform-provider-aws/issues/20795))

## 3.58.0 (September 09, 2021)

FEATURES:

* **New Resource:** `aws_chime_voice_connector_origination` ([#20676](https://github.com/hashicorp/terraform-provider-aws/issues/20676))
* **New Resource:** `aws_chime_voice_connector_termination` ([#20667](https://github.com/hashicorp/terraform-provider-aws/issues/20667))
* **New Resource:** `aws_quicksight_group_membership` ([#20687](https://github.com/hashicorp/terraform-provider-aws/issues/20687))

## 3.57.0 (September 02, 2021)

FEATURES:

* **New Resource:** `aws_service_discovery_instance` ([#17498](https://github.com/hashicorp/terraform-provider-aws/issues/17498))

ENHANCEMENTS:

* data-source/aws_instance: Add `ipv6_addresses` attribute ([#17859](https://github.com/hashicorp/terraform-provider-aws/issues/17859))
* resource/aws_athena_database: Read the database name from the `AwsDataCatalog` ([#19765](https://github.com/hashicorp/terraform-provider-aws/issues/19765))
* resource/aws_cloudformation_stack_set: Retry when `OperationInProgress` errors are returned from the AWS API ([#10969](https://github.com/hashicorp/terraform-provider-aws/issues/10969))
* resource/aws_cloudformation_stack_set_instance: Retry when `OperationInProgress` errors are returned from the AWS API ([#10969](https://github.com/hashicorp/terraform-provider-aws/issues/10969))
* resource/aws_config_organization_conformance_pack: Add configurable timeouts ([#20560](https://github.com/hashicorp/terraform-provider-aws/issues/20560))
* resource/aws_redshift_cluster: Add `cluster_nodes` attribute ([#4563](https://github.com/hashicorp/terraform-provider-aws/issues/4563))
* resource/aws_s3_bucket: Retry on `PutBucketEncryption` HTTP 409 errors due to eventual consistency ([#11795](https://github.com/hashicorp/terraform-provider-aws/issues/11795))
* resource/aws_sagemaker_notebook_instance: Add `platform_identifier` argument ([#20711](https://github.com/hashicorp/terraform-provider-aws/issues/20711))
* resource/aws_service_discovery_service: Add `force_destroy` argument ([#3538](https://github.com/hashicorp/terraform-provider-aws/issues/3538))
* resource_aws_route53_health_check: Add `RECOVERY_CONTROL` health check type and `routing_control_arn` argument ([#20731](https://github.com/hashicorp/terraform-provider-aws/issues/20731))
* resource_vpn_connection: Handle paginated response when reading Transit Gateway Attachments ([#20775](https://github.com/hashicorp/terraform-provider-aws/issues/20775))

BUG FIXES:

* resource/aws_ecs_cluster: Ensure that `setting` attribute is set consistently ([#20720](https://github.com/hashicorp/terraform-provider-aws/issues/20720))
* resource/aws_pinpoint_email_channel: When specifying the `configuration_set` parameter, use the name of the set instead of the ARN. ([#20691](https://github.com/hashicorp/terraform-provider-aws/issues/20691))
* resource/aws_route53_record: Support `set_identifier` values containing `_` ([#13453](https://github.com/hashicorp/terraform-provider-aws/issues/13453))

## 3.56.0 (August 26, 2021)

FEATURES:

* **New Data Source:** `aws_dx_connection` ([#17852](https://github.com/hashicorp/terraform-provider-aws/issues/17852))
* **New Data Source:** `aws_dx_location` ([#9735](https://github.com/hashicorp/terraform-provider-aws/issues/9735))
* **New Data Source:** `aws_dx_locations` ([#9735](https://github.com/hashicorp/terraform-provider-aws/issues/9735))
* **New Resource:** `aws_appstream_stack` ([#20547](https://github.com/hashicorp/terraform-provider-aws/issues/20547))
* **New Resource:** `aws_autoscaling_group_tag` ([#20009](https://github.com/hashicorp/terraform-provider-aws/issues/20009))
* **New Resource:** `aws_dynamodb_tag` ([#13783](https://github.com/hashicorp/terraform-provider-aws/issues/13783))
* **New Resource:** `aws_ecs_tag` ([#13783](https://github.com/hashicorp/terraform-provider-aws/issues/13783))
* **New Resource:** `aws_route53recoverycontrolconfig_cluster` ([#20568](https://github.com/hashicorp/terraform-provider-aws/issues/20568))
* **New Resource:** `aws_route53recoverycontrolconfig_control_panel` ([#20568](https://github.com/hashicorp/terraform-provider-aws/issues/20568))
* **New Resource:** `aws_route53recoverycontrolconfig_routing_control` ([#20568](https://github.com/hashicorp/terraform-provider-aws/issues/20568))
* **New Resource:** `aws_route53recoverycontrolconfig_safety_rule` ([#20568](https://github.com/hashicorp/terraform-provider-aws/issues/20568))
* **New Resource:** `aws_route53recoveryreadiness_cell` ([#20526](https://github.com/hashicorp/terraform-provider-aws/issues/20526))
* **New Resource:** `aws_route53recoveryreadiness_readiness_check` ([#20526](https://github.com/hashicorp/terraform-provider-aws/issues/20526))
* **New Resource:** `aws_route53recoveryreadiness_recovery_group` ([#20526](https://github.com/hashicorp/terraform-provider-aws/issues/20526))
* **New Resource:** `aws_route53recoveryreadiness_resource_set` ([#20526](https://github.com/hashicorp/terraform-provider-aws/issues/20526))

ENHANCEMENTS:

* data-source/aws_elasticache_user: Mark `passwords` attribute as sensitive. ([#20629](https://github.com/hashicorp/terraform-provider-aws/issues/20629))
* data-source/aws_efs_file_system: Add ability to filter results by `tags` ([#20399](https://github.com/hashicorp/terraform-provider-aws/issues/20399))
* data-source/aws_route53_delegation_set: Add `arn` attribute ([#20664](https://github.com/hashicorp/terraform-provider-aws/issues/20664))
* data-source/aws_route53_zone: Add `arn` attribute ([#20652](https://github.com/hashicorp/terraform-provider-aws/issues/20652))
* resource/aws_dx_connection: Add `owner_account_id` attribute ([#17852](https://github.com/hashicorp/terraform-provider-aws/issues/17852))
* resource/aws_dx_connection: Add `provider_name` argument ([#17852](https://github.com/hashicorp/terraform-provider-aws/issues/17852))
* resource/aws_dx_lag: Add `owner_account_id` attribute ([#17852](https://github.com/hashicorp/terraform-provider-aws/issues/17852))
* resource/aws_dx_lag: Add `provider_name` argument ([#17852](https://github.com/hashicorp/terraform-provider-aws/issues/17852))
* resource/aws_eks_node_group: Add `update_config` argument to support parallel node upgrades ([#20137](https://github.com/hashicorp/terraform-provider-aws/issues/20137))
* resource/aws_elasticache_user: Mark `passwords` argument as sensitive. ([#20629](https://github.com/hashicorp/terraform-provider-aws/issues/20629))
* resource/aws_fsx_lustre_filesystem: Allow creating filesystem from backup using `backup_id`. ([#20614](https://github.com/hashicorp/terraform-provider-aws/issues/20614))
* resource/aws_fsx_windows_filesystem: Allow creating filesystem from backup using `backup_id`. ([#20643](https://github.com/hashicorp/terraform-provider-aws/issues/20643))
* resource/aws_route53_delegation_set: Add `arn` attribute ([#20664](https://github.com/hashicorp/terraform-provider-aws/issues/20664))
* resource/aws_route53_delegation_set: Add plan time validation for `reference_name` ([#20664](https://github.com/hashicorp/terraform-provider-aws/issues/20664))
* resource/aws_route53_health_check: Add `arn` attribute. ([#20653](https://github.com/hashicorp/terraform-provider-aws/issues/20653))
* resource/aws_route53_health_check: Add plan time validation for `failure_threshold`, `ip_address`, `fqdn`, `port`, `resource_path`, `search_string`, `child_healthchecks`. ([#20653](https://github.com/hashicorp/terraform-provider-aws/issues/20653))
* resource/aws_route53_query_log: Add `arn` attribute. ([#20666](https://github.com/hashicorp/terraform-provider-aws/issues/20666))
* resource/aws_route53_zone: Add `arn` attribute ([#20652](https://github.com/hashicorp/terraform-provider-aws/issues/20652))
* resource/aws_route53_zone: Add plan time validation for `comment` ([#20652](https://github.com/hashicorp/terraform-provider-aws/issues/20652))
* resource/aws_s3_bucket_inventory: Add missing values to `optional_fields` argument ([#20658](https://github.com/hashicorp/terraform-provider-aws/issues/20658))

BUG FIXES:

* data-source/aws_kms_public_key: Correctly base64 encode `public_key` value ([#19944](https://github.com/hashicorp/terraform-provider-aws/issues/19944))
* data-source/aws_route53_resolver_rule: Fix lack of pagination when listing rules ([#20642](https://github.com/hashicorp/terraform-provider-aws/issues/20642))
* resource/aws_codebuild_webhook: Only update `build_type` if a value is specified ([#20671](https://github.com/hashicorp/terraform-provider-aws/issues/20671))
* resource/aws_route53_delegation_set: Properly remove from state when resource does not exist ([#20664](https://github.com/hashicorp/terraform-provider-aws/issues/20664))
* resource/aws_route53_query_log: Properly remove from state when resource does not exist ([#20666](https://github.com/hashicorp/terraform-provider-aws/issues/20666))

## 3.55.0 (August 19, 2021)

FEATURES:

* **New Data Source:** `aws_iam_roles` ([#18585](https://github.com/hashicorp/terraform-provider-aws/issues/18585))
* **New Data Source:** `aws_subnets` ([#18803](https://github.com/hashicorp/terraform-provider-aws/issues/18803))
* **New Resource:** `aws_chime_voice_connector_group` ([#20565](https://github.com/hashicorp/terraform-provider-aws/issues/20565))
* **New Resource:** `aws_fsx_backup` ([#20569](https://github.com/hashicorp/terraform-provider-aws/issues/20569))
* **New Resource:** `aws_sagemaker_device_fleet` ([#20058](https://github.com/hashicorp/terraform-provider-aws/issues/20058))
* **New Resource:** `aws_sagemaker_human_task_ui` ([#20570](https://github.com/hashicorp/terraform-provider-aws/issues/20570))

ENHANCEMENTS:

* aws/resource_aws_appconfig_deployment: Add `state` attribute ([#20288](https://github.com/hashicorp/terraform-provider-aws/issues/20288))
* resource/aws_db_parameter_group: Allow parameter values to be mixed case, prioritize certain parameters when chunking, and avoid diffs with mixed-case parameter names ([#18818](https://github.com/hashicorp/terraform-provider-aws/issues/18818))
* resource/aws_dms_endpoint: Add `s3_settings.data_format`, `s3_settings.parquet_timestamp_in_millisecond`, `s3_settings.parquet_version`, `s3_settings.encryption_mode` and `s3_settings.server_side_encryption_kms_key_id` arguments. ([#17591](https://github.com/hashicorp/terraform-provider-aws/issues/17591))
* resource/aws_lambda_function: Add support for `python3.9` `runtime` value ([#20593](https://github.com/hashicorp/terraform-provider-aws/issues/20593))
* resource/aws_lambda_layer_version: Add support for `python3.9` `compatible_runtimes` value ([#20593](https://github.com/hashicorp/terraform-provider-aws/issues/20593))
* resource/aws_wafv2: Add missing values to `text_transformation` argument for `aws_wafv2_web_acl` and `aws_wafv2_rule_group` resources ([#20564](https://github.com/hashicorp/terraform-provider-aws/issues/20564))

BUG FIXES:

* aws/resource_aws_appconfig_deployment: Remove internal waiter after start of deployment ([#20288](https://github.com/hashicorp/terraform-provider-aws/issues/20288))
* aws/resource_aws_cloudwatch_event_rule: Correctly handle ARN in `event_bus_name` argument ([#20312](https://github.com/hashicorp/terraform-provider-aws/issues/20312))
* aws/resource_aws_cloudwatch_event_target: Correctly handle ARN in `event_bus_name` argument ([#20312](https://github.com/hashicorp/terraform-provider-aws/issues/20312))
* resource/aws_eks_addon: Treat `DEGRADED` as a pending state during creation ([#20562](https://github.com/hashicorp/terraform-provider-aws/issues/20562))
* resource/aws_eks_identity_provider_config: Increase Create and Delete timeouts to 40 minutes ([#20561](https://github.com/hashicorp/terraform-provider-aws/issues/20561))
* resource/aws_elasticache_user: Correctly update `passwords` ([#20530](https://github.com/hashicorp/terraform-provider-aws/issues/20530))
* resource/aws_lambda_function: Fix `handler`, `runtime` attribute validation for `package_type` is `Zip` ([#20575](https://github.com/hashicorp/terraform-provider-aws/issues/20575))
* resource/aws_lambda_function: fix Osaka ap-northeast-3 lambda function creation, failing due to code signer service not available in the region. ([#20555](https://github.com/hashicorp/terraform-provider-aws/issues/20555))
* resource/aws_rds_cluster_parameter_group: Handle paginated response when reading parameters from RDS cluster parameter group. ([#16010](https://github.com/hashicorp/terraform-provider-aws/issues/16010))
* resource/aws_storagegateway_smb_file_share: Only set `oplocks_enabled` if a value is specified in configuration ([#20579](https://github.com/hashicorp/terraform-provider-aws/issues/20579))

## 3.54.0 (August 12, 2021)

FEATURES:

* **New Resource:** `aws_chime_voice_connector` ([#19504](https://github.com/hashicorp/terraform-provider-aws/issues/19504))
* **New Resource:** `aws_shield_protection_group` ([#20491](https://github.com/hashicorp/terraform-provider-aws/issues/20491))

ENHANCEMENTS:

* data-source/aws_workspaces_directory: Add `workspace_access_properties.device_type_linux` attribute ([#20462](https://github.com/hashicorp/terraform-provider-aws/issues/20462))
* resource/aws_athena_workgroup: Add `requester_pays_enabled` argument ([#20457](https://github.com/hashicorp/terraform-provider-aws/issues/20457))
* resource/aws_cloudwatch_metric_alarm: Add support for `account_id` ([#20541](https://github.com/hashicorp/terraform-provider-aws/issues/20541))
* resource/aws_codebuild_webhook: Add support for `build_type` ([#20480](https://github.com/hashicorp/terraform-provider-aws/issues/20480))
* resource/aws_db_instance: Use engine_version and engine_version_actual to set and track engine versions ([#20207](https://github.com/hashicorp/terraform-provider-aws/issues/20207))
* resource/aws_workspaces_directory: Add `workspace_access_properties.device_type_linux` argument ([#20462](https://github.com/hashicorp/terraform-provider-aws/issues/20462))

BUG FIXES:

* aws/resource_aws_imagebuilder_infrastructure_configuration: Always set `terminate_instance_on_failure` on create and update ([#20464](https://github.com/hashicorp/terraform-provider-aws/issues/20464))
* resource/aws_iot_topic_rule: Correctly update resource on `error_action` change ([#16471](https://github.com/hashicorp/terraform-provider-aws/issues/16471))
* resource/aws_iot_topic_rule: Enhance handling of IAM eventual consistency errors during create ([#20467](https://github.com/hashicorp/terraform-provider-aws/issues/20467))
* resource/aws_synthetics_canary: Correctly report any resource creation errors ([#20463](https://github.com/hashicorp/terraform-provider-aws/issues/20463))

## 3.53.0 (August 05, 2021)

ENHANCEMENTS:

* data-source/aws_acm_certificate: Add status attribute ([#20232](https://github.com/hashicorp/terraform-provider-aws/issues/20232))
* data-source/aws_ec2_coip_pool: Add `arn` attribute ([#17046](https://github.com/hashicorp/terraform-provider-aws/issues/17046))
* resource/aws_appconfig_deployment: Include predefined strategies in plan time validation of `deployment_strategy_id` ([#20420](https://github.com/hashicorp/terraform-provider-aws/issues/20420))
* resource/aws_autoscaling_schedule: Add `time_zone` argument ([#19829](https://github.com/hashicorp/terraform-provider-aws/issues/19829))
* resource/aws_db_instance: Add `customer_owned_ip_enabled` argument ([#17864](https://github.com/hashicorp/terraform-provider-aws/issues/17864))
* resource/aws_db_instance: Add `nchar_character_set_name` argument ([#20437](https://github.com/hashicorp/terraform-provider-aws/issues/20437))
* resource/aws_kms_external_key: Add `bypass_policy_lockout_safety_check` argument ([#18117](https://github.com/hashicorp/terraform-provider-aws/issues/18117))
* resource/aws_kms_key: Add `bypass_policy_lockout_safety_check` argument ([#18117](https://github.com/hashicorp/terraform-provider-aws/issues/18117))
* resource/aws_launch_template: Allow all supported resource types `tag_specifications.resource_type` ([#20409](https://github.com/hashicorp/terraform-provider-aws/issues/20409))
* resource/aws_redshift_parameter_group: Make Redshift parameters case sensitive. ([#19772](https://github.com/hashicorp/terraform-provider-aws/issues/19772))

BUG FIXES:

* aws/resource_aws_amplify_branch: Correctly handle branch names that contain '/' ([#20426](https://github.com/hashicorp/terraform-provider-aws/issues/20426))
* aws/resource_aws_apigateway_vpc_link: Ensure deletion does not return an error when resource is not found ([#20441](https://github.com/hashicorp/terraform-provider-aws/issues/20441))
* aws/resource_aws_instance: Fix running `terraform plan` with with `skip_credentials_validation=true` ([#20357](https://github.com/hashicorp/terraform-provider-aws/issues/20357))
* aws/resource_aws_instance: Fix state refresh when launch template was deleted ([#20357](https://github.com/hashicorp/terraform-provider-aws/issues/20357))

## 3.52.0 (July 29, 2021)

FEATURES:

* **New Resource:** `aws_sagemaker_workforce` ([#20065](https://github.com/hashicorp/terraform-provider-aws/issues/20065))
* **New Resource:** `aws_sagemaker_workteam` ([#20122](https://github.com/hashicorp/terraform-provider-aws/issues/20122))
* **New Resource:** `aws_storagegateway_file_system_association` ([#20082](https://github.com/hashicorp/terraform-provider-aws/issues/20082))

ENHANCEMENTS:

* data-source/aws_ec2_instance_type_offerings: Add `locations` and `location_types` attributes ([#16704](https://github.com/hashicorp/terraform-provider-aws/issues/16704))
* data-source/aws_lb: Add ability to filter results by `tags` ([#6458](https://github.com/hashicorp/terraform-provider-aws/issues/6458))
* data-source/aws_qldb_ledger: Add `permissions_mode` attribute ([#20302](https://github.com/hashicorp/terraform-provider-aws/issues/20302))
* resource/aws_budgets_budget: Add the `cost_filter` argument which allows multiple `values` to be specified per filter. This new argument will eventually replace the `cost_filters` argument ([#9092](https://github.com/hashicorp/terraform-provider-aws/issues/9092))
* resource/aws_budgets_budget: Change `time_period_start` to an optional argument. If you don't specify a start date, AWS defaults to the start of your chosen time period ([#9092](https://github.com/hashicorp/terraform-provider-aws/issues/9092))
* resource/aws_cognito_user_pool_client: Set `callback_urls` and `logout_urls` as computed. ([#20065](https://github.com/hashicorp/terraform-provider-aws/issues/20065))
* resource/aws_dx_connection: Add support for `100Gbps` `bandwidth` [#20364](https://github.com/hashicorp/terraform-provider-aws/issues/20364))
* resource/aws_dx_lag: Add support for `100Gbps` `connections_bandwidth` [#20364](https://github.com/hashicorp/terraform-provider-aws/issues/20364))
* resource/aws_qldb_ledger: Add `permissions_mode` support ([#20302](https://github.com/hashicorp/terraform-provider-aws/issues/20302))
* resource/aws_rds_cluster: Use engine_version and engine_version_actual to set and track engine versions ([#20211](https://github.com/hashicorp/terraform-provider-aws/issues/20211))
* resource/aws_rds_cluster_instance: Use engine_version and engine_version_actual to set and track engine versions ([#20211](https://github.com/hashicorp/terraform-provider-aws/issues/20211))
* resource/aws_s3_bucket_object: Existing resource can now be imported ([#10036](https://github.com/hashicorp/terraform-provider-aws/issues/10036))
* resource/aws_sagemaker_model: Add `inference_execution_config`. ([#20066](https://github.com/hashicorp/terraform-provider-aws/issues/20066))
* resource/aws_secretsmanager_secret: Add replica support ([#20293](https://github.com/hashicorp/terraform-provider-aws/issues/20293))
* resource/aws_storagegateway_gateway: Add new option for gateway_type, `FILE_FSX_SMB`, to be used with `aws_storagegateway_file_system_association` ([#20082](https://github.com/hashicorp/terraform-provider-aws/issues/20082))

BUG FIXES:

* resource/aws_elasticache_user: Correctly handle user modifications and deletion ([#20339](https://github.com/hashicorp/terraform-provider-aws/issues/20339))
* resource/aws_budgets_budget: Change the service name in the `arn` attribute from `budgetservice` to `budgets` ([#9092](https://github.com/hashicorp/terraform-provider-aws/issues/9092))
* resource/aws_budgets_budget: Suppress plan differences with trailing zeroes for `limit_amount` ([#9092](https://github.com/hashicorp/terraform-provider-aws/issues/9092))
* resource/aws_budgets_budget_action: Change the service name in the `arn` attribute from `budgetservice` to `budgets` ([#9092](https://github.com/hashicorp/terraform-provider-aws/issues/9092))
* resource/aws_lex_bot: Fix computed `version` for dependent resources ([#20336](https://github.com/hashicorp/terraform-provider-aws/issues/20336))
* resource/aws_lex_intent: Fix computed `version` for dependent resources ([#20336](https://github.com/hashicorp/terraform-provider-aws/issues/20336))
* resource/aws_lex_slot_type: Fix computed `version` for dependent resources ([#20336](https://github.com/hashicorp/terraform-provider-aws/issues/20336))

## 3.51.0 (July 22, 2021)

FEATURES:

* **New Data Source:** `aws_elasticache_user` ([#16629](https://github.com/hashicorp/terraform-provider-aws/issues/16629))
* **New Resource:** `aws_appconfig_deployment` ([#20172](https://github.com/hashicorp/terraform-provider-aws/issues/20172))
* **New Resource:** `aws_elasticache_user` ([#16629](https://github.com/hashicorp/terraform-provider-aws/issues/16629))
* **New Resource:** `aws_elasticache_user_group` ([#16504](https://github.com/hashicorp/terraform-provider-aws/issues/16504))

ENHANCEMENTS:

* resource/aws_cloudwatch_event_target: Add support for Redshift event target. ([#20256](https://github.com/hashicorp/terraform-provider-aws/issues/20256))
* resource/aws_glue_crawler: Add `sample_size` argument in `s3_target` block. ([#20203](https://github.com/hashicorp/terraform-provider-aws/issues/20203))
* resource/aws_instance: Add support for configuration with Launch Template ([#10807](https://github.com/hashicorp/terraform-provider-aws/issues/10807))
* resource/aws_servicecatalog_provisioned_product: Increase timeouts to align with CloudFormation (30 min.) ([#20254](https://github.com/hashicorp/terraform-provider-aws/issues/20254))
* resource/aws_storagegateway_smb_file_share: Add `bucket_region`, `oplocks_enabled` and `vpc_endpoint_dns_name` arguments ([#20234](https://github.com/hashicorp/terraform-provider-aws/issues/20234))

BUG FIXES:

* aws/resource_aws_lambda_event_source_mapping: Ignore `InvalidParameterValueException` error caused by IAM propagation when creating Lambda event source mapping with Kinesis stream source ([#20229](https://github.com/hashicorp/terraform-provider-aws/issues/20229))
* aws/resource_aws_route_table_association: Correctly handle `associated` as a pending state when waiting for deletion of an association ([#20265](https://github.com/hashicorp/terraform-provider-aws/issues/20265))

## 3.50.0 (July 15, 2021)

NOTES:

* resource/aws_dx_gateway_association_proposal: If an accepted Proposal reaches end-of-life and is removed by AWS do not recreate the resource, instead refreshing Terraform state from the resource's Direct Connect Gateway ID and Associated Gateway ID. ([#19741](https://github.com/hashicorp/terraform-provider-aws/issues/19741))

FEATURES:

* **New Resource:** `aws_appconfig_application` ([#19307](https://github.com/hashicorp/terraform-provider-aws/issues/19307))
* **New Resource:** `aws_appconfig_configuration_profile` ([#19320](https://github.com/hashicorp/terraform-provider-aws/issues/19320))
* **New Resource:** `aws_appconfig_deployment_strategy` ([#19359](https://github.com/hashicorp/terraform-provider-aws/issues/19359))
* **New Resource:** `aws_appconfig_environment` ([#19307](https://github.com/hashicorp/terraform-provider-aws/issues/19307))
* **New Resource:** `aws_appconfig_hosted_configuration_version` ([#19324](https://github.com/hashicorp/terraform-provider-aws/issues/19324))
* **New Resource:** `aws_config_organization_conformance_pack` ([#17298](https://github.com/hashicorp/terraform-provider-aws/issues/17298))
* **New Resource:** `aws_securityhub_organization_configuration` ([#19108](https://github.com/hashicorp/terraform-provider-aws/issues/19108))
* **New Resource:** `aws_securityhub_standards_control` ([#14714](https://github.com/hashicorp/terraform-provider-aws/issues/14714))

ENHANCEMENTS:

* resource/aws_cloudwatch_event_target: Add `enable_ecs_managed_tags`, `enable_execute_command`, `placement_constraints`, `propagate_tags`, and `tags` arguments to `ecs_target` block. ([#19975](https://github.com/hashicorp/terraform-provider-aws/issues/19975))
* resource/aws_cognito_user_pool_client: Add the `enable_token_revocation` argument to support targeted sign out ([#20031](https://github.com/hashicorp/terraform-provider-aws/issues/20031))
* resource/aws_fsx_windows_file_system: Add `aliases` argument ([#20054](https://github.com/hashicorp/terraform-provider-aws/issues/20054))
* resource/aws_guardduty_detector: Add `datasources` argument ([#19954](https://github.com/hashicorp/terraform-provider-aws/issues/19954))
* resource/aws_guardduty_organization_configuration: Add `datasources` argument ([#15241](https://github.com/hashicorp/terraform-provider-aws/issues/15241))
* resource/aws_iam_access_key: Add encrypted SES SMTP password ([#19579](https://github.com/hashicorp/terraform-provider-aws/issues/19579))
* resource/aws_kms_key: Add plan time validation to `description`. ([#19967](https://github.com/hashicorp/terraform-provider-aws/issues/19967))
* resource/aws_s3_bucket: Add the delete_marker_replication_status argument for V2 replication configurations ([#19323](https://github.com/hashicorp/terraform-provider-aws/issues/19323))
* resource/aws_s3_bucket_object: Add `source_hash` argument to compliment `etag`'s encryption limitations ([#11522](https://github.com/hashicorp/terraform-provider-aws/issues/11522))
* resource/aws_sagemaker_domain: Add support for `retention_policy` ([#18562](https://github.com/hashicorp/terraform-provider-aws/issues/18562))
* resource/aws_wafv2_web_acl: Support `scope_down_statement` on `managed_rule_group_statement` ([#19407](https://github.com/hashicorp/terraform-provider-aws/issues/19407))

BUG FIXES:

* resource/aws_cognito_user_pool_client: Allow the `default_redirect_uri` argument value to be an empty string ([#20031](https://github.com/hashicorp/terraform-provider-aws/issues/20031))
* resource/aws_cognito_user_pool_client: Retry on `ConcurrentModificationException` ([#20031](https://github.com/hashicorp/terraform-provider-aws/issues/20031))
* resource/aws_datasync_location_s3: Correctly parse S3 on Outposts location URI ([#19859](https://github.com/hashicorp/terraform-provider-aws/issues/19859))
* resource/aws_db_instance: Ignore allocated_storage for replica at creation time ([#12548](https://github.com/hashicorp/terraform-provider-aws/issues/12548))
* resource/aws_elasticache_replication_group: Cannot set `cluster_mode.replicas_per_node_group` when member of Global Replication Group ([#20111](https://github.com/hashicorp/terraform-provider-aws/issues/20111))

## 3.49.0 (July 08, 2021)

FEATURES:

* **New Resource:** `aws_eks_identity_provider_config` ([#17959](https://github.com/hashicorp/terraform-provider-aws/issues/17959))
* **New Resource:** `aws_rds_cluster_role_association` ([#12370](https://github.com/hashicorp/terraform-provider-aws/issues/12370))

ENHANCEMENTS:

* aws_rds_cluster: Set `iam_roles` as Computed to prevent drift when the `aws_rds_cluster_role_association` resource is used ([#12370](https://github.com/hashicorp/terraform-provider-aws/issues/12370))
* resource/aws_transfer_server: Add `security_group_ids` argument to `endpoint_details` configuration block. ([#17539](https://github.com/hashicorp/terraform-provider-aws/issues/17539))

BUG FIXES:

* data-source/aws_lakeformation_permissions: Fix various problems with permissions including select-only ([#20108](https://github.com/hashicorp/terraform-provider-aws/issues/20108))
* resource/aws_eks_cluster: Don't associate an `encryption_config` if there's already one ([#19986](https://github.com/hashicorp/terraform-provider-aws/issues/19986))
* resource/aws_lakeformation_permissions: Fix various problems with permissions including select-only ([#20108](https://github.com/hashicorp/terraform-provider-aws/issues/20108))
* resource/aws_ram_resource_share_accepter: Allow destroy even where AWS API provides no way to disassociate ([#19718](https://github.com/hashicorp/terraform-provider-aws/issues/19718))

## 3.48.0 (July 02, 2021)

FEATURES:

* **New Data Source:** `aws_iam_session_context` ([#19957](https://github.com/hashicorp/terraform-provider-aws/issues/19957))
* **New Data Source:** `aws_servicecatalog_launch_paths` ([#19572](https://github.com/hashicorp/terraform-provider-aws/issues/19572))
* **New Data Source:** `aws_servicecatalog_portfolio_constraints` ([#19813](https://github.com/hashicorp/terraform-provider-aws/issues/19813))
* **New Resource:** `aws_cloudfront_monitoring_subscription` ([#18083](https://github.com/hashicorp/terraform-provider-aws/issues/18083))
* **New Resource:** `aws_servicecatalog_provisioned_product` ([#19459](https://github.com/hashicorp/terraform-provider-aws/issues/19459))

ENHANCEMENTS:

* resource/aws_fsx_windows_file_system: Add `audit_log_configuration` argument. ([#19970](https://github.com/hashicorp/terraform-provider-aws/issues/19970))

BUG FIXES:

* resource/aws_cloudwatch_event_target: Don't crash if `sqs_target` configuration block is empty. ([#19946](https://github.com/hashicorp/terraform-provider-aws/issues/19946))
* resource/aws_mwaa_environment: Changes to the `kms_key` argument force resource recreation ([#19994](https://github.com/hashicorp/terraform-provider-aws/issues/19994))

## 3.47.0 (June 24, 2021)

FEATURES:

* **New Resource:** `aws_cloudwatch_event_bus_policy` ([#16874](https://github.com/hashicorp/terraform-provider-aws/issues/16874))
* **New Resource:** `aws_efs_backup_policy` ([#18006](https://github.com/hashicorp/terraform-provider-aws/issues/18006))
* **New Resource:** `aws_elasticsearch_domain_saml_options` ([#19497](https://github.com/hashicorp/terraform-provider-aws/issues/19497))
* **New Resource:** `aws_neptune_cluster_endpoint` ([#19898](https://github.com/hashicorp/terraform-provider-aws/issues/19898))

ENHANCEMENTS:

* resource/aws_default_route_table: Add retries when creating, deleting and replacing routes ([#19426](https://github.com/hashicorp/terraform-provider-aws/issues/19426))
* resource/aws_default_route_table: Add retries when creating, deleting and replacing routes ([#19426](https://github.com/hashicorp/terraform-provider-aws/issues/19426))
* resource/aws_ecs_capacity_provider: Allow updates to the `auto_scaling_group_provider` argument without recreating the resource ([#16942](https://github.com/hashicorp/terraform-provider-aws/issues/16942))
* resource/aws_eks_cluster: Allow updates to `encryption_config` ([#19144](https://github.com/hashicorp/terraform-provider-aws/issues/19144))
* resource/aws_lb_target_group: Add support for `app_cookie` stickiness type and `cookie_name` argument ([#18102](https://github.com/hashicorp/terraform-provider-aws/issues/18102))
* resource/aws_main_route_table_association: Wait for association to reach the required state ([#19426](https://github.com/hashicorp/terraform-provider-aws/issues/19426))
* resource/aws_neptune_cluster: Add `copy_tags_to_snapshot` argument ([#19899](https://github.com/hashicorp/terraform-provider-aws/issues/19899))
* resource/aws_route: Add retries when creating, deleting and replacing routes ([#19426](https://github.com/hashicorp/terraform-provider-aws/issues/19426))
* resource/aws_route_table: Add retries when creating, deleting and replacing routes ([#19426](https://github.com/hashicorp/terraform-provider-aws/issues/19426))
* resource/aws_route_table_association: Wait for association to reach the required state ([#19426](https://github.com/hashicorp/terraform-provider-aws/issues/19426))

BUG FIXES:

* resource/aws_backup_vault_policy: Correctly handle deleting policy of deleted vault ([#19854](https://github.com/hashicorp/terraform-provider-aws/issues/19854))
* resource/aws_backup_vault_policy: Correctly handle reading policy of deleted vault ([#19749](https://github.com/hashicorp/terraform-provider-aws/issues/19749))
* resource/aws_glue_catalog_database: Set `location_uri` as compute to prevent drift when `target_table` has `location_uri` set. ([#19743](https://github.com/hashicorp/terraform-provider-aws/issues/19743))
* resource/aws_glue_catalog_table: Fix updating `schema_reference` when columns are present. ([#19742](https://github.com/hashicorp/terraform-provider-aws/issues/19742))

## 3.46.0 (June 17, 2021)

FEATURES:

* **New Data Source:** `aws_appmesh_virtual_service` ([#19774](https://github.com/hashicorp/terraform-provider-aws/issues/19774))
* **New Data Source:** `aws_servicecatalog_portfolio` ([#19500](https://github.com/hashicorp/terraform-provider-aws/issues/19500))
* **New Resource:** `aws_budgets_budget_action` ([#19554](https://github.com/hashicorp/terraform-provider-aws/issues/19554))
* **New Resource:** `aws_route53_resolver_firewall_config` ([#18733](https://github.com/hashicorp/terraform-provider-aws/issues/18733))

ENHANCEMENTS:

* resource/aws_cloudwatch_log_metric_filter: Add support for `unit` in the `metric_transformation` block. ([#19804](https://github.com/hashicorp/terraform-provider-aws/issues/19804))
* resource/aws_datasync_location_nfs: Add `mount_options` argument. ([#19767](https://github.com/hashicorp/terraform-provider-aws/issues/19767))
* resource/aws_datasync_location_nfs: Add plan time validation for `on_prem_config.agent_arns`, `server_hostname`, and `subdirectory`. ([#19767](https://github.com/hashicorp/terraform-provider-aws/issues/19767))
* resource/aws_datasync_location_nfs: Add support for updating. ([#19767](https://github.com/hashicorp/terraform-provider-aws/issues/19767))
* resource/aws_ecs_cluster: Add plan time validation for `name`. ([#19785](https://github.com/hashicorp/terraform-provider-aws/issues/19785))
* resource/aws_ecs_cluster: Add support for `configuration`. ([#19785](https://github.com/hashicorp/terraform-provider-aws/issues/19785))
* resource/aws_eks_node_group: Allow minimum value of `0` for `desired_size` and `min_size` in the `scaling_config` configuration block ([#19810](https://github.com/hashicorp/terraform-provider-aws/issues/19810))
* resource/aws_spot_fleet_request: Add `on_demand_allocation_strategy`, `on_demand_max_total_price`, and `on_demand_target_capacity` arguments ([#13127](https://github.com/hashicorp/terraform-provider-aws/issues/13127))

BUG FIXES:

* data-source/aws_directory_service_directory: Check VpcSettings and ConnectSettings for nil values ([#19820](https://github.com/hashicorp/terraform-provider-aws/issues/19820))
* data-source/aws_lakeformation_permissions: Fix diffs resulting from order of column names and exclude column names ([#19817](https://github.com/hashicorp/terraform-provider-aws/issues/19817))
* resource/aws_cognito_identity_provider: Fix updating `idp_identifiers` crash. ([#19819](https://github.com/hashicorp/terraform-provider-aws/issues/19819))
* resource/aws_glue_trigger: Fix default timeouts for Create and Delete operations ([#19827](https://github.com/hashicorp/terraform-provider-aws/issues/19827))
* resource/aws_lakeformation_permissions: Fix bug preventing updates (inconsistent result) ([#19817](https://github.com/hashicorp/terraform-provider-aws/issues/19817))
* resource/aws_lakeformation_permissions: Fix bug where resource is not properly removed from state ([#19817](https://github.com/hashicorp/terraform-provider-aws/issues/19817))
* resource/aws_lakeformation_permissions: Fix diffs resulting only from order of column names and exclude column names ([#19817](https://github.com/hashicorp/terraform-provider-aws/issues/19817))
* resource/aws_lambda_event_source_mapping: Enhance handling of IAM eventual consistency errors during create ([#19831](https://github.com/hashicorp/terraform-provider-aws/issues/19831))
* resource/aws_sqs_queue: Correctly handle the default `kms_data_key_reuse_period_seconds` value of `300` for unencrypted queues ([#19834](https://github.com/hashicorp/terraform-provider-aws/issues/19834))

## 3.45.0 (June 10, 2021)

FEATURES:

* **New Data Source:** `aws_appmesh_mesh` ([#19577](https://github.com/hashicorp/terraform-provider-aws/issues/19577))
* **New Data Source:** `aws_globalaccelerator_accelerator` ([#19647](https://github.com/hashicorp/terraform-provider-aws/issues/19647))

ENHANCEMENTS:

* data-source/aws_nat_gateway: Add `connectivity_type` attribute ([#19758](https://github.com/hashicorp/terraform-provider-aws/issues/19758))
* data-source/aws_transfer_server: Add `domain` attribute. ([#19691](https://github.com/hashicorp/terraform-provider-aws/issues/19691))
* resource/aws_cognito_user_pool: Add `custom_domain`, `domain`, and `estimated_number_of_users` attributes ([#16502](https://github.com/hashicorp/terraform-provider-aws/issues/16502))
* resource/aws_cognito_user_pool: Add `custom_email_sender`, `custom_sms_sender`, and `kms_key_id` to `lambda_config` ([#16502](https://github.com/hashicorp/terraform-provider-aws/issues/16502))
* resource/aws_cognito_user_pool: Add plan time validation for `name` ([#16502](https://github.com/hashicorp/terraform-provider-aws/issues/16502))
* resource/aws_cognito_user_pool_client: Add plan time validation for `id_token_validity` and `access_token_validity`. ([#19702](https://github.com/hashicorp/terraform-provider-aws/issues/19702))
* resource/aws_cur_report_definition: Add `arn` attribute. ([#19705](https://github.com/hashicorp/terraform-provider-aws/issues/19705))
* resource/aws_cur_report_definition: Add plan time validation for `report_name`. ([#19705](https://github.com/hashicorp/terraform-provider-aws/issues/19705))
* resource/aws_cur_report_definition: Support updating definition. ([#19705](https://github.com/hashicorp/terraform-provider-aws/issues/19705))
* resource/aws_datasync_location_smb: Add plan time validation for `domain`, `agent_arns`, `password`, `server_hostname`, `subdirectory`, and `user`. ([#19753](https://github.com/hashicorp/terraform-provider-aws/issues/19753))
* resource/aws_datasync_location_smb: Add support for updating. ([#19753](https://github.com/hashicorp/terraform-provider-aws/issues/19753))
* resource/aws_default_vpc_dhcp_options: Add `owner_id` argument. ([#19656](https://github.com/hashicorp/terraform-provider-aws/issues/19656))
* resource/aws_ecs_task_definition: Add plan time validation for `family` and `requires_compatibilities`. ([#19670](https://github.com/hashicorp/terraform-provider-aws/issues/19670))
* resource/aws_ecs_task_definition: Add support for `ephemeral_storage`. ([#19694](https://github.com/hashicorp/terraform-provider-aws/issues/19694))
* resource/aws_ecs_task_definition: Add support for `fsx_windows_file_server_volume_configuration`. ([#19670](https://github.com/hashicorp/terraform-provider-aws/issues/19670))
* resource/aws_fsx_lustre_filesystem: Add `data_compression_type` argument. ([#19664](https://github.com/hashicorp/terraform-provider-aws/issues/19664))
* resource/aws_nat_gateway: Add `connectivity_type` argument ([#19758](https://github.com/hashicorp/terraform-provider-aws/issues/19758))
* resource/aws_sqs_queue: Add `deduplication_scope` and `fifo_throughput_limit` arguments ([#19639](https://github.com/hashicorp/terraform-provider-aws/issues/19639))
* resource/aws_sqs_queue: Add `url` attribute ([#19639](https://github.com/hashicorp/terraform-provider-aws/issues/19639))
* resource/aws_transfer_server: Add `domain` argument. ([#19691](https://github.com/hashicorp/terraform-provider-aws/issues/19691))
* resource/aws_transfer_user: Add `posix_profile` argument. ([#19693](https://github.com/hashicorp/terraform-provider-aws/issues/19693))

BUG FIXES:

* data-source/aws_acmpca_certificate_authority: Fix `error setting tags` ([#19681](https://github.com/hashicorp/terraform-provider-aws/issues/19681))
* data-source/aws_servicequotas_service_quota: Correctly handle errors embedded in API struct ([#19722](https://github.com/hashicorp/terraform-provider-aws/issues/19722))
* resource/aws_batch_job_definition: Suppress differences for empty `linuxParameters.devices` and `linuxParameters.tmpfs` arrays in the `container_properties` argument ([#19666](https://github.com/hashicorp/terraform-provider-aws/issues/19666))
* resource/aws_cloudwatch_event_target: Fix `ecs_target.launch_type` not allowing empty string values. ([#19703](https://github.com/hashicorp/terraform-provider-aws/issues/19555))
* resource/aws_cloudwatch_event_target: Increase the maximum allowed value for the `input_transformer` `input_paths` argument to 100 ([#19703](https://github.com/hashicorp/terraform-provider-aws/issues/19703))
* resource/aws_cloudwatch_metric_alarm: Allow extended statistics in the `stat` argument of the `metric` configuration block ([#19668](https://github.com/hashicorp/terraform-provider-aws/issues/19668))
* resource/aws_cognito_user_pool: Suppress diff for empty `account_recovery_setting`. ([#19704](https://github.com/hashicorp/terraform-provider-aws/issues/19704))
* resource/aws_cognito_user_pool_client: Fix plan time validation for `refresh_token_validity` ([#19702](https://github.com/hashicorp/terraform-provider-aws/issues/19702))
* resource/aws_iot_topic_rule: Allow tags containing `@` character ([#19677](https://github.com/hashicorp/terraform-provider-aws/issues/19677))
* resource/aws_lambda_function: Prevents perpetual diff in `vpc_config` ([#17610](https://github.com/hashicorp/terraform-provider-aws/issues/17610))
* resource/aws_servicequotas_service_quota: Correctly handle errors embedded in API struct ([#19722](https://github.com/hashicorp/terraform-provider-aws/issues/19722))
* resource/aws_sqs_queue: Allow `visibility_timeout_seconds` to be `0` when creating queue ([#19639](https://github.com/hashicorp/terraform-provider-aws/issues/19639))
* resource/aws_sqs_queue: Ensure that queue attributes propagate completely during Create and Update ([#19639](https://github.com/hashicorp/terraform-provider-aws/issues/19639))

## 3.44.0 (June 03, 2021)

FEATURES:

* **New Resource:** `aws_amplify_branch` ([#11937](https://github.com/hashicorp/terraform-provider-aws/issues/11937))
* **New Resource:** `aws_amplify_domain_association` ([#11938](https://github.com/hashicorp/terraform-provider-aws/issues/11938))
* **New Resource:** `aws_amplify_webhook` ([#11939](https://github.com/hashicorp/terraform-provider-aws/issues/11939))
* **New Resource:** `aws_servicecatalog_principal_portfolio_association` ([#19470](https://github.com/hashicorp/terraform-provider-aws/issues/19470))

ENHANCEMENTS:

* data-source/aws_launch_configuration: Add `throughput` attribute to `ebs_block_device` and `root_block_device` configuration blocks to support GP3 volumes ([#19632](https://github.com/hashicorp/terraform-provider-aws/issues/19632))
* resource/aws_acmpca_certificate_authority: Add `s3_object_acl` argument to `revocation_configuration.crl_configuration` configuration block ([#19578](https://github.com/hashicorp/terraform-provider-aws/issues/19578))
* resource/aws_cloudwatch_log_metric_filter: Add `dimensions` argument to `metric_transformation` configuration block ([#19625](https://github.com/hashicorp/terraform-provider-aws/issues/19625))
* resource/aws_cloudwatch_metric_alarm: Add plan time validation to `metric_query.metric.stat`. ([#19571](https://github.com/hashicorp/terraform-provider-aws/issues/19571))
* resource/aws_devicefarm_project: Add `default_job_timeout_minutes` and `tags` argument ([#19574](https://github.com/hashicorp/terraform-provider-aws/issues/19574))
* resource/aws_devicefarm_project: Add plan time validation for `name` ([#19574](https://github.com/hashicorp/terraform-provider-aws/issues/19574))
* resource/aws_fsx_lustre_filesystem: Allow updating `storage_capacity`. ([#19568](https://github.com/hashicorp/terraform-provider-aws/issues/19568))
* resource/aws_launch_configuration: Add `throughput` argument to `ebs_block_device` and `root_block_device` configuration blocks to support GP3 volumes ([#19632](https://github.com/hashicorp/terraform-provider-aws/issues/19632))

BUG FIXES:

* resource/aws_amplify_app: Mark the `enable_performance_mode` argument in the `auto_branch_creation_config` configuration block as `ForceNew` ([#11937](https://github.com/hashicorp/terraform-provider-aws/issues/11937))
* resource/aws_cloudwatch_event_api_destination: Fix crash on resource update ([#19654](https://github.com/hashicorp/terraform-provider-aws/issues/19654))
* resource/aws_elasticache_cluster: Fix provider-level `default_tags` support for resource ([#19615](https://github.com/hashicorp/terraform-provider-aws/issues/19615))
* resource/aws_iam_access_key: Fix status not defaulting to Active ([#19606](https://github.com/hashicorp/terraform-provider-aws/issues/19606))

## 3.43.0 (June 01, 2021)

FEATURES:

* **New Data Source:** `aws_cloudwatch_event_connection` ([#18905](https://github.com/hashicorp/terraform-provider-aws/issues/18905))
* **New Resource:** `aws_amplify_app` ([#15966](https://github.com/hashicorp/terraform-provider-aws/issues/15966))
* **New Resource:** `aws_amplify_backend_environment` ([#11936](https://github.com/hashicorp/terraform-provider-aws/issues/11936))
* **New Resource:** `aws_cloudwatch_event_api_destination` ([#18905](https://github.com/hashicorp/terraform-provider-aws/issues/18905))
* **New Resource:** `aws_cloudwatch_event_connection` ([#18905](https://github.com/hashicorp/terraform-provider-aws/issues/18905))
* **New Resource:** `aws_schemas_discoverer` ([#19100](https://github.com/hashicorp/terraform-provider-aws/issues/19100))
* **New Resource:** `aws_schemas_registry` ([#19100](https://github.com/hashicorp/terraform-provider-aws/issues/19100))
* **New Resource:** `aws_schemas_schema` ([#19100](https://github.com/hashicorp/terraform-provider-aws/issues/19100))
* **New Resource:** `aws_servicecatalog_budget_resource_association` ([#19452](https://github.com/hashicorp/terraform-provider-aws/issues/19452))
* **New Resource:** `aws_servicecatalog_provisioning_artifact` ([#19316](https://github.com/hashicorp/terraform-provider-aws/issues/19316))
* **New Resource:** `aws_servicecatalog_tag_option_resource_association` ([#19448](https://github.com/hashicorp/terraform-provider-aws/issues/19448))

ENHANCEMENTS:

* data-source/aws_msk_cluster: Add `bootstrap_brokers_sasl_iam` attribute ([#19404](https://github.com/hashicorp/terraform-provider-aws/issues/19404))
* resource/aws_cloudfront_distribution: Add `connection_attempts`, `connection_timeout`, and `origin_shield`. ([#16049](https://github.com/hashicorp/terraform-provider-aws/issues/16049))
* resource/aws_cloudtrail: Add `AWS::DynamoDB::Table` as an option for `event_selector`.`data_resource`.`type` ([#19559](https://github.com/hashicorp/terraform-provider-aws/issues/19559))
* resource/aws_ec2_capacity_reservation: Add `outpost_arn` argument ([#19535](https://github.com/hashicorp/terraform-provider-aws/issues/19535))
* resource/aws_ecs_service: Add support for ECS Anywhere with the `launch_type` `EXTERNAL` ([#19557](https://github.com/hashicorp/terraform-provider-aws/issues/19557))
* resource/aws_eks_node_group: Add `taint` argument ([#19482](https://github.com/hashicorp/terraform-provider-aws/issues/19482))
* resource/aws_elasticache_parameter_group: Add `tags` argument and `arn` and `tags_all` attributes ([#19551](https://github.com/hashicorp/terraform-provider-aws/issues/19551))
* resource/aws_lambda_event_source_mapping: Add `function_response_types` argument to support AWS Lambda checkpointing ([#19425](https://github.com/hashicorp/terraform-provider-aws/issues/19425))
* resource/aws_lambda_event_source_mapping: Add `queues` argument to support Amazon MQ for Apache ActiveMQ event sources ([#19425](https://github.com/hashicorp/terraform-provider-aws/issues/19425))
* resource/aws_lambda_event_source_mapping: Add `self_managed_event_source` and `source_access_configuration` arguments to support self-managed Apache Kafka event sources ([#19425](https://github.com/hashicorp/terraform-provider-aws/issues/19425))
* resource/aws_lambda_event_source_mapping: Add `tumbling_window_in_seconds` argument to support AWS Lambda streaming analytics calculations ([#19425](https://github.com/hashicorp/terraform-provider-aws/issues/19425))
* resource/aws_msk_cluster: Add `bootstrap_brokers_sasl_iam` attribute ([#19404](https://github.com/hashicorp/terraform-provider-aws/issues/19404))
* resource/aws_msk_cluster: Add `iam` argument to `client_authentication.sasl` configuration block ([#19404](https://github.com/hashicorp/terraform-provider-aws/issues/19404))
* resource/aws_msk_configuration: `kafka_versions` argument is optional ([#17571](https://github.com/hashicorp/terraform-provider-aws/issues/17571))
* resource/aws_sns_topic: Add `firehose_success_feedback_role_arn`, `firehose_success_feedback_sample_rate` and `firehose_failure_feedback_role_arn` arguments. ([#19528](https://github.com/hashicorp/terraform-provider-aws/issues/19528))
* resource/aws_sns_topic: Add `owner` attribute. ([#19528](https://github.com/hashicorp/terraform-provider-aws/issues/19528))
* resource/aws_sns_topic: Add plan time validation for `application_success_feedback_role_arn`, `application_failure_feedback_role_arn`, `http_success_feedback_role_arn`, `http_failure_feedback_role_arn`, `lambda_success_feedback_role_arn`, `lambda_failure_feedback_role_arn`, `sqs_success_feedback_role_arn`, `sqs_failure_feedback_role_arn`. ([#19528](https://github.com/hashicorp/terraform-provider-aws/issues/19528))

BUG FIXES:

* data-source/aws_launch_template: Add `interface_type` to `network_interfaces` attribute ([#19492](https://github.com/hashicorp/terraform-provider-aws/issues/19492))
* data-source/aws_mq_broker: Correct type for `logs.audit` attribute ([#19502](https://github.com/hashicorp/terraform-provider-aws/issues/19502))
* resource/aws_apprunner_service: Correctly configure `authentication_configuration`, `code_configuration`, and `image_configuration` nested arguments in API requests ([#19471](https://github.com/hashicorp/terraform-provider-aws/issues/19471))
* resource/aws_apprunner_service: Handle asynchronous IAM eventual consistency error on creation ([#19483](https://github.com/hashicorp/terraform-provider-aws/issues/19483))
* resource/aws_apprunner_service: Suppress `instance_configuration` `cpu` and `memory` differences ([#19483](https://github.com/hashicorp/terraform-provider-aws/issues/19483))
* resource/aws_batch_job_definition: Don't crash when setting `timeout.attempt_duration_seconds` to `null` ([#19505](https://github.com/hashicorp/terraform-provider-aws/issues/19505))
* resource/aws_cloudformation_stack: Avoid conflicts with `on_failure` and `disable_rollback` ([#10539](https://github.com/hashicorp/terraform-provider-aws/issues/10539))
* resource/aws_cloudwatch_event_api_destination: Reduce the maximum allowed value for the `invocation_rate_limit_per_second` argument to `300` ([#19594](https://github.com/hashicorp/terraform-provider-aws/issues/19594))
* resource/aws_ec2_managed_prefix_list: Fix crash with multiple description-only updates ([#19517](https://github.com/hashicorp/terraform-provider-aws/issues/19517))
* resource/aws_eks_addon: Use `service_account_role_arn`, if set, on updates ([#19454](https://github.com/hashicorp/terraform-provider-aws/issues/19454))
* resource/aws_glue_connection: `connection_properties` are optional ([#19375](https://github.com/hashicorp/terraform-provider-aws/issues/19375))
* resource/aws_lb_listener_rule: Allow blank string for `action.redirect.query` nested argument ([#19496](https://github.com/hashicorp/terraform-provider-aws/issues/19496))
* resource/aws_synthetics_canary: Change minimum `timeout_in_seconds` in `run_config` from `60` to `3` ([#19515](https://github.com/hashicorp/terraform-provider-aws/issues/19515))
* resource/aws_vpn_connection: Allow `local_ipv4_network_cidr`, `remote_ipv4_network_cidr`, `local_ipv6_network_cidr`, and `remote_ipv6_network_cidr` to be CIDRs of any size ([#17573](https://github.com/hashicorp/terraform-provider-aws/issues/17573))

## 3.42.0 (May 20, 2021)

FEATURES:

* **New Data Source:** `aws_service_discovery_dns_namespace` ([#6856](https://github.com/hashicorp/terraform-provider-aws/issues/6856))
* **New Resource:** `aws_cloudwatch_metric_stream` ([#18870](https://github.com/hashicorp/terraform-provider-aws/issues/18870))
* **New Resource:** `aws_servicecatalog_constraint` ([#19385](https://github.com/hashicorp/terraform-provider-aws/issues/19385))
* **New Resource:** `aws_servicecatalog_product_portfolio_association` ([#19385](https://github.com/hashicorp/terraform-provider-aws/issues/19385))
* **New Resource:** `aws_servicecatalog_service_action` ([#19369](https://github.com/hashicorp/terraform-provider-aws/issues/19369))

ENHANCEMENTS:

* resource/aws_autoscaling_policy: Add `PredictiveScaling` `policy_type` and `predictive_scaling_configuration` argument ([#19447](https://github.com/hashicorp/terraform-provider-aws/issues/19447))

BUG FIXES:

* resource/aws_networkfirewall_rule_group: Correctly update resource on `rules` change ([#19430](https://github.com/hashicorp/terraform-provider-aws/issues/19430))

## 3.41.0 (May 19, 2021)

FEATURES:

* **New Data Source:** `aws_cloudfront_function` ([#19315](https://github.com/hashicorp/terraform-provider-aws/issues/19315))
* **New Data Source:** `aws_glue_connection` ([#18802](https://github.com/hashicorp/terraform-provider-aws/issues/18802))
* **New Data Source:** `aws_glue_data_catalog_encryption_settings` ([#18802](https://github.com/hashicorp/terraform-provider-aws/issues/18802))
* **New Data Source:** `aws_organizations_delegated_administrators` ([#19389](https://github.com/hashicorp/terraform-provider-aws/issues/19389))
* **New Data Source:** `aws_organizations_delegated_services` ([#19389](https://github.com/hashicorp/terraform-provider-aws/issues/19389))
* **New Resource:** `aws_apprunner_auto_scaling_configuration_version` ([#19432](https://github.com/hashicorp/terraform-provider-aws/issues/19432))
* **New Resource:** `aws_apprunner_connection` ([#19432](https://github.com/hashicorp/terraform-provider-aws/issues/19432))
* **New Resource:** `aws_apprunner_custom_domain_association` ([#19432](https://github.com/hashicorp/terraform-provider-aws/issues/19432))
* **New Resource:** `aws_apprunner_service` ([#19432](https://github.com/hashicorp/terraform-provider-aws/issues/19432))
* **New Resource:** `aws_cloudfront_function` ([#19315](https://github.com/hashicorp/terraform-provider-aws/issues/19315))
* **New Resource:** `aws_macie2_invitation_accepter` ([#19304](https://github.com/hashicorp/terraform-provider-aws/issues/19304))
* **New Resource:** `aws_macie2_member` ([#19304](https://github.com/hashicorp/terraform-provider-aws/issues/19304))
* **New Resource:** `aws_macie2_organization_admin_account` ([#19303](https://github.com/hashicorp/terraform-provider-aws/issues/19303))
* **New Resource:** `aws_organizations_delegated_administrator` ([#19389](https://github.com/hashicorp/terraform-provider-aws/issues/19389))
* **New Resource:** `aws_servicecatalog_organizations_access` ([#19278](https://github.com/hashicorp/terraform-provider-aws/issues/19278))
* **New Resource:** `aws_servicecatalog_portfolio_share` ([#19278](https://github.com/hashicorp/terraform-provider-aws/issues/19278))

ENHANCEMENTS:

* data-source/aws_outposts_outpost: `owner_id` is now an optional argument ([#17585](https://github.com/hashicorp/terraform-provider-aws/issues/17585))
* data-source/aws_outposts_outposts: Add `owner_id` argument ([#17585](https://github.com/hashicorp/terraform-provider-aws/issues/17585))
* resource/aws_cloudfront_distribution: Add `function_association` argument to `ordered_cache_behavior` and `default_cache_behavior` configuration blocks ([#19315](https://github.com/hashicorp/terraform-provider-aws/issues/19315))
* resource/aws_glue_catalog_database: Add `target_database` argument ([#19371](https://github.com/hashicorp/terraform-provider-aws/issues/19371))
* resource/aws_glue_catalog_table: Add `target_table` argument ([#19372](https://github.com/hashicorp/terraform-provider-aws/issues/19372))
* resource/aws_launch_template: Add `interface_type` argument to `network_interfaces` configuration block ([#18841](https://github.com/hashicorp/terraform-provider-aws/issues/18841))
* resource/aws_network_interface: Add `interface_type` argument ([#18841](https://github.com/hashicorp/terraform-provider-aws/issues/18841))

BUG FIXES:

* resource/aws_lambda_function: Wait for successful completion of function code update ([#19386](https://github.com/hashicorp/terraform-provider-aws/issues/19386))
* resource/aws_pinpoint_email_channel: `role_arn` argument is optional ([#19361](https://github.com/hashicorp/terraform-provider-aws/issues/19361))

## 3.40.0 (May 13, 2021)

FEATURES:

* **New Resource:** `aws_macie2_custom_data_identifier` ([#19254](https://github.com/hashicorp/terraform-provider-aws/issues/19254))
* **New Resource:** `aws_macie2_findings_filter` ([#19283](https://github.com/hashicorp/terraform-provider-aws/issues/19283))
* **New Resource:** `aws_servicecatalog_tag_option` ([#19300](https://github.com/hashicorp/terraform-provider-aws/issues/19300))
* **New Resource:** `aws_timestreamwrite_database` ([#15463](https://github.com/hashicorp/terraform-provider-aws/issues/15463))
* **New Resource:** `aws_timestreamwrite_table` ([#19354](https://github.com/hashicorp/terraform-provider-aws/issues/19354))

ENHANCEMENTS:

* data-source/aws_codestarconnections_connection: Add `host_arn` attribute ([#19284](https://github.com/hashicorp/terraform-provider-aws/issues/19284))
* data-source/aws_lb_listener: Add `tags` attribute. ([#19286](https://github.com/hashicorp/terraform-provider-aws/issues/19286))
* resource/aws_ami_copy: Add `destination_outpost_arn` argument ([#17735](https://github.com/hashicorp/terraform-provider-aws/issues/17735))
* resource/aws_cloudwatch_event_target: Add `http_target` argument ([#19337](https://github.com/hashicorp/terraform-provider-aws/issues/19337))
* resource/aws_codestarconnections_connection: Add `host_arn` argument ([#19284](https://github.com/hashicorp/terraform-provider-aws/issues/19284))
* resource/aws_datasync_location_s3: Add `agent_arns` argument ([#18547](https://github.com/hashicorp/terraform-provider-aws/issues/18547))
* resource/aws_datasync_option: Add `private_link_endpoint`, `security_group_arns`, `subnet_arns` and `vpc_endpoint_id` arguments ([#16207](https://github.com/hashicorp/terraform-provider-aws/issues/16207))
* resource/aws_datasync_task: Add `excludes` argument and `overwrite_mode`, `task_queueing`, and `transfer_mode` to the `options` configuration block ([#16204](https://github.com/hashicorp/terraform-provider-aws/issues/16204))
* resource/aws_datasync_task: Add `schedule` argument ([#14452](https://github.com/hashicorp/terraform-provider-aws/issues/14452))
* resource/aws_datasync_task: Add plan time validation to `cloudwatch_log_group_arn`, `destination_location_arn` and `source_location_arn` ([#14452](https://github.com/hashicorp/terraform-provider-aws/issues/14452))
* resource/aws_eks_node_group: Add `node_group_name_prefix` argument ([#13938](https://github.com/hashicorp/terraform-provider-aws/issues/13938))
* resource/aws_lambda_event_source_mapping: Support reading `starting_position` and `starting_position_timestamp` attributes ([#19253](https://github.com/hashicorp/terraform-provider-aws/issues/19253))
* resource/aws_lb_listener: Add `tags` argument & `tags_all` attribute. ([#19286](https://github.com/hashicorp/terraform-provider-aws/issues/19286))
* resource/aws_lb_listener_rule: Add plan time validation to `listener_arn`, `action.target_group_arn`, `action.forward.target_group.arn`, `action.redirect.host`, `action.redirect.path`, `action.redirect.query`, `action.redirect.status_code`, `action.fixed_response.message_body`, `action.authenticate_cognito.user_pool_arn`. ([#19285](https://github.com/hashicorp/terraform-provider-aws/issues/19285))
* resource/aws_lb_listener_rule: Add tagging support. ([#19285](https://github.com/hashicorp/terraform-provider-aws/issues/19285))

## 3.39.0 (May 06, 2021)

FEATURES:

* **New Data Source:** `aws_cloudwatch_event_source` ([#19219](https://github.com/hashicorp/terraform-provider-aws/issues/19219))
* **New Resource:** `aws_dynamodb_kinesis_streaming_destination` ([#16743](https://github.com/hashicorp/terraform-provider-aws/issues/16743))
* **New Resource:** `aws_macie2_classification_job` ([#19165](https://github.com/hashicorp/terraform-provider-aws/issues/19165))

ENHANCEMENTS:

* data-source/aws_transfer_server: Add `certificate`, `endpoint_type`, `protocols` and `security_policy_name` attributes ([#13371](https://github.com/hashicorp/terraform-provider-aws/issues/13371))
* resource/aws_cloudwatch_event_bus: Support partner event bus creation ([#19072](https://github.com/hashicorp/terraform-provider-aws/issues/19072))
* resource/aws_cloudwatch_event_rule: Support partner event bus names ([#18491](https://github.com/hashicorp/terraform-provider-aws/issues/18491))
* resource/aws_cloudwatch_event_target: Support partner event bus names ([#18491](https://github.com/hashicorp/terraform-provider-aws/issues/18491))
* resource/aws_codebuild_project: Add `file_system_locations` argument ([#12130](https://github.com/hashicorp/terraform-provider-aws/issues/12130))
* resource/aws_cognito_identity_pool: Add allow_classic_flow argument ([#19176](https://github.com/hashicorp/terraform-provider-aws/issues/19176))
* resource/aws_datasync_location_s3: Add `s3_storage_class` argument ([#19190](https://github.com/hashicorp/terraform-provider-aws/issues/19190))
* resource/aws_glue_connection: Add plan time validation for `connection_properties`, `description`, `match_criteria`, `name`, and `physical_connection_requirements.security_group_id_list` ([#19172](https://github.com/hashicorp/terraform-provider-aws/issues/19172))
* resource/aws_msk_cluster: Support in-place `instance_type` updates ([#17447](https://github.com/hashicorp/terraform-provider-aws/issues/17447))
* resource/aws_sfn_state_machine: Add `tracing_configuration` attribute ([#15434](https://github.com/hashicorp/terraform-provider-aws/issues/15434))
* resource/aws_shield_protection: Add `tags` argument ([#19168](https://github.com/hashicorp/terraform-provider-aws/issues/19168))
* resource/aws_transfer_server: Add `protocols` argument ([#13371](https://github.com/hashicorp/terraform-provider-aws/issues/13371))
* resource/aws_transfer_server: Add `security_policy_name` argument ([#15375](https://github.com/hashicorp/terraform-provider-aws/issues/15375))

BUG FIXES:

* aws_batch_compute_environment: Allow update of just `service_role` for managed compute environments ([#19205](https://github.com/hashicorp/terraform-provider-aws/issues/19205))
* aws_batch_compute_environment: `service_role` argument is optional ([#19205](https://github.com/hashicorp/terraform-provider-aws/issues/19205))
* provider: Prevent `Provider produced inconsistent final plan` errors when lifecycle arguments apply to resource `tags` not known until apply ([#19251](https://github.com/hashicorp/terraform-provider-aws/issues/19251))
* resource/aws_appautoscaling_target: Ignore `ObjectNotFoundException` on deletion ([#18115](https://github.com/hashicorp/terraform-provider-aws/issues/18115))
* resource/aws_batch_job_definition: Prevent diff with default value of `fargatePlatformConfiguration` ([#19207](https://github.com/hashicorp/terraform-provider-aws/issues/19207))
* resource/aws_lakeformation_permissions: Fix issues related to permissions not being revoked and attempts to revoke non-existent permissions ([#18505](https://github.com/hashicorp/terraform-provider-aws/issues/18505))
* resource/aws_mwaa_environment: Correctly apply `plugins_s3_object_version` change ([#19266](https://github.com/hashicorp/terraform-provider-aws/issues/19266))
* resource/aws_sfn_state_machine: Handle eventual consistency of state machine updates ([#15434](https://github.com/hashicorp/terraform-provider-aws/issues/15434))
* resource/aws_ssoadmin_managed_policy_attachment: Retry attachment/detachment when other permission-set attachment event was not yet propagated, to avoid ConflictException. ([#19216](https://github.com/hashicorp/terraform-provider-aws/issues/19216))

## 3.38.0 (April 30, 2021)

NOTES:

* provider: `default_tags` support generally available to all provider resources that support `tags` with the exception of `aws_autoscaling_group` ([#19084](https://github.com/hashicorp/terraform-provider-aws/issues/19084))

FEATURES:

* **New Data Source:** `aws_cloudformation_type` ([#18579](https://github.com/hashicorp/terraform-provider-aws/issues/18579))
* **New Data Source:** `aws_kms_public_key` ([#18873](https://github.com/hashicorp/terraform-provider-aws/issues/18873))
* **New Data Source:** `aws_resourcegroupstaggingapi_resources` ([#17804](https://github.com/hashicorp/terraform-provider-aws/issues/17804))
* **New Resource:** `aws_cloudformation_type` ([#18579](https://github.com/hashicorp/terraform-provider-aws/issues/18579))
* **New Resource:** `aws_codestarconnections_host` ([#16918](https://github.com/hashicorp/terraform-provider-aws/issues/16918))
* **New Resource:** `aws_macie2_account` ([#19069](https://github.com/hashicorp/terraform-provider-aws/issues/19069))
* **New Resource:** `aws_rds_proxy_endpoint` ([#18881](https://github.com/hashicorp/terraform-provider-aws/issues/18881))
* **New Resource:** `aws_route53_resolver_firewall_rule` ([#18712](https://github.com/hashicorp/terraform-provider-aws/issues/18712))
* **New Resource:** `aws_route53_resolver_firewall_rule_group_association` ([#19164](https://github.com/hashicorp/terraform-provider-aws/issues/19164))
* **New Resource:** `aws_servicecatalog_product` ([#19122](https://github.com/hashicorp/terraform-provider-aws/issues/19122))

ENHANCEMENTS:

* data-source/aws_efs_mount_target: Add `access_point_id`, `file_system_id` arguments ([#18918](https://github.com/hashicorp/terraform-provider-aws/issues/18918))
* data-source/aws_iam_policy: Add support for lookup by `arn`, `name`, and/or `path_prefix` ([#6084](https://github.com/hashicorp/terraform-provider-aws/issues/6084))
* data-source/aws_launch_template: Add `placement` `host_resource_group_arn` attribute ([#15785](https://github.com/hashicorp/terraform-provider-aws/issues/15785))
* data/source_aws_eks_addon: added validation for `cluster_name` ([#19078](https://github.com/hashicorp/terraform-provider-aws/issues/19078))
* data/source_aws_eks_cluster: added validation for `cluster_name` ([#19078](https://github.com/hashicorp/terraform-provider-aws/issues/19078))
* resource/aws_appsync_resolver: Mark `request_template` and `response_template` as optional (support Lambda) ([#14710](https://github.com/hashicorp/terraform-provider-aws/issues/14710))
* resource/aws_batch_compute_environment: Additional supported value `FARGATE` and `FARGATE_SPOT` for the `type` argument in the `compute_resources` configuration block ([#16819](https://github.com/hashicorp/terraform-provider-aws/issues/16819))
* resource/aws_batch_compute_environment: The `instance_role`, `instance_type` and `min_vcpus` arguments in the `compute_resources` configuration block are now optional ([#16819](https://github.com/hashicorp/terraform-provider-aws/issues/16819))
* resource/aws_batch_compute_environment: The `security_group_ids` and `subnets` arguments in the `compute_resources` configuration block can now be updated in-place for Fargate compute resources ([#16819](https://github.com/hashicorp/terraform-provider-aws/issues/16819))
* resource/aws_batch_job_definition: Add `propagate_tags` argument ([#18336](https://github.com/hashicorp/terraform-provider-aws/issues/18336))
* resource/aws_codebuild_project: Add `build_batch_config` argument ([#14534](https://github.com/hashicorp/terraform-provider-aws/issues/14534))
* resource/aws_codebuild_project: Add `build_status_config` attribute to `source` and `secondary_sources` configuration blocks ([#15442](https://github.com/hashicorp/terraform-provider-aws/issues/15442))
* resource/aws_codebuild_project: Add `concurrent_build_limit` argument to specify build concurrency. ([#18320](https://github.com/hashicorp/terraform-provider-aws/issues/18320))
* resource/aws_codebuild_project: Add plan time validation for `secondary_artifacts`, `secondary_sources`, `service_role` ([#18843](https://github.com/hashicorp/terraform-provider-aws/issues/18843))
* resource/aws_eip: Add `address` argument to recover or an IPv4 address from an address pool, supporting BYOIP ([#8876](https://github.com/hashicorp/terraform-provider-aws/issues/8876))
* resource/aws_eks_addon: added validation for `cluster_name` ([#19078](https://github.com/hashicorp/terraform-provider-aws/issues/19078))
* resource/aws_eks_cluster: added validation for `name` ([#19078](https://github.com/hashicorp/terraform-provider-aws/issues/19078))
* resource/aws_eks_fargate_profile: added validation for `cluster_name` ([#19078](https://github.com/hashicorp/terraform-provider-aws/issues/19078))
* resource/aws_eks_node_group: added validation for `cluster_name` ([#19078](https://github.com/hashicorp/terraform-provider-aws/issues/19078))
* resource/aws_elasticache_global_replication_group: Adds parameter `engine_version_actual` to match other ElastiCache resources ([#18920](https://github.com/hashicorp/terraform-provider-aws/issues/18920))
* resource/aws_elasticache_subnet_group: Add `tags` argument ([#19119](https://github.com/hashicorp/terraform-provider-aws/issues/19119))
* resource/aws_instance: Make `instance_initiated_shutdown_behavior` also computed, allowing value to be read ([#18880](https://github.com/hashicorp/terraform-provider-aws/issues/18880))
* resource/aws_lambda_event_source_mapping: Don't incorrectly update unspecified `maximum_batching_window_in_seconds`, `maximum_record_age_in_seconds` and `maximum_retry_attempts` arguments from their default values ([#17933](https://github.com/hashicorp/terraform-provider-aws/issues/17933))
* resource/aws_lambda_event_source_mapping: Fix update of `batch_size` for MSK event source mappings ([#17933](https://github.com/hashicorp/terraform-provider-aws/issues/17933))
* resource/aws_launch_template: Add `placement` `host_resource_group_arn` argument ([#15785](https://github.com/hashicorp/terraform-provider-aws/issues/15785))
* resource/aws_organizations_organizational_unit: Add `tags` argument ([#18861](https://github.com/hashicorp/terraform-provider-aws/issues/18861))
* resource/aws_rds_global_cluster: Allow `engine_version` to be upgraded in place. ([#18598](https://github.com/hashicorp/terraform-provider-aws/issues/18598))
* resource/aws_s3outposts_endpoint: Extends creation timeout to 20 minutes ([#18454](https://github.com/hashicorp/terraform-provider-aws/issues/18454))
* resource/aws_ses_configuration_set: Adds `reputation_metrics_enabled` and `sending_enabled` arguments and `last_fresh_start` attribute ([#17608](https://github.com/hashicorp/terraform-provider-aws/issues/17608))
* resource/aws_ses_receipt_rule: Add `encoding` argument to `sns_action` configuration block. ([#17654](https://github.com/hashicorp/terraform-provider-aws/issues/17654))
* resource/aws_sns_topic_policy: Add `owner` attribute ([#14123](https://github.com/hashicorp/terraform-provider-aws/issues/14123))
* resource/aws_sns_topic_policy: Add plan time validation to `arn` ([#14123](https://github.com/hashicorp/terraform-provider-aws/issues/14123))
* resource/aws_wafv2_web_acl_logging_configuration: Add `logging_filter` argument ([#19051](https://github.com/hashicorp/terraform-provider-aws/issues/19051))

BUG FIXES:

* provider: Prevent `Provider produced inconsistent final plan` errors when resource `tags` are not known until apply ([#18958](https://github.com/hashicorp/terraform-provider-aws/issues/18958))
* resource/aws_batch_job_definition: Treat empty `container_properties.logConfiguration.secretOptions` array as `null` to prevent continual diffs ([#16120](https://github.com/hashicorp/terraform-provider-aws/issues/16120))
* resource/aws_batch_job_queue: Recreate batch job queue if the `name` changes ([#19121](https://github.com/hashicorp/terraform-provider-aws/issues/19121))
* resource/aws_codebuild_project: Allow fetching submodules for bitbucket source types ([#18843](https://github.com/hashicorp/terraform-provider-aws/issues/18843))
* resource/aws_codebuild_project: Fix removing `secondary_sources` and `secondary_artifacts` ([#18843](https://github.com/hashicorp/terraform-provider-aws/issues/18843))
* resource/aws_ec2_managed_prefix_list: Prevent `entry` `description` update errors ([#19095](https://github.com/hashicorp/terraform-provider-aws/issues/19095))
* resource/aws_elasticache_cluster: Allows specifying Redis 6.x ([#18920](https://github.com/hashicorp/terraform-provider-aws/issues/18920))
* resource/aws_elasticache_replication_group: Allows specifying Redis 6.x ([#18920](https://github.com/hashicorp/terraform-provider-aws/issues/18920))
* resource/aws_glue_crawler: Allow '/' in `name` argument ([#19160](https://github.com/hashicorp/terraform-provider-aws/issues/19160))
* resource/aws_lambda_event_source_mapping: Support -1 (forever) as a valid value for `maximum_record_age_in_seconds` ([#16113](https://github.com/hashicorp/terraform-provider-aws/issues/16113))
* resource/aws_lambda_event_source_mapping: Support -1 (forever) as a valid value for `maximum_retry_attempts` ([#16113](https://github.com/hashicorp/terraform-provider-aws/issues/16113))
* resource/aws_ram_principal_association: Improve handling of eventual consistency ([#17032](https://github.com/hashicorp/terraform-provider-aws/issues/17032))
* resource/aws_ram_resource_share: Improve handling of eventual consistency ([#17032](https://github.com/hashicorp/terraform-provider-aws/issues/17032))
* resource/aws_ram_resource_share_accepter: Improve handling of eventual consistency ([#17032](https://github.com/hashicorp/terraform-provider-aws/issues/17032))
* resource/aws_storagegateway_gateway: Correctly handle additional error message returned in some regions ([#19116](https://github.com/hashicorp/terraform-provider-aws/issues/19116))
* resource/aws_vpc_endpoint: Fix auto_accept failing while waiting for the VPC Endpoint Connection acceptance ([#19059](https://github.com/hashicorp/terraform-provider-aws/issues/19059))
* resource/aws_vpn_connection: Prevent flipped `tunnel1_*` and `tunnel2_*` ordering when `tunnel1_inside_cidr`, `tunnel1_inside_ipv6_cidr`, or `tunnel1_preshared_key` is configured ([#19077](https://github.com/hashicorp/terraform-provider-aws/issues/19077))

## 3.37.0 (April 16, 2021)

NOTES:

* provider: The HTTP User-Agent header has been reordered so the AWS SDK Go product is last, except when using the TF_APPEND_USER_AGENT environment variable. Environments dependent on the previous User-Agent header ordering may require updates. ([#18855](https://github.com/hashicorp/terraform-provider-aws/issues/18855))

FEATURES:

* **New Data Source:** `aws_eks_addon` ([#16972](https://github.com/hashicorp/terraform-provider-aws/issues/16972))
* **New Resource:** `aws_eks_addon` ([#16972](https://github.com/hashicorp/terraform-provider-aws/issues/16972))
* **New Resource:** `aws_route53_resolver_firewall_domain_list` ([#18558](https://github.com/hashicorp/terraform-provider-aws/issues/18558))
* **New Resource:** `aws_securityhub_insight` ([#18494](https://github.com/hashicorp/terraform-provider-aws/issues/18494))

ENHANCEMENTS:

* resource/aws_autoscaling_group: Add Warm Pool support ([#18734](https://github.com/hashicorp/terraform-provider-aws/issues/18734))
* resource/aws_cloudfront_distribution: Add `trusted_key_groups` argument ([#18644](https://github.com/hashicorp/terraform-provider-aws/issues/18644))
* resource/aws_codedeploy_app: Add `arn`, `linked_to_github`, `github_account_name`, `application_id` attributes ([#18564](https://github.com/hashicorp/terraform-provider-aws/issues/18564))
* resource/aws_codedeploy_app: Add `tags` argument ([#18564](https://github.com/hashicorp/terraform-provider-aws/issues/18564))
* resource/aws_codedeploy_app: Add plan time validation for `name` ([#18564](https://github.com/hashicorp/terraform-provider-aws/issues/18564))
* resource/aws_codedeploy_deployment_group: Add `arn`, `compute_platform`, and `deployment_group_id` attributes ([#18716](https://github.com/hashicorp/terraform-provider-aws/issues/18716))
* resource/aws_codedeploy_deployment_group: Add `tags` argument ([#18716](https://github.com/hashicorp/terraform-provider-aws/issues/18716))
* resource/aws_codedeploy_deployment_group: Add plan time validation for `terminate_blue_instances_on_deployment_success.termination_wait_time_in_minutes`, `service_role_arn`, `load_balancer_info.target_group_pair_info.prod_traffic_route.listener_arns`, `load_balancer_info.target_group_pair_info.test_traffic_route.listener_arns`, `trigger_configuration.trigger_target_arn` ([#18716](https://github.com/hashicorp/terraform-provider-aws/issues/18716))
* resource/aws_codedeploy_deployment_group: Updating `deployment_group_name` doesnt recreate group ([#18716](https://github.com/hashicorp/terraform-provider-aws/issues/18716))
* resource/aws_dynamodb_table: Add `kms_key_arn` argument to `replica` configuration block ([#18373](https://github.com/hashicorp/terraform-provider-aws/issues/18373))
* resource/aws_emr_cluster: Adds support for multiple subnets ([#17219](https://github.com/hashicorp/terraform-provider-aws/issues/17219))
* resource/aws_rds_cluster: Database port is updated in-place ([#18081](https://github.com/hashicorp/terraform-provider-aws/issues/18081))
* resource/aws_servicequotas_service_quota: Add plan time validation to `quota_code` and `service_code` ([#17992](https://github.com/hashicorp/terraform-provider-aws/issues/17992))
* resource/aws_sns_topic: Add `fifo_topic` and `content_based_deduplication` attributes ([#15828](https://github.com/hashicorp/terraform-provider-aws/issues/15828))

BUG FIXES:

* resource/aws_dynamodb_table: Update Global Secondary Index provisioned throughput settings on new changes ([#18215](https://github.com/hashicorp/terraform-provider-aws/issues/18215))
* resource/aws_ecr_replication_configuration: Remove relication rules on resource deletion ([#18882](https://github.com/hashicorp/terraform-provider-aws/issues/18882))
* resource/aws_eip: Tags are created for EIPs which default to vpc domain ([#18909](https://github.com/hashicorp/terraform-provider-aws/issues/18909))
* resource/aws_fms_policy: Use API model regular expression for `resource_type` and `resource_type_list` argument plan time validation ([#18600](https://github.com/hashicorp/terraform-provider-aws/issues/18600))
* resource/aws_sqs_queue: Append `.fifo` suffix for Terraform-assigned FIFO queue names ([#17164](https://github.com/hashicorp/terraform-provider-aws/issues/17164))

## 3.36.0 (April 09, 2021)

FEATURES:

* **New Resource:** `aws_cloudfront_key_group` ([#17041](https://github.com/hashicorp/terraform-provider-aws/issues/17041))
* **New Resource:** `aws_ecr_registry_policy` ([#16831](https://github.com/hashicorp/terraform-provider-aws/issues/16831))
* **New Resource:** `aws_ecr_replication_configuration` ([#16853](https://github.com/hashicorp/terraform-provider-aws/issues/16853))
* **New Resource:** `aws_kinesisanalyticsv2_application_snapshot` ([#18056](https://github.com/hashicorp/terraform-provider-aws/issues/18056))
* **New Resource:** `aws_mwaa_environment` ([#16616](https://github.com/hashicorp/terraform-provider-aws/issues/16616))

ENHANCEMENTS:

* data-source/aws_lb_listener: Add `alpn_policy` argument ([#14462](https://github.com/hashicorp/terraform-provider-aws/issues/14462))
* data-source/aws_s3_bucket_object: Add `bucket_key_enabled` attribute (Support S3 Bucket Keys) ([#16581](https://github.com/hashicorp/terraform-provider-aws/issues/16581))
* resource/aws_eip: Tags are set on create ([#17612](https://github.com/hashicorp/terraform-provider-aws/issues/17612))
* resource/aws_kinesisanalyticsv2_application: Add `force_stop` attribute ([#18056](https://github.com/hashicorp/terraform-provider-aws/issues/18056))
* resource/aws_kinesisanalyticsv2_application: Add `run_configuration` attribute for starting a Flink application ([#18056](https://github.com/hashicorp/terraform-provider-aws/issues/18056))
* resource/aws_kinesisanalyticsv2_application: Add `start_application` attribute ([#18056](https://github.com/hashicorp/terraform-provider-aws/issues/18056))
* resource/aws_kinesisanalyticsv2_application: `starting_position_configuration` can be specified when starting a SQL application ([#18056](https://github.com/hashicorp/terraform-provider-aws/issues/18056))
* resource/aws_lb_listener: Add `alpn_policy` argument ([#14462](https://github.com/hashicorp/terraform-provider-aws/issues/14462))
* resource/aws_s3_bucket: Add `bucket_key_enabled` argument to `server_side_encryption_configuration` `rule` configuration block (Support S3 Bucket Keys) ([#16581](https://github.com/hashicorp/terraform-provider-aws/issues/16581))
* resource/aws_s3_bucket_object: Add `bucket_key_enabled` attribute (Support S3 Bucket Keys) ([#16581](https://github.com/hashicorp/terraform-provider-aws/issues/16581))
* resource/aws_s3_object_copy: Add `bucket_key_enabled` argument ([#18611](https://github.com/hashicorp/terraform-provider-aws/issues/18611))

BUG FIXES:

* resource/aws_appmesh_gateway_route: Handle read-after-create eventual consistency ([#18529](https://github.com/hashicorp/terraform-provider-aws/issues/18529))
* resource/aws_appmesh_mesh: Handle read-after-create eventual consistency ([#18529](https://github.com/hashicorp/terraform-provider-aws/issues/18529))
* resource/aws_appmesh_route: Handle read-after-create eventual consistency ([#18529](https://github.com/hashicorp/terraform-provider-aws/issues/18529))
* resource/aws_appmesh_virtual_gateway: Handle read-after-create eventual consistency ([#18529](https://github.com/hashicorp/terraform-provider-aws/issues/18529))
* resource/aws_appmesh_virtual_node: Handle read-after-create eventual consistency ([#18529](https://github.com/hashicorp/terraform-provider-aws/issues/18529))
* resource/aws_appmesh_virtual_router: Handle read-after-create eventual consistency ([#18529](https://github.com/hashicorp/terraform-provider-aws/issues/18529))
* resource/aws_appmesh_virtual_service: Handle read-after-create eventual consistency ([#18529](https://github.com/hashicorp/terraform-provider-aws/issues/18529))
* resource/aws_cloudhsm_v2_hsm: Prevent orphaned HSM Instances by additionally matching on ENI identifier during lookup ([#18580](https://github.com/hashicorp/terraform-provider-aws/issues/18580))
* resource/aws_dms_replication_task: Handle read-only attributes in `replication_task_settings` to avoid unnecessary diffs. ([#13476](https://github.com/hashicorp/terraform-provider-aws/issues/13476))
* resource/aws_docdb_cluster_parameter_group: Read all user parameters and parameters specified in the configuration. ([#18486](https://github.com/hashicorp/terraform-provider-aws/issues/18486))
* resource/aws_ecr_lifecycle_policy: Handle read-after-create eventual consistency ([#18464](https://github.com/hashicorp/terraform-provider-aws/issues/18464))
* resource/aws_ecr_repository: Handle read-after-create eventual consistency ([#18464](https://github.com/hashicorp/terraform-provider-aws/issues/18464))
* resource/aws_ecr_repository_policy: Handle read-after-create eventual consistency ([#18464](https://github.com/hashicorp/terraform-provider-aws/issues/18464))
* resource/aws_elasticache_replication_group: Remmoves incorrect plan-time validation for `automatic_failover_enabled` ([#18635](https://github.com/hashicorp/terraform-provider-aws/issues/18635))
* resource/aws_iam_group: Handle read-after-create eventual consistency ([#18459](https://github.com/hashicorp/terraform-provider-aws/issues/18459))
* resource/aws_iam_group_membership: Handle read-after-create eventual consistency ([#18459](https://github.com/hashicorp/terraform-provider-aws/issues/18459))
* resource/aws_iam_group_policy: Handle read-after-create eventual consistency ([#18459](https://github.com/hashicorp/terraform-provider-aws/issues/18459))
* resource/aws_iam_group_policy_attachment: Handle read-after-create eventual consistency ([#18459](https://github.com/hashicorp/terraform-provider-aws/issues/18459))
* resource/aws_iam_user: Handle read-after-create eventual consistency ([#18458](https://github.com/hashicorp/terraform-provider-aws/issues/18458))
* resource/aws_iam_user_group_membership: Handle read-after-create eventual consistency ([#18458](https://github.com/hashicorp/terraform-provider-aws/issues/18458))
* resource/aws_iam_user_login_profile: Handle read-after-create eventual consistency ([#18458](https://github.com/hashicorp/terraform-provider-aws/issues/18458))
* resource/aws_iam_user_policy: Handle read-after-create eventual consistency ([#18458](https://github.com/hashicorp/terraform-provider-aws/issues/18458))
* resource/aws_iam_user_policy_attachment: Handle read-after-create eventual consistency ([#18458](https://github.com/hashicorp/terraform-provider-aws/issues/18458))
* resource/aws_iam_user_ssh_key: Handle read-after-create eventual consistency ([#18458](https://github.com/hashicorp/terraform-provider-aws/issues/18458))
* resource/aws_lb_target_group: Handle read-after-create eventual consistency ([#18634](https://github.com/hashicorp/terraform-provider-aws/issues/18634))
* resource/aws_secretsmanager_secret: Handle read-after-create eventual consistency ([#18462](https://github.com/hashicorp/terraform-provider-aws/issues/18462))
* resource/aws_secretsmanager_secret_policy: Handle read-after-create eventual consistency ([#18462](https://github.com/hashicorp/terraform-provider-aws/issues/18462))
* resource/aws_secretsmanager_secret_rotation: Handle read-after-create eventual consistency ([#18462](https://github.com/hashicorp/terraform-provider-aws/issues/18462))
* resource/aws_secretsmanager_secret_version: Handle read-after-create eventual consistency ([#18462](https://github.com/hashicorp/terraform-provider-aws/issues/18462))
* resource/aws_ssm_parameter: Allow `allowed_pattern` and `description` arguments to be empty strings ([#18588](https://github.com/hashicorp/terraform-provider-aws/issues/18588))
* resource/aws_ssm_parameter: Allow `tags` to be applied to resource when `overwrite` is configured ([#18640](https://github.com/hashicorp/terraform-provider-aws/issues/18640))
* resource/aws_vpc_endpoint_route_table_association: Handle read-after-create eventual consistency ([#18465](https://github.com/hashicorp/terraform-provider-aws/issues/18465))
* resource/aws_xray_sampling_rule: Change the maximum length of `rule_name` from 128 to 32 ([#18667](https://github.com/hashicorp/terraform-provider-aws/issues/18667))

## 3.35.0 (April 01, 2021)

FEATURES:

* **New Resource:** `aws_cloudwatch_query_definition` ([#17899](https://github.com/hashicorp/terraform-provider-aws/issues/17899))

ENHANCEMENTS:

* data-source/aws_efs_file_system: Add `availability_zone_id` and `availability_zone_name` attributes ([#18319](https://github.com/hashicorp/terraform-provider-aws/issues/18319))
* data-source/aws_iam_policy: Add `policy_id` and `tags` attributes ([#18276](https://github.com/hashicorp/terraform-provider-aws/issues/18276))
* resource/aws_apigatewayv2_route: Add `request_parameter` attribute ([#18410](https://github.com/hashicorp/terraform-provider-aws/issues/18410))
* resource/aws_appmesh_virtual_gateway: Add `spec.backend_defaults.client_policy.tls.certificate`, `spec.backend_defaults.client_policy.tls.validation.subject_alternative_names`, `spec.listener.tls.certificate` and `spec.listener.tls.validation.subject_alternative_names` attributes to support mutual TLS authentication ([#18106](https://github.com/hashicorp/terraform-provider-aws/issues/18106))
* resource/aws_appmesh_virtual_gateway: Add `spec.backend_defaults.client_policy.tls.validation.trust.sds` and `spec.listener.tls.validation.trust.sds` attributes to support Envoy Service Discovery Service certificates ([#18106](https://github.com/hashicorp/terraform-provider-aws/issues/18106))
* resource/aws_appmesh_virtual_node: Add `spec.backend.virtual_service.client_policy.tls.certificate`, `spec.backend.virtual_service.client_policy.tls.validation.subject_alternative_names`, `spec.backend_defaults.client_policy.tls.certificate`, `spec.backend_defaults.client_policy.tls.validation.subject_alternative_names`, `spec.listener.tls.certificate` and `spec.listener.tls.validation.subject_alternative_names` attributes to support mutual TLS authentication ([#18127](https://github.com/hashicorp/terraform-provider-aws/issues/18127))
* resource/aws_appmesh_virtual_node: Add `spec.backend.virtual_service.client_policy.tls.validation.trust.sds`, `spec.backend_defaults.client_policy.tls.validation.trust.sds` and `spec.listener.tls.validation.trust.sds` attributes to support Envoy Service Discovery Service certificates ([#18127](https://github.com/hashicorp/terraform-provider-aws/issues/18127))
* resource/aws_backup_plan: Add `enable_continuous_backup` argument ([#18315](https://github.com/hashicorp/terraform-provider-aws/issues/18315))
* resource/aws_cloudformation_stack_set: Add `auto_deployment` configuration block and `permissions_model` arguments (support service managed permissions) ([#12423](https://github.com/hashicorp/terraform-provider-aws/issues/12423))
* resource/aws_cognito_user_pool: Allow `schema` items to be added without recreating resource. ([#18512](https://github.com/hashicorp/terraform-provider-aws/issues/18512))
* resource/aws_ecs_service: Add `deployment_circuit_breaker` ([#16936](https://github.com/hashicorp/terraform-provider-aws/issues/16936))
* resource/aws_efs_file_system: Add `availability_zone_id` attribute and `availability_zone_name` argument ([#18319](https://github.com/hashicorp/terraform-provider-aws/issues/18319))
* resource/aws_efs_file_system: Add `number_of_mount_targets`, `size_in_bytes` and `owner_id` attributes ([#17969](https://github.com/hashicorp/terraform-provider-aws/issues/17969))
* resource/aws_elasticsearch_domain: Add `domain_endpoint_options` configuration block `custom_endpoint`, `custom_endpoint_certificate_arn`, and `custom_endpoint_enabled` arguments ([#16192](https://github.com/hashicorp/terraform-provider-aws/issues/16192))
* resource/aws_iam_policy: Add `policy_id` attribute ([#18276](https://github.com/hashicorp/terraform-provider-aws/issues/18276))
* resource/aws_iam_policy: Add tagging support ([#18276](https://github.com/hashicorp/terraform-provider-aws/issues/18276))
* resource/aws_lb_target_group: Add preserve_client_ip target attribute support ([#17731](https://github.com/hashicorp/terraform-provider-aws/issues/17731))
* resource/aws_route: `destination_prefix_list_id` attribute can be specified for managed prefix list destinations ([#17291](https://github.com/hashicorp/terraform-provider-aws/issues/17291))
* resource/aws_ssm_parameter: Add plan time validation to `name`, `description` and `allowed_pattern` ([#17830](https://github.com/hashicorp/terraform-provider-aws/issues/17830))
* resource/aws_ssm_parameter: Tag on create ([#17830](https://github.com/hashicorp/terraform-provider-aws/issues/17830))

BUG FIXES:

* resource/aws_ec2_transit_gateway_route_table_propagation: Wait for enable and disable operations to complete ([#18470](https://github.com/hashicorp/terraform-provider-aws/issues/18470))
* resource/aws_ecs_service: Improve handling of eventual consistency including security group dependency violations on deletion ([#16936](https://github.com/hashicorp/terraform-provider-aws/issues/16936))
* resource/aws_iam_role: Handle read-after-create eventual consistency ([#18435](https://github.com/hashicorp/terraform-provider-aws/issues/18435))
* resource/aws_iam_role_policy: Handle read-after-create eventual consistency ([#18435](https://github.com/hashicorp/terraform-provider-aws/issues/18435))
* resource/aws_iam_role_policy_attachment: Handle read-after-create eventual consistency ([#18435](https://github.com/hashicorp/terraform-provider-aws/issues/18435))
* resource/aws_network_interface_sg_attachment: Handle read-after-create eventual consistency ([#18466](https://github.com/hashicorp/terraform-provider-aws/issues/18466))
* resource/aws_route_table: Improve eventual consistency handling and handling of out-of-band resource removal ([#17319](https://github.com/hashicorp/terraform-provider-aws/issues/17319))
* resource/aws_route_table_association: Improve eventual consistency handling and handling of out-of-band resource removal ([#17319](https://github.com/hashicorp/terraform-provider-aws/issues/17319))
* resource/aws_s3_bucket_object: Handle read-after-create eventual consistency ([#17236](https://github.com/hashicorp/terraform-provider-aws/issues/17236))
* resource/aws_securityhub_organization_admin_account: Retry on `ResourceConflictException` error during creation ([#18341](https://github.com/hashicorp/terraform-provider-aws/issues/18341))
* resource/aws_sns_topic_subscription: Enforce lowercase `protocol` argument validation to match API and prevent resource errors ([#18475](https://github.com/hashicorp/terraform-provider-aws/issues/18475))
* resource/aws_sns_topic_subscription: Handle read-after-create eventual consistency ([#18475](https://github.com/hashicorp/terraform-provider-aws/issues/18475))
* resource/aws_spot_instance_request: Handle read-after-create eventual consistency ([#18473](https://github.com/hashicorp/terraform-provider-aws/issues/18473))
* resource/aws_synthetics_canary: Handle asynchronous IAM eventual consistency error on creation ([#18404](https://github.com/hashicorp/terraform-provider-aws/issues/18404))
* resource/aws_vpc_dhcp_options_association: Handle read-after-create eventual consistency ([#18472](https://github.com/hashicorp/terraform-provider-aws/issues/18472))
* resource/aws_vpn_gateway_route_propagation: Improve eventual consistency handling and handling of out-of-band resource removal ([#17319](https://github.com/hashicorp/terraform-provider-aws/issues/17319))

## 3.34.0 (March 26, 2021)

NOTES:

* resource/aws_storagegateway_upload_buffer: The Storage Gateway `ListLocalDisks` API operation has been implemented to support the `disk_path` attribute for Cached and VTL gateway types. Environments using restrictive IAM permissions may require updates. ([#18313](https://github.com/hashicorp/terraform-provider-aws/issues/18313))

FEATURES:

* **New Data Source:** `aws_codestarconnections_connection` ([#18129](https://github.com/hashicorp/terraform-provider-aws/issues/18129))
* **New Resource:** `aws_lightsail_instance_public_ports` ([#8611](https://github.com/hashicorp/terraform-provider-aws/issues/8611))

ENHANCEMENTS:

* resource/aws_ami_from_instance: Tag on create. ([#17968](https://github.com/hashicorp/terraform-provider-aws/issues/17968))
* resource/aws_ecr_repository_policy: Add plan time validation for `policy` ([#14193](https://github.com/hashicorp/terraform-provider-aws/issues/14193))
* resource/aws_fms_admin_account: Extend creation timeout to 10 minutes ([#17596](https://github.com/hashicorp/terraform-provider-aws/issues/17596))
* resource/aws_iam_instance_profile: Add tagging support ([#17962](https://github.com/hashicorp/terraform-provider-aws/issues/17962))
* resource/aws_iam_openid_connect_provider: Add plan time validation for `client_id_list` and `thumbprint_list` ([#17964](https://github.com/hashicorp/terraform-provider-aws/issues/17964))
* resource/aws_iam_openid_connect_provider: Add tagging support ([#17964](https://github.com/hashicorp/terraform-provider-aws/issues/17964))
* resource/aws_iam_saml_provider: Add plan time validation for `name` and `saml_metadata_document` ([#17965](https://github.com/hashicorp/terraform-provider-aws/issues/17965))
* resource/aws_iam_saml_provider: Add tagging support ([#17965](https://github.com/hashicorp/terraform-provider-aws/issues/17965))
* resource/aws_iam_server_certificate: Add `expiration` and `upload_date` attributes ([#17967](https://github.com/hashicorp/terraform-provider-aws/issues/17967))
* resource/aws_iam_server_certificate: Add tagging support ([#17967](https://github.com/hashicorp/terraform-provider-aws/issues/17967))
* resource/aws_light_instance_public_ports: Add `cidrs` argument to `port_info` ([#14905](https://github.com/hashicorp/terraform-provider-aws/issues/14905))
* resource/aws_pinpoint_email_channel: Add `configuration_set` argument ([#18314](https://github.com/hashicorp/terraform-provider-aws/issues/18314))
* resource/aws_pinpoint_email_channel: Add plan time validation for `identity` and `role_arn` ([#18314](https://github.com/hashicorp/terraform-provider-aws/issues/18314))
* resource/aws_pinpoint_event_stream: Plan time validations for `destination_stream_arn` and `role_arn` ([#18305](https://github.com/hashicorp/terraform-provider-aws/issues/18305))
* resource/aws_route: Validate route destination and target attributes ([#16930](https://github.com/hashicorp/terraform-provider-aws/issues/16930))
* resource/aws_sns_topic_subscription: Add plan time validation for `subscription_role_arn` and `topic_arn` ([#14101](https://github.com/hashicorp/terraform-provider-aws/issues/14101))
* resource/aws_storagegateway_upload_buffer: Add `disk_path` argument for Cached and VTL gateways ([#18313](https://github.com/hashicorp/terraform-provider-aws/issues/18313))

BUG FIXES:

* data-source/aws_storagegateway_local_disk: Allow `disk_path` reference on `disk_node` lookup and vice-versa ([#18313](https://github.com/hashicorp/terraform-provider-aws/issues/18313))
* resource/aws_api_gateway_vpc_link: Persist ID of failed VPC Link to state ([#18382](https://github.com/hashicorp/terraform-provider-aws/issues/18382))
* resource/aws_apigatewayv2_domain_name: Allow update of mutual TLS S3 object version ([#18351](https://github.com/hashicorp/terraform-provider-aws/issues/18351))
* resource/aws_cloudfront_distribution: Allow `forwarded_values` to be set to empty when values were previously set ([#18042](https://github.com/hashicorp/terraform-provider-aws/issues/18042))
* resource/aws_cloudwatch_event_permission: Fix error in Event Bridge/CloudWatch Events bus name validation ([#16815](https://github.com/hashicorp/terraform-provider-aws/issues/16815))
* resource/aws_cloudwatch_event_rule: Fix error in Event Bridge/CloudWatch Events bus name validation ([#16815](https://github.com/hashicorp/terraform-provider-aws/issues/16815))
* resource/aws_cloudwatch_event_target: Fix error in Event Bridge/CloudWatch Events bus name validation ([#16815](https://github.com/hashicorp/terraform-provider-aws/issues/16815))
* resource/aws_config_configuration_aggregator: Allow name to have uppercase characters ([#14247](https://github.com/hashicorp/terraform-provider-aws/issues/14247))
* resource/aws_ecs_service: Re-create service when `service_registries` changes ([#17387](https://github.com/hashicorp/terraform-provider-aws/issues/17387))
* resource/aws_elasticache_replication_group: Prevents re-creation of secondary replication groups when encryption is enabled ([#18361](https://github.com/hashicorp/terraform-provider-aws/issues/18361))
* resource/aws_mq_configuration: Add `ldap` as an `authentication_strategy` and `RabbitMQ` as an `engine_type` ([#18070](https://github.com/hashicorp/terraform-provider-aws/issues/18070))
* resource/aws_network_acl: Handle EC2 eventual consistency errors on creation ([#18388](https://github.com/hashicorp/terraform-provider-aws/issues/18388))
* resource/aws_network_acl_rule: Handle EC2 eventual consistency errors on creation ([#18388](https://github.com/hashicorp/terraform-provider-aws/issues/18388))
* resource/aws_pinpoint_event_stream: Retry on eventual consistency error ([#18305](https://github.com/hashicorp/terraform-provider-aws/issues/18305))
* resource/aws_pinpoint_sms_channel: Set all params on update ([#18281](https://github.com/hashicorp/terraform-provider-aws/issues/18281))
* resource/aws_route: Correctly handle updates to the route target attributes (`egress_only_gateway_id`, `gateway_id`, `instance_id`,  `local_gateway_id`, `nat_gateway_id`, `network_interface_id`, `transit_gateway_id`, `vpc_peering_connection_id`) ([#16930](https://github.com/hashicorp/terraform-provider-aws/issues/16930))
* resource/aws_sns_topic_subscription: recreate subscription if topic is deleted ([#14101](https://github.com/hashicorp/terraform-provider-aws/issues/14101))
* resource/aws_subnet: Handle EC2 eventual consistency errors on creation ([#18392](https://github.com/hashicorp/terraform-provider-aws/issues/18392))
* resource/aws_vpc: Handle EC2 eventual consistency errors on creation ([#18391](https://github.com/hashicorp/terraform-provider-aws/issues/18391))
* resource/aws_wafv2_web_acl_logging_configuration: Remove deprecation warning for `redacted_fields` `single_header` argument ([#18384](https://github.com/hashicorp/terraform-provider-aws/issues/18384))

## 3.33.0 (March 18, 2021)

NOTES:

* data-source/aws_vpc_endpoint_service: The `service_type` argument filtering has been switched from client-side to new EC2 API functionality ([#17641](https://github.com/hashicorp/terraform-provider-aws/issues/17641))
* provider: New `default_tags` argument as a public preview for applying tags across all resources under a provider. Support for the functionality must be added to individual resources in the codebase and is only implemented for the `aws_subnet` and `aws_vpc` resources at this time. Until a general availability announcement, no compatibility promises are made with these provider arguments and their functionality. ([#17974](https://github.com/hashicorp/terraform-provider-aws/issues/17974))
* resource/aws_codebuild_project: The `source` and `secondary_sources` configuration block `auth` attributes have been deprecated to match the CodeBuild API documentation. Use the `aws_codebuild_source_credential` resource instead. ([#17465](https://github.com/hashicorp/terraform-provider-aws/issues/17465))
* resource/aws_wafv2_web_acl_logging_configuration: The `redacted_fields` configuration block `all_query_arguments`, `body`, and `single_query_argument` arguments have been deprecated to match the WAF API documentation ([#14319](https://github.com/hashicorp/terraform-provider-aws/issues/14319))

FEATURES:

* **New Data Source:** `aws_ec2_transit_gateway_route_tables` ([#17589](https://github.com/hashicorp/terraform-provider-aws/issues/17589))
* **New Data Source:** `aws_kinesis_stream_consumer` ([#17149](https://github.com/hashicorp/terraform-provider-aws/issues/17149))
* **New Resource:** `aws_kinesis_stream_consumer` ([#17149](https://github.com/hashicorp/terraform-provider-aws/issues/17149))

ENHANCEMENTS:

* provider: Add `default_tags` argument (in public preview, see note above) ([#17974](https://github.com/hashicorp/terraform-provider-aws/issues/17974))
* resource/aws_db_parameter_group: Store all values in lowercase to prevent unexpected diffs ([#17909](https://github.com/hashicorp/terraform-provider-aws/issues/17909))
* resource/aws_ssm_parameter: Add support for `Intelligent-Tiering` ([#11967](https://github.com/hashicorp/terraform-provider-aws/issues/11967))
* resource/aws_storagegateway_gateway: Add support for `smb_file_share_visibility`. ([#18076](https://github.com/hashicorp/terraform-provider-aws/issues/18076))
* resource/aws_subnet: Support provider-wide default tags (in public preview, see note above) ([#17974](https://github.com/hashicorp/terraform-provider-aws/issues/17974))
* resource/aws_vpc: Support provider-wide default tags (in public preview, see note above) ([#17974](https://github.com/hashicorp/terraform-provider-aws/issues/17974))

BUG FIXES:

* data-source/aws_vpc_endpoint_service: Prevent panic with incorrect `service_type` argument values ([#17641](https://github.com/hashicorp/terraform-provider-aws/issues/17641))
* resource/aws_dms_certificate: Correctly base64 decode `certificate_wallet` value ([#17958](https://github.com/hashicorp/terraform-provider-aws/issues/17958))
* resource/aws_globalaccelerator_accelerator: Correct length for `name` attribute validation ([#17985](https://github.com/hashicorp/terraform-provider-aws/issues/17985))
* resource/aws_lakeformation_permissions: Properly serialize SELECT permission for `permissions` and `permissions_with_grant_option` fields ([#18203](https://github.com/hashicorp/terraform-provider-aws/issues/18203))
* resource/aws_ssm_patch_group: Allow for a single patch group to be registered with multiple patch baselines ([#15213](https://github.com/hashicorp/terraform-provider-aws/issues/15213))
* resource/aws_ssm_patch_group: Replace `Provider produced inconsistent result after apply` with actual error message ([#15213](https://github.com/hashicorp/terraform-provider-aws/issues/15213))
* resource/aws_waf_rule: Fix rule deletion when still referenced by a WebACL ([#17876](https://github.com/hashicorp/terraform-provider-aws/issues/17876))
* resource/aws_wafv2_web_acl_logging_configuration: Ensure `redacted_fields` are applied to the resource ([#14319](https://github.com/hashicorp/terraform-provider-aws/issues/14319))

## 3.32.0 (March 12, 2021)

FEATURES:

* **New Data Source:** `aws_acmpca_certificate` ([#10213](https://github.com/hashicorp/terraform-provider-aws/issues/10213))
* **New Resource:** `aws_acmpca_certificate` ([#10213](https://github.com/hashicorp/terraform-provider-aws/issues/10213))
* **New Resource:** `aws_acmpca_certificate_authority_certificate` ([#17850](https://github.com/hashicorp/terraform-provider-aws/issues/17850))

ENHANCEMENTS:

* resource/aws_appautoscaling_scheduled_action: Adds `timezone` support ([#17689](https://github.com/hashicorp/terraform-provider-aws/issues/17689))
* resource/aws_appautoscaling_scheduled_action: Allows any timezone to be specified for `start_time` and `end_time` ([#17689](https://github.com/hashicorp/terraform-provider-aws/issues/17689))
* resource/aws_appautoscaling_scheduled_action: Allows setting leaving `min_capacity` or `max_capacity` unset. ([#8777](https://github.com/hashicorp/terraform-provider-aws/issues/8777))
* resource/aws_appautoscaling_scheduled_action: No longer re-creates when changes can be updated in-place. ([#8777](https://github.com/hashicorp/terraform-provider-aws/issues/8777))
* resource/aws_cognito_user_pool: Add support for `configuration_set` in `email_configuration` ([#14935](https://github.com/hashicorp/terraform-provider-aws/issues/14935))
* resource/aws_cognito_user_pool_client: Add plan time validation for `name`, `default_redirect_uri`, `supported_identity_providers` ([#14935](https://github.com/hashicorp/terraform-provider-aws/issues/14935))
* resource/aws_cognito_user_pool_client: Add support for `access_token_validity` and `id_token_validity`, `token_validity_units` ([#14935](https://github.com/hashicorp/terraform-provider-aws/issues/14935))
* resource/aws_db_instance: Allow `snapshot_identifier` to be removed from configuration without resource recreation ([#18013](https://github.com/hashicorp/terraform-provider-aws/issues/18013))
* resource/aws_elasticache_replication_group: Allows creating a Replication Group as part of a Global Replication Group ([#17725](https://github.com/hashicorp/terraform-provider-aws/issues/17725))
* resource/aws_kinesis_analytics_application: Add `start_application` attribute ([#17784](https://github.com/hashicorp/terraform-provider-aws/issues/17784))
* resource/aws_kinesis_analytics_application: `starting_position_configuration` can be specified when starting an application ([#17784](https://github.com/hashicorp/terraform-provider-aws/issues/17784))
* resource/aws_mq_broker: Add RabbitMQ as option for `engine_type`, and new arguments `authentication_strategy`, `ldap_server_metadata`, and `storage_type`. Improve handling of eventual consistency. ([#16108](https://github.com/hashicorp/terraform-provider-aws/issues/16108))
* resource/aws_mq_broker: Support updating broker engine version without recreating broker ([#12758](https://github.com/hashicorp/terraform-provider-aws/issues/12758))

BUG FIXES:

* resource/aws_rds_cluster_instance: Add `configuring-iam-database-auth` pending state ([#17982](https://github.com/hashicorp/terraform-provider-aws/issues/17982))
* resource/aws_storagegateway_upload_buffer: Replace `Provider produced inconsistent result after apply` with actual error message ([#17880](https://github.com/hashicorp/terraform-provider-aws/issues/17880))

## 3.31.0 (March 04, 2021)

FEATURES:

* **New Resource:** `aws_route53_hosted_zone_dnssec` ([#17474](https://github.com/hashicorp/terraform-provider-aws/issues/17474))

ENHANCEMENTS:

* data-source/aws_msk_cluster: Orders `bootstrap_brokers`, `bootstrap_brokers_sasl_scram`, `bootstrap_brokers_tls`, and `zookeeper_connect_string` ([#17579](https://github.com/hashicorp/terraform-provider-aws/issues/17579))
* provider: Support automatic region validation for `ap-northeast-3` ([#17934](https://github.com/hashicorp/terraform-provider-aws/issues/17934))
* resource/aws_globalaccelerator_accelerator: Add plan time validation to `name`, `flow_logs_s3_bucket` and `flow_logs_s3_prefix` attributes ([#17739](https://github.com/hashicorp/terraform-provider-aws/issues/17739))
* resource/aws_msk_cluster: Orders `bootstrap_brokers`, `bootstrap_brokers_sasl_scram`, `bootstrap_brokers_tls`, and `zookeeper_connect_string` ([#17579](https://github.com/hashicorp/terraform-provider-aws/issues/17579))
* resource/aws_route53_record: Support `DS` value for `type` argument ([#17040](https://github.com/hashicorp/terraform-provider-aws/issues/17040))

BUG FIXES:

* resource/aws_acm_certificate: Trigger resource recreation with `VALIDATION_TIMED_OUT` status ([#17869](https://github.com/hashicorp/terraform-provider-aws/issues/17869))
* resource/aws_globalaccelerator_accelerator: Allow update of flow log attribute for active flow logs ([#17739](https://github.com/hashicorp/terraform-provider-aws/issues/17739))
* resource/aws_kms_grant: Adds support for operations on asymmetric keys ([#17836](https://github.com/hashicorp/terraform-provider-aws/issues/17836))
* resource/aws_neptune_cluster_instance: Add "storage-optimization" to Neptune cluster instance create/update pending states ([#17901](https://github.com/hashicorp/terraform-provider-aws/issues/17901))
* resource/aws_neptune_cluster_parameter_group: Correctly update resource by `id` ([#17872](https://github.com/hashicorp/terraform-provider-aws/issues/17872))
* resource/aws_ssm_maintenance_window_task: Prevent `ValidationException` error on update when priority is not set or 0 ([#17885](https://github.com/hashicorp/terraform-provider-aws/issues/17885))

## 3.30.0 (February 26, 2021)

FEATURES:

* **New Data Source:** `aws_apigatewayv2_api` ([#13883](https://github.com/hashicorp/terraform-provider-aws/issues/13883))
* **New Data Source:** `aws_apigatewayv2_apis` ([#13883](https://github.com/hashicorp/terraform-provider-aws/issues/13883))
* **New Resource:** `aws_cognito_user_pool_ui_customization` ([#8114](https://github.com/hashicorp/terraform-provider-aws/issues/8114))
* **New Resource:** `aws_ecrpublic_repository` ([#16865](https://github.com/hashicorp/terraform-provider-aws/issues/16865))
* **New Resource:** `aws_sagemaker_app` ([#17251](https://github.com/hashicorp/terraform-provider-aws/issues/17251))

ENHANCEMENTS:

* provider: Add validation for `role_arn`, `policy_arns`, and `policy` ([#12642](https://github.com/hashicorp/terraform-provider-aws/issues/12642))
* resource/aws_autoscaling_group: Added support Auto Scaling groups with multiple launch templates using a mixed instances policy ([#16325](https://github.com/hashicorp/terraform-provider-aws/issues/16325))
* resource/aws_dms_certificate: Add `tags` argument ([#17163](https://github.com/hashicorp/terraform-provider-aws/issues/17163))
* resource/aws_gamelift_build: Support all valid operating system values ([#17764](https://github.com/hashicorp/terraform-provider-aws/issues/17764))
* resource/aws_sagemaker_domain: Make `default_resource_spec` optional for the `tensor_board_app_settings`, `jupyter_server_app_settings` and `kernel_gateway_app_settings` config blocks. ([#17251](https://github.com/hashicorp/terraform-provider-aws/issues/17251))
* resource/aws_sns_topic_subscription: Add `email`, `email-json`, and `firehose` to protocol values. Add `subscription_role_arn` argument for Firehose support. Add `confirmation_was_authenticated`, `owner_id`, and `pending_confirmation` attributes. ([#14923](https://github.com/hashicorp/terraform-provider-aws/issues/14923))

BUG FIXES:

* provider: Underlying Terraform Plugin SDK update to ensure data source errors include configuration source (file and line) ([#17801](https://github.com/hashicorp/terraform-provider-aws/issues/17801))
* resource/aws_backup_plan: `backup_options` and `resource_type` attributes in `advanced_backup_setting` configuration block are both required ([#17692](https://github.com/hashicorp/terraform-provider-aws/issues/17692))
* resource/aws_glue_trigger: Support starting ON_DEMAND triggers via `enabled` flag. ([#17488](https://github.com/hashicorp/terraform-provider-aws/issues/17488))
* resource/aws_sagemaker_domain: Wait for update to finish. ([#17251](https://github.com/hashicorp/terraform-provider-aws/issues/17251))
* resource/aws_sagemaker_user_profile: Wait for update to finish. ([#17251](https://github.com/hashicorp/terraform-provider-aws/issues/17251))
* resource/aws_sns_topic_subscription: Fix to avoid `delivery_policy` always showing diff. ([#14255](https://github.com/hashicorp/terraform-provider-aws/issues/14255))

## 3.29.1 (February 23, 2021)

ENHANCEMENTS:

* resource/aws_iam_role: Add `inline_policy` and `managed_policy_arns` arguments to support exclusive policy management ([#5904](https://github.com/hashicorp/terraform-provider-aws/issues/5904))

BUG FIXES:

* data-source/aws_iam_policy_document: Keep empty conditions ([#17752](https://github.com/hashicorp/terraform-provider-aws/issues/17752))
* resource/aws_db_instance: Fix conflicting argument validation error ([#17755](https://github.com/hashicorp/terraform-provider-aws/issues/17755))
* resource/aws_instance: Prevent error with `iam_instance_profile` containing additional forward slashes from path ([#17734](https://github.com/hashicorp/terraform-provider-aws/issues/17734))
* resource/aws_lb_target_group_attachment: Retry InvalidTarget errors when creating ([#8538](https://github.com/hashicorp/terraform-provider-aws/issues/8538))
* resource/aws_synthetics_canary: Fix Canary Update when in running state ([#17704](https://github.com/hashicorp/terraform-provider-aws/issues/17704))

## 3.29.0 (February 19, 2021)

FEATURES:

* **New Resource:** `aws_cloudwatch_event_archive` ([#17270](https://github.com/hashicorp/terraform-provider-aws/issues/17270))
* **New Resource:** `aws_elasticache_global_replication_group` ([#15885](https://github.com/hashicorp/terraform-provider-aws/issues/15885))
* **New Resource:** `aws_s3_object_copy` ([#15461](https://github.com/hashicorp/terraform-provider-aws/issues/15461))
* **New Resource:** `aws_securityhub_invite_accepter` ([#12684](https://github.com/hashicorp/terraform-provider-aws/issues/12684))

ENHANCEMENTS:

* data-source/aws_ami: Add `usage_operation`, `platform_details`, `ena_support` attributes ([#13971](https://github.com/hashicorp/terraform-provider-aws/issues/13971))
* data-source/aws_security_groups: Adds `arns` attribute ([#13944](https://github.com/hashicorp/terraform-provider-aws/issues/13944))
* data-source/aws_subnet: Add `available_ip_address_count` attributes ([#13554](https://github.com/hashicorp/terraform-provider-aws/issues/13554))
* resource/aws_ami: Add `usage_operation`, `platform_details`, `image_owner_alias`, `image_type`, `hypervisor`, `owner_id`, `platform`, `public` attributes ([#13971](https://github.com/hashicorp/terraform-provider-aws/issues/13971))
* resource/aws_ami_copy: Add `usage_operation`, `platform_details`, `image_owner_alias`, `image_type`, `hypervisor`, `owner_id`, `platform`, `public` attributes ([#13971](https://github.com/hashicorp/terraform-provider-aws/issues/13971))
* resource/aws_ami_from_instance: Add `usage_operation`, `platform_details`, `image_owner_alias`, `image_type`, `hypervisor`, `owner_id`, `platform`, `public` attributes ([#13971](https://github.com/hashicorp/terraform-provider-aws/issues/13971))
* resource/aws_cloudwatch_event_target: Adds `dead_letter_config` attributes ([#17241](https://github.com/hashicorp/terraform-provider-aws/issues/17241))
* resource/aws_cloudwatch_event_target: Adds `retry_policy` attributes ([#17241](https://github.com/hashicorp/terraform-provider-aws/issues/17241))
* resource/aws_cloudwatch_metric_alarm: Add plan time validation to `alarm_name`, `comparison_operator`, `metric_name`, `metric_query.id`, `metric_query.expression`, `metric_query.metric.metric_name`, `metric_query.metric.namespace`, `metric_query.metric.unit`, `namespace`, `period`, `statistic`, `alarm_description`, `insufficient_data_actions`, `ok_actions`, `unit`, and `extended_statistic` ([#12817](https://github.com/hashicorp/terraform-provider-aws/issues/12817))
* resource/aws_cognito_user_pool_client: Add support for `application_arn` in the `analytics_configuration` block. ([#16734](https://github.com/hashicorp/terraform-provider-aws/issues/16734))
* resource/aws_db_instance: Adds plan-time validation for `username` and `name` when `snapshot_identifier` is set ([#17156](https://github.com/hashicorp/terraform-provider-aws/issues/17156))
* resource/aws_dx_gateway_association: Changes to `proposal_id` do not force resource recreation ([#12482](https://github.com/hashicorp/terraform-provider-aws/issues/12482))
* resource/aws_ecs_capacity_provider: Add `managed_scaling` block `instance_warmup_period` argument ([#16941](https://github.com/hashicorp/terraform-provider-aws/issues/16941))
* resource/aws_lambda_function: Handle eventual consistency issues after publishing a version ([#14578](https://github.com/hashicorp/terraform-provider-aws/issues/14578))
* resource/aws_spot_instance_request: Add import support ([#12787](https://github.com/hashicorp/terraform-provider-aws/issues/12787))
* resource/aws_spot_instance_request: Add plan time validation for `spot_type` and `block_duration_minutes` ([#12787](https://github.com/hashicorp/terraform-provider-aws/issues/12787))
* resource/ses_receipt_rule_set: Add `arn` attribute ([#17611](https://github.com/hashicorp/terraform-provider-aws/issues/17611))
* resource/ses_receipt_rule_set: Add plan time validation to `name` ([#17611](https://github.com/hashicorp/terraform-provider-aws/issues/17611))

BUG FIXES:

* resource/aws_ebs_volume: Only specify throughput on update for `gp3` volumes ([#17646](https://github.com/hashicorp/terraform-provider-aws/issues/17646))
* resource/aws_fms_policy: Update `resource_type_list` plan-time validation to include `AWS::EC2::VPC`. ([#17595](https://github.com/hashicorp/terraform-provider-aws/issues/17595))
* resource/aws_lb_cookie_stickiness_policy: Allow zero value for `cookie_expiration_period` ([#17204](https://github.com/hashicorp/terraform-provider-aws/issues/17204))
* resource/aws_lb_listener_certificate: Prevent resource ID parsing error with IAM Server Certificate names containing underscores ([#17645](https://github.com/hashicorp/terraform-provider-aws/issues/17645))
* resource/aws_lb_target_group: Use gRPC matcher when using gRPC protocol ([#17534](https://github.com/hashicorp/terraform-provider-aws/issues/17534))
* resource/aws_ses_receipt_rule: Fix name validation regex to include `.` (period) ([#17627](https://github.com/hashicorp/terraform-provider-aws/issues/17627))
* resource/aws_ssm_document: Recreate resource on `name` update ([#17582](https://github.com/hashicorp/terraform-provider-aws/issues/17582))
* resource/aws_transfer_ssh_key: Corrects user_name validation ([#17621](https://github.com/hashicorp/terraform-provider-aws/issues/17621))
* resource/aws_transfer_user: Corrects user_name validation ([#17621](https://github.com/hashicorp/terraform-provider-aws/issues/17621))

## 3.28.0 (February 12, 2021)

FEATURES:

* **New Data Source:** `aws_cloudfront_cache_policy` ([#17336](https://github.com/hashicorp/terraform-provider-aws/issues/17336))
* **New Resource:** `aws_cloudfront_cache_policy` ([#17336](https://github.com/hashicorp/terraform-provider-aws/issues/17336))
* **New Resource:** `aws_cloudfront_realtime_log_config` ([#14974](https://github.com/hashicorp/terraform-provider-aws/issues/14974))
* **New Resource:** `aws_config_conformance_pack` ([#17313](https://github.com/hashicorp/terraform-provider-aws/issues/17313))
* **New Resource:** `aws_sagemaker_model_package_group` ([#17366](https://github.com/hashicorp/terraform-provider-aws/issues/17366))
* **New Resource:** `aws_securityhub_organization_admin_account` ([#17501](https://github.com/hashicorp/terraform-provider-aws/issues/17501))
* **New Resource:** `aws_synthetics_canary` ([#13140](https://github.com/hashicorp/terraform-provider-aws/issues/13140))

ENHANCEMENTS:

* data-source/aws_customer_gateway: Add `device_name` attribute ([#14786](https://github.com/hashicorp/terraform-provider-aws/issues/14786))
* data-source/aws_iam_policy_document: Support merging policy documents by adding `source_policy_documents` and `override_policy_documents` arguments ([#12055](https://github.com/hashicorp/terraform-provider-aws/issues/12055))
* provider: Add terraform-provider-aws version to HTTP User-Agent header ([#17486](https://github.com/hashicorp/terraform-provider-aws/issues/17486))
* resource/aws_budgets_budget: Add `arn` attribute ([#13139](https://github.com/hashicorp/terraform-provider-aws/issues/13139))
* resource/aws_budgets_budget: Add plan time validation for `budget_type`, `time_unit`, and `subscriber_sns_topic_arns` arguments ([#13139](https://github.com/hashicorp/terraform-provider-aws/issues/13139))
* resource/aws_cloudfront_distribution: Add `cache_policy_id` attribute ([#17336](https://github.com/hashicorp/terraform-provider-aws/issues/17336))
* resource/aws_cloudfront_distribution: Add `realtime_log_config_arn` attribute to `default_cache_behavior` and `ordered_cache_behavior` configuration blocks ([#14974](https://github.com/hashicorp/terraform-provider-aws/issues/14974))
* resource/aws_cloudfront_public_key: Add import support ([#17044](https://github.com/hashicorp/terraform-provider-aws/issues/17044))
* resource/aws_cloudwatch_log_destination: Add plan time validation to `role_arn`, `name` and `target_arn`. ([#11687](https://github.com/hashicorp/terraform-provider-aws/issues/11687))
* resource/aws_cloudwatch_log_group: Add plan time validation for `retention_in_days` argument ([#14673](https://github.com/hashicorp/terraform-provider-aws/issues/14673))
* resource/aws_codebuild_report_group: Add `delete_reports` argument ([#17338](https://github.com/hashicorp/terraform-provider-aws/issues/17338))
* resource/aws_codestarconnections_connection: Add `tags` argument ([#16835](https://github.com/hashicorp/terraform-provider-aws/issues/16835))
* resource/aws_customer_gateway: Add `device_name` argument ([#14786](https://github.com/hashicorp/terraform-provider-aws/issues/14786))
* resource/aws_dynamodb_table: Add plan-time validation for indexes on undefined attributes ([#6364](https://github.com/hashicorp/terraform-provider-aws/issues/6364))
* resource/aws_ec2_capacity_reservation: Add `owner_id` attribute ([#17129](https://github.com/hashicorp/terraform-provider-aws/issues/17129))
* resource/aws_ec2_traffic_mirror_filter: Add `arn` attribute. ([#13948](https://github.com/hashicorp/terraform-provider-aws/issues/13948))
* resource/aws_ec2_traffic_mirror_filter_rule: Add arn attribute. ([#13949](https://github.com/hashicorp/terraform-provider-aws/issues/13949))
* resource/aws_ec2_traffic_mirror_filter_rule: Add plan time validation to `destination_port_range.from_port`,
`destination_port_range.to_port`, `source_port_range.from_port`, and `source_port_range.to_port`. ([#13949](https://github.com/hashicorp/terraform-provider-aws/issues/13949))
* resource/aws_elastictranscoder_pipeline: Add plan time validations to `content_config.storage_class`, `content_config_permissions.access`, `content_config_permissions.grantee_type`,
`notifications.completed`, `notifications.error`, `notifications.progressing`, `notifications.warning`,
`thumbnail_config.storage_class`, `thumbnail_config_permissions.access`, `thumbnail_config_permissions.grantee_type` ([#13973](https://github.com/hashicorp/terraform-provider-aws/issues/13973))
* resource/aws_fms_policy: Allow use of `resource_type` or `resource_type_list` attributes ([#17418](https://github.com/hashicorp/terraform-provider-aws/issues/17418))
* resource/aws_imagebuilder_image_recipe: Add `gp3` as a valid value for the `volume_type` attribute ([#17286](https://github.com/hashicorp/terraform-provider-aws/issues/17286))
* resource/aws_lambda_event_source_mapping: Add `topics` attribute to support Amazon MSK as an event source ([#14746](https://github.com/hashicorp/terraform-provider-aws/issues/14746))
* resource/aws_lb_listener_certificate: Add import support ([#16474](https://github.com/hashicorp/terraform-provider-aws/issues/16474))
* resource/aws_licensemanager_license_configuration: Add `arn` and `owner_account_id` attributes ([#17160](https://github.com/hashicorp/terraform-provider-aws/issues/17160))
* resource/aws_ses_active_receipt_rule_set: Add `arn` attribute ([#13962](https://github.com/hashicorp/terraform-provider-aws/issues/13962))
* resource/aws_ses_active_receipt_rule_set: Add plan time validation for `rule_set_name` argument ([#13962](https://github.com/hashicorp/terraform-provider-aws/issues/13962))
* resource/aws_ses_configuration_set: Add `arn` attribute. ([#13972](https://github.com/hashicorp/terraform-provider-aws/issues/13972))
* resource/aws_ses_configuration_set: Add `delivery_options` argument ([#11600](https://github.com/hashicorp/terraform-provider-aws/issues/11600))
* resource/aws_ses_configuration_set: Add plan time validation to `name`. ([#13972](https://github.com/hashicorp/terraform-provider-aws/issues/13972))
* resource/aws_ses_event_destination: Add `arn` attribute ([#13964](https://github.com/hashicorp/terraform-provider-aws/issues/13964))
* resource/aws_ses_event_destination: Add plan time validation for `name`, `cloudwatch_destination.default_value`, `cloudwatch_destination.default_name`, `kinesis_destination.role_arn`, `kinesis_destination.stream_arn`, and `sns_destination.topic_arn` attributes ([#13964](https://github.com/hashicorp/terraform-provider-aws/issues/13964))
* resource/aws_ses_receipt_rule: Add `arn` attribute ([#13960](https://github.com/hashicorp/terraform-provider-aws/issues/13960))
* resource/aws_ses_receipt_rule: Add plan time validations for `name`, `tls_policy`, `add_header_action.header_name`, `add_header_action.header_value`, `bounce_action.topic_arn`, `lambda_action.function_arn`, `lambda_action.topic_arn`, `lambda_action.invocation_type`, `s3_action,topic_arn`, `sns_action.topic_arn`, `stop_action.scope`, `stop_action.topic_arn`, `workmail_action.topic_arn`, and `workmail_action.organization_arn` attributes ([#13960](https://github.com/hashicorp/terraform-provider-aws/issues/13960))
* resource/aws_ses_template: Add `arn` attribute ([#13963](https://github.com/hashicorp/terraform-provider-aws/issues/13963))
* resource/aws_sns_topic_subscription: Add `redrive_policy` argument ([#11770](https://github.com/hashicorp/terraform-provider-aws/issues/11770))
* resource/aws_ssm_association: Add `apply_only_at_cron_interval` argument ([#15038](https://github.com/hashicorp/terraform-provider-aws/issues/15038))
* resource/aws_ssm_document: Add `version_name` argument ([#14128](https://github.com/hashicorp/terraform-provider-aws/issues/14128))
* resource/aws_ssm_maintenance_window_task: Add `task_invocation_parameters` `run_command_parameters` block `cloudwatch_config` and `document_version` arguments ([#11774](https://github.com/hashicorp/terraform-provider-aws/issues/11774))
* resource/aws_ssm_maintenance_window_task: Add plan time validation to `max_concurrency`, `max_errors`, `priority`, `service_role_arn`, `targets`, `targets.notification_arn`, `targets.service_role_arn`, `task_type`, `task_invocation_parameters.run_command_parameters.comment`, `task_invocation_parameters.run_command_parameters.document_hash`, `task_invocation_parameters.run_command_parameters.timeout_seconds`, and `task_invocation_parameters.run_command_parameters.notification_config.notification_events` arguments ([#11774](https://github.com/hashicorp/terraform-provider-aws/issues/11774))
* resource/aws_ssm_maintenance_window_task: Make `service_role_arn` optional ([#12200](https://github.com/hashicorp/terraform-provider-aws/issues/12200))
* resource/aws_ssm_patch_baseline: Add `approval_rule` block `approve_until_date` argument ([#13850](https://github.com/hashicorp/terraform-provider-aws/issues/13850))
* resource/aws_ssm_patch_baseline: Add `approved_patches_enable_non_security` and `rejected_patches_action` arguments ([#11772](https://github.com/hashicorp/terraform-provider-aws/issues/11772))
* resource/aws_ssm_patch_baseline: Add `source` configuration block ([#11879](https://github.com/hashicorp/terraform-provider-aws/issues/11879))
* resource/aws_ssm_patch_baseline: Adds `arn` attribute. ([#11772](https://github.com/hashicorp/terraform-provider-aws/issues/11772))
* resource/aws_ssm_patch_baseline: Adds plan time validation for `name`, `description`, `global_filter.key`, `global_filter.values`,
`approved_patches`, `rejected_patches`, `approval_rule.approve_after_days`, `approval_rule.patch_filter.key`, and `approval_rule.patch_filter.values`. ([#11772](https://github.com/hashicorp/terraform-provider-aws/issues/11772))

BUG FIXES:

* resource/aws_glue_catalog_database: Use Catalog Id when deleting Databases. ([#17489](https://github.com/hashicorp/terraform-provider-aws/issues/17489))
* resource/aws_iam_instance_profile: Detach role when role doesn't exist + remove when deleted from state. ([#16188](https://github.com/hashicorp/terraform-provider-aws/issues/16188))
* resource/aws_instance: Fix use of `throughput` and `iops` for `gp3` volumes at the same time ([#17380](https://github.com/hashicorp/terraform-provider-aws/issues/17380))
* resource/aws_lambda_event_source_mapping: Wait for create and update operations to complete ([#14765](https://github.com/hashicorp/terraform-provider-aws/issues/14765))
* resource/aws_lambda_function: Prevent crash when using `Image` package type ([#17082](https://github.com/hashicorp/terraform-provider-aws/issues/17082))
* resource/aws_ssm_parameter: Use ARN value from API response rather than generating the value ([#16618](https://github.com/hashicorp/terraform-provider-aws/issues/16618))
* resource/aws_wafv2_web_acl_association: Increase creation timeout value from 2 to 5 minutes to prevent WAFUnavailableEntityException ([#17545](https://github.com/hashicorp/terraform-provider-aws/issues/17545))

## 3.27.0 (February 05, 2021)

FEATURES:

* **New Resource:** `aws_ec2_transit_gateway_prefix_list_reference` ([#16823](https://github.com/hashicorp/terraform-provider-aws/issues/16823))
* **New Resource:** `aws_route53_key_signing_key` ([#16840](https://github.com/hashicorp/terraform-provider-aws/issues/16840))
* **New Resource:** `aws_cloudfront_origin_request_policy` ([#17342](https://github.com/hashicorp/terraform-provider-aws/issues/17342))
* **New Data Source:** `aws_cloudfront_origin_request_policy` ([#17342](https://github.com/hashicorp/terraform-provider-aws/issues/17342))

ENHANCEMENTS:

* data-source/aws_subnet: Add `customer_owned_ipv4_pool` and `map_customer_owned_ip_on_launch` attributes ([#16676](https://github.com/hashicorp/terraform-provider-aws/issues/16676))
* resource/aws_glacier_vault: Add plan-time validation for `notification` configuration block `events` and `sns_topic_arn` arguments ([#12645](https://github.com/hashicorp/terraform-provider-aws/issues/12645))
* resource/aws_glue_catalog_table: Adds support for specifying schema from schema registry. ([#17335](https://github.com/hashicorp/terraform-provider-aws/issues/17335))
* resource/aws_iam_access_key: Add `create_date` attribute ([#17318](https://github.com/hashicorp/terraform-provider-aws/issues/17318))
* resource/aws_iam_access_key: Support resource import ([#17321](https://github.com/hashicorp/terraform-provider-aws/issues/17321))
* resource/aws_subnet: Add `customer_owned_ipv4_pool` and `map_customer_owned_ip_on_launch` attributes ([#16676](https://github.com/hashicorp/terraform-provider-aws/issues/16676))
* resource/aws_lb: Add `ipv6_address` attribute ([#17229](https://github.com/hashicorp/terraform-provider-aws/issues/17229))
* resource/aws_sfn_state_machine: Add support for `EXPRESS` state machine `type` ([#12249](https://github.com/hashicorp/terraform-provider-aws/issues/12249))
* resource/aws_lb_target_group: Add `protocol_version` attribute ([#17260](https://github.com/hashicorp/terraform-provider-aws/issues/17260))
* resource/aws_cloudfront_distribution: Add `cloudfront_origin_request_policy_id` attribute ([#17342](https://github.com/hashicorp/terraform-provider-aws/issues/17342))

BUG FIXES:

* data-source/aws_partition: Correct `reverse_dns_prefix` value in AWS China, C2S, and SC2S ([#17142](https://github.com/hashicorp/terraform-provider-aws/issues/17142))
* provider: Only validate AWS shared configuration profile SSO configuration when attempting to use SSO cached credentials ([#17469](https://github.com/hashicorp/terraform-provider-aws/issues/17469))
* resource/aws_api_gateway_method_settings: Ignore non-existent resource errors during deletion ([#17234](https://github.com/hashicorp/terraform-provider-aws/issues/17234))
* resource/aws_api_gateway_method_settings: Prevent confusing Terraform error on resource disappearance during creation ([#17234](https://github.com/hashicorp/terraform-provider-aws/issues/17234))
* resource/aws_cloudwatch_event_rule: Prevent perpetual differences with `name_prefix` argument values beginning with `terraform-` ([#17030](https://github.com/hashicorp/terraform-provider-aws/issues/17030))
* resource/aws_glacier_vault: Prevent crash with `GetVaultAccessPolicy` API errors ([#12645](https://github.com/hashicorp/terraform-provider-aws/issues/12645))
* resource/aws_glacier_vault: Properly remove from state when resource does not exist ([#12645](https://github.com/hashicorp/terraform-provider-aws/issues/12645))
* resource/aws_glue_crawler: Use standard retry timeout for IAM eventual consistency and retry on LakeFormation permissions errors ([#17256](https://github.com/hashicorp/terraform-provider-aws/issues/17256))
* resource/aws_glue_partition: Fix `partition_values` to preserve order. ([#17344](https://github.com/hashicorp/terraform-provider-aws/issues/17344))
* resource/aws_iam_access_key: Ensure `Inactive` `status` is properly configured during resource creation ([#17322](https://github.com/hashicorp/terraform-provider-aws/issues/17322))
* resource/aws_kinesis_firehose_delivery_stream: Use standard retry timeout for IAM eventual consistency and retry on LakeFormation access errors ([#17254](https://github.com/hashicorp/terraform-provider-aws/issues/17254))
* resource/aws_security_group: Prevent perpetual differences with `name_prefix` argument values beginning with `terraform-` ([#17030](https://github.com/hashicorp/terraform-provider-aws/issues/17030))
* resource/aws_ssoadmin_permission_set: Properly update resource with `relay_state` argument ([#17423](https://github.com/hashicorp/terraform-provider-aws/issues/17423))
* resource/aws_vpc_endpoint: Return unsuccessful deletion information immediately as an error instead of timing out while waiting for deletion ([#16656](https://github.com/hashicorp/terraform-provider-aws/issues/16656))
* resource/aws_vpc_endpoint_service: Return unsuccessful deletion information immediately as an error instead of timing out while waiting for deletion ([#16656](https://github.com/hashicorp/terraform-provider-aws/issues/16656))

## 3.26.0 (January 28, 2021)

NOTES:

* data-source/aws_route53_zone: The Route 53 `ListResourceRecordSets` API call has been implemented to support the `name_servers` attribute for private Hosted Zones similar to the resource implementation. Environments using restrictive IAM permissions may require updates. ([#17002](https://github.com/hashicorp/terraform-provider-aws/issues/17002))

FEATURES:

* **New Data Source:** `aws_imagebuilder_image` ([#16710](https://github.com/hashicorp/terraform-provider-aws/issues/16710))
* **New Resource:** `aws_imagebuilder_image` ([#16710](https://github.com/hashicorp/terraform-provider-aws/issues/16710))
* **New Resource:** `aws_prometheus_workspace` ([#16882](https://github.com/hashicorp/terraform-provider-aws/issues/16882))
* **New Resource:** `aws_sagemaker_app_image_config` ([#17221](https://github.com/hashicorp/terraform-provider-aws/issues/17221))

ENHANCEMENTS:

* data-source/aws_elasticache_replication_group: Add `multi_az_enabled` argument ([#17320](https://github.com/hashicorp/terraform-provider-aws/issues/17320))
* data-source/aws_vpc_peering_connection: Add `cidr_block_set` and `peer_cidr_block_set` attributes ([#13420](https://github.com/hashicorp/terraform-provider-aws/issues/13420))
* provider: Support AWS Single-Sign On (SSO) cached credentials ([#17340](https://github.com/hashicorp/terraform-provider-aws/issues/17340))
* resource/aws_codeartifact_domain: Make `encryption_key` optional ([#17262](https://github.com/hashicorp/terraform-provider-aws/issues/17262))
* resource/aws_elasticache_replication_group: Add `multi_az_enabled` argument ([#17320](https://github.com/hashicorp/terraform-provider-aws/issues/17320))
* resource/aws_elasticache_replication_group: Allow changing `cluster_mode.replica_count` without re-creation ([#17301](https://github.com/hashicorp/terraform-provider-aws/issues/17301))

BUG FIXES:

* data-source/aws_elb_hosted_zone_id: Correct values for `cn-north-1` and `cn-northwest-1` regions ([#17226](https://github.com/hashicorp/terraform-provider-aws/issues/17226))
* data-source/aws_lb_listener: Prevent error when retrieving a listener whose default action contains weighted target groups ([#17238](https://github.com/hashicorp/terraform-provider-aws/issues/17238))
* data-source/aws_route53_zone: Ensure `name_servers` is populated for private Hosted Zones ([#17002](https://github.com/hashicorp/terraform-provider-aws/issues/17002))
* resource/aws_ebs_volume: Allow both `size` and `snapshot_id` attributes to be specified ([#17243](https://github.com/hashicorp/terraform-provider-aws/issues/17243))
* resource/aws_elasticache_replication_group: Correctly update computed `member_clusters` values ([#17201](https://github.com/hashicorp/terraform-provider-aws/issues/17201))
* resource/aws_sagemaker_code_repository: fix doc name ([#17221](https://github.com/hashicorp/terraform-provider-aws/issues/17221))

## 3.25.0 (January 22, 2021)

NOTES

* resource/aws_lightsail_instance: The `ipv6_address` attribute has been deprecated. Use the `ipv6_addresses` attribute instead. This is due to a backwards incompatible change in the Lightsail API. ([#17155](https://github.com/hashicorp/terraform-provider-aws/issues/17155))

FEATURES

* **New Resource:** `aws_backup_global_settings` ([#16475](https://github.com/hashicorp/terraform-provider-aws/issues/16475))
* **New Resource:** `aws_sagemaker_feature_group` ([#16728](https://github.com/hashicorp/terraform-provider-aws/issues/16728))
* **New Resource:** `aws_sagemaker_image_version` ([#17141](https://github.com/hashicorp/terraform-provider-aws/issues/17141))
* **New Resource:** `aws_sagemaker_user_profile` ([#17123](https://github.com/hashicorp/terraform-provider-aws/issues/17123))

ENHANCEMENTS

* data-source/aws_ami: Add `throughput` attribute to `block_device_mappings` `ebs` attribute ([#16631](https://github.com/hashicorp/terraform-provider-aws/issues/16631))
* data-source/aws_ebs_volume: Add `throughput` attribute ([#16517](https://github.com/hashicorp/terraform-provider-aws/issues/16517))
* data-source/aws_elasticache_replication_group: Adds `arn` attribute ([#15348](https://github.com/hashicorp/terraform-provider-aws/issues/15348))
* data-source/aws_iam_user: Add `tags` attribute ([#13287](https://github.com/hashicorp/terraform-provider-aws/issues/13287))
* resource/aws_ami: Support `volume_type` value of `gp3` and add `throughput` argument to `ebs_block_device` configuration block ([#16631](https://github.com/hashicorp/terraform-provider-aws/issues/16631))
* resource/aws_ami_copy: Add `throughput` argument to `ebs_block_device` configuration block ([#16631](https://github.com/hashicorp/terraform-provider-aws/issues/16631))
* resource/aws_ami_from_instance: Add `throughput` argument to `ebs_block_device` configuration block ([#16631](https://github.com/hashicorp/terraform-provider-aws/issues/16631))
* resource/aws_ebs_volume: Add `throughput` argument ([#16517](https://github.com/hashicorp/terraform-provider-aws/issues/16517))
* resource/aws_elasticache_replication_group: Adds `arn` attribute ([#15348](https://github.com/hashicorp/terraform-provider-aws/issues/15348))
* resource/aws_lightsail_instance: Add `ipv6_addresses` attribute ([#17155](https://github.com/hashicorp/terraform-provider-aws/issues/17155))
* resource/aws_sagemaker_domain: Delete implicit EFS file system ([#17123](https://github.com/hashicorp/terraform-provider-aws/issues/17123))

BUG FIXES

* data-source/aws_lambda_function: Prevent error when getting Code Signing Config for container image based lambdas during read ([#17180](https://github.com/hashicorp/terraform-provider-aws/issues/17180))
* provider: Fix error messages for missing required blocks not including the block name ([#17211](https://github.com/hashicorp/terraform-provider-aws/issues/17211))
* provider: Prevent panic when sending Ctrl-C (SIGINT) to Terraform ([#17211](https://github.com/hashicorp/terraform-provider-aws/issues/17211))
* resource/aws_api_gateway_authorizer: Ensure `authorizer_credentials` are configured when `type` is `COGNITO_USER_POOLS` ([#16614](https://github.com/hashicorp/terraform-provider-aws/issues/16614))
* resource/aws_api_gateway_rest_api: Allow `api_key_source`, `binary_media_types`, and `description` arguments to be omitted from configuration with OpenAPI specification import (`body` argument) ([#17099](https://github.com/hashicorp/terraform-provider-aws/issues/17099))
* resource/aws_api_gateway_rest_api: Ensure `api_key_source`, `binary_media_types`, `description`, `minimum_compression_size`, `name`, and `policy` configuration values are correctly applied as an override after OpenAPI specification import (`body` argument) ([#17099](https://github.com/hashicorp/terraform-provider-aws/issues/17099))
* resource/aws_api_gateway_rest_api: Fix `disable_execute_api_endpoint` and `endpoint_configuration` `vpc_endpoint_ids` handling with OpenAPI specification import (`body` argument) ([#17209](https://github.com/hashicorp/terraform-provider-aws/issues/17209))
* resource/aws_lakeformation_data_lake_settings: Avoid unnecessary resource cycling ([#17189](https://github.com/hashicorp/terraform-provider-aws/issues/17189))
* resource/aws_lakeformation_permissions: Handle resources with multiple permissions ([#17189](https://github.com/hashicorp/terraform-provider-aws/issues/17189))
* resource/aws_lambda_function: Prevent panic with missing `FunctionConfiguration` `PackageType` attribute in API response ([#16544](https://github.com/hashicorp/terraform-provider-aws/issues/16544))
* resource/aws_lambda_function: Prevent panic with missing environment variable value ([#17056](https://github.com/hashicorp/terraform-provider-aws/issues/17056))
* resource/aws_sagemaker_image: Fix catching image not found on read error ([#17141](https://github.com/hashicorp/terraform-provider-aws/issues/17141))


## 3.24.1 (January 15, 2021)

BUG FIXES

* data-source/instance: Fix EBS and root block device tags issue with "Invalid address to set" ([#17136](https://github.com/hashicorp/terraform-provider-aws/issues/17136))

## 3.24.0 (January 14, 2021)

FEATURES

* **New Data Source:** `aws_api_gateway_domain_name` ([#12489](https://github.com/hashicorp/terraform-provider-aws/issues/12489))
* **New Data Source:** `aws_identitystore_group` ([#15322](https://github.com/hashicorp/terraform-provider-aws/issues/15322))
* **New Data Source:** `aws_identitystore_user` ([#15322](https://github.com/hashicorp/terraform-provider-aws/issues/15322))
* **New Resource:** `aws_cloudwatch_composite_alarm` ([#15023](https://github.com/hashicorp/terraform-provider-aws/issues/15023))
* **New Resource:** `aws_fms_policy` ([#9594](https://github.com/hashicorp/terraform-provider-aws/issues/9594))
* **New Resource:** `aws_route53_resolver_dnssec_config` ([#17012](https://github.com/hashicorp/terraform-provider-aws/issues/17012))
* **New Resource:** `aws_sagemaker_domain` ([#16077](https://github.com/hashicorp/terraform-provider-aws/issues/16077))
* **New Resource:** `aws_ssoadmin_account_assignment` ([#15322](https://github.com/hashicorp/terraform-provider-aws/issues/15322))

ENHANCEMENTS

* data-source/aws_workspaces_directory: Add access properties ([#16688](https://github.com/hashicorp/terraform-provider-aws/issues/16688))
* resource/aws_api_gateway_base_path_mapping: Support in-place updates for `api_id`, `base_path`, and `stage_name` ([#16147](https://github.com/hashicorp/terraform-provider-aws/issues/16147))
* resource/aws_api_gateway_domain_name: Add `mutual_tls_authentication` configuration block ([#15258](https://github.com/hashicorp/terraform-provider-aws/issues/15258))
* resource/aws_api_gateway_integration: Add `tls_config` configuration block ([#15499](https://github.com/hashicorp/terraform-provider-aws/issues/15499))
* resource/aws_api_gateway_method: Add `operation_name` argument ([#13282](https://github.com/hashicorp/terraform-provider-aws/issues/13282))
* resource/aws_api_gateway_rest_api: Add `disable_execute_api_endpoint` argument ([#16198](https://github.com/hashicorp/terraform-provider-aws/issues/16198))
* resource/aws_api_gateway_rest_api: Add `parameters` argument ([#7374](https://github.com/hashicorp/terraform-provider-aws/issues/7374))
* resource/aws_apigatewayv2_integration: Add `response_parameters` attribute ([#17043](https://github.com/hashicorp/terraform-provider-aws/issues/17043))
* resource/aws_codepipeline: Deprecates GitHub v1 (OAuth token) authentication and removes hashing of GitHub token ([#16959](https://github.com/hashicorp/terraform-provider-aws/issues/16959))
* resource/aws_codepipeline: Adds GitHub v2 (CodeStar Connetion) authentication ([#16959](https://github.com/hashicorp/terraform-provider-aws/issues/16959))
* resource/aws_dms_endpoint: Add `s3_settings` `date_partition_enabled` argument ([#16827](https://github.com/hashicorp/terraform-provider-aws/issues/16827))
* resource/aws_elasticache_cluster: Add support for final snapshot with Redis engine ([#15592](https://github.com/hashicorp/terraform-provider-aws/issues/15592))
* resource/aws_elasticache_replication_group: Add support for final snapshot ([#15592](https://github.com/hashicorp/terraform-provider-aws/issues/15592))
* resource/aws_globalaccelerator_accelerator: Add custom timeouts ([#17112](https://github.com/hashicorp/terraform-provider-aws/issues/17112))
* resource/aws_globalaccelerator_endpoint_group: Add custom timeouts ([#17112](https://github.com/hashicorp/terraform-provider-aws/issues/17112))
* resource/aws_globalaccelerator_endpoint_listener: Add custom timeouts ([#17112](https://github.com/hashicorp/terraform-provider-aws/issues/17112))
* resource/aws_instance: Add `tags` parameter to `root_block_device`, `ebs_block_device` blocks.([#15474](https://github.com/hashicorp/terraform-provider-aws/issues/15474))
* resource/aws_workspaces_directory: Add access properties ([#16688](https://github.com/hashicorp/terraform-provider-aws/issues/16688))

BUG FIXES

* resource/aws_appmesh_route: Allow an empty `match` attribute to specified for a `grpc_route`, indicating that any service should be matched ([#16867](https://github.com/hashicorp/terraform-provider-aws/issues/16867))
* resource/aws_db_instance: Correctly validate `final_snapshot_identifier` argument at plan-time ([#16885](https://github.com/hashicorp/terraform-provider-aws/issues/16885))
* resource/aws_dms_endpoint: Support `extra_connection_attributes` for all engine names during create and read ([#16827](https://github.com/hashicorp/terraform-provider-aws/issues/16827))
* resource/aws_instance: Prevent `volume_tags` from improperly interfering with `tags` in `aws_ebs_volume` ([#15474](https://github.com/hashicorp/terraform-provider-aws/issues/15474))
* resource/aws_networkfirewall_rule_group: Prevent resource recreation due to `stateful_rule` changes after creation ([#16884](https://github.com/hashicorp/terraform-provider-aws/issues/16884))
* resource/aws_route53_zone_association: Prevent deletion errors for missing Hosted Zone or VPC association ([#17023](https://github.com/hashicorp/terraform-provider-aws/issues/17023))
* resource/aws_sagemaker_image: fix error on wait for delete when image does not exist ([#16077](https://github.com/hashicorp/terraform-provider-aws/issues/16077))
* resource/aws_s3_bucket_inventory: Prevent crashes with empty `destination`, `filter`, and `schedule` configuration blocks ([#17055](https://github.com/hashicorp/terraform-provider-aws/issues/17055))
* service/apigateway: All operations will now automatically retry on `ConflictException: Unable to complete operation due to concurrent modification. Please try again later.` errors.

## 3.23.0 (January 08, 2021)

FEATURES

* **New Data Source:** `aws_ssoadmin_instances` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))
* **New Data Source:** `aws_ssoadmin_permission_set` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))
* **New Resource:** `aws_sagemaker_image` ([#16082](https://github.com/hashicorp/terraform-provider-aws/issues/16082))
* **New Resource:** `aws_ssoadmin_managed_policy_attachment` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))
* **New Resource:** `aws_ssoadmin_permission_set` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))
* **New Resource:** `aws_ssoadmin_permission_set_inline_policy` ([#15808](https://github.com/hashicorp/terraform-provider-aws/issues/15808))

ENHANCEMENTS

* data-source/aws_imagebuilder_image_recipe: Add `working_directory` attribute ([#16947](https://github.com/hashicorp/terraform-provider-aws/issues/16947))
* data-source/aws_elasticache_replication_group: Add reader_endpoint_address attribute ([#9979](https://github.com/hashicorp/terraform-provider-aws/issues/9979))
* resource/aws_elasticache_replication_group: Add reader_endpoint_address attribute ([#9979](https://github.com/hashicorp/terraform-provider-aws/issues/9979))
* resource/aws_elasticache_replication_group: Allows configuring `replicas_per_node_group` for "Redis (cluster mode disabled)" ([#16829](https://github.com/hashicorp/terraform-provider-aws/issues/16829))
* resource/aws_imagebuilder_image_recipe: Add `working_directory` argument ([#16947](https://github.com/hashicorp/terraform-provider-aws/issues/16947))
* resource/aws_glue_crawler: add support for `lineage_configuration` and `recrawl_policy` ([#16714](https://github.com/hashicorp/terraform-provider-aws/issues/16714))
* resource/aws_glue_crawler: add plan time validations to `name`, `description` and `table_prefix` ([#16714](https://github.com/hashicorp/terraform-provider-aws/issues/16714))
* resource/aws_kinesis_stream: Update `retention_period` argument plan-time validation to include up to 8760 hours ([#16608](https://github.com/hashicorp/terraform-provider-aws/issues/16608))
* resource/aws_msk_cluster: Support `PER_TOPIC_PER_PARTITION` value for `enhanced_monitoring` argument plan-time validation ([#16914](https://github.com/hashicorp/terraform-provider-aws/issues/16914))
* resource/aws_route53_zone: Add length validations for `delegation_set_id` and `name` arguments ([#12340](https://github.com/hashicorp/terraform-provider-aws/issues/12340))
* resource/aws_vpc_endpoint_service: Make `private_dns_name` configurable and add `private_dns_name_configuration` attribute ([#16495](https://github.com/hashicorp/terraform-provider-aws/issues/16495))

BUG FIXES

* resource/aws_emr_cluster: Remove from state instead of returning an error on long terminated cluster ([#16924](https://github.com/hashicorp/terraform-provider-aws/issues/16924))
* resource/aws_glue_catalog_table: Glue table partition keys should be set to empty list instead of being unset ([#16727](https://github.com/hashicorp/terraform-provider-aws/issues/16727))
* resource/aws_imagebuilder_distribution_configuration: Remove `user_ids` argument maximum limit ([#16905](https://github.com/hashicorp/terraform-provider-aws/issues/16905))
* resource/aws_transfer_user: Update `user_name` argument validation to support 100 characters ([#16938](https://github.com/hashicorp/terraform-provider-aws/issues/16938))

## 3.22.0 (December 18, 2020)

FEATURES

* **New Data Source:** `aws_ec2_managed_prefix_list` ([#16738](https://github.com/hashicorp/terraform-provider-aws/issues/16738))
* **New Data Source:** `aws_lakeformation_data_lake_settings` ([#13250](https://github.com/hashicorp/terraform-provider-aws/issues/13250))
* **New Data Source:** `aws_lakeformation_permissions` ([#13396](https://github.com/hashicorp/terraform-provider-aws/issues/13396))
* **New Data Source:** `aws_lakeformation_resource` ([#13396](https://github.com/hashicorp/terraform-provider-aws/issues/13396))
* **New Resource:** `aws_codestarconnections_connection` ([#15990](https://github.com/hashicorp/terraform-provider-aws/issues/15990))
* **New Resource:** `aws_ec2_managed_prefix_list` ([#14068](https://github.com/hashicorp/terraform-provider-aws/issues/14068))
* **New Resource:** `aws_lakeformation_data_lake_settings` ([#13250](https://github.com/hashicorp/terraform-provider-aws/issues/13250))
* **New Resource:** `aws_lakeformation_permissions` ([#13396](https://github.com/hashicorp/terraform-provider-aws/issues/13396))
* **New Resource:** `aws_lakeformation_resource` ([#13267](https://github.com/hashicorp/terraform-provider-aws/issues/13267))

ENHANCEMENTS

* data-source/aws_autoscaling_group: Adds `launch_template` attribute ([#16297](https://github.com/hashicorp/terraform-provider-aws/issues/16297))
* data-source/aws_availability_zone: Add `parent_zone_id`, `parent_zone_name`, and `zone_type` attributes (additional support for Local and Wavelength Zones) ([#16770](https://github.com/hashicorp/terraform-provider-aws/issues/16770))
* data-source/aws_eip: Add `carrier_ip` attribute ([#16724](https://github.com/hashicorp/terraform-provider-aws/issues/16724))
* data-source/aws_instance: Add `enclave_options` attribute (Nitro Enclaves) ([#16361](https://github.com/hashicorp/terraform-provider-aws/issues/16361))
* data-source/aws_instance: Add `ebs_block_device` and `root_block_device` configuration block `throughput` attribute ([#16620](https://github.com/hashicorp/terraform-provider-aws/issues/16620))
* data-source/aws_launch_configuration: Add `metadata_options` attribute ([#14637](https://github.com/hashicorp/terraform-provider-aws/issues/14637))
* data-source/aws_launch_template: Add `enclave_options` attribute (Nitro Enclaves) ([#16361](https://github.com/hashicorp/terraform-provider-aws/issues/16361))
* data-source/aws_network_interface: Add `association` `carrier_ip` and `customer_owned_ip` attributes ([#16723](https://github.com/hashicorp/terraform-provider-aws/issues/16723))
* resource/aws_autoscaling_group: Adds support for Instance Refresh ([#16678](https://github.com/hashicorp/terraform-provider-aws/issues/16678))
* resource/aws_eip: Add `carrier_ip` attribute ([#16724](https://github.com/hashicorp/terraform-provider-aws/issues/16724))
* resource/aws_instance: Add `enclave_options` configuration block (Nitro Enclaves) ([#16361](https://github.com/hashicorp/terraform-provider-aws/issues/16361))
* resource/aws_instance: Add `ebs_block_device` and `root_block_device` configuration block `throughput` attribute ([#16620](https://github.com/hashicorp/terraform-provider-aws/issues/16620))
* resource/aws_kinesis_firehose_delivery_stream: Mark `http_endpoint_configuration` `access_key` as sensitive ([#16684](https://github.com/hashicorp/terraform-provider-aws/issues/16684))
* resource/aws_launch_configuration: Add `metadata_options` configuration block ([#14637](https://github.com/hashicorp/terraform-provider-aws/issues/14637))
* resource/aws_launch_template: Add `enclave_options` configuration block (Nitro Enclaves) ([#16361](https://github.com/hashicorp/terraform-provider-aws/issues/16361))
* resource/aws_vpn_connection: Add support for VPN tunnel options and enable acceleration, DPDTimeoutAction, StartupAction, local/remote IPv4/IPv6 network CIDR and tunnel inside IP version. ([#14740](https://github.com/hashicorp/terraform-provider-aws/issues/14740))

BUG FIXES

* data-source/aws_ec2_coip_pools: Ensure all results from large environments are returned ([#16669](https://github.com/hashicorp/terraform-provider-aws/issues/16669))
* data-source/aws_ec2_local_gateways: Ensure all results from large environments are returned ([#16669](https://github.com/hashicorp/terraform-provider-aws/issues/16669))
* data-source/aws_ec2_local_gateway_route_tables: Ensure all results from large environments are returned ([#16669](https://github.com/hashicorp/terraform-provider-aws/issues/16669))
* data-source/aws_ec2_local_gateway_virtual_interface_groups: Ensure all results from large environments are returned ([#16669](https://github.com/hashicorp/terraform-provider-aws/issues/16669))
* data-source/aws_prefix_list: Using `name` argument no longer overrides other arguments ([#16739](https://github.com/hashicorp/terraform-provider-aws/issues/16739))
* resource/aws_db_instance: Fix missing `db_subnet_group_name` in API request when using `restore_to_point_in_time` ([#16830](https://github.com/hashicorp/terraform-provider-aws/issues/16830))
* resource/aws_eip_association: Handle eventual consistency when creating resource ([#16808](https://github.com/hashicorp/terraform-provider-aws/issues/16808))
* resource/aws_main_route_table_association: Prevent crash on creation when VPC main route table association is not found ([#16680](https://github.com/hashicorp/terraform-provider-aws/issues/16680))
* resource/aws_workspaces_workspace: Prevent panic from terminated WorkSpace ([#16692](https://github.com/hashicorp/terraform-provider-aws/issues/16692))

## 3.21.0 (December 11, 2020)

NOTES

* resource/aws_imagebuilder_image_recipe: Previously the ordering of `component` configuration blocks was not properly handled by the resource, which could cause unexpected behavior with multiple Components. These configurations may see the ordering difference being fixed after upgrade. ([#16566](https://github.com/hashicorp/terraform-provider-aws/issues/16566))

FEATURES

* **New Resource:** `aws_ec2_carrier_gateway` ([#16252](https://github.com/hashicorp/terraform-provider-aws/issues/16252))
* **New Resource:** `aws_glue_schema` ([#16612](https://github.com/hashicorp/terraform-provider-aws/issues/16612))

ENHANCEMENTS

* data-source/aws_launch_template: Add `associate_carrier_ip_address` attribute to `network_interfaces` configuration block ([#16707](https://github.com/hashicorp/terraform-provider-aws/issues/16707))
* data-source/aws_launch_template: Add `throughput` attribute to `block_device_mappings.ebs` configuration block ([#16649](https://github.com/hashicorp/terraform-provider-aws/issues/16649))
* data-source/aws_launch_template: Support `id` as argument ([#16457](https://github.com/hashicorp/terraform-provider-aws/issues/16457))
* resource/aws_appmesh_virtual_node: Add `listener.connection_pool` attribute ([#16167](https://github.com/hashicorp/terraform-provider-aws/issues/16167))
* resource/aws_appmesh_virtual_node: Add `listener.outlier_detection` attribute ([#16167](https://github.com/hashicorp/terraform-provider-aws/issues/16167))
* resource/aws_launch_template: Add `associate_carrier_ip_address` attribute to `network_interfaces` configuration block ([#16707](https://github.com/hashicorp/terraform-provider-aws/issues/16707))
* resource/aws_launch_template: Add `throughput` attribute to `block_device_mappings.ebs` configuration block ([#16649](https://github.com/hashicorp/terraform-provider-aws/issues/16649))
* resource/aws_spot_fleet_request: Add `throughput` attribute to `launch_specification.ebs_block_device` and `launch_specification.root_block_device` configuration blocks ([#16652](https://github.com/hashicorp/terraform-provider-aws/issues/16652))
* resource/aws_ssm_maintenance_window: Add `schedule_offset` argument ([#16569](https://github.com/hashicorp/terraform-provider-aws/issues/16569))
* resource/aws_workspaces_workspace: Add failed request error code along with message ([#16459](https://github.com/hashicorp/terraform-provider-aws/issues/16459))

BUG FIXES

* data-source/aws_customer_gateway: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_ec2_transit_gateway: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_ec2_transit_gateway_peering_attachment: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_ec2_transit_gateway_route_table: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_ec2_transit_gateway_vpc_attachment: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_guardduty_detector: Prevent missing `id` attribute when not configured as argument ([#16667](https://github.com/hashicorp/terraform-provider-aws/issues/16667))
* data-source/aws_imagebuilder_image_recipe: Ensure proper ordering of `component` attribute ([#16566](https://github.com/hashicorp/terraform-provider-aws/issues/16566))
* resource/aws_backup_plan: Prevent plan-time validation error for pre-existing resources with `lifecycle` `delete_after` and/or `copy_action` `lifecycle` `delete_after` arguments configured ([#16605](https://github.com/hashicorp/terraform-provider-aws/issues/16605))
* resource/aws_imagebuilder_image_recipe: Ensure proper ordering of `component` configuration blocks ([#16566](https://github.com/hashicorp/terraform-provider-aws/issues/16566))
* resource/aws_workspaces_directory: Fix empty custom_security_group_id & default_ou ([#16589](https://github.com/hashicorp/terraform-provider-aws/issues/16589))

## 3.20.0 (December 03, 2020)

ENHANCEMENTS

* resource/aws_backup_plan: Add plan-time validation for various arguments ([#16476](https://github.com/hashicorp/terraform-provider-aws/issues/16476))
* resource/aws_eks_node_group: Make `capacity_type` a `Computed` attribute ([#16552](https://github.com/hashicorp/terraform-provider-aws/issues/16552))
* resource/aws_lambda_event_source_mapping: Add support for updating `maximum_batching_window_in_seconds` for SQS queue event sources ([#16518](https://github.com/hashicorp/terraform-provider-aws/issues/16518))
* resource/aws_ssm_maintenance_window_target: Add plan-time validation for `owner_information` and `targets` arguments ([#16478](https://github.com/hashicorp/terraform-provider-aws/issues/16478))
* resource/aws_storagegateway_gateway: add `timeout_in_seconds`, `organizational_unit`, `domain_controllers` arguments for `smb_active_directory_settings` block. ([#16472](https://github.com/hashicorp/terraform-provider-aws/issues/16472))
* resource/aws_storagegateway_gateway: add `smb_active_directory_settings. active_directory_status`, `ec2_instance_id`, `endpoint_type`, `host_environment`, and `gateway_network_interface` attributes. ([#16472](https://github.com/hashicorp/terraform-provider-aws/issues/16472))
* resource/aws_storagegateway_gateway: add plan time validations for `smb_guest_password`, `smb_active_directory_settings. username`, `smb_active_directory_settings. password`, `smb_active_directory_settings. domain_name`, `gateway_timezone`, and `gateway_name`. ([#16472](https://github.com/hashicorp/terraform-provider-aws/issues/16472))
* resource/aws_storagegateway_gateway: add support for `medium_changer_type`  value `medium_changer_type`. ([#16472](https://github.com/hashicorp/terraform-provider-aws/issues/16472))

BUG FIXES

* resource/aws_backup_plan: Retry on eventual consistency error during deletion ([#16476](https://github.com/hashicorp/terraform-provider-aws/issues/16476))
* resource/aws_cloudwatch_event_target: Prevent potential panic and prevent recreation after state upgrade with custom `event_bus_name` value ([#16484](https://github.com/hashicorp/terraform-provider-aws/issues/16484))
* resource/aws_ec2_client_vpn_network_association: Increase associate and disassociate timeouts from 10min to 30min ([#16522](https://github.com/hashicorp/terraform-provider-aws/issues/16522))
* resource/aws_instance: Automatically retry instance restart on eventual consistency error during `instance_type` in-place update ([#16443](https://github.com/hashicorp/terraform-provider-aws/issues/16443))
* resource/aws_lambda_function: Prevent error during deletion when resource not found ([#16183](https://github.com/hashicorp/terraform-provider-aws/issues/16183))
* resource/aws_ssm_maintenance_window_target: Remove from state if not found ([#16478](https://github.com/hashicorp/terraform-provider-aws/issues/16478))

## 3.19.0 (December 01, 2020)

FEATURES

* **New Resource:** `aws_glue_registry` ([#16418](https://github.com/hashicorp/terraform-provider-aws/issues/16418))

ENHANCEMENTS

* resource/aws_apigatewayv2_domain_name: Add `mutual_tls_authentication` attribute to support mutual TLS authentication ([#15249](https://github.com/hashicorp/terraform-provider-aws/issues/15249))
* resource/aws_appmesh_virtual_gateway: Add `listener.connection_pool` attribute ([#16168](https://github.com/hashicorp/terraform-provider-aws/issues/16168))
* data-source/aws_eks_cluster: add `kubernetes_network_config` attribute ([#15518](https://github.com/hashicorp/terraform-provider-aws/issues/15518))
* resource/aws_storagegateway_smb_file_share: add support for `notification_policy` and `access_based_enumeration`. ([#16414](https://github.com/hashicorp/terraform-provider-aws/issues/16414))
* resource/aws_storagegateway_smb_file_share: add plan time validation to `invalid_user_list` and `valid_user_list`. ([#16414](https://github.com/hashicorp/terraform-provider-aws/issues/16414))
* resource/aws_cognito_user_pool: add support for account recovery setting. ([#12444](https://github.com/hashicorp/terraform-provider-aws/issues/12444))
* resource/aws_eks_cluster: add `kubernetes_network_config` argument ([#15518](https://github.com/hashicorp/terraform-provider-aws/issues/15518))
* resource/aws_eks_node_group: Add `capacity_type` argument and support multiple `instance_types` (Support Spot Node Groups) ([#16510](https://github.com/hashicorp/terraform-provider-aws/issues/16510))
* resource/aws_lambda_function: Add support for Container Images ([#16512](https://github.com/hashicorp/terraform-provider-aws/issues/16512))

BUG FIXES

* resource/aws_fsx_windows_file_system: Prevent potential panics, unexpected errors, and use correct operation timeout on update ([#16488](https://github.com/hashicorp/terraform-provider-aws/issues/16488))

## 3.18.0 (November 25, 2020)

FEATURES

* **New Data Source:** `aws_imagebuilder_image_pipeline` ([#16299](https://github.com/hashicorp/terraform-provider-aws/issues/16299))
* **New Data Source:** `aws_imagebuilder_image_recipe` ([#16218](https://github.com/hashicorp/terraform-provider-aws/issues/16218))
* **New Data Source:** `aws_serverlessrepository_application` ([#15874](https://github.com/hashicorp/terraform-provider-aws/issues/15874))
* **New Resource:** `aws_backup_region_settings` ([#16114](https://github.com/hashicorp/terraform-provider-aws/issues/16114))
* **New Resource:** `aws_imagebuilder_image_pipeline` ([#16299](https://github.com/hashicorp/terraform-provider-aws/issues/16299))
* **New Resource:** `aws_imagebuilder_image_recipe` ([#16218](https://github.com/hashicorp/terraform-provider-aws/issues/16218))
* **New Resource:** `aws_msk_scram_secret_association` ([#15302](https://github.com/hashicorp/terraform-provider-aws/issues/15302))
* **New Resource:** `aws_networkfirewall_resource_policy` ([#16279](https://github.com/hashicorp/terraform-provider-aws/issues/16279))
* **New Resource:** `aws_serverlessrepository_stack` ([#15874](https://github.com/hashicorp/terraform-provider-aws/issues/15874))

ENHANCEMENTS

* data-source/aws_codeartifact_repository_endpoint: Support `nuget` value in `format` argument plan-time validation ([#16422](https://github.com/hashicorp/terraform-provider-aws/issues/16422))
* data-source/aws_msk_cluster: Add `bootstrap_brokers_sasl_scram` attribute ([#15302](https://github.com/hashicorp/terraform-provider-aws/issues/15302))
* resource/aws_db_proxy_default_target_group: Make `connection_pool_config` optional ([#16303](https://github.com/hashicorp/terraform-provider-aws/issues/16303))
* resource/aws_kinesisanalyticsv2_application: `runtime_environment` now supports `FLINK-1_11` ([#16389](https://github.com/hashicorp/terraform-provider-aws/issues/16389))
* resource/aws_msk_cluster: Add `bootstrap_brokers_sasl_scram` attribute ([#15302](https://github.com/hashicorp/terraform-provider-aws/issues/15302))
* resource/aws_msk_cluster: Add `client_authentication` `sasl` `scram` argument ([#15302](https://github.com/hashicorp/terraform-provider-aws/issues/15302))
* resource/aws_networkfirewall_firewall: Add `firewall_status` attribute to expose VPC endpoints ([#16399](https://github.com/hashicorp/terraform-provider-aws/issues/16399))

BUG FIXES

* data-source/aws_lambda_function: Prevent Lambda `GetFunctionCodeSigningConfig` API call error outside AWS Commercial regions ([#16412](https://github.com/hashicorp/terraform-provider-aws/issues/16412))
* resource/aws_cloudwatch_event_permission: Prevent `arn: invalid prefix` error during read in some environments ([#16319](https://github.com/hashicorp/terraform-provider-aws/issues/16319))
* resource/aws_kinesis_analytics_application: Respect the order of 'record_column' attributes ([#16260](https://github.com/hashicorp/terraform-provider-aws/issues/16260))
* resource/aws_kinesisanalyticsv2_application: Respect the order of 'record_column' attributes ([#16260](https://github.com/hashicorp/terraform-provider-aws/issues/16260))
* resource/aws_lambda_function: Prevent Lambda `GetFunctionCodeSigningConfig` API call error outside AWS Commercial regions ([#16412](https://github.com/hashicorp/terraform-provider-aws/issues/16412))
* resource/aws_lb_listener: Mark `port` argument as optional and only default `protocol` argument to `HTTP` for Application Load Balancers (Support Gateway Load Balancer) ([#16306](https://github.com/hashicorp/terraform-provider-aws/issues/16306))
* resource/aws_securityhub_member: Prevent `invited` attribute updates due to recent API changes ([#16404](https://github.com/hashicorp/terraform-provider-aws/issues/16404))

## 3.17.0 (November 24, 2020)

FEATURES

* **New Data Source:** `aws_lambda_code_signing_config` ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* **New Data Source:** `aws_signer_signing_job` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))
* **New Data Source:** `aws_signer_signing_profile` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))
* **New Resource:** `aws_lambda_code_signing_config` ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* **New Resource:** `aws_signer_signing_job` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))
* **New Resource:** `aws_signer_signing_profile` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))
* **New Resource:** `aws_signer_signing_profile_permission` ([#16383](https://github.com/hashicorp/terraform-provider-aws/issues/16383))

ENHANCEMENTS

* data-source/aws_lambda_function: Add `code_signing_config_arn`, `signing_profile_version_arn`, and `signing_job_arn` attributes ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* data-source/aws_lambda_layer_version: Add `signing_profile_version_arn` and `signing_job_arn` attributes ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* resource/aws_accessanalyzer_analyzer: Adds plan time validation to `analyzer_name` ([#16265](https://github.com/hashicorp/terraform-provider-aws/issues/16265))
* resource/aws_accessanalyzer_analyzer: Adds plan time validation to `analyzer_name` ([#16265](https://github.com/hashicorp/terraform-provider-aws/issues/16265))
* resource/aws_fsx_windows_file_system: Support updating `throughput_capacity` and `storage_capacity` ([#15582](https://github.com/hashicorp/terraform-provider-aws/issues/15582))
* resource/aws_glue_catalog_table: Add partition index support ([#16194](https://github.com/hashicorp/terraform-provider-aws/issues/16194))
* resource/aws_lambda_function: Add `code_signing_config_arn` argument and `signing_profile_version_arn` and `signing_job_arn` attributes ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* resource/aws_lambda_layer_version: Add `signing_profile_version_arn` and `signing_job_arn` attributes ([#16384](https://github.com/hashicorp/terraform-provider-aws/issues/16384))
* resource/aws_storagegateway_nfs_file_share: Add support for `notification_policy`. ([#16340](https://github.com/hashicorp/terraform-provider-aws/issues/16340))
* resource/aws_storagegateway_nfs_file_share: Add plan time validation for `client_list`, `nfs_file_share_defaults. directory_mode`, `nfs_file_share_defaults. file_mode`, `nfs_file_share_defaults. group_id`, `nfs_file_share_defaults. owner_id` ([#16340](https://github.com/hashicorp/terraform-provider-aws/issues/16340))
* resource/aws_workspaces_directory: Allows assigning IP group ([#14451](https://github.com/hashicorp/terraform-provider-aws/issues/14451))

BUG FIXES

* resource/aws_fsx_windows_file_system: Update the default creation timeout from 30 to 45 minutes ([#16363](https://github.com/hashicorp/terraform-provider-aws/issues/16363))
* resource/aws_lb: Fix `enable_cross_zone_load_balancing` argument handling with Gateway Load Balancers ([#16314](https://github.com/hashicorp/terraform-provider-aws/issues/16314))

## 3.16.0 (November 18, 2020)

* **New Data Source:** `aws_imagebuilder_component` ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* **New Data Source:** `aws_imagebuilder_distribution_configuration` ([#16180](https://github.com/hashicorp/terraform-provider-aws/issues/16180))
* **New Data Source:** `aws_imagebuilder_infrastructure_configuration` ([#16186](https://github.com/hashicorp/terraform-provider-aws/issues/16186))
* **New Resource:** `aws_api_gateway_rest_api_policy` ([#13619](https://github.com/hashicorp/terraform-provider-aws/issues/13619))
* **New Resource:** `aws_backup_vault_policy` ([#16112](https://github.com/hashicorp/terraform-provider-aws/issues/16112))
* **New Resource:** `aws_glue_dev_endpoint` ([#7895](https://github.com/hashicorp/terraform-provider-aws/pull/7895))
* **New Resource:** `aws_imagebuilder_component` ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* **New Resource:** `aws_imagebuilder_distribution_configuration` ([#16180](https://github.com/hashicorp/terraform-provider-aws/issues/16180))
* **New Resource:** `aws_imagebuilder_infrastructure_configuration` ([#16186](https://github.com/hashicorp/terraform-provider-aws/issues/16186))
* **New Resource:** `aws_networkfirewall_firewall` ([#16277](https://github.com/hashicorp/terraform-provider-aws/issues/16277))
* **New Resource:** `aws_networkfirewall_firewall_policy` ([#16277](https://github.com/hashicorp/terraform-provider-aws/issues/16277))
* **New Resource:** `aws_networkfirewall_logging_configuration` ([#16277](https://github.com/hashicorp/terraform-provider-aws/issues/16277))
* **New Resource:** `aws_networkfirewall_rule_group` ([#16277](https://github.com/hashicorp/terraform-provider-aws/issues/16277))

ENHANCEMENTS

* resource/aws_globalaccelerator_endpoint_group: Add `arn` and `port_override` attributes ([#16121](https://github.com/hashicorp/terraform-provider-aws/issues/16121))
* resource/aws_glue_catalog_table: Add support for `parameters` argument to `storage_descriptor.columns` block ([#16052](https://github.com/hashicorp/terraform-provider-aws/issues/16052))
* resource/aws_glue_catalog_table: Add plan time validation for `description`, `name`, `partition_keys.name`, `partition_keys.comment`, `partition_keys.type`, `retention`, `view_original_text`, `view_expanded_text`, `storage_descriptor.name`, `storage_descriptor.comment`, `storage_descriptor.type`, `storage_descriptor.bucket_columns`, `storage_descriptor.ser_de_info.name`, `storage_descriptor.skewed_info.skewed_column_names`, `storage_descriptor.sort_columns.column`, `storage_descriptor.sort_columns.sort_order` ([#16052](https://github.com/hashicorp/terraform-provider-aws/issues/16052))
* resource/aws_msk_cluster: Support in-place `kafka_version` upgrade ([#13654](https://github.com/hashicorp/terraform-provider-aws/issues/13654))
* resource/aws_storagegateway_smb_file_share: Add `file_share_name` argument ([#16008](https://github.com/hashicorp/terraform-provider-aws/issues/16008))
* resource_aws_storagegateway_nfs_file_share: Add `file_share_name` argument ([#16072](https://github.com/hashicorp/terraform-provider-aws/issues/16072))

BUG FIXES

* data-source/aws_s3_bucket: Use provider credentials when getting the bucket region (fix AWS China non-ICP S3 Buckets and other restrictive environments) ([#15481](https://github.com/hashicorp/terraform-provider-aws/issues/15481))
* resource/aws_apigatewayv2_stage: Correctly handle deletion of route_settings ([#16133](https://github.com/hashicorp/terraform-provider-aws/issues/16133))
* resource/aws_backup_plan: `lifecycle` block in `copy_action` is optional ([#16116](https://github.com/hashicorp/terraform-provider-aws/issues/16116))
* resource/aws_eks_fargate_profile: Serialize multiple profile creation and deletion to prevent `ResourceInUseException` errors ([#14020](https://github.com/hashicorp/terraform-provider-aws/issues/14020))
* resource/aws_organizations_organization: Prevent recreation when `feature_set` is updated to `ALL` ([#15473](https://github.com/hashicorp/terraform-provider-aws/issues/15473))
* resource/aws_s3_bucket: Use provider credentials when getting the bucket region (fix AWS China non-ICP S3 Buckets and other restrictive environments) ([#15481](https://github.com/hashicorp/terraform-provider-aws/issues/15481))
* resource/aws_s3_bucket_object: Correctly updates `version_id` when certain configuration keys are changed ([#14900](https://github.com/hashicorp/terraform-provider-aws/issues/14900))

## 3.15.0 (November 12, 2020)

ENHANCEMENTS

* data-source/aws_ec2_transit_gateway_route_table: Add `arn` attribute ([#13921](https://github.com/hashicorp/terraform-provider-aws/issues/13921))
* data-source/aws_ec2_transit_gateway_vpc_attachment: Add `appliance_mode_support` attribute ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* data-source/aws_route_table: Add `route` `vpc_endpoint_id` attribute ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_db_instance: Add `restore_to_point_in_time` argument and `latest_restorable_time` attribute ([#15969](https://github.com/hashicorp/terraform-provider-aws/issues/15969))
* resource/aws_default_route_table: Add `route` configuration block `vpc_endpoint_id` argument ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_ec2_transit_gateway: Support in-place updates for most arguments ([#15556](https://github.com/hashicorp/terraform-provider-aws/issues/15556))
* resource/aws_ec2_transit_gateway_route_table: Add `arn` attribute ([#13921](https://github.com/hashicorp/terraform-provider-aws/issues/13921))
* resource/aws_ec2_transit_gateway_vpc_attachment: Add `appliance_mode_support` argument ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* resource/aws_ec2_transit_gateway_vpc_attachment_accepter: Add `appliance_mode_support` attribute ([#16159](https://github.com/hashicorp/terraform-provider-aws/issues/16159))
* resource/aws_kinesis_firehose_delivery_stream: Add `http_endpoint_configuration` configuration block ([#15356](https://github.com/hashicorp/terraform-provider-aws/issues/15356))
* resource/aws_lb: Support `load_balancer_type` argument value of `gateway` ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_lb_target_group: Support `protocol` argument value of `GENEVE` ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_rds_cluster: Add `restore_to_point_in_time` argument ([#7031](https://github.com/hashicorp/terraform-provider-aws/issues/7031))
* resource/aws_route: Add `vpc_endpoint_id` argument ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_route_table: Add `route` configuration block `vpc_endpoint_id` argument ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_vpc_endpoint: Support `vpc_endpoint_type` argument value `GatewayLoadBalancer` ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_vpc_endpoint_service: Add `gateway_load_balancer_arns` argument ([#16131](https://github.com/hashicorp/terraform-provider-aws/issues/16131))
* resource/aws_workspaces_workspace: Add configurable timeouts ([#15479](https://github.com/hashicorp/terraform-provider-aws/issues/15479))

BUG FIXES

* data-source/aws_network_interface: Prevent crash with ENI attachments missing DeviceIndex or AttachmentID ([#15567](https://github.com/hashicorp/terraform-provider-aws/issues/15567))
* resource/aws_cognito_identity_pool: Update `identity_pool_name` argument validation to include additional characters supported by the API ([#15773](https://github.com/hashicorp/terraform-provider-aws/issues/15773))
* resource/aws_db_instance: Ignore `DBInstanceNotFound` error during deletion ([#15942](https://github.com/hashicorp/terraform-provider-aws/issues/15942))
* resource/aws_ecs_service: Properly remove resource from Terraform state with `ClusterNotFoundException` error ([#15927](https://github.com/hashicorp/terraform-provider-aws/issues/15927))
* resource/aws_eip: In EC2-Classic, wait until Instance returns as associated during create or update ([#16032](https://github.com/hashicorp/terraform-provider-aws/issues/16032))
* resource/aws_eip_association: Retry on additional EC2 Address eventual consistency errors on creation ([#16032](https://github.com/hashicorp/terraform-provider-aws/issues/16032))
* resource/aws_eip_association: In EC2-Classic, wait until Instance returns as associated during creation ([#16032](https://github.com/hashicorp/terraform-provider-aws/issues/16032))
* resource/aws_kinesis_analytics_application: Handle IAM role eventual consistency issues ([#16125](https://github.com/hashicorp/terraform-provider-aws/issues/16125))
* resource/aws_kinesisanalyticsv2_application: Handle IAM role eventual consistency issues ([#16125](https://github.com/hashicorp/terraform-provider-aws/issues/16125))
* resource/aws_lb_target_group: Allow invalid configurations that were allowed prior to 3.10. ([#15613](https://github.com/hashicorp/terraform-provider-aws/issues/15613))
* resource/aws_network_interface: Prevent crash with ENI attachments missing DeviceIndex or AttachmentID ([#15567](https://github.com/hashicorp/terraform-provider-aws/issues/15567))
* resource/aws_s3_bucket: Add plan-time validation to `acl` ([#15327](https://github.com/hashicorp/terraform-provider-aws/issues/15327))
* resource/aws_workspaces_bundle: Fix empty (private) owner ([#14535](https://github.com/hashicorp/terraform-provider-aws/issues/14535))

## 3.14.1 (November 06, 2020)

BUG FIXES

* resource/aws_cloudwatch_event_target: Prevent regression from version 3.14.0 with `ListTargetsByRuleInput.EventBusName` error ([#16075](https://github.com/hashicorp/terraform-provider-aws/issues/16075))

## 3.14.0 (November 06, 2020)

FEATURES

* **New Data Source:** `aws_route53_resolver_endpoint` ([#8628](https://github.com/hashicorp/terraform-provider-aws/issues/8628))
* **New Data Source:** `aws_sagemaker_prebuilt_ecr_image` ([#15924](https://github.com/hashicorp/terraform-provider-aws/pull/15924))
* **New Data Source:** `aws_workspaces_workspace` ([#14135](https://github.com/hashicorp/terraform-provider-aws/issues/14135))
* **New Resource:** `aws_secretsmanager_secret_policy` ([#14468](https://github.com/hashicorp/terraform-provider-aws/pull/14468))

ENHANCEMENTS

* resource/aws_apigatewayv2_integration: `timeout_milliseconds` has different valid ranges and default values between HTTP and WebSocket APIs. `timeout_milliseconds` is now `Computed`, meaning Terraform will only perform drift detection of its value when present in a configuration. ([#16017](https://github.com/hashicorp/terraform-provider-aws/issues/16017))
* resource/aws_cloudwatch_event_permission: Add `event_bus_name` ([#15922](https://github.com/hashicorp/terraform-provider-aws/issues/15922))
* resource/aws_cloudwatch_event_target: Add plan time validation to `arn`, `role_arn`, `launch_type`, `task_definition_arn` ([#11685](https://github.com/hashicorp/terraform-provider-aws/issues/11685))
* resource/aws_cloudwatch_event_target: Add `event_bus_name` ([#15799](https://github.com/hashicorp/terraform-provider-aws/issues/15799))
* resource/aws_codeartifact_domain: add `tags` argument. ([#16006](https://github.com/hashicorp/terraform-provider-aws/issues/16006))
* resource/aws_codeartifact_repository: add `tags` argument. ([#16006](https://github.com/hashicorp/terraform-provider-aws/issues/16006))
* resource/aws_eip: Add `network_border_group` argument ([#14028](https://github.com/hashicorp/terraform-provider-aws/issues/14028))
* resource/aws_glue_catalog_database: add plan time validations for `description` and `name`. ([#15956](https://github.com/hashicorp/terraform-provider-aws/issues/15956))
* resource/aws_glue_crawler: Support MongoDB target ([#15934](https://github.com/hashicorp/terraform-provider-aws/issues/15934))
* resource/aws_glue_trigger: Add plan time validation to `name` ([#15793](https://github.com/hashicorp/terraform-provider-aws/issues/15793))
* resource/aws_glue_trigger: Add `security_configuration` and `notification_property` arguments to `actions` block ([#15793](https://github.com/hashicorp/terraform-provider-aws/issues/15793))
* resource/aws_kinesis_analytics_application: Wait for resource deletion. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: `inputs.parallelism` is a computed attribute. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: Handle `inputs.processing_configuration` addition and deletion. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: Handle `reference_data_sources` deletion. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: Handle `cloudwatch_logging_options` deletion. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_kinesis_analytics_application: Set the `description` attribute on creation. ([#16005](https://github.com/hashicorp/terraform-provider-aws/issues/16005))
* resource/aws_sagemaker_endpoint_configuration: Add support for `data_capture_config`. ([#15887](https://github.com/hashicorp/terraform-provider-aws/issues/15887))
* resource/aws_sagemaker_endpoint_configuration: Add plan time validation for `production_variants.accelerator_type`, `production_variants.instance_type`. ([#15887](https://github.com/hashicorp/terraform-provider-aws/issues/15887))
* resource/aws_sagemaker_model: Add support for `primary_container. image_config` and `containers.image_config` ([#15957](https://github.com/hashicorp/terraform-provider-aws/issues/15957))
* resource/aws_sagemaker_model: Add plan time validation for `execution_role_arn`  ([#15957](https://github.com/hashicorp/terraform-provider-aws/issues/15957))

BUG FIXES

* resource/aws_datasync_task: Allow `UNAVAILABLE` as pending status during creation ([#15949](https://github.com/hashicorp/terraform-provider-aws/issues/15949))
* resource/aws_glue_classifier: Fix `quote_symbol` being optional ([#15948](https://github.com/hashicorp/terraform-provider-aws/issues/15948))
* resource/aws_lambda_function: Publish version if value of `publish` is only change ([#15020](https://github.com/hashicorp/terraform-provider-aws/issues/15020))
* resource/aws_rds_cluster: Prevent error removing cluster from global cluster when not found ([#15938](https://github.com/hashicorp/terraform-provider-aws/issues/15938))
* resource/aws_rds_cluster: Prevent recreation when using `snapshot_identifier` and `kms_key_id` without `storage_encrypted = true` ([#15915](https://github.com/hashicorp/terraform-provider-aws/issues/15915))
* resource/aws_rds_cluster_instance: Add Cluster Identifier to creation error message ([#15939](https://github.com/hashicorp/terraform-provider-aws/issues/15939))
* resource/aws_rds_global_cluster: Prevent error removing cluster from global cluster when not found ([#15938](https://github.com/hashicorp/terraform-provider-aws/issues/15938))

## 3.13.0 (October 29, 2020)

NOTES

* data-source/aws_autoscaling_groups: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_caller_identity: The `id` attribute has changed to the ID of the AWS Account. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ebs_snapshot_ids: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ebs_volumes: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_coip_pools: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_instance_type_offerings: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateway_route_tables: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateway_virtual_interface_groups: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateways: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_spot_price: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_efs_access_points: The `id` attribute has changed to the EFS File System identifier. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_glue_script: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_inspector_rules_packages: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_instances: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_kms_ciphertext: The `id` attribute has changed to the KMS Key. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_kms_secrets: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15725](https://github.com/hashicorp/terraform-provider-aws/issues/15725))
* data-source/aws_network_acls: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_network_interfaces: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_organizations_organizational_units: The `id` attribute has changed to the parent identifier. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_outposts_outposts: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_outposts_sites: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_route_tables: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_route53_resolver_rules: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_s3_bucket_objects: The `id` attribute has changed to the name of the S3 Bucket. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_security_groups: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_vpc_peering_connections: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_vpcs: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))

FEATURES

* **New Resource:** `aws_glue_resource_policy` ([#10361](https://github.com/hashicorp/terraform-provider-aws/issues/10361))
* **New Resource:** `aws_s3control_bucket` ([#15510](https://github.com/hashicorp/terraform-provider-aws/issues/15510))
* **New Resource:** `aws_s3control_bucket_lifecycle_configuration` ([#15604](https://github.com/hashicorp/terraform-provider-aws/issues/15604))
* **New Resource:** `aws_s3control_bucket_policy` ([#15575](https://github.com/hashicorp/terraform-provider-aws/issues/15575))
* **New Resource:** `aws_s3outposts_endpoint` ([#15585](https://github.com/hashicorp/terraform-provider-aws/issues/15585))
* **New Resource:** `aws_sagemaker_code_repository` ([#15809](https://github.com/hashicorp/terraform-provider-aws/issues/15809))
* **New Resource:** `aws_storagegateway_tape_pool` ([#15370](https://github.com/hashicorp/terraform-provider-aws/issues/15370))

ENHANCEMENTS

* resource/aws_cloudwatch_event_rule: Add `event_bus_name` ([#15727](https://github.com/hashicorp/terraform-provider-aws/issues/15727))
* resource/aws_ecs_service: Add `wait_for_steady_state` argument ([#3485](https://github.com/hashicorp/terraform-provider-aws/issues/3485))
* resource/aws_s3_access_point: Support S3 on Outposts ([#15621](https://github.com/hashicorp/terraform-provider-aws/issues/15621))
* resource/aws_sagemaker_model: Add `container` configuration block `mode` argument ([#15371](https://github.com/hashicorp/terraform-provider-aws/issues/15371))
* resource/aws_sagemaker_notebook_instance: Add support for `additional_code_repositories` ([#15830](https://github.com/hashicorp/terraform-provider-aws/issues/15830))
* resource/aws_sagemaker_notebook_instance: Add `url` and `network_interface_id` attributes ([#15802](https://github.com/hashicorp/terraform-provider-aws/issues/15802))

BUG FIXES

* data-source/aws_autoscaling_groups: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_caller_identity: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ebs_snapshot_ids: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ebs_volumes: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_coip_pools: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_instance_type_offerings: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateway_route_tables: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateway_virtual_interface_groups: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_local_gateways: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_ec2_spot_price: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_efs_access_points: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_glue_script: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_inspector_rules_packages: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_instances: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_kms_ciphertext: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_kms_secrets: Prevent plan differences with the `id` attribute ([#15725](https://github.com/hashicorp/terraform-provider-aws/issues/15725))
* data-source/aws_network_acls: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_network_interfaces: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_organizations_organizational_units: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_outposts_outposts: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_outposts_sites: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_route_tables: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_route53_resolver_rules: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_s3_bucket_objects: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_security_groups: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_vpc_peering_connections: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* data-source/aws_vpcs: Prevent plan differences with the `id` attribute ([#15896](https://github.com/hashicorp/terraform-provider-aws/issues/15896))
* resource/aws_apigatewayv2_integration: Correctly handle update of AWS service integrations ([#15894](https://github.com/hashicorp/terraform-provider-aws/issues/15894))
* resource/aws_api_gateway_usage_plan: Change `api_stages` to from List to Set ([#14345](https://github.com/hashicorp/terraform-provider-aws/issues/14345))
* resource/aws_lambda_function: Update published `version` and `qualified_arn` on config changes ([#15121](https://github.com/hashicorp/terraform-provider-aws/issues/15121))
* resource/aws_rds_global_cluster: Prevent recreation when using encrypted `source_db_cluster_identifier` without `storage_encrypted` ([#15916](https://github.com/hashicorp/terraform-provider-aws/issues/15916))
* resource/aws_vpc_peering_connection_options: Only modify options that have changed ([#12126](https://github.com/hashicorp/terraform-provider-aws/issues/12126))

## 3.12.0 (October 22, 2020)

FEATURES

* **New Data Source:** `aws_rds_certificate` ([#15789](https://github.com/hashicorp/terraform-provider-aws/issues/15789))
* **New Resource:** `aws_autoscalingplans_scaling_plan` ([#8965](https://github.com/hashicorp/terraform-provider-aws/issues/8965))
* **New Resource:** `aws_cloudwatch_event_bus` ([#10256](https://github.com/hashicorp/terraform-provider-aws/issues/10256))
* **New Resource:** `aws_kinesisanalyticsv2_application` ([#11652](https://github.com/hashicorp/terraform-provider-aws/issues/11652))
* **New Resource:** `aws_storagegateway_stored_iscsi_volume` ([#12027](https://github.com/hashicorp/terraform-provider-aws/issues/12027))

ENHANCEMENTS

* resource/aws_cloudwatch_event_target: Add validation to `input_transformer.input_paths` map ([#15669](https://github.com/hashicorp/terraform-provider-aws/issues/15669))
* resource/aws_codeartifact_repository: support external connections ([#15569](https://github.com/hashicorp/terraform-provider-aws/issues/15569))
* resource/aws_fsx_lustre_file_system: Add `copy_tags_to_backups` support ([#15687](https://github.com/hashicorp/terraform-provider-aws/issues/15687))
* resource/aws_fsx_lustre_file_system: Increased maximum `automatic_backup_retention_days` from 35 to 90 ([#15641](https://github.com/hashicorp/terraform-provider-aws/issues/15641))
* resource/aws_fsx_windows_file_system: Increased maximum `automatic_backup_retention_days` from 35 to 90 ([#15641](https://github.com/hashicorp/terraform-provider-aws/issues/15641))
* resource/aws_glue_catalog_table: add validation checks for resource properties ([#12523](https://github.com/hashicorp/terraform-provider-aws/issues/12523))
* resource/aws_network_interface: Add `ipv6_addresses` and `ipv6_address_count` arguments ([#12281](https://github.com/hashicorp/terraform-provider-aws/issues/12281))
* resource/aws_sagemaker_notebook_instance: `lifecycle_config_name` and `root_access`  are updateable. ([#15385](https://github.com/hashicorp/terraform-provider-aws/issues/15385))
* resource/aws_sagemaker_notebook_instance: plan time validation for `role_arn`, `instance_type`. ([#15385](https://github.com/hashicorp/terraform-provider-aws/issues/15385))

BUGFIXES

* resource/aws_workspaces_workspace: Fix terminated state resolution ([#15705](https://github.com/hashicorp/terraform-provider-aws/issues/15705))
* resource/aws_glue_table_catalog_table: Prevent errors on `unset` of `ser_de_info.name` ([#15127](https://github.com/hashicorp/terraform-provider-aws/issues/15127))
* resource/aws_glue_security_configuration: Don't send empty `kms_arn` if mode is `DISABLED` ([#13618](https://github.com/hashicorp/terraform-provider-aws/issues/13618))

## 3.11.0 (October 15, 2020)

FEATURES

* **New Data Source:** `aws_codeartifact_repository_endpoint` ([#15566](https://github.com/hashicorp/terraform-provider-aws/issues/15566))
* **New Resource:** `aws_appmesh_gateway_route` ([#15638](https://github.com/hashicorp/terraform-provider-aws/issues/15638))
* **New Resource:** `aws_appmesh_virtual_gateway` ([#15611](https://github.com/hashicorp/terraform-provider-aws/issues/15611))

BUG FIXES

* resource/aws_ec2_transit_gateway_route: Prevent plan errors with compressed IPv6 addresses ([#14846](https://github.com/hashicorp/terraform-provider-aws/issues/14846))

ENHANCEMENTS

* data-source/aws_workspaces_directory: Add workspaces creation properties ([#14577](https://github.com/hashicorp/terraform-provider-aws/issues/14577))
* resource/aws_backup_plan: Add support for AdvancedBackupSettings ([#15341](https://github.com/hashicorp/terraform-provider-aws/issues/15341))
* resource/aws_sagemaker_notebook_instance: Add `default_code_repository` attribute ([#13772](https://github.com/hashicorp/terraform-provider-aws/issues/13772))
* resource/aws_sagemaker_notebook_instance: Add `volume_size` attribute ([#15521](https://github.com/hashicorp/terraform-provider-aws/issues/15521))
* resource/aws_workspaces_directory: Add workspaces creation properties ([#14577](https://github.com/hashicorp/terraform-provider-aws/issues/14577))

## 3.10.0 (October 09, 2020)

FEATURES

* **New Data Source:** `aws_codeartifact_authorization_token` ([#15425](https://github.com/hashicorp/terraform-provider-aws/issues/15425))
* **New Data Source:** `aws_ec2_instance_type` ([#13124](https://github.com/hashicorp/terraform-provider-aws/issues/13124))
* **New Data Source:** `aws_lex_bot_alias` ([#8919](https://github.com/hashicorp/terraform-provider-aws/issues/8919))
* **New Data Source:** `aws_redshift_orderable_cluster` ([#15438](https://github.com/hashicorp/terraform-provider-aws/issues/15438))
* **New Resource:** `aws_codeartifact_repository_permissions_policy` ([#15562](https://github.com/hashicorp/terraform-provider-aws/issues/15562))
* **New Resource:** `aws_lex_bot_alias` ([#8919](https://github.com/hashicorp/terraform-provider-aws/issues/8919))
* **New Resource:** `aws_s3_bucket_ownership_controls` ([#15482](https://github.com/hashicorp/terraform-provider-aws/issues/15482))

NOTES

* data-source/aws_acm_certificate: The `id` attribute has changed to the ARN of the ACM Certificate. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_autoscaling_group: The `id` attribute has changed to the name of the Auto Scaling Group. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_availability_zones: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_db_event_categories: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ebs_default_kms_key: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ebs_encryption_by_default: The `id` attribute has changed to the name of the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ec2_instance_type_offering: The `id` attribute has changed to the EC2 Instance Type. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ecr_authorization_token: The `id` attribute has changed to the AWS Region. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ecr_image: The `id` attribute has changed to the SHA256 digest of the ECR Image. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_eks_cluster_auth: The `id` attribute has changed to the name of the EKS Cluster. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_iam_account_alias: The `id` attribute has changed to the AWS Account Alias. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_kms_alias: The `id` attribute has changed to the ARN of the KMS Alias. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_partition: The `id` attribute has changed to the identifier of the AWS Partition. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_regions: The `id` attribute has changed to the identifier of the AWS Partition. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_sns_topic: The `id` attribute has changed to the ARN of the SNS Topic. The first apply of this updated data source may show this difference. ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))

ENHANCEMENTS

* data-source/aws_batch_compute_environment: Add `tags` attribute ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* data-source/aws_batch_job_queue: Add `tags` attribute ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* data-source/aws_vpc_endpoint_service: Accept `service_type` as argument ([#15467](https://github.com/hashicorp/terraform-provider-aws/issues/15467))
* resource/aws_appmesh_route: Add `timeout` configuration block to `grpc_route`, `http_route`, `http2_route` and `tcp_route` attributes. ([#14361](https://github.com/hashicorp/terraform-provider-aws/issues/14361))
* resource/aws_appmesh_virtual_node: Add `timeout` configuration block to `listener` attribute. ([#14361](https://github.com/hashicorp/terraform-provider-aws/issues/14361))
* resource/aws_batch_compute_environment: Add `tags` argument ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* resource/aws_batch_job_definition: Add `tags` argument ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* resource/aws_batch_job_queue: Add `tags` argument ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* resource/aws_lb_target_group: Add `source_ip` as an option for the `stickiness.type` argument. ([#15295](https://github.com/hashicorp/terraform-provider-aws/issues/15295))
* resource/aws_sns_topic_subscription: Create subscriptions with attributes (delivery policy, filter policy, etc.) instead of separate API calls ([#10496](https://github.com/hashicorp/terraform-provider-aws/issues/10496))

BUG FIXES

* data-source/aws_acm_certificate: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_autoscaling_group: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_availability_zones: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_db_event_categories: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ebs_default_kms_key: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ebs_encryption_by_default: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ec2_instance_type_offering: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ecr_authorization_token: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_ecr_image: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_eks_cluster_auth: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_iam_account_alias: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_kms_alias: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_partition: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_regions: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* data-source/aws_sns_topic: Prevent plan differences with the `id` attribute ([#15399](https://github.com/hashicorp/terraform-provider-aws/issues/15399))
* resource/aws_acm_certificate: Prevent unexpected timeout error on deletion due to API retries ([#15522](https://github.com/hashicorp/terraform-provider-aws/issues/15522))
* resource/aws_batch_job_definition: Prevent unexpected plan difference for `container_properties` argument value with new secrets support ([#15470](https://github.com/hashicorp/terraform-provider-aws/issues/15470))
* resource/aws_codestarnotifications_notification_rule: Prevent unexpected timeout error during target deletion due to API retries ([#15523](https://github.com/hashicorp/terraform-provider-aws/issues/15523))
* resource/aws_config_remediation_configuration: Prevent unexpected timeout error on deletion due to API retries ([#15524](https://github.com/hashicorp/terraform-provider-aws/issues/15524))
* resource/aws_db_proxy: Increase default deletion timeout to 60 minutes ([#15537](https://github.com/hashicorp/terraform-provider-aws/issues/15537))
* resource/aws_db_proxy_target: Ensure `db_proxy_name` and `target_group_name` attributes are properly imported ([#15537](https://github.com/hashicorp/terraform-provider-aws/issues/15537))
* resource/aws_ecs_cluster: Prevent IAM Service Linked Role error on first ECS provision ([#15457](https://github.com/hashicorp/terraform-provider-aws/issues/15457))
* resource/aws_emr_instance_fleet: Prevent error on deletion when EMR Cluster is no longer running ([#15548](https://github.com/hashicorp/terraform-provider-aws/issues/15548))
* resource/aws_emr_managed_scaling_policy: Ensure `cluster_id` attribute is properly saved during import ([#15541](https://github.com/hashicorp/terraform-provider-aws/issues/15541))
* resource/aws_emr_managed_scaling_policy: Handle additional cases where resource should be removed from Terraform state ([#15541](https://github.com/hashicorp/terraform-provider-aws/issues/15541))
* resource/aws_gamelift_fleet: Prevent unexpected timeout error on creation due to API retries ([#15526](https://github.com/hashicorp/terraform-provider-aws/issues/15526))
* resource/aws_glue_workflow: Ensure `max_concurrent_runs` attribute is properly saved during import ([#15538](https://github.com/hashicorp/terraform-provider-aws/issues/15538))
* resource/aws_lex_bot: Prevent unexpected timeout error on creation due to API retries ([#15527](https://github.com/hashicorp/terraform-provider-aws/issues/15527))
* resource/aws_lex_bot_alias: Prevent unexpected timeout error on creation due to API retries ([#15527](https://github.com/hashicorp/terraform-provider-aws/issues/15527))
* resource/aws_lex_intent: Prevent unexpected timeout error on creation due to API retries ([#15527](https://github.com/hashicorp/terraform-provider-aws/issues/15527))
* resource/aws_lex_slot_type: Prevent unexpected timeout error on creation due to API retries ([#15527](https://github.com/hashicorp/terraform-provider-aws/issues/15527))
* resource/aws_organizations_policy: Prevent errors with imported AWS-managed Organizations policies ([#15446](https://github.com/hashicorp/terraform-provider-aws/issues/15446))
* resource/aws_s3_bucket: Correctly handle provider-level ignored tag configuration ([#12013](https://github.com/hashicorp/terraform-provider-aws/issues/12013))
* resource/aws_s3_bucket: Correctly set expiration for lifecycle_rule with abort_incomplete_multipart_upload_days set ([#15263](https://github.com/hashicorp/terraform-provider-aws/issues/15263))
* resource/aws_s3_bucket_analytics_configuration: Prevent unexpected timeout error on deletion due to API retries ([#15529](https://github.com/hashicorp/terraform-provider-aws/issues/15529))
* resource/aws_s3_bucket_object: Correctly handle provider-level ignored tag configuration ([#12013](https://github.com/hashicorp/terraform-provider-aws/issues/12013))

## 3.9.0 (October 02, 2020)

FEATURES

* **New Resource:** `aws_backup_vault_notifications` ([#12501](https://github.com/hashicorp/terraform-provider-aws/issues/12501))
* **New Resource:** `aws_codeartifact_domain` ([#13743](https://github.com/hashicorp/terraform-provider-aws/issues/13743))
* **New Resource:** `aws_codeartifact_domain_permissions` ([#13753](https://github.com/hashicorp/terraform-provider-aws/issues/13753))
* **New Resource:** `aws_codeartifact_repository` ([#14429](https://github.com/hashicorp/terraform-provider-aws/issues/14429))
* **New Resource:** `aws_db_proxy_target` ([#12784](https://github.com/hashicorp/terraform-provider-aws/issues/12784))
* **New Resource:** `aws_glue_data_catalog_encryption_settings` ([#14916](https://github.com/hashicorp/terraform-provider-aws/issues/14916))
* **New Resource:** `aws_glue_ml_transform` ([#14909](https://github.com/hashicorp/terraform-provider-aws/issues/14909))
* **New Resource:** `aws_glue_partition` ([#12547](https://github.com/hashicorp/terraform-provider-aws/issues/12547))
* **New Resource:** `aws_lex_bot` ([#8918](https://github.com/hashicorp/terraform-provider-aws/issues/8918))
* **New Resource:** `aws_lex_intent` ([#8917](https://github.com/hashicorp/terraform-provider-aws/issues/8917))
* **New Data Source:** `aws_lex_bot` ([#8918](https://github.com/hashicorp/terraform-provider-aws/issues/8918))
* **New Data Source:** `aws_lex_intent` ([#8917](https://github.com/hashicorp/terraform-provider-aws/issues/8917))

ENHANCEMENTS

* resource/aws_appmesh_route: Add `grpc_route` and `http2_route` attributes to support gRPC and HTTP/2 services ([#11669](https://github.com/hashicorp/terraform-provider-aws/issues/11669))
* resource/aws_appmesh_route: Add `retry_policy` attribute to support App Mesh retry policies ([#11660](https://github.com/hashicorp/terraform-provider-aws/issues/11660))
* resource/aws_appmesh_virtual_node: Add `grpc` and `http2` as valid values for the `protocol` attribute ([#11669](https://github.com/hashicorp/terraform-provider-aws/issues/11669))
* resource/aws_appmesh_virtual_node: Add `spec.backend_defaults`, `spec.backend.virtual_service.client_policy` and `spec.listener.tls` attributes to support TLS in transit encryption ([#12541](https://github.com/hashicorp/terraform-provider-aws/issues/12541))
* resource/aws_appmesh_virtual_router: Add `grpc` and `http2` as valid values for the `protocol` attribute ([#11669](https://github.com/hashicorp/terraform-provider-aws/issues/11669))
* resource/aws_fsx_lustre_file_system: Add `auto_import_policy`  argument ([#15231](https://github.com/hashicorp/terraform-provider-aws/issues/15231))
* resource/aws_fsx_lustre_file_system: Support `daily_automatic_backup_start_time` ([#15299](https://github.com/hashicorp/terraform-provider-aws/issues/15299))
* resource/aws_fsx_lustre_file_system: Add `storage_type` and `drive_cache_type` ([#14727](https://github.com/hashicorp/terraform-provider-aws/issues/14727))
* resource/aws_glue_crawler: Add `connection_name` field to `s3_target` block ([#15350](https://github.com/hashicorp/terraform-provider-aws/issues/15350))
* resource/aws_sagemaker_notebook_instance: Ability to configure root access for Sagemaker notebook instances ([#14184](https://github.com/hashicorp/terraform-provider-aws/issues/14184))

BUG FIXES

* data-source/aws_s3_bucket_object: Prevent crash when S3 HeadObject returns empty response ([#14154](https://github.com/hashicorp/terraform-provider-aws/issues/14154))
* resource/aws_db_instance: Prevent ordering differences with `enabled_cloudwatch_logs_exports` argument ([#15404](https://github.com/hashicorp/terraform-provider-aws/issues/15404))
* resource/aws_ec2_client_vpn_authorization_rule: Increased active and revoked timeouts from 5 to 10 minutes ([#15367](https://github.com/hashicorp/terraform-provider-aws/issues/15367))
* resource/aws_rds_cluster: Prevent ordering differences with `enabled_cloudwatch_logs_exports` argument ([#15404](https://github.com/hashicorp/terraform-provider-aws/issues/15404))
* resource/aws_redshift_cluster: Increase default update timeout to 75 minutes ([#15339](https://github.com/hashicorp/terraform-provider-aws/issues/15339))

## 3.8.0 (September 24, 2020)

FEATURES

* **New Resource:** `aws_datasync_location_fsx_windows` ([#12686](https://github.com/hashicorp/terraform-provider-aws/pull/12686))
* **New Resource:** `aws_route53_resolver_query_log_config`. ([#14897](https://github.com/hashicorp/terraform-provider-aws/pull/14897))
* **New Resource:** `aws_route53_resolver_query_log_config_association`. ([#14901](https://github.com/hashicorp/terraform-provider-aws/pull/14901))
* **New Data Source:** `aws_rds_engine_version` ([#15228](https://github.com/hashicorp/terraform-provider-aws/pull/15228))
* **New Data Source:** `aws_docdb_engine_version` ([#15253](https://github.com/hashicorp/terraform-provider-aws/pull/15253))
* **New Data Source:** `aws_neptune_engine_version` ([#15259](https://github.com/hashicorp/terraform-provider-aws/pull/15259))
* **New Data Source:** `aws_workspaces_image` ([#11428](https://github.com/hashicorp/terraform-provider-aws/issues/11428))

ENHANCEMENTS

* data-source/aws_lb: Add `customer_owned_ipv4_pool` and `subnet_mapping` `outpost_id` attributes ([#15170](https://github.com/hashicorp/terraform-provider-aws/issues/15170))
* resource/aws_apigatewayv2_api: Add `disable_execute_api_endpoint` attribute ([#15250](https://github.com/hashicorp/terraform-provider-aws/issues/15250))
* resource/aws_apigatewayv2_authorizer: Add `authorizer_payload_format_version`, `authorizer_result_ttl_in_seconds` and `enable_simple_responses` attribute to support Lambda authorizers for HTTP APIs ([#15232](https://github.com/hashicorp/terraform-provider-aws/issues/15232))
* resource/aws_apigatewayv2_authorizer: Change `identity_sources` to an optional attribute ([#15232](https://github.com/hashicorp/terraform-provider-aws/issues/15232))
* resource/aws_appmesh_mesh: Add `mesh_owner` and `resource_owner` attributes ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_appmesh_route: Add `mesh_owner` argument and `resource_owner` attribute ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_appmesh_virtual_node: Add `mesh_owner` argument and `resource_owner` attribute ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_appmesh_virtual_router: Add `mesh_owner` argument and `resource_owner` attribute ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_appmesh_virtual_service: Add `mesh_owner` argument and `resource_owner` attribute ([#14349](https://github.com/hashicorp/terraform-provider-aws/issues/14349))
* resource/aws_elasticsearch_domain: Support `AUDIT_LOGS` log type ([#15218](https://github.com/hashicorp/terraform-provider-aws/issues/15218))
* resource/aws_glue_connection: Support `NETWORK` connection type ([#14818](https://github.com/hashicorp/terraform-provider-aws/issues/14818))
* resource/aws_glue_crawler: Add support for `scan_all` and `scan_rate` arguments for ddb targets ([#14819](https://github.com/hashicorp/terraform-provider-aws/issues/14819))
* resource/aws_glue_crawler: Allow removing `table_prefix` ([#15268](https://github.com/hashicorp/terraform-provider-aws/issues/15268))
* resource/aws_glue_job: Add `non_overridable_arguments` argument ([#14793](https://github.com/hashicorp/terraform-provider-aws/issues/14793))
* resource/aws_glue_workflow: Add `tags` argument ([#14910](https://github.com/hashicorp/terraform-provider-aws/issues/14910))
* resource/aws_glue_workflow: Add `arn` attribute ([#14910](https://github.com/hashicorp/terraform-provider-aws/issues/14910))
* resource/aws_glue_workflow: Add `max_concurrent_runs` argument ([#14910](https://github.com/hashicorp/terraform-provider-aws/issues/14910))
* resource/aws_glue_workflow: Plan time validation for `name` ([#14910](https://github.com/hashicorp/terraform-provider-aws/issues/14910))
* resource/aws_fsx_lustre_file_system: Add support for backup retention ([#14446](https://github.com/hashicorp/terraform-provider-aws/issues/14446))
* resource/aws_fsx_lustre_file_system: Add `kms_key_id` argument ([#15057](https://github.com/hashicorp/terraform-provider-aws/issues/15057))
* resource/aws_fsx_lustre_file_system: Add `mount_name` argument ([#14313](https://github.com/hashicorp/terraform-provider-aws/issues/14313))
* resource/aws_lb: Add `customer_owned_ipv4_pool` argument and `subnet_mapping` `outpost_id` attribute ([#15170](https://github.com/hashicorp/terraform-provider-aws/issues/15170))
* resource/aws_organizations_policy: Add `tags` argument ([#15316](https://github.com/hashicorp/terraform-provider-aws/issues/15316))
* resource/aws_rds_cluster: Add `allow_major_version_upgrade` argument ([#14709](https://github.com/hashicorp/terraform-provider-aws/issues/14709))
* resource/aws_storagegateway_smb_file_share: Add `admin_user_list` argument ([#12196](https://github.com/hashicorp/terraform-provider-aws/issues/12196))
* resource/aws_transfer_server: Support `VPC` value for `endpoint_type` argument and add `endpoint_details` configuration block `address_allocation_ids`, `subnet_ids`, and `vpc_id` arguments ([#12599](https://github.com/hashicorp/terraform-provider-aws/issues/12599))
* resource/aws_transfer_user: Add `home_directory_mappings` configuration blocks and `home_directory_type` argument ([#13591](https://github.com/hashicorp/terraform-provider-aws/issues/13591))

BUG FIXES

* resource/aws_dynamodb_table: Ensure changes in `name`, `range_key`, `projection_type`, or `non_key_attributes` of a `local_secondary_index` configuration block force resource recreation ([#12335](https://github.com/hashicorp/terraform-provider-aws/issues/12335))
* resource/aws_dynamodb_table: Ensure `local_secondary_index` `non_key_attributes` are sent through API requests on resource creation ([#15115](https://github.com/hashicorp/terraform-provider-aws/issues/15115))
* resource/aws_efs_mount_target: Increase create timeout to 30 minutes ([#15293](https://github.com/hashicorp/terraform-provider-aws/issues/15293))
* resource/aws_fsx_lustre_file_system: Change `aws_fsx_lustre_file_system's`'s `network_interface_ids` to `TypeList` to preserve ordering. ([#14314](https://github.com/hashicorp/terraform-provider-aws/issues/14314))
* resource/aws_neptune_cluster_instance: Add `configuring-enhanced-monitoring` to expected states when creating and updating ([#15284](https://github.com/hashicorp/terraform-provider-aws/issues/15284))
* resource/aws_vpn_gateway: Increase VPC detachment timeout to 30 minutes ([#15201](https://github.com/hashicorp/terraform-provider-aws/issues/15201))
* resource/aws_vpn_gateway_attachment: Increase VPC detachment timeout to 30 minutes ([#15201](https://github.com/hashicorp/terraform-provider-aws/issues/15201))

## 3.7.0 (September 17, 2020)

FEATURES

* **New Resource:** `aws_config_remediation_configuration` ([#13884](https://github.com/hashicorp/terraform-provider-aws/issues/13884))

ENHANCEMENTS

* resource/aws_db_cluster_snapshot: Add plan-time validation for `db_cluster_snapshot_identifier` argument ([#15132](https://github.com/hashicorp/terraform-provider-aws/issues/15132))
* resource/aws_kinesis_firehose_delivery_stream: Add `server_side_encryption` `key_arn` and `key_type` arguments (support KMS Customer Managed Key encryption) ([#11954](https://github.com/hashicorp/terraform-provider-aws/issues/11954))

BUG FIXES

* data-source/aws_kms_secrets: Prevent `plaintext` values to appear in CLI output with Terraform 0.13 ([#15169](https://github.com/hashicorp/terraform-provider-aws/issues/15169))
* resource/aws_acm_certificate: Prevent tagging is not permitted on re-import error ([#15060](https://github.com/hashicorp/terraform-provider-aws/issues/15060))
* resource/aws_cognito_identity_pool: Prevent ordering differences for `openid_connect_provider_arns` argument ([#15178](https://github.com/hashicorp/terraform-provider-aws/issues/15178))

## 3.6.0 (September 11, 2020)

FEATURES

* **New Resource:** `aws_db_proxy_default_target_group` ([#12743](https://github.com/hashicorp/terraform-provider-aws/issues/12743))

BUG FIXES

* resource/aws_ec2_client_vpn_authorization_rule: Increase active and revoked timeouts from 1 to 5 minutes ([#15037](https://github.com/hashicorp/terraform-provider-aws/issues/15037))

## 3.5.0 (September 03, 2020)

FEATURES

* **New Data Source:** `aws_docdb_orderable_db_instance` ([#14931](https://github.com/hashicorp/terraform-provider-aws/issues/14931))
* **New Data Source:** `aws_lex_slot_type` ([#8916](https://github.com/hashicorp/terraform-provider-aws/issues/8916))
* **New Data Source:** `aws_neptune_orderable_db_instance` ([#14953](https://github.com/hashicorp/terraform-provider-aws/issues/14953))
* **New Data Source:** `aws_rds_orderable_db_instance` ([#14834](https://github.com/hashicorp/terraform-provider-aws/issues/14834))
* **New Data Source:** `aws_vpc_peering_connections` ([#9491](https://github.com/hashicorp/terraform-provider-aws/issues/9491))
* **New Resource:** `aws_codebuild_report_group` ([#12573](https://github.com/hashicorp/terraform-provider-aws/issues/12573))
* **New Resource:** `aws_db_proxy` ([#12704](https://github.com/hashicorp/terraform-provider-aws/issues/12704))
* **New Resource:** `aws_emr_instance_fleet` ([#14813](https://github.com/hashicorp/terraform-provider-aws/issues/14813))
* **New Resource:** `aws_glue_user_defined_function` ([#12537](https://github.com/hashicorp/terraform-provider-aws/issues/12537))
* **New Resource:** `aws_guardduty_filter` ([#14876](https://github.com/hashicorp/terraform-provider-aws/issues/14876))
* **New Resource:** `aws_lex_slot_type` ([#8916](https://github.com/hashicorp/terraform-provider-aws/issues/8916))

ENHANCEMENTS

* data-source/aws_cur_report_definition: Add `refresh_closed_reports` and `report_versioning` attributes ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* data-source/aws_outposts_outpost: Add `arn` argument ([#14967](https://github.com/hashicorp/terraform-provider-aws/issues/14967))
* data-source/aws_route: Add `local_gateway_id` attribute ([#14864](https://github.com/hashicorp/terraform-provider-aws/issues/14864))
* data-source/aws_route_table: Add `route` `local_gateway_id` attribute ([#14864](https://github.com/hashicorp/terraform-provider-aws/issues/14864))
* resource/aws_acm_certificate: Provide additional plan-time validation for `subject_alternative_names` argument values ([#14782](https://github.com/hashicorp/terraform-provider-aws/issues/14782))
* resource/aws_ami: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/hashicorp/terraform-provider-aws/issues/14906))
* resource/aws_autoscaling_group: Support provider-level `ignore_tags` configuration ([#13868](https://github.com/hashicorp/terraform-provider-aws/issues/13868))
* resource/aws_cloudtrail: Add `insight_selector` configuration block ([#12390](https://github.com/hashicorp/terraform-provider-aws/issues/12390))
* resource/aws_cur_report_definition: Add `refresh_closed_reports` and `report_versioning` arguments ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `ATHENA` value in `additional_artifacts` argument plan-time validation ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `Parquet` value in `compression` and `format` argument plan-time validations ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* resource/aws_cur_report_definition: Support `MONTHLY` value in `time_unit` argument plan-time validation ([#12428](https://github.com/hashicorp/terraform-provider-aws/issues/12428))
* resource/aws_ebs_volume: Support io2 type ([#14894](https://github.com/hashicorp/terraform-provider-aws/issues/14894))
* resource/aws_ec2_client_vpn_endpoint: Support `authentication_options` `type` argument `federated-authentication` value and new `saml_provider_arn` argument ([#14171](https://github.com/hashicorp/terraform-provider-aws/issues/14171))
* resource/aws_emr_cluster: Add `core_instance_fleet` and `master_instance_fleet` configuration blocks ([#14788](https://github.com/hashicorp/terraform-provider-aws/issues/14788))
* resource/aws_instance: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/hashicorp/terraform-provider-aws/issues/14906))
* resource/aws_kinesis_firehose_delivery_stream: Add `elasticsearch_configuration` `vpc_config` configuration block ([#13269](https://github.com/hashicorp/terraform-provider-aws/issues/13269))
* resource/aws_kinesis_firehose_delivery_stream: Add `elasticsearch_configuration` `cluster_endpoint` argument ([#12484](https://github.com/hashicorp/terraform-provider-aws/issues/12484))
* resource/aws_kinesis_firehose_delivery_stream: Add various plan-time validations for arguments ([#12484](https://github.com/hashicorp/terraform-provider-aws/issues/12484))
* resource/aws_launch_template: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/hashicorp/terraform-provider-aws/issues/14906))
* resource/aws_msk_configuration: Support resource in-place updates and deletion ([#14826](https://github.com/hashicorp/terraform-provider-aws/issues/14826))
* resource/aws_route: Add `local_gateway_id` argument ([#14864](https://github.com/hashicorp/terraform-provider-aws/issues/14864))
* resource/aws_route_table: Add `route` `local_gateway_id` argument ([#14864](https://github.com/hashicorp/terraform-provider-aws/issues/14864))
* resource/aws_spot_fleet_request: Support `io2` value for `volume_type` argument plan-time validation ([#14906](https://github.com/hashicorp/terraform-provider-aws/issues/14906))
* resource/aws_wafv2_rule_group: Add `ip_set_forwarded_ip_config` configuration block to `ip_set_reference_statement` ([#14902](https://github.com/hashicorp/terraform-provider-aws/issues/14902))
* resource/aws_wafv2_web_acl: Add `ip_set_forwarded_ip_config` configuration block to `ip_set_reference_statement` ([#14902](https://github.com/hashicorp/terraform-provider-aws/issues/14902))

BUG FIXES

* resource/aws_autoscaling_group: Prevent unnecessary tag removal and recreation within tag updates ([#13868](https://github.com/hashicorp/terraform-provider-aws/issues/13868))
* resource/aws_cloudfront_distribution: Prevent panic with missing `ForwardedValues` ([#14993](https://github.com/hashicorp/terraform-provider-aws/issues/14993))
* resource/aws_dynamodb_table: Properly update `global_secondary_index` `non_key_attributes` values ([#9988](https://github.com/hashicorp/terraform-provider-aws/issues/9988))
* resource/aws_emr_cluster: Prevent recreation when `ebs_config.volumes_per_instance` is greater than 1 ([#14858](https://github.com/hashicorp/terraform-provider-aws/issues/14858))
* resource/aws_lambda_function_event_invoke_config: Prevent unexpected format of function resource error ([#14851](https://github.com/hashicorp/terraform-provider-aws/issues/14851))
* resource/aws_lightsail_instance: Prevent panic with key-only tags ([#13868](https://github.com/hashicorp/terraform-provider-aws/issues/13868))
* resource/aws_mq_configuration: Prevent additional revision creation with `tags` only updates ([#14850](https://github.com/hashicorp/terraform-provider-aws/issues/14850))
* resource/aws_opsworks_stack: Suppress equivalent `custom_json` differences ([#14886](https://github.com/hashicorp/terraform-provider-aws/issues/14886))
* resource/aws_rds_cluster_endpoint: Increase creation timeout to 30 minutes ([#14862](https://github.com/hashicorp/terraform-provider-aws/issues/14862))
* resource/aws_route53_resolver_rule: Correct handling for single period (`.`) value in `domain_name` argument ([#15015](https://github.com/hashicorp/terraform-provider-aws/issues/15015))
* resource/aws_route53_zone_association: Correctly handle zones with over 100 VPC associations ([#14885](https://github.com/hashicorp/terraform-provider-aws/issues/14885))
* resource/aws_waf_rate_based_rule: Properly update `rate_limit` value ([#14964](https://github.com/hashicorp/terraform-provider-aws/issues/14964))
* resource/aws_workspaces_workspace: Prevent error when `workspace_properties` `running_mode` is set to `ALWAYS_ON` ([#13976](https://github.com/hashicorp/terraform-provider-aws/issues/13976))

## 3.4.0 (August 27, 2020)

FEATURES

* **New Data Source:** `aws_db_subnet_group` ([#9525](https://github.com/hashicorp/terraform-provider-aws/issues/9525))
* **New Resource:** `aws_emr_managed_scaling_policy` ([#13965](https://github.com/hashicorp/terraform-provider-aws/issues/13965))
* **New Resource:** `aws_guardduty_publishing_destination` ([#13894](https://github.com/hashicorp/terraform-provider-aws/issues/13894))
* **New Resource:** `aws_securityhub_action_target` ([#10493](https://github.com/hashicorp/terraform-provider-aws/issues/10493))
* **New Resource:** `aws_xray_encryption_config` ([#13600](https://github.com/hashicorp/terraform-provider-aws/issues/13600))
* **New Resource:** `aws_xray_group` ([#13597](https://github.com/hashicorp/terraform-provider-aws/issues/13597))

ENHANCEMENTS

* resource/aws_apigatewayv2_integration: Add `integration_subtype` argument (Support AWS service integrations for HTTP APIs) ([#14860](https://github.com/hashicorp/terraform-provider-aws/issues/14860))
* resource/aws_elasticache_replication_group: Add plan-time validation for `notification_topic_arn` and `snapshot_arns` arguments ([#12974](https://github.com/hashicorp/terraform-provider-aws/issues/12974))
* resource/aws_globalaccelerator_endpoint_group: Add `client_ip_preservation_enabled` argument to the `endpoint_configuration` configuration block ([#14486](https://github.com/hashicorp/terraform-provider-aws/issues/14486))
* resource/aws_storagegateway_cached_iscsi_volume: Add `kms_encrypted` and `kms_key` arguments ([#12066](https://github.com/hashicorp/terraform-provider-aws/issues/12066))
* resource/aws_storagegateway_gateway: Add `smb_security_strategy` argument ([#13563](https://github.com/hashicorp/terraform-provider-aws/issues/13563))
* resource/aws_storagegateway_gateway: Add plan-time validation for `gateway_ip_address` argument ([#13563](https://github.com/hashicorp/terraform-provider-aws/issues/13563))
* resource/aws_storagegateway_gateway: Add `average_download_rate_limit_in_bits_per_sec` and `average_upload_rate_limit_in_bits_per_sec` arguments ([#13568](https://github.com/hashicorp/terraform-provider-aws/issues/13568))
* resource/aws_storagegateway_nfs_file_share: Add `cache_attributes` configuration block ([#14759](https://github.com/hashicorp/terraform-provider-aws/issues/14759))
* resource/aws_storagegateway_nfs_file_share: Support `S3_INTELLIGENT_TIERING` value in `default_storage_class` argument plan-time validation ([#14759](https://github.com/hashicorp/terraform-provider-aws/issues/14759))
* resource/aws_storagegateway_smb_file_share: Add `cache_attributes` configuration block and `case_sensitivity` argument ([#14790](https://github.com/hashicorp/terraform-provider-aws/issues/14790))
* resource/aws_storagegateway_smb_file_share: Support `S3_INTELLIGENT_TIERING` value in `default_storage_class` argument plan-time validation ([#14790](https://github.com/hashicorp/terraform-provider-aws/issues/14790))
* resource/aws_xray_sampling_rule: Add `tags` argument ([#14831](https://github.com/hashicorp/terraform-provider-aws/issues/14831))

BUG FIXES

* resource/aws_acmpca_certificate_authority: Ensure `DELETED` status triggers state removal ([#13684](https://github.com/hashicorp/terraform-provider-aws/issues/13684))
* resource/aws_appmesh_virtual_node: Prevent panics with empty `backend` configuration blocks ([#14074](https://github.com/hashicorp/terraform-provider-aws/issues/14074))
* resource/aws_cloudfront_distribution: Preview panics during resource import with empty `forwarded_values.query_string` ([#14844](https://github.com/hashicorp/terraform-provider-aws/issues/14844))
* resource/aws_elasticache_replication_group: Ensure `tags` are stored in Terraform state and properly updated ([#12974](https://github.com/hashicorp/terraform-provider-aws/issues/12974))
* resource/aws_emr_instance_group: Increase creation and update timeout to 30 minutes ([#13077](https://github.com/hashicorp/terraform-provider-aws/issues/13077)] / [[#14106](https://github.com/hashicorp/terraform-provider-aws/issues/14106))
* resource/aws_globalaccelerator_accelerator: Increase creation timeout to 10 minutes ([#14486](https://github.com/hashicorp/terraform-provider-aws/issues/14486))
* resource/aws_globalaccelerator_endpoint_group: Prevent differences with `health_check_path` defaults ([#14486](https://github.com/hashicorp/terraform-provider-aws/issues/14486))
* resource/aws_glue_crawler: Properly update `schedule` value ([#14792](https://github.com/hashicorp/terraform-provider-aws/issues/14792))

## 3.3.0 (August 20, 2020)

ENHANCEMENTS

* data-source/aws_lambda_layer_version: Support `java8.al2` and `provided.al2` in `runtime` argument plan-time validation ([#14663](https://github.com/hashicorp/terraform-provider-aws/issues/14663))
* provider: Support for appending information to User-Agent request headers with the `TF_APPEND_USER_AGENT` environment variable ([#14555](https://github.com/hashicorp/terraform-provider-aws/issues/14555))
* resource/aws_apigatewayv2_api: Add `body` argument ([#12567](https://github.com/hashicorp/terraform-provider-aws/issues/12567))
* resource/aws_customer_gateway: Support tag on create ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_dms_replication_instance: Add `allow_major_version_upgrade` argument ([#14550](https://github.com/hashicorp/terraform-provider-aws/issues/14550))
* resource/aws_ec2_client_vpn_network_association: Allow specifying custom security groups ([#14146](https://github.com/hashicorp/terraform-provider-aws/issues/14146))
* resource/aws_ec2_client_vpn_network_association: Support resource import ([#14146](https://github.com/hashicorp/terraform-provider-aws/issues/14146))
* resource/aws_egress_only_intrenet_gateway:-Ssupport tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_eks_node_group: Support `AL2_ARM_64` value for `ami_type` argument plan-time validation ([#14729](https://github.com/hashicorp/terraform-provider-aws/issues/14729))
* resource/aws_eks_node_group: Add `launch_template` configuration block ([#14639](https://github.com/hashicorp/terraform-provider-aws/issues/14639))
* resource/aws_internet_gateway: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_lambda_function: Support `java8.al2` and `provided.al2` in `runtime` argument plan-time validation ([#14663](https://github.com/hashicorp/terraform-provider-aws/issues/14663))
* resource/aws_lambda_layer_version: Support `java8.al2` and `provided.al2` in `compatible_runtimes` argument plan-time validation ([#14663](https://github.com/hashicorp/terraform-provider-aws/issues/14663))
* resource/aws_launch_template: Support `elastic-gpu` and `spot-instances-request` in `tag_specifications` `resource_type` argument plan-time validation ([#14662](https://github.com/hashicorp/terraform-provider-aws/issues/14662))
* resource/aws_network_acl: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_network_interface: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_route_table: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_security_group: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_spot_instance_request: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_storagegatway_smb_file_share: Add `audit_destination_arn` and `smb_acl_enabled` arguments ([#13572](https://github.com/hashicorp/terraform-provider-aws/issues/13572))
* resource/aws_subnet: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_subnet: Add plan-time validation to `ipv6_cidr_block` argument ([#12303](https://github.com/hashicorp/terraform-provider-aws/issues/12303))
* resource/aws_vpc_dhcp_options: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_vpc_peering_connection: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_vpn_connection: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_vpn_gateway: Support tag on create  ([#14501](https://github.com/hashicorp/terraform-provider-aws/issues/14501))
* resource/aws_wafv2_rule_group: Add `forwarded_ip_config` configuration block to `geo_match_statement` ([#14685](https://github.com/hashicorp/terraform-provider-aws/issues/14685))
* resource/aws_wafv2_web_acl: Add `forwarded_ip_config` configuration block to `rate_based_statement` and `geo_match_statement` ([#14685](https://github.com/hashicorp/terraform-provider-aws/issues/14685))
* resource/aws_wafv2_web_acl: Support `FORWARDED_IP` value for `rate_based_statement` `aggregate_key_type` argument plan-time validation ([#14685](https://github.com/hashicorp/terraform-provider-aws/issues/14685))

BUG FIXES

* resource/aws_api_gateway_vpc_link: Increase create, update, and delete timeouts to 20 minutes ([#10407](https://github.com/hashicorp/terraform-provider-aws/issues/10407))
* resource/aws_apigatewayv2_stage: Set `execution_arn` attribute for HTTP APIs ([#14638](https://github.com/hashicorp/terraform-provider-aws/issues/14638))
* resource/aws_db_parameter_group: Restore ability to update `parameter` configuration values ([#12112](https://github.com/hashicorp/terraform-provider-aws/issues/12112))
* resource/aws_user_pool_domain: Ensure state removal when deleted outside Terraform ([#14732](https://github.com/hashicorp/terraform-provider-aws/issues/14732))
* resource/aws_rds_cluster_parameter_group: Restore ability to update `parameter` configuration values ([#12112](https://github.com/hashicorp/terraform-provider-aws/issues/12112))
* resource/aws_ssm_parameter: Handle retries after creation for asynchronous `data_type` validation process ([#14514](https://github.com/hashicorp/terraform-provider-aws/issues/14514))
* resource/aws_storagegateway_nfs_file_share: Skip `UpdateSMBFileShare` API call when only `tags` change and remove extraneous `ListTagsForResource` API call during read ([#13590](https://github.com/hashicorp/terraform-provider-aws/issues/13590))
* resource/aws_subnet: Ensure `ipv6_cidr_block` argument performs removal when removed from configuration ([#12303](https://github.com/hashicorp/terraform-provider-aws/issues/12303))

## 3.2.0 (August 14, 2020)

ENHANCEMENTS

* data-source/aws_launch_configuration: Add `ebs_block_device` `no_device` attribute ([#14583](https://github.com/hashicorp/terraform-provider-aws/issues/14583))
* data-source/aws_lb: Add `subnet_mapping` `private_ipv4_address` attribute ([#14545](https://github.com/hashicorp/terraform-provider-aws/issues/14545))
* provider: Upgrade to Terraform Plugin SDK V2. There should be no breaking changes from a practitioner's perspective. Some validation errors should now feature enhanced messaging. ([#14432](https://github.com/hashicorp/terraform-provider-aws/issues/14432))
* resource/aws_accessanalyzer_analyzer: Support `ORGANIZATION` value in `type` argument ([#14493](https://github.com/hashicorp/terraform-provider-aws/issues/14493))
* resource/aws_codebuild_project: Support `WINDOWS_SERVER_2019_CONTAINER` value in `environment` `type` argument plan-time validation ([#14532](https://github.com/hashicorp/terraform-provider-aws/issues/14532))
* resource/aws_organizations_organization: Support `AISERVICES_OPT_OUT_POLICY` value in `enabled_policy_types` argument plan-time validation (Support AI Opt Out policies) ([#14650](https://github.com/hashicorp/terraform-provider-aws/issues/14650))
* resource/aws_organizations_policy: Support `AISERVICES_OPT_OUT_POLICY` value in `type` argument plan-time validation (Support AI Opt Out policies) ([#14528](https://github.com/hashicorp/terraform-provider-aws/issues/14528))
* resource/aws_route53_health_check: Add `disabled` argument ([#14614](https://github.com/hashicorp/terraform-provider-aws/issues/14614))

BUG FIXES

* data-source/aws_launch_template: Prevent type error with `network_interfaces` `delete_on_termination` attribute ([#14599](https://github.com/hashicorp/terraform-provider-aws/issues/14599))
* resource/aws_acm_certificate_validation: Prevent panic with missing `DomainValidationOptions` `ResourceRecord` attribute in API response [[#14590](https://github.com/hashicorp/terraform-provider-aws/issues/14590)]
* resource/aws_ecr_repository: Prevent panic with missing `EncryptionConfiguration` attribute in API response ([#14584](https://github.com/hashicorp/terraform-provider-aws/issues/14584))
* resource/aws_wafv2_rule_group: Prevent unnecessary resource recreation with `rule` updates ([#14617](https://github.com/hashicorp/terraform-provider-aws/issues/14617))
* resource/aws_wafv2_web_acl: Prevent unnecessary resource recreation with `rule` updates ([#14616](https://github.com/hashicorp/terraform-provider-aws/issues/14616))

## 3.1.0 (August 07, 2020)

NOTES:

* resource/aws_route53_zone_association: The addition of cross-account zone association support required the use of new `ListHostedZonesByVPC` API call and adding the VPC Region to the resource ID for new resources. Restrictive IAM permissions for Terraform and cross-region imports may require updates. ([#14215](https://github.com/hashicorp/terraform-provider-aws/issues/14215))

FEATURES

* **New Data Source:** `aws_ec2_spot_price` ([#12504](https://github.com/hashicorp/terraform-provider-aws/issues/12504))
* **New Resource**: `aws_route53_vpc_association_authorization` ([#14215](https://github.com/hashicorp/terraform-provider-aws/issues/14215))

ENHANCEMENTS

* data-source/aws_ecr_repository: Allow `registry_id` as an argument ([#14368](https://github.com/hashicorp/terraform-provider-aws/issues/14368))
* data-source/aws_ecr_repository: Add `image_scanning_configuration` and `image_tag_mutability` attributes ([#14368](https://github.com/hashicorp/terraform-provider-aws/issues/14368))
* data-source/aws_ecr_repository: Add `encryption_configuration` attribute ([#14520](https://github.com/hashicorp/terraform-provider-aws/issues/14520))
* resource/aws_api_gateway_method_settings: Plan-time validation added to `settings` `unauthorized_cache_control_header_strategy` and `logging_level` arguments ([#12651](https://github.com/hashicorp/terraform-provider-aws/issues/12651))
* resource/aws_ecr_repository: Add `encryption_configuration` attribute ([#14520](https://github.com/hashicorp/terraform-provider-aws/issues/14520))
* resource/aws_lb: Add `subnet_mapping` configuration block `private_ipv4_address` argument ([#11404](https://github.com/hashicorp/terraform-provider-aws/issues/11404))
* resource/aws_rds_global_cluster: Add `force_destroy` and `source_db_cluster_identifier` arguments ([#14487](https://github.com/hashicorp/terraform-provider-aws/issues/14487))
* resource/aws_rds_global_cluster: Add `global_cluster_members` attribute ([#14487](https://github.com/hashicorp/terraform-provider-aws/issues/14487))
* resource/aws_route53_zone_association: Cross-account zone associations can now be created in conjunction with the new `aws_route53_vpc_association_authorization` resource ([#14215](https://github.com/hashicorp/terraform-provider-aws/issues/14215))
* resource/aws_ssm_parameter: Add `data_type` argument (support `aws:ec2:image` parameters) ([#13326](https://github.com/hashicorp/terraform-provider-aws/issues/13326))

BUG FIXES

* data-source/aws_availability_zones: Prevent unexpected plan output every apply with `group_names` attribute ([#14412](https://github.com/hashicorp/terraform-provider-aws/issues/14412))
* data-source/aws_s3_bucket: Ensure provider `s3_force_path_style` configuration is passed through for getting S3 Bucket location with non-AWS implementations ([#14481](https://github.com/hashicorp/terraform-provider-aws/issues/14481))
* resource/aws_api_gateway_method_settings: Allow `settings` `cache_ttl_in_seconds` argument to be set to 0 ([#12651](https://github.com/hashicorp/terraform-provider-aws/issues/12651))
* resource/aws_elastictranscoder_preset: Prevent empty configuration block panics ([#14092](https://github.com/hashicorp/terraform-provider-aws/issues/14092))
* resource/aws_lambda_event_source_mapping: Allow `maximum_retry_attempts` argument to be set to 0 ([#12479](https://github.com/hashicorp/terraform-provider-aws/issues/12479))
* resource/aws_rds_cluster: Add an `InvalidDBClusterStateFault` retryable error condition for clusters part of a global cluster ([#14420](https://github.com/hashicorp/terraform-provider-aws/issues/14420))
* resource/aws_rds_cluster: Increase retry timeout for deletion to 2 minutes ([#14420](https://github.com/hashicorp/terraform-provider-aws/issues/14420))
* resource/aws_rds_cluster: Prevent error when both `global_cluster_identifier` and `replication_source_identifier` are configured on creation ([#14490](https://github.com/hashicorp/terraform-provider-aws/issues/14490))
* resource/aws_s3_bucket: Ensure provider `s3_force_path_style` configuration is passed through for getting S3 Bucket location with non-AWS implementations ([#14481](https://github.com/hashicorp/terraform-provider-aws/issues/14481))
* resource/aws_secretsmanager_secret: Allow retries for IAM eventual consistency errors ([#14459](https://github.com/hashicorp/terraform-provider-aws/issues/14459))
* resource/aws_security_group: Ensure `name_prefix` argument with hex digits `a` through `f` is properly imported ([#14475](https://github.com/hashicorp/terraform-provider-aws/issues/14475))
* resource/aws_spot_fleet_request: Allow `target_capacity` argument to be updated to 0 ([#12759](https://github.com/hashicorp/terraform-provider-aws/issues/12759))
* resource/aws_spot_fleet_request: Wait for modify operation completion (default timeout of 10 minutes) ([#12759](https://github.com/hashicorp/terraform-provider-aws/issues/12759))
* resource/aws_vpc_dhcp_options_association: Properly trigger resource recreation when VPC is deleted outside Terraform ([#14367](https://github.com/hashicorp/terraform-provider-aws/issues/14367))

## 3.0.0 (July 31, 2020)

NOTES:
* provider: This version is built using Go 1.14.5, including security fixes to the crypto/x509 and net/http packages.

BREAKING CHANGES

* provider: New versions of the provider can only be automatically installed on Terraform 0.12 and later ([#14143](https://github.com/hashicorp/terraform-provider-aws/issues/14143))
* provider: All "removed" attributes are cut, using them would result in a Terraform Core level error ([#14001](https://github.com/hashicorp/terraform-provider-aws/issues/14001))
* provider: Credential ordering has changed from static, environment, shared credentials, EC2 metadata, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) to static, environment, shared credentials, default AWS Go SDK (shared configuration, web identity, ECS, EC2 Metadata) ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: The `AWS_METADATA_TIMEOUT` environment variable no longer has any effect as we now depend on the default AWS Go SDK EC2 Metadata client timeout of one second with two retries ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: Remove deprecated `kinesis_analytics` and `r53` custom service endpoint arguments ([#14238](https://github.com/hashicorp/terraform-provider-aws/issues/14238))
* data-source/aws_availability_zones: Remove deprecated `blacklisted_names` and `blacklisted_zone_ids` arguments ([#14134](https://github.com/hashicorp/terraform-provider-aws/issues/14134))
* data-source/aws_directory_service_directory: Return an error when a single result is not found ([#14006](https://github.com/hashicorp/terraform-provider-aws/issues/14006))
* data-source/aws_ecr_repository: Return an error when a single result is not found ([#10520](https://github.com/hashicorp/terraform-provider-aws/issues/10520))
* data-source/aws_efs_file_system: Return an error when a single result is not found ([#14005](https://github.com/hashicorp/terraform-provider-aws/issues/14005))
* data-source/aws_launch_template: Return an error when a single result is not found ([#10521](https://github.com/hashicorp/terraform-provider-aws/issues/10521))
* data-source/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in data-source ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* data-source/aws_route53_zone: Trailing period removed from `name` argument set in data-source ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_acm_certificate: `certificate_body`, `certificate_chain`, and `private_key` attributes are no longer stored in the Terraform state with hash values ([#9685](https://github.com/hashicorp/terraform-provider-aws/issues/9685))
* resource/aws_acm_certificate: `domain_validation_options` attribute changed from list to set ([#14199](https://github.com/hashicorp/terraform-provider-aws/issues/14199))
* resource/aws_acm_certificate: Plan-time validation added to `domain_name` and `subject_alternative_names` arguments to prevent usage of strings with trailing periods ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_api_gateway_method_settings: Remove `Computed` property from `throttling_burst_limit` and `throttling_rate_limit` arguments, enabling drift detection ([#14266](https://github.com/hashicorp/terraform-provider-aws/issues/14266))
* resource/aws_api_gateway_method_settings: Update `throttling_burst_limit` and `throttling_rate_limit` argument defaults to match API default of `-1` to keep throttling disabled ([#14266](https://github.com/hashicorp/terraform-provider-aws/issues/14266))
* resource/aws_autoscaling_group: `availability_zones` and `vpc_zone_identifier` argument conflict now reported at plan-time ([#12927](https://github.com/hashicorp/terraform-provider-aws/issues/12927))
* resource/aws_autoscaling_group: Remove `Computed` property from `load_balancers` and `target_group_arns` arguments, enabling drift detection ([#14064](https://github.com/hashicorp/terraform-provider-aws/issues/14064))
* resource/aws_cloudfront_distribution: `active_trusted_signers` argument renamed to `trusted_signers` to support accessing `items` in Terraform 0.12 ([#14339](https://github.com/hashicorp/terraform-provider-aws/issues/14339))
* resource/aws_cloudwatch_log_group: Automatically trim `:*` suffix from `arn` attribute ([#14214](https://github.com/hashicorp/terraform-provider-aws/issues/14214))
* resource/aws_codepipeline: Removes `GITHUB_TOKEN` environment variable ([#14175](https://github.com/hashicorp/terraform-provider-aws/issues/14175))
* resource/aws_cognito_user_pool: Remove deprecated `admin_create_user_config` configuration block `unused_account_validity_days` argument ([#14294](https://github.com/hashicorp/terraform-provider-aws/issues/14294))
* resource/aws_dx_gateway: Remove automatic `aws_dx_gateway_association` resource import ([#14124](https://github.com/hashicorp/terraform-provider-aws/issues/14124))
* resource/aws_dx_gateway_association: Remove deprecated `vpn_gateway_id` argument ([#14144](https://github.com/hashicorp/terraform-provider-aws/issues/14144))
* resource/aws_dx_gateway_association_proposal: Remove deprecated `vpn_gateway_id` argument ([#14144](https://github.com/hashicorp/terraform-provider-aws/issues/14144))
* resource/aws_ebs_volume: Return an error when `iops` argument set to a value greater than 0 for volume types other than `io1` ([#14310](https://github.com/hashicorp/terraform-provider-aws/issues/14310))
* resource/aws_elastic_transcoder_preset: Remove `video` configuration block `max_frame_rate` argument default value ([#7141](https://github.com/hashicorp/terraform-provider-aws/issues/7141))
* resource/aws_emr_cluster: Remove deprecated `instance_group` configuration block, `core_instance_count`, `core_instance_type`, and `master_instance_type` arguments ([#14137](https://github.com/hashicorp/terraform-provider-aws/issues/14137))
* resource/aws_glue_job: Remove deprecated `allocated_capacity` argument ([#14296](https://github.com/hashicorp/terraform-provider-aws/issues/14296))
* resource/aws_iam_access_key: Remove deprecated `ses_smtp_password` attribute ([#14299](https://github.com/hashicorp/terraform-provider-aws/issues/14299))
* resource/aws_iam_instance_profile: Remove deprecated `roles` argument ([#14303](https://github.com/hashicorp/terraform-provider-aws/issues/14303))
* resource/aws_iam_server_certificate: Remove state hashing from `certificate_body`, `certificate_chain`, and `private_key` arguments for new or recreated resources ([#14187](https://github.com/hashicorp/terraform-provider-aws/issues/14187))
* resource/aws_instance: Return an error when `ebs_block_device` `iops` or `root_block_device` `iops` argument set to a value greater than `0` for volume types other than `io1` ([#14310](https://github.com/hashicorp/terraform-provider-aws/issues/14310))
* resource/aws_lambda_alias: Resource import no longer converts Lambda Function name to ARN ([#12876](https://github.com/hashicorp/terraform-provider-aws/issues/12876))
* resource/aws_launch_template: `network_interfaces` `delete_on_termination` argument changed from `bool` to `string` type ([#8612](https://github.com/hashicorp/terraform-provider-aws/issues/8612))
* resource/aws_lb_listener_rule: Remove deprecated `condition` configuration block `field` and `values` arguments ([#14309](https://github.com/hashicorp/terraform-provider-aws/issues/14309))
* resource/aws_msk_cluster: Update `encryption_info` `encryption_in_transit` `client_broker` argument default to match API default of `TLS` ([#14132](https://github.com/hashicorp/terraform-provider-aws/issues/14132))
* resource/aws_rds_cluster: Update `scaling_configuration` `min_capacity` argument default to match API default of `1` ([#14268](https://github.com/hashicorp/terraform-provider-aws/issues/14268))
* resource/aws_route53_resolver_rule: Trailing period removed from `domain_name` argument set in resource ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_route53_zone: Trailing period removed from `name` argument set in resource ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_s3_bucket: Remove automatic `aws_s3_bucket_policy` resource import ([#14121](https://github.com/hashicorp/terraform-provider-aws/issues/14121))
* resource/aws_s3_bucket: Convert `region` to read-only attribute ([#14127](https://github.com/hashicorp/terraform-provider-aws/issues/14127))
* resource/aws_s3_bucket_metric: Update `filter` argument to require at least one of the `prefix` or `tags` nested arguments ([#14230](https://github.com/hashicorp/terraform-provider-aws/issues/14230))
* resource/aws_security_group: Remove automatic `aws_security_group_rule` resource import ([#12616](https://github.com/hashicorp/terraform-provider-aws/issues/12616))
* resource/aws_ses_domain_identity: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_ses_domain_identity_verification: Plan-time validation added to `domain` argument to prevent usage of strings with trailing periods ([#14220](https://github.com/hashicorp/terraform-provider-aws/issues/14220))
* resource/aws_sns_platform_application: `platform_credential` and `platform_principal` attributes are no longer stored in the Terraform state with hash values ([#3894](https://github.com/hashicorp/terraform-provider-aws/issues/3894))
* resource/aws_spot_fleet_request: Remove 24 hour default for `valid_until` argument ([#9718](https://github.com/hashicorp/terraform-provider-aws/issues/9718))
* resource/aws_ssm_maintenance_window_task: Remove deprecated `logging_info` and `task_parameters` configuration blocks ([#14311](https://github.com/hashicorp/terraform-provider-aws/issues/14311))

FEATURES

* **New Data Source:** aws_workspaces_directory ([#13529](https://github.com/hashicorp/terraform-provider-aws/issues/13529))

ENHANCEMENTS

* provider: Always enable shared configuration file support (no longer require `AWS_SDK_LOAD_CONFIG` environment variable) ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: Add `assume_role` configuration block `duration_seconds`, `policy_arns`, `tags`, and `transitive_tag_keys` arguments ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* data-source/aws_instance: Add `secondary_private_ips` attribute ([#14079](https://github.com/hashicorp/terraform-provider-aws/issues/14079))
* data-source/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support ([#14221](https://github.com/hashicorp/terraform-provider-aws/issues/14221))
* resource/aws_acm_certificate: Enable `domain_validation_options` usage in downstream resource `count` and `for_each` references ([#14199](https://github.com/hashicorp/terraform-provider-aws/issues/14199))
* resource/aws_api_gateway_authorizer: Add plan-time validation to `authorizer_credentials` argument ([#12643](https://github.com/hashicorp/terraform-provider-aws/issues/12643))
* resource/aws_api_gateway_method_settings: Add import support ([#14266](https://github.com/hashicorp/terraform-provider-aws/issues/14266))
* resource/aws_apigatewayv2_integration: Add `request_parameters` attribute ([#14080](https://github.com/hashicorp/terraform-provider-aws/issues/14080))
* resource/aws_apigatewayv2_integration: Add `tls_config` attribute ([#13013](https://github.com/hashicorp/terraform-provider-aws/issues/13013))
* resource/aws_apigatewayv2_route: Support for updating route key ([#13833](https://github.com/hashicorp/terraform-provider-aws/issues/13833))
* resource/aws_apigatewayv2_stage: Make `deployment_id` a `Computed` attribute ([#13644](https://github.com/hashicorp/terraform-provider-aws/issues/13644))
* resource/aws_fsx_lustre_file_system: Add `deployment_type` and `per_unit_storage_throughput` attributes ([#13639](https://github.com/hashicorp/terraform-provider-aws/issues/13639))
* resource_aws_fsx_windows_file_system - add `storage_type` argument. ([#14316](https://github.com/hashicorp/terraform-provider-aws/issues/14316))
* resource_aws_fsx_windows_file_system: add support for multi-az ([#12676](https://github.com/hashicorp/terraform-provider-aws/issues/12676))
* resource_aws_fsx_windows_file_system: add `SINGLE_AZ_2` deployment type ([#12676](https://github.com/hashicorp/terraform-provider-aws/issues/12676))
* resource_aws_fsx_windows_file_system: adds `preferred_file_server_ip`, `remote_administration_endpoint` attributes ([#12676](https://github.com/hashicorp/terraform-provider-aws/issues/12676))
* resource/aws_instance: Add `secondary_private_ips` argument (conflicts with `network_interface` configuration block) ([#14079](https://github.com/hashicorp/terraform-provider-aws/issues/14079))

BUG FIXES

* provider: Ensure nil is not passed to RetryError helpers, may result in some bug fixes ([#14104](https://github.com/hashicorp/terraform-provider-aws/issues/14104))
* provider: Ensure configured STS endpoint is used during `AssumeRole` API calls ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: Prefer AWS shared configuration over EC2 metadata credentials by default ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* provider: Prefer CodeBuild, ECS, EKS credentials over EC2 metadata credentials by default ([#14077](https://github.com/hashicorp/terraform-provider-aws/issues/14077))
* data-source/aws_lb: `enable_http2` now properly set ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))
* resource/aws_acm_certificate: Prevent unexpected ordering differences with `domain_validation_options` attribute ([#14199](https://github.com/hashicorp/terraform-provider-aws/issues/14199))
* resource/aws_api_gateway_authorizer: Allow `authorizer_result_ttl_in_seconds` to be set to 0 ([#12643](https://github.com/hashicorp/terraform-provider-aws/issues/12643))
* resource/aws_apigatewayv2_integration: Correctly handle the `integration_method` attribute for AWS Lambda integrations([#13266](https://github.com/hashicorp/terraform-provider-aws/issues/13266))
* resource/aws_apigatewayv2_integration: Correctly handle the `passthrough_behavior` attribute for HTTP APIs ([#13062](https://github.com/hashicorp/terraform-provider-aws/issues/13062))
* resource/aws_apigatewayv2_stage: Correctly handle `default_route_setting` and `route_setting` `data_trace_enabled` and `logging_level` for HTTP APIs. `logging_level` is now `Computed`, meaning Terraform will only perform drift detection of its value when present in a configuration. ([#13809](https://github.com/hashicorp/terraform-provider-aws/issues/13809))
* resource/aws_appautoscaling_target: Only retry `DeregisterScalableTarget` retries on all errors on deletion ([#14259](https://github.com/hashicorp/terraform-provider-aws/issues/14259))
* resource/aws_dx_gateway_association: Increase default create/update/delete timeouts to 30 minutes ([#14144](https://github.com/hashicorp/terraform-provider-aws/issues/14144))
* resource/aws_codepipeline: Only retry `CreatePipeline` errors for IAM eventual consistency errors ([#14264](https://github.com/hashicorp/terraform-provider-aws/issues/14264))
* resource/aws_elasticsearch_domain: Update method to properly set `advanced_security_options` ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))
* resource/aws_lambda_function: Increase IAM retry timeout for creation to standard 2 minute timeout ([#14291](https://github.com/hashicorp/terraform-provider-aws/issues/14291))
* resource/aws_lb_cookie_stickiness_policy: `lb_port` now properly set ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))
* resource/aws_network_acl_rule: Immediately return `DescribeNetworkAcls` errors on creation ([#14261](https://github.com/hashicorp/terraform-provider-aws/issues/14261))
* resource/aws_s3_bucket: Replace `GetBucketLocation` API call with custom HTTP call for FIPS endpoint support ([#14221](https://github.com/hashicorp/terraform-provider-aws/issues/14221))
* resource/aws_sns_topic_subscription: Immediately return `ListSubscriptionsByTopic` errors ([#14262](https://github.com/hashicorp/terraform-provider-aws/issues/14262))
* resource/aws_spot_fleet_request: Only retry `RequestSpotFleet` on IAM eventual consistency errors and use standard 2 minute timeout ([#14265](https://github.com/hashicorp/terraform-provider-aws/issues/14265))
* resource/aws_spot_instance_request: `primary_network_interface_id` now properly set ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))
* resource/aws_ssm_activation: Only retry `CreateActivation` on IAM eventual consistency errors and use standard 2 minute timeout ([#14263](https://github.com/hashicorp/terraform-provider-aws/issues/14263))
* resource/aws_ssm_association: `parameters` now properly set ([#14167](https://github.com/hashicorp/terraform-provider-aws/issues/14167))

## Previous Releases

For information on prior major releases, see their changelogs:

* [3.74 and earlier](https://github.com/hashicorp/terraform-provider-aws/blob/release/3.x/CHANGELOG.md)
* [2.70 and earlier](https://github.com/hashicorp/terraform-provider-aws/blob/release/2.x/CHANGELOG.md)
