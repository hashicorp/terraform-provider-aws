---
subcategory: "EventBridge"
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_endpoint"
description: |-
  Provides a resource to create an EventBridge Global Endpoint.
---

# Resource: aws_cloudwatch_event_endpoint

Provides a resource to create an EventBridge Global Endpoint.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```hcl
resource "aws_cloudwatch_event_endpoint" "this" {
  name        = "global-endpoint"
  role_arn    = aws_iam_role.replication.arn
  event_buses = [aws_cloudwatch_event_bus.primary.arn, aws_cloudwatch_event_bus.secondary.arn]

  replication_config {
    is_enabled = false
  }

  routing_config {
    failover_config {
      primary {
        health_check_arn = aws_route53_health_check.primary.arn
      }
      secondary {
        route = "us-east-2"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the global endpoint.
* `description` - (Optional) A description of the global endpoint.
* `role_arn` - (Required) The ARN of the role used for replication between event buses.
* `event_buses` - (Required) The event buses to use. The names of the event buses must be identical in each Region. Minimum of two event buses are required.
* `replication_config` - (Optional) Parameters used for replication. A maximum of 1 are allowed. Documented below.
* `routing_config` - (Required) Parameters used for routing. A maximum of 1 are allowed. Documented below.

`replication_config` support the following:

* `is_enabled` - (Optional) Enable or disable event replication between buses. The default state is DISABLED. If you want event replication enabled, set the state to ENABLED. You will need a `role_arn` if replication is ENABLED.

`routing_config` support the following:

* `failover_config` - (Required) Parameters used for failover. This includes what triggers failover and what happens when it's triggered. A maximum of 1 are allowed. Documented below.

`failover_config` support the following:

* `primary` - (Required) Parameters used for primary region. A maximum of 1 are allowed. Documented below.
* `secondary` - (Required) Parameters used for secondary region. The Region that events are routed to when failover is triggered or event replication is enabled. A maximum of 1 are allowed. Documented below.

`primary` support the following:

* `health_check_arn` - (Required) The ARN of the health check used by the endpoint to determine whether failover is triggered.

`secondary` support the following:

* `route` - (Required) The name of the secondary region.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The ARN of the endpoint that was created.
* `endpoint_url` - The URL of the endpoint that was created.

## Import

EventBridge Global Endpoints can be imported using the `name`, e.g.,

```shell
$ terraform import aws_cloudwatch_event_endpoint.imported_endpoint example-endpoint
```
