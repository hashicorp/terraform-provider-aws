resource "aws_directory_service_ip_routes_exclusive" "test" {
{{- template "region" }}
  directory_id = aws_directory_service_directory.test.id

  ip_route {
    cidr_ip     = "192.0.2.0/24"
    description = var.rName
  }
}

resource "aws_directory_service_directory" "test" {
{{- template "region" }}
  name     = var.domain
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
