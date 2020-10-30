---
subcategory: "Serverless Application Repository"
layout: "aws"
page_title: "AWS: aws_serverlessrepository_application"
description: |-
  Get information on a AWS Serverless Application Repository application
---

# Data Source: aws_serverlessrepository_application

Use this data source to get the source code URL and template URL of an AWS Serverless Application Repository application.

## Example Usage

```hcl
data "aws_serverlessrepository_application" "example" {
  application_id = "arn:aws:serverlessrepo:us-east-1:123456789012:applications/ExampleApplication"
}

resource "aws_cloudformation_stack" "example" {
  name = "Example"

  capabilities = data.aws_serverlessrepository_application.example.required_capabilities

  template_url = data.aws_serverlessrepository_application.example.template_url
}
```

## Argument Reference

* `application_id` - (Required) The ARN of the application.
* `semantic_version` - (Optional) The requested version of the application. By default, retrieves the latest version.

## Attributes Reference

* `application_id` - The ARN of the application.
* `semantic_version` - The version of the application retrieved.
* `name` - The name of the application.
* `required_capabilities` - A list of capabilities describing the permissions needed to deploy the application.
* `source_code_url` - A URL pointing to the source code of the application version.
* `template_url` - A URL pointing to the Cloud Formation template for the application version.
