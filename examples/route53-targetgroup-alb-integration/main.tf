provider "aws" {
  region = "ap-south-1"
}
# Route53 A record creation
resource "aws_route53_record" "new_record" {
  zone_id = var.hostedzone_id
  name    = var.record_name
  type    = "A"

  alias {
    name                   = var.loadbalancer_name
    zone_id                = var.loadbalancer_zoneid
    evaluate_target_health = true
  }
}

# Target group creation
resource "aws_lb_target_group" "new_target_group" {
  name        = var.target_group_name
  port        = var.target_group_port
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = var.vpc_id
}

# Attaching target group to the ALB
resource "aws_lb_target_group_attachment" "attach_target_group" {
  target_group_arn = aws_lb_target_group.new_target_group.arn
  target_id        = var.target_ins_ip
  port             = var.target_group_port
}


# Adding rule to alb listener
resource "aws_lb_listener_rule" "host_header_rule" {
  listener_arn = var.alb_https_listener_arn
  priority     = var.rule_priority
  condition {
    host_header {
      values = [var.record_name]
    }
  }

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.new_target_group.arn
  }
}
