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

### AppConfig Application

```hcl
resource "aws_appconfig_application" "test" {
  name        = "test-application-tf"
  description = "Test AppConfig Application"
  tags = {
    Type = "AppConfig Application"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name to use for the application. Must be between 1 and 64 characters in length.
- `description` - (Optional) The description of the application. Can be at most 1024 characters.
- `tags` - (Optional) A map of tags to assign to the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `arn` - The Amazon Resource Name (ARN) of the AppConfig Application.
- `id` - The AppConfig Application ID

## Import

Applications can be imported using their ID, e.g.

```
$ terraform import aws_appconfig_application.bar 71rxuzt
```
