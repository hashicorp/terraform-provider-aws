---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_route"
description: |-
  Provides an AWS App Mesh route resource.
---

# Resource: aws_appmesh_route

Provides an AWS App Mesh route resource.

## Example Usage

### HTTP Routing

```terraform
resource "aws_appmesh_route" "serviceb" {
  name                = "serviceB-route"
  mesh_name           = aws_appmesh_mesh.simple.id
  virtual_router_name = aws_appmesh_virtual_router.serviceb.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.serviceb1.name
          weight       = 90
        }

        weighted_target {
          virtual_node = aws_appmesh_virtual_node.serviceb2.name
          weight       = 10
        }
      }
    }
  }
}
```

### HTTP Header Routing

```terraform
resource "aws_appmesh_route" "serviceb" {
  name                = "serviceB-route"
  mesh_name           = aws_appmesh_mesh.simple.id
  virtual_router_name = aws_appmesh_virtual_router.serviceb.name

  spec {
    http_route {
      match {
        method = "POST"
        prefix = "/"
        scheme = "https"

        header {
          name = "clientRequestId"

          match {
            prefix = "123"
          }
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.serviceb.name
          weight       = 100
        }
      }
    }
  }
}
```

### Retry Policy

```terraform
resource "aws_appmesh_route" "serviceb" {
  name                = "serviceB-route"
  mesh_name           = aws_appmesh_mesh.simple.id
  virtual_router_name = aws_appmesh_virtual_router.serviceb.name

  spec {
    http_route {
      match {
        prefix = "/"
      }

      retry_policy {
        http_retry_events = [
          "server-error",
        ]
        max_retries = 1

        per_retry_timeout {
          unit  = "s"
          value = 15
        }
      }

      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.serviceb.name
          weight       = 100
        }
      }
    }
  }
}
```

### TCP Routing

```terraform
resource "aws_appmesh_route" "serviceb" {
  name                = "serviceB-route"
  mesh_name           = aws_appmesh_mesh.simple.id
  virtual_router_name = aws_appmesh_virtual_router.serviceb.name

  spec {
    tcp_route {
      action {
        weighted_target {
          virtual_node = aws_appmesh_virtual_node.serviceb1.name
          weight       = 100
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name to use for the route. Must be between 1 and 255 characters in length.
* `mesh_name` - (Required) Name of the service mesh in which to create the route. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `virtual_router_name` - (Required) Name of the virtual router in which to create the route. Must be between 1 and 255 characters in length.
* `spec` - (Required) Route specification to apply.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `spec` object supports the following:

* `grpc_route` - (Optional) GRPC routing information for the route.
* `http2_route` - (Optional) HTTP/2 routing information for the route.
* `http_route` - (Optional) HTTP routing information for the route.
* `priority` - (Optional) Priority for the route, between `0` and `1000`.
Routes are matched based on the specified value, where `0` is the highest priority.
* `tcp_route` - (Optional) TCP routing information for the route.

The `grpc_route` object supports the following:

* `action` - (Required) Action to take if a match is determined.
* `match` - (Required) Criteria for determining an gRPC request match.
* `retry_policy` - (Optional) Retry policy.
* `timeout` - (Optional) Types of timeouts.

The `http2_route` and `http_route` objects supports the following:

* `action` - (Required) Action to take if a match is determined.
* `match` - (Required) Criteria for determining an HTTP request match.
* `retry_policy` - (Optional) Retry policy.
* `timeout` - (Optional) Types of timeouts.

The `tcp_route` object supports the following:

* `action` - (Required) Action to take if a match is determined.
* `timeout` - (Optional) Types of timeouts.

The `action` object supports the following:

* `weighted_target` - (Required) Targets that traffic is routed to when a request matches the route.
You can specify one or more targets and their relative weights with which to distribute traffic.

The `timeout` object supports the following:

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle.

The `idle` object supports the following:

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

The `grpc_route`'s `match` object supports the following:

* `metadata` - (Optional) Data to match from the gRPC request.
* `method_name` - (Optional) Method name to match from the request. If you specify a name, you must also specify a `service_name`.
* `service_name` - (Optional) Fully qualified domain name for the service to match from the request.
* `port`- (Optional) The port number to match from the request.

The `metadata` object supports the following:

* `name` - (Required) Name of the route. Must be between 1 and 50 characters in length.
* `invert` - (Optional) If `true`, the match is on the opposite of the `match` criteria. Default is `false`.
* `match` - (Optional) Data to match from the request.

The `metadata`'s `match` object supports the following:

* `exact` - (Optional) Value sent by the client must match the specified value exactly. Must be between 1 and 255 characters in length.
* `prefix` - (Optional) Value sent by the client must begin with the specified characters. Must be between 1 and 255 characters in length.
* `port`- (Optional) The port number to match from the request.
* `range`- (Optional) Object that specifies the range of numbers that the value sent by the client must be included in.
* `regex` - (Optional) Value sent by the client must include the specified characters. Must be between 1 and 255 characters in length.
* `suffix` - (Optional) Value sent by the client must end with the specified characters. Must be between 1 and 255 characters in length.

The `grpc_route`'s `retry_policy` object supports the following:

* `grpc_retry_events` - (Optional) List of gRPC retry events.
Valid values: `cancelled`, `deadline-exceeded`, `internal`, `resource-exhausted`, `unavailable`.
* `http_retry_events` - (Optional) List of HTTP retry events.
Valid values: `client-error` (HTTP status code 409), `gateway-error` (HTTP status codes 502, 503, and 504), `server-error` (HTTP status codes 500, 501, 502, 503, 504, 505, 506, 507, 508, 510, and 511), `stream-error` (retry on refused stream).
* `max_retries` - (Required) Maximum number of retries.
* `per_retry_timeout` - (Required) Per-retry timeout.
* `tcp_retry_events` - (Optional) List of TCP retry events. The only valid value is `connection-error`.

The `grpc_route`'s `timeout` object supports the following:

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle.
* `per_request` - (Optional) Per request timeout.

The `idle` and `per_request` objects support the following:

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

The `http2_route` and `http_route`'s `match` object supports the following:

* `prefix` - (Required) Path with which to match requests.
This parameter must always start with /, which by itself matches all requests to the virtual router service name.
* `port`- (Optional) The port number to match from the request.
* `header` - (Optional) Client request headers to match on.
* `method` - (Optional) Client request header method to match on. Valid values: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`.
* `scheme` - (Optional) Client request header scheme to match on. Valid values: `http`, `https`.

