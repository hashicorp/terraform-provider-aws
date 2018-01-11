---
layout: "aws"
page_title: "AWS: sagemaker_training_job"
sidebar_current: "docs-aws-resource-sagemaker-training-job"
description: |-
  Provides a Sagemaker Training Job resource.
---

# aws\_sagemaker\_training\_job

Provides a Sagemaker Training Job resource.

Training Job cannot be deleted once they are create. Check the [AWS SageMaker docs](https://docs.aws.amazon.com/sagemaker/latest/dg/whatis.html) for more details.

## Example Usage

Basic usage:

```hcl
resource "aws_sagemaker_training_job" "tj" {
  name = "my-training-job"
  role_arn = "${aws_iam_role.role.arn}"

  algorithm_specification {
    image = "174872318107.dkr.ecr.us-west-2.amazonaws.com/kmeans:1"
    input_mode = "File"
  }

  resource_config {
    instance_type = "ml.c4.8xlarge"
    instance_count = 2
    volume_size_in_gb = 30
  }

  stopping_condition {
    max_runtime_in_seconds = 3600
  }

  hyper_parameters {
    epochs = "3"
    feature_dim = "784"
    force_dense = "True"
    k = "10"
    mini_batch_size = "500"
  }

  input_data_config = [{
    name = "train"
    data_source {
      s3_data_source {
        s3_data_type = "ManifestFile"
        s3_uri = "s3://bucket/kmeans_highlevel_example/data/KMeans-2018-01-10-14-45-00-000/.amazon.manifest"
        s3_data_distribution_type = "ShardedByS3Key"
      }
    }
  }]

  output_data_config {
    s3_output_path = "s3://bucket/kmeans_example/output"
  }

  tags {
    Name = "foo"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the training job (must be unique).
* `role_arn` - (Required) The ARN of the IAM role to be used by the training job which allows SageMaker to call other services on your behalf.
* `algorithm_specification` - (Required) Fields are documented below.
* `resource_config` - (Required) Fields are documented below.
* `stopping_condition` - (Optional) Fields are documented below.
* `hyper_parameters` - (Optional) A map of string keys and string values that will be passed in to the training algorithm.
* `input_data_config` - (Required) Fields are documented below.
* `output_data_config` - (Required) Fields are documented below.
* `tags` - (Optional) A mapping of tags to assign to the resource.

The `algorithm_specification` block supports:

* `image` - (Required) The registry path where the training code image is stored in Amazon ECR.
* `input_mode` - (Optional) The input mode that the training algorithm supports.

See [AWS SageMaker Algorithm Specification doc](https://docs.aws.amazon.com/sagemaker/latest/dg/API_AlgorithmSpecification.html) for details.

The `resource_config` block supports:

* `instance_type` - (Required) The name of ML compute instance type.
* `instance_count` - (Required) The number of ML compute instances to be used for training.
* `volume_size_in_gb` - (Required) The size in GB of the the ML volumn storage.

See [AWS SageMaker Resource Config doc](https://docs.aws.amazon.com/sagemaker/latest/dg/API_ResourceConfig.html) for details.

The `stopping_condition` block supports:

* `max_runtime_in_seconds` - (Optional) The maximum length of time in seconds that training algorithm is allowed to run. Defaults to 86,400 sec (24 hours.)

See [AWS SageMaker Stopping Condition doc](https://docs.aws.amazon.com/sagemaker/latest/dg/API_StoppingCondition.html) for details.

The `input_data_config` block supports a list of AWS SageMaker Channels each of which supports: ??TODO

* `name` - (Required) Name of the channel
* `data_source` - (Required) A structure with the details of the data location. Fields are documented below.
* `content_type` - (Optional) The MIME type of the data.
* `compression_type` - (Optional) If training data is compressed, the compression type. Defaults to `None`.
* `record_wrapper_type` - (Optional) Type of the record wrapper used for when reading the training data. Defaults to `None`.

See [AWS SageMaker Channel doc](https://docs.aws.amazon.com/sagemaker/latest/dg/API_Channel.html) for details.

The `data_source` block supports:

* `s3_data_source`  - (Optional) The S3 location of the data source that is associated with a channel. Fields are documented below.

See [AWS SageMaker Data Source doc](https://docs.aws.amazon.com/sagemaker/latest/dg/API_DataSource.html) for details.

The `s3_data_source` block supports:

* `s3_data_type`  - (Required) A string to indicate whether to a S# prefix or a manifest to create the list of S3 objects that contain the training data.
* `s3_uri` - (Required) Depending on the value specified for the `s3_data_type`, identifies either a key name prefix or a manifest.
* `s3_data_distribution_type` - (Required) A string to indicate whether the entire dataset or a subset of data must be replicated on each ML compute instance.

See [AWS SageMaker S3 Data Source doc](https://docs.aws.amazon.com/sagemaker/latest/dg/API_S3DataSource.html) for details.

The `output_data_config` block supports:

* `s3_output_path` - (Required) The S3 where the model artifacts will be stored in S3.
* `kms_key_id` - (Optional) The AWS Key Management Service (AWS KMS) key that Amazon SageMaker uses to encrypt the model artifacts at rest using Amazon S3 server-side encryption.

See [AWS SageMaker Output Data Config doc](https://docs.aws.amazon.com/sagemaker/latest/dg/API_OutputDataConfig.html) for details.


## Attributes Reference

The following attributes are exported:

* `id` - The name of the training jab.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this training job.
* `creation_time` - The creation timestamp of the training job.
* `last_modified_time` - The last modification timestamp of the training job.
* `training_start_time` - The timestamp when the training job started.
* `training_end_time` - The timestamp when the training job ended.
* `status` -  The status of the training job.
* `secondary_status` -  The secondary status of the training job.

## Import

Models can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_training_job.test_training_job my-training-job
```