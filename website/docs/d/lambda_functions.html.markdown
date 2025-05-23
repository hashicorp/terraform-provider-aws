---
subcategory: "Lambda"
layout: "aws"
page_title: "AWS: aws_lambda_functions"
description: |-
  Terraform data resource to get a list of Lambda Functions.
---

# Data Source: aws_lambda_functions

Terraform data resource to get a list of Lambda Functions.

## Example Usage

```terraform
data "aws_lambda_functions" "all" {}
```

## Argument Reference

This data source supports the following arguments:

* `region` – (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `function_names` - A list of Lambda Function names.
* `function_arns` - A list of Lambda Function ARNs.
