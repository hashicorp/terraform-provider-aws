resource "aws_ecs_daemon_task_definition" "test" {
{{- template "region" }}
  family             = var.rName
  execution_role_arn = aws_iam_role.test.arn

  container_definition {
    name      = "test"
    image     = "nginx:latest"
    essential = true
    memory    = 128
  }
{{- template "tags" . }}
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}
