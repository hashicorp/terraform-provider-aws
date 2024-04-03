variable "random_name" {
  type    = string
  default = "tf-acc-test-6224304883743208792"
}

variable "engine_type" {
  type    = string
  default = "RabbitMQ"
}

variable "engine_version" {
  type    = string
  default = "3.11.20"
}

variable "host_instance_type" {
  type    = string
  default = "mq.m5.large"
}

variable "storage_type" {
  type    = string
  default = "ebs"
}

variable "deployment_mode" {
  type    = string
  default = "CLUSTER_MULTI_AZ"
}

variable "username" {
  type    = string
  default = "Test"
}

variable "password" {
  type    = string
  default = "TestTest1234"
}

variable "name" {
  type    = string
  default = "vpc-id"
}

variable "default" {
  type    = bool
  default = true
}

