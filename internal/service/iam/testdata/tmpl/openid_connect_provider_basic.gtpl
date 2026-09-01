resource "aws_iam_openid_connect_provider" "test" {
  url = "https://accounts.testle.com/${var.rName}"

  client_id_list = [
    "266362248691-re108qaeld573ia0l6clj2i5ac7r7291.apps.testleusercontent.com",
  ]

  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]
{{- template "tags" . }}
}
