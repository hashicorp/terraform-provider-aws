---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_engine_version"
description: |-
  Information about a Neptune engine version.
---

# Data Source: aws_neptune_engine_version

Information about a Neptune engine version.

~> **Note:** If AWS returns multiple matching engine versions, this data source will produce a `multiple Neptune engine versions` error. To avoid this, provide additional criteria to narrow the results or use the `latest` argument to select a single version. See the [Argument Reference](#argument-reference) for details.

## Example Usage

```terraform
data "aws_neptune_engine_version" "test" {
  preferred_versions = ["1.4.5.0", "1.4.4.0", "1.4.3.0"]
}
```

## Argument Reference

This data source supports the following arguments:

* `default_only` – (Optional) Whether to return only default engine versions that match all other criteria. AWS may define multiple default versions for a given engine, so using `default_only` alone does not guarantee that only one version will be returned. To ensure a single version is selected, consider combining this with `latest`. Note that default versions are defined by AWS and may not reflect the most recent engine version available.
* `engine` - (Optional) DB engine. Must be `neptune`. Default is `neptune`.
* `has_major_target` - (Optional) Whether to filter for engine versions that have a major target.
* `has_minor_target` - (Optional) Whether to filter for engine versions that have a minor target.
* `latest` – (Optional) Whether to return only the latest engine version that matches all other criteria. This differs from `default_only`: AWS may define multiple defaults, and the latest version is not always marked as the default. As a result, `default_only` may still return multiple versions, while `latest` selects a single version. The two options can be used together. **Note:** This argument uses a best-effort approach. Because AWS does not consistently provide version dates or standardized identifiers, the result may not always reflect the true latest version.
* `parameter_group_family` - (Optional) Name of a specific DB parameter group family. An example parameter group family is `neptune1.4`. For some versions, if this is provided, AWS returns no results.
* `preferred_major_targets` - (Optional) Ordered list of preferred major engine versions.
* `preferred_upgrade_targets` - (Optional) Ordered list of preferred upgrade engine versions.
* `preferred_versions` - (Optional) Ordered list of preferred engine versions. The first match in this list will be returned. If no preferred matches are found and the original search returned more than one result, an error is returned. If both the `version` and `preferred_versions` arguments are not configured, the data source will return the default version for the engine.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `version` - (Optional) Version of the DB engine. For example, `1.0.1.0`, `1.0.2.2`, and `1.0.3.0`. If both the `version` and `preferred_versions` arguments are not configured, the data source will return the default version for the engine.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `default_character_set` - Default character set for the engine version.
* `engine_description` - Description of the database engine.
* `exportable_log_types` - Set of log types that the database engine has available for export to CloudWatch Logs.
* `supported_character_sets` - Set of character sets supported by this engine version.
* `supported_timezones` - Set of time zones supported by this engine.
* `supports_global_databases` - Whether the engine version supports global databases.
* `supports_log_exports_to_cloudwatch` - Whether the engine version supports exporting the log types specified by `exportable_log_types` to CloudWatch Logs.
* `supports_read_replica` - Whether the database engine version supports read replicas.
* `valid_major_targets` - Set of valid major engine versions that this version can be upgraded to.
* `valid_minor_targets` - Set of valid minor engine versions that this version can be upgraded to.
* `valid_upgrade_targets` - Set of engine versions that this database engine version can be upgraded to.
* `version_actual` - Actual engine version returned by the API.
* `version_description` - Description of the database engine version.
