---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_configuration_set"
description: |-
  Terraform data source for managing an AWS SESv2 (Simple Email V2) Configuration Set.
---

# Data Source: aws_sesv2_configuration_set

Terraform data source for managing an AWS SESv2 (Simple Email V2) Configuration Set.

## Example Usage

### Basic Usage

```terraform
data "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"
}
```

## Argument Reference

This data source supports the following arguments:

* `configuration_set_name` - (Required) Name of the configuration set.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the configuration set.
* `delivery_options` - Object that defines the dedicated IP pool used to send emails with the configuration set.
    * `max_delivery_seconds` - Maximum amount of time, in seconds, that Amazon SES API v2 attempts delivery of email. If specified, the value must be greater than or equal to 300 seconds (5 minutes) and less than or equal to 50400 seconds (840 minutes).
    * `sending_pool_name` - Name of the dedicated IP pool to associate with the configuration set.
    * `tls_policy` - Whether messages that use the configuration set are required to use TLS.
* `reputation_options` - Object that defines whether Amazon SES collects reputation metrics for emails sent with the configuration set.
    * `last_fresh_start` - Date and time (in Unix time) when the reputation metrics were last given a fresh start.
    * `reputation_metrics_enabled` - Whether tracking of reputation metrics is enabled.
* `sending_options` - Object that defines whether Amazon SES can send email sent with the configuration set.
    * `sending_enabled` - Whether email sending is enabled.
* `suppression_options` - Object that contains information about the suppression list preferences for your account.
    * `suppressed_reasons` - List that contains the reasons that email addresses are automatically added to the suppression list for your account.
* `tags` - Key-value map of resource tags.
* `tracking_options` - Object that defines the open and click tracking options for emails sent with the configuration set.
    * `custom_redirect_domain` - Domain used for tracking open and click events.
    * `https_policy` - HTTPS policy used for tracking open and click events. Valid values are `REQUIRE`, `REQUIRE_OPEN_ONLY`, or `OPTIONAL`.
* `vdm_options` - Object that contains information about the VDM preferences for your configuration set.
    * `dashboard_options` - Additional settings for your VDM configuration as applicable to the Dashboard.
        * `engagement_metrics` - Status of your VDM engagement metrics collection.
    * `guardian_options` - Additional settings for your VDM configuration as applicable to the Guardian.
        * `optimized_shared_delivery` - Status of your VDM optimized shared delivery.
