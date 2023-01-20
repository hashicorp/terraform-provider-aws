//go:generate go run ../../generate/tags/main.go -ListTags -ListTagsOutTagsElem=ResourceTags -ServiceTagsSlice -TagOp=UpdateTagsForResource -TagInTagsElem=TagsToAdd -UntagOp=UpdateTagsForResource -UntagInTagsElem=TagsToRemove -UpdateTags -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elasticbeanstalk
