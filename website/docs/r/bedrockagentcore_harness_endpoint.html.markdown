---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_harness_endpoint"
description: |-
  Manages an AWS Bedrock AgentCore Harness Endpoint.
---

# Resource: aws_bedrockagentcore_harness_endpoint

Manages an AWS Bedrock AgentCore Harness Endpoint. Harness Endpoints provide a versioned, network-accessible interface for a harness, enabling external systems to invoke a specific version of the harness.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_harness_endpoint" "example" {
  name       = "example_endpoint"
  harness_id = aws_bedrockagentcore_harness.example.harness_id
}
```

### With a Pinned Target Version

```terraform
resource "aws_bedrockagentcore_harness_endpoint" "example" {
  name           = "example_endpoint"
  harness_id     = aws_bedrockagentcore_harness.example.harness_id
  target_version = "1"
  description    = "Endpoint pinned to version 1"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the harness endpoint. Must begin with a letter and can contain letters, numbers, and underscores (up to 48 characters).
* `harness_id` - (Required) ID of the harness this endpoint belongs to.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `target_version` - (Optional) Version of the harness the endpoint should target. If omitted, the service resolves the version and exports it as a computed value.
* `description` - (Optional) Description of the harness endpoint.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `harness_endpoint_arn` - ARN of the Harness Endpoint.
* `live_version` - Version of the harness the endpoint is currently serving.
* `status` - Current status of the harness endpoint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore Harness Endpoint using the `harness_id` and `name` separated by a comma. For example:

```terraform
import {
  to = aws_bedrockagentcore_harness_endpoint.example
  id = "harness_example-abcde12345,example_endpoint"
}
```

Using `terraform import`, import Bedrock AgentCore Harness Endpoint using the `harness_id` and `name` separated by a comma. For example:

```console
% terraform import aws_bedrockagentcore_harness_endpoint.example harness_example-abcde12345,example_endpoint
```
