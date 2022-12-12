//go:generate go run ../../generate/tagresource/main.go -IDAttribName=resource_id
//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsOp=DescribeTags -ListTagsInFiltIDName=resource-id -ListTagsInIDElem=Resources -ServiceTagsSlice -TagOp=CreateTags -TagInIDElem=Resources -TagInIDNeedSlice=yes -TagType2=TagDescription -UntagOp=DeleteTags -UntagInNeedTagType -UntagInTagsElem=Tags -UpdateTags
//go:generate go run generate/createtags/main.go
//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeSpotFleetInstances,DescribeSpotFleetRequestHistory,DescribeVpcEndpointServices
// ONLY generate directives and package declaration! Do not add anything else to this file.

package ec2
