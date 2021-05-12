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

### AppConfig Hosted Configuration Version

```hcl
resource "aws_appconfig_hosted_configuration_version" "production" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.id
  description              = "test"
  content_type             = "application/json"
  content = jsonencode({
    foo = "foo"
  })
}

resource "aws_appconfig_application" "test" {
  name = "test"
}

resource "aws_appconfig_configuration_profile" "test" {
  name           = "test"
  application_id = aws_appconfig_application.test.id
  location_uri   = "hosted"
}
```

## Argument Reference

The following arguments are supported:

- `application_id` - (Required) The application id.
- `configuration_profile_id` - (Required) The configuration profile ID.
- `content` - (Required) The content of the configuration or the configuration data.
- `content_type` - (Required) A standard MIME type describing the format of the configuration content.
- `description` - (Optional) A description of the configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `version_number` - hosted configuration version
- `id` - `<application id>/<configuration profile id>/<version number>`

## Import

`aws_appconfig_hosted_configuration_version` can be imported by the Application ID and Configuration Profile ID and Hosted Configuration Version Number, e.g.

```
$ terraform import aws_appconfig_hosted_configuration_version.test 71abcde/11xxxxx/2
```
