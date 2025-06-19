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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `data_set_id` - (Required) Identifier for the data set.
* `aws_account_id` - (Optional) AWS account ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

See the [Data Set Resource](/docs/providers/aws/r/quicksight_data_set.html) for details on the
returned attributes - they are identical.
