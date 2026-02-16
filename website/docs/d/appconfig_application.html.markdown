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
  name = "my-appconfig-application"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `id` - (Optional) ID of the Application. Either `id` or `name` must be specified.
* `name` - (Optional) AWS AppConfig Application name. Either `name` or `id` must be specified.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Application.
* `description` - Description of the Application.
