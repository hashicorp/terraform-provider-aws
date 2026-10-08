resource "aws_msk_serverless_cluster" "test" {
{{- template "region" }}
  cluster_name = var.rName

  client_authentication {
    sasl {
      iam {
        enabled = true
      }
    }
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

{{- template "tags" . }}
}

{{ template "acctest.ConfigSubnets" 2 }}

resource "aws_vpc" "test" {
{{- template "region" }}
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
}