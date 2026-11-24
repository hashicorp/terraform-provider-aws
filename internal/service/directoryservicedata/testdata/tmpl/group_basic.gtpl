resource "aws_directoryservicedata_group" "test" {
{{ template "region" }}
  directory_id     = aws_directory_service_directory.test.id
  sam_account_name = "tf-acc-test-group"
  group_scope      = "Global"
  group_type       = "Security"
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
