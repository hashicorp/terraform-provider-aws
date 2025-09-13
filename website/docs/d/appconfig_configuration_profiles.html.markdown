---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_configuration_profiles"
description: |-
    Terraform data source for managing an AWS AppConfig Configuration Profiles.
---

# Data Source: aws_appconfig_configuration_profiles

Provides access to all Configuration Properties for an AppConfig Application. This will allow you to pass Configuration
Profile IDs to another resource.

## Example Usage

### Basic Usage

```terraform
data "aws_appconfig_configuration_profiles" "example" {
  application_id = "a1d3rpe"
}

data "aws_appconfig_configuration_profile" "example" {
  for_each                 = data.aws_appconfig_configuration_profiles.example.configuration_profile_ids
  configuration_profile_id = each.value
  application_id           = aws_appconfig_application.example.id
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `application_id` - (Required) ID of the AppConfig Application.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `configuration_profile_ids` - Set of Configuration Profile IDs associated with the AppConfig Application.
