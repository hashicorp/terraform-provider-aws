resource "aws_ec2_network_insights_access_scope" "test" {
{{- template "region" }}

  match_paths {
    source {
      resource_statement {
        resource_types = ["AWS::EC2::NetworkInterface"]
      }
    }
  }

{{- template "tags" . }}
}
