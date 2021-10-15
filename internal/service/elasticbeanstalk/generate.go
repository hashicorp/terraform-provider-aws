//go:generate go run -tags generate ../../generate/tags/main.go -ListTags=yes -ListTagsOutTagsElem=ResourceTags -ServiceTagsSlice=yes -TagOp=UpdateTagsForResource -TagInTagsElem=TagsToAdd -UntagOp=UpdateTagsForResource -UntagInTagsElem=TagsToRemove -UpdateTags=yes
// ONLY generate directives and package declaration! Do not add anything else to this file.

package elasticbeanstalk
