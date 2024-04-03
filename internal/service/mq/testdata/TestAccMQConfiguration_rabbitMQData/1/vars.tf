variable "description" {
  type    = string
  default = "TfAccTest MQ Configuration"
}

variable "random_name" {
  type    = string
  default = "tf-acc-test-5927946624527564031"
}

variable "engine_type" {
  type    = string
  default = "RabbitMQ"
}

variable "engine_version" {
  type    = string
  default = "3.11.16"
}

