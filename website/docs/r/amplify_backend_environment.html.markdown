---
subcategory: "Amplify"
layout: "aws"
page_title: "AWS: aws_amplify_backend_environment"
description: |-
  Provides an Amplify backend environment resource.
---

# Resource: aws_amplify_backend_environment

Provides an Amplify backend environment resource.

## Example Usage

```hcl
resource "aws_amplify_app" "app" {
  name = "app"
}

resource "aws_amplify_backend_environment" "prod" {
  app_id           = aws_amplify_app.app.id
  environment_name = "prod"

  deployment_artifacts = "app-prod-deployment"
  stack_name           = "amplify-app-prod"
}
```

## Argument Reference

The following arguments are supported:

* `app_id` - (Required) Unique Id for an Amplify App.
* `environment_name` - (Required) Name for the backend environment.
* `deployment_artifacts` - (Optional) Name of deployment artifacts.
* `stack_name` - (Optional) CloudFormation stack name of backend environment.

## Attribute Reference

The following attributes are exported:

* `arn` - Arn for the backend environment.

## Import

Amplify backend environment can be imported using `app_id` and `environment_name`, e.g.

```
$ terraform import aws_amplify_backend_environment.prod d2ypk4k47z8u6/backendenvironments/prod
```
