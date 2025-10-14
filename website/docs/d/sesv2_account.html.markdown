---
subcategory: "SESv2 (Simple Email V2)"
layout: "aws"
page_title: "AWS: aws_sesv2_account"
description: |-
  Provides details about an AWS SESv2 (Simple Email V2) Account.
---

# Data Source: aws_sesv2_account

Provides details about an AWS SESv2 (Simple Email V2) Account.

## Example Usage

### Basic Usage

```terraform
data "aws_sesv2_account" "example" {
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `dedicated_ip_auto_warmup_enabled` - Whether the automatic warm-up feature is enabled for associated dedicated IP addresses.
* `details` - Information about the account.
    * `additional_contact_email_addresses` - Additional email addresses where updates about the review process are sent.
    * `contact_language` - Preferred language for the case. Valid values are `EN` and `JA`.
    * `mail_type` - Type of email sent from the account. Valid values are `MARKETING` and `TRANSACTIONAL`.
    * `review_details`: Information about the latest review.
        * `case_id`: Associated support center case ID (if any).
        * `status`: Status of the review. Valid values are `PENDING`, `GRANTED`, `DENIED` and `FAILED`.
    * `use_case_description`: Description of the types of email sent from the account.
    * `website_url`: URL of a website helpful to understand the type of content sent from the account.
* `enforcement_status` - Reputation status. Valid values are `HEALTHY`, `PROBATION`, or `SHUTDOWN`.
* `production_access_enabled` - Whether production access has been granted. If the value is `false`, then the account is in the _sandbox_.
* `send_quota` - Information about the per-day and per-second sending limits.
    * `max_24_hour_send` - Maximum number of emails that can be sent over a 24-hour period (_sending quota_). A value of `-1` means unlimited quota.
    * `max_send_rate` - Maximum number of emails that can be sent per second (_maximum sending rate_ or _maximum TPS (transactions per second) rate_).
    * `sent_last_24_hours` - Number of emails sent over the past 24 hours.
* `sending_enabled` - Whether email sending is enabled.
* `suppression_attributes` - Information about the email address suppression configuration.
    * `suppressed_reasons` - List of the reasons for automatically adding email addresses to the suppression list. Valid values are `BOUNCE` and `COMPLAINT`.
* `vdm_attributes` - Information about the VDM configuration.
    * `dashboard_attributes` - Additional VDM configuration settings as applicable to the Dashboard.
        * `engagement_metrics` - Status of the VDM engagement metrics collection.
    * `guardian_attributes` - Additional VDM configuration settings as applicable to the Guardian.
        * `optimized_shared_delivery` - Status of the VDM optimized shared delivery.
    * `vdm_enabled` - Status of the VDM configuration. Valid values are `ENABLED` and `DISABLED`.
