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

```terraform
resource "aws_cloudwatch_event_endpoint" "this" {
  name     = "global-endpoint"
  role_arn = aws_iam_role.replication.arn

  event_bus {
    event_bus_arn = aws_cloudwatch_event_bus.primary.arn
  }
  event_bus {
    event_bus_arn = aws_cloudwatch_event_bus.secondary.arn
  }

  replication_config {
    state = "DISABLED"
  }

  routing_config {
    failover_config {
      primary {
        health_check = aws_route53_health_check.primary.arn
      }

      secondary {
        route = "us-east-2"
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) A description of the global endpoint.
* `event_bus` - (Required) The event buses to use. The names of the event buses must be identical in each Region. Exactly two event buses are required. Documented below.
* `name` - (Required) The name of the global endpoint.
* `replication_config` - (Optional) Parameters used for replication. Documented below.
* `role_arn` - (Optional) The ARN of the IAM role used for replication between event buses.
* `routing_config` - (Required) Parameters used for routing, including the health check and secondary Region. Documented below.

`event_bus` supports the following:

* `event_bus_arn` - (Required) The ARN of the event bus the endpoint is associated with.

`replication_config` supports the following:

* `state` - (Optional) The state of event replication. Valid values: `ENABLED`, `DISABLED`. The default state is `ENABLED`, which means you must supply a `role_arn`. If you don't have a `role_arn` or you don't want event replication enabled, set `state` to `DISABLED`.

`routing_config` support the following:

* `failover_config` - (Required) Parameters used for failover. This includes what triggers failover and what happens when it's triggered. Documented below.

`failover_config` support the following:

* `primary` - (Required) Parameters used for the primary Region. Documented below.
* `secondary` - (Required) Parameters used for the secondary Region, the Region that events are routed to when failover is triggered or event replication is enabled. Documented below.

`primary` support the following:

* `health_check` - (Required) The ARN of the health check used by the endpoint to determine whether failover is triggered.

`secondary` support the following:

* `route` - (Required) The name of the secondary Region.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the endpoint that was created.
* `endpoint_url` - The URL of the endpoint that was created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge Global Endpoints using the `name`. For example:

```terraform
import {
  to = aws_cloudwatch_event_endpoint.imported_endpoint
  id = "example-endpoint"
}
```

Using `terraform import`, import EventBridge Global Endpoints using the `name`. For example:

```console
% terraform import aws_cloudwatch_event_endpoint.imported_endpoint example-endpoint
```
