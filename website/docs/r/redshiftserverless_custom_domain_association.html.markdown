---
subcategory: "Redshift Serverless"
layout: "aws"
page_title: "AWS: aws_redshiftserverless_custom_domain_association"
description: |-
  Terraform resource for managing an AWS Redshift Serverless Custom Domain Association.
---
# Resource: aws_redshiftserverless_custom_domain_association

Terraform resource for managing an AWS Redshift Serverless Custom Domain Association.

## Example Usage

```terraform
resource "aws_acm_certificate" "example" {
  domain_name = "example.com"
  # ...
}

resource "aws_redshiftserverless_namespace" "example" {
  namespace_name = "example-namespace"
}

resource "aws_redshiftserverless_workgroup" "example" {
  workgroup_name = "example-workgroup"
  namespace_name = aws_redshiftserverless_namespace.example.namespace_name
}

resource "aws_redshiftserverless_custom_domain_association" "example" {
  workgroup_name                = aws_redshiftserverless_workgroup.example.workgroup_name
  custom_domain_name            = "example.com"
  custom_domain_certificate_arn = aws_acm_certificate.example.arn
}
```

## Argument Reference

The following arguments are required:

* `workgroup_name` - (Required) Name of the workgroup.
* `custom_domain_name` - (Required) Custom domain to associate with the workgroup.
* `custom_domain_certificate_arn` - (Required) ARN of the certificate for the custom domain association.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `custom_domain_certificate_expiry_time` - Expiration time for the certificate.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Redshift Serverless Custom Domain Association using the `workgroup_name` and `custom_domain_name`, separated by the coma. For example:

```terraform
import {
  to = aws_redshiftserverless_custom_domain_association.example
  id = "example-workgroup,example.com"
}
```

Using `terraform import`, import Redshift Serverless Custom Domain Association using the `workgroup_name` and `custom_domain_name`, separated by the coma. For example:

```console
% terraform import aws_redshiftserverless_custom_domain_association.example example-workgroup,example.com
```
