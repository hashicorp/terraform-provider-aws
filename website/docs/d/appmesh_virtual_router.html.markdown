---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_router"
description: |-
    Provides an AWS App Mesh virtual router resource.
---

# Data Source: aws_appmesh_virtual_router

The App Mesh Virtual Router data source allows details of an App Mesh Virtual Service to be retrieved by its name and mesh_name.

## Example Usage

```hcl
data "aws_appmesh_virtual_router" "test" {
    name = "example-router-name"
    mesh_name = "example-mesh-name"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the virtual router.
* `mesh_name` - (Required) Name of the mesh in which the virtual router exists

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the virtual router.
* `created_date` - Creation date of the virtual router.
* `last_updated_date` - Last update date of the virtual router.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Virtual routers specification.
* `tags` - Map of tags.

### Spec

* `listener` - The listener assigned to the virtual router

### Listener

* `port_mapping` - The port mapping of the listener

### Port_Mapping

* `port` - Port number the listener is listening on.
* `protocol` - Protocol the listener is listening for.
