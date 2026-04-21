resource "aws_ecs_daemon" "test" {
{{- template "region" }}
  name                   = var.rName
  cluster                = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]
{{- template "tags" . }}
}

resource "aws_ecs_cluster" "test" {
{{- template "region" }}
  name = var.rName
}

resource "aws_ecs_daemon_task_definition" "test" {
{{- template "region" }}
  family             = var.rName
  execution_role_arn = aws_iam_role.test.arn

  container_definition {
    name   = "test"
    image  = "nginx:latest"
    memory = 128
  }
}

resource "aws_iam_role" "test" {
{{- template "region" }}
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

data "aws_partition" "current" {
{{- template "region" }}
}

resource "aws_ecs_capacity_provider" "test" {
{{- template "region" }}
  name    = var.rName
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.infra.arn

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = [aws_subnet.test.id]
        security_groups = [aws_security_group.test.id]
      }
    }
  }
}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
  tags = { Name = var.rName }
}

resource "aws_subnet" "test" {
{{- template "region" }}
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  vpc_id            = aws_vpc.test.id
  tags = { Name = var.rName }
}

resource "aws_security_group" "test" {
{{- template "region" }}
  name   = var.rName
  vpc_id = aws_vpc.test.id
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = { Name = var.rName }
}

resource "aws_iam_role" "infra" {
{{- template "region" }}
  name = "${var.rName}-infra"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "infra" {
{{- template "region" }}
  role       = aws_iam_role.infra.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AdministratorAccess"
}

resource "aws_iam_role" "instance" {
{{- template "region" }}
  name = "${var.rName}-instance"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "instance" {
{{- template "region" }}
  role       = aws_iam_role.instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "test" {
{{- template "region" }}
  name = var.rName
  role = aws_iam_role.instance.name
}

{{ template "acctest.ConfigAvailableAZsNoOptInDefaultExclude" }}
