---
subcategory: "FIS (Fault Injection Simulator)"
layout: "aws"
page_title: "AWS: aws_fis_experiment_template"
description: |-
  Provides an FIS Experiment Template.
---

# Resource: aws_fis_experiment_template

Provides an FIS Experiment Template, which can be used to run an experiment.
An experiment template contains one or more actions to run on specified targets during an experiment.
It also contains the stop conditions that prevent the experiment from going out of bounds.
See [Amazon Fault Injection Simulator](https://docs.aws.amazon.com/fis/index.html)
for more information.

## Example Usage

```terraform
resource "aws_fis_experiment_template" "example" {
  description = "example"
  role_arn    = aws_iam_role.example.arn

  stop_condition {
    source = "none"
  }

  action {
    name      = "example-action"
    action_id = "aws:ec2:terminate-instances"

    target {
      key   = "Instances"
      value = "example-target"
    }
  }

  target {
    name           = "example-target"
    resource_type  = "aws:ec2:instance"
    selection_mode = "COUNT(1)"

    resource_tag {
      key   = "env"
      value = "example"
    }
  }
}
```

## Example Usage with Report Configuration

```terraform
data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = "example"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

data "aws_iam_policy_document" "report_access" {
  version = "2012-10-17"

  statement {
    sid       = "logsDelivery"
    effect    = "Allow"
    actions   = ["logs:CreateLogDelivery"]
    resources = ["*"]
  }

  statement {
    sid       = "ReportsBucket"
    effect    = "Allow"
    actions   = ["s3:PutObject", "s3:GetObject"]
    resources = ["*"]
  }

  statement {
    sid       = "GetDashboard"
    effect    = "Allow"
    actions   = ["cloudwatch:GetDashboard"]
    resources = ["*"]
  }

  statement {
    sid       = "GetDashboardData"
    effect    = "Allow"
    actions   = ["cloudwatch:getMetricWidgetImage"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "report_access" {
  name   = "report_access"
  policy = data.aws_iam_policy_document.report_access.json
}

resource "aws_iam_role_policy_attachment" "report_access" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.report_access.arn
}

resource "aws_fis_experiment_template" "example" {
  description = "example"
  role_arn    = aws_iam_role.example.arn

  stop_condition {
    source = "none"
  }

  action {
    name      = "example-action"
    action_id = "aws:ec2:terminate-instances"

    target {
      key   = "Instances"
      value = "example-target"
    }
  }

  target {
    name           = "example-target"
    resource_type  = "aws:ec2:instance"
    selection_mode = "COUNT(1)"

    resource_tag {
      key   = "env"
      value = "example"
    }
  }

  experiment_report_configuration {
    data_sources {
      cloudwatch_dashboard {
        dashboard_arn = aws_cloudwatch_dashboard.example.dashboard_arn
      }
    }

    outputs {
      s3_configuration {
        bucket_name = aws_s3_bucket.example.bucket
        prefix      = "fis-example-reports"
      }
    }

    post_experiment_duration = "PT10M"
    pre_experiment_duration  = "PT10M"
  }

  tags = {
    Name = "example"
  }
}
```

## Argument Reference

The following arguments are required:

* `action` - (Required) Action to be performed during an experiment. See below.
* `description` - (Required) Description for the experiment template.
* `role_arn` - (Required) ARN of an IAM role that grants the AWS FIS service permission to perform service actions on your behalf.
* `stop_condition` - (Required) When an ongoing experiment should be stopped. See below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `experiment_options` - (Optional) The experiment options for the experiment template. See [experiment_options](#experiment_options) below for more details!
* `tags` - (Optional) Key-value mapping of tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `target` - (Optional) Target of an action. See below.
* `log_configuration` - (Optional) The configuration for experiment logging. See below.
* `experiment_report_configuration` - (Optional) The configuration for [experiment reporting](https://docs.aws.amazon.com/fis/latest/userguide/experiment-report-configuration.html). See below.

### experiment_options

The `experiment_options` block supports the following:

* `account_targeting` - (Optional) Specifies the account targeting setting for experiment options. Supports `single-account` and `multi-account`.
* `empty_target_resolution_mode` - (Optional) Specifies the empty target resolution mode for experiment options. Supports `fail` and `skip`.

### `action`

* `action_id` - (Required) ID of the action. To find out what actions are supported see [AWS FIS actions reference](https://docs.aws.amazon.com/fis/latest/userguide/fis-actions-reference.html).
* `name` - (Required) Friendly name of the action.
* `description` - (Optional) Description of the action.
* `parameter` - (Optional) Parameter(s) for the action, if applicable. See below.
* `start_after` - (Optional) Set of action names that must complete before this action can be executed.
* `target` - (Optional) Action's target, if applicable. See below.

#### `parameter`

* `key` - (Required) Parameter name.
* `value` - (Required) Parameter value.

For a list of parameters supported by each action, see [AWS FIS actions reference](https://docs.aws.amazon.com/fis/latest/userguide/fis-actions-reference.html).

#### `target` (`action.*.target`)

* `key` - (Required) Target type. Valid values are `AutoScalingGroups` (EC2 Auto Scaling groups), `Buckets` (S3 Buckets), `Cluster` (EKS Cluster), `Clusters` (ECS Clusters), `DBInstances` (RDS DB Instances), `Instances` (EC2 Instances), `ManagedResources` (EKS clusters, Application and Network Load Balancers, and EC2 Auto Scaling groups that are enabled for ARC zonal shift), `Nodegroups` (EKS Node groups), `Pods` (EKS Pods), `ReplicationGroups`(ElastiCache Redis Replication Groups), `Roles` (IAM Roles), `SpotInstances` (EC2 Spot Instances), `Subnets` (VPC Subnets), `Tables` (DynamoDB encrypted global tables), `Tasks` (ECS Tasks), `TransitGateways` (Transit gateways), `Volumes` (EBS Volumes). See the [documentation](https://docs.aws.amazon.com/fis/latest/userguide/actions.html#action-targets) for more details.
* `value` - (Required) Target name, referencing a corresponding target.

### `stop_condition`

* `source` - (Required) Source of the condition. One of `none`, `aws:cloudwatch:alarm`.
* `value` - (Optional) ARN of the CloudWatch alarm. Required if the source is a CloudWatch alarm.

### `target`

* `name` - (Required) Friendly name given to the target.
* `resource_type` - (Required) AWS resource type. The resource type must be supported for the specified action. To find out what resource types are supported, see [Targets for AWS FIS](https://docs.aws.amazon.com/fis/latest/userguide/targets.html#resource-types).
* `selection_mode` - (Required) Scopes the identified resources. Valid values are `ALL` (all identified resources), `COUNT(n)` (randomly select `n` of the identified resources), `PERCENT(n)` (randomly select `n` percent of the identified resources).
* `filter` - (Optional) Filter(s) for the target. Filters can be used to select resources based on specific attributes returned by the respective describe action of the resource type. For more information, see [Targets for AWS FIS](https://docs.aws.amazon.com/fis/latest/userguide/targets.html#target-filters). See below.
* `resource_arns` - (Optional) Set of ARNs of the resources to target with an action. Conflicts with `resource_tag`.
* `resource_tag` - (Optional) Tag(s) the resources need to have to be considered a valid target for an action. Conflicts with `resource_arns`. See below.
* `parameters` - (Optional) The resource type parameters.

~> **NOTE:** The `target` configuration block requires either `resource_arns` or `resource_tag`.

#### `filter`

* `path` - (Required) Attribute path for the filter.
* `values` - (Required) Set of attribute values for the filter.

~> **NOTE:** Values specified in a `filter` are joined with an `OR` clause, while values across multiple `filter` blocks are joined with an `AND` clause. For more information, see [Targets for AWS FIS](https://docs.aws.amazon.com/fis/latest/userguide/targets.html#target-filters).

#### `resource_tag`

* `key` - (Required) Tag key.
* `value` - (Required) Tag value.

### `log_configuration`

* `log_schema_version` - (Required) The schema version. See [documentation](https://docs.aws.amazon.com/fis/latest/userguide/monitoring-logging.html#experiment-log-schema) for the list of schema versions.
* `cloudwatch_logs_configuration` - (Optional) The configuration for experiment logging to Amazon CloudWatch Logs. See below.
* `s3_configuration` - (Optional) The configuration for experiment logging to Amazon S3. See below.

#### `cloudwatch_logs_configuration`

* `log_group_arn` - (Required) The Amazon Resource Name (ARN) of the destination Amazon CloudWatch Logs log group.

#### `s3_configuration`

* `bucket_name` - (Required) The name of the destination bucket.
* `prefix` - (Optional) The bucket prefix.

### `experiment_report_configuration`

* `data_sources` - (Required) The data sources for the experiment report. See below.
* `outputs` - (Required) The outputs for the experiment report. See below.
* `post_experiment_duration` - (Optional) The duration of the post-experiment period. Defaults to `PT20M`.
* `pre_experiment_duration` - (Optional) The duration of the pre-experiment period. Defaults to `PT20M`.

#### `data_sources`

* `cloudwatch_dashboard` - (Required) The data sources for the experiment report. See below.

#### `cloudwatch_dashboard`

* `dashboard_arn` - (Required) The ARN of the CloudWatch dashboard.

#### `outputs`

* `s3_configuration` - (Required) The data sources for the experiment report. See below.

#### `s3_configuration`

* `bucket_name` - (Required) The name of the destination bucket.
* `prefix` - (Optional) The bucket prefix.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Experiment Template ID.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import FIS Experiment Templates using the `id`. For example:

```terraform
import {
  to = aws_fis_experiment_template.template
  id = "EXT123AbCdEfGhIjK"
}
```

Using `terraform import`, import FIS Experiment Templates using the `id`. For example:

```console
% terraform import aws_fis_experiment_template.template EXT123AbCdEfGhIjK
```
