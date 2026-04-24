# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_ecs_daemon" "test" {
  count = var.resource_count

  name                   = "${var.rName}-${count.index}"
  cluster                = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]
}

resource "aws_ecs_cluster" "test" {
  name = var.rName
}

resource "aws_ecs_daemon_task_definition" "test" {
  family             = var.rName
  execution_role_arn = aws_iam_role.test.arn

  container_definition {
    name   = "test"
    image  = "nginx:latest"
    essential = true
    memory = 128
  }
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

data "aws_partition" "current" {}

resource "aws_ecs_capacity_provider" "test" {
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
  cidr_block = "10.0.0.0/16"
  tags       = { Name = var.rName }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  vpc_id            = aws_vpc.test.id
  tags              = { Name = var.rName }
}

resource "aws_security_group" "test" {
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
  role       = aws_iam_role.infra.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRolePolicyForManagedInstances"
}

resource "aws_iam_role" "instance" {
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
  role       = aws_iam_role.instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "test" {
  name = var.rName
  role = aws_iam_role.instance.name
}

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "The number of resources to create"
  type        = number
  nullable    = false
}
