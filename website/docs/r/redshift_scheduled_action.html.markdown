---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_scheduled_action"
description: |-
  Provides a Redshift Scheduled Action resource.
---

# Resource: aws_redshift_scheduled_action

## Example Usage

### Pause Cluster Action

```terraform
resource "aws_iam_role" "example" {
  name               = "redshift_scheduled_action"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": [
          "scheduler.redshift.amazonaws.com"
        ]
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "example" {
  name   = "redshift_scheduled_action"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Sid": "VisualEditor0",
          "Effect": "Allow",
          "Action": [
              "redshift:PauseCluster",
              "redshift:ResumeCluster",
              "redshift:ResizeCluster"
          ],
          "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "example" {
  policy_arn = aws_iam_policy.example.arn
  role       = aws_iam_role.example.name
}

resource "aws_redshift_scheduled_action" "example" {
  name     = "tf-redshift-scheduled-action"
  schedule = "cron(00 23 * * ? *)"
  iam_role = aws_iam_role.example.arn

  target_action {
    pause_cluster {
      cluster_identifier = "tf-redshift001"
    }
  }
}
```

### Resize Cluster Action

```terraform
resource "aws_redshift_scheduled_action" "example" {
  name     = "tf-redshift-scheduled-action"
  schedule = "cron(00 23 * * ? *)"
  iam_role = aws_iam_role.example.arn

  target_action {
    resize_cluster {
      cluster_identifier = "tf-redshift001"
      cluster_type       = "multi-node"
      node_type          = "dc1.large"
      number_of_nodes    = 2
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The scheduled action name.
* `description` - (Optional) The description of the scheduled action.
* `enable` - (Optional) Whether to enable the scheduled action. Default is `true` .
* `start_time` - (Optional) The start time in UTC when the schedule is active, in UTC RFC3339 format(for example, YYYY-MM-DDTHH:MM:SSZ).
* `end_time` - (Optional) The end time in UTC when the schedule is active, in UTC RFC3339 format(for example, YYYY-MM-DDTHH:MM:SSZ).
* `schedule` - (Required) The schedule of action. The schedule is defined format of "at expression" or "cron expression", for example `at(2016-03-04T17:27:00)` or `cron(0 10 ? * MON *)`. See [Scheduled Action](https://docs.aws.amazon.com/redshift/latest/APIReference/API_ScheduledAction.html) for more information.
* `iam_role` - (Required) The IAM role to assume to run the scheduled action.
* `target_action` - (Required) Target action. Documented below.

### Nested Blocks

#### `target_action`

* `pause_cluster` - (Optional) An action that runs a `PauseCluster` API operation. Documented below.
* `resize_cluster` - (Optional) An action that runs a `ResizeCluster` API operation. Documented below.
* `resume_cluster` - (Optional) An action that runs a `ResumeCluster` API operation. Documented below.

### `pause_cluster`

* `cluster_identifier` - (Required) The identifier of the cluster to be paused.

### `resize_cluster`

* `cluster_identifier` - (Required) The unique identifier for the cluster to resize.
* `classic` - (Optional) A boolean value indicating whether the resize operation is using the classic resize process. Default: `false`.
* `cluster_type` - (Optional)　The new cluster type for the specified cluster.
* `node_type` - (Optional) The new node type for the nodes you are adding.
* `number_of_nodes` - (Optional) The new number of nodes for the cluster.

### `resume_cluster`

* `cluster_identifier` - (Required) The identifier of the cluster to be resumed.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The Redshift Scheduled Action name.

## Import

Redshift Scheduled Action can be imported using the `name`, e.g.,

```
$ terraform import aws_redshift_scheduled_action.example tf-redshift-scheduled-action
```
