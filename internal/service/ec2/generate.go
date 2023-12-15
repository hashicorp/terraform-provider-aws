// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/tagresource/main.go -IDAttribName=resource_id
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsOp=DescribeTags -ListTagsInFiltIDName=resource-id -ListTagsInIDElem=Resources -ServiceTagsSlice -TagOp=CreateTags -TagInIDElem=Resources -TagInIDNeedSlice=yes -TagType2=TagDescription -UntagOp=DeleteTags -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags
//go:generate go run ../../generate/tags/main.go -AWSSDKVersion=2 -ServiceTagsSlice -TagsFunc=TagsV2 -KeyValueTagsFunc=keyValueTagsV2 -GetTagsInFunc=getTagsInV2 -SetTagsOutFunc=setTagsOutV2 -TagOp=CreateTags -TagInIDElem=Resources -TagInIDNeedValueSlice=yes -UntagOp=DeleteTags -UpdateTagsFunc=updateTagsV2 -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags -- tagsv2_gen.go
//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeSpotFleetInstances,DescribeSpotFleetRequestHistory,DescribeVpcEndpointServices
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ec2
