---
subcategory: "CodeBuild"
layout: "aws"
page_title: "AWS: aws_codebuild_report_group"
description: |-
  Provides a CodeBuild Report Group resource.
---

# Resource: aws_codebuild_report_group

Provides a CodeBuild Report Groups Resource.

## Example Usage

```terraform
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "Enable IAM User Permissions"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
    }

    actions   = ["kms:*"]
    resources = ["*"]
  }
}
resource "aws_kms_key" "example" {
  description             = "my test kms key"
  deletion_window_in_days = 7
  policy                  = data.aws_iam_policy_document.example.json
}

resource "aws_s3_bucket" "example" {
  bucket = "my-test"
}

resource "aws_codebuild_report_group" "example" {
  name = "my test report group"
  type = "TEST"

  export_config {
    type = "S3"

    s3_destination {
      bucket              = aws_s3_bucket.example.id
      encryption_disabled = false
      encryption_key      = aws_kms_key.example.arn
      packaging           = "NONE"
      path                = "/some"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of a Report Group.
* `type` - (Required) The type of the Report Group. Valid value are `TEST` and `CODE_COVERAGE`.
* `export_config` - (Required) Information about the destination where the raw data of this Report Group is exported. see [Export Config](#export-config) documented below.
* `delete_reports` - (Optional) If `true`, deletes any reports that belong to a report group before deleting the report group. If `false`, you must delete any reports in the report group before deleting it. Default value is `false`.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Export Config

* `type` - (Required) The export configuration type. Valid values are `S3` and `NO_EXPORT`.
* `s3_destination` - (Required) contains information about the S3 bucket where the run of a report is exported. see [S3 Destination](#s3-destination) documented below.

#### S3 Destination

* `bucket`- (Required) The name of the S3 bucket where the raw data of a report are exported.
* `encryption_key` - (Required) The encryption key for the report's encrypted raw data. The KMS key ARN.
* `encryption_disabled`- (Optional) A boolean value that specifies if the results of a report are encrypted.
 **Note: the API does not currently allow setting encryption as disabled**
* `packaging` - (Optional) The type of build output artifact to create. Valid values are: `NONE` (default) and `ZIP`.
* `path` - (Optional) The path to the exported report's raw data results.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ARN of Report Group.
* `arn` - The ARN of Report Group.
* `created` - The date and time this Report Group was created.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CodeBuild Report Group using the CodeBuild Report Group arn. For example:

```terraform
import {
  to = aws_codebuild_report_group.example
  id = "arn:aws:codebuild:us-west-2:123456789:report-group/report-group-name"
}
```

Using `terraform import`, import CodeBuild Report Group using the CodeBuild Report Group arn. For example:

```console
% terraform import aws_codebuild_report_group.example arn:aws:codebuild:us-west-2:123456789:report-group/report-group-name
```
