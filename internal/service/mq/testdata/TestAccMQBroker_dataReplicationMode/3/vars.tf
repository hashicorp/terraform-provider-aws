variable "region" {
  type    = string
  default = "us-east-1"
}

variable "random_name" {
  type    = string
  default = "tf-acc-test-1492773231363384190"
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
  default = "mq.m5.large"
}

variable "deployment_mode" {
  type    = string
  default = "ACTIVE_STANDBY_MULTI_AZ"
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

variable "username_2" {
  type    = string
  default = "Test-ReplicationUser"
}

variable "replication_user" {
  type    = bool
  default = true
}

variable "data_replication_mode" {
  type    = string
  default = "NONE"
}

variable "username_3" {
  type    = string
  default = "Test-ReplicationUser"
}

