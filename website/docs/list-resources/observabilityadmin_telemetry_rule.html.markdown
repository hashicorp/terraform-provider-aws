---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_rule"
description: |-
  Lists CloudWatch Observability Admin Telemetry Rule resources.
---

# List Resource: aws_observabilityadmin_telemetry_rule

Lists CloudWatch Observability Admin Telemetry Rule resources.

## Example Usage

```terraform
list "aws_observabilityadmin_telemetry_rule" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
