//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=DescribeDirectConnectGateways,DescribeDirectConnectGatewayAssociations,DescribeDirectConnectGatewayAssociationProposals -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceArns -ListTagsInIDNeedSlice=yes -ListTagsOutTagsElem=ResourceTags[0].Tags -ServiceTagsSlice=yes -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package directconnect
