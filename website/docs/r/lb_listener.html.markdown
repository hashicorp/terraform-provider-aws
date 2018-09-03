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

### Forward Action

```hcl
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
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.front_end.arn}"
  }
}
```

### Redirect Action

```hcl
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.front_end.arn}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port = "443"
      protocol = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}
```

### Fixed-response Action

```hcl
resource "aws_lb" "front_end" {
  # ...
}

resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.front_end.arn}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code = "200"
    }
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

## Attributes Reference

The following attributes are exported in addition to the arguments listed above:

* `id` - The ARN of the listener (matches `arn`)
* `arn` - The ARN of the listener (matches `id`)

## Import

Listeners can be imported using their ARN, e.g.

```
$ terraform import aws_lb_listener.front_end arn:aws:elasticloadbalancing:us-west-2:187416307283:listener/app/front-end-alb/8e4497da625e2d8a/9ab28ade35828f96
```
