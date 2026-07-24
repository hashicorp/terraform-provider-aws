---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_delivery_source"
description: |-
  Terraform resource for managing an AWS CloudWatch Logs Delivery Source.
---

# Resource: aws_cloudwatch_log_delivery_source

Terraform resource for managing an AWS CloudWatch Logs Delivery Source.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_delivery_source" "example" {
  name         = "example"
  log_type     = "APPLICATION_LOGS"
  resource_arn = aws_bedrockagent_knowledge_base.example.arn
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `log_type` - (Required) The type of log that the source is sending. For Amazon Bedrock, the valid value is `APPLICATION_LOGS`. For Amazon Bedrock AgentCore, the valid values are `APPLICATION_LOGS`, `TRACES`, and `USAGE_LOGS`. For Amazon CodeWhisperer, the valid value is `EVENT_LOGS`. For IAM Identity Center, the valid value is `ERROR_LOGS`. For Amazon WorkMail, the valid values are `ACCESS_CONTROL_LOGS`, `AUTHENTICATION_LOGS`, `WORKMAIL_AVAILABILITY_PROVIDER_LOGS`, and `WORKMAIL_MAILBOX_ACCESS_LOGS`. See [Enable logging from AWS services](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/AWS-logs-and-resource-policy.html) for the log types supported by each service.
* `name` - (Required) The name for this delivery source.
* `resource_arn` - (Required) The ARN of the AWS resource that is generating and sending logs.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the delivery source.
* `service` - The AWS service that is sending logs.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_log_delivery_source.example
  identity = {
    name = "example"
  }
}

resource "aws_cloudwatch_log_delivery_source" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) Name of the delivery source.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Delivery Sources using `name`. For example:

```terraform
import {
  to = aws_cloudwatch_log_delivery_source.example
  id = "example"
}
```

Using `terraform import`, import Delivery Sources using `name`. For example:

```console
% terraform import aws_cloudwatch_log_delivery_source.example example
```
