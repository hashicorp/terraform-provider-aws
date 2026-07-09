---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_guardrails"
description: |-
  Terraform data source for listing AWS Bedrock Guardrails.
---

# Data Source: aws_bedrock_guardrails

Terraform data source for listing AWS Bedrock Guardrails.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_guardrails" "example" {}
```

### Filter by Tag Using HCL

```terraform
data "aws_bedrock_guardrails" "example" {}

locals {
  prod_guardrails = [
    for g in data.aws_bedrock_guardrails.example.guardrails :
    g if lookup(g.tags, "environment", "") == "prod"
  ]
}
```

### List All Versions of a Specific Guardrail

```terraform
data "aws_bedrock_guardrails" "versions" {
  guardrail_identifier = aws_bedrock_guardrail.example.guardrail_id
}
```

## Argument Reference

This data source supports the following arguments:

* `guardrail_identifier` - (Optional) ID or ARN of a specific guardrail. When set, returns all versions (DRAFT and numbered versions) of that guardrail instead of the DRAFT version of all guardrails. Must be a lowercase alphanumeric guardrail ID or a full guardrail ARN in the format `arn:aws(-[^:]+)?:bedrock:{region}:{account-id}:guardrail/{id}`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS region.
* `guardrails` - List of guardrail summary objects. See [`guardrails`](#guardrails).

### `guardrails`

* `arn` - ARN of the guardrail.
* `created_at` - RFC 3339 creation timestamp.
* `description` - Description of the guardrail.
* `guardrail_id` - Unique identifier of the guardrail.
* `name` - Name of the guardrail.
* `status` - Status of the guardrail.
* `tags` - Map of resource tags.
* `updated_at` - RFC 3339 last-updated timestamp.
* `version` - Version identifier. `DRAFT` when no `guardrail_identifier` is set; `DRAFT` or a numbered version (e.g. `1`) when `guardrail_identifier` is set.
