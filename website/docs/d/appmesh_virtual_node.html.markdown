---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_node"
description: |-
    Terraform data source for managing an AWS App Mesh Virtual Node.
---

# Data Source: aws_appmesh_virtual_node

Terraform data source for managing an AWS App Mesh Virtual Node.

## Example Usage

```hcl
data "aws_appmesh_virtual_node" "test" {
  name      = "serviceBv1"
  mesh_name = "example-mesh"
}
```

## Argument Reference

This data source supports the following arguments:

* `name` - (Required) Name of the virtual node.
* `mesh_name` - (Required) Name of the service mesh in which the virtual node exists.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the virtual node.
* `created_date` - Creation date of the virtual node.
* `last_updated_date` - Last update date of the virtual node.
* `resource_owner` - Resource owner's AWS account ID.
* `spec` - Virtual node specification. See the [`aws_appmesh_virtual_node`](/docs/providers/aws/r/appmesh_virtual_node.html#spec) resource for details.
* `tags` - Map of tags.
