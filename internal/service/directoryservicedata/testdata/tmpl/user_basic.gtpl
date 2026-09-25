resource "aws_directoryservicedata_user" "test" {
{{ template "region" }}
  directory_id     = aws_directory_service_directory.test.id
  sam_account_name = "tfacctest-user"
  email_address    = var.emailAddress
  given_name       = "Test"
  surname          = "User"
}

resource "aws_directory_service_directory" "test" {
{{ template "region" }}
  edition                      = "Standard"
  name                         = var.directoryDomain
  password                     = "SuperSecretPassw0rd"
  type                         = "MicrosoftAD"
  enable_directory_data_access = true

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
