---
subcategory: "CloudWatch"
layout: "aws"
page_title: "AWS: aws_cloudwatch_contributor_managed_insight_rule"
description: |-
  Terraform resource for managing an AWS CloudWatch Contributor Managed Insight Rule.
---

# Resource: aws_cloudwatch_contributor_managed_insight_rule

Terraform resource for managing an AWS CloudWatch Contributor Managed Insight Rule.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_contributor_managed_insight_rule" "example" {
  resource_arn  = aws_vpc_endpoint_service.test.arn
  template_name = "VpcEndpointService-BytesByEndpointId-v1"
  rule_state    = "DISABLED"
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required) ARN of an Amazon Web Services resource that has managed Contributor Insights rules.
* `template_name` - (Required) Template name for the managed Contributor Insights rule, as returned by ListManagedInsightRules.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `rule_state` - (Optional) State of the rule. Valid values are `ENABLED` and `DISABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Contributor Managed Insight Rule.
* `rule_name` - Name of the Contributor Insights rule that contains data for the specified AWS resource.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatch_contributor_managed_insight_rule.example
  identity = {
    resource_arn  = "arn:aws:ec2:us-east-1:123456789012:vpc-endpoint-service/vpce-svc-0123456789abcdef0"
    template_name = "VpcEndpointService-BytesByEndpointId-v1"
  }
}

resource "aws_cloudwatch_contributor_managed_insight_rule" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` (String) ARN of the resource.
* `template_name` (String) Name of the template.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Contributor Managed Insight Rules using `resource_arn` and `template_name` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_cloudwatch_contributor_managed_insight_rule.example
  id = "arn:aws:ec2:us-east-1:123456789012:vpc-endpoint-service/vpce-svc-0123456789abcdef0,VpcEndpointService-BytesByEndpointId-v1"
}
```

Using `terraform import`, import Contributor Managed Insight Rules using `resource_arn` and `template_name` separated by a comma (`,`). For example:

```console
% terraform import aws_cloudwatch_contributor_managed_insight_rule.example arn:aws:ec2:us-east-1:123456789012:vpc-endpoint-service/vpce-svc-0123456789abcdef0,VpcEndpointService-BytesByEndpointId-v1
```
