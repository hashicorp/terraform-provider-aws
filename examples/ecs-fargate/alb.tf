resource "aws_alb" "alb" {
  name            = "tf-example-ecs-fargate"
  subnets         = ["${aws_subnet.public.*.id}"]
  security_groups = ["${aws_security_group.lb_sg.id}"]
}

resource "aws_alb_target_group" "atg" {
  name        = "tf-example-ecs-fargate"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = "${aws_vpc.main.id}"
  target_type = "ip"
}

resource "aws_alb_listener" "front_end" {
  load_balancer_arn = "${aws_alb.alb.id}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = "${aws_alb_target_group.atg.id}"
    type             = "forward"
  }
}
