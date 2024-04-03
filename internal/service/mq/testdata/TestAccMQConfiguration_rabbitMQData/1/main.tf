
resource "aws_mq_configuration" "test" {
  description    = var.description
  name           = var.random_name
  engine_type    = var.engine_type
  engine_version = var.engine_version

  data = <<DATA
consumer_timeout = 60000
DATA
}
