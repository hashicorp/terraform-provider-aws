variable "random_name" {
  type    = string
  default = "tf-acc-test-6399918457541333348"
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

variable "console_access" {
  type    = bool
  default = true
}

variable "username" {
  type    = string
  default = "first"
}

variable "password" {
  type    = string
  default = "TestTest1111updated"
}

variable "username_2" {
  type    = string
  default = "second"
}

variable "password_2" {
  type    = string
  default = "TestTest2222"
}

