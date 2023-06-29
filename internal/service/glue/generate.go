//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=GetTags -ServiceTagsMap -TagInTagsElem=TagsToAdd -UntagInTagsElem=TagsToRemove -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package glue
