# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_apigatewayv2_portal_product_sharing_policy" "test" {
  portal_product_id = aws_apigatewayv2_portal_product.test.portal_product_id

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
        aws_apigatewayv2_portal_product.test.portal_product_arn,
        "${aws_apigatewayv2_portal_product.test.portal_product_arn}/*",
      ]
      Condition = {
        # Required alongside an organization condition: the policy must not share within the owning account.
        StringNotEquals = {
          "aws:PrincipalAccount" = data.aws_caller_identity.current.account_id
        }
        "ForAnyValue:StringLike" = {
          "aws:PrincipalOrgPaths" = ["${data.aws_organizations_organization.current.id}/${data.aws_organizations_organization.current.roots[0].id}/*"]
        }
      }
    }]
  })
}

resource "aws_apigatewayv2_portal_product" "test" {
  display_name = var.rName
}

data "aws_caller_identity" "current" {}

data "aws_organizations_organization" "current" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
