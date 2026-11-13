---
subcategory: "Continuum"
layout: "aws"
page_title: "AWS: aws_continuum_code_findings"
description: |-
  Lists the code findings of an AWS Continuum environment.
---

# Data Source: aws_continuum_code_findings

Lists the code findings of an AWS Continuum environment.

## Example Usage

```terraform
data "aws_continuum_code_findings" "example" {
  environment_id = aws_continuum_environment.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `environment_id` - (Required) ID of the environment.
* `region` - (Optional) Region where this data source will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `findings` - List of code findings. Each has:
    * `category`
    * `completed_at`
    * `created_at`
    * `created_issue_id`
    * `discard_reason`
    * `discard_reasoning`
    * `file_path`
    * `finding_arn`
    * `finding_id`
    * `scan_run_id`
    * `severity`
    * `source`
    * `status`
    * `status_message`
    * `title`
    * `type`
* `id` - ID of the environment.
