---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_engine_version"
description: |-
  Information about a Neptune engine version.
---

# Data Source: aws_neptune_engine_version

Information about a Neptune engine version.

## Example Usage

```terraform
data "aws_neptune_engine_version" "test" {
  preferred_versions = ["1.0.3.0", "1.0.2.2", "1.0.2.1"]
}
```

## Argument Reference

This data source supports the following arguments:

* `engine` - (Optional) DB engine. (Default: `neptune`)
* `parameter_group_family` - (Optional) Name of a specific DB parameter group family. An example parameter group family is `neptune1`.
* `preferred_versions` - (Optional) Ordered list of preferred engine versions. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned. If both the `version` and `preferred_versions` arguments are not configured, the data source will return the default version for the engine.
* `version` - (Optional) Version of the DB engine. For example, `1.0.1.0`, `1.0.2.2`, and `1.0.3.0`. If both the `version` and `preferred_versions` arguments are not configured, the data source will return the default version for the engine.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `engine_description` - Description of the database engine.
* `exportable_log_types` - Set of log types that the database engine has available for export to CloudWatch Logs.
* `supported_timezones` - Set of the time zones supported by this engine.
* `supports_log_exports_to_cloudwatch` - Indicates whether the engine version supports exporting the log types specified by `exportable_log_types` to CloudWatch Logs.
* `supports_read_replica` - Indicates whether the database engine version supports read replicas.
* `valid_upgrade_targets` - Set of engine versions that this database engine version can be upgraded to.
* `version_description` - Description of the database engine version.
