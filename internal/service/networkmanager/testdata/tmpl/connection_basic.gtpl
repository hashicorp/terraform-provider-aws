resource "aws_networkmanager_connection" "test" {
  global_network_id   = aws_networkmanager_global_network.test.id
  device_id           = aws_networkmanager_device.test1.id
  connected_device_id = aws_networkmanager_device.test2.id
{{- template "tags" . }}
}

# testAccConnectionBaseConfig

resource "aws_networkmanager_global_network" "test" {}

resource "aws_networkmanager_site" "test" {
  global_network_id = aws_networkmanager_global_network.test.id
}

resource "aws_networkmanager_device" "test1" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id
}

resource "aws_networkmanager_device" "test2" {
  global_network_id = aws_networkmanager_global_network.test.id
  site_id           = aws_networkmanager_site.test.id

  # Create one device at a time.
  depends_on = [aws_networkmanager_device.test1]
}
