resource "aws_appautoscaling_policy" "test" {
{{- template "region" }}
  name               = var.rName
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension
  service_namespace  = aws_appautoscaling_target.test.service_namespace

  step_scaling_policy_configuration {
    adjustment_type         = "ChangeInCapacity"
    cooldown                = 60
    metric_aggregation_type = "Average"

    step_adjustment {
      metric_interval_lower_bound = 0
      scaling_adjustment          = 1
    }
  }
}

resource "aws_appautoscaling_target" "test" {
{{- template "region" }}
  max_capacity       = 4
  min_capacity       = 0
  resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_ecs_service" "test" {
{{- template "region" }}
  cluster                            = aws_ecs_cluster.test.id
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
  desired_count                      = 2
  name                               = var.rName
  task_definition                    = aws_ecs_task_definition.test.arn
}

resource "aws_ecs_cluster" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_ecs_task_definition" "test" {
{{- template "region" }}
  family = var.rName

  container_definitions = <<EOF
[
  {
    "name": "busybox",
    "image": "busybox:latest",
    "cpu": 10,
    "memory": 128,
    "essential": true
  }
]
EOF
}
