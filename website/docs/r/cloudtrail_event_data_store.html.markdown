---
subcategory: "CloudTrail"
layout: "aws"
page_title: "AWS: aws_cloudtrail_event_data_store"
description: |-
  Provides a CloudTrail Event Data Store.
---

# Resource: aws_cloudtrail_event_data_store

Provides a CloudTrail Event Data Store.

More information about event data stores see [Event Data Store User Guide](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/query-event-data-store.html).

-> **Tip:** For an organization event data store, this resource must be in the master account of the organization.

## Example Usage

### Basic

The most simple event data store configuration is as following. The event data store will automatically capture all management events. To capture management events from other regions, `multi_region_enabled` must be `true`.

```terraform
resource "aws_cloudtrail_event_data_store" "example" {
  name                  = "example-event-data-store"
  retention_period      = 7
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the event data store.
- `retention_period` - (Required) The retention period of the event data store, in days. You can set a retention period of up to 2555 days, the equivalent of seven years.
- `advanced_event_selector` - (Required) The advanced event selectors to use to select the events for the data store. For more information about how to use advanced event selectors, see [Log events by using advanced event selectors](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/logging-data-events-with-cloudtrail.html#creating-data-event-selectors-advanced) in the CloudTrail User Guide.
- `multi_region_enabled` - (Optional) Specifies whether the event data store includes events from all regions, or only from the region in which the event data store is created.
- `organization_enabled` - (Optional) Specifies whether an event data store collects events logged for an organization in AWS Organizations.
- `termination_protection_enabled` - (Optional) Specifies whether termination protection is enabled for the event data store. If termination protection is enabled, you cannot delete the event data store until termination protection is disabled.
- `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](/docs/providers/aws/index.html#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### Advanced Event Selector Arguments

For **advanced_event_selector** the following attributes are supported.

* `name` (Optional) - Specifies the name of the advanced event selector.
* `field_selector` (Required) - Specifies the selector statements in an advanced event selector. Fields documented below.

#### Field Selector Arguments

For **field_selector** the following attributes are supported.

* `field` (Required) - Specifies a field in an event record on which to filter events to be logged. You can specify only the following values: `readOnly`, `eventSource`, `eventName`, `eventCategory`, `resources.type`, `resources.ARN`.
* `equals` (Optional) - A list of values that includes events that match the exact value of the event record field specified as the value of `field`. This is the only valid operator that you can use with the `readOnly`, `eventCategory`, and `resources.type` fields.
* `not_equals` (Optional) - A list of values that excludes events that match the exact value of the event record field specified as the value of `field`.
* `starts_with` (Optional) - A list of values that includes events that match the first few characters of the event record field specified as the value of `field`.
* `not_starts_with` (Optional) - A list of values that excludes events that match the first few characters of the event record field specified as the value of `field`.
* `ends_with` (Optional) - A list of values that includes events that match the last few characters of the event record field specified as the value of `field`.
* `not_ends_with` (Optional) - A list of values that excludes events that match the last few characters of the event record field specified as the value of `field`.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:


## Import
