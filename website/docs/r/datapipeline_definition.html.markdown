---
layout: "aws"
page_title: "AWS: aws_datapipeline_definition"
sidebar_current: "docs-aws-resource-datapipeline-definition"
description: |-
  Provides a AWS DataPipeline Definition.
---

# Resource: aws_datapipeline_definition

Provides a Data Pipeline Definition resource.

## Example Usage

```hcl
resource "aws_iam_role" "role" {
  name = "tf-test-datapipeline-role-%[1]s"
      
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "elasticmapreduce.amazonaws.com",
          "datapipeline.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

data "aws_iam_policy" "role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSDataPipelineRole"
}

resource "aws_iam_role_policy_attachment" "role" {
  role       = "${aws_iam_role.role.name}"
  policy_arn = "${data.aws_iam_policy.role.arn}"
}

resource "aws_iam_role" "resource_role" {
  name = "tf-test-datapipeline-resource-role-%[1]s"
      
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

data "aws_iam_policy" "resource_role" {
  arn = "arn:aws:iam::aws:policy/service-role/AWSDataPipelineRole"
}

resource "aws_iam_role_policy_attachment" "resource_role" {
  role       = "${aws_iam_role.resource_role.name}"
  policy_arn = "${data.aws_iam_policy.resource_role.arn}"
}

resource "aws_iam_instance_profile" "resource_role" {
  name = "tf-test-datapipeline-resource-role-profile-%[1]s"
  role = "${aws_iam_role.resource_role.name}"
}

resource "aws_datapipeline_pipeline" "default" {
	name      	= "tf-pipeline-default"
}

resource "aws_datapipeline_definition" "default" {
  pipeline_id = "${aws_datapipeline_pipeline.default.id}"

  default {
	schedule_type = "ondemand"
	role          = "${aws_iam_role.role.arn}"
	resource_role = "${aws_iam_instance_profile.resource_role.arn}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `pipeline_id` - (Required) The id of Pipeline.
* `default` - (Required) The default configuration to assign to the resource. Documented below.
* `schedule` - (Optional) A list of schedule configuration to assign to the resource. Defines the timing of a scheduled event, such as when an activity runs. Documented below.
* `parameter_object` - (Optional) A list of parameter object configuration to assign to the resource. Documented below.
* `parameter_value` - (Optional) A list of parameter value configuration to assign to the resource. Documented below.

The `default` configuration supports the following:

* `resource_role` - (Required) The arn of iam instance profile attached to resources.
* `role` - (Required) The arn of iam role, execute pipeline.
* `schedule_type` - (Required) The parameters for the RUN_COMMAND task execution. Valid values are `cron`, `ondemand` and `none`. Default to `none`.
* `failure_and_rerun_mode` - (Optional) The date and time to start the scheduled runs. Valid values are `cascade` and `timeseries`.
* `pipeline_log_uri` - (Optional) The s3 bucket uri for default to output logs. Example: `s3://<bucket-name>/<prefix>`.
* `schedule` - (Optional) The id of schedule pipeline object.

The `schedule` configuration supports the following:
For more information, see the [Schedule - AWS Data Pipeline](https://docs.aws.amazon.com/datapipeline/latest/DeveloperGuide/dp-object-schedule.html).

* `id` - (Required) The ID of schedule pipeline object.
* `name` - (Required) The Name of schedule pipeline object.
* `period` - (Required) How often the pipeline should run. The format is "N [minutes|hours|days|weeks|months]", where N is a number followed by one of the time specifiers.
* `start_at` - (Optional) The date and time at which to start the scheduled pipeline runs. Valid value is `FIRST_ACTIVATION_DATE_TIME`, which is deprecated in favor of creating an on-demand pipeline. Conflicts with `start_date_time`.
* `start_date_time` - (Optional) The date and time to start the scheduled runs. Conflicts with `start_at`.
* `end_date_time` - (Optional) The date and time to end the scheduled runs. Must be a date and time later than the value of startDateTime or startAt. The default behavior is to schedule runs until the pipeline is shut down. Conflicts with `occurrences`.
* `occurrences` - (Optional) The number of times to execute the pipeline after it's activated. Conflicts with `end_date_time`.
* `parent` - (Optional) Parent of the current object from which slots will be inherited.

The `parameter_object` configuration supports the following:
For more information, see the [ParameterObject - AWS Data Pipeline](https://docs.aws.amazon.com/ja_jp/datapipeline/latest/APIReference/API_ParameterObject.html).

* `id` - (Required) The ID of the parameter object.
* `description` - (Optional) The parameters for the RUN_COMMAND task execution. Documented below.

`parameter_value` supports the following:
For more information, see the [ParameterValue - AWS Data Pipeline](https://docs.aws.amazon.com/ja_jp/datapipeline/latest/APIReference/API_ParameterValue.html).

* `id` - (Required) The ID of the parameter value.
* `string_value` - (Required) The field value, expressed as a String.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The identifier of the pipeline.

## Import

`aws_datapipeline_definition` can be imported by using the id (Pipeline ID), e.g.

```
$ terraform import aws_datapipeline_definition.default df-1234567890
```