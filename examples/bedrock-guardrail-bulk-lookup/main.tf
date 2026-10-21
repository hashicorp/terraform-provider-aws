terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# Map of logical names to existing guardrail ARNs (or bare guardrail IDs).
# Populate this with the guardrails you want to inspect.
locals {
  guardrail_arns = {
    "GR-001" = "arn:aws:bedrock:us-east-1:123456789012:guardrail/aaa111bbb222"
    "GR-002" = "arn:aws:bedrock:us-east-1:123456789012:guardrail/ccc333ddd444"
    "GR-003" = "arn:aws:bedrock:us-east-1:123456789012:guardrail/eee555fff666"
    # Add more guardrail ARNs here ...
  }
}

# Read the latest published version of each guardrail in parallel.
data "aws_bedrock_guardrail" "all" {
  for_each             = local.guardrail_arns
  guardrail_identifier = each.value
  latest               = true
}

output "guardrails" {
  description = "Summary of each guardrail's latest published version."
  value = {
    for k, v in data.aws_bedrock_guardrail.all : k => {
      guardrail_id            = v.guardrail_id
      name                    = v.name
      version                 = v.version
      status                  = v.status
      arn                     = v.arn
      has_cross_region_config = length(v.cross_region_config) > 0
    }
  }
}
