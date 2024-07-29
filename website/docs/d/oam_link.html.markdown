---
subcategory: "CloudWatch Observability Access Manager"
layout: "aws"
page_title: "AWS: aws_oam_link"
description: |-
  Terraform data source for managing an AWS CloudWatch Observability Access Manager Link.
---

# Data Source: aws_oam_link

Terraform data source for managing an AWS CloudWatch Observability Access Manager Link.

## Example Usage

### Basic Usage

```terraform
data "aws_oam_link" "example" {
  link_identifier = "arn:aws:oam:us-west-1:111111111111:link/abcd1234-a123-456a-a12b-a123b456c789"
}
```

## Argument Reference

The following arguments are required:

* `link_identifier` - (Required) ARN of the link.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the link.
* `id` - ARN of the link.
* `label` - Label that is assigned to this link.
* `label_template` - Human-readable name used to identify this source account when you are viewing data from it in the monitoring account.
* `link_configuration` - Configuration for creating filters that specify that only some metric namespaces or log groups are to be shared from the source account to the monitoring account. See [`link_configuration` Block](#link_configuration-block) for details.
* `link_id` - ID string that AWS generated as part of the link ARN.
* `resource_types` - Types of data that the source account shares with the monitoring account.
* `sink_arn` - ARN of the sink that is used for this link.

### `link_configuration` Block

The `link_configuration` configuration block supports the following arguments:

* `log_group_configuration` - Configuration for filtering which log groups are to send log events from the source account to the monitoring account. See [`log_group_configuration` Block](#log_group_configuration-block) for details.
* `metric_configuration` - Configuration for filtering which metric namespaces are to be shared from the source account to the monitoring account. See [`metric_configuration` Block](#metric_configuration-block) for details.

### `log_group_configuration` Block

The `log_group_configuration` configuration block supports the following arguments:

* `filter` - Filter string that specifies which log groups are to share their log events with the monitoring account. See [LogGroupConfiguration](https://docs.aws.amazon.com/OAM/latest/APIReference/API_LogGroupConfiguration.html) for details.

### `metric_configuration` Block

The `metric_configuration` configuration block supports the following arguments:

* `filter` - Filter string that specifies  which metrics are to be shared with the monitoring account. See [MetricConfiguration](https://docs.aws.amazon.com/OAM/latest/APIReference/API_MetricConfiguration.html) for details.
