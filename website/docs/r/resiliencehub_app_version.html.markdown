---
subcategory: "Resilience Hub"
layout: "aws"
page_title: "AWS: aws_resiliencehub_app_version"
description: |-
  Terraform resource for managing an AWS Resilience Hub App Version.
---

# Resource: aws_resiliencehub_app_version

Terraform resource for managing an AWS Resilience Hub App Version.

This resource defines the structure of an [`aws_resiliencehub_app`](resiliencehub_app.html.markdown) by setting the draft application template, importing resource sources, and publishing a new application version. The application template is provided as a JSON string that closely matches the [`PutDraftAppVersionTemplate` API](https://docs.aws.amazon.com/resilience-hub/latest/APIReference/API_PutDraftAppVersionTemplate.html), and can be assembled with the [`jsonencode`](https://developer.hashicorp.com/terraform/language/functions/jsonencode) function.

Because published application versions are immutable, changing any argument on this resource publishes a new version, replacing the resource. There is no API to delete an individual published version; removing this resource only removes it from Terraform state. Published versions are removed when the parent application is deleted.

## Example Usage

### Importing a CloudFormation Stack

```terraform
resource "aws_cloudformation_stack" "example" {
  name = "example-stack"

  template_body = jsonencode({
    Resources = {
      Queue = {
        Type = "AWS::SQS::Queue"
      }
    }
  })
}

resource "aws_resiliencehub_app" "example" {
  name = "example-app"
}

resource "aws_resiliencehub_app_version" "example" {
  app_arn     = aws_resiliencehub_app.example.arn
  source_arns = [aws_cloudformation_stack.example.id]

  app_template_body = jsonencode({
    resources = [{
      logicalResourceId = {
        identifier       = "Queue"
        logicalStackName = aws_cloudformation_stack.example.name
      }
      type = "AWS::SQS::Queue"
      name = "Queue"
    }]
    appComponents = [{
      name          = "appcommon"
      type          = "AWS::ResilienceHub::AppCommonAppComponent"
      resourceNames = []
      }, {
      name          = "queue"
      type          = "AWS::ResilienceHub::QueueAppComponent"
      resourceNames = ["Queue"]
    }]
    excludedResources = {
      logicalResourceIds = []
    }
    version = 2
  })
}
```

### Importing a Terraform State File

```terraform
resource "aws_resiliencehub_app_version" "example" {
  app_arn           = aws_resiliencehub_app.example.arn
  app_template_body = jsonencode({ /* ... */ })

  terraform_source {
    s3_state_file_url = "s3://example-bucket/terraform.tfstate"
  }
}
```

## Argument Reference

The following arguments are required:

* `app_arn` - (Required) ARN of the Resilience Hub application. Changing this forces a new resource to be created.
* `app_template_body` - (Required) JSON string that provides information about the application structure. See the [`appTemplateBody` API documentation](https://docs.aws.amazon.com/resilience-hub/latest/APIReference/API_PutDraftAppVersionTemplate.html) for the supported structure, including `resources`, `appComponents`, and `excludedResources`.

The following arguments are optional:

* `import_strategy` - (Optional) Strategy used to import resources into the application. Valid values are `AddOnly` and `ReplaceAll`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `source_arns` - (Optional) Set of Amazon Resource Names (ARNs) of the resources to import, such as CloudFormation stacks.
* `terraform_source` - (Optional) Terraform S3 state file sources to import. See [`terraform_source` Block](#terraform_source-block) below.
* `version_name` - (Optional) Name of the published application version.

### `terraform_source` Block

* `s3_state_file_url` - (Required) URL of the Terraform S3 state file to import.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `app_version` - Published application version.
* `identifier` - Numeric identifier of the published application version.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub App Version using the `app_arn` and `app_version` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_resiliencehub_app_version.example
  id = "arn:aws:resiliencehub:us-east-1:123456789012:app/example-app-id,release"
}
```

Using `terraform import`, import Resilience Hub App Version using the `app_arn` and `app_version` separated by a comma (`,`). For example:

```console
% terraform import aws_resiliencehub_app_version.example arn:aws:resiliencehub:us-east-1:123456789012:app/example-app-id,release
```
