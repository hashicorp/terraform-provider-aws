# Route53 variables
variable "hostedzone_id" {
  description = "value of hosted zone id"
  type        = string
}

variable "record_name" {
  description = "value of record name"
  type        = string
}

variable "loadbalancer_name" {
  description = "value of load balancer name"
  type        = string
}

variable "loadbalancer_zoneid" {
  description = "value of load balancer zone id"
  type        = string
}

# Target group variables
variable "target_group_name" {
  description = "value of target group name"
  type        = string
}

variable "target_group_port" {
  description = "value of target group port"
  type        = number
}

variable "vpc_id" {
  description = "value of vpc id"
  type        = string
}

variable "target_ins_ip" {
  description = "value of target instance ip"
  type        = string
}

# ALB variables
variable "alb_https_listener_arn" {
  description = "value of alb https listener arn"
  type        = string
}

variable "rule_priority" {
  description = "value of rule priority"
  type        = string
}