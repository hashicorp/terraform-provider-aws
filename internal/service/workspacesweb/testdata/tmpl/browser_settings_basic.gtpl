resource "aws_workspacesweb_browser_settings" "test" {
  browser_policy = jsonencode({
    chromePolicies = {
      DefaultDownloadDirectory = {
        value = "/home/as2-streaming-user/MyFiles/TemporaryFiles1"
      }
    }
  })

{{- template "tags" . }}

}
