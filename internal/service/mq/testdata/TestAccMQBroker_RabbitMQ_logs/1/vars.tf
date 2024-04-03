variable "random_name" {
  type    = string
  default = "tf-acc-test-2423182977932074655"
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
  default = "mq.t3.micro"
}

variable "general" {
  type    = bool
  default = true
}

variable "username" {
  type    = string
  default = "Test"
}

variable "password" {
  type    = string
  default = "TestTest1234"
}

