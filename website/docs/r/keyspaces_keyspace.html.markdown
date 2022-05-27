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

* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the keyspace.
* `arn` - The ARN of the keyspace.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

`aws_keyspaces_keyspace` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

- `create` - (Default `1 minute`) Used for keyspace creation
- `delete` - (Default `1 minute`) Used for keyspace deletion

## Import

Use the `name` to import a keyspace. For example:

```
$ terraform import aws_keyspaces_keyspace.example my_keyspace
```
