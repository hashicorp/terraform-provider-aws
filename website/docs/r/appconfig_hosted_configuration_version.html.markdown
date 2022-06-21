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

The following arguments are supported:

* `application_id` - (Required, Forces new resource) The application ID.
* `configuration_profile_id` - (Required, Forces new resource) The configuration profile ID.
* `content` - (Required, Forces new resource) The content of the configuration or the configuration data.
* `content_type` - (Required, Forces new resource) A standard MIME type describing the format of the configuration content. For more information, see [Content-Type](https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.17).
* `description` - (Optional, Forces new resource) A description of the configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the AppConfig  hosted configuration version.
* `id` - The AppConfig application ID, configuration profile ID, and version number separated by a slash (`/`).
* `version_number` - The version number of the hosted configuration.

## Import

AppConfig Hosted Configuration Versions can be imported by using the application ID, configuration profile ID, and version number separated by a slash (`/`), e.g.,

```
$ terraform import aws_appconfig_hosted_configuration_version.example 71abcde/11xxxxx/2
```
