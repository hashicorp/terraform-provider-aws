---
subcategory: "VPC Lattice"
layout: "aws"
page_title: "AWS: aws_vpclattice_resource_policy"
description: |-
  Terraform resource for managing an AWS VPC Lattice Resource Policy.
---

# Resource: aws_vpclattice_resource_policy

Terraform resource for managing an AWS VPC Lattice Resource Policy.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "current" {}

resource "aws_vpclattice_service_network" "example" {
  name = "example-vpclattice-service-network"
}

resource "aws_vpclattice_resource_policy" "example" {
  resource_arn  = aws_vpclattice_service_network.example.arn
  policy        = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Sid    = "example-policy-statement"
        Effect = "Allow"
        Principal = {
          "AWS" = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action = [
          "vpc-lattice:CreateServiceNetworkVpcAssociation",
          "vpc-lattice:CreateServiceNetworkServiceAssociation",
          "vpc-lattice:GetServiceNetwork"
        ]
        Resource = "${aws_vpclattice_service_network.example.arn}"
      }
    ]
  })
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) The ID or Amazon Resource Name (ARN) of the service network or service for which the policy is created.
* `policy` - (Required) An IAM policy. The policy string in JSON must not contain newlines or blank lines.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

VPC Lattice Resource Policy can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_vpclattice_resource_policy.example rft-8012925589
```
