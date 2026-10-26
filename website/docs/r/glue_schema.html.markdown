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

* `compatibility` - (Required) Compatibility mode of the schema. Values values are: `NONE`, `DISABLED`, `BACKWARD`, `BACKWARD_ALL`, `FORWARD`, `FORWARD_ALL`, `FULL`, and `FULL_ALL`.
* `data_format` - (Required) Data format of the schema definition. Valid values are `AVRO`, `JSON` and `PROTOBUF`.
* `description` - (Optional) Description of the schema.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `registry_arn` - (Optional) ARN of the Glue Registry to create the schema in.
* `schema_definition` - (Required) Schema definition using the `data_format` setting for `schema_name`.
* `schema_name` - (Required) Name of the schema.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the schema.
* `id` - ARN of the schema.
* `latest_schema_version` - Latest version of the schema associated with the returned schema definition.
* `next_schema_version` - Next version of the schema associated with the returned schema definition.
* `registry_name` - Name of the Glue Registry.
* `schema_checkpoint` - Version number of the checkpoint (the last time the compatibility mode was changed).
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_glue_schema.example
  identity = {
    "arn" = "arn:aws:glue:us-west-2:123456789012:schema/example-registry/example-schema"
  }
}

resource "aws_glue_schema" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `arn` (String) ARN of the Glue schema.

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
