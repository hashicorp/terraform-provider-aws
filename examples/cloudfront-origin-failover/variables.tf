variable "aws_region" {
  description = "The AWS region to create things in."
  default     = "us-east-1"
}

variable "primary_bucket_name" {
  description = "The bucket name for the primary s3 bucket."
  default     = "terraform-testing-origin-failover"
}

variable "backup_bucket_name" {
  description = "The bucket name for the backup s3 bucket."
  default     = "terraform-testing-backup-origin-failover"
}
