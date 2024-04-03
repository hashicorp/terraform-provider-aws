
resource "aws_kms_key" "test" {
  description             = var.random_name
  deletion_window_in_days = var.deletion_window_in_days
}

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

  encryption_options {
    kms_key_id        = aws_kms_key.test.arn
    use_aws_owned_key = var.use_aws_owned_key
  }

  logs {
    general = var.general
  }

  user {
    username = var.username
    password = var.password
  }
}
