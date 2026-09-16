resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = var.domain
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}

resource "aws_directoryservicedata_user" "test" {
  directory_id     = aws_directory_service_directory.test.id
  sam_account_name = "testuser"
  email_address    = "testuser@example.com"
  given_name       = "Test"
  surname          = "User"
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}