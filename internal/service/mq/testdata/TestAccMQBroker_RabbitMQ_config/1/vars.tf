variable "random_name" {
  type    = string
  default = "tf-acc-test-8772785443490268916"
}

variable "description" {
  type    = string
  default = "TfAccTest MQ Configuration"
}

variable "engine_type" {
  type    = string
  default = "RabbitMQ"
}

variable "engine_version" {
  type    = string
  default = "3.11.20"
}

variable "consumer_timeout" {
  type    = number
  default = 1800000
}

variable "host_instance_type" {
  type    = string
  default = "mq.t3.micro"
}

variable "username" {
  type    = string
  default = "Test"
}

variable "password" {
  type    = string
  default = "TestTest1234"
}

