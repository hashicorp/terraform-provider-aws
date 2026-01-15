---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_domain_name_api_association"
description: |-
  Provides an AppSync API Association.
---

# Resource: aws_appsync_domain_name_api_association

Provides an AppSync API Association.

## Example Usage

```terraform
resource "aws_appsync_domain_name_api_association" "example" {
  api_id      = aws_appsync_graphql_api.example.id
  domain_name = aws_appsync_domain_name.example.domain_name
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `api_id` - (Required) API ID.
* `domain_name` - (Required) Appsync domain name.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Appsync domain name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_appsync_domain_name_api_association` using the AppSync domain name. For example:

```terraform
import {
  to = aws_appsync_domain_name_api_association.example
  id = "example.com"
}
```

Using `terraform import`, import `aws_appsync_domain_name_api_association` using the AppSync domain name. For example:

```console
% terraform import aws_appsync_domain_name_api_association.example example.com
```
