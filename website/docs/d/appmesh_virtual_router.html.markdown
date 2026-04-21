---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_router"
description: |-
    Terraform data source for managing an AWS App Mesh Virtual Router.
---

# Data Source: aws_appmesh_virtual_router

The App Mesh Virtual Router data source allows details of an App Mesh Virtual Service to be retrieved by its name and mesh_name.

## Example Usage

```terraform
data "aws_appmesh_virtual_router" "test" {
  name      = "example-router-name"
  mesh_name = "example-mesh-name"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the virtual router.
* `mesh_name` - (Required) Name of the mesh in which the virtual router exists

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the virtual router.
* `created_date` - Creation date of the virtual router.
* `last_updated_date` - Last update date of the virtual router.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Virtual routers specification. See the [`aws_appmesh_virtual_router`](/docs/providers/aws/r/appmesh_virtual_router.html#spec) resource for details.
* `tags` - Map of tags.
