---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_policy"
description: |-
  Provides a VPC Endpoint Policy resource.
---

# Resource: aws_vpc_endpoint_policy

Provides a VPC Endpoint Policy resource.

## Example Usage

```terraform
data "aws_vpc_endpoint_service" "example" {
  service = "dynamodb"
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_vpc_endpoint" "example" {
  service_name = data.aws_vpc_endpoint_service.example.service_name
  vpc_id       = aws_vpc.example.id
}

resource "aws_vpc_endpoint_policy" "example" {
  vpc_endpoint_id = aws_vpc_endpoint.example.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "AllowAll",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "*"
        },
        "Action" : [
          "dynamodb:*"
        ],
        "Resource" : "*"
      }
    ]
  })
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_endpoint_id` - (Required) The VPC Endpoint ID.
* `policy` - (Optional) A policy to attach to the endpoint that controls access to the service. Defaults to full access. All `Gateway` and some `Interface` endpoints support policies - see the [relevant AWS documentation](https://docs.aws.amazon.com/vpc/latest/userguide/vpc-endpoints-access.html) for more details. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/terraform/aws/iam-policy).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the VPC endpoint.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import VPC Endpoint Policies using the `id`. For example:

```terraform
import {
  to = aws_vpc_endpoint_policy.example
  id = "vpce-3ecf2a57"
}
```

Using `terraform import`, import VPC Endpoint Policies using the `id`. For example:

```console
% terraform import aws_vpc_endpoint_policy.example vpce-3ecf2a57
```
