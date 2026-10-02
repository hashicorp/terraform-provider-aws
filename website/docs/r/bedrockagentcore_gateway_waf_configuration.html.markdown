---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_gateway_waf_configuration"
description: |-
  Manages the AWS WAF configuration for an Amazon Bedrock AgentCore Gateway.
---

# Resource: aws_bedrockagentcore_gateway_waf_configuration

Manages the AWS WAF configuration for an Amazon Bedrock AgentCore Gateway.

AWS WAF protection for a gateway is enabled by associating an AWS WAF web ACL with the gateway ARN using [`aws_wafv2_web_acl_association`](wafv2_web_acl_association.html.markdown). This resource then sets the gateway's `failure_mode`, which controls how the gateway behaves when AWS WAF is unreachable.

~> **NOTE:** The gateway's WAF configuration can only be set once an AWS WAF web ACL is associated with the gateway. Use `depends_on` to ensure the [`aws_wafv2_web_acl_association`](wafv2_web_acl_association.html.markdown) is created before this resource. Removing this resource resets the gateway's failure mode.

## Example Usage

```terraform
resource "aws_bedrockagentcore_gateway" "example" {
  name     = "example"
  role_arn = aws_iam_role.example.arn

  authorizer_type = "AWS_IAM"
  protocol_type   = "MCP"
}

resource "aws_wafv2_web_acl" "example" {
  name  = "example"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "example"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl_association" "example" {
  resource_arn = aws_bedrockagentcore_gateway.example.gateway_arn
  web_acl_arn  = aws_wafv2_web_acl.example.arn
}

resource "aws_bedrockagentcore_gateway_waf_configuration" "example" {
  gateway_identifier = aws_bedrockagentcore_gateway.example.gateway_id
  failure_mode       = "FAIL_OPEN"

  depends_on = [aws_wafv2_web_acl_association.example]
}
```

## Argument Reference

The following arguments are required:

* `failure_mode` - (Required) Behavior when AWS WAF is unreachable or times out. Valid values: `FAIL_OPEN` (allow requests), `FAIL_CLOSE` (block requests).
* `gateway_identifier` - (Required) Identifier of the gateway to configure. Forces replacement if changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `web_acl_arn` - ARN of the AWS WAF web ACL associated with the gateway.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Amazon Bedrock AgentCore Gateway WAF Configuration using the `gateway_identifier`. For example:

```terraform
import {
  to = aws_bedrockagentcore_gateway_waf_configuration.example
  id = "gateway-abc123"
}
```

Using `terraform import`, import an Amazon Bedrock AgentCore Gateway WAF Configuration using the `gateway_identifier`. For example:

```console
% terraform import aws_bedrockagentcore_gateway_waf_configuration.example gateway-abc123
```
