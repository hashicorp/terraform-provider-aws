---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_schema"
description: |-
  Provides a Glue Schema resource.
---

# Resource: aws_glue_schema

Provides a Glue Schema resource.

## Example Usage

```terraform
resource "aws_glue_schema" "example" {
  schema_name       = "example"
  registry_arn      = aws_glue_registry.test.arn
  data_format       = "AVRO"
  compatibility     = "NONE"
  schema_definition = "{\"type\": \"record\", \"name\": \"r1\", \"fields\": [ {\"name\": \"f1\", \"type\": \"int\"}, {\"name\": \"f2\", \"type\": \"string\"} ]}"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `schema_name` - (Required) The Name of the schema.
* `registry_arn` - (Required) The ARN of the Glue Registry to create the schema in.
* `data_format` - (Required) The data format of the schema definition. Valid values are `AVRO`, `JSON` and `PROTOBUF`.
* `compatibility` - (Required) The compatibility mode of the schema. Values values are: `NONE`, `DISABLED`, `BACKWARD`, `BACKWARD_ALL`, `FORWARD`, `FORWARD_ALL`, `FULL`, and `FULL_ALL`.
* `schema_definition` - (Required) The schema definition using the `data_format` setting for `schema_name`.
* `description` - (Optional) A description of the schema.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Amazon Resource Name (ARN) of the schema.
* `id` - Amazon Resource Name (ARN) of the schema.
* `registry_name` - The name of the Glue Registry.
* `latest_schema_version` - The latest version of the schema associated with the returned schema definition.
* `next_schema_version` - The next version of the schema associated with the returned schema definition.
* `schema_checkpoint` - The version number of the checkpoint (the last time the compatibility mode was changed).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Glue Registries using `arn`. For example:

```terraform
import {
  to = aws_glue_schema.example
  id = "arn:aws:glue:us-west-2:123456789012:schema/example/example"
}
```

Using `terraform import`, import Glue Registries using `arn`. For example:

```console
% terraform import aws_glue_schema.example arn:aws:glue:us-west-2:123456789012:schema/example/example
```
