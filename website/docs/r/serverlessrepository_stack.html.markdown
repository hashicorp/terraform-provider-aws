---
layout: "aws"
page_title: "AWS: aws_serverlessrepository_stack"
sidebar_current: "docs-aws-resource-serverlessrepository-stack"
description: |-
  Provides a Serverless Application Repository Stack resource.
---

# Resource: aws_serverlessrepository_application

Provides a Serverless Application Repository CloudFormation Stack resource

## Example Usage

```hcl
resource "aws_serverlessrepository_application" "postgres-rotator" {
  name           = "postgres-rotator"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  parameters = {
    functionName = "func-postgres-rotator"
    endpoint     = "secretsmanager.us-west-2.amazonaws.com"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the stack to create. AWS prefixes this name with `serverlessrepo-`
* `application_id` - (Required) The ARN of the application from the Serverless Application Repository.
* `capabilities` - A list of capabilities.
  Valid values: `CAPABILITY_IAM`, `CAPABILITY_NAMED_IAM`, `CAPABILITY_RESOURCE_POLICY`, or `CAPABILITY_AUTO_EXPAND`
* `parameters` - (Optional) A map of Parameter structures that specify input parameters for the stack.
* `semantic_version` - (Optional) The version of the application to deploy. If not supplied, deploys the latest version.
* `tags` - (Optional) A list of tags to associate with this stack.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier of the stack.
* `outputs` - A map of outputs from the stack.

## Import

Import is not yet implemented.