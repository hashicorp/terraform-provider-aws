## 1.0.0 (Unreleased)

IMPROVEMENTS:

* resource/aws_instance: New `user_data_base64` attribute that allows non-UTF8 data (such as gzip) to be assigned to user-data without corruption [GH-850]
* data-source/aws_vpc: Expose enable_dns_* in aws_vpc data_source [GH-1373]
* resource/aws_rds_cluster: Allow setting of rds cluster engine [GH-1415]
* resource/aws_ssm_association: now supports update for `parameters`, `schedule_expression`,`output_location` [GH-1421]
* resource/aws_ssm_patch_baseline: now supports update for multiple attributes [GH-1421]
* resource/aws_cloudformation_stack: Add support for Import [GH-1432]
* resource/aws_rds_cluster_instance: Expose availability_zone attribute [GH-1439]
* resource/aws_efs_file_system: Add support for encryption [GH-1420]
* resource/aws_db_parameter_group: Allow underscores in names [GH-1460]

BUG FIXES:

* resource/aws_instance: Fix `associate_public_ip_address` [GH-1340]
* resource/aws_instance: Fix import in EC2 Classic [GH-1453]
* resource/aws_emr_cluster: Avoid spurious diff of `log_uri` [GH-1374]
* resource/aws_cloudwatch_log_subscription_filter: Add support for ResourceNotFound [GH-1414]
* resource/aws_alb: Cleanup ENIs after deleting ALB [GH-1427]

## 0.1.4 (August 08, 2017)

FEATURES:

