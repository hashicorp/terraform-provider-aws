---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_workspace"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Workspace
---

# Resource: aws_prometheus_workspace

Manages an Amazon Managed Service for Prometheus (AMP) Workspace.

## Example Usage

```terraform
resource "aws_prometheus_workspace" "example" {
  alias = "example"

  tags = {
    Environment = "production"
  }
}
```

### CloudWatch Logging

```terraform
resource "aws_cloudwatch_log_group" "example" {
  name = "example"
}

resource "aws_prometheus_workspace" "example" {
  logging_configuration {
    log_group_arn = "${aws_cloudwatch_log_group.example.arn}:*"
  }
}
```

### AWS KMS Customer Managed Keys (CMK)

```terraform
resource "aws_prometheus_workspace" "example" {
  alias       = "example"
  kms_key_arn = aws_kms_key.example.arn
}

resource "aws_kms_key" "example" {
  description             = "example"
  deletion_window_in_days = 7
}
```

## Argument Reference

This resource supports the following arguments:

* `alias` - (Optional) The alias of the prometheus workspace. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-onboard-create-workspace.html).
* `kms_key_arn` - (Optional) The ARN for the KMS encryption key. If this argument is not provided, then the AWS owned encryption key will be used to encrypt the data in the workspace. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/encryption-at-rest-Amazon-Service-Prometheus.html)
* `logging_configuration` - (Optional) Logging configuration for the workspace. See [Logging Configuration](#logging-configuration) below for details.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Logging Configuration

The `logging_configuration` block supports the following arguments:

* `log_group_arn` - (Required) The ARN of the CloudWatch log group to which the vended log data will be published. This log group must exist.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the workspace.
* `id` - Identifier of the workspace
* `prometheus_endpoint` - Prometheus endpoint available for this workspace.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AMP Workspaces using the identifier. For example:

```terraform
import {
  to = aws_prometheus_workspace.demo
  id = "ws-C6DCB907-F2D7-4D96-957B-66691F865D8B"
}
```

Using `terraform import`, import AMP Workspaces using the identifier. For example:

```console
% terraform import aws_prometheus_workspace.demo ws-C6DCB907-F2D7-4D96-957B-66691F865D8B
```
