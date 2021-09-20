//go:generate go run -tags generate ../../generate/tags/main.go -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceName -ServiceTagsSlice=yes -TagOp=CreateTags -TagInIDElem=ResourceName -UntagOp=DeleteTags -UpdateTags=yes

package redshift
