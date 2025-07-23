---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_route"
description: |-
    Terraform data source for managing an AWS App Mesh Route.
---

# Data Source: aws_appmesh_route

The App Mesh Route data source allows details of an App Mesh Route to be retrieved by its name, mesh_name, virtual_router_name, and optionally the mesh_owner.

## Example Usage

```terraform
data "aws_appmesh_route" "test" {
  name                = "test-route"
  mesh_name           = "test-mesh"
  virtual_router_name = "test-router"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the route.
* `mesh_name` - (Required) Name of the service mesh in which the virtual router exists.
* `virtual_router_name` - (Required) Name of the virtual router in which the route exists.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the route.
* `created_date` - Creation date of the route.
* `last_updated_date` - Last update date of the route.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Route specification. See the [`aws_appmesh_route`](/docs/providers/aws/r/appmesh_route.html#spec) resource for details.
* `tags` - Map of tags.
