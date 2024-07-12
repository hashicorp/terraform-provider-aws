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
  user_pool_id = "us-west-2_aaaaaaaaa"
}
```

## Argument Reference

The following arguments are required:

* `user_pool_id` - (Required) The cognito pool ID

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the User Pool.
* [account_recovery_setting](#account-recover-setting) - The available verified method a user can use to recover their password when they call ForgotPassword. You can use this setting to define a preferred method when a user has more than one method available. With this setting, SMS doesn't qualify for a valid password recovery mechanism if the user also has SMS multi-factor authentication (MFA) activated. In the absence of this setting, Amazon Cognito uses the legacy behavior to determine the recovery method where SMS is preferred through email.
* [admin_create_user_config](#admin-create-user-config) - The configuration for AdminCreateUser requests.
* `auto_verified_attributes` - The attributes that are auto-verified in a user pool.
* `creation_date` - The date and time, in ISO 8601 format, when the item was created.
* `custom_domain` - A custom domain name that you provide to Amazon Cognito. This parameter applies only if you use a custom domain to host the sign-up and sign-in pages for your application. An example of a custom domain name might be auth.example.com.
* `deletion_protection` - When active, DeletionProtection prevents accidental deletion of your user pool. Before you can delete a user pool that you have protected against deletion, you must deactivate this feature.
* [device_configuration](#device-configuration) - The device-remembering configuration for a user pool. A null value indicates that you have deactivated device remembering in your user pool.
* `domain` - The domain prefix, if the user pool has a domain associated with it.
* [email_configuration](#email-configuration) - The email configuration of your user pool. The email configuration type sets your preferred sending method, AWS Region, and sender for messages from your user pool.
* `estimated_number_of_users` - A number estimating the size of the user pool.
* [lambda_config](#lambda-config) - The AWS Lambda triggers associated with the user pool.
* `last_modified_date` - The date and time, in ISO 8601 format, when the item was modified.
* `mfa_configuration` - Can be one of the following values: `OFF` | `ON` | `OPTIONAL`
* `name` - The name of the user pool.
* [schema_attributes](#schema-attributes) - A list of the user attributes and their properties in your user pool. The attribute schema contains standard attributes, custom attributes with a custom: prefix, and developer attributes with a dev: prefix. For more information, see User pool attributes.
* `sms_authentication_message` - The contents of the SMS authentication message.
* `sms_configuration_failure` - The reason why the SMS configuration can't send the messages to your users.
* `sms_verification_message` - The contents of the SMS authentication message.
* `user_pool_tags` - The tags that are assigned to the user pool. A tag is a label that you can apply to user pools to categorize and manage them in different ways, such as by purpose, owner, environment, or other criteria.
* `username_attributes` - Specifies whether a user can use an email address or phone number as a username when they sign up.

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
* [number_attribute_constraints](#number-attribute-constraints) - Constraints for numeric attributes.
* [string_attribute_constraints](#string-attribute-constraints) - Constraints for string attributes.

### number attribute constraints

* `max_value` - Maximum allowed value.
* `min_value` - Minimum allowed value.

### string attribute constraints

* `max_length` - Maximum allowed length.
* `min_length` - Minimum allowed length.