* **New Resource:** `aws_cloudwatch_dashboard` ([#1172](https://github.com/terraform-providers/terraform-provider-aws/issues/1172))
* **New Data Source:** `aws_internet_gateway` ([#1196](https://github.com/terraform-providers/terraform-provider-aws/issues/1196))
* **New Data Source:** `aws_efs_mount_target` ([#1255](https://github.com/terraform-providers/terraform-provider-aws/issues/1255))

IMPROVEMENTS:

* AWS SDK to log extra debug details on request errors ([#1210](https://github.com/terraform-providers/terraform-provider-aws/issues/1210))
* resource/aws_spot_fleet_request: Add support for  `wait_for_fulfillment` ([#1241](https://github.com/terraform-providers/terraform-provider-aws/issues/1241))
* resource/aws_autoscaling_schedule: Allow empty value ([#1268](https://github.com/terraform-providers/terraform-provider-aws/issues/1268))
* resource/aws_ssm_association: Add support for OutputLocation and Schedule Expression ([#1253](https://github.com/terraform-providers/terraform-provider-aws/issues/1253))
* resource/aws_ssm_patch_baseline: Update support for Operating System ([#1260](https://github.com/terraform-providers/terraform-provider-aws/issues/1260))
* resource/aws_db_instance: Expose db_instance ca_cert_identifier ([#1256](https://github.com/terraform-providers/terraform-provider-aws/issues/1256))
* resource/aws_rds_cluster: Add support for iam_roles to rds_cluster ([#1258](https://github.com/terraform-providers/terraform-provider-aws/issues/1258))
* resource/aws_rds_cluster_parameter_group: Support > 20 parameters ([#1298](https://github.com/terraform-providers/terraform-provider-aws/issues/1298))
* data-source/aws_iam_role: Normalize the IAM role data source ([#1330](https://github.com/terraform-providers/terraform-provider-aws/issues/1330))
* resource/aws_kinesis_stream: Increase Timeouts, add Timeout Support ([#1345](https://github.com/terraform-providers/terraform-provider-aws/issues/1345))

BUG FIXES:

* resource/aws_instance: Guard check for aws_instance UserData to prevent panic ([#1288](https://github.com/terraform-providers/terraform-provider-aws/issues/1288))
* resource/aws_config: Set AWS Config Configuration recorder & Delivery channel names as ForceNew ([#1247](https://github.com/terraform-providers/terraform-provider-aws/issues/1247))
* resource/aws_cloudtrail: Retry if IAM role isn't propagated yet ([#1312](https://github.com/terraform-providers/terraform-provider-aws/issues/1312))
* resource/aws_cloudtrail: Fix CloudWatch role ARN/group updates ([#1357](https://github.com/terraform-providers/terraform-provider-aws/issues/1357))
* resource/aws_eip_association: Avoid crash in EC2 Classic ([#1344](https://github.com/terraform-providers/terraform-provider-aws/issues/1344))
* resource/aws_elasticache_parameter_group: Allow removing parameters ([#1309](https://github.com/terraform-providers/terraform-provider-aws/issues/1309))
* resource/aws_kinesis: add retries for Kinesis throttling exceptions ([#1085](https://github.com/terraform-providers/terraform-provider-aws/issues/1085))
* resource/aws_kinesis_firehose: adding support for `ExtendedS3DestinationConfiguration` ([#1015](https://github.com/terraform-providers/terraform-provider-aws/issues/1015))
* resource/aws_spot_fleet_request: Ignore empty `key_name` ([#1203](https://github.com/terraform-providers/terraform-provider-aws/issues/1203))
* resource/aws_emr_instance_group: fix crash when changing `instance_group.count` ([#1287](https://github.com/terraform-providers/terraform-provider-aws/issues/1287))
* resource/aws_elasticsearch_domain: Fix updating config when update doesn't involve EBS ([#1131](https://github.com/terraform-providers/terraform-provider-aws/issues/1131))
* resource/aws_s3_bucket: Avoid crashing when no lifecycle rule is defined ([#1316](https://github.com/terraform-providers/terraform-provider-aws/issues/1316))
* resource/elastic_transcoder_preset: Fix provider validation ([#1338](https://github.com/terraform-providers/terraform-provider-aws/issues/1338))
* resource/aws_s3_bucket: Avoid crashing when `filter` is not set ([#1350](https://github.com/terraform-providers/terraform-provider-aws/issues/1350))

## 0.1.3 (July 25, 2017)

FEATURES:

* **New Data Source:** `aws_iam_instance_profile` ([#1024](https://github.com/terraform-providers/terraform-provider-aws/issues/1024))
* **New Data Source:** `aws_alb_target_group` ([#1037](https://github.com/terraform-providers/terraform-provider-aws/issues/1037))
* **New Data Source:** `aws_iam_group` ([#1140](https://github.com/terraform-providers/terraform-provider-aws/issues/1140))
* **New Resource:** `aws_api_gateway_request_validator` ([#1064](https://github.com/terraform-providers/terraform-provider-aws/issues/1064))
* **New Resource:** `aws_api_gateway_gateway_response` ([#1168](https://github.com/terraform-providers/terraform-provider-aws/issues/1168))
* **New Resource:** `aws_iot_policy` ([#986](https://github.com/terraform-providers/terraform-provider-aws/issues/986))
* **New Resource:** `aws_iot_certificate` ([#1225](https://github.com/terraform-providers/terraform-provider-aws/issues/1225))

IMPROVEMENTS:

* resource/aws_sqs_queue: Add support for Server-Side Encryption ([#962](https://github.com/terraform-providers/terraform-provider-aws/issues/962))
* resource/aws_vpc: Add support for classiclink_dns_support ([#1079](https://github.com/terraform-providers/terraform-provider-aws/issues/1079))
* resource/aws_lambda_function: Add support for lambda_function vpc_config update ([#1080](https://github.com/terraform-providers/terraform-provider-aws/issues/1080))
* resource/aws_lambda_function: Add support for lambda_function dead_letter_config update ([#1080](https://github.com/terraform-providers/terraform-provider-aws/issues/1080))
* resource/aws_route53_health_check: add support for health_check regions ([#1116](https://github.com/terraform-providers/terraform-provider-aws/issues/1116))
* resource/aws_spot_instance_request: add support for request launch group ([#1097](https://github.com/terraform-providers/terraform-provider-aws/issues/1097))
* resource/aws_rds_cluster_instance: Export the RDI Resource ID for the instance ([#1142](https://github.com/terraform-providers/terraform-provider-aws/issues/1142))
* resource/aws_sns_topic_subscription: Support password-protected HTTPS endpoints ([#861](https://github.com/terraform-providers/terraform-provider-aws/issues/861))

BUG FIXES:

* provider: Remove assumeRoleHash ([#1227](https://github.com/terraform-providers/terraform-provider-aws/issues/1227))
* resource/aws_ami: Retry on `InvalidAMIID.NotFound` ([#1035](https://github.com/terraform-providers/terraform-provider-aws/issues/1035))
* resource/aws_iam_server_certificate: Fix restriction on length of `name_prefix` ([#1217](https://github.com/terraform-providers/terraform-provider-aws/issues/1217))
* resource/aws_autoscaling_group: Fix handling of empty `vpc_zone_identifier` (EC2 classic & default VPC) ([#1191](https://github.com/terraform-providers/terraform-provider-aws/issues/1191))
* resource/aws_ecr_repository_policy: Add retry logic to work around IAM eventual consistency ([#1165](https://github.com/terraform-providers/terraform-provider-aws/issues/1165))
* resource/aws_ecs_service: Fixes normalization issues in placement_strategy ([#1025](https://github.com/terraform-providers/terraform-provider-aws/issues/1025))
* resource/aws_eip: Retry reading EIPs on creation ([#1053](https://github.com/terraform-providers/terraform-provider-aws/issues/1053))
* resource/aws_elastic_beanstalk_environment: Avoid spurious diffs of JSON-based `setting`s ([#901](https://github.com/terraform-providers/terraform-provider-aws/issues/901))
* resource/aws_opsworks_permission: Fix 'set permissions' failing to set ssh access ([#1038](https://github.com/terraform-providers/terraform-provider-aws/issues/1038))
* resource/aws_s3_bucket_notification: Fix missing `bucket` field after import ([#978](https://github.com/terraform-providers/terraform-provider-aws/issues/978))
* resource/aws_sfn_state_machine: Handle another NotFound exception type ([#1062](https://github.com/terraform-providers/terraform-provider-aws/issues/1062))
* resource/aws_ssm_parameter: ForceNew on ssm_parameter rename ([#1022](https://github.com/terraform-providers/terraform-provider-aws/issues/1022))
* resource/aws_instance: Update SourceDestCheck modification on new resources ([#1065](https://github.com/terraform-providers/terraform-provider-aws/issues/1065))
* resource/aws_spot_instance_request: fixed and issue with network interfaces configuration ([#1070](https://github.com/terraform-providers/terraform-provider-aws/issues/1070))
* resource/aws_rds_cluster: Modify RDS Cluster after restoring from snapshot, if required ([#926](https://github.com/terraform-providers/terraform-provider-aws/issues/926))
* resource/aws_kms_alias: Retry lookups after creation ([#1040](https://github.com/terraform-providers/terraform-provider-aws/issues/1040))
* resource/aws_internet_gateway: Retry deletion properly on `DependencyViolation` ([#1021](https://github.com/terraform-providers/terraform-provider-aws/issues/1021))
* resource/aws_elb: Cleanup ENIs after deleting ELB ([#1036](https://github.com/terraform-providers/terraform-provider-aws/issues/1036))
* resource/aws_kms_key: Retry lookups after creation ([#1039](https://github.com/terraform-providers/terraform-provider-aws/issues/1039))
* resource/aws_dms_replication_instance: Add modifying as a pending creation state ([#1114](https://github.com/terraform-providers/terraform-provider-aws/issues/1114))
* resource/aws_redshift_cluster: Trigger ForceNew aws_redshift_cluster on encrypted change ([#1120](https://github.com/terraform-providers/terraform-provider-aws/issues/1120))
* resource/aws_default_network_acl: Add support for ipv6_cidr_block ([#1113](https://github.com/terraform-providers/terraform-provider-aws/issues/1113))
* resource/aws_autoscaling_group: Suppress diffs when an empty set is specified for `availability_zones` ([#1190](https://github.com/terraform-providers/terraform-provider-aws/issues/1190))
* resource/aws_vpc: Ignore ClassicLink DNS support in unsupported regions ([#1176](https://github.com/terraform-providers/terraform-provider-aws/issues/1176))
* resource/elastic_beanstalk_configuration_template: Handle missing platform ([#1222](https://github.com/terraform-providers/terraform-provider-aws/issues/1222))
* r/elasticache_parameter_group: support more than 20 parameters ([#1221](https://github.com/terraform-providers/terraform-provider-aws/issues/1221))
* data-source/aws_db_instance: Fix the output of subnet_group_name ([#1141](https://github.com/terraform-providers/terraform-provider-aws/issues/1141))
* data-source/aws_iam_server_certificate: Fix restriction on length of `name_prefix` ([#1217](https://github.com/terraform-providers/terraform-provider-aws/issues/1217))

## 0.1.2 (June 30, 2017)

FEATURES:

* **New Resource**: `aws_network_interface_sg_attachment` ([#860](https://github.com/terraform-providers/terraform-provider-aws/issues/860))
* **New Data Source**: `aws_ecr_repository` ([#944](https://github.com/terraform-providers/terraform-provider-aws/issues/944))

IMPROVEMENTS:

* Added ability to change the deadline for the EC2 metadata API endpoint ([#950](https://github.com/terraform-providers/terraform-provider-aws/issues/950))
* resource/aws_api_gateway_integration: Add support for specifying cache key parameters ([#893](https://github.com/terraform-providers/terraform-provider-aws/issues/893))
* resource/aws_cloudwatch_event_target: Add ecs_target ([#977](https://github.com/terraform-providers/terraform-provider-aws/issues/977))
* resource/aws_vpn_connection: Add BGP related information on aws_vpn_connection ([#973](https://github.com/terraform-providers/terraform-provider-aws/issues/973))
* resource/aws_cloudformation_stack: Add timeout support ([#994](https://github.com/terraform-providers/terraform-provider-aws/issues/994))
* resource/aws_ssm_parameter: Add support for ssm parameter overwrite ([#1006](https://github.com/terraform-providers/terraform-provider-aws/issues/1006))
* resource/aws_codebuild_project: Add support for environment privileged_mode [GH1009]
* resource/aws_dms_endpoint: Add support for dynamodb as an endpoint target ([#1002](https://github.com/terraform-providers/terraform-provider-aws/issues/1002))
* resource/aws_s3_bucket: Support lifecycle tags filter ([#899](https://github.com/terraform-providers/terraform-provider-aws/issues/899))
* resource/aws_s3_bucket_object: Allow to set WebsiteRedirect on S3 object ([#1020](https://github.com/terraform-providers/terraform-provider-aws/issues/1020))

BUG FIXES:

* resource/aws_waf: Only set FieldToMatch.Data if not empty ([#953](https://github.com/terraform-providers/terraform-provider-aws/issues/953))
* resource/aws_elastic_beanstalk_application_version: Scope labels to application ([#956](https://github.com/terraform-providers/terraform-provider-aws/issues/956))
* resource/aws_s3_bucket: Allow use of `days = 0` with lifecycle transition ([#957](https://github.com/terraform-providers/terraform-provider-aws/issues/957))
* resource/aws_ssm_maintenance_window_task: Make task_parameters updateable on aws_ssm_maintenance_window_task resource ([#965](https://github.com/terraform-providers/terraform-provider-aws/issues/965))
* resource/aws_kinesis_stream: don't force stream destroy on shard_count update ([#894](https://github.com/terraform-providers/terraform-provider-aws/issues/894))
* resource/aws_cloudfront_distribution: Remove validation from custom_origin params ([#987](https://github.com/terraform-providers/terraform-provider-aws/issues/987))
* resource_aws_route53_record: Allow import of Route 53 records with underscores in the name ([#14717](https://github.com/hashicorp/terraform/pull/14717))
* d/aws_db_snapshot: Id was being set incorrectly ([#992](https://github.com/terraform-providers/terraform-provider-aws/issues/992))
* resource/aws_spot_fleet_request: Raise the create timeout to be 10m ([#993](https://github.com/terraform-providers/terraform-provider-aws/issues/993))
* d/aws_ecs_cluster: Add ARN as an exported param for aws_ecs_cluster ([#991](https://github.com/terraform-providers/terraform-provider-aws/issues/991))
* resource/aws_ebs_volume: Not setting the state for ebs_volume correctly ([#999](https://github.com/terraform-providers/terraform-provider-aws/issues/999))
* resource/aws_network_acl: Make action in ingress / egress case insensitive ([#1000](https://github.com/terraform-providers/terraform-provider-aws/issues/1000))

## 0.1.1 (June 21, 2017)

BUG FIXES:

* Fixing malformed ARN attribute for aws_security_group data source ([#910](https://github.com/terraform-providers/terraform-provider-aws/issues/910))

## 0.1.0 (June 20, 2017)

BACKWARDS INCOMPATIBILITIES / NOTES:

FEATURES:

* **New Resource:** `aws_vpn_gateway_route_propagation` [[#15137](https://github.com/terraform-providers/terraform-provider-aws/issues/15137)](https://github.com/hashicorp/terraform/pull/15137)

IMPROVEMENTS:

* resource/ebs_snapshot: Add support for tags ([#3](https://github.com/terraform-providers/terraform-provider-aws/issues/3))
* resource/aws_elasticsearch_domain: now retries on IAM role association failure ([#12](https://github.com/terraform-providers/terraform-provider-aws/issues/12))
* resource/codebuild_project: Increase timeout for creation retry (IAM) ([#904](https://github.com/terraform-providers/terraform-provider-aws/issues/904))
* resource/dynamodb_table: Expose stream_label attribute ([#20](https://github.com/terraform-providers/terraform-provider-aws/issues/20))
* resource/opsworks: Add support for configurable timeouts in AWS OpsWorks Instances. ([#857](https://github.com/terraform-providers/terraform-provider-aws/issues/857))
* Fix handling of AdRoll's hologram clients ([#17](https://github.com/terraform-providers/terraform-provider-aws/issues/17))
* resource/sqs_queue: Add support for name_prefix to aws_sqs_queue ([#855](https://github.com/terraform-providers/terraform-provider-aws/issues/855))
* resource/iam_role: Add support for iam_role tp force_detach_policies ([#890](https://github.com/terraform-providers/terraform-provider-aws/issues/890))

BUG FIXES:

* fix aws cidr validation error [[#15158](https://github.com/terraform-providers/terraform-provider-aws/issues/15158)](https://github.com/hashicorp/terraform/pull/15158)
* resource/elasticache_parameter_group: Retry deletion on InvalidCacheParameterGroupState ([#8](https://github.com/terraform-providers/terraform-provider-aws/issues/8))
* resource/security_group: Raise creation timeout ([#9](https://github.com/terraform-providers/terraform-provider-aws/issues/9))
* resource/rds_cluster: Retry modification on InvalidDBClusterStateFault ([#18](https://github.com/terraform-providers/terraform-provider-aws/issues/18))
* resource/lambda: Fix incorrect GovCloud regexes ([#16](https://github.com/terraform-providers/terraform-provider-aws/issues/16))
* Allow ipv6_cidr_block to be assigned to peering_connection ([#879](https://github.com/terraform-providers/terraform-provider-aws/issues/879))
* resource/rds_db_instance: Correctly create cross-region encrypted replica ([#865](https://github.com/terraform-providers/terraform-provider-aws/issues/865))
* resource/eip: dissociate EIP on update ([#878](https://github.com/terraform-providers/terraform-provider-aws/issues/878))
* resource/iam_server_certificate: Increase deletion timeout ([#907](https://github.com/terraform-providers/terraform-provider-aws/issues/907))
