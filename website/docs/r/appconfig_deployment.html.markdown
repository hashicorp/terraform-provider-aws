---
subcategory: "AppConfig"
layout: "aws"
page_title: "AWS: aws_appconfig_deployment"
description: |-
  Provides an AppConfig Deployment resource.
---

# Resource: aws_appconfig_deployment

Provides an AppConfig Deployment resource for an [`aws_appconfig_application` resource](appconfig_application.html.markdown).

## Example Usage

```terraform
resource "aws_appconfig_deployment" "test" {
  application_id           = aws_appconfig_application.example.id
  configuration_profile_id = aws_appconfig_configuration_profile.example.configuration_profile_id
  configuration_version    = "00000000-0000-0000-0000-000000000000"
  deployment_strategy_id   = aws_appconfig_deployment_strategy.example.id
  description              = "My test deployment"
  environment_id           = aws_appconfig_environment.example.environment_id
  tags = {
    Env = "test"
  }
}
```

## Argument Reference

The following arguments are supported:

* `application_id` - (Required, Forces new resource) The application ID. Must be between 4 and 7 characters in length.
* `configuration_profile_id` - (Required, Forces new resource) The configuration profile ID. Must be between 4 and 7 characters in length.
* `configuration_version` - (Required, Forces new resource) The configuration version to deploy. Can be at most 1024 characters.
* `deployment_strategy_id` - (Required, Forces new resource) The deployment strategy ID. Must be between 4 and 7 characters in length.
* `description` - (Optional, Fources new resource) The description of the deployment. Can be at most 1024 characters.
* `environment_id` - (Required, Forces new resource) The environment ID. Must be between 4 and 7 characters in length.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the AppConfig Environment.
* `id` - The AppConfig application ID and environment ID and deployment number separated by a colon (`/`).
* `deployment_number` - The deployment number.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

AppConfig Deployment can be imported by using the application ID and environment ID and deployment number separated by a colon (`/`), e.g.

```
$ terraform import aws_appconfig_deployment.example 71abcde/11xxxxx/1
```