The `http2_route` and `http_route`'s `retry_policy` object supports the following:

* `http_retry_events` - (Optional) List of HTTP retry events.
Valid values: `client-error` (HTTP status code 409), `gateway-error` (HTTP status codes 502, 503, and 504), `server-error` (HTTP status codes 500, 501, 502, 503, 504, 505, 506, 507, 508, 510, and 511), `stream-error` (retry on refused stream).
* `max_retries` - (Required) Maximum number of retries.
* `per_retry_timeout` - (Required) Per-retry timeout.
* `tcp_retry_events` - (Optional) List of TCP retry events. The only valid value is `connection-error`.

You must specify at least one value for `http_retry_events`, or at least one value for `tcp_retry_events`.

The `http2_route` and `http_route`'s `timeout` object supports the following:

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle.
* `per_request` - (Optional) Per request timeout.

The `idle` and `per_request` objects support the following:

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

The `per_retry_timeout` object supports the following:

* `unit` - (Required) Retry unit. Valid values: `ms`, `s`.
* `value` - (Required) Retry value.

The `weighted_target` object supports the following:

* `virtual_node` - (Required) Virtual node to associate with the weighted target. Must be between 1 and 255 characters in length.
* `weight` - (Required) Relative weight of the weighted target. An integer between 0 and 100.
* `port` - (Optional) The targeted port of the weighted object.

The `header` object supports the following:

* `name` - (Required) Name for the HTTP header in the client request that will be matched on.
* `invert` - (Optional) If `true`, the match is on the opposite of the `match` method and value. Default is `false`.
* `match` - (Optional) Method and value to match the header value sent with a request. Specify one match method.

The `header`'s `match` object supports the following:

* `exact` - (Optional) Header value sent by the client must match the specified value exactly.
* `prefix` - (Optional) Header value sent by the client must begin with the specified characters.
* `port`- (Optional) The port number to match from the request.
* `range`- (Optional) Object that specifies the range of numbers that the header value sent by the client must be included in.
* `regex` - (Optional) Header value sent by the client must include the specified characters.
* `suffix` - (Optional) Header value sent by the client must end with the specified characters.

The `range` object supports the following:

* `end` - (Required) End of the range.
* `start` - (Requited) Start of the range.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the route.
* `arn` - ARN of the route.
* `created_date` - Creation date of the route.
* `last_updated_date` - Last update date of the route.
* `resource_owner` - Resource owner's AWS account ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

App Mesh virtual routes can be imported using `mesh_name` and `virtual_router_name` together with the route's `name`,
e.g.,

```
$ terraform import aws_appmesh_route.serviceb simpleapp/serviceB/serviceB-route
```

[1]: /docs/providers/aws/index.html
