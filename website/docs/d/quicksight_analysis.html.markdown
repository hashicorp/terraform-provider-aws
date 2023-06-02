---
subcategory: "QuickSight"
layout: "aws"
page_title: "AWS: aws_quicksight_analysis"
description: |-
  Use this data source to fetch information about a QuickSight Analysis.
---

# Data Source: aws_quicksight_analysis

Terraform data source for managing an AWS QuickSight Analysis.

## Example Usage

### Basic Usage

```terraform
data "aws_quicksight_analysis" "example" {
  analysis_id = "example-id"
}
```

## Argument Reference

The following arguments are required:

* `analysis_id` - (Required) Identifier for the analysis.

The following arguments are optional:

* `aws_account_id` - (Optional) AWS account ID.

## Attributes Reference

See the [Analysis Resource](/docs/providers/aws/r/quicksight_analysis.html) for details on the
returned attributes - they are identical.
