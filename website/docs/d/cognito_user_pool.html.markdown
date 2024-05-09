---

subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool"
description: |-
  Terraform data source for managing an AWS Cognito User Pool.

---

# Data Source: aws_cognito_user_pool

Terraform data source for managing an AWS Cognito User Pool.

## Example Usage

### Basic Usage

```terraform
data "aws_cognito_user_pool" "example" {
  id = "us-west-2_aaaaaaaaa"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) The cognito pool id

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the User Pool. 
* [account_recovery_setting](#account-recover-setting)
* [admin_create_user_config](#admin-create-user-config
* `auto_verified_attributes`
* `creation_date`
* `custom_domain`
* `deletion_protection`
* [device_configuration](#device-configuration)
* `domain`
* [email_configuration](#email-configuration)
* `email_verification_message`
* `email_verification_subject`
* `estimated_number_of_users`
* [lambda_config](#lambda-config)
* `last_modified_date`
* `mfa_configuration`
* `name`
* [schema_attributes](#schema-attributes)
* `sms_authentication_message`
* `sms_configuration_failure`
* `sms_verification_message`
* `user_pool_tags`
* `username_attributes`

### account recover setting

* [recovery_mechanism](#recovery-mechanism) - Details about an individual recovery mechanism.

### recovery mechanism
* `name` - Name of the recovery mechanism (e.g., email, phone number).
* `priority` - Priority of this mechanism in the recovery process (lower numbers are higher priority).

### admin create user config

* `allow_admin_create_user_only` - Whether only admins can create users.
* `unused_account_validity_days` - Number of days an unconfirmed user account remains valid.
* [invite_message_template](#invite-message-template) - Templates for invitation messages.

### invite message template
* `email_message` - Email message content.
* `email_subject` - Email message subject.
* `sms_message` - SMS message content.

### device configuration

* `challenge_required_on_new_device` - Whether a challenge is required on new devices.
* `device_only_remembered_on_user_prompt` - Whether devices are only remembered if the user prompts it.

### email configuration

* `configuration_set` - Configuration set used for sending emails.
* `email_sending_account` - Email sending account.
* `from` - Email sender address.
* `reply_to_email_address` - Reply-to email address.
* `source_arn` - Source Amazon Resource Name (ARN) for emails.

### lambda config
* [custom_email_sender](#lambda-function) - Configuration for a custom email sender Lambda function.

* [custom_sms_sender](#lambda-function) - Configuration for a custom SMS sender Lambda function
* [pre_token_generation_config](#lambda-function) - Configuration for a Lambda function that executes before token generation.

### lambda function
* `lambda_arn` - ARN of the Lambda function.
* `lambda_version` - Version of the Lambda function.

### schema attributes

* `attribute_data_type` - Data type of the attribute (e.g., string, number).
* `developer_only_attribute` - Whether the attribute is for developer use only.
* `mutable` - Whether the attribute can be changed after user creation.
* `name` - Name of the attribute.
* `required` - Whether the attribute is required during user registration.
* [number_attribute_constraints](#number-attribute-constraints) - Constraints for numeric attributes (if applicable).
* [string_attribute_constraints](#string-attribute-constraints) - Constraints for string attributes (if applicable).

### number attribute constraints
* `max_value` - Maximum allowed value.
* `min_value` - Minimum allowed value.

### string attribute constraints
* `max_length` - Maximum allowed length.
* `min_length` - Minimum allowed length.
