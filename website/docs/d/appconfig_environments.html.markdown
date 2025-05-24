---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_environments"
description: |-
    Terraform data source for managing an AWS AppConfig Environments.
---

# Data Source: aws_appconfig_environments

Provides access to all Environments for an AppConfig Application. This will allow you to pass Environment IDs to another
resource.

## Example Usage

### Basic Usage

```terraform
data "aws_appconfig_environments" "example" {
  application_id = "a1d3rpe"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `application_id` - (Required) ID of the AppConfig Application.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `environment_ids` - Set of Environment IDs associated with this AppConfig Application.
