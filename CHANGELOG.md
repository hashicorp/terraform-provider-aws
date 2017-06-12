## 0.1.0 (Unreleased)

BACKWARDS INCOMPATIBILITIES / NOTES:

FEATURES:

* **New Resource:** `aws_vpn_gateway_route_propagation` [GH-15137](https://github.com/hashicorp/terraform/pull/15137)

IMPROVEMENTS:

* resource/ebs_snapshot: Add support for tags [GH-3]
* resource/aws_elasticsearch_domain: now retries on IAM role association failure [GH-12]

BUG FIXES:

* fix aws cidr validation error [GH-15158](https://github.com/hashicorp/terraform/pull/15158)
* resource/elasticache_parameter_group: Retry deletion on InvalidCacheParameterGroupState [GH-8]
* resource/security_group: Raise creation timeout [GH-9]
