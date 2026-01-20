resource "aws_networkmonitor_probe" "test" {
  monitor_name = aws_networkmonitor_monitor.test.monitor_name
  destination  = "10.0.0.1"
  protocol     = "ICMP"
  source_arn   = aws_subnet.test[0].arn
{{- template "tags" . }}
}

# testAccProbeConfig_base

resource "aws_networkmonitor_monitor" "test" {
  monitor_name = var.rName
}

{{ template "acctest.ConfigVPCWithSubnets" 1 }}
