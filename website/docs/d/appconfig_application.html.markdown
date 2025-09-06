---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_application"
description: |-
  Retrieves an AWS AppConfig Application by name.
---

# Data Source: aws_appconfig_application

Provides details about an AWS AppConfig Application.

## Example Usage

### Basic Usage

```terraform
data "aws_appconfig_application" "example" {
  name = "my-appconfig-application
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) AWS AppConfig Application name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `description` - Description of the Application.
* `id` - ID of the Application.
* `name` - Name of the Application
