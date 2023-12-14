---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_resource_policy"
description: |-
  Provides a CodeBuild Resource Policy resource.
---

# Resource: aws_codebuild_resource_policy

Provides a CodeBuild Resource Policy Resource.

## Example Usage

```terraform
resource "aws_codebuild_report_group" "example" {
  name = "example"
  type = "TEST"

  export_config {
    type = "NO_EXPORT"
  }
}

data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_codebuild_resource_policy" "example" {
  resource_arn = aws_codebuild_report_group.example.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = "default"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action = [
        "codebuild:BatchGetReportGroups",
        "codebuild:BatchGetReports",
        "codebuild:ListReportsForReportGroup",
        "codebuild:DescribeTestCases",
      ]
      Resource = aws_codebuild_report_group.example.arn
    }]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `resource_arn` - (Required) The ARN of the Project or ReportGroup resource you want to associate with a resource policy.
* `policy` - (Required) A JSON-formatted resource policy. For more information, see [Sharing a Projec](https://docs.aws.amazon.com/codebuild/latest/userguide/project-sharing.html#project-sharing-share) and [Sharing a Report Group](https://docs.aws.amazon.com/codebuild/latest/userguide/report-groups-sharing.html#report-groups-sharing-share).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of Resource.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeBuild Resource Policy using the CodeBuild Resource Policy arn. For example:

```terraform
import {
  to = aws_codebuild_resource_policy.example
  id = "arn:aws:codebuild:us-west-2:123456789:report-group/report-group-name"
}
```

Using `terraform import`, import CodeBuild Resource Policy using the CodeBuild Resource Policy arn. For example:

```console
% terraform import aws_codebuild_resource_policy.example arn:aws:codebuild:us-west-2:123456789:report-group/report-group-name
```
