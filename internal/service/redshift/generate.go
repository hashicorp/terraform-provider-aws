//go:generate go run -tags generate ../../generate/tags/main.go -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceName -ServiceTagsSlice=yes -TagOp=CreateTags -TagInIDElem=ResourceName -UntagOp=DeleteTags -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package redshift
