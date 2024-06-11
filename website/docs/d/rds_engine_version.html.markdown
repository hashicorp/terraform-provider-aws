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

* `default_only` - (Optional) Whether the engine version must be an AWS-defined default version. Some engines have multiple default versions, such as for each major version. Using `default_only` may help avoid `multiple RDS engine versions` errors. See also `latest`.
* `filter` - (Optional) One or more name/value pairs to use in filtering versions. There are several valid keys; for a full reference, check out [describe-db-engine-versions in the AWS CLI reference](https://awscli.amazonaws.com/v2/documentation/api/latest/reference/rds/describe-db-engine-versions.html).
* `has_major_target` - (Optional) Whether the engine version must have one or more major upgrade targets. Not including `has_major_target` or setting it to `false` doesn't imply that there's no corresponding major upgrade target for the engine version.
* `has_minor_target` - (Optional) Whether the engine version must have one or more minor upgrade targets. Not including `has_minor_target` or setting it to `false` doesn't imply that there's no corresponding minor upgrade target for the engine version.
* `include_all` - (Optional) Whether the engine version `status` can either be `deprecated` or `available`. When not set or set to `false`, the engine version `status` will always be `available`.
* `latest` - (Optional) Whether the engine version is the most recent version matching the other criteria. This is different from `default_only` in important ways: "default" relies on AWS-defined defaults, the latest version isn't always the default, and AWS might have multiple default versions for an engine. As a result, `default_only` might not prevent errors from `multiple RDS engine versions`, while `latest` will. (`latest` can be used with `default_only`.) **Note:** The data source uses a best-effort approach at selecting the latest version. Due to the complexity of version identifiers across engines and incomplete version date information provided by AWS, using `latest` may not always result in the engine version being the actual latest version.
* `parameter_group_family` - (Optional) Name of a specific database parameter group family. Examples of parameter group families are `mysql8.0`, `mariadb10.4`, and `postgres12`.
* `preferred_major_targets` - (Optional) Ordered list of preferred major version upgrade targets. The engine version will be the first match in the list unless the `latest` parameter is set to `true`. The engine version will be the default version if you don't include any criteria, such as `preferred_major_targets`.
* `preferred_upgrade_targets` - (Optional) Ordered list of preferred version upgrade targets. The engine version will be the first match in this list unless the `latest` parameter is set to `true`. The engine version will be the default version if you don't include any criteria, such as `preferred_upgrade_targets`.
* `preferred_versions` - (Optional) Ordered list of preferred versions. The engine version will be the first match in this list unless the `latest` parameter is set to `true`. The engine version will be the default version if you don't include any criteria, such as `preferred_versions`.
* `version` - (Optional) Engine version. For example, `5.7.22`, `10.1.34`, or `12.3`. `version` can be a partial version identifier which can result in `multiple RDS engine versions` errors unless the `latest` parameter is set to `true`. The engine version will be the default version if you don't include any criteria, such as `version`. **NOTE:** In a future Terraform AWS provider version, `version` will only contain the version information you configure and not the complete version information that the data source gets from AWS. Instead, that version information will be available in the `version_actual` attribute.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `default_character_set` - Default character set for new instances of the engine version.
* `engine_description` - Description of the engine.
* `exportable_log_types` - Set of log types that the engine version has available for export to CloudWatch Logs.
* `status` - Status of the engine version, either `available` or `deprecated`.
* `supported_character_sets` - Set of character sets supported by th engine version.
* `supported_feature_names` - Set of features supported by the engine version.
* `supported_modes` - Set of supported engine version modes.
* `supported_timezones` - Set of the time zones supported by the engine version.
* `supports_global_databases` - Whether you can use Aurora global databases with the engine version.
* `supports_log_exports_to_cloudwatch` - Whether the engine version supports exporting the log types specified by `exportable_log_types` to CloudWatch Logs.
* `supports_limitless_database` - Whether the engine version supports Aurora Limitless Database.
* `supports_parallel_query` - Whether you can use Aurora parallel query with the engine version.
* `supports_read_replica` - Whether the engine version supports read replicas.
* `valid_major_targets` - Set of versions that are valid major version upgrades for the engine version.
* `valid_minor_targets` - Set of versions that are valid minor version upgrades for the engine version.
* `valid_upgrade_targets` - Set of versions that are valid major or minor upgrades for the engine version.
* `version_actual` - Complete engine version.
* `version_description` - Description of the engine version.
