---
layout: "aws"
page_title: "AWS: aws_lb_listener_rule"
sidebar_current: "docs-aws-resource-elbv2-listener-rule"
description: |-
  Provides a Load Balancer Listener Rule resource.
---

# aws_lb_listener_rule

Provides a Load Balancer Listener Rule resource.

~> **Note:** `aws_alb_listener_rule` is known as `aws_lb_listener_rule`. The functionality is identical.

## Example Usage

```hcl
# Create a new load balancer
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  # Other parameters
}

resource "aws_lb_listener_rule" "static" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.static.arn}"
  }

  condition {
    field  = "path-pattern"
    values = ["/static/*"]
  }
}

resource "aws_lb_listener_rule" "host_based_routing" {
  listener_arn = "${aws_lb_listener.front_end.arn}"
  priority     = 99

  action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.static.arn}"
  }

  condition {
    field  = "host-header"
    values = ["my-service.*.terraform.io"]
  }
}

```

## Argument Reference

The following arguments are supported:

* `listener_arn` - (Required, Forces New Resource) The ARN of the listener to which to attach the rule.
* `priority` - (Optional) The priority for the rule between `1` and `50000`. Leaving it unset will automatically set the rule with next available priority after currently existing highest rule. A listener can't have multiple rules with the same priority.
* `action` - (Required) An Action block. Action blocks are documented below.
* `condition` - (Required) A Condition block. Condition blocks are documented below.

Action Blocks (for `action`) support the following:

* `target_group_arn` - (Optional) The ARN of the Target Group to which to route traffic.
* `type` - (Required) The type of routing action. The valid values are `forward`, `authenticate-oidc` and `authenticate-cognito`.
* `order` - (Optional) The order for the action.
* `authenticate_cognito_config` - (Optional) The [authenticate cognito config](#authenticate-cognito-config) configuration.
* `authenticate_oidc_config` - (Optional) The [authenticate oidc  config](#authenticate-oidc-config) configuration.

Condition Blocks (for `condition`) support the following:

* `field` - (Required) The name of the field. Must be one of `path-pattern` for path based routing or `host-header` for host based routing.
* `values` - (Required) The path patterns to match. A maximum of 1 can be defined.

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

* `id` - The ARN of the rule (matches `arn`)
* `arn` - The ARN of the rule (matches `id`)

## Import

Rules can be imported using their ARN, e.g.

```
$ terraform import aws_lb_listener_rule.front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:listener-rule/app/test/8e4497da625e2d8a/9ab28ade35828f96/67b3d2d36dd7c26b
```
