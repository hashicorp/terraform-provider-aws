---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_gateway_route"
description: |-
    Terraform data source for managing an AWS App Mesh Gateway Route.
---

# Data Source: aws_appmesh_gateway_route

The App Mesh Gateway Route data source allows details of an App Mesh Gateway Route to be retrieved by its name, mesh_name, virtual_gateway_name, and optionally the mesh_owner.

## Example Usage

```hcl
data "aws_appmesh_gateway_route" "test" {
  name                 = "test-route"
  mesh_name            = "test-mesh"
  virtual_gateway_name = "test-gateway"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the gateway route.
* `mesh_name` - (Required) Name of the service mesh in which the virtual gateway exists.
* `virtual_gateway_name` - (Required) Name of the virtual gateway in which the route exists.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the gateway route.
* `created_date` - Creation date of the gateway route.
* `last_updated_date` - Last update date of the gateway route.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Gateway route specification. See the [`aws_appmesh_gateway_route`](/docs/providers/aws/r/appmesh_gateway_route.html#spec) resource for details.
* `tags` - Map of tags.
