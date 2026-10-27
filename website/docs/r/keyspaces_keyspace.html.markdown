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

This resource supports the following arguments:

* `name` - (Required, Forces new resource) Name of the keyspace to be created.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `replication_specification` - (Optional) Replication specification of the keyspace. [See below](#replication_specification-block).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `replication_specification` Block

* `region_list` - (Optional) Replication regions. If `replication_strategy` is `MULTI_REGION`, `region_list` requires the current Region and at least one additional AWS Region where the keyspace is going to be replicated in.
* `replication_strategy` - (Optional) Replication strategy. Valid values: `SINGLE_REGION` and `MULTI_REGION`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the keyspace.
* `id` - Name of the keyspace.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
