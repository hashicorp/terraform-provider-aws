variable "random_name" {
  type    = string
  default = "tf-acc-test-5435825179708208917"
}

variable "apply_immediately" {
  type    = bool
  default = true
}

variable "engine_type" {
  type    = string
  default = "ActiveMQ"
}

variable "engine_version" {
  type    = string
  default = "5.17.6"
}

variable "host_instance_type" {
  type    = string
  default = "mq.t2.micro"
}

variable "username" {
  type    = string
  default = "Test"
}

variable "password" {
  type    = string
  default = "TestTest1234"
}

variable "key1_value" {
  type    = string
  default = "value1"
}

