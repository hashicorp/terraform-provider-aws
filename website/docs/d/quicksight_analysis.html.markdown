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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `analysis_id` - (Required) Identifier for the analysis.
* `aws_account_id` - (Optional) AWS account ID.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

See the [Analysis Resource](/docs/providers/aws/r/quicksight_analysis.html) for details on the
returned attributes - they are identical.
