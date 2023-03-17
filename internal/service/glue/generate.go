//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOp=GetTags -ServiceTagsMap -TagInTagsElem=TagsToAdd -UntagInTagsElem=TagsToRemove -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package glue
