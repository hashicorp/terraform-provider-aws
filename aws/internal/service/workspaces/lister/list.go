//go:generate go run ../../../generators/listpages/main.go -function=DescribeIpGroups -paginator=NextToken github.com/aws/aws-sdk-go/service/workspaces

package lister
