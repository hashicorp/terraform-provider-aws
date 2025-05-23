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

# acctest.ConfigVPCWithSubnets(rName, 1)

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
{{- template "region" }}
  count = 1

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude()

data "aws_availability_zones" "available" {
{{- template "region" }}
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}
