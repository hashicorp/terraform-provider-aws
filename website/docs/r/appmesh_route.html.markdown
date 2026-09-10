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

* `mesh_name` - (Required) Name of the service mesh in which to create the route. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider](/docs/providers/aws/index.html) is currently connected to.
* `name` - (Required) Name to use for the route. Must be between 1 and 255 characters in length.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `spec` - (Required) Route specification to apply. See [`spec` Block](#spec-block) for details.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `virtual_router_name` - (Required) Name of the virtual router in which to create the route. Must be between 1 and 255 characters in length.

### `spec` Block

* `grpc_route` - (Optional) GRPC routing information for the route. See [`spec.grpc_route` Block](#specgrpc_route-block) for details.
* `http2_route` - (Optional) HTTP/2 routing information for the route. See [`spec.http2_route` Block](#spechttp2_route-block) for details.
* `http_route` - (Optional) HTTP routing information for the route. See [`spec.http_route` Block](#spechttp_route-block) for details.
* `priority` - (Optional) Priority for the route, between `0` and `1000`. Routes are matched based on the specified value, where `0` is the highest priority.
* `tcp_route` - (Optional) TCP routing information for the route. See [`spec.tcp_route` Block](#spectcp_route-block) for details.

### `spec.grpc_route` Block

* `action` - (Required) Action to take if a match is determined. See [`spec.grpc_route.action` Block](#specgrpc_routeaction-block) for details.
* `match` - (Optional) Criteria for determining an gRPC request match. See [`spec.grpc_route.match` Block](#specgrpc_routematch-block) for details.
* `retry_policy` - (Optional) Retry policy. See [`spec.grpc_route.retry_policy` Block](#specgrpc_routeretry_policy-block) for details.
* `timeout` - (Optional) Types of timeouts. See [`spec.grpc_route.timeout` Block](#specgrpc_routetimeout-block) for details.

### `spec.grpc_route.action` Block

* `weighted_target` - (Required) Targets that traffic is routed to when a request matches the route. You can specify one or more targets and their relative weights with which to distribute traffic. See [`spec.grpc_route.action.weighted_target` Block](#specgrpc_routeactionweighted_target-block) for details.

### `spec.grpc_route.action.weighted_target` Block

* `port` - (Optional) Targeted port of the weighted object.
* `virtual_node` - (Required) Virtual node to associate with the weighted target. Must be between 1 and 255 characters in length.
* `weight` - (Required) Relative weight of the weighted target. An integer between 0 and 100.

### `spec.grpc_route.match` Block

* `metadata` - (Optional) Data to match from the gRPC request. See [`spec.grpc_route.match.metadata` Block](#specgrpc_routematchmetadata-block) for details.
* `method_name` - (Optional) Method name to match from the request. If you specify a name, you must also specify a `service_name`.
* `port` - (Optional) Port number to match from the request.
* `prefix` - (Optional) Value sent by the client must begin with the specified characters.
* `service_name` - (Optional) Fully qualified domain name for the service to match from the request.

### `spec.grpc_route.match.metadata` Block

* `invert` - (Optional) Whether to match on the opposite of the `match` criteria. Default is `false`.
* `match` - (Optional) Data to match from the request. See [`spec.grpc_route.match.metadata.match` Block](#specgrpc_routematchmetadatamatch-block) for details.
* `name` - (Required) Name of the route. Must be between 1 and 50 characters in length.

### `spec.grpc_route.match.metadata.match` Block

* `exact` - (Optional) Value sent by the client must match the specified value exactly. Must be between 1 and 255 characters in length.
* `prefix` - (Optional) Value sent by the client must begin with the specified characters. Must be between 1 and 255 characters in length.
* `range` - (Optional) Object that specifies the range of numbers that the value sent by the client must be included in. See [`spec.grpc_route.match.metadata.match.range` Block](#specgrpc_routematchmetadatamatchrange-block) for details.
* `regex` - (Optional) Value sent by the client must include the specified characters. Must be between 1 and 255 characters in length.
* `suffix` - (Optional) Value sent by the client must end with the specified characters. Must be between 1 and 255 characters in length.

### `spec.grpc_route.match.metadata.match.range` Block

* `end` - (Required) End of the range.
* `start` - (Required) Start of the range.

### `spec.grpc_route.retry_policy` Block

* `grpc_retry_events` - (Optional) List of gRPC retry events. Valid values: `cancelled`, `deadline-exceeded`, `internal`, `resource-exhausted`, `unavailable`.
* `http_retry_events` - (Optional) List of HTTP retry events. Valid values: `client-error` (HTTP status code 409), `gateway-error` (HTTP status codes 502, 503, and 504), `server-error` (HTTP status codes 500, 501, 502, 503, 504, 505, 506, 507, 508, 510, and 511), `stream-error` (retry on refused stream).
* `max_retries` - (Required) Maximum number of retries.
* `per_retry_timeout` - (Required) Per-retry timeout. See [`spec.grpc_route.retry_policy.per_retry_timeout` Block](#specgrpc_routeretry_policyper_retry_timeout-block) for details.
* `tcp_retry_events` - (Optional) List of TCP retry events. The only valid value is `connection-error`.

### `spec.grpc_route.retry_policy.per_retry_timeout` Block

* `unit` - (Required) Retry unit. Valid values: `ms`, `s`.
* `value` - (Required) Retry value.

### `spec.grpc_route.timeout` Block

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle. See [`spec.grpc_route.timeout.idle` Block](#specgrpc_routetimeoutidle-block) for details.
* `per_request` - (Optional) Per request timeout. See [`spec.grpc_route.timeout.per_request` Block](#specgrpc_routetimeoutper_request-block) for details.

### `spec.grpc_route.timeout.idle` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `spec.grpc_route.timeout.per_request` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `spec.http2_route` Block

* `action` - (Required) Action to take if a match is determined. See [`spec.http2_route.action` Block](#spechttp2_routeaction-block) for details.
* `match` - (Required) Criteria for determining an HTTP request match. See [`spec.http2_route.match` Block](#spechttp2_routematch-block) for details.
* `retry_policy` - (Optional) Retry policy. See [`spec.http2_route.retry_policy` Block](#spechttp2_routeretry_policy-block) for details.
* `timeout` - (Optional) Types of timeouts. See [`spec.http2_route.timeout` Block](#spechttp2_routetimeout-block) for details.

### `spec.http2_route.action` Block

* `weighted_target` - (Required) Targets that traffic is routed to when a request matches the route. You can specify one or more targets and their relative weights with which to distribute traffic. See [`spec.http2_route.action.weighted_target` Block](#spechttp2_routeactionweighted_target-block) for details.

### `spec.http2_route.action.weighted_target` Block

* `port` - (Optional) Targeted port of the weighted object.
* `virtual_node` - (Required) Virtual node to associate with the weighted target. Must be between 1 and 255 characters in length.
* `weight` - (Required) Relative weight of the weighted target. An integer between 0 and 100.

### `spec.http2_route.match` Block

* `header` - (Optional) Client request headers to match on. See [`spec.http2_route.match.header` Block](#spechttp2_routematchheader-block) for details.
* `method` - (Optional) Client request header method to match on. Valid values: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`.
* `path` - (Optional) Client request path to match on. See [`spec.http2_route.match.path` Block](#spechttp2_routematchpath-block) for details.
* `port` - (Optional) Port number to match from the request.
* `prefix` - (Optional) Path with which to match requests. This parameter must always start with /, which by itself matches all requests to the virtual router service name.
* `query_parameter` - (Optional) Client request query parameters to match on. See [`spec.http2_route.match.query_parameter` Block](#spechttp2_routematchquery_parameter-block) for details.
* `scheme` - (Optional) Client request header scheme to match on. Valid values: `http`, `https`.

### `spec.http2_route.match.header` Block

* `invert` - (Optional) Whether to match on the opposite of the `match` method and value. Default is `false`.
* `match` - (Optional) Method and value to match the header value sent with a request. Specify one match method. See [`spec.http2_route.match.header.match` Block](#spechttp2_routematchheadermatch-block) for details.
* `name` - (Required) Name for the HTTP header in the client request that will be matched on.

### `spec.http2_route.match.header.match` Block

* `exact` - (Optional) Header value sent by the client must match the specified value exactly.
* `prefix` - (Optional) Header value sent by the client must begin with the specified characters.
* `range` - (Optional) Object that specifies the range of numbers that the header value sent by the client must be included in. See [`spec.http2_route.match.header.match.range` Block](#spechttp2_routematchheadermatchrange-block) for details.
* `regex` - (Optional) Header value sent by the client must include the specified characters.
* `suffix` - (Optional) Header value sent by the client must end with the specified characters.

### `spec.http2_route.match.header.match.range` Block

* `end` - (Required) End of the range.
* `start` - (Required) Start of the range.

### `spec.http2_route.match.path` Block

* `exact` - (Optional) Exact path to match on.
* `regex` - (Optional) Regex used to match the path.

### `spec.http2_route.match.query_parameter` Block

* `match` - (Optional) Query parameter to match on. See [`spec.http2_route.match.query_parameter.match` Block](#spechttp2_routematchquery_parametermatch-block) for details.
* `name` - (Required) Name for the query parameter that will be matched on.

### `spec.http2_route.match.query_parameter.match` Block

* `exact` - (Optional) Exact query parameter to match on.

### `spec.http2_route.retry_policy` Block

* `http_retry_events` - (Optional) List of HTTP retry events. Valid values: `client-error` (HTTP status code 409), `gateway-error` (HTTP status codes 502, 503, and 504), `server-error` (HTTP status codes 500, 501, 502, 503, 504, 505, 506, 507, 508, 510, and 511), `stream-error` (retry on refused stream).
* `max_retries` - (Required) Maximum number of retries.
* `per_retry_timeout` - (Required) Per-retry timeout. See [`spec.http2_route.retry_policy.per_retry_timeout` Block](#spechttp2_routeretry_policyper_retry_timeout-block) for details.
* `tcp_retry_events` - (Optional) List of TCP retry events. The only valid value is `connection-error`. You must specify at least one value for `http_retry_events`, or at least one value for `tcp_retry_events`.

### `spec.http2_route.retry_policy.per_retry_timeout` Block

* `unit` - (Required) Retry unit. Valid values: `ms`, `s`.
* `value` - (Required) Retry value.

### `spec.http2_route.timeout` Block

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle. See [`spec.http2_route.timeout.idle` Block](#spechttp2_routetimeoutidle-block) for details.
* `per_request` - (Optional) Per request timeout. See [`spec.http2_route.timeout.per_request` Block](#spechttp2_routetimeoutper_request-block) for details.

### `spec.http2_route.timeout.idle` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `spec.http2_route.timeout.per_request` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `spec.http_route` Block

* `action` - (Required) Action to take if a match is determined. See [`spec.http_route.action` Block](#spechttp_routeaction-block) for details.
* `match` - (Required) Criteria for determining an HTTP request match. See [`spec.http_route.match` Block](#spechttp_routematch-block) for details.
* `retry_policy` - (Optional) Retry policy. See [`spec.http_route.retry_policy` Block](#spechttp_routeretry_policy-block) for details.
* `timeout` - (Optional) Types of timeouts. See [`spec.http_route.timeout` Block](#spechttp_routetimeout-block) for details.

### `spec.http_route.action` Block

* `weighted_target` - (Required) Targets that traffic is routed to when a request matches the route. You can specify one or more targets and their relative weights with which to distribute traffic. See [`spec.http_route.action.weighted_target` Block](#spechttp_routeactionweighted_target-block) for details.

### `spec.http_route.action.weighted_target` Block

* `port` - (Optional) Targeted port of the weighted object.
* `virtual_node` - (Required) Virtual node to associate with the weighted target. Must be between 1 and 255 characters in length.
* `weight` - (Required) Relative weight of the weighted target. An integer between 0 and 100.

### `spec.http_route.match` Block

* `header` - (Optional) Client request headers to match on. See [`spec.http_route.match.header` Block](#spechttp_routematchheader-block) for details.
* `method` - (Optional) Client request header method to match on. Valid values: `GET`, `HEAD`, `POST`, `PUT`, `DELETE`, `CONNECT`, `OPTIONS`, `TRACE`, `PATCH`.
* `path` - (Optional) Client request path to match on. See [`spec.http_route.match.path` Block](#spechttp_routematchpath-block) for details.
* `port` - (Optional) Port number to match from the request.
* `prefix` - (Optional) Path with which to match requests. This parameter must always start with /, which by itself matches all requests to the virtual router service name.
* `query_parameter` - (Optional) Client request query parameters to match on. See [`spec.http_route.match.query_parameter` Block](#spechttp_routematchquery_parameter-block) for details.
* `scheme` - (Optional) Client request header scheme to match on. Valid values: `http`, `https`.

### `spec.http_route.match.header` Block

* `invert` - (Optional) Whether to match on the opposite of the `match` method and value. Default is `false`.
* `match` - (Optional) Method and value to match the header value sent with a request. Specify one match method. See [`spec.http_route.match.header.match` Block](#spechttp_routematchheadermatch-block) for details.
* `name` - (Required) Name for the HTTP header in the client request that will be matched on.

### `spec.http_route.match.header.match` Block

* `exact` - (Optional) Header value sent by the client must match the specified value exactly.
* `prefix` - (Optional) Header value sent by the client must begin with the specified characters.
* `range` - (Optional) Object that specifies the range of numbers that the header value sent by the client must be included in. See [`spec.http_route.match.header.match.range` Block](#spechttp_routematchheadermatchrange-block) for details.
* `regex` - (Optional) Header value sent by the client must include the specified characters.
* `suffix` - (Optional) Header value sent by the client must end with the specified characters.

### `spec.http_route.match.header.match.range` Block

* `end` - (Required) End of the range.
* `start` - (Required) Start of the range.

### `spec.http_route.match.path` Block

* `exact` - (Optional) Exact path to match on.
* `regex` - (Optional) Regex used to match the path.

### `spec.http_route.match.query_parameter` Block

* `match` - (Optional) Query parameter to match on. See [`spec.http_route.match.query_parameter.match` Block](#spechttp_routematchquery_parametermatch-block) for details.
* `name` - (Required) Name for the query parameter that will be matched on.

### `spec.http_route.match.query_parameter.match` Block

* `exact` - (Optional) Exact query parameter to match on.

### `spec.http_route.retry_policy` Block

* `http_retry_events` - (Optional) List of HTTP retry events. Valid values: `client-error` (HTTP status code 409), `gateway-error` (HTTP status codes 502, 503, and 504), `server-error` (HTTP status codes 500, 501, 502, 503, 504, 505, 506, 507, 508, 510, and 511), `stream-error` (retry on refused stream).
* `max_retries` - (Required) Maximum number of retries.
* `per_retry_timeout` - (Required) Per-retry timeout. See [`spec.http_route.retry_policy.per_retry_timeout` Block](#spechttp_routeretry_policyper_retry_timeout-block) for details.
* `tcp_retry_events` - (Optional) List of TCP retry events. The only valid value is `connection-error`. You must specify at least one value for `http_retry_events`, or at least one value for `tcp_retry_events`.

### `spec.http_route.retry_policy.per_retry_timeout` Block

* `unit` - (Required) Retry unit. Valid values: `ms`, `s`.
* `value` - (Required) Retry value.

### `spec.http_route.timeout` Block

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle. See [`spec.http_route.timeout.idle` Block](#spechttp_routetimeoutidle-block) for details.
* `per_request` - (Optional) Per request timeout. See [`spec.http_route.timeout.per_request` Block](#spechttp_routetimeoutper_request-block) for details.

### `spec.http_route.timeout.idle` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `spec.http_route.timeout.per_request` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

### `spec.tcp_route` Block

* `action` - (Required) Action to take if a match is determined. See [`spec.tcp_route.action` Block](#spectcp_routeaction-block) for details.
* `match` - (Optional) Criteria for determining a TCP request match. See [`spec.tcp_route.match` Block](#spectcp_routematch-block) for details.
* `timeout` - (Optional) Types of timeouts. See [`spec.tcp_route.timeout` Block](#spectcp_routetimeout-block) for details.

### `spec.tcp_route.action` Block

* `weighted_target` - (Required) Targets that traffic is routed to when a request matches the route. You can specify one or more targets and their relative weights with which to distribute traffic. See [`spec.tcp_route.action.weighted_target` Block](#spectcp_routeactionweighted_target-block) for details.

### `spec.tcp_route.action.weighted_target` Block

* `port` - (Optional) Targeted port of the weighted object.
* `virtual_node` - (Required) Virtual node to associate with the weighted target. Must be between 1 and 255 characters in length.
* `weight` - (Required) Relative weight of the weighted target. An integer between 0 and 100.

### `spec.tcp_route.match` Block

* `port` - (Optional) Port number to match from the request.

### `spec.tcp_route.timeout` Block

* `idle` - (Optional) Idle timeout. An idle timeout bounds the amount of time that a connection may be idle. See [`spec.tcp_route.timeout.idle` Block](#spectcp_routetimeoutidle-block) for details.

### `spec.tcp_route.timeout.idle` Block

* `unit` - (Required) Unit of time. Valid values: `ms`, `s`.
* `value` - (Required) Number of time units. Minimum value of `0`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the route.
* `created_date` - Creation date of the route.
* `id` - ID of the route.
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
