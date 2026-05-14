---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_rule_for_organization"
description: |-
  Lists CloudWatch Observability Admin Telemetry Rule For Organization resources.
---

# List Resource: aws_observabilityadmin_telemetry_rule_for_organization

Lists CloudWatch Observability Admin Telemetry Rule For Organization resources.

## Example Usage

### Basic Usage

```terraform
list "aws_observabilityadmin_telemetry_rule_for_organization" "example" {
  provider = aws
}
```

### Include Resource Objects

```terraform
list "aws_observabilityadmin_telemetry_rule_for_organization" "example" {
  provider         = aws
  include_resource = true
}
```

## Argument Reference

This list resource supports the following arguments:

* `include_resource` - (Optional) Whether to include the full resource object in the results. Defaults to `false`.
* `region` - (Optional) Region to query. Defaults to provider region.