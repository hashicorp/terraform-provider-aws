resource "aws_security_group" "test" {
  name = var.random_name

  tags = {
    Name = var.random_name
  }
}

resource "aws_mq_broker" "test" {
  broker_name        = var.random_name
  engine_type        = var.engine_type
  engine_version     = var.engine_version
  host_instance_type = var.host_instance_type
  security_groups    = [aws_security_group.test.id]
  storage_type       = var.storage_type
  deployment_mode    = var.deployment_mode

  user {
    username = var.username
    password = var.password
  }
}

data "aws_subnets" "default" {
  filter {
    name   = var.name
    values = [data.aws_vpc.default.id]
  }
}

data "aws_vpc" "default" {
  default = var.default
}
