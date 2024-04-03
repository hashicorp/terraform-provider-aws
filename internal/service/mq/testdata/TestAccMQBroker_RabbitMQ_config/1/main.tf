resource "aws_security_group" "test" {
  name = var.random_name

  tags = {
    Name = var.random_name
  }
}

resource "aws_mq_configuration" "test" {
  description    = var.description
  name           = var.random_name
  engine_type    = var.engine_type
  engine_version = var.engine_version

  data = <<DATA
  # Default RabbitMQ delivery acknowledgement timeout is 30 minutes
  consumer_timeout = var.consumer_timeout
  
  DATA
}

resource "aws_mq_broker" "test" {
  broker_name        = var.random_name
  engine_type        = var.engine_type
  engine_version     = var.engine_version
  host_instance_type = var.host_instance_type
  security_groups    = [aws_security_group.test.id]

  configuration {
    id       = aws_mq_configuration.test.id
    revision = aws_mq_configuration.test.latest_revision
  }

  user {
    username = var.username
    password = var.password
  }
}
