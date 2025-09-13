---
subcategory: "Service Catalog"
layout: "aws"
page_title: "AWS: aws_servicecatalog_principal_portfolio_association"
description: |-
  Manages a Service Catalog Principal Portfolio Association
---

# Resource: aws_servicecatalog_principal_portfolio_association

Manages a Service Catalog Principal Portfolio Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_servicecatalog_principal_portfolio_association" "example" {
  portfolio_id  = "port-68656c6c6f"
  principal_arn = "arn:aws:iam::123456789012:user/Eleanor"
}
```

## Argument Reference

The following arguments are required:

* `portfolio_id` - (Required) Portfolio identifier.
* `principal_arn` - (Required) Principal ARN.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `accept_language` - (Optional) Language code. Valid values: `en` (English), `jp` (Japanese), `zh` (Chinese). Default value is `en`.
* `principal_type` - (Optional) Principal type. Setting this argument empty (e.g., `principal_type = ""`) will result in an error. Valid values are `IAM` and `IAM_PATTERN`. Default is `IAM`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Identifier of the association.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

- `create` - (Default `3m`)
- `read` - (Default `10m`)
- `delete` - (Default `3m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_servicecatalog_principal_portfolio_association` using `accept_language`, `principal_arn`, `portfolio_id`, and `principal_type` separated by a comma. For example:

```terraform
import {
  to = aws_servicecatalog_principal_portfolio_association.example
  id = "en,arn:aws:iam::123456789012:user/Eleanor,port-68656c6c6f,IAM"
}
```

Using `terraform import`, import `aws_servicecatalog_principal_portfolio_association` using `accept_language`, `principal_arn`, `portfolio_id`, and `principal_type` separated by a comma. For example:

```console
% terraform import aws_servicecatalog_principal_portfolio_association.example en,arn:aws:iam::123456789012:user/Eleanor,port-68656c6c6f,IAM
```
