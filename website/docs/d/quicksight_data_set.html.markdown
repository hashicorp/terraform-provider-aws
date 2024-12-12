---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_data_set"
description: |-
  Use this data source to fetch information about a QuickSight Data Set.
---

# Data Source: aws_quicksight_data_set

Data source for managing a QuickSight Data Set.

## Example Usage

### Basic Usage

```terraform
data "aws_quicksight_data_set" "example" {
  data_set_id = "example-id"
}
```

## Argument Reference

The following arguments are required:

* `data_set_id` - (Required) Identifier for the data set.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID.

## Attribute Reference

See the [Data Set Resource](/docs/providers/aws/r/quicksight_data_set.html) for details on the
returned attributes - they are identical.
