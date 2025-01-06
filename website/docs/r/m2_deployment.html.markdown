---
subcategory: "Mainframe Modernization"
layout: "aws"
page_title: "AWS: aws_m2_deployment"
description: |-
  Terraform resource for managing an AWS Mainframe Modernization Deployment.
---
# Resource: aws_m2_deployment

Terraform resource for managing an [AWS Mainframe Modernization Deployment.](https://docs.aws.amazon.com/m2/latest/userguide/applications-m2-deploy.html)

## Example Usage

### Basic Usage

```terraform
resource "aws_m2_deployment" "test" {
  environment_id      = "01234567890abcdef012345678"
  application_id      = "34567890abcdef012345678012"
  application_version = 1
  start               = true
}
```

## Argument Reference

The following arguments are required:

* `environment_id` - (Required) Environment to deploy application to.
* `application_id` - (Required) Application to deploy.
* `application_version` - (Required) Version to application to deploy
* `start` - (Required) Start the application once deployed.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `60m`)
* `delete` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Mainframe Modernization Deployment using the `APPLICATION-ID,DEPLOYMENT-ID`. For example:

```terraform
import {
  to = aws_m2_deployment.example
  id = "APPLICATION-ID,DEPLOYMENT-ID"
}
```

Using `terraform import`, import Mainframe Modernization Deployment using the `APPLICATION-ID,DEPLOYMENT-ID`. For example:

```console
% terraform import aws_m2_deployment.example APPLICATION-ID,DEPLOYMENT-ID
```
