resource "aws_datasync_location_fsx_windows_file_system" "test" {
{{- template "region" }}
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test.arn]
{{- template "tags" . }}
}

# testAccLocationFSxWindowsFileSystemConfig_baseFS

resource "aws_fsx_windows_file_system" "test" {
{{- template "region" }}
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8
}

resource "aws_security_group" "test" {
{{- template "region" }}
  name   = var.rName
  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

# testAccLocationFSxWindowsFileSystemConfig_base

resource "aws_directory_service_directory" "test" {
{{- template "region" }}
  edition  = "Standard"
  name     = var.domain
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 2 }}
