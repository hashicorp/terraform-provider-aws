variable "random_name" {
  type    = string
  default = "tf-acc-test-2214292245511532902"
}

variable "apply_immediately" {
  type    = bool
  default = true
}

variable "authentication_strategy" {
  type    = string
  default = "ldap"
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

variable "hosts" {
  type    = list(string)
  default = ["my.ldap.server-1.com", "my.ldap.server-2.com"]
}

variable "role_base" {
  type    = string
  default = "role.base"
}

variable "role_name" {
  type    = string
  default = "role.name"
}

variable "role_search_matching" {
  type    = string
  default = "role.search.matching"
}

variable "role_search_subtree" {
  type    = bool
  default = true
}

variable "service_account_password" {
  type    = string
  default = "supersecret"
}

variable "service_account_username" {
  type    = string
  default = "anyusername"
}

variable "user_base" {
  type    = string
  default = "user.base"
}

variable "user_role_name" {
  type    = string
  default = "user.role.name"
}

variable "user_search_matching" {
  type    = string
  default = "user.search.matching"
}

variable "user_search_subtree" {
  type    = bool
  default = true
}

