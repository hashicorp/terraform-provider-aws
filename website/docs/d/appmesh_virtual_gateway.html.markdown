---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_gateway"
description: |-
  Terraform data source for managing an AWS App Mesh Virtual Gateway.
---

# Data Source: aws_appmesh_virtual_gateway

Terraform data source for managing an AWS App Mesh Virtual Gateway.

## Example Usage

### Basic Usage

```hcl
data "aws_appmesh_virtual_gateway" "example" {
  mesh_name = "mesh-gateway"
  name      = "example-mesh"
}
```

```hcl
data "aws_caller_identity" "current" {}

data "aws_appmesh_virtual_gateway" "test" {
  name       = "example.mesh.local"
  mesh_name  = "example-mesh"
  mesh_owner = data.aws_caller_identity.current.account_id
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the virtual gateway.
* `mesh_name` - (Required) Name of the service mesh in which the virtual gateway exists.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the virtual gateway.
* `created_date` - Creation date of the virtual gateway.
* `last_updated_date` - Last update date of the virtual gateway.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Virtual gateway specification. See the [`aws_appmesh_virtual_gateway`](/docs/providers/aws/r/appmesh_virtual_gateway.html#spec) resource for details.
* `tags` - Map of tags.
