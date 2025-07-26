---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_service"
description: |-
  Provides an AWS App Mesh virtual service resource.
---

# Resource: aws_appmesh_virtual_service

Provides an AWS App Mesh virtual service resource.

## Example Usage

### Virtual Node Provider

```terraform
resource "aws_appmesh_virtual_service" "servicea" {
  name      = "servicea.simpleapp.local"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    provider {
      virtual_node {
        virtual_node_name = aws_appmesh_virtual_node.serviceb1.name
      }
    }
  }
}
```

### Virtual Router Provider

```terraform
resource "aws_appmesh_virtual_service" "servicea" {
  name      = "servicea.simpleapp.local"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    provider {
      virtual_router {
        virtual_router_name = aws_appmesh_virtual_router.serviceb.name
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name to use for the virtual service. Must be between 1 and 255 characters in length.
* `mesh_name` - (Required) Name of the service mesh in which to create the virtual service. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `spec` - (Required) Virtual service specification to apply.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `spec` object supports the following:

* `provider`- (Optional) App Mesh object that is acting as the provider for a virtual service. You can specify a single virtual node or virtual router.

The `provider` object supports the following:

* `virtual_node` - (Optional) Virtual node associated with a virtual service.
* `virtual_router` - (Optional) Virtual router associated with a virtual service.

The `virtual_node` object supports the following:

* `virtual_node_name` - (Required) Name of the virtual node that is acting as a service provider. Must be between 1 and 255 characters in length.

The `virtual_router` object supports the following:

* `virtual_router_name` - (Required) Name of the virtual router that is acting as a service provider. Must be between 1 and 255 characters in length.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the virtual service.
* `arn` - ARN of the virtual service.
* `created_date` - Creation date of the virtual service.
* `last_updated_date` - Last update date of the virtual service.
* `resource_owner` - Resource owner's AWS account ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Mesh virtual services using `mesh_name` together with the virtual service's `name`. For example:

```terraform
import {
  to = aws_appmesh_virtual_service.servicea
  id = "simpleapp/servicea.simpleapp.local"
}
```

Using `terraform import`, import App Mesh virtual services using `mesh_name` together with the virtual service's `name`. For example:

```console
% terraform import aws_appmesh_virtual_service.servicea simpleapp/servicea.simpleapp.local
```

[1]: /docs/providers/aws/index.html
