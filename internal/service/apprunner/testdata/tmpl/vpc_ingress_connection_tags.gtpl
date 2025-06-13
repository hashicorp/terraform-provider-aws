resource "aws_apprunner_vpc_ingress_connection" "test" {
{{- template "region" }}
  name        = var.rName
  service_arn = aws_apprunner_service.test.arn

  ingress_vpc_configuration {
    vpc_id          = aws_vpc.test.id
    vpc_endpoint_id = aws_vpc_endpoint.test.id
  }

{{- template "tags" . }}
}

# testAccVPCIngressConnectionConfig_base

data "aws_region" "current" {
{{- template "region" }}
}

resource "aws_apprunner_service" "test" {
{{- template "region" }}
  service_name = var.rName

  source_configuration {
    image_repository {
      image_configuration {
        port = "8000"
      }
      image_identifier      = "public.ecr.aws/aws-containers/hello-app-runner:latest"
      image_repository_type = "ECR_PUBLIC"
    }
    auto_deployments_enabled = false
  }

  network_configuration {
    ingress_configuration {
      is_publicly_accessible = false
    }
  }
}

resource "aws_vpc_endpoint" "test" {
{{- template "region" }}
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.apprunner.requests"
  vpc_endpoint_type = "Interface"

  subnet_ids = aws_subnet.test[*].id

  security_group_ids = [
    aws_vpc.test.default_security_group_id,
  ]
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
