resource "aws_sagemaker_hyper_parameter_tuning_job" "test" {
{{- template "region" }}
  hyper_parameter_tuning_job_name = var.rName

  hyper_parameter_tuning_job_config {
    strategy = "Bayesian"

    resource_limits {
      max_parallel_training_jobs  = 1
    }
  }

{{- template "tags" . }}
}
