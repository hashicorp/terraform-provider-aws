---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_workspace"
description: |-
  Gets information on an Amazon Managed Prometheus workspace.
---

# Data Source: aws_amp_workspace

Provides an Amazon Managed Prometheus workspace data source.

## Example Usage

### Basic configuration

```terraform
data "aws_amp_workspace" "example" {
  workspace_id = "ws-41det8a1-2c67-6a1a-9381-9b83d3d78ef7"
}
```

## Argument Reference

The following arguments are required:

* `workspace_id` - (Required) The Prometheus workspace ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the Prometheus workspace.
* `created_date` - The creation date of the Prometheus workspace.
* `prometheus_endpoint` - The endpoint of the Prometheus workspace.
* `alias` - The Prometheus workspace alias.
* `status` - The status of the Prometheus workspace.
* `tags` - The tags assigned to the resource.
