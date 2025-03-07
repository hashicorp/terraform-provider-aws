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

* `data` - (Required) the rule group namespace data that you want to be applied. See more [in AWS Docs](https://docs.aws.amazon.com/prometheus/latest/userguide/AMP-Ruler.html).
* `name` - (Required) The name of the rule group namespace.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `workspace_id` - (Required) ID of the prometheus workspace the rule group namespace should be linked to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the rule group namespace.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

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
