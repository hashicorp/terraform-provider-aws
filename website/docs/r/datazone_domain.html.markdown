---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_domain"
description: |-
  Terraform resource for managing an AWS DataZone Domain.
---

# Resource: aws_datazone_domain

Terraform resource for managing an AWS DataZone Domain.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "domain_execution_role" {
  name = "my_domain_execution_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "datazone.amazonaws.com"
        }
      },
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = "domain_execution_policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          # Consider scoping down
          Action = [
            "datazone:*",
            "ram:*",
            "sso:*",
            "kms:*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}

resource "aws_datazone_domain" "example" {
  name                  = "example"
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the Domain.
* `domain_execution_role` - (Required) ARN of the role used by DataZone to configure the Domain.

The following arguments are optional:

* `description` - (Optional) Description of the Domain.
* `kms_key_identifier` - (Optional) ARN of the KMS key used to encrypt the Amazon DataZone domain, metadata and reporting data.
* `single_sign_on` - (Optional) Single sign on options, used to [enable AWS IAM Identity Center](https://docs.aws.amazon.com/datazone/latest/userguide/enable-IAM-identity-center-for-datazone.html) for DataZone.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Domain.
* `id` - ID of the Domain.
* `portal_url` - URL of the data portal for the Domain.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Domain using the `domain_id`. For example:

```terraform
import {
  to = aws_datazone_domain.example
  id = "domain-id-12345678"
}
```

Using `terraform import`, import DataZone Domain using the `domain_id`. For example:

```console
% terraform import aws_datazone_domain.example domain-id-12345678
```
