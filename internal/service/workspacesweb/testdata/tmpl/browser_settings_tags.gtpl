locals {
    var1 = "{"
    var2 = <<EOF
        "chromePolicies": {
        "DefaultDownloadDirectory": {
            "value": "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
        }
    }
}
EOF
}
resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy = "${local.var1} ${local.var2}"

{{- template "tags" . }}

}
