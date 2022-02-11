---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_data_set"
description: |-
  Manages a Resource QuickSight Data Set.
---

# Resource: aws_quicksight_data_set

Resource for managing QuickSight Data Set

## Example Usage

```terraform
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id       = "example-id"
	name              = "example-name"
	import_mode       = "SPICE"
    physical_table_id = "example-physical-table-id"
	physical_table_map {
		s3_source {
			data_source_arn = "my data source arn"
			input_columns {
				name = "example-column"
				type = "STRING"
			}
		}
	}
}
```

## Argument Reference

The following arguments are required:

* `data_set_id` - (Required, Forces new resource) An identifier for the data set.
* `import_mode` - (Required) Indicates whether you want to import the data into SPICE. Must be either `SPICE` or `DIRECT_QUERY`
* `name` - (Required) The display name for the dataset. maximum length of 128 characters.
* `physical_table_map` - (Required) Declares the physical tables that are available in the underlying data sources. Maximum of 1 entry.
* `physical_table_id` - (Required) Declares the ID of the physical table map.

The following arguments are optional:

* `aws_account_id` - (Optional, Forces new resource) The ID for the AWS account that the data set is in. Currently, you use the ID for the AWS account that contains your Amazon QuickSight account.
* `column_groups` - (Optional) Groupings of columns that work together in certain Amazon QuickSight features. Currently, only geospatial hierarchy is supported. Maximum number of 8 items.
* `column_level_permission_rules` - (Optional) A set of 1 or more definitions of a [ColumnLevelPermissionRule](https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnLevelPermissionRule.html)
* `data_set_usage_configuration` - (Optional) The usage configuration to apply to child datasets that reference this dataset as a source.
* `field_folders` - (Optional) The folder that contains fields and nested subfolders for your dataset. Maximum of 1 entry.
* `logical_table_map` - (Optional) Configures the combination and transformation of the data from the physical tables. Maximum of 1 entry.
* `permission` - (Optional) A set of resource permissions on the data source. Maximum of 64 items. See [Permission](#permission-argument-reference) below for more details.
* `row_level_permission_data_set` - (Optional) The row-level security configuration for the data that you want to create.
* `row_level_permission_tag_configuration` - (Optional) The configuration of tags on a dataset to set row-level security. Row-level security tags are currently supported for anonymous embedding only.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the data source
* `output_columns` - The columns in the data set.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

A QuickSight data source can be imported using the AWS account ID, and data source ID name separated by a slash (`/`) e.g.,

```
$ terraform import aws_quicksight_data_set.example 123456789123/my-data-set-id
```
