---
subcategory: "CloudWatch Observability Admin"
layout: "aws"
page_title: "AWS: aws_observabilityadmin_telemetry_rule_for_organization"
description: |-
  List CloudWatch Observability Admin Telemetry Rules For Organization.
---

# List Resource: aws_observabilityadmin_telemetry_rule_for_organization

List CloudWatch Observability Admin Telemetry Rules For Organization.

## Example Usage

### Basic Usage

```terraform
query "aws_observabilityadmin_telemetry_rule_for_organization" "example" {}
```

### Include Resource Objects

```terraform
query "aws_observabilityadmin_telemetry_rule_for_organization" "example" {
  include_resource = true
}
```

## Argument Reference

The following arguments are optional:

* `include_resource` - (Optional) Whether to include the full resource object in the results. Defaults to `false`.
* `region` - (Optional) AWS region. If not specified, the provider region is used.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `resources` - List of telemetry rules for organization. Each resource has the following attributes:
  * `display_name` - Display name of the telemetry rule.
  * `identity` - Unique identifier for the telemetry rule.
  * `resource` - (Only when `include_resource` is `true`) Full resource object with all attributes from the [`aws_observabilityadmin_telemetry_rule_for_organization` resource](/docs/providers/aws/r/observabilityadmin_telemetry_rule_for_organization.html).