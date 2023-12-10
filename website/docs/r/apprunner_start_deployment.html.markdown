---
subcategory: "App Runner"
layout: "aws"
page_title: "AWS: aws_apprunner_start_deployment"
description: |-
  Manages an App Runner Start Deployment.
---

# Resource: aws_apprunner_start_deployment

Manages an App Runner Start Deployment.

## Example Usage

```terraform
resource "aws_apprunner_start_deployment" "example" {
  service_arn = aws_apprunner_service.example.arn
}
```

## Argument Reference

The following arguments supported:

* `service_arn` - (Required, Forces new resource) The Amazon Resource Name (ARN) of the App Runner service that you want to manually deploy to.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique ID of the asynchronous operation that the deployment request started.
* `started_at` - The time when the operation started. It's in the Unix time stamp format.
* `status` - The current state of the operation.
* `ended_at` - The time when the operation ended. It's in the Unix time stamp format.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import App Runner Start Deployment using the `id`. For example:

```terraform
import {
  to = aws_apprunner_start_deployment.example
  id = "3663487ba02648919ec45d85df6a3ad7"
}
```

Using `terraform import`, import App Runner Start Deployment using the `id`. For example:

```console
% terraform import aws_apprunner_start_deployment.example 3663487ba02648919ec45d85df6a3ad7
```
