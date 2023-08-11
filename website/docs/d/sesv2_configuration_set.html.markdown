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

The following arguments are required:

* `configuration_set_name` - (Required) The name of the configuration set.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `delivery_options` - An object that defines the dedicated IP pool that is used to send emails that you send using the configuration set.
    * `sending_pool_name` - The name of the dedicated IP pool to associate with the configuration set.
    * `tls_policy` - Specifies whether messages that use the configuration set are required to use Transport Layer Security (TLS).
* `reputation_options` - An object that defines whether or not Amazon SES collects reputation metrics for the emails that you send that use the configuration set.
    * `last_fresh_start` - The date and time (in Unix time) when the reputation metrics were last given a fresh start.
    * `reputation_metrics_enabled` - Specifies whether tracking of reputation metrics is enabled.
* `sending_options` - An object that defines whether or not Amazon SES can send email that you send using the configuration set.
    * `sending_enabled` - Specifies whether email sending is enabled.
* `suppression_options` - An object that contains information about the suppression list preferences for your account.
    * `suppressed_reasons` - A list that contains the reasons that email addresses are automatically added to the suppression list for your account.
* `tags` - Key-value map of resource tags for the container recipe.
* `tracking_options` - An object that defines the open and click tracking options for emails that you send using the configuration set.
    * `custom_redirect_domain` - The domain to use for tracking open and click events.
* `vdm_options` - An object that contains information about the VDM preferences for your configuration set.
    * `dashboard_options` - Specifies additional settings for your VDM configuration as applicable to the Dashboard.
        * `engagement_metrics` - Specifies the status of your VDM engagement metrics collection.
    * `guardian_options` - Specifies additional settings for your VDM configuration as applicable to the Guardian.
        * `optimized_shared_delivery` - Specifies the status of your VDM optimized shared delivery.
