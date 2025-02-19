variable "enable_delay_after_policy_creation" {
  description = "Indicates whether to enable a delay after policy creation"
  type        = bool
  default     = false
}

variable "delay_after_policy_creation" {
  description = "Duration of the delay after policy creation, specified in seconds"
  type        = number
  default     = 3
}