---
subcategory: "Serverless Application Repository"
layout: "aws"
page_title: "AWS: aws_serverlessapplicationrepository_cloudformation_stack"
description: |-
  Manages an Application CloudFormation Stack from the Serverless Application Repository.
---

# Resource: aws_serverlessapplicationrepository_cloudformation_stack

Manages an Application CloudFormation Stack from the Serverless Application Repository.

!> **Warning:** CloudFormation masks `NoEcho` parameter values as `****` in API responses, which may set an expectation that they remain hidden. They do not — like any other argument, the configured value is persisted to state. To mask a specific parameter in plan and `terraform show` output, wrap it with Terraform's [`sensitive()`](https://developer.hashicorp.com/terraform/language/functions/sensitive) function, for example `parameters = { password = sensitive(var.password) }`.

## Example Usage

### Basic Usage

```terraform
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "postgres-rotator"
  application_id = "arn:aws:serverlessrepo:us-east-1:297356227824:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-postgres-rotator"
    endpoint     = "secretsmanager.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
```

## Argument Reference

This resource supports the following arguments:

* `application_id` - (Required) ARN of the application from the Serverless Application Repository.
* `capabilities` - (Optional) List of capabilities. Valid values are `CAPABILITY_IAM`, `CAPABILITY_NAMED_IAM`, `CAPABILITY_RESOURCE_POLICY`, or `CAPABILITY_AUTO_EXPAND`. If the application contains IAM resources, IAM resources with custom names, resource-based policies, or nested applications, the corresponding capability must be specified. If omitted, the value applied by AWS is tracked in state.
* `name` - (Required) Name of the stack to create. The resource deployed in AWS will be prefixed with `serverlessrepo-`
* `parameters` - (Optional) Map of Parameter structures that specify input parameters for the stack.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `semantic_version` - (Optional) Version of the application to deploy. If not supplied, deploys the latest version.
* `tags` - (Optional) Map of tags assigned to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Unique identifier of the stack.
* `outputs` - Map of outputs from the stack.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Serverless Application Repository Stack using the CloudFormation Stack name (with or without the `serverlessrepo-` prefix) or the CloudFormation Stack ID. For example:

```terraform
import {
  to = aws_serverlessapplicationrepository_cloudformation_stack.example
  id = "serverlessrepo-postgres-rotator"
}
```

Using `terraform import`, import Serverless Application Repository Stack using the CloudFormation Stack name (with or without the `serverlessrepo-` prefix) or the CloudFormation Stack ID. For example:

```console
% terraform import aws_serverlessapplicationrepository_cloudformation_stack.example serverlessrepo-postgres-rotator
```
