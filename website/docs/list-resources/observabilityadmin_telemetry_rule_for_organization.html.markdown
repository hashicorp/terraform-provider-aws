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

```terraform
list "aws_observabilityadmin_telemetry_rule_for_organization" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
