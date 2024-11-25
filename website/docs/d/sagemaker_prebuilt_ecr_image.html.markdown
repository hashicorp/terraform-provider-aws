---
subcategory: "SageMaker"
layout: "aws"
page_title: "AWS: aws_sagemaker_prebuilt_ecr_image"
description: |-
  Get information about prebuilt Amazon SageMaker Docker images.
---

# Data Source: aws_sagemaker_prebuilt_ecr_image

Get information about prebuilt Amazon SageMaker Docker images.

~> **NOTE:** The AWS provider creates a validly constructed `registry_path` but does not verify that the `registry_path` corresponds to an existing image. For example, using a `registry_path` containing an `image_tag` that does not correspond to a Docker image in the ECR repository, will result in an error.

## Example Usage

Basic usage:

```terraform
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-scikit-learn"
  image_tag       = "2.2-1.0.11.0"
}
```

## Argument Reference

This data source supports the following arguments:

* `repository_name` - (Required) Name of the repository, which is generally the algorithm or library. Values include `autogluon-inference`, `autogluon-training`, `blazingtext`, `djl-inference`, `factorization-machines`, `forecasting-deepar`, `huggingface-pytorch-inference`, `huggingface-pytorch-inference-neuron`, `huggingface-pytorch-inference-neuronx`, `huggingface-pytorch-tgi-inference`, `huggingface-pytorch-training`, `huggingface-pytorch-training-neuronx`, `huggingface-pytorch-trcomp-training`, `huggingface-tensorflow-inference`, `huggingface-tensorflow-training`, `huggingface-tensorflow-trcomp-training`, `image-classification`, `image-classification-neo`, `ipinsights`, `kmeans`, `knn`, `lda`, `linear-learner`, `mxnet-inference`, `mxnet-inference-eia`, `mxnet-training`, `ntm`, `object-detection`, `object2vec`, `pca`, `pytorch-inference`, `pytorch-inference-eia`, `pytorch-inference-graviton`, `pytorch-inference-neuronx`, `pytorch-training`, `pytorch-training-neuronx`, `pytorch-trcomp-training`, `randomcutforest`, `sagemaker-base-python`, `sagemaker-chainer`, `sagemaker-clarify-processing`, `sagemaker-data-wrangler-container`, `sagemaker-debugger-rules`, `sagemaker-geospatial-v1-0`, `sagemaker-inference-mxnet`, `sagemaker-inference-pytorch`, `sagemaker-inference-tensorflow`, `sagemaker-model-monitor-analyzer`, `sagemaker-mxnet`, `sagemaker-mxnet-eia`, `sagemaker-mxnet-serving`, `sagemaker-mxnet-serving-eia`, `sagemaker-neo-mxnet`, `sagemaker-neo-pytorch`, `sagemaker-neo-tensorflow`, `sagemaker-pytorch`, `sagemaker-rl-coach-container`, `sagemaker-rl-mxnet`, `sagemaker-rl-ray-container`, `sagemaker-rl-tensorflow`, `sagemaker-rl-vw-container`, `sagemaker-scikit-learn`, `sagemaker-spark-processing`, `sagemaker-sparkml-serving`, `sagemaker-tensorflow`, `sagemaker-tensorflow-eia`, `sagemaker-tensorflow-scriptmode`, `sagemaker-tensorflow-serving`, `sagemaker-tensorflow-serving-eia`, `sagemaker-tritonserver`, `sagemaker-xgboost`, `semantic-segmentation`, `seq2seq`, `stabilityai-pytorch-inference`, `tei`, `tei-cpu`, `tensorflow-inference`, `tensorflow-inference-eia`, `tensorflow-inference-graviton`, `tensorflow-training`, and `xgboost-neo`.
* `dns_suffix` - (Optional) DNS suffix to use in the registry path. If not specified, the AWS provider sets it to the DNS suffix for the current region.
* `image_tag` - (Optional) Image tag for the Docker image. If not specified, the AWS provider sets the value to `1`, which for many repositories indicates the latest version. Some repositories, such as XGBoost, do not support `1` or `latest` and specific version must be used.
* `region` (Optional) - Region to use in the registry path. If not specified, the AWS provider sets it to the current region.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `registry_id` - Account ID containing the image. For example, `469771592824`.
* `registry_path` - Docker image URL. For example, `341280168497.dkr.ecr.ca-central-1.amazonaws.com/sagemaker-sparkml-serving:2.4`.
