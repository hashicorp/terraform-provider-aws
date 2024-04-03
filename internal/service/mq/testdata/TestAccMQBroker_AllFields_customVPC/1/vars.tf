variable "exclude_zone_ids" {
  type    = list(string)
  default = ["usw2-az4", "usgw1-az2"]
}

variable "state" {
  type    = string
  default = "available"
}

variable "name" {
  type    = string
  default = "opt-in-status"
}

variable "values" {
  type    = list(string)
  default = ["opt-in-not-required"]
}

variable "cidr_block" {
  type    = string
  default = "10.0.0.0/16"
}

variable "random_name" {
  type    = string
  default = "tf-acc-test-5865867048482345256"
}

variable "vcount" {
  type    = number
  default = 2
}

variable "cidr_block_2" {
  type    = string
  default = "0.0.0.0/0"
}

variable "engine_type" {
  type    = string
  default = "ActiveMQ"
}

variable "engine_version" {
  type    = string
  default = "5.17.6"
}

variable "auto_minor_version_upgrade" {
  type    = bool
  default = true
}

variable "apply_immediately" {
  type    = bool
  default = true
}

variable "deployment_mode" {
  type    = string
  default = "ACTIVE_STANDBY_MULTI_AZ"
}

variable "storage_type" {
  type    = string
  default = "efs"
}

variable "host_instance_type" {
  type    = string
  default = "mq.t2.micro"
}

variable "general" {
  type    = bool
  default = true
}

variable "audit" {
  type    = bool
  default = true
}

variable "day_of_week" {
  type    = string
  default = "TUESDAY"
}

variable "time_of_day" {
  type    = string
  default = "02:00"
}

variable "time_zone" {
  type    = string
  default = "CET"
}

variable "publicly_accessible" {
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
  default = "SecondTest"
}

variable "password_2" {
  type    = string
  default = "SecondTestTest1234"
}

variable "console_access" {
  type    = bool
  default = true
}

variable "groups" {
  type    = list(string)
  default = ["first", "second", "third"]
}

