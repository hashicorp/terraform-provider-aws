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

This resource supports the following arguments:

* `name` - (Required) The name of the rule group namespace
* `workspace_id` - (Required) ID of the prometheus workspace the rule group namespace should be linked to
* `data` - (Required) the rule group namespace data that you want to be applied. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-Ruler.html).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the prometheus rule group namespace using the arn. For example:

```terraform
import {
  to = aws_prometheus_rule_group_namespace.demo
  id = "arn:aws:aps:us-west-2:123456789012:rulegroupsnamespace/IDstring/namespace_name"
}
```

Using `terraform import`, import the prometheus rule group namespace using the arn. For example:

```console
% terraform import aws_prometheus_rule_group_namespace.demo arn:aws:aps:us-west-2:123456789012:rulegroupsnamespace/IDstring/namespace_name
```
