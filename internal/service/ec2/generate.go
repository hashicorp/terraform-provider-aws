// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tagresource/main.go -IDAttribName=resource_id
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsOp=DescribeTags -ListTagsOpPaginated -ListTagsInFiltIDName=resource-id -ServiceTagsSlice -TagOp=CreateTags -TagInIDElem=Resources -TagInIDNeedValueSlice -TagType2=TagDescription -UntagOp=DeleteTags -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags
//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeSpotFleetInstances,DescribeSpotFleetRequestHistory,DescribeVpcBlockPublicAccessExclusions,DescribeVpcEndpointAssociations,DescribeVpcEndpointServices,GetSubnetCidrReservations
//go:generate go run ../../generate/paginator/main.go -ListOp=DescribeVpcEncryptionControls
//go:generate go run ../../generate/paginator/main.go -ListOp=GetVpcResourcesBlockingEncryptionEnforcement
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
//go:generate go run ../../generate/identitytests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ec2
