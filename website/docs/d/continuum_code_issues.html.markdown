---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_code_issues"
description: |-
  Lists the code issues of an AWS Continuum environment.
---

# Data Source: aws_continuum_code_issues

Lists the code issues of an AWS Continuum environment.

## Example Usage

```terraform
data "aws_continuum_code_issues" "example" {
  environment_id = aws_continuum_environment.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `environment_id` - (Required) ID of the environment.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the environment.
* `issues` - List of code issues. Each has:
    * `confidence`
    * `created_at`
    * `issue_arn`
    * `issue_id`
    * `name`
    * `phase`
    * `severity`
    * `status`
    * `status_message`
    * `status_reason`
    * `status_updated_at`
    * `updated_at`
