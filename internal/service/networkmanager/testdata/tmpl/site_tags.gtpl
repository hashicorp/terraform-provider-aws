resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
{{- template "tags" . }}
}

resource "aws_networkmanager_global_network" "test" {}
