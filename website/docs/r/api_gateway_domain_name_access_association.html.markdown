---
subcategory: "API Gateway"
layout: "aws"
page_title: "AWS: aws_api_gateway_domain_name_access_association"
description: |-
  Creates a domain name access association resource between an access association source and a private custom domain name.
---

# Resource: aws_api_gateway_domain_name_access_association

Creates a domain name access association resource between an access association source and a private custom domain name.

## Example Usage

```terraform
resource "aws_api_gateway_domain_name_access_association" "example" {
  access_association_source      = aws_vpc_endpoint.example.id
  access_association_source_type = "VPCE"
  domain_name_arn                = aws_api_gateway_domain_name.example.domain_name_arn
}
```

## Argument Reference

This resource supports the following arguments:

* `access_association_source` - (Required) The identifier of the domain name access association source. For a `VPCE`, the value is the VPC endpoint ID.
* `access_association_source_type` - (Required) The type of the domain name access association source. Valid values are `VPCE`.
* `domain_name_arn` - (Required) The ARN of the domain name.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the domain name access association.
* `id` - (**Deprecated**, use `arn` instead) Internal identifier assigned to this domain name access association.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway domain name acces associations using their `arn`. For example:

```terraform
import {
  to = aws_api_gateway_domain_name_access_association.example
  id = "arn:aws:apigateway:us-west-2:123456789012:/domainnameaccessassociations/domainname/12qmzgp2.9m7ilski.test+hykg7a12e7/vpcesource/vpce-05de3f8f82740a748"
}
```

Using `terraform import`, import API Gateway domain name acces associations as using their `arn`. For example:

```console
% terraform import aws_api_gateway_domain_name_access_association.example arn:aws:apigateway:us-west-2:123456789012:/domainnameaccessassociations/domainname/12qmzgp2.9m7ilski.test+hykg7a12e7/vpcesource/vpce-05de3f8f82740a748
```
