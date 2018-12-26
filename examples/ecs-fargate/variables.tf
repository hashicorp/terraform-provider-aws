variable "aws_region" {
  description = "The AWS region to create things in."
  default     = "eu-west-1"
}

variable "az_count" {
  description = "Number of AZs to cover in a given AWS region."
  default     = "2"
}

variable "vpc_cidr_block" {
  description = "The CIDR block for the VPC"
  default     = "10.10.0.0/16"
}

variable "app_docker_image" {
  description = "Docker image to use."
  default     = "ghost:latest"
}

variable "app_name" {
  description = "Name of application."
  default     = "ghost"
}

variable "app_port" {
  description = "Port the container listens to."
  default     = 2368
}

variable "task_cpu" {
  description = "The number of cpu units to use (256 is .25 vCPU)."
  default     = 256
}

variable "task_memory" {
  description = "The amount (in MiB) of memory to use."
  default     = 512
}

variable "service_desired" {
  description = "Desired numbers of instances in the ecs service."
  default     = "1"
}
