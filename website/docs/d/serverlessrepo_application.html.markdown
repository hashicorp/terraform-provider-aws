---
layout: "aws"
page_title: "AWS: aws_serverlessrepo_application"
sidebar_current: "docs-aws-datasource-serverlessrepo-application"
description: |-
  Get information on a AWS Serverless Application Repository application
---

# Data Source: aws_serverlessrepo_application

Use this data source to get the source code URL and template URL of an AWS Serverless Application Repository application.

## Example Usage

```hcl
data "aws_serverlessrepo_application" "example" {
  application_id = "arn:aws:serverlessrepo:us-east-1:123456789012:applications/ExampleApplication"
}

resource "aws_cloudformation_stack" "example" {
  name = "Example"

  capabilities = ["CAPABILITY_NAMED_IAM"]

  template_url = data.aws_serverlessrepo_application.example.template_url
}

```

## Argument Reference

* `application_id` - (Required) The ARN of the application.
* `semantic_version` - (Optional) The requested version of the application. By default, uses the latest version.

## Attributes Reference

* `application_id` - The ARN of the application.
* `semantic_version` - The version of the application retrieved.
* `name` - The name of the application.
* `source_code_url` - A URL pointing to the source code of the application version.
* `template_url` - A URL pointing to the Cloud Formation template for the application version.
