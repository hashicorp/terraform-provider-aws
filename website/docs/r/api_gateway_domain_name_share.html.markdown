---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_domain_name_share"
description: |-
  Manages cross-account sharing for a private API Gateway custom domain name.
---

# Resource: aws_api_gateway_domain_name_share

Manages cross-account sharing for a private API Gateway custom domain name by configuring the domain name management policy.

For more information, see [Share a private custom domain name in API Gateway](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-private-custom-domains-provider-share-cli.html) in the AWS documentation.

## Example Usage

```terraform
resource "aws_api_gateway_domain_name_share" "example" {
  domain_name_id   = aws_api_gateway_domain_name.example.domain_name_id
  allowed_accounts = [
    "111122223333",
    "444455556666",
  ]
}

resource "aws_api_gateway_domain_name" "example" {
  domain_name     = "private.example.com"
  certificate_arn = aws_acm_certificate.example.arn

  endpoint_configuration {
    types = ["PRIVATE"]
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `domain_name_id` - (Required, Forces new resource) Domain name ID of the private custom domain name to share.
* `allowed_accounts` - (Required) Set of AWS account IDs allowed to create domain name access associations for the private custom domain name.

## Attribute Reference

This resource exports no additional attributes beyond the arguments above.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_api_gateway_domain_name_share.example
  identity = {
    domain_name_id = "abcd1234"
  }
}
```

### Identity Schema

#### Required

* `domain_name_id` (String) Domain name ID of the private custom domain name.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway domain name shares using the `domain_name_id`. For example:

```terraform
import {
  to = aws_api_gateway_domain_name_share.example
  id = "abcd1234"
}
```

Using `terraform import`, import API Gateway domain name shares using the `domain_name_id`. For example:

```console
% terraform import aws_api_gateway_domain_name_share.example abcd1234
```
