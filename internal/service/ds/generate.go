//go:generate go run -tags generate ../../generate/listpages/main.go -ListOps=DescribeDirectories -Export=yes
//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsInIDElem=ResourceId -ServiceTagsSlice=yes -TagOp=AddTagsToResource -TagInIDElem=ResourceId -UntagOp=RemoveTagsFromResource -UpdateTags=yes

package ds
