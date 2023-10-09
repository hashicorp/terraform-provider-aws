---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_custom_model"
description: |-
  Manages a Bedrock custom model
---

# Resource: aws_bedrock_custom_model

Manages a Bedrock custom model.

## Example Usage

```terraform
resource "random_id" "id" {
  byte_length = 8
}

data aws_caller_identity current {}

resource aws_s3_bucket training_data {
  bucket = "bedrock-training-data-${random_id.id.hex}"
}

resource aws_s3_bucket validation_data {
  bucket = "bedrock-validation-data-${random_id.id.hex}"
}

resource aws_s3_bucket output_data {
  bucket = "bedrock-output-data-${random_id.id.hex}"
}

resource "aws_s3_bucket_object" "training_data" {
  bucket = aws_s3_bucket.training_data.id
  key    = "myfolder/training_data.jsonl"
  source = "./testdata/training_data.jsonl"
  etag   = filemd5("./testdata/training_data.jsonl")
}

resource "aws_s3_bucket_object" "validation_data" {
  bucket = aws_s3_bucket.validation_data.id
  key    = "myfolder/validation_data.jsonl"
  source = "./testdata/validation_data.jsonl"
  etag   = filemd5("./testdata/validation_data.jsonl")
}

resource "aws_iam_role" "bedrock_fine_tuning" {
  name = "bedrock-fine-tuning-${random_id.id.hex}"

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"Service": "bedrock.amazonaws.com"
			},
			"Action": "sts:AssumeRole",
			"Condition": {
				"StringEquals": {
					"aws:SourceAccount": "${data.aws_caller_identity.current.account_id}"
				},
				"ArnEquals": {
					"aws:SourceArn": "arn:aws:bedrock:us-east-1:${data.aws_caller_identity.current.account_id}:model-customization-job/*"
				}
			}
		}
	] 
}
EOF
}

resource "aws_iam_policy" "BedrockAccessTrainingValidationS3Policy" {
  name        = "BedrockAccessTrainingValidationS3Policy_${random_id.id.hex}"
  path        = "/"
  description = "BedrockAccessTrainingValidationS3Policy"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket",
          "s3:ListObjects"
        ],
        "Resource" : [
          "${aws_s3_bucket.training_data.arn}",
          "${aws_s3_bucket.training_data.arn}/myfolder",
          "${aws_s3_bucket.training_data.arn}/myfolder/*",
          "${aws_s3_bucket.validation_data.arn}/myfolder",
          "${aws_s3_bucket.validation_data.arn}/myfolder/*"
        ]
      }
    ]
  })
}

resource "aws_iam_policy" "BedrockAccessOutputS3Policy" {
  name        = "BedrockAccessOutputS3Policy_${random_id.id.hex}"
  path        = "/"
  description = "BedrockAccessOutputS3Policy"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket",
          "s3:ListObjects"
        ],
        "Resource" : [
          "${aws_s3_bucket.output_data.arn}/myfolder",
          "${aws_s3_bucket.output_data.arn}/myfolder/*"
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "bedrock_attachment_1" {
  role       = aws_iam_role.bedrock_fine_tuning.name
  policy_arn = aws_iam_policy.BedrockAccessTrainingValidationS3Policy.arn
}

resource "aws_iam_role_policy_attachment" "bedrock_attachment_2" {
  role       = aws_iam_role.bedrock_fine_tuning.name
  policy_arn = aws_iam_policy.BedrockAccessOutputS3Policy.arn
}

resource "aws_bedrock_custom_model" "test" {
  custom_model_name = "tf-test-${random_id.id.hex}"
  job_name          = "tf-test-${random_id.id.hex}"
  base_model_arn    = "amazon.titan-text-express-v1"
  hyper_parameters = {
    "epochCount"              = "1"
    "batchSize"               = "1"
    "learningRate"            = "0.005"
    "learningRateWarmupSteps" = "0"
  }
  output_data_config   = "s3://${aws_s3_bucket.output_data.id}/myfolder/"
  role_arn             = aws_iam_role.bedrock_fine_tuning.arn
  training_data_config = "s3://${aws_s3_bucket.training_data.id}/myfolder/training_data.jsonl"
}
```

## Argument Reference

The following arguments are required:

* `base_model_identifier` - Name of the base model. Type: String. Required: Yes. Length Constraints: Minimum length of 1. Maximum length of 2048. Pattern: ^(arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}/[a-z0-9]{12})|(:foundation-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([a-z0-9-]{1,63}[.]){0,2}[a-z0-9-]{1,63}([:][a-z0-9-]{1,63}){0,2})))|([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})|(([0-9a-zA-Z][_-]?)+)$
* `custom_model_name` - Name for the custom model. Type: String. Required: Yes. Length Constraints: Minimum length of 1. Maximum length of 63. Pattern: ^([0-9a-zA-Z][_-]?)+$
* `job_name` - Enter a unique name for the fine-tuning job. Type: String. Required: Yes. Length Constraints: Minimum length of 1. Maximum length of 63. Pattern: ^[a-zA-Z0-9](-*[a-zA-Z0-9\+\-\.])*$
* `output_data_config` - S3 location for the output data. Type: String. Required: Yes
* `role_arn` - The Amazon Resource Name (ARN) of an IAM role that Bedrock can assume to perform tasks on your behalf. Type: String. Required: Yes. Length Constraints: Minimum length of 0. Maximum length of 2048. Pattern: ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$
* `training_data_config` - Information about the training dataset. Type: String. Required: Yes

The following arguments are optional:

* `client_request_token` - Unique token value that you can provide. The GetModelCustomizationJob response includes the same token value. Type: String. Required: No. Length Constraints: Minimum length of 1. Maximum length of 256. Pattern: ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$
* `custom_model_kms_key_id` - The custom model is encrypted at rest using this key. Type: String. Required: No. Length Constraints: Minimum length of 1. Maximum length of 2048. Pattern: ^arn:aws(-[^:]+)?:kms:[a-zA-Z0-9-]*:[0-9]{12}:((key/[a-zA-Z0-9-]{36})|(alias/[a-zA-Z0-9-_/]+))$
* `validation_data_config` - Information about the validation dataset. Type: string. Required: No.
* `vpc_config` - VPC configuration (optional). Configuration parameters for the private Virtual Private Cloud (VPC) that contains the resources you are using for this job. Type: VpcConfig object. Required: No.
* `job_tags` - (Optional) Key-value mapping of tags for the fine-tuning job.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

### VpcConfig Object

* `addon_name` – (Required) Name of the EKS add-on. The name must match one of
  the names returned by [describe-addon-versions](https://docs.aws.amazon.com/cli/latest/reference/eks/describe-addon-versions.html).
* `cluster_name` – (Required) Name of the EKS Cluster. Must be between 1-100 characters in length. Must begin with an alphanumeric character, and must only contain alphanumeric characters, dashes and underscores (`^[0-9A-Za-z][A-Za-z0-9\-_]+$`).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Custom Model using the `model_id`:

```terraform
import {
  to       = aws_bedrock_custom_model.my_model
  model_id = "my_model_arn"
}
```

Using `terraform import`, import Bedrock custom model using the `model_id`:

```console
% terraform import aws_bedrock_custom_model.my_model my_model_arn
```
