---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_servicecatalog_portfolio_status"
description: |-
  Manages status of Service Catalog in SageMaker. Service Catalog is used to create SageMaker AI projects.
---

# Resource: aws_sagemaker_servicecatalog_portfolio_status

Manages status of Service Catalog in SageMaker. Service Catalog is used to create SageMaker AI projects.

## Example Usage

Usage:

```terraform
resource "aws_sagemaker_servicecatalog_portfolio_status" "example" {
  status = "Enabled"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `status` - (Required) Whether Service Catalog is enabled or disabled in SageMaker. Valid values are `Enabled` and `Disabled`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The AWS Region the Servicecatalog portfolio status resides in.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to       = aws_sagemaker_servicecatalog_portfolio_status.example
  identity = {}
}

resource "aws_sagemaker_servicecatalog_portfolio_status" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Optional

* `account_id` (String) Account ID where this resource is managed.
* `region` (String) Region where this resource is managed.

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
