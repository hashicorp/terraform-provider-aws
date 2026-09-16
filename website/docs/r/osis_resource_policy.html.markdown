---
subcategory: "OpenSearch Ingestion (OSIS)"
layout: "aws"
page_title: "AWS: aws_osis_resource_policy"
description: |-
  Terraform resource for managing an AWS OpenSearch Ingestion Resource Policy.
---

# Resource: aws_osis_resource_policy

Terraform resource for managing an AWS OpenSearch Ingestion Resource Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_osis_resource_policy" "example" {
  resource_arn = aws_osis_pipeline.example.pipeline_arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "osis:CreatePipelineEndpoint"
        Resource = aws_osis_pipeline.example.pipeline_arn
      }
    ]
  })
}

resource "aws_osis_pipeline" "example" {
  pipeline_name               = "example"
  pipeline_configuration_body = <<-EOT
            version: "2"
            example-pipeline:
              source:
                http:
                  path: "/example"
              sink:
                - s3:
                    aws:
                      sts_role_arn: "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/Example"
                      region: "us-east-1"
                    bucket: "example"
                    threshold:
                      event_collect_timeout: "60s"
                    codec:
                      ndjson:
        EOT
  max_units                   = 1
  min_units                   = 1
}

data "aws_caller_identity" "current" {}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) JSON-formatted policy to attach to the resource.
* `resource_arn` - (Required) ARN of the resource to attach the policy to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_osis_resource_policy.example
  identity = {
    resource_arn = "arn:aws:osis:us-east-1:123456789012:pipeline/example"
  }
}

resource "aws_osis_resource_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` (String) ARN of the resource the policy is attached to.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Ingestion Resource Policy using the `resource_arn`. For example:

```terraform
import {
  to = aws_osis_resource_policy.example
  id = "arn:aws:osis:us-east-1:123456789012:pipeline/example"
}
```

Using `terraform import`, import OpenSearch Ingestion Resource Policy using the `resource_arn`. For example:

```console
% terraform import aws_osis_resource_policy.example arn:aws:osis:us-east-1:123456789012:pipeline/example
```
