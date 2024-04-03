variable "random_name" {
  type    = string
  default = "tf-acc-test-540622394625720983"
}

variable "engine_type" {
  type    = string
  default = "ActiveMQ"
}

variable "engine_version" {
  type    = string
  default = "5.17.6"
}

variable "storage_type" {
  type    = string
  default = "ebs"
}

variable "host_instance_type" {
  type    = string
  default = "mq.m5.large"
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

