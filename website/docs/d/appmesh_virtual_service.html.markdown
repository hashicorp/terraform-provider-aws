---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_service"
description: |-
    Terraform data source for managing an AWS App Mesh Virtual Service.
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

This data source supports the following arguments:

* `name` - (Required) Name of the virtual service.
* `mesh_name` - (Required) Name of the service mesh in which the virtual service exists.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the virtual service.
* `created_date` - Creation date of the virtual service.
* `last_updated_date` - Last update date of the virtual service.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Virtual service specification. See the [`aws_appmesh_virtual_service`](/docs/providers/aws/r/appmesh_virtual_service.html#spec) resource for details.
* `tags` - Map of tags.
