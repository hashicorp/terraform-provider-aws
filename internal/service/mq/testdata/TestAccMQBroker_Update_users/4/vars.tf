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

variable "username" {
  type    = string
  default = "second"
}

variable "password" {
  type    = string
  default = "TestTest2222"
}

variable "groups" {
  type    = list(string)
  default = ["admin"]
}

