---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_application"
description: |-
  Provides an AppConfig Application resource.
---

# Resource: aws_appconfig_application

Provides an AppConfig Application resource.

## Example Usage

```terraform
resource "aws_appconfig_application" "example" {
  name        = "example-application-tf"
  description = "Example AppConfig Application"

  tags = {
    Type = "AppConfig Application"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name for the application. Must be between 1 and 64 characters in length.
* `description` - (Optional) The description of the application. Can be at most 1024 characters.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the AppConfig Application.
* `id` - The AppConfig application ID.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

AppConfig Applications can be imported using their application ID, e.g.,

```
$ terraform import aws_appconfig_application.example 71rxuzt
```
