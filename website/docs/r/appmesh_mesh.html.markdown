---
subcategory: "App Mesh"
layout: "aws"
page_title: "AWS: aws_appmesh_mesh"
description: |-
  Provides an AWS App Mesh service mesh resource.
---

# Resource: aws_appmesh_mesh

Provides an AWS App Mesh service mesh resource.

## Example Usage

### Basic

```terraform
resource "aws_appmesh_mesh" "simple" {
  name = "simpleapp"
}
```

### Egress Filter

```terraform
resource "aws_appmesh_mesh" "simple" {
  name = "simpleapp"

  spec {
    egress_filter {
      type = "ALLOW_ALL"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name to use for the service mesh. Must be between 1 and 255 characters in length.
* `spec` - (Optional) Service mesh specification to apply.
    * `egress_filter`- (Optional) Egress filter rules for the service mesh.
        * `type` - (Optional) Egress filter type. By default, the type is `DROP_ALL`. Valid values are `ALLOW_ALL` and `DROP_ALL`.
    * `service_discovery`- (Optional) The service discovery information for the service mesh.
        * `ip_preference` - (Optional) The IP version to use to control traffic within the mesh. Valid values are `IPv6_PREFERRED`, `IPv4_PREFERRED`, `IPv4_ONLY`, and `IPv6_ONLY`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - ID of the service mesh.
* `arn` - ARN of the service mesh.
* `created_date` - Creation date of the service mesh.
* `last_updated_date` - Last update date of the service mesh.
* `mesh_owner` - AWS account ID of the service mesh's owner.
* `resource_owner` - Resource owner's AWS account ID.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Mesh service meshes using the `name`. For example:

```terraform
import {
  to = aws_appmesh_mesh.simple
  id = "simpleapp"
}
```

Using `terraform import`, import App Mesh service meshes using the `name`. For example:

```console
% terraform import aws_appmesh_mesh.simple simpleapp
```
