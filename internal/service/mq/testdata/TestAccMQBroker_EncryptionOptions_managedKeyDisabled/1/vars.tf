variable "random_name" {
  type    = string
  default = "tf-acc-test-298925321046406872"
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

