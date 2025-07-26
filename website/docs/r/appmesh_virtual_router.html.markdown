---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_virtual_router"
description: |-
  Provides an AWS App Mesh virtual router resource.
---

# Resource: aws_appmesh_virtual_router

Provides an AWS App Mesh virtual router resource.

## Breaking Changes

Because of backward incompatible API changes (read [here](https://github.com/awslabs/aws-app-mesh-examples/issues/92) and [here](https://github.com/awslabs/aws-app-mesh-examples/issues/94)), `aws_appmesh_virtual_router` resource definitions created with provider versions earlier than v2.3.0 will need to be modified:

* Remove service `service_names` from the `spec` argument. AWS has created a `aws_appmesh_virtual_service` resource for each service name. Import these resource using `terraform import`.

* Add a `listener` configuration block to the `spec` argument.

The Terraform state associated with existing resources will automatically be migrated.

## Example Usage

```terraform
resource "aws_appmesh_virtual_router" "serviceb" {
  name      = "serviceB"
  mesh_name = aws_appmesh_mesh.simple.id

  spec {
    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name to use for the virtual router. Must be between 1 and 255 characters in length.
* `mesh_name` - (Required) Name of the service mesh in which to create the virtual router. Must be between 1 and 255 characters in length.
* `mesh_owner` - (Optional) AWS account ID of the service mesh's owner. Defaults to the account ID the [AWS provider][1] is currently connected to.
* `spec` - (Required) Virtual router specification to apply.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `spec` object supports the following:

* `listener` - (Optional) Listeners that the virtual router is expected to receive inbound traffic from.
Currently only one listener is supported per virtual router.

The `listener` object supports the following:

* `port_mapping` - (Required) Port mapping information for the listener.

The `port_mapping` object supports the following:

* `port` - (Required) Port used for the port mapping.
* `protocol` - (Required) Protocol used for the port mapping. Valid values are `http`,`http2`, `tcp` and `grpc`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the virtual router.
* `arn` - ARN of the virtual router.
* `created_date` - Creation date of the virtual router.
* `last_updated_date` - Last update date of the virtual router.
* `resource_owner` - Resource owner's AWS account ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Mesh virtual routers using `mesh_name` together with the virtual router's `name`. For example:

```terraform
import {
  to = aws_appmesh_virtual_router.serviceb
  id = "simpleapp/serviceB"
}
```

Using `terraform import`, import App Mesh virtual routers using `mesh_name` together with the virtual router's `name`. For example:

```console
% terraform import aws_appmesh_virtual_router.serviceb simpleapp/serviceB
```

[1]: /docs/providers/aws/index.html
