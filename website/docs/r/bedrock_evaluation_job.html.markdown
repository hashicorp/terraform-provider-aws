---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_evaluation_job"
description: ,-
  Terraform resource for managing an AWS Amazon Bedrock Evaluation Job.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_bedrock_evaluation_job

Terraform resource for managing an AWS Amazon Bedrock Evaluation Job.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrock_evaluation_job" "example" {
}
```

## Argument Reference

The following arguments are required:

* `client_request_token` -  (Required) Unique case-sensitive identifier to ensure that the API request completes no more than one time. Bedrock will ignore requests with previous tokens and not return an error.
* `customer_encryption_key` -  (Required) Key ARN that will be used to encrypt the evaluation job. It must have a Length of between 1 and 2048 and follow regex of ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$.
* `evaluation_config` -  (Required) Object that contains one of the following objects to describe the evaluation config.
    * `automated` -  (Required) Object that is used to specify and describe an automated evaluation job through the following attributes.
        * `automated_evaluation_config` - (Required) Object that is used to specify and describe the automated evaluation job through the following attributes.
            * `dataset_metric_configs` -  (Required) Array of EvaluationDatasetMetricConfig objects that specify the prompt datasets, task type and metric names through the following attributes. Array must have a length between 1 and 5.
                * `metric_names` -  (Required) Array of names of the metrics used. Valid values are Builtin.Accuracy, Builtin.Robustness, and Builtin.Toxicity.
                * `task_type` -  (Required) Task type that you want the model to carry out. Valid Values are Summarization , Classification, QuestionAndAnswer, Generation, and Custom.
                * `dataset` -  (Required) Object that is used to specify and describe the prompt dataset through the following attributes.
                  * `name` -  (Required) Name of built-in prompt datasets. Valid values are Builtin.Bold, Builtin.BoolQ, Builtin.NaturalQuestions, Builtin.Gigaword, Builtin.RealToxicityPrompts, Builtin.TriviaQa, Builtin.T-Rex, Builtin.WomensEcommerceClothingReviews and Builtin.Wikitext2.
                  * `dataset_location` -  (Optional) Object that is used to specify and describe the location in Amazon S3 where the prompt dataset is saved through the following attributes.
                    * `s3_uri` -  (Required) Amazon S3 URI where the prompt dataset is stored. It must have a Length of between 1 and 1024 and follow regex of ^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$.

* `inference_config` -  (Required) Array of objects that define the models you want used in your model evaluation job through. Automated model evaluation jobs support only a single model. In a human-based model evaluation job, your annotator can compare the responses for up to two different models.
    * `model` - (Required) Defines a model used in the model evaluation jobs through the following attributes.
        * ``

* `name` -  (Required) Name of the model evaluation job. Must be unique within your AWS account and account region. It must hve a length of between 1 and 63 and follow the regex of ^[a-z0-9](-*[a-z0-9]){0,62}$.
* `output_data_config` -  (Required) Object with the following attributes that defines wat Amazon S3 will have your results saved.
    * `s3_uri` -  (Required) Amazon S3 URI where the results of model evaluation job are saved. It must have a Length of between 1 and 1024 and follow regex of ^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$.
  






The following arguments are optional:

* `description` -  (Optional) Description of the evaluation job. It must hve a length of between 1 and 200 and follow the regex of ^.+$.
* `tags` -  (Optional) Tags to attach to the model evaluation job. Maximum 200 items. Each tag key must have a length of between 1 and 128 and follow the regex of ^[a-zA-Z0-9\s._:/=+@-]*$. Each tag value must have a length of between 1 and 256 and follow the regex of ^[a-zA-Z0-9\s._:/=+@-]*$.




## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Evaluation Job. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Bedrock Evaluation Job using the `example_id_arg`. For example:

```terraform
import {
  to = aws_bedrock_evaluation_job.example
  id = "evaluation_job-id-12345678"
}
```

Using `terraform import`, import Amazon Bedrock Evaluation Job using the `example_id_arg`. For example:

```console
% terraform import aws_bedrock_evaluation_job.example evaluation_job-id-12345678
```
