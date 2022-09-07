---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_alert_manager_definition"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Alert Manager Definition
---

# Resource: aws_prometheus_alert_manager_definition

Manages an Amazon Managed Service for Prometheus (AMP) Alert Manager Definition

## Example Usage

```terraform
resource "aws_prometheus_workspace" "demo" {
}

resource "aws_prometheus_alert_manager_definition" "demo" {
  workspace_id = aws_prometheus_workspace.demo.id
  definition   = <<EOF
alertmanager_config: |
  route:
    receiver: 'default'
  receivers:
    - name: 'default'
EOF
}
```

## Argument Reference

The following arguments are supported:

* `workspace_id` - (Required) ID of the prometheus workspace the alert manager definition should be linked to
* `definition` - (Required) the alert manager definition that you want to be applied. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-alert-manager.html).

## Attributes Reference

No additional attributes are exported.

## Import

The prometheus alert manager definition can be imported using the workspace identifier, e.g.,

```
$ terraform import aws_prometheus_alert_manager_definition.demo ws-C6DCB907-F2D7-4D96-957B-66691F865D8B
```
