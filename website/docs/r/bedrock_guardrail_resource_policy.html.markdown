---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_guardrail_resource_policy"
description: |-
  Terraform resource for managing an AWS Bedrock Guardrail Resource Policy.
---
# Resource: aws_bedrock_guardrail_resource_policy

Manages a resource-based policy (RBP) on an Amazon Bedrock guardrail or system-defined guardrail profile.

Resource-based policies are required for organization-level enforced guardrails (`BEDROCK_POLICY` in AWS Organizations Service Control Policies). Member accounts must be granted `bedrock:ApplyGuardrail` on the guardrail (and, when Cross-Region Inference is enabled, on the associated guardrail profile) that lives in the management account.

## Example Usage

### Guardrail RBP for Organization Enforcement

```terraform
resource "aws_bedrock_guardrail" "example" {
  name                      = "org-guardrail"
  blocked_input_messaging   = "Blocked."
  blocked_outputs_messaging = "Blocked."

  content_policy_config {
    filters_config {
      type            = "VIOLENCE"
      input_strength  = "HIGH"
      output_strength = "HIGH"
    }
  }
}

data "aws_organizations_organization" "current" {}

resource "aws_bedrock_guardrail_resource_policy" "example" {
  resource_arn = aws_bedrock_guardrail.example.guardrail_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = ["bedrock:GetGuardrail", "bedrock:ApplyGuardrail"]
      Resource  = aws_bedrock_guardrail.example.guardrail_arn
      Condition = {
        StringEquals = { "aws:PrincipalOrgID" = data.aws_organizations_organization.current.id }
      }
    }]
  })
}
```

### Guardrail Profile RBP for Cross-Region Inference

```terraform
resource "aws_bedrock_guardrail_resource_policy" "profile" {
  resource_arn = var.guardrail_profile_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = ["bedrock:ApplyGuardrail"]
      Resource  = var.guardrail_profile_arn
      Condition = {
        StringEquals = { "aws:PrincipalOrgID" = data.aws_organizations_organization.current.id }
      }
    }]
  })
}
```

## Argument Reference

The following arguments are required:

* `resource_arn` - (Required, Forces new resource) ARN of the Bedrock resource to attach the policy to. This can be a guardrail ARN or a system-defined guardrail profile ARN.
* `policy` - (Required) JSON policy document. This is a resource-based policy granting other principals (e.g., all principals in an AWS Organization) permission to call Bedrock APIs on this resource.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes beyond the arguments above.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Guardrail Resource Policy using the `resource_arn`. For example:

```terraform
import {
  to = aws_bedrock_guardrail_resource_policy.example
  id = "arn:aws:bedrock:us-east-1:123456789012:guardrail/abc123"
}
```

Using `terraform import`, import Bedrock Guardrail Resource Policy using the `resource_arn`. For example:

```console
% terraform import aws_bedrock_guardrail_resource_policy.example arn:aws:bedrock:us-east-1:123456789012:guardrail/abc123
```
