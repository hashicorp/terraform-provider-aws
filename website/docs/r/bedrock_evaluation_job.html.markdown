---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_evaluation_job"
description: ,-
  Terraform resource for managing an AWS Amazon Bedrock Evaluation Job.
---
# Resource: aws_bedrock_evaluation_job

Terraform resource for managing an AWS Amazon Bedrock Evaluation Job.

## Example Usage

resource "aws_bedrock_evaluation_job" "test" {

  evaluation_config {
    automated {
        dataset_metric_configs {
          dataset {
		    name = "BoolQ"
			dataset_location {
				s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.dataset.key}"
			}
          }
          metric_names = ["Builtin.Accuracy"]
          task_type    = "QuestionAndAnswer"
        }
    }
  }

  inference_config {
    models {
      bedrock_model { 
        inference_params = tolist(data.aws_bedrock_foundation_model.test.inference_types_supported)[0]
        model_identifier = data.aws_bedrock_foundation_model.test.id
		}
    }
  }

  customer_encryption_key_id = aws_kms_key.test.arn
  description = "test"
  name        = %[1]q

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.bucket}/bedrock/evaluation_jobs"
  }

  tags = {
	%[2]q = %[3]q
	%[4]q = %[5]q
  }

  role_arn = aws_iam_role.test.arn
}


### Basic Usage

resource "aws_bedrock_evaluation_job" "test" {

  evaluation_config {
    automated {
        dataset_metric_configs {
          dataset {
		    name = "BoolQ"
			dataset_location {
				s3_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.dataset.key}"
			}
          }
          metric_names = ["Builtin.Accuracy"]
          task_type    = "QuestionAndAnswer"
        }
    }
  }
  inference_config {
    models {
      bedrock_model { 
        inference_params = tolist(data.aws_bedrock_foundation_model.test.inference_types_supported)[0]
        model_identifier = data.aws_bedrock_foundation_model.test.id
		}
    }
  }

  description = "test"
  name        = %[1]q

  output_data_config {
    s3_uri = "s3://${aws_s3_bucket.test.bucket}/bedrock/evaluation_jobs"
  }

  role_arn = aws_iam_role.test.arn
}

## Argument Reference

The following arguments are required:

