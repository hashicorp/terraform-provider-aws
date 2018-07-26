---
layout: "aws"
page_title: "AWS: aws_lb_listener"
sidebar_current: "docs-aws-resource-elbv2-listener"
description: |-
  Provides a Load Balancer Listener resource.
---

# aws_lb_listener

Provides a Load Balancer Listener resource.

~> **Note:** `aws_alb_listener` is known as `aws_lb_listener`. The functionality is identical.

## Example Usage

```hcl
# Create a new load balancer
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_target_group" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.front_end.arn}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2015-05"
  certificate_arn   = "arn:aws:iam::187416307283:server-certificate/test_cert_rab3wuqwgja25ct3n4jdj2tzu4"

  default_action {
    target_group_arn = "${aws_lb_target_group.front_end.arn}"
    type             = "forward"
  }
}
```

## Argument Reference

The following arguments are supported:

* `load_balancer_arn` - (Required, Forces New Resource) The ARN of the load balancer.
* `port` - (Required) The port on which the load balancer is listening.
* `protocol` - (Optional) The protocol for connections from clients to the load balancer. Valid values are `TCP`, `HTTP` and `HTTPS`. Defaults to `HTTP`.
* `ssl_policy` - (Optional) The name of the SSL Policy for the listener. Required if `protocol` is `HTTPS`.
* `certificate_arn` - (Optional) The ARN of the default SSL server certificate. Exactly one certificate is required if the protocol is HTTPS. For adding additional SSL certificates, see the [`aws_lb_listener_certificate` resource](/docs/providers/aws/r/lb_listener_certificate.html).
* `default_action` - (Required) An Action block. Action blocks are documented below.

~> **NOTE::** Please note that listeners that are attached to Application Load Balancers must use either `HTTP` or `HTTPS` protocols while listeners that are attached to Network Load Balancers must use the `TCP` protocol.

Action Blocks (for `default_action`) support the following:

* `target_group_arn` - (Optional) The ARN of the Target Group to which to route traffic.
* `type` - (Required) The type of routing action. The valid values are `forward`, `authenticate-oidc` and `authenticate-cognito`.
* `order` - (Optional) The order for the action.
* `authenticate_cognito_config` - (Optional) The [authenticate cognito config](#authenticate-cognito-config) configuration.
* `authenticate_oidc_config` - (Optional) The [authenticate oidc  config](#authenticate-oidc-config) configuration.

### Authenticate Cognito Config

* `authentication_request_extra_params` - (Optional) The query parameters to include in the redirect request to the authorization endpoint. Max: 10. [See](#authentication-request-extra-params)
* `on_unauthenticated_request` - (Optional) The behavior if the user is not authenticated. Valid values: `deny`, `allow` and `authenticate`
* `scope` - (Optional) The set of user claims to be requested from the IdP.
* `session_cookie_name` - (Optional) The name of the cookie used to maintain session information.
* `session_time_out` - (Optional) The maximum duration of the authentication session, in seconds.
* `user_pool_arn` - (Required) The ARN of the Cognito user pool.
* `user_pool_client` - (Required) The ID of the Cognito user pool client.
* `user_pool_domain` - (Required) The domain prefix or fully-qualified domain name of the Cognito user pool.

### Authenticate OIDC Config

* `authentication_request_extra_params` - (Optional) The query parameters to include in the redirect request to the authorization endpoint. Max: 10. [See](#authentication-request-extra-params)
* `authorization_endpoint` - (Required) The authorization endpoint of the IdP.
* `client_id` - (Required) The OAuth 2.0 client identifier.
* `client_secret` - (Required) The OAuth 2.0 client secret.
* `issuer` - (Required) The OIDC issuer identifier of the IdP.
* `on_unauthenticated_request` - (Optional) The behavior if the user is not authenticated. Valid values: `deny`, `allow` and `authenticate`
* `scope` - (Optional) The set of user claims to be requested from the IdP.
* `session_cookie_name` - (Optional) The name of the cookie used to maintain session information.
* `session_time_out` - (Optional) The maximum duration of the authentication session, in seconds.
* `token_endpoint` - (Required) The token endpoint of the IdP.
* `user_info_endpoint` - (Required) The user info endpoint of the IdP.

### Authentication Request Extra Params

* `key` - (Required) The key of query parameter
* `value` - (Required) The value of query parameter

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of the listener (matches `arn`)
* `arn` - The ARN of the listener (matches `id`)

## Import

Listeners can be imported using their ARN, e.g.

```
$ terraform import aws_lb_listener.front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:listener/app/front-end-alb/8e4497da625e2d8a/9ab28ade35828f96
```
