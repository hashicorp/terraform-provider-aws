---
subcategory: "RAM (Resource Access Manager)"
layout: "aws"
page_title: "AWS: aws_ram_resource_share_association_exclusive"
description: |-
  Terraform resource for maintaining exclusive management of principal and resource associations for an AWS RAM (Resource Access Manager) Resource Share.
---

# Resource: aws_ram_resource_share_association_exclusive

Terraform resource for maintaining exclusive management of principal and resource associations for an AWS RAM (Resource Access Manager) Resource Share.

!> This resource takes exclusive ownership over principal and resource associations for a resource share. This includes removal of principals and resources which are not explicitly configured.

~> Destruction of this resource will disassociate all configured principals and resources from the resource share.

~> **NOTE:** This resource cannot be used in conjunction with [`aws_ram_principal_association`](/docs/providers/aws/r/ram_principal_association.html) or [`aws_ram_resource_association`](/docs/providers/aws/r/ram_resource_association.html) for the same resource share. Using them together will cause persistent drift and conflicts.

## Example Usage

### Basic Usage with Principals

```terraform
resource "aws_ram_resource_share" "example" {
  name                      = "example"
  allow_external_principals = true
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  vpc_id     = aws_vpc.example.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_ram_resource_share_association_exclusive" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn

  principals = [
    "111111111111",
    "222222222222",
  ]

  resource_arns = [
    aws_subnet.example.arn,
  ]
}
```

### With Organization Principal

```terraform
resource "aws_ram_resource_share" "example" {
  name = "example"
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "example" {
  count      = 2
  vpc_id     = aws_vpc.example.id
  cidr_block = cidrsubnet(aws_vpc.example.cidr_block, 8, count.index)
}

resource "aws_ram_resource_share_association_exclusive" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn

  principals = [
    aws_organizations_organization.example.arn,
  ]

  resource_arns = aws_subnet.example[*].arn
}
```

### With Service Principals

When sharing resources with AWS services, use service principals. Service principals follow the pattern `service-id.amazonaws.com` (e.g., `pca-connector-ad.amazonaws.com`, `elasticmapreduce.amazonaws.com`). The `sources` argument can be used to restrict which AWS accounts the service can access the shared resources from.

~> **NOTE:** Service principals cannot be mixed with other principal types (AWS account IDs, organization ARNs, OU ARNs, IAM role ARNs, or IAM user ARNs) in the same resource.

```terraform
resource "aws_ram_resource_share" "example" {
  name                      = "example-service-share"
  allow_external_principals = true
}

resource "aws_acmpca_certificate_authority" "example" {
  type = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "example.com"
    }
  }
}

resource "aws_ram_resource_share_association_exclusive" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn

  principals = [
    "pca-connector-ad.amazonaws.com",
  ]

  resource_arns = [
    aws_acmpca_certificate_authority.example.arn,
  ]

  sources = [
    "111111111111",
    "222222222222",
  ]
}
```

### Disallow All Associations

To automatically remove any configured associations, omit the `principals` and `resource_arns` arguments or set them to empty lists.

~> This will not **prevent** associations from being created via Terraform (or any other interface). This resource enables bringing associations into a configured state, however, this reconciliation happens only when `apply` is proactively run.

```terraform
resource "aws_ram_resource_share_association_exclusive" "example" {
  resource_share_arn = aws_ram_resource_share.example.arn
}
```

## Argument Reference

The following arguments are required:

* `resource_share_arn` - (Required) The Amazon Resource Name (ARN) of the resource share. Changing this value forces creation of a new resource.

The following arguments are optional:

* `principals` - (Optional) A set of principals to associate with the resource share. Principals not configured in this argument will be removed. Valid values include:
    * AWS account ID (exactly 12 digits, e.g., `123456789012`)
    * AWS Organizations Organization ARN (e.g., `arn:aws:organizations::123456789012:organization/o-exampleorgid`)
    * AWS Organizations Organizational Unit ARN (e.g., `arn:aws:organizations::123456789012:ou/o-exampleorgid/ou-examplerootid-exampleouid`)
    * IAM role ARN (e.g., `arn:aws:iam::123456789012:role/example-role`)
    * IAM user ARN (e.g., `arn:aws:iam::123456789012:user/example-user`)
    * Service principal (e.g., `ec2.amazonaws.com`)
* `resource_arns` - (Optional) A set of Amazon Resource Names (ARNs) of resources to associate with the resource share. Resources not configured in this argument will be removed.
* `sources` - (Optional) A set of AWS account IDs that restrict which accounts a service principal can access resources from. This argument can only be specified when `principals` contains only service principals. When specified, it limits the source accounts from which the service can access the shared resources.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Amazon Resource Name (ARN) of the resource share (same as `resource_share_arn`).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RAM Resource Share Association Exclusive using the `resource_share_arn`. For example:

```terraform
import {
  to = aws_ram_resource_share_association_exclusive.example
  id = "arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12"
}
```

Using `terraform import`, import RAM Resource Share Association Exclusive using the `resource_share_arn`. For example:

```console
% terraform import aws_ram_resource_share_association_exclusive.example arn:aws:ram:eu-west-1:123456789012:resource-share/73da1ab9-b94a-4ba3-8eb4-45917f7f4b12
```
