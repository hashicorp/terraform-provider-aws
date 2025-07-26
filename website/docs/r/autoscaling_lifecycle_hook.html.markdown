---
subcategory: "Auto Scaling"
layout: "aws"
page_title: "AWS: aws_autoscaling_lifecycle_hook"
description: |-
  Provides an AutoScaling Lifecycle Hook resource.
---

# Resource: aws_autoscaling_lifecycle_hook

Provides an AutoScaling Lifecycle Hook resource.

~> **NOTE:** Terraform has two types of ways you can add lifecycle hooks - via
the `initial_lifecycle_hook` attribute from the
[`aws_autoscaling_group`](/docs/providers/aws/r/autoscaling_group.html)
resource, or via this one. Hooks added via this resource will not be added
until the autoscaling group has been created, and depending on your
[capacity](/docs/providers/aws/r/autoscaling_group.html#waiting-for-capacity)
settings, after the initial instances have been launched, creating unintended
behavior. If you need hooks to run on all instances, add them with
`initial_lifecycle_hook` in
[`aws_autoscaling_group`](/docs/providers/aws/r/autoscaling_group.html),
but take care to not duplicate those hooks with this resource.

## Example Usage

```terraform
resource "aws_autoscaling_group" "foobar" {
  availability_zones   = ["us-west-2a"]
  name                 = "terraform-test-foobar5"
  health_check_type    = "EC2"
  termination_policies = ["OldestInstance"]

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_lifecycle_hook" "foobar" {
  name                   = "foobar"
  autoscaling_group_name = aws_autoscaling_group.foobar.name
  default_result         = "CONTINUE"
  heartbeat_timeout      = 2000
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_LAUNCHING"

  notification_metadata = jsonencode({
    foo = "bar"
  })

  notification_target_arn = "arn:aws:sqs:us-east-1:444455556666:queue1*"
  role_arn                = "arn:aws:iam::123456789012:role/S3Access"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the lifecycle hook.
* `autoscaling_group_name` - (Required) Name of the Auto Scaling group to which you want to assign the lifecycle hook
* `default_result` - (Optional) Defines the action the Auto Scaling group should take when the lifecycle hook timeout elapses or if an unexpected failure occurs. The value for this parameter can be either CONTINUE or ABANDON. The default value for this parameter is ABANDON.
* `heartbeat_timeout` - (Optional) Defines the amount of time, in seconds, that can elapse before the lifecycle hook times out. When the lifecycle hook times out, Auto Scaling performs the action defined in the DefaultResult parameter
* `lifecycle_transition` - (Required) Instance state to which you want to attach the lifecycle hook. For a list of lifecycle hook types, see [describe-lifecycle-hook-types](https://docs.aws.amazon.com/cli/latest/reference/autoscaling/describe-lifecycle-hook-types.html#examples)
* `notification_metadata` - (Optional) Contains additional information that you want to include any time Auto Scaling sends a message to the notification target.
* `notification_target_arn` - (Optional) ARN of the notification target that Auto Scaling will use to notify you when an instance is in the transition state for the lifecycle hook. This ARN target can be either an SQS queue or an SNS topic.
* `role_arn` - (Optional) ARN of the IAM role that allows the Auto Scaling group to publish to the specified notification target.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AutoScaling Lifecycle Hooks using the role autoscaling_group_name and name separated by `/`. For example:

```terraform
import {
  to = aws_autoscaling_lifecycle_hook.test-lifecycle-hook
  id = "asg-name/lifecycle-hook-name"
}
```

Using `terraform import`, import AutoScaling Lifecycle Hooks using the role autoscaling_group_name and name separated by `/`. For example:

```console
% terraform import aws_autoscaling_lifecycle_hook.test-lifecycle-hook asg-name/lifecycle-hook-name
```
