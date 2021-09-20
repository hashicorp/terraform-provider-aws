//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=DescribeIpGroups -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceId -ListTagsOutTagsElem=TagList -ServiceTagsSlice=yes -TagOp=CreateTags -TagInIDElem=ResourceId -UntagOp=DeleteTags -UpdateTags=yes

package workspaces
