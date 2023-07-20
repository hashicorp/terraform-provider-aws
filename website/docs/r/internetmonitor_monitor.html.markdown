---
subcategory: "CloudWatch Internet Monitor"
layout: "aws"
page_title: "AWS: aws_internetmonitor_monitor"
description: |-
  Provides a CloudWatch Internet Monitor Monitor resource
---

# Resource: aws_internetmonitor_monitor

Provides a Internet Monitor Monitor resource.

## Example Usage

```terraform
resource "aws_internetmonitor_monitor" "example" {
  monitor_name = "exmple"
}
```

## Argument Reference

The following arguments are required:

* `monitor_name` - (Required) The name of the monitor.

The following arguments are optional:

* `health_events_config` - (Optional) Health event thresholds. A health event threshold percentage, for performance and availability, determines when Internet Monitor creates a health event when there's an internet issue that affects your application end users. See [Health Events Config](#health-events-config) below.
* `internet_measurements_log_delivery` - (Optional) Publish internet measurements for Internet Monitor to an Amazon S3 bucket in addition to CloudWatch Logs.
* `max_city_networks_to_monitor` - (Optional) The maximum number of city-networks to monitor for your resources. A city-network is the location (city) where clients access your application resources from and the network or ASN, such as an internet service provider (ISP), that clients access the resources through. This limit helps control billing costs.
* `resources` - (Optional) The resources to include in a monitor, which you provide as a set of Amazon Resource Names (ARNs).
* `status` - (Optional) The status for a monitor. The accepted values for Status with the UpdateMonitor API call are the following: `ACTIVE` and `INACTIVE`.
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `traffic_percentage_to_monitor` - (Optional) The percentage of the internet-facing traffic for your application that you want to monitor with this monitor.

### Health Events Config

Defines the health event threshold percentages, for performance score and availability score. Amazon CloudWatch Internet Monitor creates a health event when there's an internet issue that affects your application end users where a health score percentage is at or below a set threshold. If you don't set a health event threshold, the default value is 95%.

* `availability_score_threshold` - (Optional) The health event threshold percentage set for availability scores.
* `performance_score_threshold` - (Optional) The health event threshold percentage set for performance scores.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Monitor.
* `id` - Name of the monitor.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

Internet Monitor Monitors can be imported using the `monitor_name`, e.g.,

```
$ terraform import aws_internetmonitor_monitor.some some-monitor
```
