---
layout: "aws"
page_title: "AWS: aws_db_engine_version"
sidebar_current: "docs-aws-datasource-db-engine-version"
description: |-
  Get information on a DB Engine Version.
---

# Data Source: aws_db_engine_version

Use this data source to get information about a DB Engine Version.

## Example Usage

```hcl
data "aws_db_engine_version" "example" {
  engine      = "postgres"
  most_recent = true
}
```

## Argument Reference

The following arguments are supported:

* `engine` - (Required) The database engine to return.
* `engine_version` - (Optional) The database engine version to return.
* `most_recent` - (Optional) If more than one result is returned, use the most recent Snapshot.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `db_engine_description` - The description of the database engine.
* `db_engine_version_description` - The description of the database engine version.
* `db_parameter_group_family` - The name of the DB parameter group family for the database engine.
* `exportable_log_types` - The types of logs that the database engine has available for export to CloudWatch Logs.
* `supported_engine_modes` - A list of the supported DB engine modes.
* `supported_feature_names` - A list of features supported by the DB engine. Supported feature names include the following.
* `supports_log_exports_to_cloudwatch_logs` - A value that indicates whether the engine version supports exporting the log types specified by ExportableLogTypes to CloudWatch Logs.
* `supports_read_replica` - Indicates whether the database engine version supports Read Replicas.
