---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_configuration_set"
description: |-
  Terraform resource for managing an AWS SESv2 (Simple Email V2) Configuration Set.
---

# Resource: aws_sesv2_configuration_set

Terraform resource for managing an AWS SESv2 (Simple Email V2) Configuration Set.

## Example Usage

### Basic Usage

```terraform
resource "aws_sesv2_configuration_set" "example" {
  configuration_set_name = "example"

  delivery_options {
    tls_policy = "REQUIRE"
  }

  reputation_options {
    reputation_metrics_enabled = false
  }

  sending_options {
    sending_enabled = true
  }

  suppression_options {
    suppressed_reasons = ["BOUNCE", "COMPLAINT"]
  }

  tracking_options {
    custom_redirect_domain = "example.com"
  }
}
```

## Argument Reference

The following arguments are supported:

* `configuration_set_name` - (Required) The name of the configuration set.
* `delivery_options` - (Optional) An object that defines the dedicated IP pool that is used to send emails that you send using the configuration set.
* `reputation_options` - (Optional) An object that defines whether or not Amazon SES collects reputation metrics for the emails that you send that use the configuration set.
* `sending_options` - (Optional) An object that defines whether or not Amazon SES can send email that you send using the configuration set.
* `suppression_options` - (Optional) An object that contains information about the suppression list preferences for your account.
* `tags` - (Optional) A map of tags to assign to the service. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `tracking_options` - (Optional) An object that defines the open and click tracking options for emails that you send using the configuration set.

### delivery_options

The following arguments are supported:

* `sending_pool_name` - (Optional) The name of the dedicated IP pool to associate with the configuration set.
* `tls_policy` - (Optional) Specifies whether messages that use the configuration set are required to use Transport Layer Security (TLS). Valid values: `REQUIRE`, `OPTIONAL`.

### reputation_options

The following arguments are supported;

* `reputation_metrics_enabled` - (Optional) If `true`, tracking of reputation metrics is enabled for the configuration set. If `false`, tracking of reputation metrics is disabled for the configuration set.

### sending_options

The following arguments are supported:

* `sending_enabled` - (Optional) If `true`, email sending is enabled for the configuration set. If `false`, email sending is disabled for the configuration set.

### suppression_options

- `suppressed_reasons` - (Optional) A list that contains the reasons that email addresses are automatically added to the suppression list for your account. Valid values: `BOUNCE`, `COMPLAINT`.

## tracking_options

- `custom_redirect_domain` - (Required) The domain to use for tracking open and click events.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Configuration Set.
* `reputation_options` - An object that defines whether or not Amazon SES collects reputation metrics for the emails that you send that use the configuration set.
    * `last_fresh_start` - The date and time (in Unix time) when the reputation metrics were last given a fresh start. When your account is given a fresh start, your reputation metrics are calculated starting from the date of the fresh start.

## Import

SESv2 (Simple Email V2) Configuration Set can be imported using the `configuration_set_name`, e.g.,

```
$ terraform import aws_sesv2_configuration_set.example example
```
