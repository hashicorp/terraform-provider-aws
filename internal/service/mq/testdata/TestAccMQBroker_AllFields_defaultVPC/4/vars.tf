variable "vcount" {
  type    = number
  default = 2
}

variable "random_name" {
  type    = string
  default = "tf-acc-test-5643589426675988144"
}

variable "random_name_2" {
  type    = string
  default = "tf-acc-test-3815919829857381909"
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
  default = false
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

