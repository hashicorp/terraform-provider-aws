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

# Forward action

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

# Redirect action

resource "aws_lb_listener_rule" "redirect_http_to_https" {
  listener_arn = "${aws_lb_listener.front_end.arn}"

  action {
    type = "redirect"
    redirect {
      port = "443"
      protocol = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  condition {
    field  = "host-header"
    values = ["my-service.*.terraform.io"]
  }
}

# Fixed-response action

resource "aws_lb_listener_rule" "health_check" {
  listener_arn = "${aws_lb_listener.front_end.arn}"

  action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "HEALTHY"
      status_code = "200"
    }
  }

  condition {
    field  = "path-pattern"
    values = ["/health"]
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

* `type` - (Required) The type of routing action. Valid values are `forward`, `redirect` and `fixed-response`.
* `target_group_arn` - (Optional) The ARN of the Target Group to which to route traffic. Required if `type` is `forward`.
* `redirect` - (Optional) Information for creating a redirect action. Required if `type` is `redirect`.
* `fixed_response` - (Optional) Information for creating an action that returns a custom HTTP response. Required if `type` is `fixed-response`.

Redirect Blocks (for `redirect`) support the following:

~> **NOTE::** You can reuse URI components using the following reserved keywords: `#{protocol}`, `#{host}`, `#{port}`, `#{path}` (the leading "/" is removed) and `#{query}`.

* `host` - (Optional) The hostname. This component is not percent-encoded. The hostname can contain `#{host}`. Defaults to `#{host}`.
* `path` - (Optional) The absolute path, starting with the leading "/". This component is not percent-encoded. The path can contain #{host}, #{path}, and #{port}. Defaults to `/#{path}`.
* `port` - (Optional) The port. Specify a value from `1` to `65535` or `#{port}`. Defaults to `#{port}`.
* `protocol` - (Optional) The protocol. Valid values are `HTTP`, `HTTPS`, or `#{protocol}`. Defaults to `#{protocol}`.
* `query` - (Optional) The query parameters, URL-encoded when necessary, but not percent-encoded. Do not include the leading "?". Defaults to `#{query}`.
* `status_code` - (Required) The HTTP redirect code. The redirect is either permanent (`HTTP_301`) or temporary (`HTTP_302`).

Fixed-response Blocks (for `fixed_response`) support the following:

* `content_type` - (Required) The content type. Valid values are `text/plain`, `text/css`, `text/html`, `application/javascript` and `application/json`.
* `message_body` - (Optional) The message body.
* `status_code` - (Optional) The HTTP response code. Valid values are `2XX`, `4XX`, or `5XX`.

Condition Blocks (for `condition`) support the following:

* `field` - (Required) The name of the field. Must be one of `path-pattern` for path based routing or `host-header` for host based routing.
* `values` - (Required) The path patterns to match. A maximum of 1 can be defined.

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of the rule (matches `arn`)
* `arn` - The ARN of the rule (matches `id`)

## Import

Rules can be imported using their ARN, e.g.

```
$ terraform import aws_lb_listener_rule.front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:listener-rule/app/test/8e4497da625e2d8a/9ab28ade35828f96/67b3d2d36dd7c26b
```
