---
subcategory: "Elemental MediaStore"
layout: "aws"
page_title: "AWS: aws_media_store_container"
description: |-
  Provides a MediaStore Container.
---

# Resource: aws_media_store_container

Provides a MediaStore Container.

## Example Usage

```terraform
resource "aws_media_store_container" "example" {
  name = "example"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the container. Must contain alphanumeric characters or underscores.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the container.
* `endpoint` - The DNS endpoint of the container.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import MediaStore Container using the MediaStore Container Name. For example:

```terraform
import {
  to = aws_media_store_container.example
  id = "example"
}
```

Using `terraform import`, import MediaStore Container using the MediaStore Container Name. For example:

```console
% terraform import aws_media_store_container.example example
```
