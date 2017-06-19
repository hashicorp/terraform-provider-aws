## 0.1.0 (Unreleased)

BACKWARDS INCOMPATIBILITIES / NOTES:

FEATURES:

* **New Resource:** `aws_vpn_gateway_route_propagation` [GH-15137](https://github.com/hashicorp/terraform/pull/15137)

IMPROVEMENTS:

* resource/ebs_snapshot: Add support for tags [GH-3]
* resource/aws_elasticsearch_domain: now retries on IAM role association failure [GH-12]
* resource/codebuild_project: Increase timeout for creation retry (IAM) [GH-904]
* resource/dynamodb_table: Expose stream_label attribute [GH-20]
* resource/opsworks: Add support for configurable timeouts in AWS OpsWorks Instances. [GH-857]
* Fix handling of AdRoll's hologram clients [GH-17]
* resource/sqs_queue: Add support for name_prefix to aws_sqs_queue [GH-855]
* resource/iam_role: Add support for iam_role tp force_detach_policies [GH-890]

BUG FIXES:

* fix aws cidr validation error [GH-15158](https://github.com/hashicorp/terraform/pull/15158)
* resource/elasticache_parameter_group: Retry deletion on InvalidCacheParameterGroupState [GH-8]
* resource/security_group: Raise creation timeout [GH-9]
* resource/rds_cluster: Retry modification on InvalidDBClusterStateFault [GH-18]
* resource/lambda: Fix incorrect GovCloud regexes [GH-16]
* Allow ipv6_cidr_block to be assigned to peering_connection [GH-879]
* resource/rds_db_instance: Correctly create cross-region encrypted replica [GH-865]
* resource/eip: dissociate EIP on update [GH-878]
* resource/iam_server_certificate: Increase deletion timeout [GH-907]
