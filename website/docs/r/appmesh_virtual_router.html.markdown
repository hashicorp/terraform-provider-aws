---
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_router"
sidebar_current: "docs-aws-resource-appmesh-virtual-router"
description: |-
  Provides an AWS App Mesh virtual router resource.
---

# aws_appmesh_virtual_router

Provides an AWS App Mesh virtual router resource.

~> **Note:** Backward incompatible API changes have been announced for AWS App Mesh which will affect this resource. Read more about the changes [here](https://github.com/awslabs/aws-app-mesh-examples/issues/92).

## Example Usage

```hcl
resource "aws_appmesh_virtual_router" "serviceb" {
  name          = "serviceB"
  mesh_name     = "simpleapp"

  spec {
    service_names = ["serviceb.simpleapp.local"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name to use for the virtual router.
* `mesh_name` - (Required) The name of the service mesh in which to create the virtual router.
* `spec` - (Required) The virtual router specification to apply.

The `spec` object supports the following:

* `service_names` - (Required) The service mesh service names to associate with the virtual router.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the virtual router.
* `arn` - The ARN of the virtual router.
* `created_date` - The creation date of the virtual router.
* `last_updated_date` - The last update date of the virtual router.
