#
# Variables Configuration
#

variable "cluster-name" {
  default = "terraform-eks-demo"
  type    = "string"
}

variable "desired-workers" {
  default = "2"
  type    = "string"
}

variable "max-workers" {
  default = "2"
  type    = "string"
}

variable "min-workers" {
  default = "1"
  type    = "string"
}
