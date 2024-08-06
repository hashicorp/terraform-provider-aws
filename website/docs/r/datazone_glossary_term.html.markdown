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
resource "aws_datazone_glossary_term" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_identifier = aws_datazone_glossary.test.id
  long_description    = "long_description"
  name                = "example-name"
  short_description   = "short_desc"
  status              = "ENABLED"
  term_relations {
    classifies = ["id of other glossary term"]
	  is_a = ["id of other glossary term"]
  }
}
```

### Basic Usage

```terraform
resource "aws_datazone_glossary_term" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_identifier = aws_datazone_glossary.test.id
  long_description    = "long_description"
  name                = "example-name"
  short_description   = "short_desc"
  status              = "ENABLED"
  term_relations {
    classifies = ["id of other glossary term"]
	  is_a       = ["id of other glossary term"]
  }
}
```

## Argument Reference

The following arguments are required:

* `glossary_identifier` - (Required) Identifier of glossary.
* `domain_identifier` - (Required) Identifier of domain.
* `name` - (Required) Name of glossary term./

The following arguments are optional:

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
* `updated_at` - Time of glossary term update.
* `updated_by` - Updater of glossary term.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Glossary Term using a comma-delimited string combining the domain id, glossary term id, and the glossary id. For example:

```terraform
import {
  to = aws_datazone_glossary_term.example
  id = "domain-id,glossary-term-id,glossary-id"
}
```

Using `terraform import`, import DataZone Glossary Term using a comma-delimited string combining the domain id, glossary term id, and the glossary id. For example:

```console
% terraform import aws_datazone_glossary_term.example domain-id,glossary-term-id,glossary-id
```
