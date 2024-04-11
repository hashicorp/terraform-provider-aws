module github.com/hashicorp/terraform-provider-aws/tools/awssdkpatch

go 1.21.8

replace github.com/hashicorp/terraform-provider-aws => ../..

require (
	github.com/hashicorp/terraform-provider-aws v0.0.0-00010101000000-000000000000
	golang.org/x/tools v0.20.0
)

require (
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
)
