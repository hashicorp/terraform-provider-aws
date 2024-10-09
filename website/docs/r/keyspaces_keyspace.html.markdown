---
subcategory: "Keyspaces (for Apache Cassandra)"
layout: "aws"
page_title: "AWS: aws_keyspaces_keyspace"
description: |-
  Provides a Keyspaces Keyspace.
---

# Resource: aws_keyspaces_keyspace

Provides a Keyspaces Keyspace.

More information about keyspaces can be found in the [Keyspaces User Guide](https://docs.aws.amazon.com/keyspaces/latest/devguide/what-is-keyspaces.html).

## Example Usage

```terraform
resource "aws_keyspaces_keyspace" "example" {
  name = "my_keyspace"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) The name of the keyspace to be created.

The following arguments are optional:

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the keyspace.
* `arn` - The ARN of the keyspace.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `1m`)
- `delete` - (Default `1m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a keyspace using the `name`. For example:

```terraform
import {
  to = aws_keyspaces_keyspace.example
  id = "my_keyspace"
}
```

Using `terraform import`, import a keyspace using the `name`. For example:

```console
% terraform import aws_keyspaces_keyspace.example my_keyspace
```
