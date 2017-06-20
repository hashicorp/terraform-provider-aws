## 0.1.1 (Unreleased)
## 0.1.0 (June 20, 2017)

BACKWARDS INCOMPATIBILITIES / NOTES:

FEATURES:

* **New Resource:** `aws_vpn_gateway_route_propagation` [[#15137](https://github.com/terraform-providers/terraform-provider-aws/15137)](https://github.com/hashicorp/terraform/pull/15137)

IMPROVEMENTS:

* resource/ebs_snapshot: Add support for tags ([#3](https://github.com/terraform-providers/terraform-provider-aws/3))
* resource/aws_elasticsearch_domain: now retries on IAM role association failure ([#12](https://github.com/terraform-providers/terraform-provider-aws/12))
* resource/codebuild_project: Increase timeout for creation retry (IAM) ([#904](https://github.com/terraform-providers/terraform-provider-aws/904))
* resource/dynamodb_table: Expose stream_label attribute ([#20](https://github.com/terraform-providers/terraform-provider-aws/20))
* resource/opsworks: Add support for configurable timeouts in AWS OpsWorks Instances. ([#857](https://github.com/terraform-providers/terraform-provider-aws/857))
* Fix handling of AdRoll's hologram clients ([#17](https://github.com/terraform-providers/terraform-provider-aws/17))
* resource/sqs_queue: Add support for name_prefix to aws_sqs_queue ([#855](https://github.com/terraform-providers/terraform-provider-aws/855))
* resource/iam_role: Add support for iam_role tp force_detach_policies ([#890](https://github.com/terraform-providers/terraform-provider-aws/890))

BUG FIXES:

* fix aws cidr validation error [[#15158](https://github.com/terraform-providers/terraform-provider-aws/15158)](https://github.com/hashicorp/terraform/pull/15158)
* resource/elasticache_parameter_group: Retry deletion on InvalidCacheParameterGroupState ([#8](https://github.com/terraform-providers/terraform-provider-aws/8))
* resource/security_group: Raise creation timeout ([#9](https://github.com/terraform-providers/terraform-provider-aws/9))
* resource/rds_cluster: Retry modification on InvalidDBClusterStateFault ([#18](https://github.com/terraform-providers/terraform-provider-aws/18))
* resource/lambda: Fix incorrect GovCloud regexes ([#16](https://github.com/terraform-providers/terraform-provider-aws/16))
* Allow ipv6_cidr_block to be assigned to peering_connection ([#879](https://github.com/terraform-providers/terraform-provider-aws/879))
* resource/rds_db_instance: Correctly create cross-region encrypted replica ([#865](https://github.com/terraform-providers/terraform-provider-aws/865))
* resource/eip: dissociate EIP on update ([#878](https://github.com/terraform-providers/terraform-provider-aws/878))
* resource/iam_server_certificate: Increase deletion timeout ([#907](https://github.com/terraform-providers/terraform-provider-aws/907))
