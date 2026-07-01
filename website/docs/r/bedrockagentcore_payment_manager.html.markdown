---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_payment_manager"
description: |-
  Manages an AWS Bedrock AgentCore Payment Manager.
---

# Resource: aws_bedrockagentcore_payment_manager

Manages an AWS Bedrock AgentCore Payment Manager. A Payment Manager governs how agents authenticate and authorize payment operations through AgentCore.

## Example Usage

### AWS IAM Authorization

```terraform
data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "example" {
  name               = "example-payment-manager"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_bedrockagentcore_payment_manager" "example" {
  name            = "example-payment-manager"
  authorizer_type = "AWS_IAM"
  role_arn        = aws_iam_role.example.arn
}
```

### Custom JWT Authorization

```terraform
resource "aws_bedrockagentcore_payment_manager" "example" {
  name            = "example-payment-manager"
  authorizer_type = "CUSTOM_JWT"
  role_arn        = aws_iam_role.example.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://example.com/.well-known/openid-configuration"
      allowed_audience = ["example-audience"]
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `authorizer_type` - (Required) Type of authorizer used by the payment manager. Valid values: `AWS_IAM`, `CUSTOM_JWT`.
* `name` - (Required, Forces new resource) Name of the payment manager.
* `role_arn` - (Required) ARN of the IAM role that the payment manager assumes.

The following arguments are optional:

* `authorizer_configuration` - (Optional) Configuration for the authorizer. Required when `authorizer_type` is `CUSTOM_JWT`. See [`authorizer_configuration`](#authorizer_configuration) below.
* `description` - (Optional) Description of the payment manager.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `authorizer_configuration`

The `authorizer_configuration` block supports the following:

* `custom_jwt_authorizer` - (Optional) Custom JWT authorizer configuration. See [`custom_jwt_authorizer`](#custom_jwt_authorizer) below.

### `custom_jwt_authorizer`

The `custom_jwt_authorizer` block supports the following:

* `discovery_url` - (Required) OpenID Connect discovery URL for the JWT authorizer.
* `allowed_audience` - (Optional) Set of allowed audiences for the JWT.
* `allowed_clients` - (Optional) Set of allowed client identifiers for the JWT.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `payment_manager_arn` - ARN of the Payment Manager.
* `payment_manager_id` - Unique identifier of the Payment Manager.
* `status` - Status of the Payment Manager.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `workload_identity_details` - Workload identity details for the Payment Manager.
    * `workload_identity_arn` - ARN of the workload identity.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_payment_manager.example
  identity = {
    payment_manager_id = "payment-manager-id-12345678"
  }
}

resource "aws_bedrockagentcore_payment_manager" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `payment_manager_id` (String) Payment manager ID.

#### Optional

* `account_id` (String) AWS account ID for this resource.
* `region` (String) AWS Region for this resource.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Bedrock AgentCore Payment Manager by payment manager ID. For example:

```terraform
import {
  to = aws_bedrockagentcore_payment_manager.example
  id = "payment-manager-id-12345678"
}
```

Using `terraform import`, import a Bedrock AgentCore Payment Manager by payment manager ID. For example:

```console
% terraform import aws_bedrockagentcore_payment_manager.example payment-manager-id-12345678
```
