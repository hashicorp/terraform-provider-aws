
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

  user {
    username = var.username
    password = var.password
  }

  tags = {
    key2 = var.key2_value
  }
}
