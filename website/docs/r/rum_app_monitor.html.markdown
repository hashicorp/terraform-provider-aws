---
subcategory: "CloudWatch RUM"
layout: "aws"
page_title: "AWS: aws_rum_app_monitor"
description: |-
  Provides a CloudWatch RUM App Monitor resource.
---

# Resource: aws_rum_app_monitor

Provides a CloudWatch RUM App Monitor resource.

## Example Usage

```terraform
resource "aws_rum_app_monitor" "example" {
  name   = "example"
  domain = "localhost"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) The name of the log stream.
* `app_monitor_configuration` - (Optional) configuration data for the app monitor. See [app_monitor_configuration](#app_monitor_configuration) below.
* `cw_log_enabled` - (Optional) Data collected by RUM is kept by RUM for 30 days and then deleted. This parameter specifies whether RUM sends a copy of this telemetry data to Amazon CloudWatch Logs in your account. This enables you to keep the telemetry data for more than 30 days, but it does incur Amazon CloudWatch Logs charges. Default value is `false`.
* `custom_events` - (Optional) Specifies whether this app monitor allows the web client to define and send custom events. If you omit this parameter, custom events are `DISABLED`. See [custom_events](#custom_events) below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### app_monitor_configuration

* `allow_cookies` - (Optional) If you set this to `true`, RUM web client sets two cookies, a session cookie and a user cookie. The cookies allow the RUM web client to collect data relating to the number of users an application has and the behavior of the application across a sequence of events. Cookies are stored in the top-level domain of the current page.
* `domain` - (Optional) The top-level internet domain name for which your application has administrative authority. Exactly one of `domain` or `domain_list` must be specified.
* `domain_list` - (Optional) A list of internet domain names for which your application has administrative authority. Exactly one of `domain` or `domain_list` must be specified.
* `enable_xray` - (Optional) If you set this to `true`, RUM enables X-Ray tracing for the user sessions that RUM samples. RUM adds an X-Ray trace header to allowed HTTP requests. It also records an X-Ray segment for allowed HTTP requests.
* `excluded_pages` - (Optional) A list of URLs in your website or application to exclude from RUM data collection.
* `favorite_pages` - (Optional) A list of pages in the CloudWatch RUM console that are to be displayed with a "favorite" icon.
* `guest_role_arn` - (Optional) The ARN of the guest IAM role that is attached to the Amazon Cognito identity pool that is used to authorize the sending of data to RUM.
* `identity_pool_id` - (Optional) The ID of the Amazon Cognito identity pool that is used to authorize the sending of data to RUM.
* `included_pages` - (Optional)  If this app monitor is to collect data from only certain pages in your application, this structure lists those pages.
* `session_sample_rate` - (Optional) Specifies the percentage of user sessions to use for RUM data collection. Choosing a higher percentage gives you more data but also incurs more costs. The number you specify is the percentage of user sessions that will be used. Default value is `0.1`.
* `telemetries` - (Optional) An array that lists the types of telemetry data that this app monitor is to collect. Valid values are `errors`, `performance`, and `http`.

### custom_events

* `status` - (Optional) Specifies whether this app monitor allows the web client to define and send custom events. The default is for custom events to be `DISABLED`. Valid values are `DISABLED` and `ENABLED`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) specifying the app monitor.
* `id` - The CloudWatch RUM name as it is the identifier of a RUM.
* `app_monitor_id` - The unique ID of the app monitor. Useful for JS templates.
* `cw_log_group` - The name of the log group where the copies are stored.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cloudwatch RUM App Monitor using the `name`. For example:

```terraform
import {
  to = aws_rum_app_monitor.example
  id = "example"
}
```

Using `terraform import`, import Cloudwatch RUM App Monitor using the `name`. For example:

```console
% terraform import aws_rum_app_monitor.example example
```
