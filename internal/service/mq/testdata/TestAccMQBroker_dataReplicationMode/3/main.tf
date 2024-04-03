provider "awsalternate" {
  region = var.region
}

resource "aws_security_group" "primary" {
  provider = awsalternate

  name = "${var.random_name}-primary"

  tags = {
    Name = "${var.random_name}-primary"
  }
}

resource "aws_mq_broker" "primary" {
  provider = awsalternate

  apply_immediately  = var.apply_immediately
  broker_name        = "${var.random_name}-primary"
  engine_type        = var.engine_type
  engine_version     = var.engine_version
  host_instance_type = var.host_instance_type
  security_groups    = [aws_security_group.primary.id]
  deployment_mode    = var.deployment_mode

  logs {
    general = var.general
  }

  user {
    username = var.username
    password = var.password
  }
  user {
    username         = var.username_2
    password         = var.password
    replication_user = var.replication_user
  }
}

resource "aws_security_group" "test" {
  name = var.random_name

  tags = {
    Name = var.random_name
  }
}

resource "aws_mq_broker" "test" {
  apply_immediately  = var.apply_immediately
  broker_name        = var.random_name
  engine_type        = var.engine_type
  engine_version     = var.engine_version
  host_instance_type = var.host_instance_type
  security_groups    = [aws_security_group.test.id]
  deployment_mode    = var.deployment_mode

  data_replication_mode               = var.data_replication_mode
  data_replication_primary_broker_arn = aws_mq_broker.primary.arn

  logs {
    general = var.general
  }

  user {
    username = var.username
    password = var.password
  }
  user {
    username         = var.username_3
    password         = var.password
    replication_user = var.replication_user
  }
}
