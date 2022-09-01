---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_service"
description: |-
    Provides an AWS App Mesh virtual service resource.
---

# Data Source: aws_appmesh_virtual_service

The App Mesh Virtual Service data source allows details of an App Mesh Virtual Service to be retrieved by its name, mesh_name, and optionally the mesh_owner.

## Example Usage

```hcl
data "aws_appmesh_virtual_service" "test" {
  name      = "example.mesh.local"
  mesh_name = "example-mesh"
}
```

```hcl
data "aws_caller_identity" "current" {}

data "aws_appmesh_virtual_service" "test" {
  name       = "example.mesh.local"
  mesh_name  = "example-mesh"
  mesh_owner = data.aws_caller_identity.current.account_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the virtual service.
* `mesh_name` - (Required) Name of the service mesh in which the virtual service exists.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the virtual service.
* `created_date` - Creation date of the virtual service.
* `last_updated_date` - Last update date of the virtual service.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Virtual service specification
* `tags` - Map of tags.

### Spec

* `provider` - App Mesh object that is acting as the provider for a virtual service.

### Provider

* `virtual_node` - Virtual node associated with the virtual service.
* `virtual_router` - Virtual router associated with the virtual service.

### Virtual Node

* `virtual_node_name` - Name of the virtual node that is acting as a service provider.

### Virtual Router

* `virtual_router_name` - Name of the virtual router that is acting as a service provider.
