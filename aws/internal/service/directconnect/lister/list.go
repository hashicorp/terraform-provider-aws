//go:generate go run ../../../generators/listpages/main.go -function=DescribeDirectConnectGateways,DescribeDirectConnectGatewayAssociations,DescribeDirectConnectGatewayAssociationProposals -paginator=NextToken github.com/aws/aws-sdk-go/service/directconnect

package lister
