---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_code_scans"
description: |-
  Lists the code scans of an AWS Continuum environment.
---

# Data Source: aws_continuum_code_scans

Lists the code scans of an AWS Continuum environment.

## Example Usage

```terraform
data "aws_continuum_code_scans" "example" {
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
* `scans` - List of code scans. Each has:
    * `completed_at`
    * `created_at`
    * `failure_message`
    * `failure_stage`
    * `origin`
    * `scan_run_id`
    * `status`
