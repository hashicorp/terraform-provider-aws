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
resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_cloudwatch_log_transformer" "example" {
  log_group_identifier = aws_cloudwatch_log_group.example.name
  transformer_config {
    parse_json {}
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `log_group_name` - (Required) Log group name or ARN to set the transformer for.
* `transformer_config` - (Required) Specifies the configuration of the transformer. You must include at least one configuration, and 20 at most. See [`transformer_config`](#transformer_config-block) below for details.

### `transformer_config` Block

~> **Note** You must only specify a single processor per `transformer_config` block. Besides, the first processor in a transformer must be a parser.

Each `transformer_config` supports the following arguments:

* `add_keys` - (Optional) Adds new key-value pairs to the log event. See [`add_keys`](#add_keys-block) below for details.
* `copy_value` - (Optional) Copies values within a log event. See [`copy_value`](#copy_value-block) below for details.
* `csv` - (Optional) Parses comma-separated values (CSV) from the log events into columns. See [`csv`](#csv-block) below for details.
* `date_time_converter` - (Optional) Converts a datetime string into a format that you specify. See [`date_time_converter`](#date_time_converter-block) below for details.
* `delete_keys` - (Optional) Deletes entries from a log event. See [`delete_keys`](#delete_keys-block) below for details.
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

~> **Note** You can only add a single `add_keys` processor.

The `add_keys` block supports the following arguments:

* `entries` - (Required) Objects containing the information about the keys to add to the log event. You must include at least one entry, and five at most. See [`add_keys` `entries`](#add_keys-entries-block) below for details.

### `add_keys` `entries` Block

Each `entries` block supports the following arguments:

* `key` - (Required) Specifies the key of the new entry to be added to the log event.
* `overwrite_if_exists` - (Optional) Specifies whether to overwrite the value if the key already exists in the log event. Defaults to `false`.
* `value` - (Required) Specifies the value of the new entry to be added to the log event.

### `copy_value` Block

Each `copy_value` block supports the following arguments:

* `entries` - (Required) Objects containing the information about the values to copy to the log event. You must include at least one entry, and five at most. See [`copy_value` `entries`](#substitute_string-block) below for details.

### `copy_value` `entries` Block

Each `entries` block supports the following arguments:

* `overwrite_if_exists` - (Optional) Specifies whether to overwrite the value if the destination key already exists. Defaults to `false`.
* `source` - (Required) Specifies the key to copy.
* `target` - (Required) Specifies the key of the field to copy the value to.

### `csv` Block

Each `csv` block supports the following arguments:

* `columns` - (Optional) Specifies the names to use for the columns in the transformed log event. If not specified, default column names (`[column_1, column_2 ...]`) are used.
* `delimiter` - (Optional) Specifies the character used to separate each column in the original comma-separated value log event. Defaults to the comma `,` character.
* `quote_character` - (Optional) Specifies the character used as a text qualifier for a single column of data. Defaults to the double quotation mark `"` character.
* `source` - (Optional) Specifies the path to the field in the log event that has the comma separated values to be parsed. If ommited, the whole log message is processed.

### `date_time_converter` Block

Each `date_time_converter` block supports the following arguments:

* `locale` - (Optional) Specifies the locale of the source field. Defaults to `locale.ROOT`.
* `match_patterns` - (Required) Specifies the list of patterns to match against the `source` field.
* `source` - (Required) Specifies the key to apply the date conversion to.
* `source_timezone` - (Optional) Specifies the time zone of the source field. Defaults to `UTC`.
* `target` - (Required) Specifies the JSON field to store the result in.
* `target_format` - (Optional) Specifies the datetime format to use for the converted data in the target field. Defaults to `yyyy-MM-dd'T'HH:mm:ss.SSS'Z`.
* `target_timezone` - (Optional) Specifies the time zone of the target field. Defaults to `UTC`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs Transformer using the `log_group_identifier`. For example:

```terraform
import {
  to = aws_cloudwatch_log_transformer.example
  id = "/aws/log/group/name"
}
```

Using `terraform import`, import CloudWatch Logs Transformer using the `log_group_identifier`. For example:

```console
% terraform import aws_cloudwatch_log_transformer.example /aws/log/group/name
```