* `client_request_token` -  (Required) Unique case-sensitive identifier to ensure that the API request completes no more than one time. Bedrock will ignore requests with previous tokens and not return an error.
* `customer_encryption_key` -  (Required) Key ARN that will be used to encrypt the evaluation job. It must have a Length of between 1 and 2048 and follow regex of ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$.
* `evaluation_config` -  (Required) Object that contains only one of the following objects to describe the evaluation config.
    * `automated` -  (Required) Object that is used to specify and describe an automated evaluation job through the following attributes.
        * `automated_evaluation_config` - (Required) Object that is used to specify and describe the automated evaluation job through the following attributes.
            * `dataset_metric_configs` -  (Required) Array of EvaluationDatasetMetricConfig objects that specify the prompt datasets, task type and metric names through the following attributes. Array must have a length between 1 and 5.
                * `metric_names` -  (Required) Array of names of the metrics used. Valid values are Builtin.Accuracy, Builtin.Robustness, and Builtin.Toxicity.
                * `task_type` -  (Required) Task type that you want the model to carry out. Valid Values are Summarization , Classification, QuestionAndAnswer, Generation, and Custom.
                * `dataset` -  (Required) Object that is used to specify and describe the prompt dataset through the following attributes.
                  * `name` -  (Required) Name of built-in prompt datasets. Valid values are Builtin.Bold, Builtin.BoolQ, Builtin.NaturalQuestions, Builtin.Gigaword, Builtin.RealToxicityPrompts, Builtin.TriviaQa, Builtin.T-Rex, Builtin.WomensEcommerceClothingReviews and Builtin.Wikitext2.
                  * `dataset_location` -  (Optional) Object that is used to specify and describe the location in Amazon S3 where the prompt dataset is saved through the following attributes.
                    * `s3_uri` -  (Required) Amazon S3 URI where the prompt dataset is stored. It must have a Length of between 1 and 1024 and follow regex of ^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$.
    * `human` -  (Required) Object that is used to specify and describe a human evaluation job through the following attributes. This feature is not tested.
        * `dataset_metric_configs` -  (Required) Array of EvaluationDatasetMetricConfig objects that specify the prompt datasets, task type and metric names through the following attributes. Array must have a length between 1 and 5.
                * `metric_names` -  (Required) Array of names of the metrics used. Valid values are Builtin.Accuracy, Builtin.Robustness, and Builtin.Toxicity.
                * `task_type` -  (Required) Task type that you want the model to carry out. Valid Values are Summarization , Classification, QuestionAndAnswer, Generation, and Custom.
                * `dataset` -  (Required) Object that is used to specify and describe the prompt dataset through the following attributes.
                  * `name` -  (Required) Name of built-in prompt datasets. Valid values are Builtin.Bold, Builtin.BoolQ, Builtin.NaturalQuestions, Builtin.Gigaword, Builtin.RealToxicityPrompts, Builtin.TriviaQa, Builtin.T-Rex, Builtin.WomensEcommerceClothingReviews and Builtin.Wikitext2.
                  * `dataset_location` -  (Optional) Object that is used to specify and describe the location in Amazon S3 where the prompt dataset is saved through the following attributes.
                    * `s3_uri` -  (Required) Amazon S3 URI where the prompt dataset is stored. It must have a Length of between 1 and 1024 and follow regex of ^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$.
        * `custom_metrics` -  (Optional) Array of HumanEvaluationCustomMetric objects with a minimum length of 1 and maximum of 10 items
            * `name` -  (Required) Name of them metric that your human evaluatiors will see in the evaluation UI. It must have a length of betweeen 1 and 63 and follow regex of ^[0-9a-zA-Z-_.]+$.
            * `rating_method` -  (Required) Rating method of how you want humans to evaluate your model. Valid values are ThumbsUpDown, IndividualLikertScale,ComparisonLikertScale, ComparisonChoice, and ComparisonRank.
            * `description` -  (Optional) Description about method.
        * `human_workflow_config` -  (Optional) SageMakerFlowDefinition object that names metrics and describes their evaluation.
            * `name` -  (Required) Name of them metric that your human evaluatiors will see in the evaluation UI. It must have a length of betweeen 1 and 63 and follow regex of ^[0-9a-zA-Z-_.]+$.
            * `rating_method` -  (Required) Rating method of how you want humans to evaluate your model. Valid values are ThumbsUpDown, IndividualLikertScale,ComparisonLikertScale, ComparisonChoice, and ComparisonRank.
            * `description` -  (Optional) Description about method.
* `inference_config` -  (Required) Object that describes the models you want to use in your model evaluation job.
    * `models` - (Required) Array of model objects that define the models used in your model evaluation job through. Automated model evaluation jobs support only a single model. In a human-based model evaluation job, your annotator can compare the responses for up to two different models.
        * `bedrock_model` (Required) Defines the model that is used through the following attributes.
            * `inference_params` (Required) Bedrock model inference parameters.
            * `model_identifier` (Required) ARN of the Bedrock model. Must follow the regex of ^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}(([:][a-z0-9-]{1,63}){0,2})?/[a-z0-9]{12})|(:foundation-model/([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})))|(([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2}))|(([0-9a-zA-Z][_-]?)+)$.
* `name` -  (Required) Name of the model evaluation job. Must be unique within your AWS account and account region. It must hve a length of between 1 and 63 and follow the regex of ^[a-z0-9](-*[a-z0-9]){0,62}$.
* `output_data_config` -  (Required) Object with the following attributes that defines wat Amazon S3 will have your results saved.
    * `s3_uri` -  (Required) Amazon S3 URI where the results of model evaluation job are saved. It must have a Length of between 1 and 1024 and follow regex of ^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$.
  
The following arguments are optional:
* `description` -  (Optional) Description of the evaluation job. It must hve a length of between 1 and 200 and follow the regex of ^.+$.
* `tags` -  (Optional) Tags to attach to the model evaluation job. Maximum 200 items. Each tag key must have a length of between 1 and 128 and follow the regex of ^[a-zA-Z0-9\s._:/=+@-]*$. Each tag value must have a length of between 1 and 256 and follow the regex of ^[a-zA-Z0-9\s._:/=+@-]*$.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Evaluation Job. 

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `read` - (Default `60m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Bedrock Evaluation Job using the `evaluation-job-arn`. For example:

```terraform
import {
  to = aws_bedrock_evaluation_job.example
  id = "evaluation_job-arn-12345678"
}
```

Using `terraform import`, import Amazon Bedrock Evaluation Job using the `evaluation-job-arn`. For example:

```console
% terraform import aws_bedrock_evaluation_job.example evaluation_job-id-12345678
```
