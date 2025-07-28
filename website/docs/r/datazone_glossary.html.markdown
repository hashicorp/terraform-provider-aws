---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_glossary"
description: |-
  Terraform resource for managing an AWS DataZone Glossary.
---
# Resource: aws_datazone_glossary

Terraform resource for managing an AWS DataZone Glossary.

## Example Usage

```terraform

resource "aws_iam_role" "domain_execution_role" {
  name = "example_name"
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
    name = "example_name"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
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

resource "aws_datazone_domain" "test" {
  name                  = "example_name"
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}

resource "aws_security_group" "test" {
  name = "example_name"
}

resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_terms      = ["2N8w6XJCwZf"]
  name                = "example_name"
  description         = "desc"
  skip_deletion_check = true
}


resource "aws_datazone_glossary" "test" {
  description               = "description"
  name                      = "example_name"
  owning_project_identifier = aws_datazone_project.test.id
  status                    = "DISABLED"
  domain_identifier         = aws_datazone_project.test.domain_identifier
}
```

### Basic Usage

```terraform
resource "aws_datazone_glossary" "test" {
  description               = "description"
  name                      = "example_name"
  owning_project_identifier = aws_datazone_project.test.id
  status                    = "DISABLED"
  domain_identifier         = aws_datazone_project.test.domain_identifier
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the glossary. Must have length between 1 and 256.
* `owning_project_identifier` - (Required) ID of the project that owns business glossary. Must follow regex of ^[a-zA-Z0-9_-]{1,36}$.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `description` - (Optional) Description of the glossary. Must have a length between 0 and 4096.
* `status` - (Optional) Status of business glossary. Valid values are DISABLED and ENABLED.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Id of the Glossary.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Glossary using the `example_id_arg`. For example:

```terraform
import {
  to = aws_datazone_glossary.example
  id = "domain-id,glossary-id,owning_project_identifier"
}
```

Using `terraform import`, import DataZone Glossary using the import Datazone Glossary using a comma-delimited string combining the domain id, glossary id, and the id of the project it's under. For example:

```console
% terraform import aws_datazone_glossary.example domain-id,glossary-id,owning-project-identifier
```
