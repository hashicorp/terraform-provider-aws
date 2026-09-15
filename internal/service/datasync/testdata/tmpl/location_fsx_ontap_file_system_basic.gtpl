resource "aws_datasync_location_fsx_ontap_file_system" "test" {
{{- template "region" }}
  security_group_arns         = [aws_security_group.test.arn]
  storage_virtual_machine_arn = aws_fsx_ontap_storage_virtual_machine.test.arn

  protocol {
    nfs {
      mount_options {
        version = "NFS3"
      }
    }
  }
{{- template "tags" . }}
}

# testAccFSxOntapFileSystemConfig_baseNFS

resource "aws_fsx_ontap_storage_virtual_machine" "test" {
{{- template "region" }}
  file_system_id = aws_fsx_ontap_file_system.test.id
  name           = var.rName
}

# testAccFSxOntapFileSystemConfig_base(rName, 1, "SINGLE_AZ_1")

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

resource "aws_fsx_ontap_file_system" "test" {
{{- template "region" }}
  storage_capacity    = 1024
  subnet_ids          = aws_subnet.test[*].id
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 512
  preferred_subnet_id = aws_subnet.test[0].id
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
