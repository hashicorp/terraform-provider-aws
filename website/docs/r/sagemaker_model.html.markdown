---
subcategory: "SageMaker AI"
layout: "aws"
page_title: "AWS: aws_sagemaker_model"
description: |-
  Provides a SageMaker AI model resource.
---

# Resource: aws_sagemaker_model

Provides a SageMaker AI model resource.

## Example Usage

Basic usage:

```terraform
resource "aws_sagemaker_model" "example" {
  name               = "my-model"
  execution_role_arn = aws_iam_role.example.arn

  primary_container {
    image = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
  }
}

resource "aws_iam_role" "example" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional) The name of the model (must be unique). If omitted, Terraform will assign a random, unique name.
* `primary_container` - (Optional) The primary docker image containing inference code that is used when the model is deployed for predictions.  If not specified, the `container` argument is required. Fields are documented below.
* `execution_role_arn` - (Required) A role that SageMaker AI can assume to access model artifacts and docker images for deployment.
* `inference_execution_config` - (Optional) Specifies details of how containers in a multi-container endpoint are called. see [Inference Execution Config](#inference-execution-config).
* `container` (Optional) -  Specifies containers in the inference pipeline. If not specified, the `primary_container` argument is required. Fields are documented below.
* `enable_network_isolation` (Optional) - Isolates the model container. No inbound or outbound network calls can be made to or from the model container.
* `vpc_config` (Optional) - Specifies the VPC that you want your model to connect to. VpcConfig is used in hosting services and in batch transform.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

The `primary_container` and `container` block both support:

* `image` - (Optional) The registry path where the inference code image is stored in Amazon ECR.
* `mode` - (Optional) The container hosts value `SingleModel/MultiModel`. The default value is `SingleModel`.
* `model_data_url` - (Optional) The URL for the S3 location where model artifacts are stored.
* `model_package_name` - (Optional) The Amazon Resource Name (ARN) of the model package to use to create the model.
* `model_data_source` - (Optional) The location of model data to deploy. Use this for uncompressed model deployment. For information about how to deploy an uncompressed model, see [Deploying uncompressed models](https://docs.aws.amazon.com/sagemaker/latest/dg/large-model-inference-uncompressed.html) in the _AWS SageMaker AI Developer Guide_.
* `container_hostname` - (Optional) The DNS host name for the container.
* `environment` - (Optional) Environment variables for the Docker container.
   A list of key value pairs.
* `image_config` - (Optional) Specifies whether the model container is in Amazon ECR or a private Docker registry accessible from your Amazon Virtual Private Cloud (VPC). For more information see [Using a Private Docker Registry for Real-Time Inference Containers](https://docs.aws.amazon.com/sagemaker/latest/dg/your-algorithms-containers-inference-private.html). see [Image Config](#image-config).
* `inference_specification_name` - (Optional) The inference specification name in the model package version.
* `multi_model_config` - (Optional) Specifies additional configuration for multi-model endpoints. see [Multi Model Config](#multi-model-config).

### Image Config

* `repository_access_mode` - (Required) Specifies whether the model container is in Amazon ECR or a private Docker registry accessible from your Amazon Virtual Private Cloud (VPC). Allowed values are: `Platform` and `Vpc`.
* `repository_auth_config` - (Optional) Specifies an authentication configuration for the private docker registry where your model image is hosted. Specify a value for this property only if you specified Vpc as the value for the RepositoryAccessMode field, and the private Docker registry where the model image is hosted requires authentication. see [Repository Auth Config](#repository-auth-config).

#### Repository Auth Config

* `repository_credentials_provider_arn` - (Required) The Amazon Resource Name (ARN) of an AWS Lambda function that provides credentials to authenticate to the private Docker registry where your model image is hosted. For information about how to create an AWS Lambda function, see [Create a Lambda function with the console](https://docs.aws.amazon.com/lambda/latest/dg/getting-started-create-function.html) in the _AWS Lambda Developer Guide_.

### Model Data Source

* `s3_data_source` - (Required) The S3 location of model data to deploy.

#### S3 Data Source

* `compression_type` - (Required) How the model data is prepared. Allowed values are: `None` and `Gzip`.
* `s3_data_type` - (Required) The type of model data to deploy. Allowed values are: `S3Object` and `S3Prefix`.
* `s3_uri` - (Required) The S3 path of model data to deploy.
* `model_access_config` - (Optional) Specifies the access configuration file for the ML model. You can explicitly accept the model end-user license agreement (EULA) within the [`model_access_config` configuration block]. see [Model Access Config](#model-access-config).

##### Model Access Config

* `accept_eula` - (Required) Specifies agreement to the model end-user license agreement (EULA). The AcceptEula value must be explicitly defined as `true` in order to accept the EULA that this model requires. You are responsible for reviewing and complying with any applicable license terms and making sure they are acceptable for your use case before downloading or using a model.

### Multi Model Config

* `model_cache_setting` - (Optional) Whether to cache models for a multi-model endpoint. By default, multi-model endpoints cache models so that a model does not have to be loaded into memory each time it is invoked. Some use cases do not benefit from model caching. For example, if an endpoint hosts a large number of models that are each invoked infrequently, the endpoint might perform better if you disable model caching. To disable model caching, set the value of this parameter to `Disabled`. Allowed values are: `Enabled` and `Disabled`.

## Inference Execution Config

* `mode` - (Required) How containers in a multi-container are run. The following values are valid `Serial` and `Direct`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `name` - The name of the model.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this model.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import models using the `name`. For example:

```terraform
import {
  to = aws_sagemaker_model.test_model
  id = "model-foo"
}
```

Using `terraform import`, import models using the `name`. For example:

```console
% terraform import aws_sagemaker_model.test_model model-foo
```
