variable "description" {
  type    = string
  default = "TfAccTest MQ Configuration"
}

variable "random_name" {
  type    = string
  default = "tf-acc-test-8333933185166084985"
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

variable "key1_value" {
  type    = string
  default = "value1updated"
}

variable "key2_value" {
  type    = string
  default = "value2"
}

