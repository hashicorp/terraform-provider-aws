---
subcategory: "DocumentDB"
layout: "aws"
page_title: "AWS: aws_docdb_engine_version"
description: |-
  Information about a DocumentDB engine version.
---

# Data Source: aws_docdb_engine_version

Information about a DocumentDB engine version.

## Example Usage

```terraform
data "aws_docdb_engine_version" "test" {
  version = "3.6.0"
}
```

## Argument Reference

This data source supports the following arguments:

* `engine` - (Optional) DB engine. (Default: `docdb`)
* `parameter_group_family` - (Optional) Name of a specific DB parameter group family. An example parameter group family is `docdb3.6`.
* `preferred_versions` - (Optional) Ordered list of preferred engine versions. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned. If both the `version` and `preferred_versions` arguments are not configured, the data source will return the default version for the engine.
* `version` - (Optional) Version of the DB engine. For example, `3.6.0`. If `version` and `preferred_versions` are not set, the data source will provide information for the AWS-defined default version. If both the `version` and `preferred_versions` arguments are not configured, the data source will return the default version for the engine.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `engine_description` - Description of the database engine.
* `exportable_log_types` - Set of log types that the database engine has available for export to CloudWatch Logs.
* `supports_log_exports_to_cloudwatch` - Indicates whether the engine version supports exporting the log types specified by `exportable_log_types` to CloudWatch Logs.
* `valid_upgrade_targets` - A set of engine versions that this database engine version can be upgraded to.
* `version_description` - Description of the database engine version.
