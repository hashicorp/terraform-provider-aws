resource "aws_apprunner_service" "test" {
{{- template "region" }}
  service_name = var.rName
  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }

{{- template "tags" . }}
}
