//go:generate go run -tags generate ../../generate/tagresource/main.go -IDAttribName=resource_id
//go:generate go run -tags generate ../../generate/tags/main.go -GetTag=yes -ListTags=yes -ListTagsOp=DescribeTags -ListTagsInFiltIDName=resource-id -ListTagsInIDElem=Resources -ServiceTagsSlice=yes -TagOp=CreateTags -TagInIDElem=Resources -TagInIDNeedSlice=yes -TagType2=TagDescription -UntagOp=DeleteTags -UntagInNeedTagType=yes -UntagInTagsElem=Tags -UpdateTags=yes
//go:generate go run -tags generate generate/createtags/main.go

package ec2
