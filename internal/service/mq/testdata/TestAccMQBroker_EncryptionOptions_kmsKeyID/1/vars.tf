variable "random_name" {
  type    = string
  default = "tf-acc-test-6959298119541701151"
}

variable "deletion_window_in_days" {
  type    = number
  default = 7
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

variable "use_aws_owned_key" {
  type    = bool
  default = false
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

