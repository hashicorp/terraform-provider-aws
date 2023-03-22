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

The following arguments are supported:

* `name` - (Required) Name to use for the gateway route. Must be between 1 and 255 characters in length.
* `mesh_name` - (Required) Name of the service mesh in which to create the gateway route. Must be between 1 and 255 characters in length.
* `virtual_gateway_name` - (Required) Name of the [virtual gateway](/docs/providers/aws/r/appmesh_virtual_gateway.html) to associate the gateway route with. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `spec` - (Required) Gateway route specification to apply.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `spec` object supports the following:

* `grpc_route` - (Optional) Specification of a gRPC gateway route.
* `http_route` - (Optional) Specification of an HTTP gateway route.
* `http2_route` - (Optional) Specification of an HTTP/2 gateway route.

The `grpc_route`, `http_route` and `http2_route` objects supports the following:

* `action` - (Required) Action to take if a match is determined.
* `match` - (Required) Criteria for determining a request match.

The `grpc_route`, `http_route` and `http2_route`'s `action` object supports the following:

* `target` - (Required) Target that traffic is routed to when a request matches the gateway route.

The `target` object supports the following:

* `virtual_service` - (Required) Virtual service gateway route target.

The `virtual_service` object supports the following:

* `virtual_service_name` - (Required) Name of the virtual service that traffic is routed to. Must be between 1 and 255 characters in length.

The `http_route` and `http2_route`'s `action` object additionally supports the following:

* `rewrite` - (Optional) Gateway route action to rewrite.

The `rewrite` object supports the following:

* `hostname` - (Optional) Host name to rewrite.
* `prefix` - (Optional) Specified beginning characters to rewrite.
* `port` - (Optional) The port number to match from the request.

The `hostname` object supports the following:

* `default_target_hostname` - (Required) Default target host name to write to. Valid values: `ENABLED`, `DISABLED`.

The `prefix` object supports the following:

* `default_prefix` - (Optional) Default prefix used to replace the incoming route prefix when rewritten. Valid values: `ENABLED`, `DISABLED`.
* `value` - (Optional) Value used to replace the incoming route prefix when rewritten.

The `grpc_route`'s `match` object supports the following:

* `service_name` - (Required) Fully qualified domain name for the service to match from the request.
* `port` - (Optional) The port number to match from the request.

The `http_route` and `http2_route`'s `match` object supports the following:

* `hostname` - (Optional) Host name to match on.
* `prefix` - (Required) Path to match requests with. This parameter must always start with `/`, which by itself matches all requests to the virtual service name.
* `port` - (Optional) The port number to match from the request.

The `hostname` object supports the following:

* `exact` - (Optional) Exact host name to match on.
* `suffix` - (Optional) Specified ending characters of the host name to match on.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the gateway route.
* `arn` - ARN of the gateway route.
* `created_date` - Creation date of the gateway route.
* `last_updated_date` - Last update date of the gateway route.
* `resource_owner` - Resource owner's AWS account ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

App Mesh gateway routes can be imported using `mesh_name` and `virtual_gateway_name` together with the gateway route's `name`,
e.g.,

```
$ terraform import aws_appmesh_gateway_route.example mesh/gw1/example-gateway-route
```

[1]: /docs/providers/aws/index.html
