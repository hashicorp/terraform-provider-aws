---
subcategory: "Glue"
layout: "aws"
page_title: "AWS: aws_glue_dev_endpoint"
description: |-
  Provides a Glue Development Endpoint resource.
---

# aws_glue_dev_endpoint

Provides a Glue Development Endpoint resource.

## Example Usage

Basic usage:

```hcl
resource "aws_glue_dev_endpoint" "de" {
  name     = "foo"
  role_arn = aws_iam_role.de.arn
}

resource "aws_iam_role" "de" {
  name               = "AWSGlueServiceRole-foo"
  assume_role_policy = data.aws_iam_policy_document.de.json
}

data "aws_iam_policy_document" "de" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["glue.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "foo-AWSGlueServiceRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.de.name
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of this endpoint (must be unique). If omitted, Terraform will assign a random, unique name.
* `extra_jars_s3_path` - (Optional) Path to one or more Java Jars in an S3 bucket that should be loaded in this endpoint.
* `extra_python_libs_s3_path` - (Optional) Path(s) to one or more Python libraries in an S3 bucket that should be loaded in this endpoint. Multiple values must be complete paths separated by a comma.
* `glue_version` - (Optional) -  Specifies the versions of Python and Apache Spark to use. Defaults to AWS Glue version 0.9.
* `number_of_nodes` - (Optional) The number of AWS Glue Data Processing Units (DPUs) to allocate to this endpoint. Conflicts with `worker_type`.
* `number_of_workers` - (Optional) The number of workers of a defined worker type that are allocated to this endpoint. This field is available only when you choose worker type G.1X or G.2X.
* `public_key` - (Optional) The public key to be used by this endpoint for authentication.
* `public_keys` - (Optional) A list of public keys to be used by this endpoint for authentication.
* `role_arn` - (Required) The IAM role for this endpoint.
* `security_configuration` - (Optional) The name of the Security Configuration structure to be used with this endpoint.
* `security_group_ids` - (Optional) Security group IDs for the security groups to be used by this endpoint.
* `subnet_id` - (Optional) The subnet ID for the new endpoint to use.
* `worker_type` - (Optional) The type of predefined worker that is allocated to this endpoint. Accepts a value of Standard, G.1X, or G.2X.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the new Glue Development Endpoint.
* `private_address` - A private IP address to access the Glue Development Endpoint within a VPC, if this endpoint is created within one.
* `public_address` - The public IP address used by this endpoint. The PublicAddress field is present only when you create a non-VPC endpoint.
* `yarn_endpoint_address` - The YARN endpoint address used by this endpoint.
* `zeppelin_remote_spark_interpreter_port` - The Apache Zeppelin port for the remote Apache Spark interpreter.
* `availability_zone` - The AWS availability zone where this endpoint is located.
* `vpc_id` - he ID of the VPC used by this endpoint.
* `status` - The current status of this endpoint.
* `failure_reason` - The reason for a current failure in this endpoint.

## Import

Glue Development Endpoint can be imported using the `name`, e.g.

```
$ terraform import aws_glue_dev_endpoint.de foo
```