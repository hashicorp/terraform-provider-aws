---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_form_type"
description: |-
  Terraform resource for managing an AWS DataZone Form Type.
---

# Resource: aws_datazone_form_type

Terraform resource for managing an AWS DataZone Form Type.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role" "domain_execution_role" {
  name = "example-role"
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
    name = "example-policy"
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
  name                  = "example"
  domain_execution_role = aws_iam_role.domain_execution_role.arn
}

resource "aws_security_group" "test" {
  name = "example"
}

resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  glossary_terms      = ["2N8w6XJCwZf"]
  name                = "example name"
  description         = "desc"
  skip_deletion_check = true
}

resource "aws_datazone_form_type" "test" {
  description               = "desc"
  name                      = "SageMakerModelFormType"
  domain_identifier         = aws_datazone_domain.test.id
  owning_project_identifier = aws_datazone_project.test.id
  status                    = "DISABLED"
  model {
    smithy = <<EOF
	structure SageMakerModelFormType {
			@required
			@amazon.datazone#searchable
			modelName: String

			@required
			modelArn: String

			@required
			creationTime: String
			}
		EOF
  }
}
```

## Argument Reference

The following arguments are required:

* `domain_identifier` - (Required) Identifier of the domain.
* `name` - (Required) Name of the form type. Must be the name of the structure in smithy document.
* `owning_project_identifier` - (Required) Identifier of project that owns the form type. Must follow regex of ^[a-zA-Z0-9_-]{1,36}.
* `model` - (Required) Object of the model of the form type that contains the following attributes.
    * `smithy` - (Required) Smithy document that indicates the model of the API. Must be between the lengths 1 and 100,000 and be encoded as a smithy document.

The following arguments are optional:

* `description` - (Optional) Description of form type. Must have a length of between 1 and 2048 characters.
* `status` - (Optional) Status of form type. Must be "ENABLED" or "DISABLED" If status is set to "ENABLED" terraform cannot delete the resource until it is manually changed in the AWS console.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Creation time of the Form Type.
* `created_by` - Creator of the Form Type.
* `origin_domain_id` - Origin domain id of the Form Type.
* `origin_project_id` - Origin project id of the Form Type.
* `owning_project_id` - Owning project id of the Form Type.
* `revision` - Revision of the Form Type.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DataZone Form Type using a comma separated value of DomainIdentifier,Name,Revision. For example:

```terraform
import {
  to = aws_datazone_form_type.example
  id = "domain_identifier,name,revision"
}
```

Using `terraform import`, import DataZone Form Type using a comma separated value of `domain_identifier`,`name`,`revision`. For example:

```console
% terraform import aws_datazone_form_type.example domain_identifier,name,revision
```
