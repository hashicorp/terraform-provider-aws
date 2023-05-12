---
subcategory: "EventBridge Pipes"
layout: "aws"
page_title: "AWS: aws_pipes_pipe"
description: |-
  Terraform resource for managing an AWS EventBridge Pipes Pipe.
---

# Resource: aws_pipes_pipe

Terraform resource for managing an AWS EventBridge Pipes Pipe.

You can find out more about EventBridge Pipes in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes.html).

~> **Note:** EventBridge was formerly known as CloudWatch Events. The functionality is identical.

## Example Usage

### Basic Usage

```terraform
data "aws_caller_identity" "main" {}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = "sts:AssumeRole"
      Principal = {
        Service = "pipes.amazonaws.com"
      }
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.main.account_id
        }
      }
    }
  })
}

resource "aws_iam_role_policy" "source" {
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ReceiveMessage",
        ],
        Resource = [
          aws_sqs_queue.source.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "source" {}

resource "aws_iam_role_policy" "target" {
  role = aws_iam_role.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
        ],
        Resource = [
          aws_sqs_queue.target.arn,
        ]
      },
    ]
  })
}

resource "aws_sqs_queue" "target" {}

resource "aws_pipes_pipe" "example" {
  depends_on = [aws_iam_role_policy.source, aws_iam_role_policy.target]
  name       = "example-pipe"
  role_arn   = aws_iam_role.example.arn
  source     = aws_sqs_queue.source.arn
  target     = aws_sqs_queue.target.arn

  source_parameters {}
  target_parameters {}
}
```

## Argument Reference

The following arguments are required:

* `role_arn` - (Required) ARN of the role that allows the pipe to send data to the target.
* `source` - (Required) Source resource of the pipe (typically an ARN).
* `target` - (Required) Target resource of the pipe (typically an ARN).
* `source_parameters` - (Required) Parameters required to set up a source for the pipe. Detailed below.
* `target_parameters` - (Required) Parameters required to set up a target for your pipe. Detailed below.

The following arguments are optional:

* `description` - (Optional) A description of the pipe. At most 512 characters.
* `desired_state` - (Optional) The state the pipe should be in. One of: `RUNNING`, `STOPPED`.
* `enrichment` - (Optional) Enrichment resource of the pipe (typically an ARN). Read more about enrichment in the [User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-pipes.html#pipes-enrichment).
* `name` - (Optional) Name of the pipe. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `tags` - (Optional) Key-value mapping of resource tags. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### source_parameters Configuration Block

* `filter_criteria` - (Optional) The collection of event patterns used to filter events. Detailed below.

#### source_parameters.filter_criteria Configuration Block

* `filter` - (Optional) An array of up to 5 event patterns. Detailed below.

##### source_parameters.filter_criteria.filter Configuration Block

* `pattern` - (Required) The event pattern. At most 4096 characters.

### target_parameters Configuration Block

* `input_template` - (Optional) Valid JSON text passed to the target. In this case, nothing from the event itself is passed to the target.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of this pipe.
* `id` - Same as `name`.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

Pipes can be imported using the `name`. For example:

```
$ terraform import aws_pipes_pipe.example my-pipe
```
