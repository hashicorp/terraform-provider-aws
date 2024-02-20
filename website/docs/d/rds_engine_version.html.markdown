---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_engine_version"
description: |-
  Information about an RDS engine version.
---

# Data Source: aws_rds_engine_version

Information about an RDS engine version.

## Example Usage

### Basic Usage

```terraform
data "aws_rds_engine_version" "test" {
  engine             = "mysql"
  preferred_versions = ["8.0.27", "8.0.26"]
}
```

### With `filter`

```terraform
data "aws_rds_engine_version" "test" {
  engine      = "aurora-postgresql"
  version     = "10.14"
  include_all = true

  filter {
    name   = "engine-mode"
    values = ["serverless"]
  }
}
```

## Argument Reference

The following arguments are required:

* `engine` - (Required) Database engine. Engine values include `aurora`, `aurora-mysql`, `aurora-postgresql`, `docdb`, `mariadb`, `mysql`, `neptune`, `oracle-ee`, `oracle-se`, `oracle-se1`, `oracle-se2`, `postgres`, `sqlserver-ee`, `sqlserver-ex`, `sqlserver-se`, and `sqlserver-web`.

The following arguments are optional:

* `default_only` - (Optional) When set to `true`, the default version for the specified `engine` or combination of `engine` and major `version` will be returned. Can be used to limit responses to a single version when they would otherwise fail for returning multiple versions.
* `filter` - (Optional) One or more name/value pairs to filter off of. There are several valid keys; for a full reference, check out [describe-db-engine-versions in the AWS CLI reference](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/rds/describe-db-engine-versions.html).
* `include_all` - (Optional) When set to `true`, the specified `version` or member of `preferred_versions` will be returned even if it is `deprecated`. Otherwise, only `available` versions will be returned.
* `latest` - (Optional) When set to `true`, the data source attempts to return the most recent version matching the other criteria you provide. This differs from `default_only`. For example, the latest version is not always the default. In addition, AWS may return multiple defaults depending on the criteria. Using `latest` will avoid `multiple RDS engine versions` errors. **Note:** The data source uses a best-effort approach at selecting the latest version but due to the complexity of version identifiers across engines and incomplete version date information provided by AWS, using `latest` may _not_ return the latest version in every situation.
* `parameter_group_family` - (Optional) Name of a specific database parameter group family. Examples of parameter group families are `mysql8.0`, `mariadb10.4`, and `postgres12`.
* `preferred_major_targets` - (Optional) Ordered list of preferred major version upgrade targets. The version corresponding to the first match in this list will be returned unless the `latest` parameter is set to `true`. If you don't configure `version`, `preferred_major_targets`, `preferred_upgrade_targets`, and `preferred_versions`, the data source will return the default version for the engine. You can use this with other version criteria.
* `preferred_upgrade_targets` - (Optional) Ordered list of preferred version upgrade targets. The version corresponding to the first match in this list will be returned unless the `latest` parameter is set to `true`. If you don't configure `version`, `preferred_major_targets`, `preferred_upgrade_targets`, and `preferred_versions`, the data source will return the default version for the engine. You can use this with other version criteria.
* `preferred_versions` - (Optional) Ordered list of preferred versions. The first match in this list that matches any other criteria will be returned unless the `latest` parameter is set to `true`. If you don't configure `version`, `preferred_major_targets`, `preferred_upgrade_targets`, and `preferred_versions`, the data source will return the default version for the engine. You can use this with other version criteria.
* `version` - (Optional) Version of the database engine. For example, `5.7.22`, `10.1.34`, or `12.3`. `version` can be a major version which may result in the data source finding multiple versions and returning an error unless the `latest` parameter is set to `true`. If you don't configure `version`, `preferred_major_targets`, `preferred_upgrade_targets`, and `preferred_versions`, the data source will return the default version for the engine. You can use this with other version criteria. **NOTE:** In a future Terraform AWS provider version, `version` will only contain the version information you configure and not the complete version information that the data source gets from AWS. Instead, that version information will be available in the `version_actual` attribute.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `default_character_set` - The default character set for new instances of this engine version.
* `engine_description` - Description of the database engine.
* `exportable_log_types` - Set of log types that the database engine has available for export to CloudWatch Logs.
* `status` - Status of the database engine version, either available or deprecated.
* `supported_character_sets` - Set of the character sets supported by this engine.
* `supported_feature_names` - Set of features supported by the database engine.
* `supported_modes` - Set of the supported database engine modes.
* `supported_timezones` - Set of the time zones supported by this engine.
* `supports_global_databases` - Indicates whether you can use Aurora global databases with a specific database engine version.
* `supports_log_exports_to_cloudwatch` - Indicates whether the engine version supports exporting the log types specified by `exportable_log_types` to CloudWatch Logs.
* `supports_parallel_query` - Indicates whether you can use Aurora parallel query with a specific database engine version.
* `supports_read_replica` - Indicates whether the database engine version supports read replicas.
* `valid_upgrade_targets` - Set of engine versions that this database engine version can be upgraded to.
* `version_actual` - Version of the database engine.
* `version_description` - Description of the database engine version.
