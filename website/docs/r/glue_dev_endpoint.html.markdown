---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_dev_endpoint"
description: |-
  Provides a Glue Development Endpoint resource.
---

# Resource: aws_glue_dev_endpoint

Provides a Glue Development Endpoint resource.

## Example Usage

Basic usage:

```terraform
resource "aws_glue_dev_endpoint" "example" {
  name     = "foo"
  role_arn = aws_iam_role.example.arn
}

resource "aws_iam_role" "example" {
  name               = "AWSGlueServiceRole-foo"
  assume_role_policy = data.aws_iam_policy_document.example.json
}

data "aws_iam_policy_document" "example" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["glue.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "example-AWSGlueServiceRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.example.name
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `arguments` - (Optional) A map of arguments used to configure the endpoint.
* `extra_jars_s3_path` - (Optional) Path to one or more Java Jars in an S3 bucket that should be loaded in this endpoint.
* `extra_python_libs_s3_path` - (Optional) Path(s) to one or more Python libraries in an S3 bucket that should be loaded in this endpoint. Multiple values must be complete paths separated by a comma.
* `glue_version` - (Optional) -  Specifies the versions of Python and Apache Spark to use. Defaults to AWS Glue version 0.9.
* `name` - (Required) The name of this endpoint. It must be unique in your account.
* `number_of_nodes` - (Optional) The number of AWS Glue Data Processing Units (DPUs) to allocate to this endpoint. Conflicts with `worker_type`.
* `number_of_workers` - (Optional) The number of workers of a defined worker type that are allocated to this endpoint. This field is available only when you choose worker type G.1X or G.2X.
* `public_key` - (Optional) The public key to be used by this endpoint for authentication.
* `public_keys` - (Optional) A list of public keys to be used by this endpoint for authentication.
* `role_arn` - (Required) The IAM role for this endpoint.
* `security_configuration` - (Optional) The name of the Security Configuration structure to be used with this endpoint.
* `security_group_ids` - (Optional) Security group IDs for the security groups to be used by this endpoint.
* `subnet_id` - (Optional) The subnet ID for the new endpoint to use.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `worker_type` - (Optional) The type of predefined worker that is allocated to this endpoint. Accepts a value of Standard, G.1X, or G.2X.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the endpoint.
* `name` - The name of the new endpoint.
* `private_address` - A private IP address to access the endpoint within a VPC, if this endpoint is created within one.
* `public_address` - The public IP address used by this endpoint. The PublicAddress field is present only when you create a non-VPC endpoint.
* `yarn_endpoint_address` - The YARN endpoint address used by this endpoint.
* `zeppelin_remote_spark_interpreter_port` - The Apache Zeppelin port for the remote Apache Spark interpreter.
* `availability_zone` - The AWS availability zone where this endpoint is located.
* `vpc_id` - he ID of the VPC used by this endpoint.
* `status` - The current status of this endpoint.
* `failure_reason` - The reason for a current failure in this endpoint.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a Glue Development Endpoint using the `name`. For example:

```terraform
import {
  to = aws_glue_dev_endpoint.example
  id = "foo"
}
```

Using `terraform import`, import a Glue Development Endpoint using the `name`. For example:

```console
% terraform import aws_glue_dev_endpoint.example foo
```
