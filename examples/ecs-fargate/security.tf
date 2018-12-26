resource "aws_security_group" "lb_sg" {
  name        = "tf-example-ecs-fargate-lb"
  description = "controls access to the application ALB"

  vpc_id = "${aws_vpc.main.id}"

  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "ecs_sg" {
  name        = "tf-example-ecs-fargate-ecs"
  description = "controls direct access to ecs"

  vpc_id = "${aws_vpc.main.id}"

  ingress {
    protocol  = "tcp"
    from_port = "${var.app_port}"
    to_port   = "${var.app_port}"

    security_groups = ["${aws_security_group.lb_sg.id}"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}
