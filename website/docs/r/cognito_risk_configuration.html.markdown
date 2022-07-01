---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_risk_configuration"
description: |-
  Provides a Cognito Risk Configuration resource.
---

# Resource: aws_cognito_risk_configuration

Provides a Cognito Risk Configuration resource.

## Example Usage

```terraform
resource "aws_cognito_risk_configuration" "example" {
  user_pool_id = aws_cognito_user_pool.example.id

  risk_exception_configuration {
    blocked_ip_range_list = ["10.10.10.10/32"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `user_pool_id` - (Required) The user pool ID.
* `client_id` - (Optional) The app client ID. When the client ID is not provided, the same risk configuration is applied to all the clients in the User Pool.
* `account_takeover_risk_configuration` - (Optional) The account takeover risk configuration. See details below.
* `compromised_credentials_risk_configuration` - (Optional) The compromised credentials risk configuration. See details below.
* `risk_exception_configuration` - (Optional) The configuration to override the risk decision. See details below.

### account_takeover_risk_configuration

* `notify_configuration` - (Required) The notify configuration used to construct email notifications. See details below.
* `actions` - (Required) Account takeover risk configuration actions. See details below.

#### notify_configuration

* `block_email` - (Optional) Email template used when a detected risk event is blocked. See notify email type below.
* `mfa_email` - (Optional) The multi-factor authentication (MFA) email template used when MFA is challenged as part of a detected risk. See notify email type below.
* `no_action_email` - (Optional) The email template used when a detected risk event is allowed. See notify email type below.
* `from` - (Optional) The email address that is sending the email. The address must be either individually verified with Amazon Simple Email Service, or from a domain that has been verified with Amazon SES.
* `reply_to` - (Optional) The destination to which the receiver of an email should reply to.
* `source_arn` - (Required) The Amazon Resource Name (ARN) of the identity that is associated with the sending authorization policy. This identity permits Amazon Cognito to send for the email address specified in the From parameter.

##### notify email type

* `html_body` - (Required) The email HTML body.
* `block_email` - (Required) The email subject.
* `block_email` - (Required) The email text body.

#### actions

* `high_action` - (Optional) Action to take for a high risk. See action block below.
* `low_action` - (Optional) Action to take for a low risk. See action block below.
* `medium_action` - (Optional) Action to take for a medium risk. See action block below.

##### action

* `event_action` - (Required) The action to take in response to the account takeover action. Valid values are `BLOCK`, `MFA_IF_CONFIGURED`, `MFA_REQUIRED` and `NO_ACTION`.
* `notify` - (Required) Whether to send a notification.

### compromised_credentials_risk_configuration

* `event_filter` - (Optional) Perform the action for these events. The default is to perform all events if no event filter is specified. Valid values are `SIGN_IN`, `PASSWORD_CHANGE`, and `SIGN_UP`.
* `actions` - (Required) The compromised credentials risk configuration actions. See details below.

#### actions

* `event_action` - (Optional) The event action. Valid values are `BLOCK` or `NO_ACTION`.

### risk_exception_configuration

* `blocked_ip_range_list` - (Optional) Overrides the risk decision to always block the pre-authentication requests. The IP range is in CIDR notation, a compact representation of an IP address and its routing prefix.
* `skipped_ip_range_list` - (Optional) Risk detection isn't performed on the IP addresses in this range list. The IP range is in CIDR notation.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The user pool ID. or The user pool ID and Client Id separated by a `:` if the configuration is client specific.

## Import

Cognito Risk Configurations can be imported using the `id`, e.g.,

```
$ terraform import aws_cognito_risk_configuration.main example
```

```
$ terraform import aws_cognito_risk_configuration.main example:example
```
