---
subcategory: "AMP (Managed Prometheus)"
layout: "aws"
page_title: "AWS: aws_prometheus_rule_group_namespace"
description: |-
  Manages an Amazon Managed Service for Prometheus (AMP) Rule Group Namespace
---

# Resource: aws_prometheus_rule_group_namespace

Manages an Amazon Managed Service for Prometheus (AMP) Rule Group Namespace

## Example Usage

```terraform
resource "aws_prometheus_workspace" "demo" {
}

resource "aws_prometheus_rule_group_namespace" "demo" {
  name         = "rules"
  workspace_id = aws_prometheus_workspace.demo.id
  data         = <<EOF
groups:
  - name: test
    rules:
    - record: metric:recording_rule
      expr: avg(rate(container_cpu_usage_seconds_total[5m]))
EOF
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the rule group namespace
* `workspace_id` - (Required) ID of the prometheus workspace the rule group namespace should be linked to
* `data` - (Required) the rule group namespace data that you want to be applied. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-Ruler.html).

## Attributes Reference

No additional attributes are exported.

## Import

The prometheus rule group namespace can be imported using the arn, e.g.,

```
$ terraform import aws_prometheus_rule_group_namespace.demo arn:aws:aps:us-west-2:123456789012:rulegroupsnamespace/IDstring/namespace_name
```
