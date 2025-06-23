resource "aws_lb_listener_rule" "test" {
  listener_arn = aws_lb_listener.test.arn
  priority     = 100
  {{ template "region" }}  

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.test.arn
  }

  condition {
    path_pattern {
      values = ["/static/*"]
    }
  }

{{- template "tags" . }}
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"
  {{ template "region" }}  

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id
  {{ template "region" }}  

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lb" "test" {
  name            = var.rName
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id
  {{ template "region" }}  

  idle_timeout               = 30
  enable_deletion_protection = false
}

resource "aws_lb_target_group" "test" {
  name     = var.rName
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
  {{ template "region" }}  

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
