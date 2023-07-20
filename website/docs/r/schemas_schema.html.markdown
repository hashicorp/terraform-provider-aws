---
subcategory: "EventBridge Schemas"
layout: "aws"
page_title: "AWS: aws_schemas_schema"
description: |-
  Provides an EventBridge Schema resource.
---

# Resource: aws_schemas_schema

Provides an EventBridge Schema resource.

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

```terraform
resource "aws_schemas_registry" "test" {
  name = "my_own_registry"
}

resource "aws_schemas_schema" "test" {
  name          = "my_schema"
  registry_name = aws_schemas_registry.test.name
  type          = "OpenApi3"
  description   = "The schema definition for my event"

  content = jsonencode({
    "openapi" : "3.0.0",
    "info" : {
      "version" : "1.0.0",
      "title" : "Event"
    },
    "paths" : {},
    "components" : {
      "schemas" : {
        "Event" : {
          "type" : "object",
          "properties" : {
            "name" : {
              "type" : "string"
            }
          }
        }
      }
    }
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) The name of the schema. Maximum of 385 characters consisting of lower case letters, upper case letters, ., -, _, @.
* `content` - (Required) The schema specification. Must be a valid Open API 3.0 spec.
* `registry_name` - (Required) The name of the registry in which this schema belongs.
* `type` - (Required) The type of the schema. Valid values: `OpenApi3`.
* `description` - (Optional) The description of the schema. Maximum of 256 characters.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the discoverer.
* `last_modified` - The last modified date of the schema.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `version` - The version of the schema.
* `version_created_date` - The created date of the version of the schema.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EventBridge schema using the `name` and `registry_name`. For example:

```terraform
import {
  to = aws_schemas_schema.test
  id = "name/registry"
}
```

Using `terraform import`, import EventBridge schema using the `name` and `registry_name`. For example:

```console
% terraform import aws_schemas_schema.test name/registry
```
