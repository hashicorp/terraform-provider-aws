---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_mesh"
description: |-
    Terraform data source for managing an AWS App Mesh Mesh.
---

# Data Source: aws_appmesh_mesh

The App Mesh Mesh data source allows details of an App Mesh Mesh to be retrieved by its name and optionally the mesh_owner.

## Example Usage

```hcl
data "aws_appmesh_mesh" "simple" {
  name = "simpleapp"
}
```

```hcl
data "aws_caller_identity" "current" {}

data "aws_appmesh_mesh" "simple" {
  name       = "simpleapp"
  mesh_owner = data.aws_caller_identity.current.account_id
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the service mesh.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the service mesh.
* `created_date` - Creation date of the service mesh.
* `last_updated_date` - Last update date of the service mesh.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Service mesh specification. See the [`aws_appmesh_mesh`](/docs/providers/aws/r/appmesh_mesh.html#spec) resource for details.
* `tags` - Map of tags.
