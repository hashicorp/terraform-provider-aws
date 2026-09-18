---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_gateway_route"
description: |-
  Provides an AWS App Mesh gateway route resource.
---

# Resource: aws_appmesh_gateway_route

Provides an AWS App Mesh gateway route resource.

## Example Usage

```terraform
resource "aws_appmesh_gateway_route" "example" {
  name                 = "example-gateway-route"
  mesh_name            = "example-service-mesh"
  virtual_gateway_name = aws_appmesh_virtual_gateway.example.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = aws_appmesh_virtual_service.example.name
          }
        }
      }

      match {
        prefix = "/"
      }
    }
  }

  tags = {
    Environment = "test"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `mesh_name` - (Required) Name of the service mesh in which to create the gateway route. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider](/docs/providers/aws/index.html) is currently connected to.
* `name` - (Required) Name to use for the gateway route. Must be between 1 and 255 characters in length.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `spec` - (Required) Gateway route specification to apply.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `virtual_gateway_name` - (Required) Name of the [virtual gateway](/docs/providers/aws/r/appmesh_virtual_gateway.html) to associate the gateway route with. Must be between 1 and 255 characters in length.

### `spec` Block

* `grpc_route` - (Optional) Specification of a gRPC gateway route. See [`grpc_route` Block](#grpc_route-block) for details.
* `http2_route` - (Optional) Specification of an HTTP/2 gateway route. See [`http2_route` Block](#http2_route-block) for details.
* `http_route` - (Optional) Specification of an HTTP gateway route. See [`http_route` Block](#http_route-block) for details.
* `priority` - (Optional) Priority for the gateway route, between `0` and `1000`.

### `grpc_route` Block

* `action` - (Required) Action to take if a match is determined. See [`spec.grpc_route.action` Block](#specgrpc_routeaction-block) for details.
* `match` - (Required) Criteria for determining a request match. See [`spec.grpc_route.match` Block](#specgrpc_routematch-block) for details.

### `spec.grpc_route.action` Block

* `target` - (Required) Target that traffic is routed to when a request matches the gateway route. See [`spec.grpc_route.action.target` Block](#specgrpc_routeactiontarget-block) for details.

### `spec.grpc_route.action.target` Block

* `port` - (Optional) Port number that corresponds to the target for Virtual Service provider port. This is required when the provider (router or node) of the Virtual Service has multiple listeners.
* `virtual_service` - (Required) Virtual service gateway route target. See [`spec.grpc_route.action.target.virtual_service` Block](#specgrpc_routeactiontargetvirtual_service-block) for details.

### `spec.grpc_route.action.target.virtual_service` Block

* `virtual_service_name` - (Required) Name of the virtual service that traffic is routed to. Must be between 1 and 255 characters in length.

### `spec.grpc_route.match` Block

* `port` - (Optional) Port number to match from the request.
* `service_name` - (Required) Fully qualified domain name for the service to match from the request.

### `http_route` Block

* `action` - (Required) Action to take if a match is determined. See [`spec.http_route.action` Block](#spechttp_routeaction-block) for details.
* `match` - (Required) Criteria for determining a request match. See [`spec.http_route.match` Block](#spechttp_routematch-block) for details.

### `spec.http_route.action` Block

* `rewrite` - (Optional) Gateway route action to rewrite. See [`spec.http_route.action.rewrite` Block](#spechttp_routeactionrewrite-block) for details.
* `target` - (Required) Target that traffic is routed to when a request matches the gateway route. See [`spec.http_route.action.target` Block](#spechttp_routeactiontarget-block) for details.

### `spec.http_route.action.rewrite` Block

* `hostname` - (Optional) Host name to rewrite. See [`spec.http_route.action.rewrite.hostname` Block](#spechttp_routeactionrewritehostname-block) for details.
* `path` - (Optional) Exact path to rewrite. See [`spec.http_route.action.rewrite.path` Block](#spechttp_routeactionrewritepath-block) for details.
* `prefix` - (Optional) Specified beginning characters to rewrite. See [`spec.http_route.action.rewrite.prefix` Block](#spechttp_routeactionrewriteprefix-block) for details.

### `spec.http_route.action.rewrite.hostname` Block

* `default_target_hostname` - (Required) Default target host name to write to. Valid values: `ENABLED`, `DISABLED`.

### `spec.http_route.action.rewrite.path` Block

* `exact` - (Required) Value used to replace matched path.

### `spec.http_route.action.rewrite.prefix` Block

* `default_prefix` - (Optional) Default prefix used to replace the incoming route prefix when rewritten. Valid values: `ENABLED`, `DISABLED`.
* `value` - (Optional) Value used to replace the incoming route prefix when rewritten.

### `spec.http_route.action.target` Block

* `port` - (Optional) Port number that corresponds to the target for Virtual Service provider port. This is required when the provider (router or node) of the Virtual Service has multiple listeners.
* `virtual_service` - (Required) Virtual service gateway route target. See [`spec.http_route.action.target.virtual_service` Block](#spechttp_routeactiontargetvirtual_service-block) for details.

### `spec.http_route.action.target.virtual_service` Block

* `virtual_service_name` - (Required) Name of the virtual service that traffic is routed to. Must be between 1 and 255 characters in length.

### `spec.http_route.match` Block

* `header` - (Optional) Client request headers to match on. See [`spec.http_route.match.header` Block](#spechttp_routematchheader-block) for details.
* `hostname` - (Optional) Host name to match on. See [`spec.http_route.match.hostname` Block](#spechttp_routematchhostname-block) for details.
* `path` - (Optional) Client request path to match on. See [`spec.http_route.match.path` Block](#spechttp_routematchpath-block) for details.
* `port` - (Optional) Port number to match from the request.
* `prefix` - (Optional) Path to match requests with. This parameter must always start with `/`, which by itself matches all requests to the virtual service name.
* `query_parameter` - (Optional) Client request query parameters to match on. See [`spec.http_route.match.query_parameter` Block](#spechttp_routematchquery_parameter-block) for details.

### `spec.http_route.match.header` Block

* `invert` - (Optional) If `true`, the match is on the opposite of the `match` method and value. Default is `false`.
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

### `spec.http_route.match.hostname` Block

* `exact` - (Optional) Exact host name to match on.
* `suffix` - (Optional) Specified ending characters of the host name to match on.

### `spec.http_route.match.path` Block

* `exact` - (Optional) Exact path to match on.
* `regex` - (Optional) Regex used to match the path.

### `spec.http_route.match.query_parameter` Block

* `match` - (Optional) Query parameter to match on. See [`spec.http_route.match.query_parameter.match` Block](#spechttp_routematchquery_parametermatch-block) for details.
* `name` - (Required) Name for the query parameter that will be matched on.

### `spec.http_route.match.query_parameter.match` Block

* `exact` - (Optional) Exact query parameter to match on.

### `http2_route` Block

* `action` - (Required) Action to take if a match is determined. See [`spec.http2_route.action` Block](#spechttp2_routeaction-block) for details.
* `match` - (Required) Criteria for determining a request match. See [`spec.http2_route.match` Block](#spechttp2_routematch-block) for details.

### `spec.http2_route.action` Block

* `rewrite` - (Optional) Gateway route action to rewrite. See [`spec.http2_route.action.rewrite` Block](#spechttp2_routeactionrewrite-block) for details.
* `target` - (Required) Target that traffic is routed to when a request matches the gateway route. See [`spec.http2_route.action.target` Block](#spechttp2_routeactiontarget-block) for details.

### `spec.http2_route.action.rewrite` Block

* `hostname` - (Optional) Host name to rewrite. See [`spec.http2_route.action.rewrite.hostname` Block](#spechttp2_routeactionrewritehostname-block) for details.
* `path` - (Optional) Exact path to rewrite. See [`spec.http2_route.action.rewrite.path` Block](#spechttp2_routeactionrewritepath-block) for details.
* `prefix` - (Optional) Specified beginning characters to rewrite. See [`spec.http2_route.action.rewrite.prefix` Block](#spechttp2_routeactionrewriteprefix-block) for details.

### `spec.http2_route.action.rewrite.hostname` Block

* `default_target_hostname` - (Required) Default target host name to write to. Valid values: `ENABLED`, `DISABLED`.

### `spec.http2_route.action.rewrite.path` Block

* `exact` - (Required) Value used to replace matched path.

### `spec.http2_route.action.rewrite.prefix` Block

* `default_prefix` - (Optional) Default prefix used to replace the incoming route prefix when rewritten. Valid values: `ENABLED`, `DISABLED`.
* `value` - (Optional) Value used to replace the incoming route prefix when rewritten.

### `spec.http2_route.action.target` Block

* `port` - (Optional) Port number that corresponds to the target for Virtual Service provider port. This is required when the provider (router or node) of the Virtual Service has multiple listeners.
* `virtual_service` - (Required) Virtual service gateway route target. See [`spec.http2_route.action.target.virtual_service` Block](#spechttp2_routeactiontargetvirtual_service-block) for details.

### `spec.http2_route.action.target.virtual_service` Block

* `virtual_service_name` - (Required) Name of the virtual service that traffic is routed to. Must be between 1 and 255 characters in length.

### `spec.http2_route.match` Block

* `header` - (Optional) Client request headers to match on. See [`spec.http2_route.match.header` Block](#spechttp2_routematchheader-block) for details.
* `hostname` - (Optional) Host name to match on. See [`spec.http2_route.match.hostname` Block](#spechttp2_routematchhostname-block) for details.
* `path` - (Optional) Client request path to match on. See [`spec.http2_route.match.path` Block](#spechttp2_routematchpath-block) for details.
* `port` - (Optional) Port number to match from the request.
* `prefix` - (Optional) Path to match requests with. This parameter must always start with `/`, which by itself matches all requests to the virtual service name.
* `query_parameter` - (Optional) Client request query parameters to match on. See [`spec.http2_route.match.query_parameter` Block](#spechttp2_routematchquery_parameter-block) for details.

### `spec.http2_route.match.header` Block

* `invert` - (Optional) If `true`, the match is on the opposite of the `match` method and value. Default is `false`.
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

### `spec.http2_route.match.hostname` Block

* `exact` - (Optional) Exact host name to match on.
* `suffix` - (Optional) Specified ending characters of the host name to match on.

### `spec.http2_route.match.path` Block

* `exact` - (Optional) Exact path to match on.
* `regex` - (Optional) Regex used to match the path.

### `spec.http2_route.match.query_parameter` Block

* `match` - (Optional) Query parameter to match on. See [`spec.http2_route.match.query_parameter.match` Block](#spechttp2_routematchquery_parametermatch-block) for details.
* `name` - (Required) Name for the query parameter that will be matched on.

### `spec.http2_route.match.query_parameter.match` Block

* `exact` - (Optional) Exact query parameter to match on.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the gateway route.
* `created_date` - Creation date of the gateway route.
* `id` - ID of the gateway route.
* `last_updated_date` - Last update date of the gateway route.
* `resource_owner` - Resource owner's AWS account ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Mesh gateway routes using `mesh_name` and `virtual_gateway_name` together with the gateway route's `name`. For example:

```terraform
import {
  to = aws_appmesh_gateway_route.example
  id = "mesh/gw1/example-gateway-route"
}
```

Using `terraform import`, import App Mesh gateway routes using `mesh_name` and `virtual_gateway_name` together with the gateway route's `name`. For example:

```console
% terraform import aws_appmesh_gateway_route.example mesh/gw1/example-gateway-route
```
