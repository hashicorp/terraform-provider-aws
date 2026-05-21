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

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name to use for the route. Must be between 1 and 255 characters in length.
* `mesh_name` - (Required) Name of the service mesh in which to create the route. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `virtual_router_name` - (Required) Name of the virtual router in which to create the route. Must be between 1 and 255 characters in length.
* `spec` - (Required) Route specification to apply.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `spec` Block

* `grpc_route` - (Optional) GRPC routing information for the route. See [`grpc_route` Block](#grpc_route-block) for details.
* `http2_route` - (Optional) HTTP/2 routing information for the route. See [`http2_route` Block](#http2_route-block) for details.
* `http_route` - (Optional) HTTP routing information for the route. See [`http_route` Block](#http_route-block) for details.
* `priority` - (Optional) Priority for the route, between `0` and `1000`. Routes are matched based on the specified value, where `0` is the highest priority.
* `tcp_route` - (Optional) TCP routing information for the route. See [`tcp_route` Block](#tcp_route-block) for details.

### `grpc_route` Block

* `action` - (Required) Action to take if a match is determined. See [`action` Block](#action-block) for details.
* `match` - (Required) Criteria for determining an gRPC request match. See [`match` Block](#match-block) for details.
* `retry_policy` - (Optional) Retry policy. See [`retry_policy` Block](#retry_policy-block) for details.
* `timeout` - (Optional) Types of timeouts. See [`timeout` Block](#timeout-block) for details.

### `http2_route` and `http_route` Blocks

* `action` - (Required) Action to take if a match is determined.
* `match` - (Required) Criteria for determining an HTTP request match.
* `retry_policy` - (Optional) Retry policy.
* `timeout` - (Optional) Types of timeouts.

### `tcp_route` Block

* `action` - (Required) Action to take if a match is determined.
* `timeout` - (Optional) Types of timeouts.

### `action` Block

* `weighted_target` - (Required) Targets that traffic is routed to when a request matches the route. You can specify one or more targets and their relative weights with which to distribute traffic.

### `timeout` Block

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle. See [`idle` Block](#idle-block) for details.
* `per_request` - (Optional) Per request timeout. See [`per_request` Block](#per_request-block) for details.

#### `idle` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

#### `per_request` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `match` Block

* `metadata` - (Optional) Data to match from the gRPC request.
* `method_name` - (Optional) Method name to match from the request. If you specify a name, you must also specify a `service_name`.
* `service_name` - (Optional) Fully qualified domain name for the service to match from the request.
* `port` - (Optional) The port number to match from the request.

### `metadata` Block

* `name` - (Required) Name of the route. Must be between 1 and 50 characters in length.
* `invert` - (Optional) If `true`, the match is on the opposite of the `match` criteria. Default is `false`.
* `match` - (Optional) Data to match from the request.

### `match` Block

* `exact` - (Optional) Value sent by the client must match the specified value exactly. Must be between 1 and 255 characters in length.
* `prefix` - (Optional) Value sent by the client must begin with the specified characters. Must be between 1 and 255 characters in length.
* `port` - (Optional) The port number to match from the request.
* `range`- (Optional) Object that specifies the range of numbers that the value sent by the client must be included in.
* `regex` - (Optional) Value sent by the client must include the specified characters. Must be between 1 and 255 characters in length.
* `suffix` - (Optional) Value sent by the client must end with the specified characters. Must be between 1 and 255 characters in length.

### `retry_policy` Block

* `grpc_retry_events` - (Optional) List of gRPC retry events. Valid values: `cancelled`, `deadline-exceeded`, `internal`, `resource-exhausted`, `unavailable`.
* `http_retry_events` - (Optional) List of HTTP retry events. Valid values: `client-error` (HTTP status code 409), `gateway-error` (HTTP status codes 502, 503, and 504), `server-error` (HTTP status codes 500, 501, 502, 503, 504, 505, 506, 507, 508, 510, and 511), `stream-error` (retry on refused stream).
* `max_retries` - (Required) Maximum number of retries.
* `per_retry_timeout` - (Required) Per-retry timeout.
* `tcp_retry_events` - (Optional) List of TCP retry events. The only valid value is `connection-error`.

### `timeout` Block

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle. See [`idle` Block](#idle-block) for details.
* `per_request` - (Optional) Per request timeout. See [`per_request` Block](#per_request-block) for details.

#### `idle` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

#### `per_request` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `match` Block

* `header` - (Optional) Client request headers to match on. See [`header` Block](#header-block) for details.
* `method` - (Optional) Client request header method to match on. Valid values: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`.
* `path` - (Optional) Client request path to match on. See [`path` Block](#path-block) for details.
* `port` - (Optional) The port number to match from the request.
* `prefix` - (Optional) Path with which to match requests. This parameter must always start with /, which by itself matches all requests to the virtual router service name.
* `query_parameter` - (Optional) Client request query parameters to match on. See [`query_parameter` Block](#query_parameter-block) for details.
* `scheme` - (Optional) Client request header scheme to match on. Valid values: `http`, `https`.

### `path` Block

* `exact` - (Optional) The exact path to match on.
* `regex` - (Optional) The regex used to match the path.

### `query_parameter` Block

* `name` - (Required) Name for the query parameter that will be matched on.
* `match` - (Optional) The query parameter to match on.

### `match` Block

* `exact` - (Optional) The exact query parameter to match on.

### `retry_policy` Block

* `http_retry_events` - (Optional) List of HTTP retry events. Valid values: `client-error` (HTTP status code 409), `gateway-error` (HTTP status codes 502, 503, and 504), `server-error` (HTTP status codes 500, 501, 502, 503, 504, 505, 506, 507, 508, 510, and 511), `stream-error` (retry on refused stream).
* `max_retries` - (Required) Maximum number of retries.
* `per_retry_timeout` - (Required) Per-retry timeout.
* `tcp_retry_events` - (Optional) List of TCP retry events. The only valid value is `connection-error`. You must specify at least one value for `http_retry_events`, or at least one value for `tcp_retry_events`.

### `timeout` Block

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle. See [`idle` Block](#idle-block) for details.
* `per_request` - (Optional) Per request timeout. See [`per_request` Block](#per_request-block) for details.

#### `idle` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

#### `per_request` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `per_retry_timeout` Block

* `unit` - (Required) Retry unit. Valid values: `ms`, `s`.
* `value` - (Required) Retry value.

### `weighted_target` Block

* `virtual_node` - (Required) Virtual node to associate with the weighted target. Must be between 1 and 255 characters in length.
* `weight` - (Required) Relative weight of the weighted target. An integer between 0 and 100.
* `port` - (Optional) The targeted port of the weighted object.

### `header` Block

* `name` - (Required) Name for the HTTP header in the client request that will be matched on.
* `invert` - (Optional) If `true`, the match is on the opposite of the `match` method and value. Default is `false`.
* `match` - (Optional) Method and value to match the header value sent with a request. Specify one match method.

### `match` Block

* `exact` - (Optional) Header value sent by the client must match the specified value exactly.
* `prefix` - (Optional) Header value sent by the client must begin with the specified characters.
* `port` - (Optional) The port number to match from the request.
* `range`- (Optional) Object that specifies the range of numbers that the header value sent by the client must be included in.
* `regex` - (Optional) Header value sent by the client must include the specified characters.
* `suffix` - (Optional) Header value sent by the client must end with the specified characters.

### `range` Block

* `end` - (Required) End of the range.
* `start` - (Required) Start of the range.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the route.
* `arn` - ARN of the route.
* `created_date` - Creation date of the route.
* `last_updated_date` - Last update date of the route.
* `resource_owner` - Resource owner's AWS account ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Mesh virtual routes using `mesh_name` and `virtual_router_name` together with the route's `name`. For example:

```terraform
import {
  to = aws_appmesh_route.serviceb
  id = "simpleapp/serviceB/serviceB-route"
}
```

Using `terraform import`, import App Mesh virtual routes using `mesh_name` and `virtual_router_name` together with the route's `name`. For example:

```console
% terraform import aws_appmesh_route.serviceb simpleapp/serviceB/serviceB-route
```

[1]: /docs/providers/aws/index.html
