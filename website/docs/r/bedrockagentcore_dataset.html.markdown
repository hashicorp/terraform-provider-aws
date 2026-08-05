---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_dataset"
description: |-
  Manages an AWS Bedrock AgentCore Dataset.
---

# Resource: aws_bedrockagentcore_dataset

Manages an AWS Bedrock AgentCore Dataset. A dataset holds evaluation examples used by AgentCore evaluation workflows. The initial examples are supplied at creation time through the `source` block, either inline or from a JSONL file in Amazon S3.

## Example Usage

### Inline Examples

```terraform
resource "aws_bedrockagentcore_dataset" "example" {
  name        = "example_dataset"
  schema_type = "AGENTCORE_EVALUATION_PREDEFINED_V1"

  source {
    inline_examples {
      examples = [
        jsonencode({
          scenario_id = "scenario-1"
          turns = [
            {
              input             = "What is the capital of France?"
              expected_response = "Paris"
            }
          ]
        })
      ]
    }
  }
}
```

### Examples from Amazon S3

```terraform
resource "aws_s3_object" "examples" {
  bucket  = aws_s3_bucket.example.id
  key     = "examples.jsonl"
  content = jsonencode({ scenario_id = "s1", turns = [{ input = "hi", expected_response = "ok" }] })
}

resource "aws_bedrockagentcore_dataset" "example" {
  name        = "example_dataset"
  schema_type = "AGENTCORE_EVALUATION_PREDEFINED_V1"

  source {
    s3_source {
      s3_uri = "s3://${aws_s3_bucket.example.id}/${aws_s3_object.examples.key}"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required, Forces new resource) Name of the dataset. Must be unique within the account. Between 1 and 48 characters, beginning with a letter and containing only alphanumeric characters and underscores.
* `schema_type` - (Required, Forces new resource) Versioned schema type governing the structure of examples. Valid values are `AGENTCORE_EVALUATION_PREDEFINED_V1` (pre-written inputs per conversation turn) and `AGENTCORE_EVALUATION_SIMULATED_V1` (scenarios used to generate full conversations).
* `source` - (Required, Forces new resource) Source of the initial examples. See [`source`](#source) below.

The following arguments are optional:

* `description` - (Optional) Description of the dataset. Because the service treats an omitted description as "leave unchanged", removing this argument retains the previously configured value rather than clearing it.
* `kms_key_arn` - (Optional, Forces new resource) ARN of a customer-managed KMS key used for server-side encryption of the service's Amazon S3 writes.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### source

Exactly one of `inline_examples` or `s3_source` must be set. The source is used only when the dataset is created and is not returned by the service on read; changing it forces a new resource.

* `inline_examples` - (Optional) Examples provided directly in the configuration.
    * `examples` - (Required, Sensitive) List of one or more examples, each a JSON-encoded string whose structure is governed by `schema_type`. For `AGENTCORE_EVALUATION_PREDEFINED_V1`, each example must contain a `scenario_id` and a `turns` array.
* `s3_source` - (Optional) Examples loaded from a JSONL file in Amazon S3.
    * `s3_uri` - (Required) S3 URI of the JSONL file (for example, `s3://my-bucket/path/to/examples.jsonl`).

### timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Creation timestamp (RFC3339).
* `dataset_arn` - ARN of the dataset.
* `dataset_id` - Identifier of the dataset.
* `dataset_version` - Resolved version of the dataset. `DRAFT` by default.
* `draft_status` - Publish synchronization state of the draft. One of `MODIFIED` or `UNMODIFIED`.
* `example_count` - Number of examples in the draft.
* `failure_reason` - Reason for the failure when `status` is a failed state.
* `status` - Current status of the dataset. One of `CREATING`, `ACTIVE`, `UPDATING`, `DELETING`, `CREATE_FAILED`, `UPDATE_FAILED`, or `DELETE_FAILED`.
* `updated_at` - Last-updated timestamp (RFC3339).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a dataset using the `dataset_id`. For example:

```terraform
import {
  to = aws_bedrockagentcore_dataset.example
  id = "example-dataset-abc1234567"
}
```

Using `terraform import`, import a dataset using the `dataset_id`. For example:

```console
% terraform import aws_bedrockagentcore_dataset.example example-dataset-abc1234567
```

~> **Note:** The `source` argument is not returned by the service and cannot be read back on import. After importing, Terraform will plan to replace the resource in order to reconcile the `source` declared in your configuration; the dataset's examples are unaffected by a read-only refresh.
