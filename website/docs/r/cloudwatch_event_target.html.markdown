---
layout: "aws"
page_title: "AWS: aws_cloudwatch_event_target"
sidebar_current: "docs-aws-resource-cloudwatch-event-target"
description: |-
  Provides a CloudWatch Event Target resource.
---

# aws_cloudwatch_event_target

Provides a CloudWatch Event Target resource.

## Example Usage

```hcl
resource "aws_cloudwatch_event_target" "yada" {
  target_id = "Yada"
  rule      = "${aws_cloudwatch_event_rule.console.name}"
  arn       = "${aws_kinesis_stream.test_stream.arn}"

  run_command_targets {
    key = "tag:Name"
    values = ["FooBar"]
  }

  run_command_targets {
    key = "InstanceIds"
    values = ["i-162058cd308bffec2"]
  }
}

resource "aws_cloudwatch_event_rule" "console" {
  name        = "capture-ec2-scaling-events"
  description = "Capture all EC2 scaling events"

  event_pattern = <<PATTERN
{
  "source": [
    "aws.autoscaling"
  ],
  "detail-type": [
    "EC2 Instance Launch Successful",
    "EC2 Instance Terminate Successful",
    "EC2 Instance Launch Unsuccessful",
    "EC2 Instance Terminate Unsuccessful"
  ]
}
PATTERN
}

resource "aws_kinesis_stream" "test_stream" {
  name        = "terraform-kinesis-test"
  shard_count = 1
}
```

## Argument Reference

-> **Note:** `input` and `input_path` are mutually exclusive options.

-> **Note:** In order to be able to have your AWS Lambda function or
   SNS topic invoked by a CloudWatch Events rule, you must setup the right permissions
   using [`aws_lambda_permission`](https://www.terraform.io/docs/providers/aws/r/lambda_permission.html)
   or [`aws_sns_topic.policy`](https://www.terraform.io/docs/providers/aws/r/sns_topic.html#policy).
   More info [here](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/EventsResourceBasedPermissions.html).

The following arguments are supported:

* `rule` - (Required) The name of the rule you want to add targets to.
* `target_id` - (Optional) The unique target assignment ID.  If missing, will generate a random, unique id.
* `arn` - (Required) The Amazon Resource Name (ARN) associated of the target.
* `input` - (Optional) Valid JSON text passed to the target.
* `input_path` - (Optional) The value of the [JSONPath](http://goessner.net/articles/JsonPath/)
	that is used for extracting part of the matched event when passing it to the target.
* `role_arn` - (Optional) The Amazon Resource Name (ARN) of the IAM role to be used for this target when the rule is triggered. Required if `ecs_target` is used.
* `run_command_targets` - (Optional) Parameters used when you are using the rule to invoke Amazon EC2 Run Command. Documented below. A maximum of 5 are allowed.
* `ecs_target` - (Optional) Parameters used when you are using the rule to invoke Amazon ECS Task. Documented below. A maximum of 1 are allowed.
* `input_transformer` - (Optional) Parameters used when you are providing a custom input to a target based on certain event data.

`run_command_targets` support the following:

* `key` - (Required) Can be either `tag:tag-key` or `InstanceIds`.
* `values` - (Required) If Key is `tag:tag-key`, Values is a list of tag values. If Key is `InstanceIds`, Values is a list of Amazon EC2 instance IDs.

`ecs_target` support the following:

* `task_count` - (Optional) The number of tasks to create based on the TaskDefinition. The default is 1.
* `task_definition_arn` - (Required) The ARN of the task definition to use if the event target is an Amazon ECS cluster.

`input_transformer` support the following:

* `input_paths` - (Optional) Key value pairs specified in the form of JSONPath (for example, time = $.time)
* `input_template` - (Required) Structure containing the template body.
