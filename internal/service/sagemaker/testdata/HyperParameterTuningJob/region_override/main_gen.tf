# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
  region = var.region

  hyper_parameter_tuning_job_name = var.rName

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    resource_limits {
      max_parallel_training_jobs  = 1
    }
  }
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
