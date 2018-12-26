resource "aws_ecs_cluster" "cluster" {
  name = "tf-example-ecs-fargate"
}

resource "aws_ecs_task_definition" "app" {
  family = "tf-example-ecs-fargate-task"

  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]

  cpu    = "${var.task_cpu}"
  memory = "${var.task_memory}"

  execution_role_arn = "${aws_iam_role.ecs_execution_role.arn}"
  task_role_arn      = "${aws_iam_role.ecs_execution_role.arn}"

  container_definitions = <<DEFINITION
[
  {
    "cpu": ${var.task_cpu},
    "image": "${var.app_docker_image}",
    "memory": ${var.task_memory},
    "name": "${var.app_name}",
    "networkMode": "awsvpc",
    "portMappings": [
      {
        "containerPort": ${var.app_port},
        "hostPort": ${var.app_port}
      }
    ],
    "logConfiguration": {
      "logDriver": "awslogs",
      "options": {
        "awslogs-group": "${aws_cloudwatch_log_group.app.name}",
        "awslogs-region": "${var.aws_region}",
        "awslogs-stream-prefix": "ecs"
      }
    }
  }
]
DEFINITION
}

resource "aws_ecs_service" "service" {
  name            = "tf-example-ecs-fargate"
  cluster         = "${aws_ecs_cluster.cluster.id}"
  task_definition = "${aws_ecs_task_definition.app.arn}"
  desired_count   = "${var.service_desired}"
  launch_type     = "FARGATE"
  iam_role        = ""

  network_configuration {
    subnets         = ["${aws_subnet.private.*.id}"]
    security_groups = ["${aws_security_group.ecs_sg.id}"]
  }

  load_balancer {
    target_group_arn = "${aws_alb_target_group.atg.id}"
    container_name   = "${var.app_name}"
    container_port   = "${var.app_port}"
  }

  depends_on = [
    "aws_alb_listener.front_end",
    "aws_iam_role_policy.ecs_execution_role_policy",
  ]
}
