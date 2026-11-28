---
subcategory: "API Gateway V2"
layout: "aws"
page_title: "AWS: aws_apigatewayv2_portal_product_sharing_policy"
description: |-
  Manages the sharing policy of an Amazon API Gateway portal product.
---

# Resource: aws_apigatewayv2_portal_product_sharing_policy

Manages the sharing policy of an Amazon API Gateway portal product. The sharing policy is a resource-based policy that controls which other AWS accounts can add the portal product to their own portals. Accounts that the policy allows can view the portal product and its pages, but cannot modify them.

A portal product has at most one sharing policy. Deleting this resource removes the policy, which stops new access from other accounts.

More information can be found in the [Amazon API Gateway Developer Guide](https://docs.aws.amazon.com/apigateway/latest/developerguide/apigateway-portals-share-resources.html).

## Example Usage

### Share with an AWS Account

```terraform
resource "aws_apigatewayv2_portal_product_sharing_policy" "example" {
  portal_product_id = aws_apigatewayv2_portal_product.example.portal_product_id

  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        AWS = "arn:aws:iam::123456789012:root"
      }
      Action = [
        "apigateway:GetPortalProduct",
        "apigateway:ListProductPages",
        "apigateway:ListProductRestEndpointPages",
        "apigateway:GetProductPage",
        "apigateway:GetProductRestEndpointPage",
      ]
      Resource = [
        aws_apigatewayv2_portal_product.example.portal_product_arn,
        "${aws_apigatewayv2_portal_product.example.portal_product_arn}/*",
      ]
    }]
  })
}
```

### Share with an Organization

A statement that uses the `aws:PrincipalOrgID` or `aws:PrincipalOrgPaths` condition key must also exclude the account that owns the portal product with an `aws:PrincipalAccount` condition.

```terraform
data "aws_caller_identity" "current" {}

data "aws_organizations_organization" "current" {}

resource "aws_apigatewayv2_portal_product_sharing_policy" "example" {
  portal_product_id = aws_apigatewayv2_portal_product.example.portal_product_id

  policy_document = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action = [
        "apigateway:GetPortalProduct",
        "apigateway:ListProductPages",
        "apigateway:ListProductRestEndpointPages",
        "apigateway:GetProductPage",
        "apigateway:GetProductRestEndpointPage",
      ]
      Resource = [
        aws_apigatewayv2_portal_product.example.portal_product_arn,
        "${aws_apigatewayv2_portal_product.example.portal_product_arn}/*",
      ]
      Condition = {
        StringNotEquals = {
          "aws:PrincipalAccount" = data.aws_caller_identity.current.account_id
        }
        StringEquals = {
          "aws:PrincipalOrgID" = data.aws_organizations_organization.current.id
        }
      }
    }]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `policy_document` - (Required) JSON sharing policy document. Between 1 and 307,200 characters.
* `portal_product_id` - (Required) ID of the portal product. Changing this argument creates a new resource.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_apigatewayv2_portal_product_sharing_policy.example
  identity = {
    portal_product_id = "abcdef1234"
  }
}

resource "aws_apigatewayv2_portal_product_sharing_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `portal_product_id` (String) ID of the portal product.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import API Gateway portal product sharing policies using `portal_product_id`. For example:

```terraform
import {
  to = aws_apigatewayv2_portal_product_sharing_policy.example
  id = "abcdef1234"
}
```

Using `terraform import`, import API Gateway portal product sharing policies using `portal_product_id`. For example:

```console
% terraform import aws_apigatewayv2_portal_product_sharing_policy.example abcdef1234
```
