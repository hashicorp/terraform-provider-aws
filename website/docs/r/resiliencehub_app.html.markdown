---
subcategory: "Resilience Hub"
layout: "aws"
page_title: "AWS: aws_resiliencehub_app"
description: |-
  Terraform resource for managing an AWS Resilience Hub App.
---

# Resource: aws_resiliencehub_app

Terraform resource for managing an AWS Resilience Hub App.

## Example Usage

### Basic Usage

```terraform
resource "aws_resiliencehub_app" "example" {
  name                = "example-app"
  assessment_schedule = "Disabled"

  app_template {
    version = "2.0"

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }
  }
}
```

### Complete Usage with Resources and Terraform Source

```terraform
resource "aws_resiliencehub_app" "example" {
  name                    = "example-app"
  description             = "Example app with Terraform source"
  assessment_schedule = "Disabled"

  app_template {
    version = "2.0"

    resource {
      name = "lambda-function"
      type = "AWS::Lambda::Function"

      logical_resource_id {
        identifier            = "MyLambda"
        terraform_source_name = "my-terraform-source"
      }
    }

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }

    app_component {
      name           = "compute-tier"
      type           = "AWS::ResilienceHub::ComputeAppComponent"
      resource_names = ["lambda-function"]
    }
  }

  resource_mapping {
    mapping_type          = "Terraform"
    resource_name         = "lambda-function"
    terraform_source_name = "my-terraform-source"

    physical_resource_id {
      type       = "Native"
      identifier = "s3://${aws_s3_bucket.example.bucket}/terraform.tfstate"
    }
  }

  depends_on = [aws_s3_object.tfstate, aws_s3_bucket_policy.example]

  tags = {
    Environment = "example"
  }
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "example" {
  bucket        = "example-terraform-state-bucket"
  force_destroy = true
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "resiliencehub.amazonaws.com"
        }
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.example.arn,
          "${aws_s3_bucket.example.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_s3_object" "tfstate" {
  bucket = aws_s3_bucket.example.bucket
  key    = "terraform.tfstate"
  content = jsonencode({
    version           = 4
    terraform_version = "1.0.0"
    serial            = 1
    lineage           = "example"
    outputs           = {}
    resources = [
      {
        mode     = "managed"
        type     = "aws_lambda_function"
        name     = "example"
        provider = "provider[\"registry.terraform.io/hashicorp/aws\"]"
        instances = [
          {
            schema_version = 0
            attributes = {
              function_name = "example-function"
              arn           = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:example-function"
            }
          }
        ]
      }
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the application. Must be 2-60 characters, start with alphanumeric, contain only alphanumeric, underscore, and hyphen.
* `app_template` - (Required) Application template configuration. See [app_template](#app_template) below.

The following arguments are optional:

* `app_assessment_schedule` - (Optional) Assessment schedule for the application. Valid values are `Disabled` and `Daily`.
* `description` - (Optional) Description of the application. Maximum 500 characters.
* `region` - (Optional) AWS region where the application will be created.
* `resiliency_policy_arn` - (Optional) ARN of the resiliency policy to associate with the application.
* `resource_mapping` - (Optional) Resource mapping configuration. See [resource_mapping](#resource_mapping) below.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### app_template

* `version` - (Required) Version of the application template.
* `additional_info` - (Optional) Additional information about the application template as a map of key-value pairs.
* `app_component` - (Optional) Application components. See [app_component](#app_component) below.
* `resource` - (Optional) Resources in the application. See [resource](#resource) below.

### app_component

* `name` - (Required) Name of the application component.
* `type` - (Required) Type of the application component. Valid values include `AWS::ResilienceHub::AppCommonAppComponent`, `AWS::ResilienceHub::ComputeAppComponent`, `AWS::ResilienceHub::DatabaseAppComponent`, `AWS::ResilienceHub::NetworkingAppComponent`, and `AWS::ResilienceHub::StorageAppComponent`.
* `resource_names` - (Optional) List of resource names associated with this component.
* `additional_info` - (Optional) Additional information about the application component as a map of key-value pairs.

### resource

* `name` - (Required) Name of the resource.
* `type` - (Required) Type of the resource (e.g., `AWS::Lambda::Function`, `AWS::RDS::DBInstance`).
* `additional_info` - (Optional) Additional information about the resource as a map of key-value pairs.
* `logical_resource_id` - (Required) Logical resource identifier. See [logical_resource_id](#logical_resource_id) below.

### logical_resource_id

* `identifier` - (Required) Identifier for the logical resource.
* `logical_stack_name` - (Optional) Name of the logical stack.
* `resource_group_name` - (Optional) Name of the resource group.
* `terraform_source_name` - (Optional) Name of the Terraform source.
* `eks_source_name` - (Optional) Name of the EKS source.

### resource_mapping

* `mapping_type` - (Required) Type of resource mapping. Valid values are `CfnStack`, `Resource`, `Terraform`, and `EKS`.
* `resource_name` - (Required) Name of the resource.
* `terraform_source_name` - (Optional) Name of the Terraform source.
* `physical_resource_id` - (Required) Physical resource identifier. See [physical_resource_id](#physical_resource_id) below.

### physical_resource_id

* `type` - (Required) Type of the physical resource identifier. Valid values are `Arn` and `Native`.
* `identifier` - (Required) Identifier of the physical resource.
* `aws_account_id` - (Optional) AWS account ID where the physical resource is located.
* `aws_region` - (Optional) AWS region where the physical resource is located.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the application.
* `drift_status` - Drift status of the application.
* `id` - ARN of the application.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Resilience Hub App using the `arn`. For example:

```terraform
import {
  to = aws_resiliencehub_app.example
  id = "arn:aws:resiliencehub:us-east-1:123456789012:app/example-app-id"
}
```

Using `terraform import`, import Resilience Hub App using the `arn`. For example:

```console
% terraform import aws_resiliencehub_app.example arn:aws:resiliencehub:us-east-1:123456789012:app/example-app-id
```
