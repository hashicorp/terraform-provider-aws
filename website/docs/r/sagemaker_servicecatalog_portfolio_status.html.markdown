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

The following arguments are supported:

* `status` - (Required) Whether Service Catalog is enabled or disabled in SageMaker. Valid values are `Enabled` and `Disabled`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The AWS Region the Servicecatalog portfolio status resides in.

## Import

Models can be imported using the `id`, e.g.,

```
$ terraform import aws_sagemaker_servicecatalog_portfolio_status.example us-east-1
```
