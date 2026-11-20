---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_guardrail_resource_policy"
description: |-
  Terraform resource for managing an AWS Bedrock Guardrail Resource Policy.
---
# Resource: aws_bedrock_guardrail_resource_policy

Manages a resource-based policy (RBP) on an Amazon Bedrock guardrail or system-defined guardrail profile.

Resource-based policies are required for organization-level enforced guardrails (an AWS Organizations policy of type `BEDROCK_POLICY`). Member accounts must be granted `bedrock:ApplyGuardrail` on the guardrail (and, when Cross-Region Inference is enabled, on the associated guardrail profile) that lives in the management account.

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

  cross_region_config {
    guardrail_profile_identifier = "arn:aws:bedrock:us-east-1:123456789012:guardrail-profile/us.guardrail.v1:0"
  }
}

data "aws_organizations_organization" "current" {}

resource "aws_bedrock_guardrail_resource_policy" "profile" {
  resource_arn = aws_bedrock_guardrail.example.cross_region_config[0].guardrail_profile_identifier

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = ["bedrock:ApplyGuardrail"]
      Resource  = aws_bedrock_guardrail.example.cross_region_config[0].guardrail_profile_identifier
      Condition = {
        StringEquals = { "aws:PrincipalOrgID" = data.aws_organizations_organization.current.id }
      }
    }]
  })
}
```

## Argument Reference

The following arguments are required:

* `policy` - (Required) JSON policy document. This is a resource-based policy granting other principals (e.g., all principals in an AWS Organization) permission to call Bedrock APIs on this resource.
* `resource_arn` - (Required, Forces new resource) ARN of the Bedrock resource to attach the policy to. This can be a guardrail ARN or a system-defined guardrail profile ARN.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrock_guardrail_resource_policy.example
  identity = {
    resource_arn = "arn:aws:bedrock:us-east-1:123456789012:guardrail/abc123"
  }
}

resource "aws_bedrock_guardrail_resource_policy" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `resource_arn` (String) ARN of the guardrail or guardrail profile to which the resource policy is attached.

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
