---
subcategory: "CloudWatch Logs"
layout: "aws"
page_title: "AWS: aws_cloudwatch_log_anomaly_detector"
description: |-
  Terraform resource for managing an AWS CloudWatch Log Anomaly Detector.
---

# Resource: aws_cloudwatch_log_anomaly_detector

Terraform resource for managing an AWS CloudWatch Logs Log Anomaly Detector.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatch_log_group" "test" {
  count = 2
  name  = "testing-${count.index}"
}

resource "aws_cloudwatch_log_anomaly_detector" "test" {
  detector_name           = "testing"
  log_group_arn_list      = [aws_cloudwatch_log_group.test[0].arn]
  anomaly_visibility_time = 7
  evaluation_frequency    = "TEN_MIN"
  enabled                 = "false"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `log_group_arn_list` - (Required) Array containing the ARN of the log group that this anomaly detector will watch. You can specify only one log group ARN.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `anomaly_visibility_time` - (Optional) Number of days to have visibility on an anomaly. After this time period has elapsed for an anomaly, it will be automatically baselined and the anomaly detector will treat new occurrences of a similar anomaly as normal. Therefore, if you do not correct the cause of an anomaly during the time period specified in `anomaly_visibility_time`, it will be considered normal going forward and will not be detected as an anomaly. Valid Range: Minimum value of 7. Maximum value of 90.

* `detector_name` - (Optional) Name for this anomaly detector.

* `evaluation_frequency` - (Optional) Specifies how often the anomaly detector is to run and look for anomalies. Set this value according to the frequency that the log group receives new logs. For example, if the log group receives new log events every 10 minutes, then 15 minutes might be a good setting for `evaluation_frequency`. Valid Values: `ONE_MIN | FIVE_MIN | TEN_MIN | FIFTEEN_MIN | THIRTY_MIN | ONE_HOUR`.

* `filter_pattern` - (Optional) You can use this parameter to limit the anomaly detection model to examine only log events that match the pattern you specify here. For more information, see [Filter and Pattern Syntax](https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html).

* `kms_key_id` - (Optional) Optionally assigns a AWS KMS key to secure this anomaly detector and its findings. If a key is assigned, the anomalies found and the model used by this detector are encrypted at rest with the key. If a key is assigned to an anomaly detector, a user must have permissions for both this key and for the anomaly detector to retrieve information about the anomalies that it finds.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` -  ARN of the log anomaly detector that you just created.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Logs Anomaly Detector using the `arn`. For example:

```terraform
import {
  to = aws_cloudwatch_log_anomaly_detector.example
  id = "log_anomaly_detector-arn-12345678"
}
```

Using `terraform import`, import CloudWatch Log Anomaly Detector using the `example_id_arg`. For example:

```console
% terraform import aws_cloudwatch_log_anomaly_detector.example log_anomaly_detector-arn-12345678
```
