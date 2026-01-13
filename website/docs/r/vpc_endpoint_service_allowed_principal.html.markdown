---
subcategory: "VPC (Virtual Private Cloud)"
layout: "aws"
page_title: "AWS: aws_vpc_endpoint_service_allowed_principal"
description: |-
  Provides a resource to allow a principal to discover a VPC endpoint service.
---

# Resource: aws_vpc_endpoint_service_allowed_principal

Provides a resource to allow a principal to discover a VPC endpoint service.

~> **NOTE on VPC Endpoint Services and VPC Endpoint Service Allowed Principals:** Terraform provides
both a standalone [VPC Endpoint Service Allowed Principal](vpc_endpoint_service_allowed_principal.html) resource
and a VPC Endpoint Service resource with an `allowed_principals` attribute. Do not use the same principal ARN in both
a VPC Endpoint Service resource and a VPC Endpoint Service Allowed Principal resource. Doing so will cause a conflict
and will overwrite the association.

## Example Usage

Basic usage:

```terraform
data "aws_caller_identity" "current" {}

resource "aws_vpc_endpoint_service_allowed_principal" "allow_me_to_foo" {
  vpc_endpoint_service_id = aws_vpc_endpoint_service.foo.id
  principal_arn           = data.aws_caller_identity.current.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `vpc_endpoint_service_id` - (Required) The ID of the VPC endpoint service to allow permission.
* `principal_arn` - (Required) The ARN of the principal to allow permissions.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The ID of the association.
