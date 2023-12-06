---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_hosted_configuration_version"
description: |-
  Provides an AppConfig Hosted Configuration Version resource.
---

# Resource: aws_appconfig_hosted_configuration_version

Provides an AppConfig Hosted Configuration Version resource.

## Example Usage

### Freeform

```terraform
resource "aws_appconfig_hosted_configuration_version" "example" {
  application_id           = aws_appconfig_application.example.id
  configuration_profile_id = aws_appconfig_configuration_profile.example.configuration_profile_id
  description              = "Example Freeform Hosted Configuration Version"
  content_type             = "application/json"

  content = jsonencode({
    foo            = "bar",
    fruit          = ["apple", "pear", "orange"],
    isThingEnabled = true
  })
}
```

### Feature Flags

```terraform
resource "aws_appconfig_hosted_configuration_version" "example" {
  application_id           = aws_appconfig_application.example.id
  configuration_profile_id = aws_appconfig_configuration_profile.example.configuration_profile_id
  description              = "Example Feature Flag Configuration Version"
  content_type             = "application/json"

  content = jsonencode({
    flags : {
      foo : {
        name : "foo",
        _deprecation : {
          "status" : "planned"
        }
      },
      bar : {
        name : "bar",
        attributes : {
          someAttribute : {
            constraints : {
              type : "string",
              required : true
            }
          },
          someOtherAttribute : {
            constraints : {
              type : "number",
              required : true
            }
          }
        }
      }
    },
    values : {
      foo : {
        enabled : "true",
      },
      bar : {
        enabled : "true",
        someAttribute : "Hello World",
        someOtherAttribute : 123
      }
    },
    version : "1"
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required, Forces new resource) Application ID.
* `configuration_profile_id` - (Required, Forces new resource) Configuration profile ID.
* `content` - (Required, Forces new resource) Content of the configuration or the configuration data.
* `content_type` - (Required, Forces new resource) Standard MIME type describing the format of the configuration content. For more information, see [Content-Type](https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.17).
* `description` - (Optional, Forces new resource) Description of the configuration.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the AppConfig  hosted configuration version.
* `id` - AppConfig application ID, configuration profile ID, and version number separated by a slash (`/`).
* `version_number` - Version number of the hosted configuration.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppConfig Hosted Configuration Versions using the application ID, configuration profile ID, and version number separated by a slash (`/`). For example:

```terraform
import {
  to = aws_appconfig_hosted_configuration_version.example
  id = "71abcde/11xxxxx/2"
}
```

Using `terraform import`, import AppConfig Hosted Configuration Versions using the application ID, configuration profile ID, and version number separated by a slash (`/`). For example:

```console
% terraform import aws_appconfig_hosted_configuration_version.example 71abcde/11xxxxx/2
```
