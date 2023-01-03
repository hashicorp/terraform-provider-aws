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

## Argument Reference

The following arguments are supported:

* `alias` - (Optional) The alias of the prometheus workspace. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-onboard-create-workspace.html).
* `logging_configuration` - (Optional) Logging configuration for the workspace. See [Logging Configuration](#logging-configuration) below for details.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Logging Configuration

The `logging_configuration` block supports the following arguments:

* `log_group_arn` - (Required) The ARN of the CloudWatch log group to which the vended log data will be published. This log group must exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - Amazon Resource Name (ARN) of the workspace.
* `id` - Identifier of the workspace
* `prometheus_endpoint` - Prometheus endpoint available for this workspace.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

AMP Workspaces can be imported using the identifier, e.g.,

```
$ terraform import aws_prometheus_workspace.demo ws-C6DCB907-F2D7-4D96-957B-66691F865D8B
```
