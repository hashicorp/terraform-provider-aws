// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tagresource/main.go -IDAttribName=resource_id
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -GetTag -ListTags -ListTagsOp=DescribeTags -ListTagsOpPaginated -ListTagsInFiltIDName=resource-id -ServiceTagsSlice -KeyValueTagsFunc=keyValueTags -TagOp=CreateTags -TagInIDElem=Resources -TagInIDNeedValueSlice=yes -TagType2=TagDescription -UntagOp=DeleteTags -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags
//go:generate go run ../../generate/listpages/main.go -AWSSDKVersion=2 -ListOps=DescribeSpotFleetInstances,DescribeSpotFleetRequestHistory,DescribeVpcEndpointServices
//go:generate go run ../../generate/servicepackage/main.go
//go:generate go run ../../generate/tagstests/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ec2
