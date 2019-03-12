---
layout: "aws"
page_title: "AWS: aws_glue_dev_endpoint"
sidebar_current: "docs-aws-resource-glue-dev-endpoint"
description: |-
  Provides a Glue Development Endpoint resource.
---

# aws_glue_dev_endpoint

Provides a Glue Development Endpoint resource.

## Example Usage

Basic usage:

```hcl
resource "aws_glue_dev_endpoint" "de" {
  name = "foo"
  role_arn = "${aws_iam_role.test.arn}"
}

resource "aws_iam_role" "de" {
  name = "AWSGlueServiceRole-foo"
  assume_role_policy = "${data.aws_iam_policy_document.de.json}"
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
  role       = "${aws_iam_role.de.name}"
}

```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the Glue Development Endpoint (must be unique). If omitted, Terraform will assign a random, unique name.
* `extra_jars_s3_path` - (Optional) Path to one or more Java Jars in an S3 bucket that should be loaded in your DevEndpoint.
* `extra_python_libs_s3_path` - (Optional) Path(s) to one or more Python libraries in an S3 bucket that should be loaded in your DevEndpoint. Multiple values must be complete paths separated by a comma.
* `number_of_nodes` - (Optional) The number of AWS Glue Data Processing Units (DPUs) to allocate to this DevEndpoint.
* `public_key` - (Optional) The public key to be used by this DevEndpoint for authentication.
* `public_keys` - (Optional) A list of public keys to be used by the DevEndpoints for authentication.
* `role_arn` - (Required) The IAM role for the DevEndpoint.
* `security_configuration` - (Optional) The name of the SecurityConfiguration structure to be used with this DevEndpoint.
* `security_group_ids` - (Optional) Security group IDs for the security groups to be used by the new DevEndpoint.
* `subnet_id` - (Optional) The subnet ID for the new DevEndpoint to use.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the new Glue Development Endpoint.
* `private_address` - A private IP address to access the DevEndpoint within a VPC, if the DevEndpoint is created within one.
* `public_address` - The public IP address used by this DevEndpoint. The PublicAddress field is present only when you create a non-VPC DevEndpoint.
* `yarn_endpoint_address` - The YARN endpoint address used by this DevEndpoint.
* `zeppelin_remote_spark_interpreter_port` - The Apache Zeppelin port for the remote Apache Spark interpreter.
* `availability_zone` - The AWS availability zone where this DevEndpoint is located.
* `vpc_id` - he ID of the VPC used by this DevEndpoint.
* `status` - The current status of this DevEndpoint.
* `failure_reason` - The reason for a current failure in this DevEndpoint.

## Import

SageMaker Glue Development Endpoint can be imported using the `name`, e.g.

```
$ terraform import aws_glue_dev_endpoint.de foo
```