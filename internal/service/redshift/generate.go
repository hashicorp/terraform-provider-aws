//go:generate go run ../../generate/tags/main.go -ListTagsOp=DescribeTags -ListTagsInIDElem=ResourceName -ServiceTagsSlice -TagOp=CreateTags -TagInIDElem=ResourceName -UntagOp=DeleteTags -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package redshift
