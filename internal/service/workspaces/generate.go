//go:generate go run ../../generate/listpages/main.go -ListOps=DescribeIpGroups -ContextOnly
//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=TagList -ServiceTagsSlice -TagOp=CreateTags -TagInIDElem=ResourceId -UntagOp=DeleteTags -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package workspaces
