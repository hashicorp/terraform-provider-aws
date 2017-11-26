---
layout: "aws"
page_title: "AWS: sagemaker_model"
sidebar_current: "docs-aws-resource-sagemaker-model"
description: |-
  Provides a Sagemaker model resource.
---

# aws\_sagemaker\_model

Provides a Sagemaker model resource.

## Example Usage

Basic usage:

```hcl
resource "aws_sagemaker_model" "m" {
    name = "my-model"

    primary_container {
        image = "111111111111.ecr.us-west-2.amazonaws.com/my-docker-image:latest"
    }
}
```

Usage with supplemental container and tags:

```hcl
resource "aws_sagemaker_model" "m" {
    name = "my-model"

    primary_container {
        image = "111111111111.ecr.us-west-2.amazonaws.com/my-docker-image:latest"
        model_data_url  = "s3://111111111111-foo/model.tar.gz"
    }

    supplemental_container {
        image = "111111111111.ecr.us-west-2.amazonaws.com/my-other-docker-image:latest"
        environment = [
            KeyName1 = "value"
            KeyName2 = "value"
        ]
    }

    tags {
        Name = "foo"
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the model (must be unique). If omitted, Terraform will assign a random, unique name.
* `primary_container` - (Required) Fields are documented below.
* `supplemental_container` - (Optional) Can be specified multiple times for each
   additional docker container to use. Each supplemental block supports the same fields as the primary container.
* `execution_role_arn` - (Optional) A role to with permissions that allows SageMaker to call other services on your behalf.
* `tags` - (Optional) A mapping of tags to assign to the resource.

The `primary_container` block supports:

* `image` - (Required) The registry path where the inference code image is stored in Amazon ECR.
* `model_data_url` - (Optional) The URL for the S3 location where model artifacts are stored.
* `container_hostname` - (Optional) The DNS host name for the container.
* `environment` - (Optional) Environment variables for the Docker container.
   A list of key value pairs.


## Attributes Reference

The following attributes are exported:

* `name` - The name of the model.
* `arn` - The Amazon Resource Name (ARN) assigned by AWS to this model.
* `creation_time` - The creation timestamp of the model.

## Import

Models can be imported using the `name`, e.g.

```
$ terraform import aws_sagemaker_model.test_model model-foo
```
