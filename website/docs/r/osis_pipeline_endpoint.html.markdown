---
subcategory: "OpenSearch Ingestion (OSIS)"
layout: "aws"
page_title: "AWS: aws_osis_pipeline_endpoint"
description: |-
  Terraform resource for managing an AWS OpenSearch Ingestion Pipeline Endpoint.
---

# Resource: aws_osis_pipeline_endpoint

Terraform resource for managing an AWS OpenSearch Ingestion Pipeline Endpoint.

## Example Usage

### Basic Usage

```terraform
resource "aws_osis_pipeline_endpoint" "example" {
  pipeline_arn = aws_osis_pipeline.example.pipeline_arn

  vpc_options {
    security_group_ids = [aws_security_group.example.id]
    subnet_ids         = [aws_subnet.example.id]
  }
}
```

## Argument Reference

The following arguments are required:

* `pipeline_arn` - (Required) ARN of the pipeline to create the endpoint for.
* `vpc_options` - (Required) VPC options for the pipeline endpoint. See [`vpc_options` Block](#vpc_options-block) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `vpc_options` Block

* `security_group_ids` - (Optional) List of security groups associated with the VPC endpoint.
* `subnet_ids` - (Required) List of subnet IDs associated with the VPC endpoint.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier for the endpoint.
* `status` - Status of the endpoint.
* `vpc_id` - ID of the VPC.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `15m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import OpenSearch Ingestion Pipeline Endpoint using the `id`. For example:

```terraform
import {
  to = aws_osis_pipeline_endpoint.example
  id = "endpoint-id"
}
```

Using `terraform import`, import OpenSearch Ingestion Pipeline Endpoint using the `id`. For example:

```console
% terraform import aws_osis_pipeline_endpoint.example endpoint-id
```
