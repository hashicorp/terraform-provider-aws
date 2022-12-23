---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_workspace"
description: |-
  Gets information on an Amazon Managed Prometheus workspace.
---

# Data Source: aws_prometheus_workspace

Provides an Amazon Managed Prometheus workspace data source.

## Example Usage

### By Workspace Alias

```terraform
data "aws_prometheus_workspace" "example" {
  alias = "example"
}
```

### By Workspace ID

```terraform
data "aws_prometheus_workspace" "example" {
  workspace_id = "ws-41det8a1-2c67-6a1a-9381-9b83d3d78ef7"
}
```


## Argument Reference

* `alias` - (Optional) Prometheus workspace alias. Conflicts with `workspace_id`.
* `workspace_id` - (Optional) Prometheus workspace ID. Conflicts with `alias`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Prometheus workspace.
* `created_date` - Creation date of the Prometheus workspace.
* `prometheus_endpoint` - Endpoint of the Prometheus workspace.
* `status` - Status of the Prometheus workspace.
* `tags` - Tags assigned to the resource.
