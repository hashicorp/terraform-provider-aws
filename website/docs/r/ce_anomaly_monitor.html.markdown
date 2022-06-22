---
subcategory: "CE (Cost Explorer)"
layout: "aws"
page_title: "AWS: aws_ce_anomaly_monitor"
description: |-
  Provides a CE Cost Anomaly Monitor
---

# Resource: aws_ce_anomaly_monitor

Provides a CE Anomaly Monitor.

## Example Usage

There are two main types of a Cost Anomaly Monitor: `DIMENSIONAL` and `CUSTOM`.

### Dimensional Example

```terraform
resource "aws_ce_anomaly_monitor" "service_monitor" {
  name              = "AWSServiceMonitor"
  monitor_type      = "DIMENSIONAL"
  monitor_dimension = "SERVICE"
}
```

### Custom Example

```terraform
resource "aws_ce_anomaly_monitor" "test" {
  name         = "AWSCustomAnomalyMonitor"
  monitor_type = "CUSTOM"

  specification = <<JSON
{
	"And": null,
	"CostCategories": null,
	"Dimensions": null,
	"Not": null,
	"Or": null,
	"Tags": {
		"Key": "CostCenter",
		"MatchOptions": null,
		"Values": [
			"10000"
		]
	}
}
JSON
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) The name of the monitor.
* `monitor_type` - (Required) The possible type values. Valid values: `DIMENSIONAL` | `CUSTOM`.
* `monitor_dimension` - (Required, if `monitor_type` is `DIMENSIONAL`) The dimensions to evaluate. Valid values: `SERVICE`.
* `monitor_specification` - (Required, if `monitor_type` is `CUSTOM`) A valid JSON representation for the [Expression](https://docs.aws.amazon.com/aws-cost-management/latest/APIReference/API_Expression.html) object.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the anomaly monitor.
* `id` - Unique ID of the anomaly monitor. Same as `arn`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block).

## Import

`aws_ce_anomaly_monitor` can be imported using the `id`, e.g.

```
$ terraform import aws_ce_anomaly_monitor.example costAnomalyMonitorARN
```
