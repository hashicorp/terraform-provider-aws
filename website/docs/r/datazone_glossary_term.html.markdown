---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_glossary_term"
description: |-
  Terraform resource for managing an AWS DataZone Glossary Term.
---
# Resource: aws_datazone_glossary_term

Terraform resource for managing an AWS DataZone Glossary Term.

## Example Usage

```terraform
resource "aws_iam_role" "example" {
  name = "example"
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
    name = "example"
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

resource "aws_datazone_domain" "example" {
  name                  = "example_name"
  domain_execution_role = aws_iam_role.example.arn
}

resource "aws_security_group" "example" {
  name = "example_name"
}

resource "aws_datazone_project" "example" {
  domain_identifier   = aws_datazone_domain.example.id
  glossary_terms      = ["2N8w6XJCwZf"]
  name                = "example"
  skip_deletion_check = true
}

resource "aws_datazone_glossary" "example" {
  description               = "description"
  name                      = "example"
  owning_project_identifier = aws_datazone_project.example.id
  status                    = "ENABLED"
  domain_identifier         = aws_datazone_project.example.domain_identifier
}

resource "aws_datazone_glossary_term" "example" {
  domain_identifier   = aws_datazone_domain.example.id
  glossary_identifier = aws_datazone_glossary.example.id
  name                = "example"
  status              = "ENABLED"
}
```

## Argument Reference

The following arguments are required:

* `domain_identifier` - (Required) Identifier of domain.
* `glossary_identifier` - (Required) Identifier of glossary.
* `name` - (Required) Name of glossary term.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `long_description` - (Optional) Long description of entry.
* `short_description` - (Optional) Short description of entry.
* `status` - (Optional) If glossary term is ENABLED or DISABLED.
* `term_relations` - (Optional) Object classifying the term relations through the following attributes:
    * `classifies` - (Optional) String array that calssifies the term relations.
    * `is_as` - (Optional) The isA property of the term relations.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Id of the glossary term.
* `created_at` - Time of glossary term creation.
* `created_by` - Creator of glossary term.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30s`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Glossary Term using a comma-delimited string combining the `domain_identifier`, `id`, and the `glossary_identifier`. For example:

```terraform
import {
  to = aws_datazone_glossary_term.example
  id = "domain_identifier,id,glossary_identifier"
}
```

Using `terraform import`, import DataZone Glossary Term using a comma-delimited string combining the `domain_identifier`, `id`, and the `glossary_identifier`. For example:

```console
% terraform import aws_datazone_glossary_term.example domain-id,glossary-term-id,glossary-id
```
