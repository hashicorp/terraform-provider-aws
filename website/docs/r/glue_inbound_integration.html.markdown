---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_inbound_integration"
description: |-
  Manages an AWS Glue Inbound Integration (Zero-ETL) between a source (e.g., DynamoDB) and a target (e.g., SageMaker Lakehouse).
---

# Resource: aws_glue_inbound_integration

Manages an AWS Glue Inbound Integration (Zero-ETL) between a source and a target. Use this to configure DynamoDB to SageMaker Lakehouse zero‑ETL via Glue.

Refer to AWS documentation for prerequisites, IAM and resource policies:

- Glue InboundIntegration: https://docs.aws.amazon.com/glue/latest/webapi/API_InboundIntegration.html
- DynamoDB → SageMaker Lakehouse zero‑ETL: https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/amazon-sagemaker-lakehouse-for-DynamoDB-zero-etl.html

## Example Usage

```hcl
resource "aws_dynamodb_table" "example" {
  name           = "example"
  hash_key       = "pk"
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "pk"
    type = "S"
  }

  point_in_time_recovery { enabled = true }
}

resource "aws_glue_inbound_integration" "example" {
  integration_name = "example"
  source_arn       = aws_dynamodb_table.example.arn
  target_arn       = var.target_arn
}
```

## Argument Reference

- `integration_name` (Required) Name of the integration.
- `source_arn` (Required) ARN of the source resource (for zero‑ETL, typically a DynamoDB table ARN).
- `target_arn` (Required) ARN of the target resource (for zero‑ETL, e.g., SageMaker Lakehouse target).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `arn` – ARN of the integration.

## Timeouts

The `timeouts` block supports the following:

- `create`
- `update`
- `delete`

## Import

Glue inbound integrations can be imported by `arn`:

```sh
terraform import aws_glue_inbound_integration.example arn:aws:glue:region:account:integration/ID
```


