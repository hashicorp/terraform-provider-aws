resource "aws_directory_service_directory_setting" "test" {
{{- template "region" }}
  directory_id = aws_directory_service_directory.test.id
  name         = "TLS_1_0"
  value        = "Disable"
}

resource "aws_directory_service_directory" "test" {
{{- template "region" }}
  name     = var.domain
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  size     = "Large"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
