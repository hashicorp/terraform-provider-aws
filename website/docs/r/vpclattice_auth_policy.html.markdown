---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_auth_policy"
description: |-
  Terraform resource for managing an AWS VPC Lattice Auth Policy.
---

# Resource: aws_vpclattice_auth_policy

Terraform resource for managing an AWS VPC Lattice Auth Policy.

## Example Usage

### Basic Usage

```terraform
resource "aws_vpclattice_service" "example" {
  name               = "example-vpclattice-service"
  auth_type          = "AWS_IAM"
  custom_domain_name = "example.com"
}

resource "aws_vpclattice_auth_policy" "example" {
  resource_identifier = aws_vpclattice_service.example.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action    = "*"
        Effect    = "Allow"
        Principal = "*"
        Resource  = "*"
        Condition = {
          StringNotEqualsIgnoreCase = {
            "aws:PrincipalType" = "anonymous"
          }
        }
      }
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `resource_identifier` - (Required) The ID or Amazon Resource Name (ARN) of the service network or service for which the policy is created.
* `policy` - (Required) The auth policy. The policy string in JSON must not contain newlines or blank lines.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `policy` - The auth policy. The policy string in JSON must not contain newlines or blank lines.
* `state` - The state of the auth policy. The auth policy is only active when the auth type is set to AWS_IAM. If you provide a policy, then authentication and authorization decisions are made based on this policy and the client's IAM policy. If the Auth type is NONE, then, any auth policy you provide will remain inactive.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Lattice Auth Policy using the `example_id_arg`. For example:

```terraform
import {
  to = aws_vpclattice_auth_policy.example
  id = "rft-8012925589"
}
```

Using `terraform import`, import VPC Lattice Auth Policy using the `example_id_arg`. For example:

```console
% terraform import aws_vpclattice_auth_policy.example rft-8012925589
```
