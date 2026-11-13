resource "aws_ecs_service" "test" {
{{- template "region" }}
  name            = var.rName
  cluster         = aws_ecs_cluster.test.arn
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
}

resource "aws_ecs_cluster" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_ecs_task_definition" "test" {
{{- template "region" }}
  family = var.rName

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}
