data "aws_mq_broker_instance_type_offerings" "empty" {}

data "aws_mq_broker_instance_type_offerings" "engine" {
  engine_type = var.engine_type
}

data "aws_mq_broker_instance_type_offerings" "storage" {
  storage_type = var.storage_type
}

data "aws_mq_broker_instance_type_offerings" "instance" {
  host_instance_type = var.host_instance_type
}

data "aws_mq_broker_instance_type_offerings" "all" {
  host_instance_type = var.host_instance_type
  storage_type       = var.storage_type
  engine_type        = var.engine_type
}
