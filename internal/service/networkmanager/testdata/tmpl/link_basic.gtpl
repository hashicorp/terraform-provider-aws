resource "aws_networkmanager_link" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  bandwidth {
    download_speed = 50
    upload_speed   = 10
  }
{{- template "tags" . }}
}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_global_network" "test" {}
