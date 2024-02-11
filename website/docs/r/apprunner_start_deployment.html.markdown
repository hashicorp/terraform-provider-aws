---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_start_deployment"
description: |-
  Manages an App Runner Start Deployment Operation.
---

# Resource: aws_apprunner_start_deployment

Manages an App Runner Start Deployment Operation.

## Example Usage

```terraform
resource "aws_apprunner_start_deployment" "example" {
  service_arn = aws_apprunner_service.example.arn
}
```

## Argument Reference

The following arguments supported:

* `service_arn` - (Required) The Amazon Resource Name (ARN) of the App Runner service to start the deployment for.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `operation_id` - The unique ID of the operation associated with deployment.
* `status` - The current status of the App Runner service deployment.
