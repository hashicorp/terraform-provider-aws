---
subcategory: "OpenSearch Ingestion (OSIS)"
layout: "aws"
page_title: "AWS: aws_osis_pipeline_resource_policy"
description: |-
  Terraform resource for managing an AWS OpenSearch Ingestion Pipeline Resource Policy.
---

# Resource: aws_osis_pipeline_resource_policy

Terraform resource for managing an AWS OpenSearch Ingestion Pipeline Resource Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

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

resource "aws_osis_pipeline_resource_policy" "example" {
  resource_arn = aws_osis_pipeline.example.pipeline_arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "osis:Ingest"
        Resource = aws_osis_pipeline.example.pipeline_arn
      }
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) JSON-formatted resource policy to attach to the pipeline.
* `resource_arn` - (Required) ARN of the pipeline to attach the resource policy to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ARN of the pipeline.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Ingestion Pipeline Resource Policy using the `resource_arn`. For example:

```terraform
import {
  to = aws_osis_pipeline_resource_policy.example
  id = "arn:aws:osis:us-east-1:123456789012:pipeline/example"
}
```

Using `terraform import`, import OpenSearch Ingestion Pipeline Resource Policy using the `resource_arn`. For example:

```console
% terraform import aws_osis_pipeline_resource_policy.example arn:aws:osis:us-east-1:123456789012:pipeline/example
```
