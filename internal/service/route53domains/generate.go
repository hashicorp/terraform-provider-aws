//go:generate go run ../../generate/tags/main.go -GetTag -ListTags -ListTagsOp=ListTagsForDomain -ListTagsInIDElem=DomainName -ListTagsOutTagsElem=TagList -UntagOp=DeleteTagsForDomain -UntagInTagsElem=TagsToDelete -TagOp=UpdateTagsForDomain -TagInIDElem=DomainName -TagInTagsElem=TagsToUpdate -ServiceTagsSlice -UpdateTags
// ONLY generate directives and package declaration! Do not add anything else to this file.

package route53domains
