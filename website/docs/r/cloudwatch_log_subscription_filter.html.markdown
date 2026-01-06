---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_subscription_filter"
description: |-
  Provides a CloudWatch Logs subscription filter.
---

# Resource: aws_cloudwatch_log_subscription_filter

Provides a CloudWatch Logs subscription filter resource.

## Example Usage

```terraform
resource "aws_cloudwatch_log_subscription_filter" "test_lambdafunction_logfilter" {
  name            = "test_lambdafunction_logfilter"
  role_arn        = aws_iam_role.iam_for_lambda.arn
  log_group_name  = "/aws/lambda/example_lambda_name"
  filter_pattern  = "logtype test"
  destination_arn = aws_kinesis_stream.test_logstream.arn
  distribution    = "Random"
}
```

## Argument Reference

This resource supports the following arguments:

* `destination_arn` - (Required) ARN of the destination to deliver matching log events to. Kinesis stream or Lambda function ARN.
* `distribution` - (Optional) Method used to distribute log data to the destination. By default log data is grouped by log stream, but the grouping can be set to random for a more even distribution. This property is only applicable when the destination is an Amazon Kinesis stream. Valid values are "Random" and "ByLogStream".
* `emit_system_fields` - (Optional) List of system fields to include in the log events sent to the subscription destination. These fields provide source information for centralized log data in the forwarded payload. Valid values: `"@aws.account"`, `"@aws.region"`. To remove this argument after it has been set, specify an empty list `[]` explicitly to avoid perpetual differences.
* `filter_pattern` - (Required) Valid CloudWatch Logs filter pattern for subscribing to a filtered stream of log events. Use empty string `""` to match everything. For more information, see the [Amazon CloudWatch Logs User Guide](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html).
* `log_group_name` - (Required) Name of the log group to associate the subscription filter with.
* `name` - (Required) Name for the subscription filter.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `apply_on_transformed_logs` - (Optional) Boolean to indicate whether to apply the subscription filter on the transformed version of the log events instead of the original ingested log events. Defaults to `false`. Valid only for log groups that have an active log transformer. 
* `role_arn` - (Optional) ARN of an IAM role that grants CloudWatch Logs permissions to deliver ingested log events to the destination stream. You don't need to provide the ARN when you are working with a logical destination for cross-account delivery. If you use Lambda as a destination, you should skip this argument and use `aws_lambda_permission` resource for granting access from CloudWatch logs to the destination Lambda function.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs subscription filter using the log group name and subscription filter name separated by `|`. For example:

```terraform
import {
  to = aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter
  id = "/aws/lambda/example_lambda_name|test_lambdafunction_logfilter"
}
```

Using `terraform import`, import CloudWatch Logs subscription filter using the log group name and subscription filter name separated by `|`. For example:

```console
% terraform import aws_cloudwatch_log_subscription_filter.test_lambdafunction_logfilter "/aws/lambda/example_lambda_name|test_lambdafunction_logfilter"
```
