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
resource "aws_appconfig_deployment" "example" {
  application_id           = aws_appconfig_application.example.id
  configuration_profile_id = aws_appconfig_configuration_profile.example.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.example.version_number
  deployment_strategy_id   = aws_appconfig_deployment_strategy.example.id
  description              = "My example deployment"
  environment_id           = aws_appconfig_environment.example.environment_id

  tags = {
    Type = "AppConfig Deployment"
  }
}
```

## Argument Reference

The following arguments are supported:

* `application_id` - (Required, Forces new resource) The application ID. Must be between 4 and 7 characters in length.
* `configuration_profile_id` - (Required, Forces new resource) The configuration profile ID. Must be between 4 and 7 characters in length.
* `configuration_version` - (Required, Forces new resource) The configuration version to deploy. Can be at most 1024 characters.
* `deployment_strategy_id` - (Required, Forces new resource) The deployment strategy ID or name of a predefined deployment strategy. See [Predefined Deployment Strategies](https://docs.aws.amazon.com/appconfig/latest/userguide/appconfig-creating-deployment-strategy.html#appconfig-creating-deployment-strategy-predefined) for more details.
* `description` - (Optional, Forces new resource) The description of the deployment. Can be at most 1024 characters.
* `environment_id` - (Required, Forces new resource) The environment ID. Must be between 4 and 7 characters in length.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The AppConfig application ID, environment ID, and deployment number separated by a slash (`/`).
* `arn` - The Amazon Resource Name (ARN) of the AppConfig Deployment.
* `deployment_number` - The deployment number.
* `state` - The state of the deployment.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

AppConfig Deployments can be imported by using the application ID, environment ID, and deployment number separated by a slash (`/`), e.g.,

```
$ terraform import aws_appconfig_deployment.example 71abcde/11xxxxx/1
```
