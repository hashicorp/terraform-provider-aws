---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_evaluation_job"
description: |-
  Manages an Amazon Bedrock evaluation job.
---

# Resource: aws_bedrock_evaluation_job

Manages an Amazon Bedrock evaluation job. An evaluation job assesses model or knowledge base performance using either automated metrics or human workers.

~> Amazon Bedrock does not support permanently deleting an evaluation job. Destroying this resource stops the job (if it is still running) using the [StopEvaluationJob](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_StopEvaluationJob.html) API, then removes it from Terraform state. Set `skip_destroy` to leave the job in its current state instead.

## Example Usage

### Automated Model Evaluation

```terraform
resource "aws_bedrock_evaluation_job" "example" {
  job_name = "example-job"
  role_arn = aws_iam_role.example.arn

  evaluation_config {
    automated {
      dataset_metric_config {
        task_type = "Generation"

        dataset {
          name = "Builtin.Bold"
        }

        metric_names = ["Builtin.Robustness"]
      }
    }
  }

  inference_config {
    model {
      bedrock_model {
        model_identifier = "amazon.nova-micro-v1:0"
      }
    }
  }

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.example.id}/output/"
  }
}
```

## Argument Reference

The following arguments are required:

* `evaluation_config` - (Required) Configuration for either an automated or human-based evaluation job. See [`evaluation_config` Block](#evaluation_config-block) below.
* `inference_config` - (Required) Configuration for the inference model, or models, used for the evaluation job. See [`inference_config` Block](#inference_config-block) below.
* `job_name` - (Required) Name for the evaluation job. Must be unique within your AWS account and Region, and consist of lowercase letters, numbers, and hyphens.
* `output_data_config` - (Required) Configuration for the Amazon S3 location where the results of the evaluation job are stored. See [`output_data_config` Block](#output_data_config-block) below.
* `role_arn` - (Required) ARN of an IAM service role that Amazon Bedrock can assume to perform tasks on your behalf. See [Required permissions for model evaluations](https://docs.aws.amazon.com/bedrock/latest/userguide/model-evaluation-security.html).

The following arguments are optional:

* `application_type` - (Optional) Whether the evaluation job evaluates a model or a knowledge base. Valid values: `ModelEvaluation`, `RagEvaluation`.
* `customer_encryption_key_id` - (Optional) ARN of the customer managed KMS key to use to encrypt the evaluation job.
* `job_description` - (Optional) Description of the evaluation job.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `skip_destroy` - (Optional) Whether to leave the evaluation job in its current state when destroying the resource, instead of stopping it.
* `tags` - (Optional) Map of tags to assign to the evaluation job. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `evaluation_config` Block

The `evaluation_config` block supports the following arguments. Exactly one of `automated` or `human` is required.

* `automated` - (Optional) Configuration for an automated evaluation job that computes metrics. See [`automated` Block](#automated-block) below.
* `human` - (Optional) Configuration for an evaluation job that uses human workers. See [`human` Block](#human-block) below.

### `automated` Block

The `automated` block supports the following arguments:

* `custom_metric_config` - (Optional) Configuration for custom metrics to compute for the evaluation job. See [`custom_metric_config` Block](#custom_metric_config-block) below.
* `dataset_metric_config` - (Required) One or more configurations for the prompt datasets and metrics to use. See [`dataset_metric_config` Block](#dataset_metric_config-block) below.
* `evaluator_model_config` - (Optional) Configuration for the evaluator (judge) model. Required for automated jobs that use an LLM-as-judge metric, or that evaluate a knowledge base. See [`evaluator_model_config` Block](#evaluator_model_config-block) below.

### `dataset_metric_config` Block

The `dataset_metric_config` block, used by both the `automated` and `human` blocks, supports the following arguments:

* `dataset` - (Required) Prompt dataset to use. See [`dataset` Block](#dataset-block) below.
* `metric_names` - (Required) Names of the metrics to use for the evaluation job.
* `task_type` - (Required) Type of task to evaluate. Common values are `Summarization`, `Classification`, `QuestionAndAnswer`, `Generation`, and `Custom`. Use `General` for automated evaluation jobs that use a judge model (`evaluator_model_config`).

### `dataset` Block

The `dataset` block supports the following arguments:

* `dataset_location` - (Optional) Location of a custom prompt dataset. See [`dataset_location` Block](#dataset_location-block) below.
* `name` - (Required) Name of a built-in prompt dataset, for example `Builtin.Bold`, or a label for a custom prompt dataset.

### `dataset_location` Block

The `dataset_location` block supports the following arguments:

* `s3_uri` - (Required) S3 URI of the custom prompt dataset.

### `evaluator_model_config` Block

The `evaluator_model_config` block, used by both the `automated` and `custom_metric_config` blocks, supports the following arguments:

* `bedrock_evaluator_model` - (Required) Evaluator model. See [`bedrock_evaluator_model` Block](#bedrock_evaluator_model-block) below.

### `bedrock_evaluator_model` Block

The `bedrock_evaluator_model` block, used by both the `evaluator_model_config` blocks under `automated` and `custom_metric_config`, supports the following arguments:

* `model_identifier` - (Required) Identifier of the Amazon Bedrock model, or inference profile, used to compute the metrics.

### `custom_metric_config` Block

The `custom_metric_config` block supports the following arguments:

* `custom_metric` - (Required) One or more custom metric definitions. See [`evaluation_config.automated.custom_metric_config.custom_metric` Block](#evaluation_configautomatedcustom_metric_configcustom_metric-block) below.
* `evaluator_model_config` - (Required) Configuration for the evaluator model used to compute the custom metrics. See [`evaluator_model_config` Block](#evaluator_model_config-block) above.

### `evaluation_config.automated.custom_metric_config.custom_metric` Block

The `custom_metric` block, nested under `custom_metric_config`, supports the following arguments:

* `custom_metric_definition` - (Required) Definition of the custom metric. See [`custom_metric_definition` Block](#custom_metric_definition-block) below.

### `custom_metric_definition` Block

The `custom_metric_definition` block supports the following arguments:

* `instructions` - (Required) Prompt that instructs the evaluator model how to rate the model or RAG source under evaluation.
* `name` - (Required) Name for the custom metric. Must be unique in your AWS Region.
* `rating_scale` - (Optional) One or more items defining the rating scale for the custom metric. See [`rating_scale` Block](#rating_scale-block) below.

### `rating_scale` Block

The `rating_scale` block supports the following arguments:

* `definition` - (Required) Definition for one rating in the custom metric rating scale.
* `value` - (Required) Value for one rating in the custom metric rating scale. See [`value` Block](#value-block) below.

### `value` Block

The `value` block supports the following arguments. Exactly one of `float_value` or `string_value` is required.

* `float_value` - (Optional) Floating point number representing the rating value.
* `string_value` - (Optional) String representing the rating value.

### `human` Block

The `human` block supports the following arguments:

* `custom_metric` - (Optional) One or more custom metrics for your human workers to use. See [`evaluation_config.human.custom_metric` Block](#evaluation_confighumancustom_metric-block) below.
* `dataset_metric_config` - (Required) One or more configurations for the prompt datasets and metrics to use. See [`dataset_metric_config` Block](#dataset_metric_config-block) above.
* `human_workflow_config` - (Optional) Configuration for the human workflow. See [`human_workflow_config` Block](#human_workflow_config-block) below.

### `evaluation_config.human.custom_metric` Block

The `custom_metric` block, nested under `human`, supports the following arguments:

* `description` - (Optional) Description of the metric.
* `name` - (Required) Name of the metric.
* `rating_method` - (Required) How the metric is rated. Valid values: `ThumbsUpDown`, `IndividualLikertScale`, `ComparisonLikertScale`, `ComparisonChoice`, `ComparisonRank`.

### `human_workflow_config` Block

The `human_workflow_config` block supports the following arguments:

* `flow_definition_arn` - (Required) ARN of the Amazon SageMaker AI flow definition.
* `instructions` - (Optional) Instructions for the flow definition.

### `inference_config` Block

The `inference_config` block supports the following arguments. Exactly one of `model` or `rag_config` is required.

* `model` - (Optional) One or more inference models. Automated jobs support a single model; jobs that use human workers support up to two models. See [`model` Block](#model-block) below.
* `rag_config` - (Optional) Inference configuration for a knowledge base evaluation job. See [`rag_config` Block](#rag_config-block) below.

### `model` Block

The `model` block supports the following arguments. Exactly one of `bedrock_model` or `precomputed_inference_source` is required.

* `bedrock_model` - (Optional) Amazon Bedrock model. See [`bedrock_model` Block](#bedrock_model-block) below.
* `precomputed_inference_source` - (Optional) Model where you provide your own precomputed inference response data. See [`precomputed_inference_source` Block](#precomputed_inference_source-block) below.

### `bedrock_model` Block

The `bedrock_model` block supports the following arguments:

* `inference_params` - (Optional) JSON-formatted string of inference parameters for the model.
* `model_identifier` - (Required) Identifier of the Amazon Bedrock model, or inference profile, used for inference.
* `performance_config` - (Optional) Model's performance settings. See [`performance_config` Block](#performance_config-block) below.

### `performance_config` Block

The `performance_config` block supports the following arguments:

* `latency` - (Optional) Whether to use the latency-optimized or standard version of the model. Valid values: `standard`, `optimized`.

### `precomputed_inference_source` Block

The `precomputed_inference_source` block supports the following arguments:

* `inference_source_identifier` - (Required) Label that identifies the precomputed inference source.

### `rag_config` Block

The `rag_config` block supports the following arguments. Exactly one of `knowledge_base_config` or `precomputed_rag_source_config` is required.

* `knowledge_base_config` - (Optional) Amazon Bedrock knowledge base. See [`knowledge_base_config` Block](#knowledge_base_config-block) below.
* `precomputed_rag_source_config` - (Optional) RAG source where you provide your own precomputed inference response data. See [`precomputed_rag_source_config` Block](#precomputed_rag_source_config-block) below.

### `knowledge_base_config` Block

The `knowledge_base_config` block supports the following arguments. Exactly one of `retrieve_and_generate_config` or `retrieve_config` is required.

* `retrieve_and_generate_config` - (Optional) Configuration for retrieval with response generation. See [`retrieve_and_generate_config` Block](#retrieve_and_generate_config-block) below.
* `retrieve_config` - (Optional) Configuration for retrieval only. See [`retrieve_config` Block](#retrieve_config-block) below.

### `retrieve_and_generate_config` Block

The `retrieve_and_generate_config` block supports the following arguments:

* `knowledge_base_id` - (Required) Identifier of the knowledge base.
* `model_arn` - (Required) ARN of the foundation model, or inference profile, used to generate responses.
* `retrieval_configuration` - (Optional) Knowledge base retrieval configuration. See [`retrieval_configuration` Block](#retrieval_configuration-block) below.

### `retrieval_configuration` Block

The `retrieval_configuration` block supports the following arguments:

* `vector_search_configuration` - (Required) Vector search configuration. See [`vector_search_configuration` Block](#vector_search_configuration-block) below.

### `vector_search_configuration` Block

The `vector_search_configuration` block, used by both the `retrieval_configuration` and `knowledge_base_retrieval_configuration` blocks, supports the following arguments:

* `number_of_results` - (Optional) Number of text chunks to retrieve.

### `retrieve_config` Block

The `retrieve_config` block supports the following arguments:

* `knowledge_base_id` - (Required) Identifier of the knowledge base.
* `knowledge_base_retrieval_configuration` - (Optional) Knowledge base retrieval configuration. See [`knowledge_base_retrieval_configuration` Block](#knowledge_base_retrieval_configuration-block) below.

### `knowledge_base_retrieval_configuration` Block

The `knowledge_base_retrieval_configuration` block supports the following arguments:

* `vector_search_configuration` - (Required) Vector search configuration. See [`vector_search_configuration` Block](#vector_search_configuration-block) above.

### `precomputed_rag_source_config` Block

The `precomputed_rag_source_config` block supports the following arguments. Exactly one of `retrieve_and_generate_source_config` or `retrieve_source_config` is required.

* `retrieve_and_generate_source_config` - (Optional) Configuration for retrieval with response generation. See [`retrieve_and_generate_source_config` Block](#retrieve_and_generate_source_config-block) below.
* `retrieve_source_config` - (Optional) Configuration for retrieval only. See [`retrieve_source_config` Block](#retrieve_source_config-block) below.

### `retrieve_and_generate_source_config` Block

The `retrieve_and_generate_source_config` block supports the following arguments:

* `rag_source_identifier` - (Required) Label that identifies the precomputed RAG source.

### `retrieve_source_config` Block

The `retrieve_source_config` block supports the following arguments:

* `rag_source_identifier` - (Required) Label that identifies the precomputed RAG source.

### `output_data_config` Block

The `output_data_config` block supports the following arguments:

* `s3_uri` - (Required) S3 URI where the results of the evaluation job are stored.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Date and time the evaluation job was created.
* `failure_messages` - List of reasons the evaluation job failed to create, if applicable.
* `job_arn` - ARN of the evaluation job.
* `job_type` - Whether the evaluation job is automated or human-based.
* `last_modified_time` - Date and time the evaluation job was last modified.
* `status` - Current status of the evaluation job.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrock_evaluation_job.example
  identity = {
    "job_arn" = "arn:aws:bedrock:us-west-2:123456789012:evaluation-job/abcdefgh1234"
  }
}

resource "aws_bedrock_evaluation_job" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

- `job_arn` (String) ARN of the evaluation job.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Evaluation Job using the `job_arn`. For example:

```terraform
import {
  to = aws_bedrock_evaluation_job.example
  id = "arn:aws:bedrock:us-west-2:123456789012:evaluation-job/abcdefgh1234"
}
```

Using `terraform import`, import Bedrock Evaluation Job using the `job_arn`. For example:

```console
% terraform import aws_bedrock_evaluation_job.example arn:aws:bedrock:us-west-2:123456789012:evaluation-job/abcdefgh1234
```
