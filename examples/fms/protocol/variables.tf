variable "aws_region" {
  description = "The AWS region to create resources in."
  type        = string
  default     = "us-east-1"
}

variable "list_name" {
  description = "The name of the FMs Protocol List"
  type        = string
  default     = "tf-example-fms-protocol-list"
}

variable "list_protocols" {
  description = "An array of protocols in the AWS Firewall Manager protocols list."
  type        = list(string)
  default     = ["IPv4", "IPv6", "ICMP"]
}