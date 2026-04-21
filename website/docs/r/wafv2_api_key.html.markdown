---
subcategory: "WAF"
layout: "aws"
page_title: "AWS: aws_wafv2_api_key"
description: |-
  Provides a WAFv2 API Key resource.
---

# Resource: aws_wafv2_api_key

Provides an AWS WAFv2 API Key resource.

## Example Usage

```terraform
resource "aws_wafv2_api_key" "example" {
  scope         = "REGIONAL"
  token_domains = ["example.com"]
}
```

## Argument Reference

This resource supports the following arguments:

- `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
- `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. Changing this forces a new resource to be created. **NOTE:** WAFv2 API Keys deployed for `CLOUDFRONT` must be created within the `us-east-1` region.
- `token_domains` - (Required) The domains that you want to be able to use the API key with, for example `example.com`. You can specify up to 5 domains. Changing this forces a new resource to be created.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

- `api_key` - The generated API key. This value is sensitive.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import WAFv2 API Key using `api_key,scope`. For example:

```terraform
import {
  to = aws_wafv2_api_key.example
  id = "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111,REGIONAL"
}
```

Using `terraform import`, import WAFv2 API Key using `api_key,scope`. For example:

```console
% terraform import aws_wafv2_api_key.example a1b2c3d4-5678-90ab-cdef-EXAMPLE11111,REGIONAL
```
