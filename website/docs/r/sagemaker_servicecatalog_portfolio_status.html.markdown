---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_servicecatalog_portfolio_status"
description: |-
  Manages status of Service Catalog in SageMaker. Service Catalog is used to create SageMaker projects.
---

# Resource: aws_sagemaker_servicecatalog_portfolio_status

Manages status of Service Catalog in SageMaker. Service Catalog is used to create SageMaker projects.

## Example Usage

Usage:

```terraform
resource "aws_sagemaker_servicecatalog_portfolio_status" "example" {
  status = "Enabled"
}
```

## Argument Reference

This resource supports the following arguments:

* `status` - (Required) Whether Service Catalog is enabled or disabled in SageMaker. Valid values are `Enabled` and `Disabled`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS Region the Servicecatalog portfolio status resides in.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import models using the `id`. For example:

```terraform
import {
  to = aws_sagemaker_servicecatalog_portfolio_status.example
  id = "us-east-1"
}
```

Using `terraform import`, import models using the `id`. For example:

```console
% terraform import aws_sagemaker_servicecatalog_portfolio_status.example us-east-1
```
