---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_transformer"
description: |-
  Terraform resource for managing an AWS CloudWatch Logs Transformer.
---

# Resource: aws_cloudwatch_log_transformer

Terraform resource for managing an AWS CloudWatch Logs Transformer.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_transformer" "example" {
  log_group_arn = aws_cloudwatch_log_group.example.arn
  transformer_config {
    parse_json {}
  }
}

resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `log_group_arn` - (Required) Log group ARN to set the transformer for.
* `transformer_config` - (Required) Specifies the configuration of the transformer. You must include at least one configuration, and 20 at most. See [`transformer_config`](#transformer_config-block) below for details.

### `transformer_config` Block

~> **Note** You must only specify a single processor per `transformer_config` block. Besides, the first processor in a transformer must be a parser.

Each `transformer_config` supports the following arguments:

* `add_keys` - (Optional) Adds new key-value pairs to the log event. See [`add_keys`](#add_keys-block) below for details.
* `copy_value` - (Optional) Copies values within a log event. See [`copy_value`](#copy_value-block) below for details.
* `csv` - (Optional) Parses comma-separated values (CSV) from the log events into columns. See [`csv`](#csv-block) below for details.
* `date_time_converter` - (Optional) Converts a datetime string into a format that you specify. See [`date_time_converter`](#date_time_converter-block) below for details.
* `delete_keys` - (Optional) Deletes entry from a log event. See [`delete_keys`](#delete_keys-block) below for details.
* `grok` - (Optional) Parses and structures unstructured data by using pattern matching. See [`grok`](#grok-block) below for details.
* `list_to_map` - (Optional) Converts list of objects that contain key fields into a map of target keys. See [`list_to_map`](#list_to_map-block) below for details.
* `lower_case_string` - (Optional) Converts a string to lowercase. See [`lower_case_string`](#lower_case_string-block) below for details.
* `move_keys` - (Optional) Moves a key from one field to another. See [`move_keys`](#move_keys-block) below for details.
* `parse_cloudfront` - (Optional) Parses CloudFront vended logs, extracts fields, and converts them into JSON format. See [`parse_cloudfront`](#parse_cloudfront-block) below for details.
* `parse_json` - (Optional) Parses log events that are in JSON format. See [`parse_json`](#parse_json-block) below for details.
* `parse_key_value` - (Optional) Parses a specified field in the original log event into key-value pairs. See [`parse_key_value`](#parse_key_value-block) below for details.
* `parse_postgres` - (Optional) Parses RDS for PostgreSQL vended logs, extracts fields, and and convert them into a JSON format. See [`parse_postgres`](#parse_postgres-block) below for details.
* `parse_route53` - (Optional) Parses Route 53 vended logs, extracts fields, and converts them into JSON format. See [`parse_route53`](#parse_route53-block) below for details.
* `parse_to_ocsf` - (Optional) Parses logs events and converts them into Open Cybersecurity Schema Framework (OCSF) events. See [`parse_to_ocsf`](#parse_to_ocsf-block) below for details.
* `parse_vpc` - (Optional) Parses Amazon VPC vended logs, extracts fields, and converts them into JSON format. See [`parse_vpc`](#parse_vpc-block) below for details.
* `parse_waf` - (Optional) Parses AWS WAF vended logs, extracts fields, and converts them into JSON format. See [`parse_waf`](#parse_waf-block) below for details.
* `rename_keys` - (Optional) Renames keys in a log event. See [`rename_keys`](#rename_keys-block) below for details.
* `split_string` - (Optional) Splits a field into an array of strings using a delimiting character. See [`split_string`](#split_string-block) below for details.
* `substitute_string` - (Optional) Matches a key’s value against a regular expression and replaces all matches with a replacement string. See [`substitute_string`](#substitute_string-block) below for details.
* `trim_string` - (Optional) Removes leading and trailing whitespace from a string. See [`trim_string`](#trim_string-block) below for details.
* `type_converter` - (Optional) Converts a value type associated with the specified key to the specified type. See [`type_converter`](#type_converter-block) below for details.
* `upper_case_string` - (Optional) Converts a string to uppercase. See [`upper_case_string`](#upper_case_string-block) below for details.

### `add_keys` Block

~> **Note** You can only add a single `add_keys` processor per transformer.

The `add_keys` block supports the following arguments:

* `entry` - (Required) Objects containing the information about the keys to add to the log event. You must include at least one entry, and five at most. See [`add_keys` `entry`](#add_keys-entry-block) below for details.

### `add_keys` `entry` Block

Each `entry` block supports the following arguments:

* `key` - (Required) Specifies the key of the new entry to be added to the log event.
* `overwrite_if_exists` - (Optional) Specifies whether to overwrite the value if the key already exists in the log event. Defaults to `false`.
* `value` - (Required) Specifies the value of the new entry to be added to the log event.

### `copy_value` Block

~> **Note** You can only add a single `copy_value` processor per transformer.

The `copy_value` block supports the following arguments:

* `entry` - (Required) Objects containing the information about the values to copy to the log event. You must include at least one entry, and five at most. See [`copy_value` `entry`](#copy_value-entry-block) below for details.

### `copy_value` `entry` Block

Each `entry` block supports the following arguments:

* `overwrite_if_exists` - (Optional) Specifies whether to overwrite the value if the destination key already exists. Defaults to `false`.
* `source` - (Required) Specifies the key to copy.
* `target` - (Required) Specifies the key of the field to copy the value to.

### `csv` Block

Each `csv` block supports the following arguments:

* `columns` - (Optional) Specifies the names to use for the columns in the transformed log event. If not specified, default column names (`[column_1, column_2 ...]`) are used.
* `delimiter` - (Optional) Specifies the character used to separate each column in the original comma-separated value log event. Defaults to the comma `,` character.
* `quote_character` - (Optional) Specifies the character used as a text qualifier for a single column of data. Defaults to the double quotation mark `"` character.
* `source` - (Optional) Specifies the path to the field in the log event that has the comma separated values to be parsed. If omitted, the whole log message is processed.

### `date_time_converter` Block

Each `date_time_converter` block supports the following arguments:

* `locale` - (Optional) Specifies the locale of the source field. Defaults to `locale.ROOT`.
* `match_patterns` - (Required) Specifies the list of patterns to match against the `source` field.
* `source` - (Required) Specifies the key to apply the date conversion to.
* `source_timezone` - (Optional) Specifies the time zone of the source field. Defaults to `UTC`.
* `target` - (Required) Specifies the JSON field to store the result in.
* `target_format` - (Optional) Specifies the datetime format to use for the converted data in the target field. Defaults to `yyyy-MM-dd'T'HH:mm:ss.SSS'Z`.
* `target_timezone` - (Optional) Specifies the time zone of the target field. Defaults to `UTC`.

### `delete_keys` Block

Each `delete_keys` block supports the following arguments:

* `with_keys` - (Required) Specifies the keys to be deleted.

### `grok` Block

~> **Note** You can only add a single `grok` processor per transformer.

The `grok` block supports the following arguments:

* `match` - (Required) Specifies the grok pattern to match against the log event.
* `source` - (Optional) Specifies the path to the field in the log event that has the comma separated values to be parsed. If omitted, the whole log message is processed.

### `list_to_map` Block

Each `list_to_map` block supports the following arguments:

* `flatten` - (Optional) Specifies whether the list will be flattened into single items. Defaults to `false`.
* `flattened_element` - (Optional) Required if `flatten` is set to true. Specifies the element to keep. Allowed values are `first` and `last`.
* `key` - (Required) Specifies the key of the field to be extracted as keys in the generated map.
* `source` - (Required) Specifies the key in the log event that has a list of objects that will be converted to a map.
* `target` - (Optional) Specifies the key of the field that will hold the generated map.
* `value_key` - (Optional) Specifies the values that will be extracted from the source objects and put into the values of the generated map. If omitted, original objects in the source list will be put into the values of the generated map.

### `lower_case_string` Block

Each `lower_case_string` block supports the following arguments:

* `with_keys` - (Required) Specifies the keys of the fields to convert to lowercase.

### `move_keys` Block

Each `move_keys` block supports the following arguments:

* `entry` - (Required) Objects containing the information about the keys to move to the log event. You must include at least one entry, and five at most. See [`move_keys` `entry`](#move_keys-entry-block) below for details.

### `move_keys` `entry` Block

Each `entry` block supports the following arguments:

* `overwrite_if_exists` - (Optional) Specifies whether to overwrite the value if the destination key already exists. Defaults to `false`.
* `source` - (Required) Specifies the key to move.
* `target` - (Required) Specifies the key to move to.

### `parse_cloudfront` Block

~> **Note** You can only add a single `parse_cloudfront` processor per transformer. If specified, it must be the first processor in your transformer.

The `parse_cloudfront` block supports the following arguments:

* `source` - (Optional) Specifies the source field to be parsed. The only allowed value is `@message`. If omitted, the whole log message is processed.

### `parse_json` Block

Each `parse_json` block supports the following arguments:

* `destination` - (Optional) Specifies the location to put the parsed key value pair into. If omitted, it will be placed under the root node.
* `source` - (Optional) Specifies the path to the field in the log event that will be parsed. Defaults to `@message`.

### `parse_key_value` Block

Each `parse_key_value` block supports the following arguments:

* `destination` - (Optional) Specifies the destination field to put the extracted key-value pairs into.
* `field_delimiter` - (Optional) Specifies the field delimiter string that is used between key-value pairs in the original log events. Defaults to the ampersand `&` character.
* `key_prefix` - (Optional) Specifies a prefix that will be added to all transformed keys.
* `key_value_delimiter` - (Optional) Specifies the delimiter string to use between the key and value in each pair in the transformed log event. Defaults to the equal `=` character.
* `non_match_value` - (Optional) Specifies a value to insert into the value field in the result if a key-value pair is not successfully split.
* `overwrite_if_exists` - (Optional) Specifies whether to overwrite the value if the destination key already exists. Defaults to `false`.
* `source` - (Optional) Specifies the path to the field in the log event that will be parsed. Defaults to `@message`.

### `parse_postgres` Block

~> **Note** You can only add a single `parse_postgres` processor per transformer. If specified, it must be the first processor in your transformer.

The `parse_postgres` block supports the following arguments:

* `source` - (Optional) Specifies the source field to be parsed. The only allowed value is `@message`. If omitted, the whole log message is processed.

### `parse_route53` Block

~> **Note** You can only add a single `parse_route53` processor per transformer. If specified, it must be the first processor in your transformer.

The `parse_route53` block supports the following arguments:

* `source` - (Optional) Specifies the source field to be parsed. The only allowed value is `@message`. If omitted, the whole log message is processed.

### `parse_to_ocsf` Block

~> **Note** You can only add a single `parse_to_ocsf` processor per transformer. If specified, it must be the first processor in your transformer.

The `parse_to_ocsf` block supports the following arguments:

* `event_type` - (Required) Specifies the service or process that produces the log events. Allowed values are: `CloudTrail`, `Route53Resolver`, `VPCFlow`, `EKSAudit`, and `AWSWAF`.
* `ocsf_version` - (Required) Specifies the version of the OCSF schema to use for the transformed log events. The only allowed value is `V1.1`.
* `source` - (Optional) Specifies the source field to be parsed. The only allowed value is `@message`. If omitted, the whole log message is processed.

### `parse_vpc` Block

~> **Note** You can only add a single `parse_vpc` processor per transformer. If specified, it must be the first processor in your transformer.

The `parse_vpc` block supports the following arguments:

* `source` - (Optional) Specifies the source field to be parsed. The only allowed value is `@message`. If omitted, the whole log message is processed.

### `parse_waf` Block

~> **Note** You can only add a single `parse_waf` processor per transformer. If specified, it must be the first processor in your transformer.

The `parse_waf` block supports the following arguments:

* `source` - (Optional) Specifies the source field to be parsed. The only allowed value is `@message`. If omitted, the whole log message is processed.

### `rename_keys` Block

Each `rename_keys` block supports the following arguments:

* `entry` - (Required) Objects containing the information about the keys to rename. You must include at least one entry, and five at most. See [`rename_keys` `entry`](#rename_keys-entry-block) below for details.

### `rename_keys` `entry` Block

Each `entry` block supports the following arguments:

* `key` - (Required) Specifies the key to rename.
* `overwrite_if_exists` - (Optional) Specifies whether to overwrite the value if the destination key already exists. Defaults to `false`.
* `renameTo` - (Required) Specifies the new name of the key.

### `split_string` Block

Each `split_string` block supports the following arguments:

* `entry` - (Required) Objects containing the information about the fields to split. You must include at least one entry, and ten at most. See [`split_string` `entry`](#split_string-entry-block) below for details.

### `split_string` `entry` Block

Each `entry` block supports the following arguments:

* `delimiter` - (Required) Specifies the separator characters to split the string entry on.
* `source` - (Required) Specifies the key of the field to split.

### `substitute_string` Block

Each `substitute_string` block supports the following arguments:

* `entry` - (Required) Objects containing the information about the fields to substitute. You must include at least one entry, and ten at most. See [`substitute_string` `entry`](#substitute_string-entry-block) below for details.

### `substitute_string` `entry` Block

Each `entry` block supports the following arguments:

* `from` - (Required) Specifies the regular expression string to be replaced.
* `source` - (Required) Specifies the key to modify.
* `to` - (Required) Specifies the string to be substituted for each match of `from`.

### `trim_string` Block

Each `trim_string` block supports the following arguments:

* `with_keys` - (Required) Specifies the keys of the fields to trim.

### `type_converter` Block

Each `type_converter` block supports the following arguments:

* `entry` - (Required) Objects containing the information about the fields to change the type of. You must include at least one entry, and five at most. See [`type_converter` `entry`](#type_converter-entry-block) below for details.

### `type_converter` `entry` Block

Each `entry` block supports the following arguments:

* `key` - (Required) Specifies the key with the value that will be converted to a different type.
* `type` - (Required) Specifies the type to convert the field value to. Allowed values are: `integer`, `double`, `string` and `boolean`.

### `upper_case_string` Block

Each `upper_case_string` block supports the following arguments:

* `with_keys` - (Required) Specifies the keys of the fields to convert to uppercase.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs Transformer using the `log_group_arn`. For example:

```terraform
import {
  to = aws_cloudwatch_log_transformer.example
  id = "arn:aws:logs:us-west-2:123456789012:log-group:example"
}
```

Using `terraform import`, import CloudWatch Logs Transformer using the `log_group_arn`. For example:

```console
% terraform import aws_cloudwatch_log_transformer.example arn:aws:logs:us-west-2:123456789012:log-group:example
```
