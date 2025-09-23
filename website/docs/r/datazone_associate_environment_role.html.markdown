---
subcategory: "DataZone"
layout: "aws"
page_title: "AWS: aws_datazone_associate_environment_role"
description: |-
  Terraform resource for associates the environment role in Amazon DataZone.
---

# Resource: aws_datazone_associate_environment_role

Terraform resource for associates the environment role in Amazon DataZone.

## Example Usage

### Basic Usage

```terraform

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

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
}

resource "aws_iam_role_policy" "domain_execution_role" {
  role = aws_iam_role.domain_execution_role.name
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

resource "aws_datazone_domain" "example" {
  name                  = "example"
  domain_execution_role = aws_iam_role.domain_execution_role.arn

  skip_deletion_check = true
}

resource "aws_datazone_project" "sample" {
  domain_identifier   = aws_datazone_domain.example.id
  name                = "sample"
  skip_deletion_check = true
}

data "aws_datazone_environment_blueprint" "sample" {
  domain_id = aws_datazone_domain.example.id
  name      = "CustomAwsService"
  managed   = true
}

resource "aws_datazone_environment_blueprint_configuration" "example" {
  domain_id                = aws_datazone_domain.example.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.example.id

  enabled_regions          = [data.aws_region.current.region]
}

resource "aws_datazone_environment_profile" "example" {
  aws_account_id                   = data.aws_caller_identity.current.account_id
  aws_account_region               = data.aws_region.current.region
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.example.id
  name                             = "example"
  project_identifier               = aws_datazone_project.example.id
  domain_identifier                = aws_datazone_domain.example.id
}

resource "aws_datazone_environment" "example" {
  name                 = "example"

  account_identifier   = data.aws_caller_identity.example.account_id
  account_region       = data.aws_region.example.region

  profile_identifier   = aws_datazone_environment_profile.example.id
  project_identifier   = aws_datazone_project.example.id
  domain_identifier    = aws_datazone_domain.example.id
}

data "aws_iam_policy_document" "blueprint_role" {
  statement {
    actions = [
      "sts:AssumeRole",
      "sts:TagSession"
    ]
    principals {
      type        = "Service"
      identifiers = ["datazone.amazonaws.com", "emr-serverless.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "environment_role" {
  name = "DataZone_Environment_Example"

  assume_role_policy = data.aws_iam_policy_document.blueprint_role.json
}


resource "aws_datazone_associate_environment_role" "example" {
  domain_identifier      = aws_datazone_domain.example.id
  environment_identifier = aws_datazone_environment.example.id
  environment_role_arn = aws_iam_role.environment_role.arn
}

```

## Argument Reference

The following arguments are required:

* `domain_identifier` - (Required) The ID of the Amazon DataZone domain in which the environment role is associated.
* `environment_identifier` - (Required) The ID of the Amazon DataZone environment.
* `environment_role_arn` - (Required) The ARN of the environment role.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `delete` - (Default `10m`)
