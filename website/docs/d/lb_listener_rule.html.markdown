---
subcategory: "ELB (Elastic Load Balancing)"
layout: "aws"
page_title: "AWS: aws_lb_listener_rule"
description: |-
  Terraform data source for managing an AWS ELB (Elastic Load Balancing) Listener Rule.
---

# Data Source: aws_lb_listener_rule

Provides information about an AWS Elastic Load Balancing Listener Rule.

## Example Usage

### Match by Rule ARN

```terraform
variable "lb_rule_arn" {
  type = string
}

data "aws_lb_listener_rule" "example" {
  arn = var.lb_rule_arn
}
```

### Match by Listener ARN and Priority

```terraform
variable "lb_listener_arn" {
  type = string
}

variable "lb_rule_priority" {
  type = number
}

data "aws_lb_listener_rule" "example" {
  listener_arn = var.lb_listener_arn
  priority     = var.lb_rule_priority
}
```

## Argument Reference

This data source supports the following arguments:

* `arn` - (Optional) ARN of the Listener Rule.
  Either `arn` or `listener_arn` must be set.
* `listener_arn` - (Optional) ARN of the associated Listener.
  Either `arn` or `listener_arn` must be set.
* `priority` - (Optional) Priority of the Listener Rule within the Listener.
  Must be set if `listener_arn` is set, otherwise must not be set.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `action` - List of actions associated with the rule, sorted by `order`.
  [Detailed below](#action).
* `condition` - Set of conditions associated with the rule.
  [Detailed below](#condition).
* `tags` - Tags assigned to the Listener Rule.

### `action`

* `type` - The type of the action, indicates which sub-block will be populated.
* `order` - The evaluation order of the action.
* `authenticate_cognito` - An action to authenticate using Amazon Cognito.
  [Detailed below](#authenticate_cognito).
* `authenticate_oidc` - An action to authenticate using OIDC.
  [Detailed below](#authenticate_oidc).
* `fixed_response` - An action to return a fixed response.
  [Detailed below](#fixed_response).
* `forward` - An action to forward the request.
  [Detailed below](#forward).
* `redirect` - An action to redirect the request.
  [Detailed below](#redirect).

#### `authenticate_cognito`

* `authentication_request_extra_params` - Set of additional parameters for the request.
  [Detailed below](#authentication_request_extra_params).
* `on_unauthenticated_request` - Behavior when the client is not authenticated.
* `scope` - Set of user claims requested.
* `session_cookie_name` - Name of the cookie used to maintain session information.
* `session_timeout` - Maximum duration of the authentication session in seconds.
* `user_pool_arn` - ARN of the Cognito user pool.
* `user_pool_client_id` - ID of the Cognito user pool client.
* `user_pool_domain` - Domain prefix or fully-qualified domain name of the Cognito user pool.

#### `authenticate_oidc`

* `authentication_request_extra_params` - Set of additional parameters for the request.
  [Detailed below](#authentication_request_extra_params).
* `authorization_endpoint` -  The authorization endpoint of the IdP.
* `client_id` - OAuth 2.0 client identifier.
* `issuer` - OIDC issuer identifier of the IdP.
* `on_unauthenticated_request` - Behavior when the client is not authenticated.
* `scope` - Set of user claims requested.
* `session_cookie_name` - Name of the cookie used to maintain session information.
* `session_timeout` - Maximum duration of the authentication session in seconds.
* `token_endpoint` - The token endpoint of the IdP.
* `user_info_endpoint` - The user info endpoint of the IdP.

#### `authentication_request_extra_params`

* `key` - Key of query parameter
* `value` - Value of query parameter

#### `fixed_response`

* `content_type` - Content type of the response.
* `message_body` - Message body of the response.
* `status_code` - HTTP response code of the response.

#### `forward`

* `stickiness` - Target group stickiness for the rule.
  [Detailed below](#stickiness).
* `target_group` - Set of target groups for the action.
  [Detailed below](#target_group).

#### `stickiness`

* `duration` - The time period, in seconds, during which requests from a client should be routed to the same target group.
* `enabled` - Indicates whether target group stickiness is enabled.

#### `target_group`

* `arn` - ARN of the target group.
* `weight` - Weight of the target group.

#### `redirect`

* `host` - The hostname.
* `path` - The absolute path, starting with `/`.
* `port` - The port.
* `protocol` - The protocol.
* `query` - The query parameters.
* `status_code` - The HTTP redirect code.

### `condition`

* `host_header` - Contains a single attribute `values`, which contains a set of host names.
* `http_header` - HTTP header and values to match.
  [Detailed below](#http_header).
* `http_request_method` - Contains a single attribute `values`, which contains a set of HTTP request methods.
* `path_pattern` - Contains a single attribute `values`, which contains a set of path patterns to compare against the request URL.
* `query_string` - Query string parameters to match.
  [Detailed below](#query_string).
* `source_ip` - Contains a single attribute `values`, which contains a set of source IPs in CIDR notation.

#### `http_header`

* `http_header_name` - Name of the HTTP header to match.
* `values` - Set of values to compare against the value of the HTTP header.

#### `query_string`

* `values` - Set of `key`-`value` pairs indicating the query string parameters to match.
