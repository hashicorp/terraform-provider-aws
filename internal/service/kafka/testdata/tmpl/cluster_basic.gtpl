resource "aws_msk_cluster" "test" {
{{- template "region" }}
  cluster_name           = var.rName
  kafka_version          = "3.8.x"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.test.id]

    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }

{{- template "tags" . }}
}

{{ template "acctest.ConfigVPCWithSubnets" 3 }}

resource "aws_security_group" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}