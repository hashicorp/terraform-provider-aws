## 0.1.2 (Unreleased)

FEATURES:

* **New Data Source**: `aws_ecr_repository` [GH-944]

IMPROVEMENTS:

* resource/api_gateway_integration: Add support for specifying cache key parameters [GH-893]
* resource/cloudwatch_event_target: Add ecs_target [GH-977]
* resource/vpn_connection: Add BGP related information on aws_vpn_connection [GH-973]

BUG FIXES:

* resource/waf: Only set FieldToMatch.Data if not empty [GH-953]
* resource/aws_elastic_beanstalk_application_version: Scope labels to application [GH-956]
* resource/aws_s3_bucket: Allow use of `days = 0` with lifecycle transition [GH-957]
* resource/aws_ssm_maintenance_window_task: Make task_parameters updateable on aws_ssm_maintenance_window_task resource [GH-965]
* resource/aws_kinesis_stream: don't force stream destroy on shard_count update [GH-894]
* resource/aws_cloudfront_distribution: Remove validation from custom_origin params [GH-987]
* resource_aws_route53_record: Allow import of Route 53 records with underscores in the name ([#14717](https://github.com/hashicorp/terraform/pull/14717))
* d/aws_db_snapshot: Id was being set incorrectly [GH-992]
* resource/aws_spot_fleet_request: Raise the create timeout to be 10m [GH-993]
* d/aws_ecs_cluster: Add ARN as an exported param for aws_ecs_cluster [GH-991]

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
