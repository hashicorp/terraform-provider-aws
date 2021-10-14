//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=DescribeIpGroups -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=TagList -ServiceTagsSlice=yes -TagOp=CreateTags -TagInIDElem=ResourceId -UntagOp=DeleteTags -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package workspaces
