---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_orderable_db_instance"
description: |-
  Information about RDS orderable DB instances.
---

# Data Source: aws_rds_orderable_db_instance

Information about RDS orderable DB instances and valid parameter combinations.

## Example Usage

```terraform
data "aws_rds_orderable_db_instance" "test" {
  engine         = "mysql"
  engine_version = "5.7.22"
  license_model  = "general-public-license"
  storage_type   = "standard"

  preferred_instance_classes = ["db.r6.xlarge", "db.m4.large", "db.t3.small"]
}
```

Valid parameter combinations can also be found with `preferred_engine_versions` and/or `preferred_instance_classes`.

```terraform
data "aws_rds_orderable_db_instance" "test" {
  engine        = "mysql"
  license_model = "general-public-license"

  preferred_engine_versions  = ["5.6.35", "5.6.41", "5.6.44"]
  preferred_instance_classes = ["db.t2.small", "db.t3.medium", "db.t3.large"]
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `availability_zone_group` - (Optional) Availability zone group.
* `engine_latest_version` - (Optional) When set to `true`, the data source attempts to return the most recent version matching the other criteria you provide. You must use `engine_latest_version` with `preferred_instance_classes` and/or `preferred_engine_versions`. Using `engine_latest_version` will avoid `multiple RDS DB Instance Classes` errors. If you use `engine_latest_version` with `preferred_instance_classes`, the data source returns the latest version for the _first_ matching instance class (instance class priority). **Note:** The data source uses a best-effort approach at selecting the latest version but due to the complexity of version identifiers across engines, using `engine_latest_version` may _not_ return the latest version in every situation.
* `engine_version` - (Optional) Version of the DB engine. If none is provided, the data source tries to use the AWS-defined default version that matches any other criteria.
* `engine` - (Required) DB engine. Engine values include `aurora`, `aurora-mysql`, `aurora-postgresql`, `docdb`, `mariadb`, `mysql`, `neptune`, `oracle-ee`, `oracle-se`, `oracle-se1`, `oracle-se2`, `postgres`, `sqlserver-ee`, `sqlserver-ex`, `sqlserver-se`, and `sqlserver-web`.
* `instance_class` - (Optional) DB instance class. Examples of classes are `db.m3.2xlarge`, `db.t2.small`, and `db.m3.medium`.
* `license_model` - (Optional) License model. Examples of license models are `general-public-license`, `bring-your-own-license`, and `amazon-license`.
* `preferred_engine_versions` - (Optional) Ordered list of preferred RDS DB instance engine versions. When `engine_latest_version` is not set, the data source will return the first match in this list that matches any other criteria. If the data source finds no preferred matches or multiple matches without `engine_latest_version`, it returns an error. **CAUTION:** We don't recommend using `preferred_engine_versions` without `preferred_instance_classes` since the data source returns an arbitrary `instance_class` based on the first one AWS returns that matches the engine version and any other criteria.
* `preferred_instance_classes` - (Optional) Ordered list of preferred RDS DB instance classes. The data source will return the first match in this list that matches any other criteria. If the data source finds no preferred matches or multiple matches without `engine_latest_version`, it returns an error. If you use `preferred_instance_classes` without `preferred_engine_versions` or `engine_latest_version`, the data source returns an arbitrary `engine_version` based on the first one AWS returns matching the instance class and any other criteria.
* `read_replica_capable` - (Optional) Whether a DB instance can have a read replica.
* `storage_type` - (Optional) Storage types. Examples of storage types are `standard`, `io1`, `gp2`, and `aurora`.
* `supported_engine_modes` - (Optional) Use to limit results to engine modes such as `provisioned`.
* `supported_network_types` - (Optional) Use to limit results to network types `IPV4` or `DUAL`.
* `supports_clusters` - (Optional) Whether to limit results to instances that support clusters.
* `supports_multi_az` - (Optional) Whether to limit results to instances that are multi-AZ capable.
* `supports_enhanced_monitoring` - (Optional) Enable this to ensure a DB instance supports Enhanced Monitoring at intervals from 1 to 60 seconds.
* `supports_global_databases` - (Optional) Enable this to ensure a DB instance supports Aurora global databases with a specific combination of other DB engine attributes.
* `supports_iam_database_authentication` - (Optional) Enable this to ensure a DB instance supports IAM database authentication.
* `supports_iops` - (Optional) Enable this to ensure a DB instance supports provisioned IOPS.
* `supports_kerberos_authentication` - (Optional) Enable this to ensure a DB instance supports Kerberos Authentication.
* `supports_performance_insights` - (Optional) Enable this to ensure a DB instance supports Performance Insights.
* `supports_storage_autoscaling` - (Optional) Enable this to ensure Amazon RDS can automatically scale storage for DB instances that use the specified DB instance class.
* `supports_storage_encryption` - (Optional) Enable this to ensure a DB instance supports encrypted storage.
* `vpc` - (Optional) Boolean that indicates whether to show only VPC or non-VPC offerings.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `availability_zones` - Availability zones where the instance is available.
* `max_iops_per_db_instance` - Maximum total provisioned IOPS for a DB instance.
* `max_iops_per_gib` - Maximum provisioned IOPS per GiB for a DB instance.
* `max_storage_size` - Maximum storage size for a DB instance.
* `min_iops_per_db_instance` - Minimum total provisioned IOPS for a DB instance.
* `min_iops_per_gib` - Minimum provisioned IOPS per GiB for a DB instance.
* `min_storage_size` - Minimum storage size for a DB instance.
* `multi_az_capable` - Whether a DB instance is Multi-AZ capable.
* `outpost_capable` - Whether a DB instance supports RDS on Outposts.
