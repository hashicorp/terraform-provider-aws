---
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_service"
sidebar_current: "docs-aws-resource-appmesh-virtual-service"
description: |-
  Provides an AWS App Mesh virtual service resource.
---

# aws_appmesh_virtual_service

Provides an AWS App Mesh virtual service resource.

## Example Usage

### Virtual Node Provider

```hcl
resource "aws_appmesh_virtual_service" "servicea" {
  name                = "servicea.simpleapp.local"
  mesh_name           = "simpleapp"

  spec {
    provider {
      virtual_node {
        virtual_node_name = "serviceBv1"
      }
    }
  }
}
```

### Virtual Router Provider

```hcl
resource "aws_appmesh_virtual_service" "servicea" {
  name                = "servicea.simpleapp.local"
  mesh_name           = "simpleapp"

  spec {
    provider {
      virtual_router {
        virtual_router_name = "serviceB"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the virtual service.
* `mesh_name` - (Required) The name of the service mesh in which to create the virtual service.
* `spec` - (Required) The virtual service specification to apply.

The `spec` object supports the following:

* `provider`- (Required) The provider information for the virtual node.

The `provider` object supports the following:

* `virtual_node` - (Optional) The virtual node information for the provider.
* `virtual_router` - (Optional) The virtual router information for the provider.

The `virtual_node` object supports the following:

* `virtual_node_name` - (Required) Specifies the name of the virtual node.

The `virtual_router` object supports the following:

* `virtual_router_name` - (Required) The name of the virtual router.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual service.
* `arn` - The ARN of the virtual service.
* `created_date` - The creation date of the virtual service.
* `last_updated_date` - The last update date of the virtual service.

## Import

App Mesh virtual services can be imported using `mesh_name` together with the virtual service's `name`,
e.g.

```
$ terraform import aws_appmesh_virtual_service.servicea simpleapp/servicea.simpleapp.local
```
