variable "description" {
  type    = string
  default = "TfAccTest MQ Configuration"
}

variable "random_name" {
  type    = string
  default = "tf-acc-test-3976901205937294897"
}

variable "engine_type" {
  type    = string
  default = "ActiveMQ"
}

variable "engine_version" {
  type    = string
  default = "5.17.6"
}

variable "authentication_strategy" {
  type    = string
  default = "simple"
}

